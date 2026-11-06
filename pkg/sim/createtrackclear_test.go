package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// clearTrack: crossing pieces that share no connection pass; the same piece
// again is "already built"; a junction needs a junction-capable track object
// and level pieces at one height.
func TestTrackClearTrack(t *testing.T) {
	h := &objects.ObjectHeader{}
	copy(h.Name[:], "TRACKST ")
	withJunction := &objects.TrackObject{Header: h, TrackPieces: trackTraitJunction}
	c := &Cmd{Company: 0, Objects: &objects.ObjectManager{TrackObjects: []*objects.TrackObject{withJunction}}}
	m := &Map{}
	straight := scenario.TrackElement{TrackID: 0, Rotation: 0, SequenceIndex: 0, BaseZ: 8, Owner: 0}
	args := trackClearArgs{Z: 32, TrackID: 0, Rotation: 0, Index: 0, IsLastIndex: true, BridgeID: 0xFF}
	st := &trackCmdState{Bridge: 0xFF}
	if r := m.trackClearTrack(c, st, &straight, args); r != ClearCollisionErrorSet {
		t.Errorf("the same straight again: %v", r)
	}
	cross := args
	cross.Rotation = 1 // a straight at right angles: shares the tile's connections
	if r := m.trackClearTrack(c, st, &straight, cross); r != ClearNoCollision || st.Flags73&(1<<3) == 0 {
		t.Errorf("a junction with a junction-capable track: %v flags %b", r, st.Flags73)
	}
	low := cross
	low.Z = 48
	if r := m.trackClearTrack(c, &trackCmdState{}, &straight, low); r != ClearCollisionErrorSet {
		t.Errorf("a junction at another height: %v", r)
	}
	withJunction.TrackPieces = 0
	if r := m.trackClearTrack(c, &trackCmdState{}, &straight, cross); r != ClearCollisionErrorSet {
		t.Errorf("a junction without the trait: %v", r)
	}
	signal := straight
	signal.HasSignal = true
	withJunction.TrackPieces = trackTraitJunction
	if r := m.trackClearTrack(c, &trackCmdState{}, &signal, cross); r != ClearCollisionErrorSet {
		t.Errorf("a junction through a signal: %v", r)
	}
}

// clearRoad: a level crossing only for straight track over straight road at
// right angles and the same height, of compatible objects; marked on apply.
func TestTrackClearRoad(t *testing.T) {
	hdr := func(name string) *objects.ObjectHeader { h := &objects.ObjectHeader{}; copy(h.Name[:], name); return h }
	track := &objects.TrackObject{Header: hdr("TRACKST "), CompatibleNames: []string{"ROADTOWN"}}
	road := &objects.RoadObject{Header: hdr("ROADTOWN")}
	c := &Cmd{Company: 0, Objects: &objects.ObjectManager{TrackObjects: []*objects.TrackObject{track}, RoadObjects: []*objects.RoadObject{road}}}
	m := &Map{DefaultLevelCrossingType: 1}
	rd := scenario.RoadElement{RoadID: 0, Rotation: 1, BaseZ: 8}
	a := trackClearArgs{Z: 32, TrackID: 0, Rotation: 0, BridgeID: 0xFF, Flags: CmdApply}
	var lc bool
	st := &trackCmdState{}
	if r := m.trackClearRoad(c, st, &rd, a, &lc); r != ClearNoCollision || !lc || !rd.HasLevelCrossing || rd.LevelCrossingObjectID != 1 {
		t.Errorf("a level crossing: %v %v %+v", r, lc, rd)
	}
	along := scenario.RoadElement{RoadID: 0, Rotation: 0, BaseZ: 8}
	if r := m.trackClearRoad(c, &trackCmdState{}, &along, a, &lc); r != ClearCollisionErrorSet {
		t.Errorf("track along the road: %v", r)
	}
	ghost := scenario.RoadElement{RoadID: 0, Rotation: 1, BaseZ: 8}
	a.Flags = CmdApply | CmdGhost
	m.trackClearRoad(c, &trackCmdState{}, &ghost, a, &lc)
	if ghost.HasLevelCrossing {
		t.Error("a ghost marked the road")
	}
	track.CompatibleNames = nil
	if r := m.trackClearRoad(c, &trackCmdState{}, &scenario.RoadElement{Rotation: 1, BaseZ: 8}, a, &lc); r != ClearCollisionErrorSet {
		t.Errorf("incompatible objects: %v", r)
	}
}
