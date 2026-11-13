package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// consTestRun finds a row of n empty flat tiles of one height west of (x, y)
// on the title map (rotation 0 builds towards -x) and calls try for each
// until it reports done.
func consTestRun(t *testing.T, s *State, n int, try func(x, y int) bool) {
	t.Helper()
	for y := 8; y < s.Height-8; y++ {
	next:
		for x := 8 + n; x < s.Width-8; x++ {
			z := s.TileAt(x, y).BaseZ
			for i := range n {
				if tl := s.TileAt(x-i, y); len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 || tl.BaseZ != z {
					continue next
				}
			}
			if try(x, y) {
				return
			}
		}
	}
	t.Skip("no place on the map took the pieces")
}

// countTrack / countRoad: the real and the ghost elements on a tile.
func countTrack(tl *Tile) (real, ghost int) {
	for _, r := range tl.Order {
		if r.Kind == scenario.KindTrack {
			if tl.Tracks[r.Index].Ghost {
				ghost++
			} else {
				real++
			}
		}
	}
	return
}

func countRoad(tl *Tile) (real, ghost int) {
	for _, r := range tl.Order {
		if r.Kind == scenario.KindRoad {
			if tl.Roads[r.Index].Ghost {
				ghost++
			} else {
				real++
			}
		}
	}
	return
}

