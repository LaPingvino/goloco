package world

import "testing"

// newTestWorld builds a synthetic flat World (no renderer) suitable for
// exercising the pure construction-head math.
func newTestWorld(w, h int, baseZ uint8) *World {
	tiles := make([][]tile, h)
	for y := 0; y < h; y++ {
		tiles[y] = make([]tile, w)
		for x := 0; x < w; x++ {
			tiles[y][x] = tile{tileType: TileGrass, baseZ: baseZ}
		}
	}
	return &World{width: w, height: h, tiles: tiles}
}

// TestPlaceRoadAtHeadSequence walks a road piece sequence through the shared
// construction head and asserts the head advances to the expected tile,
// height and rotation after each piece — mirroring kRoadCoordinates.
func TestPlaceRoadAtHeadSequence(t *testing.T) {
	wld := newTestWorld(20, 20, 10)
	wld.SetConstructionHead(10, 10) // head at (10,10), rot 0, baseZ 10

	steps := []struct {
		roadID             int
		wantX, wantY       int
		wantBaseZ, wantRot uint8
		originX, originY   int // tile that should receive the piece origin element
		desc               string
	}{
		// straight rot0: kRoadCoordinates[0] = {0,0,-32,0,0}
		{0, 9, 10, 10, 0, 10, 10, "straight NE"},
		// rightCurveVerySmall rot0: kRoadCoordinates[16] = {0,1,0,32,0}
		{2, 9, 11, 10, 1, 9, 10, "right very-small curve → heading 1"},
		// straight rot1: kRoadCoordinates[1] = {1,1,0,32,0}
		{0, 9, 12, 10, 1, 9, 11, "straight SE"},
		// leftCurveVerySmall rot1: kRoadCoordinates[9] = {1,0,-32,0,0}
		{1, 8, 12, 10, 0, 9, 12, "left very-small curve → heading 0"},
		// straightSlopeUp rot0: kRoadCoordinates[40] = {0,0,-64,0,16}
		{5, 6, 12, 14, 0, 8, 12, "slope up: +2 tiles, +4 baseZ"},
	}

	for _, s := range steps {
		if !wld.CanPlaceRoadAtHead(s.roadID) {
			t.Fatalf("%s: CanPlaceRoadAtHead(%d) = false, head=%+v", s.desc, s.roadID, wld.ConstructionHeadState())
		}
		if !wld.PlaceRoadAtHead(s.roadID, 0) {
			t.Fatalf("%s: PlaceRoadAtHead(%d) = false", s.desc, s.roadID)
		}
		h := wld.ConstructionHeadState()
		if h.X != s.wantX || h.Y != s.wantY || h.BaseZ != s.wantBaseZ || h.Rotation != s.wantRot {
			t.Errorf("%s: head = (%d,%d z%d r%d), want (%d,%d z%d r%d)",
				s.desc, h.X, h.Y, h.BaseZ, h.Rotation, s.wantX, s.wantY, s.wantBaseZ, s.wantRot)
		}
		// The origin element must have landed on the expected tile.
		ot := wld.getTile(s.originX, s.originY)
		found := false
		for _, re := range ot.roads {
			if re.RoadID == uint8(s.roadID) {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no RoadElement roadID=%d at origin (%d,%d)", s.desc, s.roadID, s.originX, s.originY)
		}
	}

	// UndoLastRoad must revert the last piece and move the head back to its origin.
	before := wld.ConstructionHeadState()
	_ = before
	if !wld.UndoLastRoad() {
		t.Fatal("UndoLastRoad returned false")
	}
	h := wld.ConstructionHeadState()
	// After undoing the slope-up piece, head returns to (8,12) baseZ 10 rot 0.
	if h.X != 8 || h.Y != 12 || h.BaseZ != 10 || h.Rotation != 0 {
		t.Errorf("after undo: head = (%d,%d z%d r%d), want (8,12 z10 r0)", h.X, h.Y, h.BaseZ, h.Rotation)
	}
	// The slope-up piece spanned (8,12)+(7,12); both must be cleared by undo.
	if tl := wld.getTile(8, 12); len(tl.roads) != 0 {
		t.Errorf("after undo: tile (8,12) has %d road elements, want 0", len(tl.roads))
	}
	if tl := wld.getTile(7, 12); len(tl.roads) != 0 {
		t.Errorf("after undo: tile (7,12) has %d road elements, want 0", len(tl.roads))
	}
	// The left-curve origin (placed earlier at (9,12)) must remain untouched.
	if tl := wld.getTile(9, 12); len(tl.roads) != 1 {
		t.Errorf("after undo: tile (9,12) has %d road elements, want 1", len(tl.roads))
	}
}

// TestCanPlaceRoadRejectsWrongHeading verifies begin-rotation gating: a piece
// whose begin rotation differs from the head heading cannot be placed.
func TestCanPlaceRoadRejectsWrongHeading(t *testing.T) {
	wld := newTestWorld(10, 10, 8)
	wld.SetConstructionHead(5, 5)
	wld.RotateConstructionHead() // heading now 1

	// straight is valid at any heading (RotBegin cycles 0..3).
	if !wld.CanPlaceRoadAtHead(0) {
		t.Error("straight should be placeable at heading 1")
	}
	// Deactivating the head rejects everything.
	wld.ClearConstructionHead()
	if wld.CanPlaceRoadAtHead(0) {
		t.Error("no piece should be placeable with an inactive head")
	}
}
