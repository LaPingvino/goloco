package objects

// Object system types extracted from OpenLoco
// Clean, compilable definitions for all game object types

// ObjectType represents the type of a game object
type ObjectType uint8

const (
	InterfaceSkin ObjectType = iota
	Sound
	Currency
	Steam
	CliffEdge
	Water
	Land
	TownNames
	Cargo
	Wall
	TrackSignal
	LevelCrossing
	StreetLight
	Tunnel
	Bridge
	TrainStation
	TrackExtra
	Track
	RoadStation
	RoadExtra
	Road
	Airport
	Dock
	Vehicle
	Tree
	Snow
	Climate
	HillShapes
	Building
	Scaffolding
	Industry
	Region
	Competitor
	ScenarioText
)

const (
	MaxObjectTypes = 34
	CargoTypeNull  = 0xFF
)

// SourceGame indicates where an object originated
type SourceGame uint8

const (
	Custom   SourceGame = 0
	Data     SourceGame = 1
	Vanilla  SourceGame = 2
	OpenLoco SourceGame = 3
)

// ObjectHeader is the header for all object files
type ObjectHeader struct {
	Flags    uint32
	Name     [8]byte
	Checksum uint32
}

const ObjectHeaderFuzzyFlagsMask uint32 = 0xFFFE0000

// GetName returns the object name as a string
func (oh *ObjectHeader) GetName() string {
	// Find null terminator
	for i, b := range oh.Name {
		if b == 0 {
			return string(oh.Name[:i])
		}
	}
	return string(oh.Name[:])
}

// GetSourceGame extracts the source game from flags
func (oh *ObjectHeader) GetSourceGame() SourceGame {
	return SourceGame((oh.Flags >> 6) & 0x3)
}

// GetType extracts the object type from flags
func (oh *ObjectHeader) GetType() ObjectType {
	return ObjectType(oh.Flags & 0x3F)
}

// GetFuzzyFlags returns flags used for fuzzy matching
func (oh *ObjectHeader) GetFuzzyFlags() uint32 {
	return oh.Flags & ObjectHeaderFuzzyFlagsMask
}

// IsCustom returns true if this is a custom object
func (oh *ObjectHeader) IsCustom() bool {
	return oh.GetSourceGame() == Custom
}

// IsEmpty returns true if this is an empty object header
func (oh *ObjectHeader) IsEmpty() bool {
	return oh.Flags == 0xFFFFFFFF && oh.Checksum == 0xFFFFFFFF
}

// LoadedObjectId represents a loaded object's ID
type LoadedObjectId = uint32

const NullObjectId LoadedObjectId = 0xFFFFFFFF

// LoadedObjectHandle references a loaded object
type LoadedObjectHandle struct {
	Type ObjectType
	Id   LoadedObjectId
}
