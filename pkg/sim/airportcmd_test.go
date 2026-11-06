package sim

import "testing"

// From every live airport of the title save, its owner's search never
// returns that station (flag_6: it has its airport) — and the station does
// carry flag_6. The extents are ordered and hold the start tile.
func TestFindNearbyStationAirportSkipsAirports(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	n := 0
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull || st.Flags&stationFlag5 != 0 {
			continue
		}
		for i := 0; i < int(st.TileSize); i++ {
			p := st.Tiles[i]
			el := m.getStationElement(int32(p[0]), int32(p[1]), int32(p[2]&^3))
			if el == nil || el.StationType != 2 || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil {
				continue
			}
			n++
			a := om.AirportObjects[el.ObjectID]
			if st.Flags&stationFlagHasAirport == 0 {
				t.Errorf("airport station %d %q lacks flag_6 (flags %#x)", id, m.StationName(id), st.Flags)
			}
			x0, y0, x1, y1 := airportExtents(a, int(p[0])>>5, int(p[1])>>5, el.Rotation)
			if x0 > x1 || y0 > y1 || int(p[0])>>5 < x0 || int(p[0])>>5 > x1 || int(p[1])>>5 < y0 || int(p[1])>>5 > y1 {
				t.Errorf("airport %d: extents %d,%d–%d,%d miss its tile %v", id, x0, y0, x1, y1, p)
			}
			if got, _ := m.findNearbyStationAirport(int32(p[0]), int32(p[1]), int32(p[2]&^3), a, el.Rotation, st.Owner); got == id {
				t.Errorf("airport station %d %q found by the airport search", id, m.StationName(id))
			}
			break
		}
	}
	if n == 0 {
		t.Skip("no airports in the title save")
	}
	t.Logf("%d airports checked", n)
}
