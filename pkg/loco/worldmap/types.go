package worldmap

// World map type definitions

// coord_t is the coordinate type used in the game
type coord_t = int16

// TilePos2 represents a tile position (in tile coordinates)
type TilePos2 struct {
	X int16
	Y int16
}

// Pos2 represents a position in world coordinates (pixels/subpixels)
type Pos2 struct {
	X int16
	Y int16
}

// Pos3 represents a 3D position in world coordinates
type Pos3 struct {
	X int16
	Y int16
	Z int16
}

// TileHeight represents land and water height at a tile
type TileHeight struct {
	LandHeight  coord_t
	WaterHeight coord_t
}

// GetEffectiveHeight returns water height if present, otherwise land height
func (th *TileHeight) GetEffectiveHeight() coord_t {
	if th.WaterHeight == 0 {
		return th.LandHeight
	}
	return th.WaterHeight
}

// Tile offsets for isometric rendering
var TileOffsets = [4]Pos2{
	{X: 0, Y: 0},
	{X: 0, Y: 32},
	{X: 32, Y: 32},
	{X: 32, Y: 0},
}

// Rotation offsets for different rotation angles
var RotationOffset = [16]Pos2{
	{X: -32, Y: 0},
	{X: 0, Y: 32},
	{X: 32, Y: 0},
	{X: 0, Y: -32},
	{X: -32, Y: 0},
	{X: 0, Y: 32},
	{X: 32, Y: 0},
	{X: 0, Y: -32},
	{X: -32, Y: 0},
	{X: 0, Y: 32},
	{X: 32, Y: 0},
	{X: 0, Y: -32},
	{X: -32, Y: 32},
	{X: 32, Y: 32},
	{X: 32, Y: -32},
	{X: -32, Y: -32},
}

// ReverseRotation maps rotations to their reverse
var ReverseRotation = [16]uint8{
	2, 3, 0, 1,
	10, 11, 8, 9,
	6, 7, 4, 5,
	14, 15, 12, 13,
}
