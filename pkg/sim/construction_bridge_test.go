package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// placeTrackGhost's bridge retry and sub_4A193B (ConstructionTab.cpp): a
// raised slope with a bridge that can't carry slopes is refused; just after
// the gradient button the window takes the first bridge that can, the retry
// itself (upstream's old args) still fails and the next update's ghost
// stands with the new bridge. Without a piece button since, nothing
// changes. A ghost that meets a bridged element takes that element's bridge.
func TestConstructionBridgeRetry(t *testing.T) {
	newState := func() (*State, *Cmd) {
		hdr := &objects.ObjectHeader{}
		copy(hdr.Name[:], "TRACKST")
		track := &objects.TrackObject{Header: hdr, Name: "Railway"}
		flat := &objects.BridgeObject{Name: "flat only", MaxHeight: 30, DisabledTrackCfg: commonTraitSlope | commonTraitSteepSlope}
		any := &objects.BridgeObject{Name: "any", MaxHeight: 30}
		om := &objects.ObjectManager{TrackObjects: []*objects.TrackObject{track}, BridgeObjects: []*objects.BridgeObject{flat, any}}
		s := &State{}
		s.Width, s.Height = 16, 16
		s.Tiles = make([][]Tile, 16)
		for y := range s.Tiles {
			s.Tiles[y] = make([]Tile, 16)
			for x := range s.Tiles[y] {
				s.Tiles[y][x].BaseZ = 4
				s.Tiles[y][x].Order = []scenario.ElemRef{{Kind: scenario.KindSurface}}
			}
		}
		s.Cons = ConstructionState{X: 8 * 32, Y: 8 * 32, Z: 96, Bridge: 1, BridgeList: [9]uint8{1, 2}}
		return s, &Cmd{Company: CompanyNeutral, Objects: om}
	}

	s, c := newState()
	s.ChangeTrackPiece(c, construct.SlopeUp, true)
	s.UpdateConstruction(c, false, 0)
	if s.TrackGhost != nil || s.Cons.Bridge != 2 {
		t.Fatalf("after the gradient button: ghost %+v, bridge %d (want none and the second bridge)", s.TrackGhost, s.Cons.Bridge)
	}
	s.UpdateConstruction(c, false, 0)
	if s.TrackGhost == nil {
		t.Fatalf("the next update: no ghost with the new bridge (error %d)", c.ErrorText)
	}

	s, c = newState()
	s.Cons.Gradient = construct.SlopeUp // no piece button since the window opened
	s.UpdateConstruction(c, false, 0)
	if s.TrackGhost != nil || s.Cons.Bridge != 1 {
		t.Fatalf("without a piece button: ghost %+v, bridge %d (want none, the first)", s.TrackGhost, s.Cons.Bridge)
	}

	// sub_4A193B: a bridged straight across the ghost's tile; the ghost
	// meeting it (refused: no junctions on this track) takes its bridge.
	s, c = newState()
	s.Cons.Bridge = 1
	if _, ok := s.createTrack(c, trackPlacementArgs{X: 8 * 32, Y: 8 * 32, Z: 96, Rotation: 1, Bridge: 1, TrackObjectID: 0}, CmdApply); !ok {
		t.Fatalf("the bridged straight: error %d", c.ErrorText)
	}
	s.UpdateConstruction(c, false, 0)
	if s.Cons.Bridge != 2 {
		t.Errorf("after meeting the bridged track: bridge %d, want 2 (c.Bridge %d)", s.Cons.Bridge, c.Bridge)
	}
}
