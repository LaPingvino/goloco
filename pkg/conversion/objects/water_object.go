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
type WaterObject struct {
	Name StringId
	CostIndex uint8
	Var_03 uint8
	CostFactor int16
	Image uint32
	MapPixelImage uint32
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
}
const WaterObjectObjectType any = ObjectType.water
// static_assert(sizeof(WaterObject) == 0xE);
// namespace Water::ImageIds
const ColourPalette uint32 = 41
const ToolbarTerraformWater uint32 = 42