// The Construction tab on a track, as the real OpenLoco behaves (olref,
// scripts/olref/tours/construction-rail.txt): the tool's ghost under the
// cursor; a click starts building and the state moves past the piece; the
// Construct button builds on; the bulldozer takes the pieces up one by one
// (through removeTrack: the money comes back) and the state moves back —
// after the last one it still stands where the first piece began, building
// on, a ghost there; one press more does nothing; the red cross goes back to
// hovering without turning; a refused click hovers on, the rotation kept.
func TestConstructionTrack(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	if len(om.TrackObjects) == 0 || om.TrackObjects[0] == nil {
		t.Skip("no track object")
	}
	g1, _ := om.G1File.(G1Raw)
	c := &Cmd{Company: s.ControllingCompany, Objects: om}
	consTestRun(t, s, 4, func(x, y int) bool {
		tl := s.TileAt(x, y)
		wx, wy := int32(x)*32, int32(y)*32
		s.ResetConstruction()
		s.ConstructionToolUpdate(c, x, y, -1, false, 0)
		if s.TrackGhost == nil {
			return false // refused here (owner, clearance …): another place
		}
		if _, g := countTrack(tl); g != 1 || !s.ConsArrow.On || s.ConsArrow.X != wx || s.ConsArrow.Y != wy || !s.Cons.Hover || s.Cons.Cost <= 0 {
			t.Fatalf("hovering: %d ghosts, arrow %+v, state %+v", g, s.ConsArrow, s.Cons)
		}
		cash := s.Companies[c.Company].Cash
		if _, built, _ := s.ConstructionToolUp(c, x, y, -1, false, 0); !built {
			return false
		}
		z := s.Cons.Z
		if r, g := countTrack(tl); r != 1 || g != 0 || s.Cons.Hover || s.Cons.X != wx-32 || s.Cons.Y != wy || s.Cons.Rotation != 0 {
			t.Fatalf("after the click: %d real %d ghost, state %+v", r, g, s.Cons)
		}
		for i := range 2 {
			if _, ok := s.ConstructTrack(c, 0); !ok {
				t.Fatalf("piece %d refused: error %d", i+2, c.ErrorText)
			}
		}
		if s.Cons.X != wx-96 || s.Companies[c.Company].Cash >= cash {
			t.Fatalf("after three pieces: state %+v, cash %d → %d", s.Cons, cash, s.Companies[c.Company].Cash)
		}
		s.UpdateConstruction(c, false, 0)
		if s.TrackGhost == nil || s.TrackGhost.X != wx-96 {
			t.Fatalf("the ghost while building: %+v", s.TrackGhost)
		}
		spent := s.Companies[c.Company].Cash
		for i := range 3 {
			if removed, refused := s.ConstructionRemoveTrack(c, 0, g1); !removed || refused {
				t.Fatalf("removing piece %d: removed %v refused %v error %d", 3-i, removed, refused, c.ErrorText)
			}
			if want := wx - 64 + int32(i)*32; s.Cons.X != want || s.Cons.Rotation != 0 || s.Cons.Hover {
				t.Fatalf("after removing %d: state %+v, want x %d", i+1, s.Cons, want)
			}
		}
		if s.Companies[c.Company].Cash <= spent {
			t.Errorf("removing paid nothing back: %d → %d", spent, s.Companies[c.Company].Cash)
		}
		for i := range 4 {
			if r, g := countTrack(s.TileAt(x-i, y)); r != 0 || g != 0 {
				t.Fatalf("tile %d after removing all: %d real %d ghost", i, r, g)
			}
		}
		// The last piece gone: still building on from where it began.
		want := ConstructionState{X: wx, Y: wy, Z: z, Mods: s.Cons.Mods, Bridge: s.Cons.Bridge}
		got := s.Cons
		got.arrowFrame, got.arrowBlink, got.Cost = 0, false, 0
		if got != want {
			t.Fatalf("after removing all: %+v, want %+v", got, want)
		}
		for range 2 { // one press more: nothing; the ghost is back with the next update
			s.UpdateConstruction(c, false, 0)
			if _, g := countTrack(tl); g != 1 || s.TrackGhost == nil {
				t.Fatalf("the ghost where the first piece was: %d on the tile, %+v", g, s.TrackGhost)
			}
			if removed, refused := s.ConstructionRemoveTrack(c, 0, g1); removed || refused {
				t.Fatalf("removing with nothing behind: removed %v refused %v", removed, refused)
			}
			if _, g := countTrack(tl); g != 0 || s.Cons.X != wx || s.Cons.Hover {
				t.Fatalf("after it: %d ghosts, state %+v", g, s.Cons)
			}
		}
		// A slope going down starts below its position (its first z is -16):
		// its ghost is found and taken up too.
		s.Cons.Gradient, s.Cons.Bridge = construct.SlopeDown, 1
		s.Cons.Z += 16
		s.UpdateConstruction(c, false, 0)
		if _, g := countTrack(tl); g != 1 || s.TrackGhost == nil {
			t.Fatalf("the slope-down ghost: %d on the tile, %+v (error %d)", g, s.TrackGhost, c.ErrorText)
		}
		s.ChangeTrackPiece(c, construct.Level, true)
		s.Cons.Z -= 16
		s.Cons.Bridge = 0
		for i := range 4 {
			if _, g := countTrack(s.TileAt(x-i, y)); g != 0 {
				t.Fatalf("a slope ghost stays on tile %d", i)
			}
		}
		// The red cross: hovering again, not turned, the ghost and arrow off.
		s.UpdateConstruction(c, false, 0)
		s.ConstructionRotate90(c)
		if _, g := countTrack(tl); g != 0 || !s.Cons.Hover || s.Cons.Rotation != 0 || s.ConsArrow.On {
			t.Fatalf("after the red cross: %d ghosts, state %+v, arrow %+v", g, s.Cons, s.ConsArrow)
		}
		// While hovering the same button turns a quarter.
		s.ConstructionRotate90(c)
		if s.Cons.Rotation != 1 || !s.Cons.Hover {
			t.Fatalf("rotate while hovering: %+v", s.Cons)
		}
		// A refused click: hovering on, the rotation as it was.
		if _, built, refused := s.ConstructionToolUp(c, 0, 0, -1, false, 0); built || !refused {
			t.Fatalf("a click on the map's edge: built %v refused %v", built, refused)
		}
		if !s.Cons.Hover || s.Cons.Rotation != 1 {
			t.Errorf("after a refused click: %+v, want hovering in rotation 1", s.Cons)
		}
		return true
	})
}

