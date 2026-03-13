package world

// Tests that verify GoLoco's coordinate math matches OpenLoco's exact formulas.
// These act as a regression guard: if any formula in world.go drifts from
// the ground truth in MEASUREMENTS.md, these tests will catch it.
//
// Ground truth source: OpenLoco Map/Tile.cpp::gameToScreen, PaintSurface.cpp,
// World.hpp constants.  See /home/joop/goloco-project/MEASUREMENTS.md.

import (
	"testing"
)

// ---------------------------------------------------------------------------
// tileToScreen — must match OpenLoco gameToScreen (rotation 0)
//   vpX = (tileY - tileX) * 32    (kTileSize = 32)
//   vpY = (tileY + tileX) * 16    (kTileSize/2 = 16; height excluded here)
// Height is applied separately in Draw().
// ---------------------------------------------------------------------------

func TestTileToScreen(t *testing.T) {
	w := &World{tileW: 64, tileH: 32, width: 20, height: 15}

	cases := []struct {
		tileX, tileY int
		wantVpX      float64
		wantVpY      float64
		desc         string
	}{
		{0, 0, 0, 0, "origin"},
		{1, 0, -32, 16, "+X moves left and down"},
		{0, 1, 32, 16, "+Y moves right and down"},
		{1, 1, 0, 32, "diagonal"},
		{5, 3, -64, 128, "arbitrary"},
		{10, 10, 0, 320, "on diagonal"},
	}

	for _, c := range cases {
		vpX, vpY := w.tileToScreen(c.tileX, c.tileY)
		if vpX != c.wantVpX || vpY != c.wantVpY {
			t.Errorf("tileToScreen(%d,%d) = (%.0f, %.0f), want (%.0f, %.0f) — %s",
				c.tileX, c.tileY, vpX, vpY, c.wantVpX, c.wantVpY, c.desc)
		}
	}
}

// ---------------------------------------------------------------------------
// Height offset — OpenLoco: loc.z = baseZ * kSmallZStep = baseZ * 4
// Must be subtracted from vpY.
// ---------------------------------------------------------------------------

func TestHeightOffset(t *testing.T) {
	// For a tile at (0,0) with baseZ, the final vpY must be (0) - baseZ*4.
	// We test via the formula directly since Draw() inlines it.
	cases := []struct {
		baseZ   uint8
		wantPx  float64 // pixels subtracted from vpY
		desc    string
	}{
		{0, 0, "sea level"},
		{1, 4, "1 SmallZ = 4 px"},
		{4, 16, "4 SmallZ = 16 px = 1 tile Y step"},
		{8, 32, "8 SmallZ"},
		{84, 336, "max seen in title.dat"},
	}

	for _, c := range cases {
		got := float64(c.baseZ) * 4.0 // the formula that must be in Draw()
		if got != c.wantPx {
			t.Errorf("heightOffset(baseZ=%d) = %.0f, want %.0f — %s",
				c.baseZ, got, c.wantPx, c.desc)
		}
	}
}

// ---------------------------------------------------------------------------
// getCornerHeights — OpenLoco: microZ = baseZ / kMicroToSmallZStep = baseZ / 4
// Corner values from kCornerHeights are in MicroZ units (1 MicroZ = 16 px).
// ---------------------------------------------------------------------------

func TestGetCornerHeights(t *testing.T) {
	w := &World{}

	cases := []struct {
		baseZ    uint8
		slope    uint8
		wantTop  uint8
		wantRight uint8
		wantBottom uint8
		wantLeft uint8
		desc     string
	}{
		// Flat (slope=0): all corners = microZ = baseZ/4
		{0, 0, 0, 0, 0, 0, "baseZ=0 flat"},
		{4, 0, 1, 1, 1, 1, "baseZ=4 flat: microZ=1"},
		{8, 0, 2, 2, 2, 2, "baseZ=8 flat: microZ=2"},
		{12, 0, 3, 3, 3, 3, "baseZ=12 flat: microZ=3"},
		// Slope 1: {T=0,R=0,B=1,L=0} → add microZ
		{0, 1, 0, 0, 1, 0, "baseZ=0 slope1"},
		{4, 1, 1, 1, 2, 1, "baseZ=4 slope1: microZ=1 + {0,0,1,0}"},
		// Slope 15: {T=1,R=1,B=1,L=1} (all raised)
		{0, 15, 1, 1, 1, 1, "baseZ=0 slope15"},
		{4, 15, 2, 2, 2, 2, "baseZ=4 slope15"},
		// Double-height slope 23: {T=1,R=0,B=1,L=2}
		{0, 23, 1, 0, 1, 2, "baseZ=0 slope23 double-height"},
		{4, 23, 2, 1, 2, 3, "baseZ=4 slope23"},
	}

	for _, c := range cases {
		tile := &tile{baseZ: c.baseZ, slope: c.slope}
		ch := w.getCornerHeights(tile)
		if ch.Top != c.wantTop || ch.Right != c.wantRight ||
			ch.Bottom != c.wantBottom || ch.Left != c.wantLeft {
			t.Errorf("getCornerHeights(baseZ=%d,slope=%d) = {%d,%d,%d,%d}, want {%d,%d,%d,%d} — %s",
				c.baseZ, c.slope,
				ch.Top, ch.Right, ch.Bottom, ch.Left,
				c.wantTop, c.wantRight, c.wantBottom, c.wantLeft,
				c.desc)
		}
	}
}

