package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// findSignalledStraight: a one-tile track piece (straight, id 0) with a
// signal element after it, on the title map.
func findSignalledStraight(t *testing.T, m *Map) (x, y int, tr scenario.TrackElement) {
	t.Helper()
	for y := 4; y < m.Height-4; y++ {
		for x := 4; x < m.Width-4; x++ {
			tl := m.TileAt(x, y)
			for pos, r := range tl.Order {
				if r.Kind != scenario.KindTrack {
					continue
				}
				el := tl.Tracks[r.Index]
				if el.TrackID == 0 && el.HasSignal && !el.Ghost && signalAfter(tl, pos) >= 0 {
					return x, y, el
				}
			}
		}
	}
	t.Skip("no signalled straight on the title map")
	return 0, 0, scenario.TrackElement{}
}

func TestRemoveSignal(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	x, y, el := findSignalledStraight(t, m)
	tl := m.TileAt(x, y)
	c := &Cmd{Company: CompanyNeutral, Objects: om} // neutral: compatible with any owner
	a := SignalRemoval{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation, TrackID: 0, Index: el.SequenceIndex, TrackObjectID: el.TrackObjectID}
	pos := -1
	for p, r := range tl.Order {
		if r.Kind == scenario.KindTrack && tl.Tracks[r.Index].TrackID == 0 && tl.Tracks[r.Index].HasSignal {
			pos = p
			break
		}
	}
	sig := tl.Signals[signalAfter(tl, pos)]
	nSignals := len(tl.Signals)
	a.Left, a.Right = sig.Sides[0].HasSignal(), sig.Sides[1].HasSignal()
	if _, ok := m.RemoveSignal(c, 0, a); !ok || len(tl.Signals) != nSignals {
		t.Fatalf("check only: ok %v, signals %d → %d", ok, nSignals, len(tl.Signals))
	}
	if _, ok := m.RemoveSignal(c, CmdApply, a); !ok {
		t.Fatal("removing the signal was refused")
	}
	if len(tl.Signals) != nSignals-1 {
		t.Fatalf("signal elements: %d → %d", nSignals, len(tl.Signals))
	}
	for _, r := range tl.Order {
		if r.Kind == scenario.KindTrack && tl.Tracks[r.Index].TrackID == 0 && tl.Tracks[r.Index].BaseZ == el.BaseZ && tl.Tracks[r.Index].Rotation == el.Rotation && tl.Tracks[r.Index].HasSignal {
			t.Error("the track still says it has a signal")
		}
	}
}

// RemoveTrack on a signalled piece takes the signal off first instead of refusing.
func TestRemoveTrackWithSignal(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	x, y, el := findSignalledStraight(t, m)
	tl := m.TileAt(x, y)
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	nSignals, nTracks := len(tl.Signals), len(tl.Tracks)
	a := TrackRemoval{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation, TrackID: 0, Index: el.SequenceIndex, TrackObjectID: el.TrackObjectID}
	if _, ok := m.RemoveTrack(c, CmdApply, a, nil); !ok {
		t.Fatal("removing a signalled straight was refused")
	}
	if len(tl.Signals) != nSignals-1 || len(tl.Tracks) != nTracks-1 {
		t.Fatalf("signals %d → %d, tracks %d → %d", nSignals, len(tl.Signals), nTracks, len(tl.Tracks))
	}
}
