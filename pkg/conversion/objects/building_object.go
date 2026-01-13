package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "BuildingCommon.h"
// #include "Graphics/Colour.h"
// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type BuildingObjectFlags int

const (
	None BuildingObjectFlags = 0
	LargeTile BuildingObjectFlags = 1 << 0
	MiscBuilding BuildingObjectFlags = 1 << 1
	Indestructible BuildingObjectFlags = 1 << 2
	IsHeadquarters BuildingObjectFlags = 1 << 3
	HasShadows BuildingObjectFlags = 1 << 4
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(BuildingObjectFlags);
// // Todo this is the same as industry obj
type PartAnimation struct {
	NumFrames uint8
	AnimationSpeed uint8
}
// static_assert(sizeof(PartAnimation) == 0x2);
type BuildingObject struct {
	Name StringId
	Image uint32
	NumParts uint8
	NumVariations uint8
	PartHeightsOffset uint32
	PartAnimationsOffset uint32
// uint32_t variationPartsOffset[32];         // 0x10 Access with getBuildingParts helper method
	Colours uint32
	DesignedYear uint16
	ObsoleteYear uint16
	Flags BuildingObjectFlags
	ClearCostIndex uint8
	ClearCostFactor uint16
	ScaffoldingSegmentType uint8
	ScaffoldingColour Colour
	GeneratorFunction uint8
	AverageNumberOnMap uint8
// uint8_t producedQuantity[2];               // 0xA0
// uint8_t producedCargoType[2];              // 0xA2
// uint8_t requiredCargoType[2];              // 0xA4
// uint8_t var_A6[2];                         // 0xA6
// uint8_t var_A8[2];                         // 0xA8
	DemolishRatingReduction int16
	Var_AC uint8
	NumElevatorSequences uint8
// uint32_t elevatorHeightSequencesOffset[4]; // 0XAE Access with getElevatorHeightSequence helper method
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: void drawBuilding(Gfx::DrawingContext& drawingCtx, uint8_t buildingRotation, int16_t x, int16_t y, Colour colour) const;
	// method: void drawDescription(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y, [[maybe_unused]] const int16_t width) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects* dependencies);
	// method: void unload();
// std::span<const uint8_t> getBuildingPartHeights() const;
// std::span<const std::uint8_t> getBuildingParts(const uint8_t variation) const;
// std::span<const std::uint8_t> getElevatorHeightSequence(const uint8_t animIdx) const;
// std::span<const BuildingPartAnimation> getBuildingPartAnimations() const;
	// method: constexpr bool hasFlags(BuildingObjectFlags flagsToTest) const
// return (flags & flagsToTest) != BuildingObjectFlags::none;
}
const BuildingObjectObjectType any = ObjectType.building
// static_assert(sizeof(BuildingObject) == 0xBE);