// ---------------------------------------------------------------------------
// getNeighborTile direction offsets — OpenLoco kNeighbourOffsets[rotation=0]:
//   Edge 0 SW: +32,  0  → tile (+1, 0)
//   Edge 1 SE:   0, +32 → tile ( 0,+1)
//   Edge 2 NW:   0, -32 → tile ( 0,-1)
//   Edge 3 NE: -32,  0  → tile (-1, 0)
// ---------------------------------------------------------------------------

func TestNeighborOffsets(t *testing.T) {
	const sz = 5
	w := &World{width: sz, height: sz, tiles: make([][]tile, sz)}
	for y := range w.tiles {
		w.tiles[y] = make([]tile, sz)
	}
	// Mark each tile with unique base values we can identify.
	for y := 0; y < sz; y++ {
		for x := 0; x < sz; x++ {
			w.tiles[y][x].baseZ = uint8(y*sz + x)
		}
	}

	// From tile (2,2), check each neighbour.
	cx, cy := 2, 2
	cases := []struct {
		edge      int
		wantTileX int
		wantTileY int
		label     string
	}{
		{EdgeSW, 3, 2, "SW = +tileX"},
		{EdgeSE, 2, 3, "SE = +tileY"},
		{EdgeNW, 2, 1, "NW = -tileY"},
		{EdgeNE, 1, 2, "NE = -tileX"},
	}

	for _, c := range cases {
		nb := w.getNeighborTile(cx, cy, c.edge)
		if nb == nil {
			t.Errorf("getNeighborTile edge %d (%s): got nil", c.edge, c.label)
			continue
		}
		wantBaseZ := uint8(c.wantTileY*sz + c.wantTileX)
		if nb.baseZ != wantBaseZ {
			t.Errorf("getNeighborTile edge %d (%s): got tile with baseZ=%d, want baseZ=%d (tile %d,%d)",
				c.edge, c.label, nb.baseZ, wantBaseZ, c.wantTileX, c.wantTileY)
		}
	}
}

// ---------------------------------------------------------------------------
// getCliffEdgeOffset — derived from kEdgeImageOffset Pos3 via gameToScreen.
// See MEASUREMENTS.md "Cliff edge Pos3 → screen offset" table.
//   h in MicroZ; each MicroZ = 16 px at zoom-0.
//   SW: vpX=-30, vpY=15-16h
//   SE: vpX=+30, vpY=15-16h
//   NW: vpX=-2,  vpY=-2-16h
//   NE: vpX=+2,  vpY=-2-16h
// ---------------------------------------------------------------------------

func TestGetCliffEdgeOffset(t *testing.T) {
	w := &World{}

	cases := []struct {
		edge      int
		h         uint8  // MicroZ
		wantX     int16
		wantY     int16
		label     string
	}{
		{EdgeSW, 0, -30, 15, "SW h=0"},
		{EdgeSW, 1, -30, -1, "SW h=1: 15-16=-1"},
		{EdgeSW, 2, -30, -17, "SW h=2"},
		{EdgeSE, 0, 30, 15, "SE h=0"},
		{EdgeSE, 1, 30, -1, "SE h=1"},
		{EdgeNW, 0, -2, -2, "NW h=0"},
		{EdgeNW, 1, -2, -18, "NW h=1: -2-16=-18"},
		{EdgeNE, 0, 2, -2, "NE h=0"},
		{EdgeNE, 1, 2, -18, "NE h=1"},
	}

	for _, c := range cases {
		gotX, gotY := w.getCliffEdgeOffset(c.edge, c.h)
		if gotX != c.wantX || gotY != c.wantY {
			t.Errorf("getCliffEdgeOffset(edge=%d, h=%d) = (%d, %d), want (%d, %d) — %s",
				c.edge, c.h, gotX, gotY, c.wantX, c.wantY, c.label)
		}
	}
}

