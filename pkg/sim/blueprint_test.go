package sim

import (
	"fmt"
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// copyTrackToBlueprint / placeBlueprint / removeBlueprint (Blueprints.cpp)
// on the title map: a company's signalled straight on flat ground is copied,
// taken up, pasted back as a ghost (ghost track and signal where it stood),
// the ghost removed (the tile as before the ghost), then pasted for real (the
// straight and its signal back, not ghosts). The blueprint itself is
// unchanged by the placements; rotateBlueprint four times is the identity.
func TestBlueprintCopyPaste(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if len(om.TrainSignalObjects) == 0 || om.TrainSignalObjects[0] == nil {
		t.Skip("no signal object")
	}
	s := &State{Map: *m}
	var x, y int
	var el scenario.TrackElement
	found := false
	for ty := 4; ty < s.Height-4 && !found; ty++ {
		for tx := 4; tx < s.Width-4 && !found; tx++ {
			tl := s.TileAt(tx, ty)
			if len(tl.Tracks) != 1 || len(tl.Roads) != 0 || tl.Slope != 0 || tl.WaterLevel != 0 {
				continue
			}
			e := tl.Tracks[0]
			if e.TrackID == 0 && !e.HasSignal && !e.HasStation && !e.HasLevelCrossing && !e.HasBridge && !e.Ghost && !e.AiAllocated && e.BaseZ == tl.BaseZ {
				x, y, el, found = tx, ty, e, true
			}
		}
	}
	if !found {
		t.Skip("no plain flat straight")
	}
	s.ControllingCompany = el.Owner
	c := &Cmd{Company: el.Owner, Objects: om}
	sig := SignalPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation, TrackID: 0, Type: 0, TrackObjectID: el.TrackObjectID, Left: true}
	if _, ok := s.CreateSignal(c, CmdApply, sig); !ok {
		t.Fatalf("signal refused (error %d)", c.ErrorText)
	}

	bp := s.CopyTrackToBlueprint([2]int{x, y}, [2]int{x, y})
	if len(bp.Tracks) != 1 || len(bp.Signals) != 1 || len(bp.Stations) != 0 {
		t.Fatalf("blueprint %+v", bp)
	}
	if tr := bp.Tracks[0]; tr.X != 0 || tr.Y != 0 || tr.Z != 0 || tr.Bridge != 0xFF || tr.Rotation != el.Rotation {
		t.Errorf("track placement %+v (want at the pivot, no bridge)", tr)
	}
	if sg := bp.Signals[0]; !sg.Left || sg.Right || sg.Type != 0 {
		t.Errorf("signal placement %+v (want the left side, type 0)", sg)
	}
	saved := fmt.Sprintf("%+v", bp)
	turned := bp
	turned.Tracks, turned.Signals = append([]TrackPlacement(nil), bp.Tracks...), append([]SignalPlacement(nil), bp.Signals...)
	for range 4 {
		turned.Rotate(1)
	}
	if fmt.Sprintf("%+v", turned) != saved {
		t.Errorf("four quarter turns: %+v, want %s", turned, saved)
	}

	pos := [3]int32{int32(x) * 32, int32(y) * 32, int32(el.BaseZ) * 4}
	if p, ok := s.ConstructionPos(x, y); !ok || p != pos {
		t.Fatalf("ConstructionPos %v, want %v", p, pos)
	}
	if _, ok := s.RemoveTrack(c, CmdApply, TrackRemoval{X: pos[0], Y: pos[1], Z: pos[2], Rotation: el.Rotation, TrackID: 0, TrackObjectID: el.TrackObjectID}, nil); !ok {
		t.Fatalf("taking the straight up refused (error %d)", c.ErrorText)
	}
	tl := s.TileAt(x, y)
	before := fmt.Sprintf("%+v", *tl)

	s.SetBlueprint(&bp)
	s.PasteToolUpdate(c, pos)
	if !s.BlueprintGhostShown() || len(tl.Tracks) != 1 || !tl.Tracks[0].Ghost || len(tl.Signals) != 1 || !tl.Signals[0].LeftGhost {
		t.Fatalf("the ghost: shown %v, tracks %+v, signals %+v", s.BlueprintGhostShown(), tl.Tracks, tl.Signals)
	}
	s.PasteToolUpdate(c, pos) // the same place: the ghost stays
	s.RemoveConstructionGhosts(c)
	if s.BlueprintGhostShown() || fmt.Sprintf("%+v", *tl) != before {
		t.Fatalf("after the ghost went:\n%+v\nwant\n%s", *tl, before)
	}

	_, res, ok := s.PasteToolUp(c, pos)
	if !ok || !res.Built || res.Failed != nil {
		t.Fatalf("paste: ok %v %+v (error %d)", ok, res, c.ErrorText)
	}
	if len(tl.Tracks) != 1 || tl.Tracks[0].Ghost || !tl.Tracks[0].HasSignal || len(tl.Signals) != 1 || tl.Signals[0].LeftGhost || !tl.Signals[0].Sides[0].HasSignal() {
		t.Fatalf("pasted: tracks %+v, signals %+v", tl.Tracks, tl.Signals)
	}
	if fmt.Sprintf("%+v", bp) != saved {
		t.Errorf("the blueprint changed: %+v, want %s", bp, saved)
	}
	// On the track it was copied from the trial refuses the whole paste.
	if _, _, ok := s.PasteToolUp(c, pos); ok {
		t.Error("a paste onto the same track wasn't refused")
	}
}
