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
// TerrainIndex propagation
// ---------------------------------------------------------------------------

func TestParseTileElements_terrain_index_stored(t *testing.T) {
	// Three tiles with distinct terrain indices; verify TerrainIndex matches the
	// raw value written into byte 6 of each element.
	var data []byte
	indices := [3]uint8{0, 7, 15}
	for _, idx := range indices {
		e := makeSurfaceElement(idx, 0, 1, true)
		data = append(data, e[:]...)
	}

	sc := &Scenario{MapWidth: 3, MapHeight: 1}
	sc.parseTileElements(data)

	for i, want := range indices {
		got := sc.Tiles[0][i].TerrainIndex
		if got != want {
			t.Errorf("(%d,0) TerrainIndex = %d, want %d", i, got, want)
		}
	}
}

func TestParseTileElements_terrain_index_water_tile_preserves_index(t *testing.T) {
	// Even when water overrides the surface type, TerrainIndex must still
	// reflect the underlying terrain stored in byte 6.
	elem := makeSurfaceElement(12, 5, 3, true) // terrain=12, water=5

	sc := &Scenario{MapWidth: 1, MapHeight: 1}
	sc.parseTileElements(elem[:])

	if sc.Tiles[0][0].Surface != SurfaceWater {
		t.Fatalf("Surface = %d, want SurfaceWater", sc.Tiles[0][0].Surface)
	}
	if sc.Tiles[0][0].TerrainIndex != 12 {
		t.Errorf("TerrainIndex = %d, want 12", sc.Tiles[0][0].TerrainIndex)
	}
}

// ---------------------------------------------------------------------------
// parseLandObjectOrder
// ---------------------------------------------------------------------------

// makeObjectHeader builds a 16-byte ObjectHeader: flags(u32) name[8] checksum(u32).
func makeObjectHeader(objType uint8, name string) [16]byte {
	var h [16]byte
	// flags: type in bits [5:0]
	h[0] = objType & objectTypeMask
	// name: copy up to 8 bytes, pad with 0xFF
	for i := 0; i < 8; i++ {
		if i < len(name) {
			h[4+i] = name[i]
		} else {
			h[4+i] = 0xFF
		}
	}
	// checksum left as 0 — not checked by parser
	return h
}

// makeEmptyHeader returns an all-0xFF 16-byte header (empty slot).
func makeEmptyHeader() [16]byte {
	var h [16]byte
	for i := range h {
		h[i] = 0xFF
	}
	return h
}

// makeRequiredObjectsBlob builds a minimal RequiredObjects blob large enough
// to cover the land-object slot range.  Slots before landSlotStart are filled
// with dummy non-land headers; the land slots are filled from the provided
// array (nil entries become empty headers).
func makeRequiredObjectsBlob(landNames []string) []byte {
	totalSlots := landSlotStart + landSlotCount
	buf := make([]byte, 0, totalSlots*objectHeaderSize)

	// Fill pre-land slots with empty headers
	for i := 0; i < landSlotStart; i++ {
		h := makeEmptyHeader()
		buf = append(buf, h[:]...)
	}

	// Fill land slots
	for i := 0; i < landSlotCount; i++ {
		if i < len(landNames) && landNames[i] != "" {
			h := makeObjectHeader(objectTypeLand, landNames[i])
			buf = append(buf, h[:]...)
		} else {
			h := makeEmptyHeader()
			buf = append(buf, h[:]...)
		}
	}
	return buf
}

func TestParseLandObjectOrder_basic(t *testing.T) {
	names := []string{"GRASS1", "SAND1", "", "ROCK1"}
	blob := makeRequiredObjectsBlob(names)
	order := parseLandObjectOrder(blob)

	if len(order) != landSlotCount {
		t.Fatalf("len(order) = %d, want %d", len(order), landSlotCount)
	}
	if order[0] != "GRASS1" {
		t.Errorf("order[0] = %q, want %q", order[0], "GRASS1")
	}
	if order[1] != "SAND1" {
		t.Errorf("order[1] = %q, want %q", order[1], "SAND1")
	}
	if order[2] != "" {
		t.Errorf("order[2] = %q, want empty (empty slot)", order[2])
	}
	if order[3] != "ROCK1" {
		t.Errorf("order[3] = %q, want %q", order[3], "ROCK1")
	}
}

func TestParseLandObjectOrder_all_empty(t *testing.T) {
	blob := makeRequiredObjectsBlob(nil)
	order := parseLandObjectOrder(blob)

	if len(order) != landSlotCount {
		t.Fatalf("len(order) = %d, want %d", len(order), landSlotCount)
	}
	for i, name := range order {
		if name != "" {
			t.Errorf("order[%d] = %q, want empty", i, name)
		}
	}
}

func TestParseLandObjectOrder_wrong_type_ignored(t *testing.T) {
	// Put a non-land header (type=2, rock object) in land slot 0.
	// It should not appear in the output because its type doesn't match.
	totalSlots := landSlotStart + landSlotCount
	buf := make([]byte, totalSlots*objectHeaderSize)
	for i := range buf {
		buf[i] = 0xFF // all empty
	}
	// Overwrite slot landSlotStart with a type-2 (rock) header
	off := landSlotStart * objectHeaderSize
	h := makeObjectHeader(2, "ROCK1") // type=2, not land
	copy(buf[off:], h[:])

	order := parseLandObjectOrder(buf)
	if order[0] != "" {
		t.Errorf("order[0] = %q, want empty (wrong type should be skipped)", order[0])
	}
}

func TestParseLandObjectOrder_truncated_blob(t *testing.T) {
	// Blob too short to cover any land slots → all empty, no panic.
	blob := make([]byte, 10)
	order := parseLandObjectOrder(blob)

	if len(order) != landSlotCount {
		t.Fatalf("len(order) = %d, want %d", len(order), landSlotCount)
	}
	for i, name := range order {
		if name != "" {
			t.Errorf("order[%d] = %q, want empty (truncated blob)", i, name)
		}
	}
}

func TestParseLandObjectOrder_name_trimming(t *testing.T) {
	// Name shorter than 8 bytes; trailing 0xFF bytes must be trimmed.
	names := []string{"AB"}
	blob := makeRequiredObjectsBlob(names)
	order := parseLandObjectOrder(blob)

	if order[0] != "AB" {
		t.Errorf("order[0] = %q, want %q", order[0], "AB")
	}
}

func TestParseLandObjectOrder_space_padded_names(t *testing.T) {
	// On-disk ObjectHeader names are space-padded to 8 bytes (e.g. "GRASS1  ").
	// The parser must trim trailing spaces so the name matches the Objects map key.
	totalSlots := landSlotStart + landSlotCount
	buf := make([]byte, totalSlots*objectHeaderSize)
	for i := range buf {
		buf[i] = 0xFF
	}

	// Write a land header with space-padded name "GRASS1  " at slot 0
	off := landSlotStart * objectHeaderSize
	buf[off] = objectTypeLand // flags low byte = type
	buf[off+1] = 0
	buf[off+2] = 0
	buf[off+3] = 0
	copy(buf[off+4:off+12], []byte("GRASS1  ")) // space-padded

	order := parseLandObjectOrder(buf)
	if order[0] != "GRASS1" {
		t.Errorf("order[0] = %q, want %q (trailing spaces not trimmed)", order[0], "GRASS1")
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
