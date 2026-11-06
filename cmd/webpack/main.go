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
	Wasm   string   `json:"wasm"` // file holding the game code
	Lazy   []lazy   `json:"lazy"` // files fetched on first open (music)
	Open   bool     `json:"open"` // the open build: no game data; the page asks for the player's own
}

// lazy is a file the page does not download up front: the fs shim fetches
// URL (a .gz) the first time the game opens Path.
type lazy struct {
	Path string `json:"p"`
	URL  string `json:"u"`
	Size int    `json:"n"`
}

// mount maps a host directory onto a virtual one, with a filter.
type mount struct {
	host, virt string
	keep       func(rel string, size int64) bool
}

func main() {
	out := flag.String("out", "", "output directory (outside the repo)")
	data := flag.String("data", "../locomotion", "Locomotion data directory")
	music := flag.Bool("music", false, "include the music files (Data/*.dat over 3 MB), fetched on first use (needs -local)")
	scenarios := flag.String("scenarios", "", "comma-separated scenario file names to include (default: all)")
	chunkRaw := flag.Int("chunk", 18<<20, "raw bytes per chunk before gzip")
	args := flag.String("args", "", "space-separated command-line arguments passed to goloco")
	local := flag.Bool("local", false, "write plain .gz files for a local server (the artifact host needs base64 text)")
	open := flag.Bool("open", false, "the open build (GitHub Pages): goloco's own files only, no Locomotion data — the page takes the player's own folder")
	og := flag.String("opengraphics", "", "with -open: the unzipped OpenGraphics objects.zip (MIT; OG_*.dat only), packed flat at /w/opengraphics/ObjData — the fallback when the player brings no folder")
	flag.Parse()
	if *out == "" {
		log.Fatal("webpack: -out is required")
	}
	if *og != "" && !*open {
		log.Fatal("webpack: -opengraphics is for the open build (-open)")
	}
	if *open && *music {
		log.Fatal("webpack: -open never includes game data (the music is the player's own)")
	}
	if *music && !*local {
		log.Fatal("webpack: -music needs -local (the artifact host's 64 MB limit can't hold the music)")
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
	}
	if !*open { // the player's data (never in the open build)
		mounts = append(mounts, []mount{
			{filepath.Join(*data, "Data"), "/w/locomotion/Data", func(rel string, size int64) bool {
				return size < 3<<20
			}},
			{filepath.Join(*data, "ObjData"), "/w/locomotion/ObjData", nil},
			{filepath.Join(*data, "Scenarios"), "/w/locomotion/Scenarios", func(rel string, size int64) bool {
				return len(wanted) == 0 || wanted[strings.ToLower(filepath.Base(rel))]
			}},
		}...)
	}

	type file struct{ host, virt string }
	var list []file
	if *og != "" { // OpenGraphics, flattened (the object loader reads one folder); nothing but OG_*.dat
		err := filepath.WalkDir(*og, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			name := filepath.Base(p)
			if !strings.HasPrefix(name, "OG_") || !strings.EqualFold(filepath.Ext(name), ".dat") {
				log.Fatalf("webpack: %s in the OpenGraphics folder is not an OG_*.dat object — refusing (no original data in the open build)", p)
			}
			list = append(list, file{p, "/w/opengraphics/ObjData/" + name})
			return nil
		})
		if err != nil {
			log.Fatalf("webpack: %s: %v", *og, err)
		}
	}
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
	man.Open = *open
	if *args != "" {
		man.Args = strings.Fields(*args)
	}
	// The artifact host serves text but not gzip, so its chunks are base64 text
	// of gzip; a local server can hand out the .gz files as they are.
	ext, enc := ".b64.txt", b64
	if *local {
		ext, enc = ".gz", func(b []byte) []byte { return b }
	}
	man.Wasm = "goloco.wasm" + ext
	var raw bytes.Buffer
	flush := func() {
		if raw.Len() == 0 {
			return
		}
		name := fmt.Sprintf("pack-%d%s", len(man.Chunks), ext)
		gz := enc(gzipBytes(raw.Bytes()))
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
	if *music {
		// Music tracks (Data/*.dat of 3 MB and more): one .gz each, fetched by the
		// page only when the game opens them, so starting the game stays light.
		must(os.RemoveAll(filepath.Join(*out, "lazy"))) // tracks of an earlier build (WAV or Opus) go
		must(os.MkdirAll(filepath.Join(*out, "lazy"), 0o755))
		var opusBytes int
		var opusErr error
		ents, err := os.ReadDir(filepath.Join(*data, "Data"))
		must(err)
		for _, e := range ents {
			info, err := e.Info()
			must(err)
			if e.IsDir() || info.Size() < 3<<20 {
				continue
			}
			src := filepath.Join(*data, "Data", e.Name())
			// Opus (LOOP item 5d): a track is ~6.5× smaller (2.4 MB for 3 min
			// at 96 kb/s instead of a 16 MB WAV), the browser plays it natively;
			// the game's path stays, the player tells by the "OggS" start. No
			// ffmpeg: the WAV, gzipped, as before.
			if b, err := toOpus(src); err == nil {
				url := "lazy/" + e.Name() + ".opus"
				must(os.WriteFile(filepath.Join(*out, url), b, 0o644))
				man.Lazy = append(man.Lazy, lazy{"/w/locomotion/Data/" + e.Name(), url, len(b)})
				opusBytes += len(b)
				continue
			} else if opusErr == nil {
				opusErr = err
			}
			b, err := os.ReadFile(src)
			must(err)
			url := "lazy/" + e.Name() + ".gz"
			must(os.WriteFile(filepath.Join(*out, url), gzipBytes(b), 0o644))
			man.Lazy = append(man.Lazy, lazy{"/w/locomotion/Data/" + e.Name(), url, len(b)})
		}
		if opusErr != nil {
			log.Printf("music: kept as WAV (%v)", opusErr)
		}
		log.Printf("music: %d tracks, fetched on first use (%.1f MB as Opus)", len(man.Lazy), float64(opusBytes)/(1<<20))
	}
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
	must(os.WriteFile(filepath.Join(*out, man.Wasm), enc(gzipBytes(wb)), 0o644))
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

// toOpus converts a music track (WAV) to Ogg Opus at 96 kb/s with ffmpeg.
func toOpus(path string) ([]byte, error) {
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-i", path, "-c:a", "libopus", "-b:a", "96k", "-f", "ogg", "-")
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %v %s", err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}
