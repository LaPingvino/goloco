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
type HillShapeFlags int

const (
	None HillShapeFlags = 0
	IsHeightMap HillShapeFlags = 1 << 0
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(HillShapeFlags);
type HillShapesObject struct {
	Name StringId
	HillHeightMapCount uint8
	MountainHeightMapCount uint8
	Image uint32
	ImageHills uint32
	Flags HillShapeFlags
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
// // 0x00463BB3
func (h *HillShapesObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// void unload();
}
// static_assert(sizeof(HillShapesObject) == 0xE);
}
const HillShapesObjectObjectType any = ObjectType.hillShapes
