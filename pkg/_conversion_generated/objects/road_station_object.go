package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Engine/Limits.h"
// #include "Map/Track/TrackEnum.h"
// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <sfl/static_vector.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type RoadStationFlags int

const (
	None RoadStationFlags = 0
	Recolourable RoadStationFlags = 1 << 0
	Passenger RoadStationFlags = 1 << 1
	Freight RoadStationFlags = 1 << 2
	RoadEnd RoadStationFlags = 1 << 3
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(RoadStationFlags);
type RoadStationObject struct {
type CargoOffset = any /* std::array<World::Pos3, 2> */ 
	Name StringId
	PaintStyle uint8
	Height uint8
// World::Track::RoadTraitFlags roadPieces; // 0x04
	BuildCostFactor int16
	SellCostFactor int16
	CostIndex uint8
	Flags RoadStationFlags
	Image uint32
// uint32_t imageOffsets[4];                // 0x10 "sequenceIndexImageOffsets"
	NumCompatible uint8
// uint8_t mods[7];                         // 0x21
	DesignedYear uint16
	ObsoleteYear uint16
	CargoType uint8
	Pad_2D uint8
// uint32_t cargoOffsetBytes[4][4]; // 0x2E
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: void drawDescription(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y, [[maybe_unused]] const int16_t width) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
// sfl::static_vector<CargoOffset, Limits::kMaxStationCargoDensity> getCargoOffsets(const uint8_t rotation, const uint8_t nibble) const;
	// method: constexpr bool hasFlags(RoadStationFlags flagsToTest) const
// return (flags & flagsToTest) != RoadStationFlags::none;
}
const RoadStationObjectObjectType any = ObjectType.roadStation
// static_assert(sizeof(RoadStationObject) == 0x6E);
// namespace RoadStation::ImageIds
const Preview_image uint32 = 0
const Preview_image_windows uint32 = 1
const TotalPreviewImages uint32 = 2
// // These are relative to var_10
// // var_10 is the imageIds per sequenceIndex (for start/middle/end of the platform)
// namespace Style0
const StraightBackNE uint32 = 0
const StraightFrontNE uint32 = 1
const StraightCanopyNE uint32 = 2
const StraightCanopyTranslucentNE uint32 = 3
const StraightBackSE uint32 = 4
const StraightFrontSE uint32 = 5
const StraightCanopySE uint32 = 6
const StraightCanopyTranslucentSE uint32 = 7
const TotalNumImages uint32 = 8
