package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type WallObjectFlags int

const (
	None WallObjectFlags = 0
	HasPrimaryColour WallObjectFlags = 1 << 0
	HasGlass WallObjectFlags = 1 << 1
	OnlyOnLevelLand WallObjectFlags = 1 << 2
	TwoSided WallObjectFlags = 1 << 3
	Unk4 WallObjectFlags = 1 << 4
	Unk5 WallObjectFlags = 1 << 5
	HasSecondaryColour WallObjectFlags = 1 << 6
	HasTertiaryColour WallObjectFlags = 1 << 7
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(WallObjectFlags);
type WallObject struct {
	Name StringId
	Sprite uint32
	Var_06 uint8
	Flags WallObjectFlags
// World::SmallZ height;  // 0x08
	Var_09 uint8
// // 0x004C4AF0
func (w *WallObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// void unload();
	// void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
}
// static_assert(sizeof(WallObject) == 0xA);
// namespace WallObj::ImageIds
// // If two sided the following are for the back side of the wall
}
const WallObjectObjectType any = ObjectType.wall
const WallObjectFlatSE uint32 = 0
const WallObjectFlatNE uint32 = 1
const WallObjectSlopedSE uint32 = 2
const WallObjectSlopedNE uint32 = 3
const WallObjectSlopedNW uint32 = 4
const WallObjectSlopedSW uint32 = 5
const WallObjectGlassFlatSE uint32 = 6
const WallObjectGlassFlatNE uint32 = 7
const WallObjectGlassSlopedSE uint32 = 8
const WallObjectGlassSlopedNE uint32 = 9
const WallObjectGlassSlopedNW uint32 = 10
const WallObjectGlassSlopedSW uint32 = 11
