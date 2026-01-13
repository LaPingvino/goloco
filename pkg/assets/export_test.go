package assets

import (
	"os"
	"path/filepath"
	"testing"
)

// This test is a functional extractor test that runs only when the locomotion
// data is available at the expected path via the repository symlink. It
// writes extracted PNGs to ../../assets/extracted/<datfile>/ and will be
// skipped otherwise.
func TestExportG1Images_RealDat(t *testing.T) {
	// expected path: repository root has a 'locomotion' symlink
	// Try several known locations for locomotion Data/title.dat
	candidates := []string{
		filepath.Join("..", "..", "..", "locomotion", "Data", "title.dat"),
		filepath.Join(os.Getenv("HOME"), ".local", "share", "Steam", "steamapps", "common", "Locomotion", "Data", "title.dat"),
	}
	var datPath string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			datPath = c
			break
		}
	}
	if datPath == "" {
		t.Skip("locomotion data not found in known locations; skipping extractor test")
		return
	}
	data, err := LoadDatFile(datPath)
	if err != nil {
		t.Fatalf("failed to load dat: %v", err)
	}
	out := filepath.Join("..", "..", "assets", "extracted")
	count, err := ExportG1Images(data, filepath.Base(datPath), out)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	t.Logf("exported %d images to %s", count, out)
}
