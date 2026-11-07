package sim

import "testing"

// renameStation on a title station: a new name is taken (the query alone
// changes nothing), the same name again does nothing, an empty name
// generates one.
func TestRenameStation(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	id := -1
	for i := range m.StationNames {
		if m.StationNames[i].Name != 0 {
			id = i
			break
		}
	}
	if id < 0 {
		t.Skip("no station")
	}
	c := &Cmd{Objects: om}
	before := m.StationName(id)
	if _, ok := m.RenameStation(c, 0, id, "Goloco Central", om); !ok || m.StationName(id) != before {
		t.Fatalf("query: ok %v name %q", ok, m.StationName(id))
	}
	if _, ok := m.RenameStation(c, CmdApply, id, "Goloco Central", om); !ok || m.StationName(id) != "Goloco Central" {
		t.Fatalf("apply: ok %v name %q", ok, m.StationName(id))
	}
	n := m.StationNames[id].Name
	if _, ok := m.RenameStation(c, CmdApply, id, "Goloco Central", om); !ok || m.StationNames[id].Name != n {
		t.Fatal("the same name changed the string")
	}
	if _, ok := m.RenameStation(c, CmdApply, id, "", om); !ok || m.StationName(id) == "Goloco Central" || m.StationName(id) == "" {
		t.Fatalf("empty: %q", m.StationName(id))
	}
	t.Logf("%q → Goloco Central → %q", before, m.StationName(id))
}
