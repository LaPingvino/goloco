package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstdint>
// #include <cstring>
// #include <limits>
// #include <string_view>
// namespace OpenLoco
type ObjectType int

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
const MaxObjectTypes int = 34
const CargoTypeNull uint8 = 0xFF
type SourceGame int

const (
// // See below note for vanilla.
// // If object SourceGame is custom then object header will be fully compared
// // when looking for matches (i.e. takes into account object header checksum).
// // This means that non-custom objects might have different versions but still
// // be considered as the same object.
	Custom SourceGame = 0
	Data SourceGame = 1
// // Most custom objects set this, so can't be trusted to be only on vanilla.
// // Use the isVanilla() function to actually check for custom as that does
// // a lookup against the vanilla object list
	Vanilla SourceGame = 2
	OpenLoco SourceGame = 3
)
type ObjectHeader struct {
	Flags uint32
// char name[8];
	Checksum uint32
// std::string_view getName() const
// return std::string_view(name, sizeof(name));
}
const ObjectHeaderFuzzyFlagsMask uint32 = 0xFFFE0000
// func GetSourceGame() SourceGame
// return static_cast<SourceGame>((flags >> 6) & 0x3);
// func GetType() ObjectType
// return static_cast<ObjectType>(flags & 0x3F);
// func GetFuzzyFlags() uint32
// return flags & kFuzzyFlagsMask;
// func IsCustom() bool
// func GetSourceGame() return
// func IsEmpty() bool
// // The original game would check whether an object was part of the base game by means of the sourceGame attribute.
// // As many custom objects were being based on base game objects, most custom objects do not set this attribute
// // correctly. Therefore those objects would not be packed by Locomotion. We change this behaviour by explicitly
// // checking objects against a list of vanilla objects instead. The original logic is left in the isCustom method.
// func IsVanilla() bool
// bool operator==(const ObjectHeader& rhs) const
// // Some vanilla objects reference other vanilla objects using a
// // ObjectHeader that is set as custom. To handle those we need
// // to check both the LHS and the RHS and only if both are Custom
// // do the full check.
// if (isCustom() && rhs.isCustom())
// return std::memcmp(this, &rhs, sizeof(ObjectHeader)) == 0;
// else
// func GetType() return
// static_assert(sizeof(ObjectHeader) == 0x10);
const EmptyObjectHeader ObjectHeader = ObjectHeader{ 0xFFFFFFFFU, { '\xFF', '\xFF', '\xFF', '\xFF', '\xFF', '\xFF', '\xFF', '\xFF' }, 0xFFFFFFFFU }
// constexpr bool ObjectHeader::isEmpty() const
// // No point checking the name as its already invalid if flags is all 0xFFFFFFFFU
// return this->flags == kEmptyObjectHeader.flags && this->checksum == kEmptyObjectHeader.checksum;
// /**
// * Represents an index / ID of a specific object type.
// */
type LoadedObjectId = uint16
// /**
// * Represents an undefined index / ID for a specific object type.
// */
const NullObjectId LoadedObjectId = std.numeric_limits<LoadedObjectId>.max()
type LoadedObjectHandle struct {
	Type ObjectType
	Id LoadedObjectId
}
// func ObjectCreateIdentifierName(dst byte, header ObjectHeader) 
