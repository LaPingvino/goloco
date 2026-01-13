package objects

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Map/Track/TrackEnum.h"
// #include "Object.h"
// #include "Speed.hpp"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <span>
// namespace OpenLoco
// namespace ObjectManager
// forward: struct DependentObjects;
type TownSize int

const (
// namespace Gfx
// forward: class DrawingContext;
)
type TownSize int

const (
type RoadObjectFlags int

const (
	None RoadObjectFlags = 0
	IsOneWay RoadObjectFlags = 1 << 0
	IsRail RoadObjectFlags = 1 << 1
	Unk_02 RoadObjectFlags = 1 << 2
	AnyRoadTypeCompatible RoadObjectFlags = 1 << 3
	NoSlipSurface RoadObjectFlags = 1 << 4
	HasRackRail RoadObjectFlags = 1 << 5
	IsRoad RoadObjectFlags = 1 << 6
	AllowUseByAllCompanies RoadObjectFlags = 1 << 7
	CanHaveStreetLights RoadObjectFlags = 1 << 8
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(RoadObjectFlags);
type RoadObject struct {
	Name StringId
// World::Track::RoadTraitFlags roadPieces; // 0x02
	BuildCostFactor int16
	SellCostFactor int16
	TunnelCostFactor int16
	CostIndex uint8
	Tunnel uint8
	MaxSpeed Speed16
	Image uint32
	Flags RoadObjectFlags
	NumBridges uint8
// uint8_t bridges[7];                      // 0x15
	NumStations uint8
// uint8_t stations[7];                     // 0x1D
	PaintStyle uint8
	NumMods uint8
// uint8_t mods[2];                         // 0x26
	NumCompatible uint8
	DisplayOffset uint8
	CompatibleRoads uint16
	CompatibleTracks uint16
	TargetTownSize TownSize
	Pad_2F uint8
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects* dependencies);
	// method: void unload();
	// method: constexpr bool hasFlags(RoadObjectFlags flagsToTest) const
// return (flags & flagsToTest) != RoadObjectFlags::none;
}
const RoadObjectObjectType any = ObjectType.road
// func HasTraitFlags(flagsToTest World::Track::RoadTraitFlags) bool
// return (roadPieces & flagsToTest) != World::Track::RoadTraitFlags::none;
// static_assert(sizeof(RoadObject) == 0x30);
// namespace RoadObj::ImageIds
// namespace Style0
const Straight0NE uint32 = 34
const Straight0SE uint32 = 35
const RightCurveVerySmall0NE uint32 = 36
const RightCurveVerySmall0SE uint32 = 37
const RightCurveVerySmall0SW uint32 = 38
const RightCurveVerySmall0NW uint32 = 39
const JunctionLeft0NE uint32 = 40
const JunctionLeft0SE uint32 = 41
const JunctionLeft0SW uint32 = 42
const JunctionLeft0NW uint32 = 43
const JunctionCrossroads0NE uint32 = 44
const RightCurveSmall0NE uint32 = 45
const RightCurveSmall1NE uint32 = 46
const RightCurveSmall2NE uint32 = 47
const RightCurveSmall3NE uint32 = 48
const RightCurveSmall0SE uint32 = 49
const RightCurveSmall1SE uint32 = 50
const RightCurveSmall2SE uint32 = 51
const RightCurveSmall3SE uint32 = 52
const RightCurveSmall0SW uint32 = 53
const RightCurveSmall1SW uint32 = 54
const RightCurveSmall2SW uint32 = 55
const RightCurveSmall3SW uint32 = 56
const RightCurveSmall0NW uint32 = 57
const RightCurveSmall1NW uint32 = 58
const RightCurveSmall2NW uint32 = 59
const RightCurveSmall3NW uint32 = 60
const StraightSlopeUp0NE uint32 = 61
const StraightSlopeUp1NE uint32 = 62
const StraightSlopeUp0SE uint32 = 63
const StraightSlopeUp1SE uint32 = 64
const StraightSlopeUp0SW uint32 = 65
const StraightSlopeUp1SW uint32 = 66
const StraightSlopeUp0NW uint32 = 67
const StraightSlopeUp1NW uint32 = 68
const StraightSteepSlopeUp0NE uint32 = 69
const StraightSteepSlopeUp0SE uint32 = 70
const StraightSteepSlopeUp0SW uint32 = 71
const StraightSteepSlopeUp0NW uint32 = 72
const Turnaround0NE uint32 = 73
const Turnaround0SE uint32 = 74
const Turnaround0SW uint32 = 75
const Turnaround0NW uint32 = 76
// // Assumes rotational symmetry
// // k{TrackId}{sequenceIndex}{type}{direction}
// // type = Ballast, Sleeper, Rail
// namespace Style1
const Straight0BallastNE uint32 = 34
const Straight0BallastSE uint32 = 35
const Straight0SleeperNE uint32 = 36
const Straight0SleeperSE uint32 = 37
const Straight0RailNE uint32 = 38
const Straight0RailSE uint32 = 39
const RightCurveSmall0BallastNE uint32 = 40
const RightCurveSmall1BallastNE uint32 = 41
const RightCurveSmall2BallastNE uint32 = 42
const RightCurveSmall3BallastNE uint32 = 43
const RightCurveSmall0BallastSE uint32 = 44
const RightCurveSmall1BallastSE uint32 = 45
const RightCurveSmall2BallastSE uint32 = 46
const RightCurveSmall3BallastSE uint32 = 47
const RightCurveSmall0BallastSW uint32 = 48
const RightCurveSmall1BallastSW uint32 = 49
const RightCurveSmall2BallastSW uint32 = 50
const RightCurveSmall3BallastSW uint32 = 51
const RightCurveSmall0BallastNW uint32 = 52
const RightCurveSmall1BallastNW uint32 = 53
const RightCurveSmall2BallastNW uint32 = 54
const RightCurveSmall3BallastNW uint32 = 55
const RightCurveSmall0SleeperNE uint32 = 56
const RightCurveSmall1SleeperNE uint32 = 57
const RightCurveSmall2SleeperNE uint32 = 58
const RightCurveSmall3SleeperNE uint32 = 59
const RightCurveSmall0SleeperSE uint32 = 60
const RightCurveSmall1SleeperSE uint32 = 61
const RightCurveSmall2SleeperSE uint32 = 62
const RightCurveSmall3SleeperSE uint32 = 63
const RightCurveSmall0SleeperSW uint32 = 64
const RightCurveSmall1SleeperSW uint32 = 65
const RightCurveSmall2SleeperSW uint32 = 66
const RightCurveSmall3SleeperSW uint32 = 67
const RightCurveSmall0SleeperNW uint32 = 68
const RightCurveSmall1SleeperNW uint32 = 69
const RightCurveSmall2SleeperNW uint32 = 70
const RightCurveSmall3SleeperNW uint32 = 71
const RightCurveSmall0RailNE uint32 = 72
const RightCurveSmall1RailNE uint32 = 73
const RightCurveSmall2RailNE uint32 = 74
const RightCurveSmall3RailNE uint32 = 75
const RightCurveSmall0RailSE uint32 = 76
const RightCurveSmall1RailSE uint32 = 77
const RightCurveSmall2RailSE uint32 = 78
const RightCurveSmall3RailSE uint32 = 79
const RightCurveSmall0RailSW uint32 = 80
const RightCurveSmall1RailSW uint32 = 81
const RightCurveSmall2RailSW uint32 = 82
const RightCurveSmall3RailSW uint32 = 83
const RightCurveSmall0RailNW uint32 = 84
const RightCurveSmall1RailNW uint32 = 85
const RightCurveSmall2RailNW uint32 = 86
const RightCurveSmall3RailNW uint32 = 87
const StraightSlopeUp0BallastNE uint32 = 88
const StraightSlopeUp1BallastNE uint32 = 89
const StraightSlopeUp0BallastSE uint32 = 90
const StraightSlopeUp1BallastSE uint32 = 91
const StraightSlopeUp0BallastSW uint32 = 92
const StraightSlopeUp1BallastSW uint32 = 93
const StraightSlopeUp0BallastNW uint32 = 94
const StraightSlopeUp1BallastNW uint32 = 95
const StraightSlopeUp0SleeperNE uint32 = 96
const StraightSlopeUp1SleeperNE uint32 = 97
const StraightSlopeUp0SleeperSE uint32 = 98
const StraightSlopeUp1SleeperSE uint32 = 99
const StraightSlopeUp0SleeperSW uint32 = 100
const StraightSlopeUp1SleeperSW uint32 = 101
const StraightSlopeUp0SleeperNW uint32 = 102
const StraightSlopeUp1SleeperNW uint32 = 103
const StraightSlopeUp0RailNE uint32 = 104
const StraightSlopeUp1RailNE uint32 = 105
const StraightSlopeUp0RailSE uint32 = 106
const StraightSlopeUp1RailSE uint32 = 107
const StraightSlopeUp0RailSW uint32 = 108
const StraightSlopeUp1RailSW uint32 = 109
const StraightSlopeUp0RailNW uint32 = 110
const StraightSlopeUp1RailNW uint32 = 111
const StraightSteepSlopeUp0BallastNE uint32 = 112
const StraightSteepSlopeUp0BallastSE uint32 = 113
const StraightSteepSlopeUp0BallastSW uint32 = 114
const StraightSteepSlopeUp0BallastNW uint32 = 115
const StraightSteepSlopeUp0SleeperNE uint32 = 116
const StraightSteepSlopeUp0SleeperSE uint32 = 117
const StraightSteepSlopeUp0SleeperSW uint32 = 118
const StraightSteepSlopeUp0SleeperNW uint32 = 119
const StraightSteepSlopeUp0RailNE uint32 = 120
const StraightSteepSlopeUp0RailSE uint32 = 121
const StraightSteepSlopeUp0RailSW uint32 = 122
const StraightSteepSlopeUp0RailNW uint32 = 123
const RightCurveVerySmall0BallastNE uint32 = 124
const RightCurveVerySmall0BallastSE uint32 = 125
const RightCurveVerySmall0BallastSW uint32 = 126
const RightCurveVerySmall0BallastNW uint32 = 127
const RightCurveVerySmall0SleeperNE uint32 = 128
const RightCurveVerySmall0SleeperSE uint32 = 129
const RightCurveVerySmall0SleeperSW uint32 = 130
const RightCurveVerySmall0SleeperNW uint32 = 131
const RightCurveVerySmall0RailNE uint32 = 132
const RightCurveVerySmall0RailSE uint32 = 133
const RightCurveVerySmall0RailSW uint32 = 134
const RightCurveVerySmall0RailNW uint32 = 135
const Turnaround0BallastNE uint32 = 136
const Turnaround0BallastSE uint32 = 137
const Turnaround0BallastSW uint32 = 138
const Turnaround0BallastNW uint32 = 139
const Turnaround0SleeperNE uint32 = 140
const Turnaround0SleeperSE uint32 = 141
const Turnaround0SleeperSW uint32 = 142
const Turnaround0SleeperNW uint32 = 143
const Turnaround0RailNE uint32 = 144
const Turnaround0RailSE uint32 = 145
const Turnaround0RailSW uint32 = 146
const Turnaround0RailNW uint32 = 147
// namespace Style2
const Straight0NE uint32 = 34
const Straight0SE uint32 = 35
const LeftCurveVerySmall0NW uint32 = 36
const LeftCurveVerySmall0NE uint32 = 37
const LeftCurveVerySmall0SE uint32 = 38
const LeftCurveVerySmall0SW uint32 = 39
const JunctionLeft0NE uint32 = 40
const JunctionLeft0SE uint32 = 41
const JunctionLeft0SW uint32 = 42
const JunctionLeft0NW uint32 = 43
const JunctionCrossroads0NE uint32 = 44
const LeftCurveSmall3NW uint32 = 45
const LeftCurveSmall1NW uint32 = 46
const LeftCurveSmall2NW uint32 = 47
const LeftCurveSmall0NW uint32 = 48
const LeftCurveSmall3NE uint32 = 49
const LeftCurveSmall1NE uint32 = 50
const LeftCurveSmall2NE uint32 = 51
const LeftCurveSmall0NE uint32 = 52
const LeftCurveSmall3SE uint32 = 53
const LeftCurveSmall1SE uint32 = 54
const LeftCurveSmall2SE uint32 = 55
const LeftCurveSmall0SE uint32 = 56
const LeftCurveSmall3SW uint32 = 57
const LeftCurveSmall1SW uint32 = 58
const LeftCurveSmall2SW uint32 = 59
const LeftCurveSmall0SW uint32 = 60
const StraightSlopeUp0NE uint32 = 61
const StraightSlopeUp1NE uint32 = 62
const StraightSlopeUp0SE uint32 = 63
const StraightSlopeUp1SE uint32 = 64
const StraightSlopeUp0SW uint32 = 65
const StraightSlopeUp1SW uint32 = 66
const StraightSlopeUp0NW uint32 = 67
const StraightSlopeUp1NW uint32 = 68
const StraightSteepSlopeUp0NE uint32 = 69
const StraightSteepSlopeUp0SE uint32 = 70
const StraightSteepSlopeUp0SW uint32 = 71
const StraightSteepSlopeUp0NW uint32 = 72
const Turnaround0NE uint32 = 73
const Turnaround0SE uint32 = 74
const Turnaround0SW uint32 = 75
const Turnaround0NW uint32 = 76
const Straight0SW uint32 = 85
const Straight0NW uint32 = 86
const RightCurveVerySmall0NE uint32 = 87
const RightCurveVerySmall0SE uint32 = 88
const RightCurveVerySmall0SW uint32 = 89
const RightCurveVerySmall0NW uint32 = 90
const JunctionRight0NE uint32 = 91
const JunctionRight0SE uint32 = 92
const JunctionRight0SW uint32 = 93
const JunctionRight0NW uint32 = 94
// // Must duplicate kJunctionCrossroads0NE
const JunctionCrossroads0NE2 uint32 = 95
const RightCurveSmall0NE uint32 = 96
const RightCurveSmall1NE uint32 = 97
const RightCurveSmall2NE uint32 = 98
const RightCurveSmall3NE uint32 = 99
const RightCurveSmall0SE uint32 = 100
const RightCurveSmall1SE uint32 = 101
const RightCurveSmall2SE uint32 = 102
const RightCurveSmall3SE uint32 = 103
const RightCurveSmall0SW uint32 = 104
const RightCurveSmall1SW uint32 = 105
const RightCurveSmall2SW uint32 = 106
const RightCurveSmall3SW uint32 = 107
const RightCurveSmall0NW uint32 = 108
const RightCurveSmall1NW uint32 = 109
const RightCurveSmall2NW uint32 = 110
const RightCurveSmall3NW uint32 = 111
const StraightSlopeDown1SW uint32 = 112
const StraightSlopeDown0SW uint32 = 113
const StraightSlopeDown1NW uint32 = 114
const StraightSlopeDown0NW uint32 = 115
const StraightSlopeDown1NE uint32 = 116
const StraightSlopeDown0NE uint32 = 117
const StraightSlopeDown1SE uint32 = 118
const StraightSlopeDown0SE uint32 = 119
const StraightSteepSlopeDown0SW uint32 = 120
const StraightSteepSlopeDown0NW uint32 = 121
const StraightSteepSlopeDown0NE uint32 = 122
const StraightSteepSlopeDown0SE uint32 = 123
