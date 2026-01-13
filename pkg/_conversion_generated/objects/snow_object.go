package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type SnowObject struct {
	Name StringId
	Image uint32
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
// // 0x00469A6B
func (s *SnowObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// void unload();
}
// static_assert(sizeof(SnowObject) == 0x6);
// namespace SnowLine::ImageIds
}
const SnowObjectObjectType any = ObjectType.snow
const SnowObjectSurfaceEighthZoom uint32 = 0
const SnowObjectOutlineEighthZoom uint32 = 10
const SnowObjectSurfaceQuarterZoom uint32 = 19
const SnowObjectOutlineQuarterZoom uint32 = 38
const SnowObjectSurfaceHalfZoom uint32 = 57
const SnowObjectOutlineHalfZoom uint32 = 76
const SnowObjectSurfaceFullZoom uint32 = 95
const SnowObjectOutlineFullZoom uint32 = 114
