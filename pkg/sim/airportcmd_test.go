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

// An airport built by the player on the title save: the dry run changes
// nothing and costs what applying does; applied, every building tile has an
// airport element of the new station, which gets flag_6 and the start /
// rotation, and the nearest town counts one airport more.
func TestCreateAirport(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, _ := om.G1File.(G1Raw)
	typ, best := -1, 1<<30
	for i, a := range om.AirportObjects { // the smallest airport, likeliest to fit
		if a != nil && len(a.Buildings) > 0 && int(a.MaxX-a.MinX)*int(a.MaxY-a.MinY) < best {
			typ, best = i, int(a.MaxX-a.MinX)*int(a.MaxY-a.MinY)
		}
	}
	if typ < 0 {
		t.Skip("no airport object")
	}
	c := &Cmd{Company: m.PlayerCompanies[0], Objects: om}
	for y := 8; y < m.Height-8; y += 2 {
		for x := 8; x < m.Width-8; x += 2 {
			tl := m.TileAt(x, y)
			if tl.WaterLevel != 0 {
				continue
			}
			a := AirportPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(tl.BaseZ) * 4, Type: uint8(typ)}
			town, _, _ := m.ClosestTownAndDensity(a.X, a.Y)
			n0 := m.Towns[town].NumberOfAirports
			dry, ok := m.CreateAirport(c, 0, a, g1)
			if !ok {
				continue
			}
			if m.Towns[town].NumberOfAirports != n0 {
				t.Fatal("the dry run counted an airport")
			}
			cost, ok := m.CreateAirport(c, CmdApply, a, g1)
			if !ok || cost != dry || cost <= 0 {
				t.Fatalf("applying at %d,%d: ok %v cost %d (dry %d)", x, y, ok, cost, dry)
			}
			id := -1
			for _, st := range m.TileAt(x, y).Stations {
				if st.StationType == 2 && !st.Ghost {
					id = int(st.StationID)
				}
			}
			if id < 0 {
				t.Fatalf("no airport element on the centre tile %d,%d", x, y)
			}
			st := &m.StationNames[id]
			if st.Flags&stationFlagHasAirport == 0 || st.AirportStartPos[0] != int16(a.X) || st.AirportStartPos[1] != int16(a.Y) || st.Owner != c.Company {
				t.Errorf("station %d %q: flags %#x start %v owner %d", id, m.StationName(id), st.Flags, st.AirportStartPos, st.Owner)
			}
			if m.Towns[town].NumberOfAirports != n0+1 {
				t.Errorf("town airports %d, want %d", m.Towns[town].NumberOfAirports, n0+1)
			}
			t.Logf("airport %q (object %d) at %d,%d: %q, cost %d", om.AirportObjects[typ].Name, typ, x, y, m.StationName(id), cost)
			// removed again (its centre clicked):
			// every element goes, flag_6 is cleared, the town counts one less
			refund, ok := m.RemoveAirport(c, CmdApply, AirportRemoval{X: a.X, Y: a.Y, Z: a.Z}, g1)
			if !ok || refund >= 0 {
				t.Fatalf("removing: ok %v refund %d", ok, refund)
			}
			for yy := y - 4; yy <= y+4; yy++ {
				for xx := x - 4; xx <= x+4; xx++ {
					for _, se := range m.TileAt(xx, yy).Stations {
						if se.StationType == 2 && int(se.StationID) == id && !se.Ghost {
							t.Errorf("an airport element left on %d,%d", xx, yy)
						}
					}
				}
			}
			if st.Flags&stationFlagHasAirport != 0 || st.TileSize != 0 {
				t.Errorf("station %d after removal: flags %#x tiles %d", id, st.Flags, st.TileSize)
			}
			if m.Towns[town].NumberOfAirports != n0 {
				t.Errorf("town airports %d after removal, want %d", m.Towns[town].NumberOfAirports, n0)
			}
			t.Logf("removed: refund %d", refund)
			return
		}
	}
	t.Fatal("no spot took an airport")
}
