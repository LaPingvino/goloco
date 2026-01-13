package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "BuildingCommon.h"
// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type IndustryObjectFlags int

const (
	None IndustryObjectFlags = 0
	BuiltInClusters IndustryObjectFlags = 1 << 0
	BuiltOnHighGround IndustryObjectFlags = 1 << 1
	BuiltOnLowGround IndustryObjectFlags = 1 << 2
	BuiltOnSnow IndustryObjectFlags = 1 << 3
	BuiltBelowSnowLine IndustryObjectFlags = 1 << 4
	BuiltOnFlatGround IndustryObjectFlags = 1 << 5
	BuiltNearWater IndustryObjectFlags = 1 << 6
	BuiltAwayFromWater IndustryObjectFlags = 1 << 7
	BuiltOnWater IndustryObjectFlags = 1 << 8
	BuiltNearTown IndustryObjectFlags = 1 << 9
	BuiltAwayFromTown IndustryObjectFlags = 1 << 10
	BuiltNearTrees IndustryObjectFlags = 1 << 11
	BuiltRequiresOpenSpace IndustryObjectFlags = 1 << 12
	OilfieldStationName IndustryObjectFlags = 1 << 13
	MinesStationName IndustryObjectFlags = 1 << 14
	NotRotatable IndustryObjectFlags = 1 << 15
	CanBeFoundedByPlayer IndustryObjectFlags = 1 << 16
	RequiresAllCargo IndustryObjectFlags = 1 << 17
	CanIncreaseProduction IndustryObjectFlags = 1 << 18
	CanDecreaseProduction IndustryObjectFlags = 1 << 19
	RequiresElectricityPylons IndustryObjectFlags = 1 << 20
	HasShadows IndustryObjectFlags = 1 << 21
	KillsTrees IndustryObjectFlags = 1 << 22
	FarmTilesGrowthStageDesynchronized IndustryObjectFlags = 1 << 23
	BuiltInDesert IndustryObjectFlags = 1 << 24
	BuiltNearDesert IndustryObjectFlags = 1 << 25
	FarmTilesDrawAboveSnow IndustryObjectFlags = 1 << 26
	FarmTilesPartialCoverage IndustryObjectFlags = 1 << 27
	FarmProductionIgnoresSnow IndustryObjectFlags = 1 << 28
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(IndustryObjectFlags);
type IndustryObjectUnk38 struct {
	Var_00 uint8
	Var_01 uint8
}
// static_assert(sizeof(IndustryObjectUnk38) == 0x2);
type IndustryObjectProductionRateRange struct {
	Min uint16
	Max uint16
}
// static_assert(sizeof(IndustryObjectProductionRateRange) == 0x4);
type IndustryObject struct {
	Name StringId
	Var_02 StringId
	NameClosingDown StringId
	NameUpProduction StringId
	NameDownProduction StringId
	NameSingular StringId
	NamePlural StringId
	ShadowImageIds uint32
	BuildingImageIds uint32
	FieldImageIds uint32
	NumImagesPerFieldGrowthStage uint32
	NumBuildingParts uint8
	NumBuildingVariations uint8
	BuildingPartHeightsOffset uint32
	BuildingPartAnimationsOffset uint32
// uint32_t animationSequenceOffsets[4];      // 0x28 Access with getAnimationSequence helper method
	Var_38_Offset uint32
// uint32_t buildingVariationPartOffsets[32]; // 0x3C Access with getBuildingParts helper method
	MinNumBuildings uint8
	MaxNumBuildings uint8
	BuildingsOffset uint32
	AvailableColours uint32
	BuildingSizeFlags uint32
	DesignedYear uint16
	ObsoleteYear uint16
// // Total industries of this type that can be created in a scenario
// // Note: this is not directly comparable to total industries and varies based
// // on scenario total industries cap settings. At low industries cap this value is ~3x the
// // amount of industries in a scenario.
	TotalOfTypeInScenario uint8
	CostIndex uint8
	CostFactor int16
	ClearCostFactor int16
	ScaffoldingSegmentType uint8
	ScaffoldingColour Colour
// IndustryObjectProductionRateRange initialProductionRate[2]; // 0xD6
// uint8_t producedCargoType[2];                               // 0xDE (0xFF = null)
// uint8_t requiredCargoType[3];                               // 0xE0 (0xFF = null)
	MapColour Colour
	Flags IndustryObjectFlags
	Var_E8 uint8
	FarmTileNumImageAngles uint8
	FarmTileGrowthStageNoProduction uint8
	FarmNumFields uint8
	FarmTileNumGrowthStages uint8
// uint8_t wallTypes[4];                    // 0xED There can be up to 4 different wall types for an industry
// // Selection of wall types isn't completely random from the 4 it is biased into 2 groups of 2 (wall and entrance)
	BuildingWall uint8
	BuildingWallEntrance uint8
	MonthlyClosureChance uint8
	// method: bool requiresCargo() const;
	// method: bool producesCargo() const;
// char* getProducedCargoString(const char* buffer) const;
// char* getRequiredCargoString(const char* buffer) const;
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: void drawIndustry(Gfx::DrawingContext& drawingCtx, int16_t x, int16_t y) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects* dependencies);
	// method: void unload();
// std::span<const std::uint8_t> getBuildingParts(const uint8_t buildingType) const;
// std::span<const std::uint8_t> getAnimationSequence(const uint8_t unk) const;
// std::span<const std::uint8_t> getBuildingPartHeights() const;
// std::span<const IndustryObjectUnk38> getUnk38() const;
// std::span<const BuildingPartAnimation> getBuildingPartAnimations() const;
// std::span<const std::uint8_t> getBuildings() const;
	// method: constexpr bool hasFlags(IndustryObjectFlags flagsToTest) const
// return (flags & flagsToTest) != IndustryObjectFlags::none;
}
const IndustryObjectObjectType any = ObjectType.industry
// static_assert(sizeof(IndustryObject) == 0xF4);
