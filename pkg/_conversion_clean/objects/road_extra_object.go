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
	PaintStyle      uint8
	CostIndex       uint8
	BuildCostFactor int16
	SellCostFactor  int16
	Image           uint32
	Var_0E          uint32
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16 x, const int16 y) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, []<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
}

// MALFORMED FIELD: const RoadExtraObjectObjectType any = ObjectType.roadExtra

// static_assert(sizeof(RoadExtraObject) == 0x12);
// namespace RoadExtraObj::ImageIds
// // Note: Style imageIds are relative to 0x0A so you need to +8 to get its
// // real id relative to object at rest
// namespace Style1
// MALFORMED FIELD: const Straight0NE uint32 = 0
// MALFORMED FIELD: const Straight0SE uint32 = 1
// MALFORMED FIELD: const RightCurveSmall0NE uint32 = 2
// MALFORMED FIELD: const RightCurveSmall3NE uint32 = 3
// MALFORMED FIELD: const RightCurveSmall0SE uint32 = 4
// MALFORMED FIELD: const RightCurveSmall3SE uint32 = 5
// MALFORMED FIELD: const RightCurveSmall0SW uint32 = 6
// MALFORMED FIELD: const RightCurveSmall3SW uint32 = 7
// MALFORMED FIELD: const RightCurveSmall0NW uint32 = 8
// MALFORMED FIELD: const RightCurveSmall3NW uint32 = 9
// MALFORMED FIELD: const RightCurveVerySmall0NE uint32 = 10
// MALFORMED FIELD: const RightCurveVerySmall0SE uint32 = 11
// MALFORMED FIELD: const RightCurveVerySmall0SW uint32 = 12
// MALFORMED FIELD: const RightCurveVerySmall0NW uint32 = 13
// MALFORMED FIELD: const Turnaround0NE uint32 = 14
// MALFORMED FIELD: const Turnaround0SE uint32 = 15
// MALFORMED FIELD: const Turnaround0SW uint32 = 16
// MALFORMED FIELD: const Turnaround0NW uint32 = 17
// MALFORMED FIELD: const StraightSlopeUp0NE uint32 = 18
// MALFORMED FIELD: const StraightSlopeUp0SE uint32 = 19
// MALFORMED FIELD: const StraightSlopeUp0SW uint32 = 20
// MALFORMED FIELD: const StraightSlopeUp0NW uint32 = 21
// MALFORMED FIELD: const StraightSlopeUp1NE uint32 = 22
// MALFORMED FIELD: const StraightSlopeUp1SE uint32 = 23
// MALFORMED FIELD: const StraightSlopeUp1SW uint32 = 24
// MALFORMED FIELD: const StraightSlopeUp1NW uint32 = 25
// MALFORMED FIELD: const StraightSteepSlopeUp0NE uint32 = 26
// MALFORMED FIELD: const StraightSteepSlopeUp0SE uint32 = 27
// MALFORMED FIELD: const StraightSteepSlopeUp0SW uint32 = 28
// MALFORMED FIELD: const StraightSteepSlopeUp0NW uint32 = 29
// MALFORMED FIELD: const SupportStraight0SE uint32 = 30
// MALFORMED FIELD: const SupportConnectorStraight0SE uint32 = 31
// MALFORMED FIELD: const SupportStraight0SW uint32 = 32
// MALFORMED FIELD: const SupportConnectorStraight0SW uint32 = 33
// MALFORMED FIELD: const SupportStraight0NW uint32 = 34
// MALFORMED FIELD: const SupportConnectorStraight0NW uint32 = 35
// MALFORMED FIELD: const SupportStraight0NE uint32 = 36
// MALFORMED FIELD: const SupportConnectorStraight0NE uint32 = 37
