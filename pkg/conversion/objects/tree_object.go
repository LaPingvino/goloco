package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <array>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type TreeObjectFlags int

const (
	None TreeObjectFlags = 0
	HasSnowVariation TreeObjectFlags = 1 << 0
	Unk1 TreeObjectFlags = 1 << 1
	HighAltitude TreeObjectFlags = 1 << 2
	LowAltitude TreeObjectFlags = 1 << 3
	RequiresWater TreeObjectFlags = 1 << 4
	Unk5 TreeObjectFlags = 1 << 5
	DroughtResistant TreeObjectFlags = 1 << 6
	HasShadow TreeObjectFlags = 1 << 7
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TreeObjectFlags);
type TreeObject struct {
	Name StringId
	InitialHeight uint8
	Height uint8
	Var_04 uint8
	Var_05 uint8
	NumRotations uint8
	Growth uint8
	Flags TreeObjectFlags
// uint32_t sprites[6];             // 0x0A
// uint32_t snowSprites[6];         // 0x22
	ShadowImageOffset uint16
	Var_3C uint8
	SeasonState uint8
	CurrentSeason uint8
	CostIndex uint8
	BuildCostFactor int16
	ClearCostFactor int16
	Colours uint32
	Rating int16
	DemolishRatingReduction int16
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: uint8_t getTreeGrowthDisplayOffset() const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
	// method: constexpr bool hasFlags(TreeObjectFlags flagsToTest) const
// return (flags & flagsToTest) != TreeObjectFlags::none;
}
const TreeObjectObjectType any = ObjectType.tree
// static_assert(sizeof(TreeObject) == 0x4C);
