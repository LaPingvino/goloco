package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Engine/Limits.h"
// #include "Map/Track/TrackEnum.h"
// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <array>
// #include <cstddef>
// #include <sfl/static_vector.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type TrainStationFlags int

const (
	None TrainStationFlags = 0
	Recolourable TrainStationFlags = 1 << 0
	NoCanopy TrainStationFlags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TrainStationFlags);
type TrainStationObject struct {
type CargoOffset = any /* std::array<World::Pos3, 2> */ 
	Name StringId
	PaintStyle uint8
	Height uint8
// World::Track::TrackTraitFlags trackPieces; // 0x04
	BuildCostFactor int16
	SellCostFactor int16
	CostIndex uint8
	Var_0B uint8
	Flags TrainStationFlags
	Var_0D uint8
	Image uint32
// uint32_t imageOffsets[4];        // 0x12 "sequenceIndexImageOffsets"
	NumCompatible uint8
// uint8_t mods[7];                 // 0x23
	DesignedYear uint16
	ObsoleteYear uint16
// uint32_t cargoOffsetBytes[4][4]; // 0x2E
// uint32_t var_6E[16];
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: void drawDescription(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y, [[maybe_unused]] const int16_t width) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
// sfl::static_vector<CargoOffset, Limits::kMaxStationCargoDensity> getCargoOffsets(const uint8_t rotation, const uint8_t nibble) const;
	// method: constexpr bool hasFlags(TrainStationFlags flagsToTest) const
// return (flags & flagsToTest) != TrainStationFlags::none;
}
const TrainStationObjectObjectType any = ObjectType.trainStation
// static_assert(sizeof(TrainStationObject) == 0xAE);
// namespace TrainStation::ImageIds
const Preview_image uint32 = 0
const Preview_image_windows uint32 = 1
const TotalPreviewImages uint32 = 2
// // These are relative to var_12
// // var_12 is the imageIds per sequenceIndex (for start/middle/end of the platform)
// namespace Style0
const StraightBackNE uint32 = 0
const StraightFrontNE uint32 = 1
const StraightCanopyNE uint32 = 2
const StraightCanopyTranslucentNE uint32 = 3
const StraightBackSE uint32 = 4
const StraightFrontSE uint32 = 5
const StraightCanopySE uint32 = 6
const StraightCanopyTranslucentSE uint32 = 7
const DiagonalNE0 uint32 = 8
const DiagonalNE3 uint32 = 9
const DiagonalNE1 uint32 = 10
const DiagonalCanopyNE1 uint32 = 11
const DiagonalCanopyTranslucentNE1 uint32 = 12
const DiagonalSE1 uint32 = 13
const DiagonalSE2 uint32 = 14
const DiagonalSE3 uint32 = 15
const DiagonalCanopyTranslucentSE3 uint32 = 16
const TotalNumImages uint32 = 17
// namespace Style1
const TotalNumImages uint32 = 8
