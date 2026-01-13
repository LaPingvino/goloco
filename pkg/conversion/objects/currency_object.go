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
type CurrencyObject struct {
	Name StringId
	PrefixSymbol StringId
	SuffixSymbol StringId
	ObjectIcon uint32
	Separator uint8
	Factor uint8
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
}
const CurrencyObjectObjectType any = ObjectType.currency
// static_assert(sizeof(CurrencyObject) == 0xC);
