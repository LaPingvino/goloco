package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

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

// Station::updateCargo on the title save: every used cargo's age moves up by
// one (capped), its rating at most 2 toward calculateCargoRating, quantities
// only fall; and calculateCargoRating's own rules on a player station.
func TestUpdateStationCargo(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	before := append([]scenario.StationName(nil), m.StationNames...)
	checked := 0
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull {
			continue
		}
		targets := map[int]int{}
		for i := range st.Cargo {
			if st.Cargo[i].Origin != stationCargoOriginNull {
				c := st.Cargo[i]
				c.Age = uint8(min(int(c.Age)+1, 255))
				targets[i] = m.calculateCargoRating(st, &c)
			}
		}
		m.updateStationCargo(id, om)
		for i, target := range targets {
			b, a := before[id].Cargo[i], st.Cargo[i]
			want := int(b.Rating) + min(max(target-int(b.Rating), -2), 2)
			if int(a.Rating) != want || a.Age != uint8(min(int(b.Age)+1, 255)) || a.Quantity > b.Quantity {
				t.Errorf("station %d cargo %d: rating %d→%d (target %d, want %d), age %d→%d, quantity %d→%d", id, i, b.Rating, a.Rating, target, want, b.Age, a.Age, b.Quantity, a.Quantity)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no cargo at any station")
	}
	t.Logf("%d station cargoes stepped", checked)
	st := &scenario.StationName{Flags: stationFlag7, Owner: 0}
	for _, c := range []struct {
		cargo scenario.StationCargo
		want  int
	}{
		{scenario.StationCargo{Age: 3, Quantity: 50, VehicleSpeed: 95, VehicleAge: 0}, 200}, // 165 + 15 + 33 = 213, clamped
		{scenario.StationCargo{Age: 40, Quantity: 2000, VehicleSpeed: 20, VehicleAge: 9}, 0},
		{scenario.StationCargo{Age: 20, Quantity: 250, VehicleSpeed: 400, VehicleAge: 3}, 108}, // 85 − 40 + (250 − 35) / 4 + 10
	} {
		if !m.IsPlayerCompany(0) {
			break // the title's company 0 is an AI: the 120 rule applies, tested above through the save
		}
		if got := m.calculateCargoRating(st, &c.cargo); got != c.want {
			t.Errorf("rating of %+v: %d, want %d", c.cargo, got, c.want)
		}
	}
}
