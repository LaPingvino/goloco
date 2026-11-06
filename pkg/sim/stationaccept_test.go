package sim

import "testing"

// For every live station of the title save, the cargo it accepts (its cargo
// stats' isAccepted bit, as Station::update last set it) is what
// CalcAcceptedCargo finds in its catchment, and the industry per cargo too.
func TestCalcAcceptedCargoReproducesSave(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	same, diff := 0, 0
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull || st.Flags&stationFlag5 != 0 {
			continue
		}
		got, s := m.CalcAcceptedCargo(id, om)
		var saved uint32
		indOK := true
		for c := 0; c < 32; c++ {
			if st.Cargo[c].Flags&1 != 0 {
				saved |= 1 << c
			}
			if om.CargoObjects[c] != nil && st.Cargo[c].IndustryID != s.Industry[c] {
				indOK = false
			}
		}
		if got != saved || !indOK {
			diff++
			if diff <= 5 {
				t.Errorf("station %d %q: accepts %#x, saved %#x (industries match %v)", id, m.StationName(id), got, saved, indOK)
			}
		} else {
			same++
		}
	}
	t.Logf("%d stations match, %d differ", same, diff)
}
