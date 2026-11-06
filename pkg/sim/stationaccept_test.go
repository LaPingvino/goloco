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

// StationManager::update over a whole 1024-tick cycle on the title save leaves
// every station's acceptance as CalcAcceptedCargo gives it (each in-use
// station is visited once), and a station whose flag was cleared gets it back.
func TestStationsUpdateCycle(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	var target = -1
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull {
			continue
		}
		for c := range st.Cargo {
			if st.Cargo[c].Flags&cargoStatAccepted != 0 && target < 0 {
				target = id
				st.Cargo[c].Flags &^= cargoStatAccepted // forget one: the update must restore it
			}
		}
	}
	if target < 0 {
		t.Skip("no station accepts anything")
	}
	changed := 0
	for i := 0; i < 1024; i++ {
		m.ScenarioTicks++
		if b, a := m.stationsUpdate(om); b != a {
			changed++
		}
	}
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull {
			continue
		}
		want, _ := m.CalcAcceptedCargo(id, om)
		var got uint32
		for c := range st.Cargo {
			if st.Cargo[c].Flags&cargoStatAccepted != 0 {
				got |= 1 << c
			}
		}
		if got != want {
			t.Errorf("station %d %q: accepted %#x, CalcAcceptedCargo %#x", id, m.StationName(id), got, want)
		}
	}
	if changed < 1 {
		t.Errorf("station %d's cleared flag wasn't restored", target)
	}
	t.Logf("%d station(s) changed acceptance in one cycle", changed)
}
