package scenario

import (
	"github.com/LaPingvino/goloco/pkg/assets"
)

// Scenario represents a loaded scenario or saved game
type Scenario struct {
	// Header information
	Header   assets.S5Header
	FilePath string

	// Map data
	MapWidth  int
	MapHeight int
	Tiles     [][]Tile

	// Scenario metadata
	Options Options

	// Packed objects (DAT file references)
	PackedObjects []PackedObject

	// LandObjectOrder is the list of land-object names in terrain-slot order,
	// as specified by the RequiredObjects chunk.  Slot 0 = terrain index 0, etc.
	// Empty entries (0xFF-filled headers) are stored as "".
	LandObjectOrder []string

	// TreeObjectOrder is the list of tree-object names in tree-slot order,
	// as specified by the RequiredObjects chunk.  Slot 0 = treeObjectId 0, etc.
	TreeObjectOrder []string

	// BuildingObjectOrder is the list of building-object names in building-slot order.
	BuildingObjectOrder []string
}

// GetTile returns tile at the given coordinates
// (implementation moved to loader.go)

// IsTitleSequence returns true if this is a title sequence file
// (implementation moved to loader.go)

// Tile represents a single map tile
type Tile struct {
	Surface      SurfaceType
	Height       uint8
	Water        uint8
	TerrainIndex uint8 // raw 5-bit terrain index from the surface element (LandObject slot)
	Slope        uint8 // slope byte from surface element byte 4
	Ownership    uint8
	BuildingID   uint16
	TrackType    uint8
	Flags        uint8

	// Non-surface elements on this tile
	Trees     []TreeElement
	Buildings []BuildingElement
}

// TreeElement represents a tree on a tile.
//
// OpenLoco reference: src/OpenLoco/src/Map/TreeElement.h
type TreeElement struct {
	TreeObjectID uint8 // byte 4: which TreeObject (index into RequiredObjects tree slots)
	Rotation     uint8 // byte 0 bits [1:0]
	Quadrant     uint8 // byte 0 bits [7:6]
	Growth       uint8 // byte 5 bits [3:0]: tree size 0-15
	Season       uint8 // byte 7 bits [7:3]
	Colour       uint8 // byte 6 bits [4:0]
	HasSnow      bool  // byte 6 bit 6
	BaseZ        uint8 // byte 2: height in SmallZ
	ClearZ       uint8 // byte 3: clear height
	Unk7l        uint8 // byte 7 bits [2:0]: seasonal transition state
}

// BuildingElement represents a building on a tile.
//
// OpenLoco reference: src/OpenLoco/src/Map/BuildingElement.h
type BuildingElement struct {
	ObjectID      uint8 // byte 4: which BuildingObject
	Rotation      uint8 // byte 0 bits [1:0]
	SequenceIndex uint8 // byte 5 bits [1:0]: multi-tile position
	Variation     uint8 // building variant
	Colour        uint8 // company/building color
	IsConstructed bool  // byte 0 bit 7
	BaseZ         uint8 // byte 2
	ClearZ        uint8 // byte 3
}

// SurfaceType represents type of terrain surface
type SurfaceType uint8

const (
	SurfaceGrass SurfaceType = iota
	SurfaceSand
	SurfaceDirt
	SurfaceRock
	SurfaceSnow
	SurfaceWater
)

// Options contains scenario configuration
type Options struct {
	Name        string
	Description string
	StartYear   uint16
	EndYear     uint16
	StartMoney  int64
	MaxLoans    int64
}

// PackedObject represents a reference to a DAT object file
type PackedObject struct {
	Name     [8]byte
	Checksum uint32
	Type     ObjectType
}

// ObjectType identifies category of a DAT object
type ObjectType uint8

const (
	ObjectTypeInterface ObjectType = iota
	ObjectTypeSound
	ObjectTypeCurrency
	ObjectTypeSteam
	ObjectTypeRockObj
	ObjectTypeWaterObj
	ObjectTypeLand
	ObjectTypeTownNames
	ObjectTypeCargo
	ObjectTypeWall
	ObjectTypeTrackSignal
	ObjectTypeLevelCrossing
	ObjectTypeStreetLight
	ObjectTypeTunnel
	ObjectTypeBridge
	ObjectTypeTrackStation
	ObjectTypeTrackMod
	ObjectTypeTrack
	ObjectTypeRoadStation
	ObjectTypeRoadMod
	ObjectTypeRoad
	ObjectTypeAirport
	ObjectTypeDock
	ObjectTypeVehicle
	ObjectTypeTree
	ObjectTypeSnow
	ObjectTypeClimate
	ObjectTypeHillShapes
	ObjectTypeBuilding
	ObjectTypeScaffold
	ObjectTypeIndustry
	ObjectTypeRegion
	ObjectTypeCompetitor
	ObjectTypeScenarioText
)

// ChunkType identifies type of data chunk in an S5 file
type ChunkType uint8

const (
	ChunkTypeFlags ChunkType = iota
	ChunkTypePackedObjects
	ChunkTypeGameState
	ChunkTypeTileElements
	// Add more as needed
)
