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
type DockObjectFlags int

const (
func init() {
	None       DockObjectFlags = 0
	HasShadows DockObjectFlags = 1 << 0
}

)

// OPENLOCO_ENABLE_ENUM_OPERATORS(DockObjectFlags);
type DockObject struct {
	Name                         StringId
	BuildCostFactor              int16
	SellCostFactor               int16
	CostIndex                    uint8
	Var_07                       uint8
	Image                        uint32
	BuildingImage                uint32
	Flags                        DockObjectFlags
	NumBuildingParts             uint8
	NumBuildingVariations        uint8
	PartHeightsOffset            uint32
	BuildingPartAnimationsOffset uint32
	// uint32 buildingVariationPartsOffset[1]; // 0x1C odd that this is size 1 but that is how its used
	DesignedYear uint16
	ObsoleteYear uint16
	// World::Pos2 boatPosition;                 // 0x24
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16 x, const int16 y) const;
	// method: void drawDescription(Gfx::DrawingContext& drawingCtx, const int16 x, const int16 y, [[maybe_unused]] const int16 width) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, []<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
	// []<const std::uint8> getBuildingParts(const uint8 buildingType) const;
	// []<const BuildingPartAnimation> getBuildingPartAnimations() const;
	// []<const uint8> getBuildingPartHeights() const;
	// method: constexpr bool hasFlags(DockObjectFlags flagsToTest) const
	// return (flags & flagsToTest) != DockObjectFlags::none;

const DockObjectObjectType any = ObjectType.dock

// static_assert(sizeof(DockObject) == 0x28);
