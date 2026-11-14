package sim

import "testing"

// renameStation on a title station: a new name is taken (the query alone
// changes nothing), the same name again does nothing, an empty name
// generates one. The label's frame follows the name (updateLabel).
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
	g1, _ := om.G1File.(G1Raw)
	before := m.StationName(id)
	width := func() int { f := &m.StationNames[id].LabelFrame; return int(f.Right[0]) - int(f.Left[0]) }
	w0 := width()
	if _, ok := m.RenameStation(c, 0, id, "Goloco Central", om, g1); !ok || m.StationName(id) != before {
		t.Fatalf("query: ok %v name %q", ok, m.StationName(id))
	}
	if _, ok := m.RenameStation(c, CmdApply, id, "Goloco Central", om, g1); !ok || m.StationName(id) != "Goloco Central" {
		t.Fatalf("apply: ok %v name %q", ok, m.StationName(id))
	}
	if _, ok := m.RenameStation(c, CmdApply, id, before+" Goloco Central", om, g1); !ok || width() <= w0 {
		t.Fatalf("a longer name: frame %d wide, was %d", width(), w0)
	}
	m.RenameStation(c, CmdApply, id, "Goloco Central", om, g1)
	n := m.StationNames[id].Name
	if _, ok := m.RenameStation(c, CmdApply, id, "Goloco Central", om, g1); !ok || m.StationNames[id].Name != n {
		t.Fatal("the same name changed the string")
	}
	if _, ok := m.RenameStation(c, CmdApply, id, "", om, g1); !ok || m.StationName(id) == "Goloco Central" || m.StationName(id) == "" {
		t.Fatalf("empty: %q", m.StationName(id))
	}
	t.Logf("%q → Goloco Central → %q", before, m.StationName(id))
}
