package scenario

import (
	"fmt"
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

	// PackedObjectsRaw holds fully decoded (header + decompressed data) for
	// each packed object embedded in the scenario file.
	PackedObjectsRaw []PackedObjectRaw

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

	// TrainSignalObjectOrder is the list of train-signal-object names in slot order.
	// Slots 268-283 (16 slots, ObjectType::trackSignal).
	TrainSignalObjectOrder []string

	// BridgeObjectOrder is the list of bridge-object names in slot order.
	// Slots 305-312 (8 slots, ObjectType::bridge).
	BridgeObjectOrder []string

	// TrainStationObjectOrder is the list of train-station-object names in slot order.
	// Slots 313-328 (16 slots).
	TrainStationObjectOrder []string

	// RoadStationObjectOrder is the list of road-station-object names in slot order.
	// Slots 345-360 (16 slots).
	RoadStationObjectOrder []string

	// TrackObjectOrder is the list of track-object names in track-slot order.
	// Slots 337-344 (8 slots).
	TrackObjectOrder []string

	// RoadObjectOrder is the list of road-object names in road-slot order.
	// Slots 365-372 (8 slots).
	RoadObjectOrder []string

	// LevelCrossingObjectOrder is the list of level-crossing-object names in slot order.
	// Slots 284-287 (4 slots, ObjectType::levelCrossing).
	LevelCrossingObjectOrder []string

	// VehicleObjectOrder is the list of vehicle-object names in slot order.
	// Slots 389-612 (224 slots, ObjectType::vehicle).
	VehicleObjectOrder []string

	// Entities contains vehicle entities parsed from the GameState chunk of
	// an .SV5 saved-game file.  Only VehicleHead and VehicleBody types are
	// included — enough for world-space rendering.  Nil for .SC5 scenarios.
	//
	// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.h  GameState::entities
	CargoObjectOrder []string  // cargo slot -> DAT name (for objective cargo names)
	Objective        Objective // win condition (valid when HasObjective)
	HasObjective     bool
	Entities         []VehicleEntity
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
	TerrainIndex uint8 // byte 6 bits [4:0] — LandObject slot (0-31)
	GrowthStage  uint8 // byte 6 bits [7:5] — growth/season stage (0-7); used in sprite variation formula
	Variation    uint8 // byte 7 — alternate terrain style; non-zero in snowy/special scenarios
	IsIndustrial bool  // type byte bit 7 — surface belongs to an industry (farm fields etc.)
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
	WallObjectID  uint8 // byte 4: which WallObject (index into RequiredObjects wall slots)
	Rotation      uint8 // byte 0 bits [1:0]
	EdgeSlope     uint8 // byte 0 bits [7:6]: 0=none, 1=upwards, 2=downwards
	PrimaryColour uint8 // byte 6 bits [4:0]
	BaseZ         uint8 // byte 2: height in SmallZ
	ClearZ        uint8 // byte 3: clear height
}

// TrackElement represents a railway track section on a tile.
//
// OpenLoco reference: src/OpenLoco/src/Map/TrackElement.h
// Objective is the scenario win condition, from GameState offset 0x418.
// OpenLoco reference: include/OpenLoco/Scenario/ScenarioObjective.h
type Objective struct {
	Type                 uint8 // 0 companyValue, 1 vehicleProfit, 2 performanceIndex, 3 cargoDelivery
	Flags                uint8 // bit0 beTopCompany, bit1 topThree, bit2 withinTimeLimit
	CompanyValue         uint32
	MonthlyVehicleProfit uint32
	PerformanceIndex     uint8 // x10 = percent
	DeliveredCargoType   uint8 // cargo object slot
	DeliveredCargoAmount uint32
	TimeLimitYears       uint8
}

// DescribeWithCargo renders the objective using a cargo name when known.
func (o Objective) DescribeWithCargo(cargoName string) string {
	if o.Type == 3 && cargoName != "" {
		s := fmt.Sprintf("Deliver %d units of %s", o.DeliveredCargoAmount, cargoName)
		if o.Flags&0x04 != 0 {
			s += fmt.Sprintf(" within %d years", o.TimeLimitYears)
		}
		return s
	}
	return o.Describe()
}

// Describe renders the objective as a human-readable sentence.
func (o Objective) Describe() string {
	s := ""
	switch o.Type {
	case 0:
		s = fmt.Sprintf("Achieve a company value of £%d", o.CompanyValue)
	case 1:
		s = fmt.Sprintf("Achieve a monthly vehicle profit of £%d", o.MonthlyVehicleProfit)
	case 2:
		s = fmt.Sprintf("Achieve a performance index of %d%%", o.PerformanceIndex)
	case 3:
		s = fmt.Sprintf("Deliver %d units of cargo type %d", o.DeliveredCargoAmount, o.DeliveredCargoType)
	default:
		return "Unknown objective"
	}
	if o.Flags&0x04 != 0 {
		s += fmt.Sprintf(" within %d years", o.TimeLimitYears)
	}
	if o.Flags&0x01 != 0 {
		s += " (be the top company)"
	}
	if o.Flags&0x02 != 0 {
		s += " (be in the top three)"
	}
	return s
}

