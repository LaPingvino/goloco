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
type LevelCrossingObject struct {
	Name           StringId
	CostFactor     int16
	SellCostFactor int16
	CostIndex      uint8
	AnimationSpeed uint8
	ClosingFrames  uint8
	ClosedFrames   uint8
	Var_0A         uint8
	Pad_0B         uint8
	DesignedYear   uint16
	Image          uint32
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, []<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16 x, const int16 y) const;
	// method: void drawDescription(Gfx::DrawingContext& drawingCtx, const int16 x, const int16 y, [[maybe_unused]] const int16 width) const;
}

// MALFORMED FIELD: const LevelCrossingObjectObjectType any = ObjectType.levelCrossing

// static_assert(sizeof(LevelCrossingObject) == 0x12);
