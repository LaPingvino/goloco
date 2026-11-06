package sim

import "testing"

// Every (not AI-reserved) train-station element's sequence index in the title save (0, or 1 for
// an uncovered middle piece) is what sub_48D794 sets again.
func TestStationCanopiesReproduceSave(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	type key struct {
		tile *Tile
		idx  int
	}
	saved := map[key]uint8{}
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			tl := m.TileAt(x, y)
			for i := range tl.Stations {
				if tl.Stations[i].StationType == 0 && !tl.Stations[i].AiAllocated { // AI reservations: no station of them yet
					saved[key{tl, i}] = tl.Stations[i].SequenceIndex
					tl.Stations[i].SequenceIndex = 9 // must be rewritten
				}
			}
		}
	}
	for id := range m.StationNames {
		if st := m.StationNames[id]; st.Name != stationNameNull && st.Flags&stationFlag5 == 0 {
			m.UpdateStationCanopies(id, om)
		}
	}
	same, ones, bad := 0, 0, 0
	for k, v := range saved {
		got := k.tile.Stations[k.idx].SequenceIndex
		switch {
		case got == v:
			same++
			if v == 1 {
				ones++
			}
		default:
			bad++
			if bad <= 5 {
				e := k.tile.Stations[k.idx]
				t.Errorf("station element %d of station %d: index %d, saved %d (ai %v ghost %v, station flags %#x tiles %d)", k.idx, e.StationID, got, v, e.AiAllocated, e.Ghost, m.StationNames[e.StationID].Flags, m.StationNames[e.StationID].TileSize)
			}
		}
	}
	t.Logf("%d of %d train-station elements match (%d middle pieces)", same, len(saved), ones)
}