// The same for a road: three straights, taken up again one by one.
func TestConstructionRoad(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	obj := -1
	for i, o := range om.RoadObjects {
		if o != nil {
			obj = i
			break
		}
	}
	if obj < 0 {
		t.Skip("no road object")
	}
	o := uint8(obj)
	g1, _ := om.G1File.(G1Raw)
	c := &Cmd{Company: s.ControllingCompany, Objects: om}
	consTestRun(t, s, 4, func(x, y int) bool {
		tl := s.TileAt(x, y)
		wx := int32(x) * 32
		s.ResetConstruction()
		s.ConstructionToolUpdate(c, x, y, -1, true, o)
		if s.RoadGhost == nil {
			return false
		}
		if _, g := countRoad(tl); g != 1 {
			t.Fatalf("hovering: %d ghost roads", g)
		}
		if _, built, _ := s.ConstructionToolUp(c, x, y, -1, true, o); !built {
			return false
		}
		if r, g := countRoad(tl); r != 1 || g != 0 || s.Cons.Hover || s.Cons.X != wx-32 {
			t.Fatalf("after the click: %d real %d ghost, state %+v", r, g, s.Cons)
		}
		for _, ref := range tl.Order { // createRoad's element: the occupied quarters the clearance checks read
			if ref.Kind == scenario.KindRoad && ref.Flags&0xF == 0 {
				t.Errorf("the road's ref has no occupied quarters (flags %#x)", ref.Flags)
			}
		}
		for i := range 2 {
			if _, ok := s.ConstructRoad(c, o); !ok {
				t.Fatalf("piece %d refused: error %d", i+2, c.ErrorText)
			}
		}
		s.UpdateConstruction(c, true, o)
		if s.RoadGhost == nil || s.RoadGhost.X != wx-96 {
			t.Fatalf("the ghost while building: %+v", s.RoadGhost)
		}
		for i := range 3 {
			if removed, refused := s.ConstructionRemoveRoad(c, o, g1); !removed || refused {
				t.Fatalf("removing piece %d: removed %v refused %v error %d", 3-i, removed, refused, c.ErrorText)
			}
		}
		for i := range 4 {
			if r, g := countRoad(s.TileAt(x-i, y)); r != 0 || g != 0 {
				t.Fatalf("tile %d after removing all: %d real %d ghost", i, r, g)
			}
		}
		if s.Cons.X != wx || s.Cons.Rotation != 0 || s.Cons.Hover {
			t.Fatalf("after removing all: %+v", s.Cons)
		}
		s.UpdateConstruction(c, true, o)
		if _, g := countRoad(tl); g != 1 {
			t.Fatalf("the ghost where the first piece was: %d", g)
		}
		if removed, refused := s.ConstructionRemoveRoad(c, o, g1); removed || refused {
			t.Fatalf("removing with nothing behind: removed %v refused %v", removed, refused)
		}
		s.ConstructionRotate90(c)
		if _, g := countRoad(tl); g != 0 || !s.Cons.Hover || s.Cons.Rotation != 0 {
			t.Fatalf("after the red cross: %d ghosts, state %+v", g, s.Cons)
		}
		return true
	})
}

