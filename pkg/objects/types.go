package objects

// ObjectType identifies the type of object in a DAT file
type ObjectType uint8

const (
	ObjectTypeInterfaceSkin ObjectType = iota
	ObjectTypeSound
	ObjectTypeCurrency
	ObjectTypeSteam
	ObjectTypeCliffEdge
	ObjectTypeWater
	ObjectTypeLand
	ObjectTypeTownNames
	ObjectTypeCargo
	ObjectTypeWall
	ObjectTypeTrackSignal
	ObjectTypeLevelCrossing
	ObjectTypeStreetLight
	ObjectTypeTunnel
	ObjectTypeBridge
	ObjectTypeTrainStation
	ObjectTypeTrackExtra
	ObjectTypeTrack
	ObjectTypeRoadStation
	ObjectTypeRoadExtra
	ObjectTypeRoad
	ObjectTypeAirport
	ObjectTypeDock
	ObjectTypeVehicle
	ObjectTypeTree
	ObjectTypeSnow
	ObjectTypeClimate
	ObjectTypeHillShapes
	ObjectTypeBuilding
	ObjectTypeScaffolding
	ObjectTypeIndustry
	ObjectTypeRegion
	ObjectTypeCompetitor
	ObjectTypeScenarioText
)

// MaxObjectTypes is the number of distinct object types
const MaxObjectTypes = 34

// ObjectTypeName returns the name of an object type
func (t ObjectType) String() string {
	names := []string{
		"InterfaceSkin", "Sound", "Currency", "Steam", "CliffEdge",
		"Water", "Land", "TownNames", "Cargo", "Wall",
		"TrackSignal", "LevelCrossing", "StreetLight", "Tunnel", "Bridge",
		"TrainStation", "TrackExtra", "Track", "RoadStation", "RoadExtra",
		"Road", "Airport", "Dock", "Vehicle", "Tree",
		"Snow", "Climate", "HillShapes", "Building", "Scaffolding",
		"Industry", "Region", "Competitor", "ScenarioText",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}

// SourceGame identifies the origin of an object
type SourceGame uint8

const (
	SourceGameCustom  SourceGame = 0
	SourceGameData    SourceGame = 1
	SourceGameVanilla SourceGame = 2
	SourceGameOpenLoco SourceGame = 3
)

// TransportMode defines how a vehicle moves
type TransportMode uint8

const (
	TransportModeRail  TransportMode = 0
	TransportModeRoad  TransportMode = 1
	TransportModeAir   TransportMode = 2
	TransportModeWater TransportMode = 3
)

func (m TransportMode) String() string {
	switch m {
	case TransportModeRail:
		return "Rail"
	case TransportModeRoad:
		return "Road"
	case TransportModeAir:
		return "Air"
	case TransportModeWater:
		return "Water"
	}
	return "Unknown"
}

// VehicleType defines the specific type of vehicle
type VehicleType uint8

const (
	VehicleTypeTrain    VehicleType = 0
	VehicleTypeBus      VehicleType = 1
	VehicleTypeTruck    VehicleType = 2
	VehicleTypeTram     VehicleType = 3
	VehicleTypeAircraft VehicleType = 4
	VehicleTypeShip     VehicleType = 5
)

func (t VehicleType) String() string {
	switch t {
	case VehicleTypeTrain:
		return "Train"
	case VehicleTypeBus:
		return "Bus"
	case VehicleTypeTruck:
		return "Truck"
	case VehicleTypeTram:
		return "Tram"
	case VehicleTypeAircraft:
		return "Aircraft"
	case VehicleTypeShip:
		return "Ship"
	}
	return "Unknown"
}

// SawyerEncoding defines the compression type used in DAT files
type SawyerEncoding uint8

const (
	SawyerEncodingUncompressed    SawyerEncoding = 0
	SawyerEncodingRunLengthSingle SawyerEncoding = 1
	SawyerEncodingRunLengthMulti  SawyerEncoding = 2
	SawyerEncodingRotate          SawyerEncoding = 3
)
