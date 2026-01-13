package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Map/Track/TrackEnum.h"
// #include "Object.h"
// #include "Types.hpp"
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type RoadExtraObject struct {
	Name StringId
// World::Track::RoadTraitFlags roadPieces; // 0x02
	PaintStyle uint8
	CostIndex uint8
	BuildCostFactor int16
	SellCostFactor int16
	Image uint32
	Var_0E uint32
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
}
const RoadExtraObjectObjectType any = ObjectType.roadExtra
// static_assert(sizeof(RoadExtraObject) == 0x12);
// namespace RoadExtraObj::ImageIds
// // Note: Style imageIds are relative to 0x0A so you need to +8 to get its
// // real id relative to object at rest
// namespace Style1
const Straight0NE uint32 = 0
const Straight0SE uint32 = 1
const RightCurveSmall0NE uint32 = 2
const RightCurveSmall3NE uint32 = 3
const RightCurveSmall0SE uint32 = 4
const RightCurveSmall3SE uint32 = 5
const RightCurveSmall0SW uint32 = 6
const RightCurveSmall3SW uint32 = 7
const RightCurveSmall0NW uint32 = 8
const RightCurveSmall3NW uint32 = 9
const RightCurveVerySmall0NE uint32 = 10
const RightCurveVerySmall0SE uint32 = 11
const RightCurveVerySmall0SW uint32 = 12
const RightCurveVerySmall0NW uint32 = 13
const Turnaround0NE uint32 = 14
const Turnaround0SE uint32 = 15
const Turnaround0SW uint32 = 16
const Turnaround0NW uint32 = 17
const StraightSlopeUp0NE uint32 = 18
const StraightSlopeUp0SE uint32 = 19
const StraightSlopeUp0SW uint32 = 20
const StraightSlopeUp0NW uint32 = 21
const StraightSlopeUp1NE uint32 = 22
const StraightSlopeUp1SE uint32 = 23
const StraightSlopeUp1SW uint32 = 24
const StraightSlopeUp1NW uint32 = 25
const StraightSteepSlopeUp0NE uint32 = 26
const StraightSteepSlopeUp0SE uint32 = 27
const StraightSteepSlopeUp0SW uint32 = 28
const StraightSteepSlopeUp0NW uint32 = 29
const SupportStraight0SE uint32 = 30
const SupportConnectorStraight0SE uint32 = 31
const SupportStraight0SW uint32 = 32
const SupportConnectorStraight0SW uint32 = 33
const SupportStraight0NW uint32 = 34
const SupportConnectorStraight0NW uint32 = 35
const SupportStraight0NE uint32 = 36
const SupportConnectorStraight0NE uint32 = 37
