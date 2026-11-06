package sim

import "testing"

// Recomputed for every live title station: its modes and acceptedFromProducer
// bits (recalculateStationModes), its centre and label (recalculateStationCenter)
// match the save; a tile taken out and put back leaves its acceptance as it was.
func TestStationTilesReproduceSave(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	g1, _ := om.G1File.(G1Raw)
	modes, centres, roundTrips, n := 0, 0, 0, 0
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull || st.Flags&stationFlag5 != 0 || st.TileSize == 0 {
			continue
		}
		n++
		saved := *st
		m.RecalculateStationModes(id, om)
		ok := st.Flags == saved.Flags
		for c := range st.Cargo {
			ok = ok && st.Cargo[c].Flags&stationCargoAcceptedFromProducer == saved.Cargo[c].Flags&stationCargoAcceptedFromProducer
		}
		if ok {
			modes++
		} else if n-modes <= 3 {
			t.Errorf("station %d %q: modes %#x (saved %#x)", id, m.StationName(id), st.Flags, saved.Flags)
		}
		m.RecalculateStationCenter(id, g1)
		if st.X == saved.X && st.Y == saved.Y && st.Z == saved.Z && (g1 == nil || st.LabelFrame == saved.LabelFrame) {
			centres++
		} else if n-centres <= 3 {
			t.Errorf("station %d %q: centre %d,%d,%d (saved %d,%d,%d)", id, m.StationName(id), st.X, st.Y, st.Z, saved.X, saved.Y, saved.Z)
		}
		last := st.Tiles[st.TileSize-1]
		m.RemoveTileFromStationAndRecalcCargo(id, int32(last[0]), int32(last[1]), int32(last[2]&^3), uint8(last[2]&3), om)
		m.AddTileToStation(id, int32(last[0]), int32(last[1]), int32(last[2]), uint8(last[2]&3), om)
		if st.Tiles == saved.Tiles && st.TileSize == saved.TileSize && func() bool {
			for c := range st.Cargo {
				if st.Cargo[c].Flags&cargoStatAccepted != saved.Cargo[c].Flags&cargoStatAccepted {
					return false
				}
			}
			return true
		}() {
			roundTrips++
		} else if n-roundTrips <= 3 {
			t.Errorf("station %d %q: tile round trip changed it", id, m.StationName(id))
		}
	}
	t.Logf("%d stations: modes %d, centres %d, tile round trips %d", n, modes, centres, roundTrips)
}
