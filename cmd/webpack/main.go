// Command webpack builds the browser test page for goloco's js/wasm build:
// the gzipped wasm, wasm_exec.js, the page (web/), and a data pack — the files
// goloco reads at run time, packed into gzipped chunks that web/index.html
// unpacks into the in-memory filesystem of web/fsshim.js.
//
// The output contains the Locomotion game data, which is NOT ours: it goes to
// a directory outside the repo and may only be published as a private test
// page (see LOOP.md). Music is left out by default to keep the pack small.
//
//	go run ./cmd/webpack -out $TMPDIR/goloco-web
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// maxChunk is the largest chunk file allowed. The private page host serves
// neither gzip nor arbitrary binary types, but it serves text up to 16 MB per
// file, so each chunk ships as base64 text of its gzip (≈ +33%).
const maxChunk = 15_000_000

type entry struct {
	Path   string `json:"p"` // absolute path in the virtual filesystem
	Chunk  int    `json:"c"`
	Offset int    `json:"o"`
	Size   int    `json:"n"`
}

type manifest struct {
	Chunks []string `json:"chunks"`
	Files  []entry  `json:"files"`
	Args   []string `json:"args"`
}

// mount maps a host directory onto a virtual one, with a filter.
type mount struct {
	host, virt string
	keep       func(rel string, size int64) bool
}

func main() {
	out := flag.String("out", "", "output directory (outside the repo)")
	data := flag.String("data", "../locomotion", "Locomotion data directory")
	music := flag.Bool("music", false, "include the music files (Data/*.dat over 3 MB)")
	scenarios := flag.String("scenarios", "", "comma-separated scenario file names to include (default: all)")
	chunkRaw := flag.Int("chunk", 18<<20, "raw bytes per chunk before gzip")
	args := flag.String("args", "", "space-separated command-line arguments passed to goloco")
	flag.Parse()
	if *out == "" {
		log.Fatal("webpack: -out is required")
	}
	if abs, _ := filepath.Abs(*out); strings.HasPrefix(abs+"/", mustAbs(".")+"/") {
		log.Fatal("webpack: -out must be outside the repo (the pack holds game data)")
	}
	must(os.MkdirAll(*out, 0o755))

	wanted := map[string]bool{}
	for _, s := range strings.Split(*scenarios, ",") {
		if s = strings.TrimSpace(s); s != "" {
			wanted[strings.ToLower(s)] = true
		}
	}
	mounts := []mount{
		{"data", "/w/goloco/data", nil},
		{"assets", "/w/goloco/assets", nil},
		{filepath.Join(*data, "Data"), "/w/locomotion/Data", func(rel string, size int64) bool {
			return *music || size < 3<<20
		}},
		{filepath.Join(*data, "ObjData"), "/w/locomotion/ObjData", nil},
		{filepath.Join(*data, "Scenarios"), "/w/locomotion/Scenarios", func(rel string, size int64) bool {
			return len(wanted) == 0 || wanted[strings.ToLower(filepath.Base(rel))]
		}},
	}

	type file struct{ host, virt string }
	var list []file
	for _, m := range mounts {
		err := filepath.WalkDir(m.host, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(m.host, p)
			if m.keep != nil && !m.keep(rel, info.Size()) {
				return nil
			}
			list = append(list, file{p, m.virt + "/" + filepath.ToSlash(rel)})
			return nil
		})
		if err != nil {
			log.Fatalf("webpack: %s: %v", m.host, err)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].virt < list[j].virt })

	var man manifest
	if *args != "" {
		man.Args = strings.Fields(*args)
	}
	var raw bytes.Buffer
	flush := func() {
		if raw.Len() == 0 {
			return
		}
		name := fmt.Sprintf("pack-%d.b64.txt", len(man.Chunks))
		gz := b64(gzipBytes(raw.Bytes()))
		if len(gz) > maxChunk {
			log.Fatalf("webpack: %s is %d bytes gzipped (max %d) — lower -chunk", name, len(gz), maxChunk)
		}
		must(os.WriteFile(filepath.Join(*out, name), gz, 0o644))
		log.Printf("%s: %.1f MB raw → %.1f MB", name, float64(raw.Len())/1e6, float64(len(gz))/1e6)
		man.Chunks = append(man.Chunks, name)
		raw.Reset()
	}
	for _, f := range list {
		b, err := os.ReadFile(f.host)
		must(err)
		if raw.Len() > 0 && raw.Len()+len(b) > *chunkRaw {
			flush()
		}
		man.Files = append(man.Files, entry{f.virt, len(man.Chunks), raw.Len(), len(b)})
		raw.Write(b)
	}
	flush()
	mj, err := json.Marshal(man)
	must(err)
	must(os.WriteFile(filepath.Join(*out, "manifest.json"), mj, 0o644))
	log.Printf("pack: %d files in %d chunks", len(man.Files), len(man.Chunks))

	// The wasm itself, gzipped and base64-encoded (21 MB raw; the page unpacks it).
	wasm := filepath.Join(*out, "goloco.wasm")
	build := exec.Command("go", "build", "-ldflags=-s -w", "-o", wasm, "./cmd/goloco")
	build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	must(build.Run())
	wb, err := os.ReadFile(wasm)
	must(err)
	must(os.WriteFile(filepath.Join(*out, "goloco.wasm.b64.txt"), b64(gzipBytes(wb)), 0o644))
	must(os.Remove(wasm))

	for _, f := range []struct{ from, to string }{
		{filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js"), "wasm_exec.js"},
		{"web/fsshim.js", "fsshim.js"},
		{"web/index.html", "index.html"},
	} {
		b, err := os.ReadFile(f.from)
		must(err)
		must(os.WriteFile(filepath.Join(*out, f.to), b, 0o644))
	}
	log.Printf("webpack: done → %s", *out)
}

// b64 encodes b as base64 text (the page decodes, then gunzips).
func b64(b []byte) []byte {
	out := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(out, b)
	return out
}

func gzipBytes(b []byte) []byte {
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	_, _ = zw.Write(b)
	must(zw.Close())
	return buf.Bytes()
}

func mustAbs(p string) string {
	a, err := filepath.Abs(p)
	must(err)
	return a
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
