package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "BuildingCommon.h"
// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type AirportMovementNodeFlags int

const (
func init() {
	None AirportMovementNodeFlags = 0
	Terminal AirportMovementNodeFlags = 1 << 0
	TakeoffEnd AirportMovementNodeFlags = 1 << 1
	Flag2 AirportMovementNodeFlags = 1 << 2
	Taxiing AirportMovementNodeFlags = 1 << 3
	InFlight AirportMovementNodeFlags = 1 << 4
	HeliTakeoffBegin AirportMovementNodeFlags = 1 << 5
	TakeoffBegin AirportMovementNodeFlags = 1 << 6
	HeliTakeoffEnd AirportMovementNodeFlags = 1 << 7
	Touchdown AirportMovementNodeFlags = 1 << 8
}

)
// OPENLOCO_ENABLE_ENUM_OPERATORS(AirportMovementNodeFlags);
type AirportObjectFlags int

func init() {
	None AirportObjectFlags = 0
	HasShadows AirportObjectFlags = 1 << 0
	IsHelipad AirportObjectFlags = 1 << 1
	AcceptsLightPlanes AirportObjectFlags = 1 << 2
	AcceptsHeavyPlanes AirportObjectFlags = 1 << 3
	AcceptsHelicopter AirportObjectFlags = 1 << 4

// OPENLOCO_ENABLE_ENUM_OPERATORS(AirportObjectFlags);
type AirportBuilding struct {
	Index uint8
	Rotation uint8
	X int8
	Y int8
type AirportObject struct {

type AirportObjectMovementNode struct {
	X int16
	Y int16
	Z int16
	Flags AirportMovementNodeFlags
	// method: constexpr bool hasFlags(AirportMovementNodeFlags flagsToTest) const
// return (flags & flagsToTest) != AirportMovementNodeFlags::none;
const AirportObjectObjectType any = ObjectType.airport
type MovementEdge struct {
	Var_00 uint8
	CurNode uint8
	NextNode uint8
	Var_03 uint8
	MustBeClearEdges uint32
	AtLeastOneClearEdges uint32
// orphan member: StringId name;
// orphan member: int16 buildCostFactor; // 0x02
// orphan member: int16 sellCostFactor;  // 0x04
// orphan member: uint8 costIndex;       // 0x06
// orphan member: uint8 var_07;
// orphan member: uint32 image;                            // 0x08
// orphan member: uint32 buildingImage;                    // 0x0C
// orphan member: AirportObjectFlags flags;                  // 0x10
// orphan member: uint8 numSpriteSets;                     // 0x12
// orphan member: uint8 numTiles;                          // 0x13
// orphan member: uint32 buildingPartHeightsOffset;        // 0x14
// orphan member: uint32 buildingPartAnimationsOffset;     // 0x18
// uint32 buildingVariationPartOffsets[32]; // 0x1C
// orphan member: uint32 buildingPositionsOffset;          // 0x9C
// orphan member: uint32 largeTiles;                       // 0xA0
// orphan member: int8_t minX;                               // 0xA4
// orphan member: int8_t minY;                               // 0xA5
// orphan member: int8_t maxX;                               // 0xA6
// orphan member: int8_t maxY;                               // 0xA7
// orphan member: uint16 designedYear;                     // 0xA8
// orphan member: uint16 obsoleteYear;                     // 0xAA
// orphan member: uint8 numMovementNodes;                  // 0xAC
// orphan member: uint8 numMovementEdges;                  // 0xAD
// orphan member: uint32 movementNodesOffset;              // 0xAE
// orphan member: uint32 movementEdgesOffset;              // 0xB2
// orphan member: uint32 var_B6;
// func DrawPreviewImage(drawingCtx Gfx::DrawingContext, x int16, y int16) 
// func DrawDescription(drawingCtx Gfx::DrawingContext, x int16, y int16, width [[maybe_unused]] int16) 
// func Validate() bool
// func Load(handle LoadedObjectHandle, data any /* []<std::byte> */ , ObjectManager::DependentObjects) 
// func Unload() 
// std::pair<World::TilePos2, World::TilePos2> getAirportExtents(const World::TilePos2& centrePos, const uint8 rotation) const;
// []<const AirportBuilding> getBuildingPositions() const;
// []<const std::uint8> getBuildingParts(const uint8 buildingType) const;
// []<const uint8> getBuildingPartHeights() const;
// []<const BuildingPartAnimation> getBuildingPartAnimations() const;
// []<const MovementNode> getMovementNodes() const;
// []<const MovementEdge> getMovementEdges() const;
// func HasFlags(flagsToTest AirportObjectFlags) bool
// return (flags & flagsToTest) != AirportObjectFlags::none;
// static_assert(sizeof(AirportObject) == 0xBA);
