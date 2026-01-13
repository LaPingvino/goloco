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
type StreetLightObject struct {
	Name StringId
// uint16_t designedYear[3]; // 0x02
	Image uint32
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
// // 0x00477F5F
func (s *StreetLightObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// void unload();
}
// static_assert(sizeof(StreetLightObject) == 0xC);
// namespace Streetlight::ImageIds
}
const StreetLightObjectObjectType any = ObjectType.streetLight
const StreetLightObjectStyle0NE uint32 = 0
const StreetLightObjectStyle0SE uint32 = 1
const StreetLightObjectStyle0SW uint32 = 2
const StreetLightObjectStyle0NW uint32 = 3
const StreetLightObjectStyle1NE uint32 = 4
const StreetLightObjectStyle1SE uint32 = 5
const StreetLightObjectStyle1SW uint32 = 6
const StreetLightObjectStyle1NW uint32 = 7
const StreetLightObjectStyle2NE uint32 = 8
const StreetLightObjectStyle2SE uint32 = 9
const StreetLightObjectStyle2SW uint32 = 10
const StreetLightObjectStyle2NW uint32 = 11
