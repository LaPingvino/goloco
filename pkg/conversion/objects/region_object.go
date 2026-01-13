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
// // Towns growth can be influenced by the region objects cargo influence
// // the cargo influence can be further filtered so it only effects
// // certain types of town
type CargoInfluenceTownFilterType int

const (
	AllTowns CargoInfluenceTownFilterType = iota
	MaySnow
	InDesert
)
type RegionObjectFlags int

const (
	None RegionObjectFlags = 0
	RightHandTraffic RegionObjectFlags = 1 << 0
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(RegionObjectFlags);
type RegionObject struct {
	Name StringId
	Image uint32
	Flags RegionObjectFlags
	NumCargoInflunceObjects uint8
// uint8_t cargoInfluenceObjectIds[4];                       // 0x09
// CargoInfluenceTownFilterType cargoInfluenceTownFilter[4]; // 0x0D valid values 0, 1, 2
	Pad_11 uint8
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
// // 0x0043CB89
func (r *RegionObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects* dependencies);
	// void unload();
}
// static_assert(sizeof(RegionObject) == 0x12);
}
const RegionObjectObjectType any = ObjectType.region
