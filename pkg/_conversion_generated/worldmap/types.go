// Package worldmap provides map and tile management for goloco.
// Translated from OpenLoco/src/Map
package worldmap

// Map dimensions
const (
	MapColumns   = 384
	MapRows      = 384
	TileSize     = 32     // World units per tile
	SmallZStep   = 4      // Height units per small Z step
	MicroZStep   = 16     // Height units per micro Z step
	TileElemSize = 8      // Bytes per tile element
)

// SmallZ is a compressed height value (4 units per step)
type SmallZ = uint8

// MicroZ is a more precise height value (16 units per step)
type MicroZ = uint8

// Coord is a world coordinate
type Coord = int16

// TilePos2 represents a tile position (0-383, 0-383)
type TilePos2 struct {
	X, Y uint16
}

// Pos2 represents a world position in 2D
type Pos2 struct {
	X, Y Coord
}

// Pos3 represents a world position in 3D
type Pos3 struct {
	X, Y, Z Coord
}

// NewTilePos2 creates a TilePos2 from world coordinates
func NewTilePos2(pos Pos2) TilePos2 {
	return TilePos2{
		X: uint16(pos.X / TileSize),
		Y: uint16(pos.Y / TileSize),
	}
}

// ToPos2 converts tile position to world position
func (t TilePos2) ToPos2() Pos2 {
	return Pos2{
		X: Coord(t.X * TileSize),
		Y: Coord(t.Y * TileSize),
	}
}

// IndustryId identifies an industry on the map
type IndustryId uint8

// CompanyId identifies a company
type CompanyId uint8

// EntityId identifies an entity
type EntityId uint16

// ObjectId identifies an object type
type ObjectId uint8