type TrackElement struct {
	TrackObjectID uint8 // byte 5 bits [7:4]: index into RequiredObjects track slots
	TrackID       uint8 // byte 4 bits [5:0]: track piece type (0-63)
	Rotation      uint8 // byte 0 bits [1:0]: 0-3
	SequenceIndex uint8 // byte 5 bits [3:0]: multi-tile piece index
	HasBridge     bool  // byte 4 bit 7
	BridgeID      uint8 // byte 6 bits [7:5]: bridge object slot index (only valid if HasBridge)
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
	RoadObjectID          uint8 // byte 5 bits [7:4]: index into RequiredObjects road slots
	RoadID                uint8 // byte 4 bits [3:0]: road piece type (0-15)
	Rotation              uint8 // byte 0 bits [1:0]: 0-3
	SequenceIndex         uint8 // byte 5 bits [1:0]: multi-tile piece index
	HasBridge             bool  // byte 4 bit 7
	BridgeID              uint8 // byte 6 bits [7:5]: bridge object slot index (only valid if HasBridge)
	HasLevelCrossing      bool  // byte 7 bit 5
	LevelCrossingObjectID uint8 // byte 5 bits [3:2]: index into RequiredObjects level-crossing slots
	AnimFrame             uint8 // byte 6 bits [3:0]: level crossing animation state
	Owner                 uint8 // byte 7 bits [3:0]
	Mods                  uint8 // byte 7 bits [7:6]: modifier bitmask
	BaseZ                 uint8 // byte 2
	ClearZ                uint8 // byte 3
}

// StationElement represents a station tile (rail, road, airport, or dock).
//
// OpenLoco reference: src/OpenLoco/src/Map/StationElement.h
type StationElement struct {
	StationType   uint8  // byte 5 bits [7:5]: 0=train, 1=road, 2=airport, 3=docks
	ObjectID      uint8  // byte 5 bits [4:0]: StationObject index
	Rotation      uint8  // byte 0 bits [1:0]
	SequenceIndex uint8  // byte 0 bits [7:6]: multi-tile position
	StationID     uint16 // bytes 6-7 bits [9:0]: which station instance
	BuildingType  uint8  // bytes 6-7 bits [15:10]: airport building type
	Owner         uint8  // byte 4 bits [3:0]
	BaseZ         uint8  // byte 2
	ClearZ        uint8  // byte 3
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
	IndustryID    uint8 // byte 4: industry index
	BuildingType  uint8 // bytes 6-7 bits [10:6]
	Rotation      uint8 // byte 0 bits [1:0]
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

// ScenarioCategory is the difficulty level stored in ScenarioOptions.difficulty (offset 0x01).
//
// OpenLoco reference: src/OpenLoco/src/S5/S5Options.h  Options::difficulty
type ScenarioCategory uint8

const (
	CategoryBeginner    ScenarioCategory = 0
	CategoryEasy        ScenarioCategory = 1
	CategoryMedium      ScenarioCategory = 2
	CategoryChallenging ScenarioCategory = 3
	CategoryExpert      ScenarioCategory = 4
)

func (c ScenarioCategory) String() string {
	switch c {
	case CategoryBeginner:
		return "Beginner"
	case CategoryEasy:
		return "Easy"
	case CategoryMedium:
		return "Medium"
	case CategoryChallenging:
		return "Challenging"
	case CategoryExpert:
		return "Expert"
	default:
		return "Unknown"
	}
}

// ScenarioFlagLandscapeGenerationDone is bit 0 of ScenarioOptions.scenarioFlags (offset 0x06).
// When set, the map was hand-crafted and a 128×128 preview image is stored.
// When clear, the map is generated from stored parameters (no preview).
//
// OpenLoco reference: src/OpenLoco/src/S5/S5Options.h  ScenarioFlags::landscapeGenerationDone
const ScenarioFlagLandscapeGenerationDone uint16 = 0x0001

// PreviewSize is the side length of the scenario preview image in pixels.
const PreviewSize = 128

// Options contains scenario configuration
type Options struct {
	// Parsed from ScenarioOptions chunk (S5::Options)
	Name        string
	Description string
	Category    ScenarioCategory
	Flags       uint16 // ScenarioFlags bitmask (bit 0 = landscapeGenerationDone)
	HasPreview  bool   // true when landscapeGenerationDone flag is set
	// Preview is the 128×128 palette-indexed preview image.
	// Each byte is a palette colour index; zero-filled if no preview.
	Preview        [PreviewSize * PreviewSize]byte
	StartYear      uint16
	MaxCompetitors uint8
	ObjectiveType  uint8

	Objective          Objective
	HasObjective       bool
	ObjectiveCargoName string // DAT name of the delivery cargo (e.g. PASS)

	// Generation parameters (used when landscapeGenerationDone is clear)
	MinLandHeight   uint8 // base height (0-15)
	TopographyStyle uint8 // 0=flat,1=smallHills,2=mountains,3=halfMtHills,4=halfMtFlat
	HillDensity     uint8 // 0-100

	// Legacy / unused fields kept for compatibility
	EndYear    uint16
	StartMoney int64
	MaxLoans   int64
}

// PackedObject represents a reference to a DAT object file
type PackedObject struct {
	Name     [8]byte
	Checksum uint32
	Type     ObjectType
}

// PackedObjectRaw stores a packed object's header and decompressed data as
// embedded directly in an SC5/SV5 file.  These objects are loaded into the
// ObjectManager before reordering so that scenario-specific assets (buildings,
// vehicles, etc.) that are not present in ObjData are still available.
type PackedObjectRaw struct {
	HeaderBytes [16]byte // raw 16-byte ObjectHeader (flags, name, checksum)
	Data        []byte   // decompressed object payload
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
