package objects

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The open build's objects (item O): every OpenGraphics object (MIT, the
// unzipped objects.zip in GOLOCO_OPENGRAPHICS; skipped without it) parses
// with goloco's loader, without g1.DAT (no images registered). Reports the
// failures by object type, so the open build knows what it can't load yet.
func TestOpenGraphicsObjectsParse(t *testing.T) {
	dir := os.Getenv("GOLOCO_OPENGRAPHICS")
	if dir == "" {
		t.Skip("GOLOCO_OPENGRAPHICS (the unzipped OpenGraphics objects.zip) not set")
	}
	m := NewObjectManager(dir)
	ok, failed := 0, map[string][]string{}
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasPrefix(d.Name(), "OG_") {
			return err
		}
		kind := filepath.Base(filepath.Dir(p))
		if _, err := m.LoadObject(p); err != nil {
			failed[kind] = append(failed[kind], d.Name()+": "+err.Error())
		} else {
			ok++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var kinds []string
	for k := range failed {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	for _, k := range kinds {
		for _, f := range failed[k] {
			t.Errorf("%s: %s", k, f)
		}
	}
	t.Logf("%d OpenGraphics objects parsed; %d types with failures", ok, len(kinds))
}
