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
type CliffEdgeObject struct {
	Name StringId
	Image uint32
// // 0x004699FC
func (c *CliffEdgeObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// void unload();
	// void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
}
// static_assert(sizeof(CliffEdgeObject) == 0x6);
}
const CliffEdgeObjectObjectType any = ObjectType.cliffEdge
