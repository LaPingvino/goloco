package scene

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// Every signal block searched from both sides of every signalled piece
// (sub_4A2AD7: findOccupationByBlock + setSignalsOccupiedState), after
// clearing all occupied bits, gives back exactly the saved occupied bits.
func TestBlockSignalsReproduceSave(t *testing.T) {
	_, m, _ := titleView(t)
	type key struct{ x, y, i int }
	saved := map[key][2]scenario.SignalSide{}
	type start struct {
		x, y, z    int32
		tad        uint16
		own, ttype uint8
	}
	var starts []start
	occ := 0
	for y := range m.Height {
		for x := range m.Width {
			tl := m.TileAt(x, y)
			for i, s := range tl.Signals {
				saved[key{x, y, i}] = s.Sides
				for _, sd := range s.Sides {
					if sd.IsOccupied() {
						occ++
					}
				}
			}
			for _, el := range tl.Tracks {
				if !el.HasSignal || el.SequenceIndex != 0 || el.Ghost || el.AiAllocated {
					continue
				}
				tad := uint16(el.TrackID)<<3 | uint16(el.Rotation)
				z := int32(el.BaseZ)*4 - int32(sim.TrackPieces[el.TrackID][0].Z)
				starts = append(starts, start{int32(x * 32), int32(y * 32), z, tad, el.Owner, el.TrackObjectID})
				ex, ey, ez := sim.TrackEndTile(int32(x*32), int32(y*32), z, tad)
				starts = append(starts, start{ex, ey, ez, tad | 4, el.Owner, el.TrackObjectID})
			}
		}
	}
	for y := range m.Height { // forget the saved occupancy
		for x := range m.Width {
			for i := range m.TileAt(x, y).Signals {
				for j := range 2 {
					m.TileAt(x, y).Signals[i].Sides[j].SetIsOccupied(false)
				}
			}
		}
	}
	for _, s := range starts {
		m.UpdateBlockSignals(s.x, s.y, s.z, s.tad, s.own, s.ttype)
	}
	diff, total := 0, 0
	for k, sides := range saved {
		now := m.TileAt(k.x, k.y).Signals[k.i].Sides
		for j := range 2 {
			total++
			if now[j].IsOccupied() != sides[j].IsOccupied() {
				diff++
				if diff <= 5 {
					t.Logf("(%d,%d) side %d: saved occupied %v, recomputed %v", k.x, k.y, j, sides[j].IsOccupied(), now[j].IsOccupied())
				}
			}
		}
	}
	t.Logf("signals %d (sides %d, %d occupied saved), block searches %d: %d sides differ", len(saved), total, occ, len(starts), diff)
	if diff != 0 || occ == 0 {
		t.Errorf("%d sides differ (saved occupied %d)", diff, occ)
	}
}
