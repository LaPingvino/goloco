package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <array>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type CompetitorObject struct {
	Name StringId
	AvailableNamePrefixes StringId
	AvailablePlayStyles uint32
	Var_08 uint32
	Emotions uint32
// uint32_t images[9];             // 0x10
	Intelligence uint8
	Aggressiveness uint8
	Competitiveness uint8
	Var_37 uint8
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: void drawDescription(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y, [[maybe_unused]] const int16_t width) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
}
const CompetitorObjectObjectType any = ObjectType.competitor
// static_assert(sizeof(CompetitorObject) == 0x38);
// [[nodiscard]] StringId aiRatingToLevel(const uint8_t rating);
