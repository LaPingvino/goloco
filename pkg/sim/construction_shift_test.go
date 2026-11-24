package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// shiftTestState is a flat 16×16 map with its ground at z 128 and one
// track object, hovering in rotation 0 with a straight selected.
func shiftTestState() (*State, *Cmd) {
	hdr := &objects.ObjectHeader{}
	copy(hdr.Name[:], "TRACKST")
	om := &objects.ObjectManager{TrackObjects: []*objects.TrackObject{{Header: hdr, Name: "Railway"}}}
	s := &State{}
	s.Width, s.Height = 16, 16
	s.Tiles = make([][]Tile, 16)
	for y := range s.Tiles {
		s.Tiles[y] = make([]Tile, 16)
		for x := range s.Tiles[y] {
			s.Tiles[y][x].BaseZ = 32 // z 128
			s.Tiles[y][x].Order = []scenario.ElemRef{{Kind: scenario.KindSurface}}
		}
	}
	s.ResetConstruction()
	s.Cons.Bridge = 0
	return s, &Cmd{Company: CompanyNeutral, Objects: om}
}

// onToolUpdateSingle / constructionGhostLoop / onToolUpSingle /
// constructionLoop with shift (ConstructionTab.cpp): the arrow and the ghost
// start 32 below the ground instead of 16 and the tries go down, 16 at a
// time — under flat ground the straight finds its tunnel depth (past a
// tunnel across at that depth); without shift it stands on the ground. The
// tool's tiles are selected (enableConstruct); after the click the next
// piece's are.
func TestConstructionShift(t *testing.T) {
	s, c := shiftTestState()
	s.ConstructionToolUpdate(c, 8, 8, -1, false, 0, false)
	if s.TrackGhost == nil || s.TrackGhost.Z != 128 || s.ConsArrow.Z != 128 {
		t.Fatalf("without shift: ghost %+v arrow %+v, want both at the ground (128)", s.TrackGhost, s.ConsArrow)
	}
	if !s.ConsSelection.On || len(s.ConsSelection.Tiles) != 1 || s.ConsSelection.Tiles[0] != [2]int32{256, 256} {
		t.Fatalf("the selection: %+v, want the straight's one tile", s.ConsSelection)
	}

	s, c = shiftTestState()
	s.ConstructionToolUpdate(c, 8, 8, -1, false, 0, true)
	if s.ConsArrow.Z != 96 {
		t.Errorf("shift: the arrow at %d, want 96 (the ground less 32)", s.ConsArrow.Z)
	}
	if s.TrackGhost == nil || s.TrackGhost.Z != 96 {
		t.Fatalf("shift: ghost %+v, want it at 96 (error %d)", s.TrackGhost, c.ErrorText)
	}
	// A tunnel across at 96: the search goes on down past it.
	s.EndConstruction(c)
	if _, ok := s.CreateTrack(c, CmdApply, TrackPlacement{X: 256, Y: 256, Z: 96, Rotation: 1, Bridge: 0xFF}); !ok {
		t.Fatalf("the tunnel across: error %d", c.ErrorText)
	}
	s.ConstructionToolUpdate(c, 8, 8, -1, false, 0, true)
	if s.TrackGhost == nil || s.TrackGhost.Z >= 96 {
		t.Fatalf("shift over a tunnel: ghost %+v, want one below 96 (error %d)", s.TrackGhost, c.ErrorText)
	}
	ghostZ := s.TrackGhost.Z
	if _, built, _ := s.ConstructionToolUp(c, 8, 8, -1, false, 0, true); !built {
		t.Fatalf("shift click: refused, error %d", c.ErrorText)
	}
	// The click's tool selection went (toolCancel); building on, the
	// selection is the next piece's (activateSelectedConstructionWidgets:
	// unk_03, at the state's position).
	if s.ConsArrow.On || !s.ConsSelection.On || !s.ConsSelection.Unk03 || len(s.ConsSelection.Tiles) != 1 || s.ConsSelection.Tiles[0] != [2]int32{s.Cons.X, s.Cons.Y} {
		t.Errorf("after the click: arrow %v, selection %+v, want no arrow and the next piece's tile (%d, %d) in unk_03", s.ConsArrow.On, s.ConsSelection, s.Cons.X, s.Cons.Y)
	}
	tl := s.TileAt(8, 8)
	if r, _ := countTrack(tl); r != 2 || int32(tl.Tracks[1].BaseZ)*4 != ghostZ {
		t.Errorf("shift click: %d tracks, the new one at %d, want two, it where the ghost stood (%d)", r, int32(tl.Tracks[1].BaseZ)*4, ghostZ)
	}
}