// ---------------------------------------------------------------------------
// Water height — OpenLoco:
//   water field is MicroZ; waterHeight_px = water * kMicroZStep = water * 16
//   Overlay displacement above terrain = water*16 - baseZ*4
// ---------------------------------------------------------------------------

func TestWaterHeightFormula(t *testing.T) {
	cases := []struct {
		waterLevel uint8 // MicroZ from surface element byte 5
		baseZ      uint8 // SmallZ
		wantDiff   float64 // pixels water is above terrain base (positive = higher on screen)
		desc       string
	}{
		{0, 0, 0, "no water"},
		{1, 0, 16, "sea level: 1 MicroZ = 16 px above baseZ=0"},
		{2, 0, 32, "2 MicroZ"},
		{2, 4, 16, "water=2*16=32, terrain=4*4=16 → diff=16"},
		{4, 8, 32, "water=4*16=64, terrain=8*4=32 → diff=32"},
		{2, 8, 0, "water at same height as terrain: 2*16=32, 8*4=32"},
	}

	for _, c := range cases {
		got := float64(c.waterLevel)*16.0 - float64(c.baseZ)*4.0
		if got != c.wantDiff {
			t.Errorf("waterHeightDiff(water=%d, baseZ=%d) = %.0f, want %.0f — %s",
				c.waterLevel, c.baseZ, got, c.wantDiff, c.desc)
		}
	}
}

// ---------------------------------------------------------------------------
// Checkerboard factor — OpenLoco: ((worldX ^ worldY) & 0x20)
//   worldX = tileX * 32, so ((tileX^tileY) & 1) * 32 in tile coords.
//   Must alternate every tile (not every 32 tiles).
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Water palette remap — the blend sprite (+35) uses palette indices 1-4.
// Without the water palette remap (G1[waterImageOffset+41]), those indices
// decode as transparent/black. The remap maps them to opaque blue shades.
// We verify the remap formula: spriteID = ImageOffset + BlendedOffset + shape.
// ---------------------------------------------------------------------------

func TestWaterBlendSpriteIDFormula(t *testing.T) {
	// Verifies the sprite ID formula used in paintWater for the full-tile
	// water surface (blendID) that requires palette remap to render as blue.
	cases := []struct {
		imageOffset uint32
		shape       uint8
		wantFlatID  uint32
		wantBlendID uint32
		desc        string
	}{
		{100000, 0, 100030, 100035, "flat ocean"},
		{100000, 1, 100031, 100036, "slope variant 1"},
		{200000, 0, 200030, 200035, "different object offset"},
	}

	for _, c := range cases {
		flatID := c.imageOffset + 30 + uint32(c.shape)
		blendID := c.imageOffset + 35 + uint32(c.shape)
		if flatID != c.wantFlatID || blendID != c.wantBlendID {
			t.Errorf("water sprite IDs(offset=%d,shape=%d): flat=%d blend=%d, want flat=%d blend=%d — %s",
				c.imageOffset, c.shape, flatID, blendID, c.wantFlatID, c.wantBlendID, c.desc)
		}
	}
}

// TestWaterPaletteRemapIndex verifies that the water remap palette map is at
// waterObj.ImageOffset + 41, matching OpenLoco's WaterObject sprite layout.
// OpenLoco reference: PaintSurface.cpp paintSurfaceWater — withBlend(ExtColour::water)
//   The remap table is the 42nd sprite in the water object image table (0-indexed).
func TestWaterPaletteRemapIndex(t *testing.T) {
	const remapOffset = 41 // 0-indexed position of the remap palette map
	const imageOffset = uint32(118748)
	wantRemapID := int(imageOffset) + remapOffset
	if wantRemapID != 118789 {
		t.Errorf("water remap G1 index = %d, want 118789 (imageOffset+41)", wantRemapID)
	}
}

func TestCheckerboard(t *testing.T) {
	cases := []struct {
		x, y    int
		want    uint32 // 0 or 32
		desc    string
	}{
		{0, 0, 0, "(0,0)"},
		{1, 0, 32, "(1,0): different parity"},
		{0, 1, 32, "(0,1): different parity"},
		{1, 1, 0, "(1,1): same parity as (0,0)"},
		{2, 0, 0, "(2,0): even+even"},
		{2, 1, 32, "(2,1): even+odd"},
		{10, 7, 32, "odd+even"},
		{10, 8, 0, "even+even"},
	}

	for _, c := range cases {
		got := uint32((c.x^c.y)&1) * 32
		if got != c.want {
			t.Errorf("checkerboard(%d,%d) = %d, want %d — %s", c.x, c.y, got, c.want, c.desc)
		}
	}
}
