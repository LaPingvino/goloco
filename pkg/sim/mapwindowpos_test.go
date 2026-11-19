package sim

import "testing"

// Values worked by hand from MapWindow.cpp mapWindowPosToLocation (l. 144) and
// locationToMapWindowPos (l. 167).
func TestMapWindowPosToLocation(t *testing.T) {
	cases := []struct {
		x, y   int
		rot    uint8
		wx, wy int32
	}{
		// (400, 300): pos (12, 154), location ((154−12)·32, (12+154)·32) = (4544, 5312)
		{400, 300, 0, 4544, 5312},
		{400, 300, 1, 12287 - 5312, 4544},
		{400, 300, 2, 12287 - 4544, 12287 - 5312},
		{400, 300, 3, 5312, 12287 - 4544},
		// C++ division truncates towards zero: (1 + 8 − 384) / 2 = −187, not −188.
		{0, 0, 0, 192 * 32, -184 * 32},
		{1, 0, 0, 191 * 32, -183 * 32},
		// the map's top corner, rotations 1–3: off the map's far edges, unclamped
		{0, 0, 1, 12287 + 184*32, 192 * 32},
		{0, 0, 2, 12287 - 192*32, 12287 + 184*32},
		{0, 0, 3, -184 * 32, 12287 - 192*32},
	}
	for _, c := range cases {
		wx, wy := MapWindowPosToLocation(c.x, c.y, c.rot)
		if wx != c.wx || wy != c.wy {
			t.Errorf("MapWindowPosToLocation(%d, %d, rot %d) = (%d, %d), want (%d, %d)", c.x, c.y, c.rot, wx, wy, c.wx, c.wy)
		}
	}
}

func TestLocationToMapWindowPos(t *testing.T) {
	cases := []struct {
		wx, wy int32
		rot    uint8
		x, y   int
	}{
		{4544, 5312, 0, 408, 308},   // tile (142, 166): (−142+166+384, 142+166)
		{4544, 5312, 1, 459, 407},   // (y, 12287 − x) = (5312, 7743): tile (166, 241)
		{4544, 5312, 2, 360, 458},   // (12287 − x, 12287 − y) = (7743, 6975): tile (241, 217)
		{4544, 5312, 3, 309, 359},   // (12287 − y, x) = (6975, 4544): tile (217, 142)
		{0, 0, 0, 384, 0},           // the world's origin: the buffer's top centre
		{12287, 12287, 0, 384, 766}, // the far corner: the bottom centre
	}
	for _, c := range cases {
		x, y := LocationToMapWindowPos(c.wx, c.wy, c.rot)
		if x != c.x || y != c.y {
			t.Errorf("LocationToMapWindowPos(%d, %d, rot %d) = (%d, %d), want (%d, %d)", c.wx, c.wy, c.rot, x, y, c.x, c.y)
		}
	}
}

// TestMapWindowPosRoundTrip: rotation 0, a tile's map point goes back to the
// tile 8 further along y (the + 8s of mapWindowPosToLocation) — upstream's
// behaviour, kept.
func TestMapWindowPosRoundTrip(t *testing.T) {
	for _, tile := range [][2]int32{{10, 20}, {142, 166}, {0, 0}, {300, 50}} {
		x, y := LocationToMapWindowPos(tile[0]*32, tile[1]*32, 0)
		wx, wy := MapWindowPosToLocation(x, y, 0)
		if wx != tile[0]*32 || wy != (tile[1]+8)*32 {
			t.Errorf("tile %v: back at (%d, %d), want (%d, %d)", tile, wx, wy, tile[0]*32, (tile[1]+8)*32)
		}
	}
	// In every rotation the point a world position maps to lies in the buffer.
	for rot := range uint8(4) {
		for _, p := range [][2]int32{{0, 0}, {12287, 0}, {0, 12287}, {12287, 12287}, {6000, 3000}} {
			x, y := LocationToMapWindowPos(p[0], p[1], rot)
			if x < 0 || x >= MapRenderedWidth || y < 0 || y >= MapRenderedWidth {
				t.Errorf("rot %d %v: (%d, %d) off the buffer", rot, p, x, y)
			}
		}
	}
}
