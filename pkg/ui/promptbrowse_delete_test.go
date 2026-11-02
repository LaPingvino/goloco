package ui

import (
	"os"
	"path/filepath"
	"testing"
)

// A right click on a file asks; only OK deletes it, and the list refreshes.
func TestPromptBrowseDeleteAsks(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "old.goloco")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := NewPromptBrowse("Load Game", dir, []string{".goloco"}, 1024, 768)
	var asked string
	var then func()
	w.Confirm = func(caption, text, ok string, f func()) { asked, then = text, f }
	l := w.list()
	w.HandleRightClick(w.X+l[0]+20, w.Y+l[1]+1+2) // the first row
	if asked != "Delete file: “old”?" || then == nil {
		t.Fatalf("asked %q", asked)
	}
	if _, err := os.Stat(f); err != nil {
		t.Fatal("deleted before OK")
	}
	then()
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Error("not deleted after OK")
	}
	if len(w.entries) != 0 {
		t.Errorf("list not refreshed: %d entries", len(w.entries))
	}
}
