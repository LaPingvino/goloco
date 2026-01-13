package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Map/Track/TrackEnum.h"
// #include "Object.h"
// #include "Speed.hpp"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
// namespace Gfx
// forward: class DrawingContext;
type BridgeObjectFlags int

const (
	None BridgeObjectFlags = 0
	HasRoof BridgeObjectFlags = 1 << 0
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(BridgeObjectFlags);
type BridgeObject struct {
	Name StringId
	Flags BridgeObjectFlags
	Pad_03 uint8
	ClearHeight uint16
	DeckDepth int16
	SpanLength uint8
	PillarSpacing uint8
	MaxSpeed Speed16
// World::MicroZ maxHeight;                         // 0x0C MicroZ!
	CostIndex uint8
	BaseCostFactor int16
	HeightCostFactor int16
	SellCostFactor int16
// World::Track::CommonTraitFlags disabledTrackCfg; // 0x14
	Image uint32
	TrackNumCompatible uint8
// uint8_t trackMods[7];                            // 0x1B
	RoadNumCompatible uint8
// uint8_t roadMods[7];                             // 0x23
	DesignedYear uint16
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects*);
	// method: void unload();
}
const BridgeObjectObjectType any = ObjectType.bridge
// static_assert(sizeof(BridgeObject) == 0x2C);
// namespace Bridge::ImageIds
// // Unlike most objects bridges have a fixed number
// // of images and this number doesn't change dynamically
// // with the features of the bridge. If the image is
// // not required due to the features of the bridge an
// // empty 1x1 pixel is normal used.
// // Image displayed in the bridge selection dropdown
const UiDropdown uint32 = 0
const DeckBaseNoSupport uint32 = 1
const DeckBaseNoSupportEdge0 uint32 = 2
const DeckBaseNoSupportEdge1 uint32 = 3
const DeckBaseNoSupportEdge2 uint32 = 4
const DeckBaseNoSupportEdge3 uint32 = 5
const RoofFullTile uint32 = 6
const RoofEdge0 uint32 = 7
const RoofEdge1 uint32 = 8
const RoofEdge2 uint32 = 9
const RoofEdge3 uint32 = 10
// // Front facing edge with any support header
const DeckBaseWithSupportHeaderEdge3 uint32 = 11
const DeckWallEdge3 uint32 = 12
const DeckWallEdge1 uint32 = 13
const DeckWallEdge0 uint32 = 14
const DeckWallEdge2 uint32 = 15
const SupportSegmentLhs16NE uint32 = 16
const SupportSegmentRhs16NE uint32 = 17
const SupportSegmentLhs32NE uint32 = 18
const SupportSegmentRhs32NE uint32 = 19
const SupportSegmentLhsSlope1NE uint32 = 20
const SupportSegmentLhsSlope2NE uint32 = 21
const SupportSegmentRhsSlope1NE uint32 = 22
const SupportSegmentRhsSlope2NE uint32 = 23
const SupportSegmentLhs16SW uint32 = 24
const SupportSegmentRhs16SW uint32 = 25
const SupportSegmentLhs32SW uint32 = 26
const SupportSegmentRhs32SW uint32 = 27
const SupportSegmentLhsSlope1SW uint32 = 28
const SupportSegmentLhsSlope2SW uint32 = 29
const SupportSegmentRhsSlope1SW uint32 = 30
const SupportSegmentRhsSlope2SW uint32 = 31
const SupportSegmentEdge3Lhs16SW uint32 = 32
const SupportSegmentEdge3Rhs16SW uint32 = 33
const SupportSegmentEdge3Lhs32SW uint32 = 34
const SupportSegmentEdge3Rhs32SW uint32 = 35
const DeckBaseNEspan0 uint32 = 36
const SupportHeaderLhsNEspan0 uint32 = 37
const SupportHeaderRhsNEspan0 uint32 = 38
const DeckWallLhsNEspan0 uint32 = 39
const DeckWallRhsNEspan0 uint32 = 40
const RoofNEspan0 uint32 = 41
const DeckBaseNEspan1 uint32 = 42
const SupportHeaderLhsNEspan1 uint32 = 43
const SupportHeaderRhsNEspan1 uint32 = 44
const DeckWallLhsNEspan1 uint32 = 45
const DeckWallRhsNEspan1 uint32 = 46
const RoofNEspan1 uint32 = 47
const DeckBaseNEspan2 uint32 = 48
const SupportHeaderLhsNEspan2 uint32 = 49
const SupportHeaderRhsNEspan2 uint32 = 50
const DeckWallLhsNEspan2 uint32 = 51
const DeckWallRhsNEspan2 uint32 = 52
const RoofNEspan2 uint32 = 53
const DeckBaseNEspan3 uint32 = 54
const SupportHeaderLhsNEspan3 uint32 = 55
const SupportHeaderRhsNEspan3 uint32 = 56
const DeckWallLhsNEspan3 uint32 = 57
const DeckWallRhsNEspan3 uint32 = 58
const RoofNEspan3 uint32 = 59
const DeckBaseSWspan0 uint32 = 60
const SupportHeaderLhsSWspan0 uint32 = 61
const SupportHeaderRhsSWspan0 uint32 = 62
const DeckWallLhsSWspan0 uint32 = 63
const DeckWallRhsSWspan0 uint32 = 64
const RoofSWspan0 uint32 = 65
const DeckBaseSWspan1 uint32 = 66
const SupportHeaderLhsSWspan1 uint32 = 67
const SupportHeaderRhsSWspan1 uint32 = 68
const DeckWallLhsSWspan1 uint32 = 69
const DeckWallRhsSWspan1 uint32 = 70
const RoofSWspan1 uint32 = 71
const DeckBaseSWspan2 uint32 = 72
const SupportHeaderLhsSWspan2 uint32 = 73
const SupportHeaderRhsSWspan2 uint32 = 74
const DeckWallLhsSWspan2 uint32 = 75
const DeckWallRhsSWspan2 uint32 = 76
const RoofSWspan2 uint32 = 77
const DeckBaseSWspan3 uint32 = 78
const SupportHeaderLhsSWspan3 uint32 = 79
const SupportHeaderRhsSWspan3 uint32 = 80
const DeckWallLhsSWspan3 uint32 = 81
const DeckWallRhsSWspan3 uint32 = 82
const RoofSWspan3 uint32 = 83
// // The following are of the form:
// // deckBase [slope] [rotation] [index]
// // deckWall [side] [slope] [rotation] [index]
// //
// // Slope is either gentle or steep
// // Gentle slopes are over 2 tiles so can be index 0 or 1
const DeckBaseGentleSlopeNE0 uint32 = 84
const DeckWallLhsGentleSlopeNE0 uint32 = 85
const DeckWallRhsGentleSlopeNE0 uint32 = 86
const DeckBaseGentleSlopeNE1 uint32 = 87
const DeckWallLhsGentleSlopeNE1 uint32 = 88
const DeckWallRhsGentleSlopeNE1 uint32 = 89
const DeckBaseGentleSlopeSE0 uint32 = 90
const DeckWallLhsGentleSlopeSE0 uint32 = 91
const DeckWallRhsGentleSlopeSE0 uint32 = 92
const DeckBaseGentleSlopeSE1 uint32 = 93
const DeckWallLhsGentleSlopeSE1 uint32 = 94
const DeckWallRhsGentleSlopeSE1 uint32 = 95
const DeckBaseGentleSlopeSW0 uint32 = 96
const DeckWallLhsGentleSlopeSW0 uint32 = 97
const DeckWallRhsGentleSlopeSW0 uint32 = 98
const DeckBaseGentleSlopeSW1 uint32 = 99
const DeckWallLhsGentleSlopeSW1 uint32 = 100
const DeckWallRhsGentleSlopeSW1 uint32 = 101
const DeckBaseGentleSlopeNW0 uint32 = 102
const DeckWallLhsGentleSlopeNW0 uint32 = 103
const DeckWallRhsGentleSlopeNW0 uint32 = 104
const DeckBaseGentleSlopeNW1 uint32 = 105
const DeckWallLhsGentleSlopeNW1 uint32 = 106
const DeckWallRhsGentleSlopeNW1 uint32 = 107
const DeckBaseSteepSlopeNE uint32 = 108
const DeckWallLhsSteepSlopeNE uint32 = 109
const DeckWallRhsSteepSlopeNE uint32 = 110
const DeckBaseSteepSlopeSE uint32 = 111
const DeckWallLhsSteepSlopeSE uint32 = 112
const DeckWallRhsSteepSlopeSE uint32 = 113
const DeckBaseSteepSlopeSW uint32 = 114
const DeckWallLhsSteepSlopeSW uint32 = 115
const DeckWallRhsSteepSlopeSW uint32 = 116
const DeckBaseSteepSlopeNW uint32 = 117
const DeckWallLhsSteepSlopeNW uint32 = 118
const DeckWallRhsSteepSlopeNW uint32 = 119
// // Even though this is on a sloped curve it is only
// // ever a flat piece as there isn't an up/down variation
const DeckBaseSlopedCurveEdge0 uint32 = 120
const DeckBaseSlopedCurveEdge1 uint32 = 121
const DeckBaseSlopedCurveEdge2 uint32 = 122
const DeckBaseSlopedCurveEdge3 uint32 = 123
