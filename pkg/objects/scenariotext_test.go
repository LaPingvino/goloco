package objects

import (
	"path/filepath"
	"strings"
	"testing"
)

// The scenario text objects load a name and details (ScenarioTextObject::load);
// STEX044 is Weatherworld's.
func TestScenarioTextObjects(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "locomotion", "ObjData")
	files, _ := filepath.Glob(filepath.Join(dir, "STEX*.DAT"))
	if len(files) == 0 {
		t.Skip("no game data")
	}
	m := NewObjectManager(dir)
	for _, f := range files {
		lo, err := m.LoadObject(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		st, ok := lo.Object.(*ScenarioTextObject)
		if !ok || st.Name == "" || st.Details == "" {
			t.Errorf("%s: %+v", filepath.Base(f), lo.Object)
			continue
		}
		if strings.EqualFold(filepath.Base(f), "STEX044.DAT") {
			t.Logf("%q: %q", st.Name, st.Details)
			if st.Name != "Weatherworld" {
				t.Errorf("STEX044 name %q, want Weatherworld", st.Name)
			}
		}
	}
}
