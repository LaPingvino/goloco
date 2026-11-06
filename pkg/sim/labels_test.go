package sim

import "testing"

// Recomputed for the title save, every town's and every live station's label
// frame matches the one the game saved.
func TestLabelFramesReproduceSave(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, ok := om.G1File.(G1Raw)
	if !ok {
		t.Skip("no g1")
	}
	towns, stations := 0, 0
	for i := range m.Towns {
		saved := m.Towns[i].LabelFrame
		m.UpdateTownLabel(i, g1)
		if got := m.Towns[i].LabelFrame; got != saved {
			t.Errorf("town %d %q: %+v, saved %+v", i, m.TownName(m.Towns[i].Name, i), got, saved)
		} else {
			towns++
		}
	}
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == 0xFFFF || st.Flags&(1<<5) != 0 { // Station::empty(): name == StringIds::null
			continue
		}
		saved := st.LabelFrame
		m.UpdateStationLabel(id, g1)
		if st.LabelFrame != saved {
			if stations+1 < 5 || testing.Verbose() {
				t.Errorf("station %d %q: %+v, saved %+v", id, m.StationName(id), st.LabelFrame, saved)
			}
		} else {
			stations++
		}
	}
	t.Logf("%d towns and %d stations reproduced", towns, stations)
}
