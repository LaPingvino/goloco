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

	// WallObjectOrder is the list of wall-object names in wall-slot order,
	// as specified by the RequiredObjects chunk. Slots 236-267.
	WallObjectOrder []string

	// TrackObjectOrder is the list of track-object names in track-slot order.
	// Slots 337-344 (8 slots).
	TrackObjectOrder []string

	// RoadObjectOrder is the list of road-object names in road-slot order.
	// Slots 365-372 (8 slots).
	RoadObjectOrder []string

	// Entities contains vehicle entities parsed from the GameState chunk of
	// an .SV5 saved-game file.  Only VehicleHead and VehicleBody types are
	// included — enough for world-space rendering.  Nil for .SC5 scenarios.
	//
	// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  GameState::entities
	Entities []VehicleEntity
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
	Trees      []TreeElement
	Buildings  []BuildingElement
	Walls      []WallElement
	Tracks     []TrackElement
	Roads      []RoadElement
	Stations   []StationElement
	Signals    []SignalElement
	Industries []IndustryElement
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

// WallElement represents a wall/fence on a tile edge.
//
// OpenLoco reference: src/OpenLoco/src/Map/WallElement.h
type WallElement struct {
	WallObjectID   uint8 // byte 4: which WallObject (index into RequiredObjects wall slots)
	Rotation       uint8 // byte 0 bits [1:0]
	EdgeSlope      uint8 // byte 0 bits [7:6]: 0=none, 1=upwards, 2=downwards
	PrimaryColour  uint8 // byte 6 bits [4:0]
	BaseZ          uint8 // byte 2: height in SmallZ
	ClearZ         uint8 // byte 3: clear height
}

// TrackElement represents a railway track section on a tile.
//
// OpenLoco reference: src/OpenLoco/src/Map/TrackElement.h
type TrackElement struct {
	TrackObjectID uint8 // byte 5 bits [7:4]: index into RequiredObjects track slots
	TrackID       uint8 // byte 4 bits [5:0]: track piece type (0-63)
	Rotation      uint8 // byte 0 bits [1:0]: 0-3
	SequenceIndex uint8 // byte 5 bits [3:0]: multi-tile piece index
	HasBridge     bool  // byte 4 bit 7
	HasSignal     bool  // byte 0 bit 6
	HasStation    bool  // byte 0 bit 7
	Owner         uint8 // byte 7 bits [3:0]
	Mods          uint8 // byte 7 bits [7:4]: modifier bitmask
	BaseZ         uint8 // byte 2
	ClearZ        uint8 // byte 3
}

// RoadElement represents a road section on a tile.
//
// OpenLoco reference: src/OpenLoco/src/Map/RoadElement.h
type RoadElement struct {
	RoadObjectID  uint8 // byte 5 bits [7:4]: index into RequiredObjects road slots
	RoadID        uint8 // byte 4 bits [3:0]: road piece type (0-15)
	Rotation      uint8 // byte 0 bits [1:0]: 0-3
	SequenceIndex uint8 // byte 5 bits [1:0]: multi-tile piece index
	HasBridge     bool  // byte 4 bit 7
	Owner         uint8 // byte 7 bits [3:0]
	Mods          uint8 // byte 7 bits [7:6]: modifier bitmask
	BaseZ         uint8 // byte 2
	ClearZ        uint8 // byte 3
}

// StationElement represents a station tile (rail, road, airport, or dock).
//
// OpenLoco reference: src/OpenLoco/src/Map/StationElement.h
type StationElement struct {
	ObjectID     uint8  // byte 5 bits [4:0]: StationObject index
	Rotation     uint8  // byte 0 bits [1:0]
	SequenceIndex uint8 // byte 0 bits [7:6]: multi-tile position
	StationID    uint16 // bytes 6-7 bits [9:0]: which station instance
	BuildingType uint8  // bytes 6-7 bits [15:10]: airport building type
	Owner        uint8  // byte 4 bits [3:0]
	BaseZ        uint8  // byte 2
	ClearZ       uint8  // byte 3
}

// SignalElement represents a railway signal on a tile.
//
// OpenLoco reference: src/OpenLoco/src/Map/SignalElement.h
type SignalElement struct {
	SignalObjectID uint8 // byte 4 bits [3:0]
	Rotation       uint8 // byte 0 bits [1:0]
	LeftFrame      uint8 // byte 5 bits [3:0]: animation frame for left signal
	RightFrame     uint8 // byte 7 bits [3:0]: animation frame for right signal
	HasLeftSignal  bool  // byte 4 bit 7
	HasRightSignal bool  // byte 6 bit 7
	BaseZ          uint8 // byte 2
	ClearZ         uint8 // byte 3
}

// IndustryElement represents an industry building tile.
//
// OpenLoco reference: src/OpenLoco/src/Map/IndustryElement.h
type IndustryElement struct {
	IndustryID   uint8 // byte 4: industry index
	BuildingType uint8 // bytes 6-7 bits [10:6]
	Rotation     uint8 // byte 0 bits [1:0]
	IsConstructed bool // byte 0 bit 7
	BaseZ        uint8 // byte 2
	ClearZ       uint8 // byte 3
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
