package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// World::gameToScreen (Map/Tile.cpp) at each rotation: rotate, then
// (y − x, (y + x) >> 1 − z) — values worked from upstream's formula.
func TestGameToScreenRotations(t *testing.T) {
	want := [4][2]int32{{-32, 88}, {-224, -40}, {32, -136}, {224, -8}} // of (128, 96, 24)
	for rot := uint8(0); rot < 4; rot++ {
		if x, y := GameToScreen(128, 96, 24, rot); [2]int32{x, y} != want[rot] {
			t.Errorf("rotation %d: (128, 96, 24) → (%d, %d), want %v", rot, x, y, want[rot])
		}
	}
}

// Ui::viewportCoordToMapCoord undoes gameToScreen at the same height and
// rotation (to the 2 units the halving loses).
func TestViewportCoordToMapCoordRoundTrip(t *testing.T) {
	for rot := uint8(0); rot < 4; rot++ {
		for _, p := range [][3]int32{{128, 96, 24}, {6143, 6143, 0}, {17, 1000, 136}, {12000, 40, 508}} {
			vx, vy := GameToScreen(p[0], p[1], p[2], rot)
			x, y := ViewportCoordToMapCoord(vx, vy, p[2], rot)
			if dx, dy := x-p[0], y-p[1]; dx < -2 || dx > 2 || dy < -2 || dy > 2 {
				t.Errorf("rotation %d: %v → view (%d, %d) → (%d, %d)", rot, p, vx, vy, x, y)
			}
		}
	}
}

// EntityManager::updateSpatialIndex after a rotation: every entity's sprite
// bounds sit about its gameToScreen point at the new rotation (EntityBase::moveTo),
// and the old and new areas are invalidated.
func TestUpdateSpatialIndexFollowsRotation(t *testing.T) {
	var m Map
	m.Entities = []scenario.VehicleEntity{{PosX: 128, PosY: 96, PosZ: 24, SpriteWidth: 10, SpriteHeightNegative: 6, SpriteHeightPositive: 4}}
	var inval [][4]int16
	m.InvalidateEntity = func(l, t, r, b int16) { inval = append(inval, [4]int16{l, t, r, b}) }
	for rot := uint8(0); rot < 4; rot++ {
		inval = inval[:0]
		m.CurrentRotation = rot
		m.UpdateSpatialIndex()
		vx, vy := GameToScreen(128, 96, 24, rot)
		e := &m.Entities[0]
		if got, want := [4]int16{e.SpriteLeft, e.SpriteTop, e.SpriteRight, e.SpriteBottom}, [4]int16{int16(vx - 10), int16(vy - 6), int16(vx + 10), int16(vy + 4)}; got != want {
			t.Errorf("rotation %d: bounds %v, want %v", rot, got, want)
		}
		if len(inval) != 2 || inval[1] != [4]int16{e.SpriteLeft, e.SpriteTop, e.SpriteRight, e.SpriteBottom} {
			t.Errorf("rotation %d: invalidated %v, want the old and the new bounds", rot, inval)
		}
	}
}

// TileManager::getHeight's helpers measure a tile's slope with kTileSize 31
// (the last sub-tile coordinate): values worked from TileManager.cpp.
func TestSurfaceLandHeightTileSize31(t *testing.T) {
	for _, c := range []struct {
		slope  uint8
		xl, yl int16
		want   int16
	}{
		{slopeSideSE, 16, 16, 7},       // (31 − 16) / 2
		{slopeSideSW, 0, 0, 15},        // (31 − 0) / 2
		{slopeSideNE, 31, 0, 16},       // 31 / 2 + 1
		{slopeCornerNorth, 31, 31, 15}, // (31 + 31 − 31) / 2
		{slopeCornerSouth, 0, 0, 15},   // (31 − 0 − 0) / 2
		{slopeDownNorth, 0, 0, 31},     // (31 − 0 − 0 − 1) / 2 + 16
		{slopeValleyWE, 31, 31, 0},     // 31 − 62 < 0
	} {
		if got := surfaceLandHeight(c.slope, c.xl, c.yl); got != c.want {
			t.Errorf("slope %#x at (%d, %d): %d, want %d", c.slope, c.xl, c.yl, got, c.want)
		}
	}
}
