package sim

import "testing"

func TestAllocateNewStation(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	best := 0
	for i, tw := range m.Towns {
		if tw.Population > m.Towns[best].Population {
			best = i
		}
	}
	town := m.Towns[best]
	x, y := int32(town.X)+64, int32(town.Y)+64
	z := int32(m.TileAt(int(x)>>5, int(y)>>5).BaseZ) * 4
	used := map[uint16]bool{}
	for i := range m.StationNames {
		if st := m.StationNames[i]; st.Name != stationNameNull && st.Town == town.ID {
			used[st.Name] = true
		}
	}
	a, ok := m.AllocateNewStation(x, y, z, 1, 0, om)
	if !ok {
		t.Fatal("allocation refused")
	}
	st := m.StationNames[a]
	if st.Town != town.ID || st.Flags != stationFlag5 || st.Cargo[0].Rating != 150 || used[st.Name] {
		t.Fatalf("station %d: town %d (want %d), flags %#x, rating %d, name %d (used: %v)", a, st.Town, town.ID, st.Flags, st.Cargo[0].Rating, st.Name, used[st.Name])
	}
	b, _ := m.AllocateNewStation(x, y, z, 1, 0, om)
	air, _ := m.AllocateNewStation(x, y, z, 1, 1, om)
	if m.StationNames[b].Name == st.Name || b == a {
		t.Errorf("a second station there: slot %d name %q, the first %q", b, m.StationName(b), m.StationName(a))
	}
	if !used[toTownName(stringStationTown+snAirport)] && m.StationNames[air].Name != toTownName(stringStationTown+snAirport) {
		t.Errorf("mode 1 named %q, not the airport", m.StationName(air))
	}
	t.Logf("%s: new stations %q, %q, %q", m.TownName(town.Name, best), m.StationName(a), m.StationName(b), m.StationName(air))
	m.DeallocateStation(b)
	if m.StationNames[b].Name != stationNameNull {
		t.Error("deallocated slot not free")
	}
	if c, _ := m.AllocateNewStation(x, y, z, 1, 0, om); c != b || m.StationNames[c].Name != m.StationNames[b].Name {
		t.Errorf("the freed slot %d not reused (got %d)", b, c)
	}
}
