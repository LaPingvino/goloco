package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type LandObjectFlags int

const (
	None LandObjectFlags = 0
	HasGrowthStages LandObjectFlags = 1 << 0
	HasReplacementLandHeader LandObjectFlags = 1 << 1
	IsDesert LandObjectFlags = 1 << 2
	NoTrees LandObjectFlags = 1 << 3
	HasSharpSlopeTransition LandObjectFlags = 1 << 4
	DisableSmoothTileTransition LandObjectFlags = 1 << 5
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(LandObjectFlags);
type LandObject struct {
	Name StringId
	CostIndex uint8
	NumGrowthStages uint8
	NumImageAngles uint8
	Flags LandObjectFlags
	CliffEdgeHeader uint8
	ReplacementLandHeader uint8
	CostFactor int16
	Image uint32
	NumImagesPerGrowthStage uint32
	CliffEdgeImage uint32
	MapPixelImage uint32
	DistributionPattern uint8
	NumVariations uint8
	VariationLikelihood uint8
	Pad_1D uint8
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects* dependencies);
	// method: void unload();
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: constexpr bool hasFlags(LandObjectFlags flagsToTest) const
// return (flags & flagsToTest) != LandObjectFlags::none;
}
const LandObjectObjectType any = ObjectType.land
// static_assert(sizeof(LandObject) == 0x1E);
// namespace Land::ImageIds
const Landscape_generator_tile_icon uint32 = 1
const Toolbar_terraform_land uint32 = 3
