package audio

import (
	"os"
	"path/filepath"
	"testing"
)

// OpenLoco's names for the music files against an installation's: found
// whatever the case, the exact name first.
func TestFindFile(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"20s1.dat", "Chrysanthemum.dat", "CSS5.DAT", "css5.dat"} {
		if err := os.WriteFile(filepath.Join(dir, n), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for ask, want := range map[string]string{"20s1.DAT": "20s1.dat", "Chrysanthemum.DAT": "Chrysanthemum.dat", "CSS5.DAT": "CSS5.DAT"} {
		if got, err := FindFile(filepath.Join(dir, ask)); err != nil || got != filepath.Join(dir, want) {
			t.Errorf("%s: %q, %v; want %s", ask, got, err, want)
		}
	}
	if _, err := FindFile(filepath.Join(dir, "Rag2.DAT")); !os.IsNotExist(err) {
		t.Errorf("a missing file: %v", err)
	}
}
