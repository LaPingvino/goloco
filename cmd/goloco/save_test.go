package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// The saves live in the config folder's save/ (autosaves in save/autosave/),
// as OpenLoco's; autosaveClean keeps the newest autosave_amount autosaves
// (by name: the date) and leaves other files alone.
func TestAutosaveFolders(t *testing.T) {
	cfg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfg)
	t.Setenv("HOME", t.TempDir())
	if got, want := savesDir(), filepath.Join(cfg, "goloco", "save"); got != want {
		t.Fatalf("savesDir %s, want %s", got, want)
	}
	dir := autosaveDir()
	os.MkdirAll(dir, 0o755)
	names := []string{"autosave_2026-01-01_10-00-00.SV5", "autosave_2026-01-02_10-00-00.SV5", "autosave_2026-01-03_10-00-00.SV5", "mine.SV5"}
	for _, n := range names {
		os.WriteFile(filepath.Join(dir, n), nil, 0o644)
	}
	os.MkdirAll(filepath.Join(cfg, "goloco"), 0o755)
	os.WriteFile(configPath(), []byte("autosave_amount: 2\n"), 0o644)
	(&Game{}).autosaveClean()
	ents, _ := os.ReadDir(dir)
	var left []string
	for _, e := range ents {
		left = append(left, e.Name())
	}
	sort.Strings(left)
	if want := names[1:]; !reflect.DeepEqual(left, want) {
		t.Errorf("left %v, want %v", left, want)
	}
}