// Rotations 4–15: a large curve ends on the diagonal (rotation 12–15), a
// diagonal straight builds on from there, and the bulldozer finds both again
// (removeTrack's diagonal probe, one kRotationOffset step on) — back to the
// start in rotation 0.
func TestConstructionDiagonal(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	s := &State{Map: *m}
	if len(om.TrackObjects) == 0 || om.TrackObjects[0] == nil {
		t.Skip("no track object")
	}
	g1, _ := om.G1File.(G1Raw)
	c := &Cmd{Company: s.ControllingCompany, Objects: om}
	for y := 16; y < s.Height-16; y++ {
		for x := 16; x < s.Width-16; x++ {
			if tl := s.TileAt(x, y); len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 {
				continue
			}
			s.ResetConstruction()
			s.Cons.Piece = construct.LeftCurveLarge
			if _, built, _ := s.ConstructionToolUp(c, x, y, -1, false, 0); !built {
				continue
			}
			start := s.Cons
			if s.Cons.Rotation < 12 {
				t.Fatalf("a large curve ends in rotation %d", s.Cons.Rotation)
			}
			s.Cons.Piece = construct.Straight
			if _, ok := s.ConstructTrack(c, 0); !ok {
				// no room for the diagonal: take the curve up and look on
				if removed, _ := s.ConstructionRemoveTrack(c, 0, g1); !removed {
					t.Fatal("the curve doesn't come up")
				}
				continue
			}
			if s.Cons.Rotation != start.Rotation || (s.Cons.X == start.X && s.Cons.Y == start.Y) {
				t.Fatalf("after the diagonal: %+v (from %+v)", s.Cons, start)
			}
			if removed, _ := s.ConstructionRemoveTrack(c, 0, g1); !removed || s.Cons.X != start.X || s.Cons.Y != start.Y || s.Cons.Z != start.Z || s.Cons.Rotation != start.Rotation {
				t.Fatalf("the diagonal removed: %v, state %+v, want %+v", removed, s.Cons, start)
			}
			if removed, _ := s.ConstructionRemoveTrack(c, 0, g1); !removed || s.Cons.X != int32(x)*32 || s.Cons.Y != int32(y)*32 || s.Cons.Rotation != 0 {
				t.Fatalf("the curve removed: %v, state %+v, want (%d, %d) rotation 0", removed, s.Cons, x*32, y*32)
			}
			return
		}
	}
	t.Skip("no place on the map took a large curve and a diagonal")
}

// Every piece getTrackPieceId / getRoadPieceId gives for rotations 0–3 can
// be laid and ends somewhere else or turned (the s-bends, the turnaround,
// the slopes on small curves). getTrackPieceId gives the turnaround (track
// 42) at 0–3 though kTrackCoordinates begins it at 4–7: CreateTrack doesn't
// check rotationBegin, so it builds.
func TestPieceIDsLay(t *testing.T) {
	for _, pc := range [][2]uint8{
		{construct.SBendLeft, construct.Level}, {construct.SBendRight, construct.Level},
		{construct.SBendToDualTrack, construct.Level}, {construct.SBendToSingleTrack, construct.Level},
		{construct.Turnaround, construct.Level},
		{construct.LeftCurveSmall, construct.SlopeUp}, {construct.RightCurveSmall, construct.SteepSlopeDown},
	} {
		for rot := uint8(0); rot < 4; rot++ {
			p, ok := construct.GetTrackPieceID(pc[0], pc[1], rot)
			if !ok {
				t.Errorf("piece %d gradient %d rot %d: no track id", pc[0], pc[1], rot)
				continue
			}
			h := headAt(flatState(30, 64), 15, 15, 64, rot)
			start := h.e
			if !h.track(int(p.ID)) {
				t.Errorf("piece %d (track %d) rot %d: can't be laid", pc[0], p.ID, rot)
			} else if h.e == start {
				t.Errorf("piece %d (track %d) rot %d: ends where it starts", pc[0], p.ID, rot)
			}
		}
	}
	for _, piece := range []uint8{construct.Straight, construct.LeftCurveVerySmall, construct.RightCurveVerySmall,
		construct.LeftCurveSmall, construct.RightCurveSmall, construct.Turnaround} {
		for rot := uint8(0); rot < 4; rot++ {
			p, ok := construct.GetRoadPieceID(piece, construct.Level, rot)
			if !ok {
				t.Errorf("piece %d rot %d: no road id", piece, rot)
				continue
			}
			h := headAt(flatState(30, 64), 15, 15, 64, rot)
			start := h.e
			if !h.road(int(p.ID)) {
				t.Errorf("piece %d (road %d) rot %d: can't be laid", piece, p.ID, rot)
			} else if h.e == start {
				t.Errorf("piece %d (road %d) rot %d: ends where it starts", piece, p.ID, rot)
			}
		}
	}
}
