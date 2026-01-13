package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Object.h"
// #include "Types.hpp"
// #include <array>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type ScaffoldingObject struct {
	Name StringId
	Image uint32
// uint16_t segmentHeights[3]; // 0x06
// uint16_t roofHeights[3];    // 0x0C
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
// // 0x0042DF0B
func (s *ScaffoldingObject) Validate() bool {
	// void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// void unload();
}
// static_assert(sizeof(ScaffoldingObject) == 0x12);
// namespace Scaffolding::ImageIds
}
const ScaffoldingObjectObjectType any = ObjectType.scaffolding
const ScaffoldingObjectType01x1SegmentBack uint32 = 0
const ScaffoldingObjectType01x1SegmentFront uint32 = 1
const ScaffoldingObjectType01x1RoofNE uint32 = 2
const ScaffoldingObjectType01x1RoofSE uint32 = 3
const ScaffoldingObjectType01x1RoofSW uint32 = 4
const ScaffoldingObjectType01x1RoofNW uint32 = 5
const ScaffoldingObjectType02x2SegmentBack uint32 = 6
const ScaffoldingObjectType02x2SegmentFront uint32 = 7
const ScaffoldingObjectType02x2RoofNE uint32 = 8
const ScaffoldingObjectType02x2RoofSE uint32 = 9
const ScaffoldingObjectType02x2RoofSW uint32 = 10
const ScaffoldingObjectType02x2RoofNW uint32 = 11
const ScaffoldingObjectType11x1SegmentBack uint32 = 12
const ScaffoldingObjectType11x1SegmentFront uint32 = 13
const ScaffoldingObjectType11x1RoofNE uint32 = 14
const ScaffoldingObjectType11x1RoofSE uint32 = 15
const ScaffoldingObjectType11x1RoofSW uint32 = 16
const ScaffoldingObjectType11x1RoofNW uint32 = 17
const ScaffoldingObjectType12x2SegmentBack uint32 = 18
const ScaffoldingObjectType12x2SegmentFront uint32 = 19
const ScaffoldingObjectType12x2RoofNE uint32 = 20
const ScaffoldingObjectType12x2RoofSE uint32 = 21
const ScaffoldingObjectType12x2RoofSW uint32 = 22
const ScaffoldingObjectType12x2RoofNW uint32 = 23
const ScaffoldingObjectType21x1SegmentBack uint32 = 24
const ScaffoldingObjectType21x1SegmentFront uint32 = 25
const ScaffoldingObjectType21x1RoofNE uint32 = 26
const ScaffoldingObjectType21x1RoofSE uint32 = 27
const ScaffoldingObjectType21x1RoofSW uint32 = 28
const ScaffoldingObjectType21x1RoofNW uint32 = 29
const ScaffoldingObjectType22x2SegmentBack uint32 = 30
const ScaffoldingObjectType22x2SegmentFront uint32 = 31
const ScaffoldingObjectType22x2RoofNE uint32 = 32
const ScaffoldingObjectType22x2RoofSE uint32 = 33
const ScaffoldingObjectType22x2RoofSW uint32 = 34
const ScaffoldingObjectType22x2RoofNW uint32 = 35
type ScaffoldingImages struct {
}

type ScaffoldingImagesBuilding struct {
	Back uint32
	Front uint32
// std::array<uint32_t, 4> roof;
func (s *ScaffoldingImages) GetRoof(rotation uint8) uint32 {
}
// Building buildings[2];
// constexpr const Building& get1x1() const { return buildings[0]; }
// constexpr const Building& get2x2() const { return buildings[1]; }
}
var ScaffoldingImages = [3]ScaffoldingImages{
// ScaffoldingImages{
// ScaffoldingImages::Building{
// Scaffolding::ImageIds::type01x1SegmentBack,
// Scaffolding::ImageIds::type01x1SegmentFront,
// Scaffolding::ImageIds::type01x1RoofNE,
// Scaffolding::ImageIds::type01x1RoofSE,
// Scaffolding::ImageIds::type01x1RoofSW,
// Scaffolding::ImageIds::type01x1RoofNW,
// },
// },
// ScaffoldingImages::Building{
// Scaffolding::ImageIds::type02x2SegmentBack,
// Scaffolding::ImageIds::type02x2SegmentFront,
// Scaffolding::ImageIds::type02x2RoofNE,
// Scaffolding::ImageIds::type02x2RoofSE,
// Scaffolding::ImageIds::type02x2RoofSW,
// Scaffolding::ImageIds::type02x2RoofNW,
// },
// },
// },
// ScaffoldingImages{
// ScaffoldingImages::Building{
// Scaffolding::ImageIds::type11x1SegmentBack,
// Scaffolding::ImageIds::type11x1SegmentFront,
// Scaffolding::ImageIds::type11x1RoofNE,
// Scaffolding::ImageIds::type11x1RoofSE,
// Scaffolding::ImageIds::type11x1RoofSW,
// Scaffolding::ImageIds::type11x1RoofNW,
// },
// },
// ScaffoldingImages::Building{
// Scaffolding::ImageIds::type12x2SegmentBack,
// Scaffolding::ImageIds::type12x2SegmentFront,
// Scaffolding::ImageIds::type12x2RoofNE,
// Scaffolding::ImageIds::type12x2RoofSE,
// Scaffolding::ImageIds::type12x2RoofSW,
// Scaffolding::ImageIds::type12x2RoofNW,
// },
// },
// },
// ScaffoldingImages{
// ScaffoldingImages::Building{
// Scaffolding::ImageIds::type21x1SegmentBack,
// Scaffolding::ImageIds::type21x1SegmentFront,
// Scaffolding::ImageIds::type21x1RoofNE,
// Scaffolding::ImageIds::type21x1RoofSE,
// Scaffolding::ImageIds::type21x1RoofSW,
// Scaffolding::ImageIds::type21x1RoofNW,
// },
// },
// ScaffoldingImages::Building{
// Scaffolding::ImageIds::type22x2SegmentBack,
// Scaffolding::ImageIds::type22x2SegmentFront,
// Scaffolding::ImageIds::type22x2RoofNE,
// Scaffolding::ImageIds::type22x2RoofSE,
// Scaffolding::ImageIds::type22x2RoofSW,
// Scaffolding::ImageIds::type22x2RoofNW,
// },
// },
// },
}
// constexpr const ScaffoldingImages& getScaffoldingImages(uint8_t type) { return kScaffoldingImages[type]; }
