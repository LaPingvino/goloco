package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Signal auto placement (04e7900) on the title map: a run of track between
// junctions (its signals taken off first), a signal every 2nd piece both ways
// from a straight in it (the straight itself included), then the same pattern
// removed again.
func TestSignalsAuto(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	count := func() int {
		n := 0
		for y := range m.Tiles {
			for x := range m.Tiles[y] {
				n += len(m.Tiles[y][x].Signals)
			}
		}
		return n
	}
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	var signalled []SignalRemoval
	pieces := func(x, y, z int32, el scenario.TrackElement) (n int, clean bool) {
		signalled = nil
		clean = true
		walk := func(sx, sy, sz int32, tad uint16) {
			m.iterateTrackToJunction(sx, sy, sz, tad, el.TrackObjectID, c.Company, func(px, py, pz int32, tad uint16) bool {
				n++
				rot, id := uint8(tad&3), uint8(tad>>3)&0x3F
				p0 := TrackPieces[id][0]
				ox, oy := rotOffset(p0.X, p0.Y, rot)
				tl, ti, _ := m.signalTrackAt(SignalRemoval{Rotation: rot, TrackID: id, TrackObjectID: el.TrackObjectID}, px+int32(ox), py+int32(oy), pz+int32(p0.Z), 0)
				if tl == nil {
					clean = false
				} else if tl.Tracks[ti].HasStation { // the walk ends there (validateTrackIsSignalCompatible)
					n--
					return false
				} else if tl.Tracks[ti].HasSignal {
					signalled = append(signalled, SignalRemoval{X: px + int32(ox), Y: py + int32(oy), Z: pz + int32(p0.Z), Rotation: rot, TrackID: id,
						TrackObjectID: el.TrackObjectID, Left: true, Right: true})
				}
				return n < 200
			})
		}
		walk(x, y, z, uint16(el.Rotation)) // the straight (id 0) forward
		ts := TrackCoordinates[el.Rotation] // and back from its start, as autoSignalsWalk does
		rx, ry := x, y
		if ts.RotBegin < 12 {
			rx, ry = rx-kRotationOffset[ts.RotBegin][0], ry-kRotationOffset[ts.RotBegin][1]
		}
		if tc := m.TrackConnections(rx, ry, z, ReverseRotation[ts.RotBegin], c.Company, el.TrackObjectID, 0, 0); len(tc.Connections) == 1 {
			walk(rx, ry, z, tc.Connections[0]&TaDBasicMask)
		}
		return n, clean
	}
	for ty := 4; ty < m.Height-4; ty++ {
		for tx := 4; tx < m.Width-4; tx++ {
			tl := m.TileAt(tx, ty)
			for _, r := range tl.Order {
				if r.Kind != scenario.KindTrack {
					continue
				}
				el := tl.Tracks[r.Index]
				if el.TrackID != 0 || el.HasSignal || el.Ghost || el.HasStation {
					continue
				}
				x, y, z := int32(tx)*32, int32(ty)*32, int32(el.BaseZ)*4
				c.Company = el.Owner // the walk follows the company's own track (getTrackConnections)
				n, clean := pieces(x, y, z, el)
				if !clean || n < 8 || n > 60 {
					continue
				}
				for _, rm := range signalled { // a clean run first
					if _, ok := m.RemoveSignal(c, CmdApply, rm); !ok {
						t.Fatalf("clearing the run's signal at %d,%d refused", rm.X/32, rm.Y/32)
					}
				}
				before := count()
				a := SignalsPlacementAuto{SignalPlacement{X: x, Y: y, Z: z, Rotation: el.Rotation, TrackID: 0, TrackObjectID: el.TrackObjectID, Left: true}, 2}
				if _, ok := m.CreateSignalsAuto(c, CmdApply, a); !ok {
					t.Fatalf("placement at %d,%d refused (error %d)", tx, ty, c.ErrorText)
				}
				placed := count() - before
				if placed < 3 || !tl.Tracks[r.Index].HasSignal {
					t.Fatalf("at %d,%d: %d signals placed, the clicked straight signalled %v", tx, ty, placed, tl.Tracks[r.Index].HasSignal)
				}
				rm := SignalsRemovalAuto{SignalRemoval{X: x, Y: y, Z: z, Rotation: el.Rotation, TrackID: 0, TrackObjectID: el.TrackObjectID, Left: true}, 2}
				if _, ok := m.RemoveSignalsAuto(c, CmdApply, rm); !ok || count() != before {
					t.Fatalf("removal: ok %v, signals %d (were %d)", ok, count(), before)
				}
				t.Logf("at %d,%d: %d signals placed and removed", tx, ty, placed)
				return
			}
		}
	}
	t.Skip("no clean run of track on the title map")
}
