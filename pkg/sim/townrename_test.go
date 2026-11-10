package sim

import "testing"

// Renaming a town: the name is a new user string, the old one freed, the
// town's label recomputed (wider for a longer name), and its stations' labels
// show the new name; another town's name is refused (unique), the dry run
// leaves everything.
func TestRenameTown(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, ok := om.G1File.(G1Raw)
	if !ok {
		t.Skip("no g1")
	}
	if len(m.Towns) < 2 {
		t.Skip("fewer than two towns")
	}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	old := m.TownName(m.Towns[0].Name, 0)
	oldID := m.Towns[0].Name
	if _, ok := m.RenameTown(c, CmdApply, 0, m.TownName(m.Towns[1].Name, 1), g1); ok {
		t.Error("another town's name was taken")
	}
	if _, ok := m.RenameTown(c, 0, 0, "Joopsville", g1); !ok || m.TownName(m.Towns[0].Name, 0) != old {
		t.Fatal("the dry run renamed the town")
	}
	// A station named after town 0 ("{town}" or a template on it).
	st := -1
	for i, sn := range m.StationNames {
		if sn.Name != stationNameNull && sn.Town == 0 {
			st = i
			break
		}
	}
	w0 := m.Towns[0].LabelFrame
	if _, ok := m.RenameTown(c, CmdApply, 0, "Joopsville on the Long River", g1); !ok {
		t.Fatal("refused")
	}
	if got := m.TownName(m.Towns[0].Name, 0); got != "Joopsville on the Long River" {
		t.Fatalf("name %q", got)
	}
	if oldID >= kUserStringsStart && m.UserStrings[oldID-kUserStringsStart] != "" {
		t.Error("the old name's string wasn't freed")
	}
	if m.Towns[0].LabelFrame == w0 {
		t.Error("the town's label wasn't recomputed")
	}
	if st >= 0 {
		t.Logf("station %d: %q", st, m.StationName(st))
	}
}
