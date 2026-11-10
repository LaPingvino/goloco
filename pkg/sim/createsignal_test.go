package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A straight on the title map with no signal, station or level crossing, the
// only track on its tile.
func findPlainStraight(t *testing.T, m *Map) (int, int, scenario.TrackElement) {
	t.Helper()
	for y := 4; y < m.Height-4; y++ {
		for x := 4; x < m.Width-4; x++ {
			tl := m.TileAt(x, y)
			if len(tl.Tracks) != 1 || len(tl.Roads) != 0 {
				continue
			}
			el := tl.Tracks[0]
			if el.TrackID == 0 && !el.HasSignal && !el.HasStation && !el.HasLevelCrossing && !el.Ghost && !el.AiAllocated {
				return x, y, el
			}
		}
	}
	t.Skip("no plain straight")
	return 0, 0, scenario.TrackElement{}
}

func TestCreateSignal(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(om.TrainSignalObjects) == 0 || om.TrainSignalObjects[0] == nil {
		t.Skip("no signal object")
	}
	x, y, el := findPlainStraight(t, m)
	tl := m.TileAt(x, y)
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	a := SignalPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation, TrackID: 0, Index: el.SequenceIndex, Type: 0, TrackObjectID: el.TrackObjectID, Left: true}

	dry, ok := m.CreateSignal(c, 0, a)
	if !ok || dry <= 0 || len(tl.Signals) != 0 {
		t.Fatalf("dry run: ok %v cost %d signals %d", ok, dry, len(tl.Signals))
	}
	cost, ok := m.CreateSignal(c, CmdApply, a)
	if !ok || cost != dry {
		t.Fatalf("apply: ok %v cost %d (dry %d)", ok, cost, dry)
	}
	if !tl.Tracks[0].HasSignal || len(tl.Signals) != 1 {
		t.Fatalf("track HasSignal %v, signals %d", tl.Tracks[0].HasSignal, len(tl.Signals))
	}
	pos := -1
	for p, r := range tl.Order {
		if r.Kind == scenario.KindTrack {
			pos = p
		}
	}
	si := signalAfter(tl, pos)
	if si < 0 {
		t.Fatal("the signal element isn't right after the track")
	}
	sg := tl.Signals[si]
	if !sg.Sides[0].HasSignal() || sg.Sides[1].HasSignal() || sg.Rotation != el.Rotation || sg.BaseZ != el.BaseZ || !sg.HasLeftSignal {
		t.Errorf("signal %+v", sg)
	}
	if tl.Order[pos+1].Flags&0xF != tl.Order[pos].Flags&0xF {
		t.Error("the signal's ref doesn't carry the track's quarters")
	}
	// The same signal on the same side again costs nothing.
	if again, ok := m.CreateSignal(c, 0, a); !ok || again != 0 {
		t.Errorf("the same signal again: ok %v cost %d", ok, again)
	}
	// Removed (a negative cost: the sell factor refunds), the tile is back as it was.
	refund, ok := m.RemoveSignal(c, CmdApply, SignalRemoval{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, TrackID: 0, Index: a.Index, TrackObjectID: a.TrackObjectID, Left: true})
	if !ok || refund >= 0 || len(tl.Signals) != 0 || tl.Tracks[0].HasSignal {
		t.Errorf("removal: ok %v refund %d signals %d HasSignal %v", ok, refund, len(tl.Signals), tl.Tracks[0].HasSignal)
	}
}

// A ghost signal is a ghost element; a level crossing refuses a signal.
func TestCreateSignalGhostAndCrossing(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(om.TrainSignalObjects) == 0 || om.TrainSignalObjects[0] == nil {
		t.Skip("no signal object")
	}
	x, y, el := findPlainStraight(t, m)
	tl := m.TileAt(x, y)
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	a := SignalPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation, Index: el.SequenceIndex, TrackObjectID: el.TrackObjectID, Right: true}
	if _, ok := m.CreateSignal(c, CmdApply|CmdGhost, a); !ok || len(tl.Signals) != 1 {
		t.Fatalf("ghost: ok %v signals %d", ok, len(tl.Signals))
	}
	if sg := tl.Signals[0]; !sg.Ghost || !sg.RightGhost || !sg.Sides[1].HasSignal() {
		t.Errorf("ghost signal %+v", sg)
	}
	if _, ok := m.RemoveSignal(c, CmdApply|CmdGhost, SignalRemoval{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, Index: a.Index, TrackObjectID: a.TrackObjectID, Right: true}); !ok || len(tl.Signals) != 0 {
		t.Fatalf("ghost removal: ok %v signals %d", ok, len(tl.Signals))
	}
	tl.Tracks[0].HasLevelCrossing = true
	if _, ok := m.CreateSignal(c, 0, a); ok || c.ErrorText != i18n.LevelCrossingInTheWay {
		t.Errorf("on a crossing: ok %v error %d", ok, c.ErrorText)
	}
}
