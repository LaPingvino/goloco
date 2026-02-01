package scenario

import "testing"

// ---------------------------------------------------------------------------
// parseTileElements
// ---------------------------------------------------------------------------

// makeSurfaceElement builds an 8-byte surface TileElement with the given
// terrain, water level, baseZ, and last flag.
func makeSurfaceElement(terrain uint8, waterLevel uint8, baseZ uint8, last bool) [8]byte {
	var e [8]byte
	e[0] = elementTypeSurface << elementTypeShift // type = surface (0)
	if last {
		e[1] = elementFlagLast
	}
	e[2] = baseZ
	e[3] = baseZ // clearZ = baseZ for simplicity
	e[4] = 0     // slope (not used yet)
	e[5] = waterLevel & 0x1F
	e[6] = terrain & 0x1F
	e[7] = 0     // variation
	return e
}

func TestParseTileElements_single_tile(t *testing.T) {
	// One surface element with last flag → fills tile (0,0)
	elem := makeSurfaceElement(0, 0, 5, true) // terrain=grass, no water, baseZ=5

	sc := &Scenario{MapWidth: 2, MapHeight: 2}
	sc.parseTileElements(elem[:])

	tile := sc.Tiles[0][0]
	if tile.Surface != SurfaceGrass {
		t.Errorf("(0,0) Surface = %d, want SurfaceGrass (%d)", tile.Surface, SurfaceGrass)
	}
	if tile.Height != 5 {
		t.Errorf("(0,0) Height = %d, want 5", tile.Height)
	}
	if tile.Water != 0 {
		t.Errorf("(0,0) Water = %d, want 0", tile.Water)
	}
}

func TestParseTileElements_water_overrides_terrain(t *testing.T) {
	// Terrain = grass but water level > 0 → surface should be Water
	elem := makeSurfaceElement(0, 3, 2, true)

	sc := &Scenario{MapWidth: 1, MapHeight: 1}
	sc.parseTileElements(elem[:])

	if sc.Tiles[0][0].Surface != SurfaceWater {
		t.Errorf("Surface = %d, want SurfaceWater (%d)", sc.Tiles[0][0].Surface, SurfaceWater)
	}
	if sc.Tiles[0][0].Water != 3 {
		t.Errorf("Water = %d, want 3", sc.Tiles[0][0].Water)
	}
}

func TestParseTileElements_multiple_tiles_row(t *testing.T) {
	// Three tiles in a row: grass, dirt (terrain=1), rock (terrain=2)
	var data []byte
	e := makeSurfaceElement(0, 0, 1, true) // grass
	data = append(data, e[:]...)
	e = makeSurfaceElement(1, 0, 2, true) // dirt
	data = append(data, e[:]...)
	e = makeSurfaceElement(2, 0, 3, true) // rock
	data = append(data, e[:]...)

	sc := &Scenario{MapWidth: 3, MapHeight: 1}
	sc.parseTileElements(data)

	cases := []struct {
		x       int
		surface SurfaceType
		height  uint8
	}{
		{0, SurfaceGrass, 1},
		{1, SurfaceDirt, 2},
		{2, SurfaceRock, 3},
	}
	for _, c := range cases {
		tile := sc.Tiles[0][c.x]
		if tile.Surface != c.surface {
			t.Errorf("(%d,0) Surface = %d, want %d", c.x, tile.Surface, c.surface)
		}
		if tile.Height != c.height {
			t.Errorf("(%d,0) Height = %d, want %d", c.x, tile.Height, c.height)
		}
	}
}

func TestParseTileElements_wraps_to_next_row(t *testing.T) {
	// 2×2 grid: fill all 4 tiles, verify row wrapping
	var data []byte
	heights := [4]uint8{10, 20, 30, 40}
	for _, h := range heights {
		e := makeSurfaceElement(0, 0, h, true)
		data = append(data, e[:]...)
	}

	sc := &Scenario{MapWidth: 2, MapHeight: 2}
	sc.parseTileElements(data)

	// Tile order: (0,0) (1,0) (0,1) (1,1)
	expected := [2][2]uint8{{10, 20}, {30, 40}}
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if sc.Tiles[y][x].Height != expected[y][x] {
				t.Errorf("(%d,%d) Height = %d, want %d", x, y, sc.Tiles[y][x].Height, expected[y][x])
			}
		}
	}
}

func TestParseTileElements_stacked_elements_only_surface_matters(t *testing.T) {
	// A tile with two elements: surface (last=false) then a non-surface
	// element (last=true).  Only the surface element should affect the tile.
	var data []byte

	// Surface element, NOT last
	surf := makeSurfaceElement(2, 0, 7, false) // rock, baseZ=7
	data = append(data, surf[:]...)

	// Non-surface element (e.g. track = type 1), last=true
	var track [8]byte
	track[0] = 1 << elementTypeShift // type = track
	track[1] = elementFlagLast
	track[2] = 7
	data = append(data, track[:]...)

	sc := &Scenario{MapWidth: 1, MapHeight: 1}
	sc.parseTileElements(data)

	if sc.Tiles[0][0].Surface != SurfaceRock {
		t.Errorf("Surface = %d, want SurfaceRock (%d)", sc.Tiles[0][0].Surface, SurfaceRock)
	}
	if sc.Tiles[0][0].Height != 7 {
		t.Errorf("Height = %d, want 7", sc.Tiles[0][0].Height)
	}
}

func TestParseTileElements_empty_data(t *testing.T) {
	// Zero bytes → all tiles stay at defaults (grass, height 0)
	sc := &Scenario{MapWidth: 2, MapHeight: 2}
	sc.parseTileElements(nil)

	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if sc.Tiles[y][x].Surface != SurfaceGrass {
				t.Errorf("(%d,%d) Surface = %d, want default grass", x, y, sc.Tiles[y][x].Surface)
			}
		}
	}
}

func TestParseTileElements_partial_element_ignored(t *testing.T) {
	// 7 bytes (one short of a full element) → nothing parsed, defaults remain
	sc := &Scenario{MapWidth: 1, MapHeight: 1}
	sc.parseTileElements([]byte{0, 0x80, 5, 5, 0, 0, 2, /* missing 8th byte */})

	// Should still have defaults
	if sc.Tiles[0][0].Surface != SurfaceGrass {
		t.Errorf("partial element: Surface = %d, want default grass", sc.Tiles[0][0].Surface)
	}
}

// ---------------------------------------------------------------------------
// terrain type mapping
// ---------------------------------------------------------------------------

func TestTerrainTypes_known_values(t *testing.T) {
	cases := []struct {
		raw  uint8
		want SurfaceType
	}{
		{0, SurfaceGrass},
		{1, SurfaceDirt},
		{2, SurfaceRock},
		{5, SurfaceSnow},
	}
	for _, c := range cases {
		got := terrainTypes[c.raw]
		if got != c.want {
			t.Errorf("terrainTypes[%d] = %d, want %d", c.raw, got, c.want)
		}
	}
}
