package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <cstdint>
// namespace OpenLoco::World::Track
type TrackId int

const (
	Straight TrackId = iota
	Diagonal
	LeftCurveVerySmall
	RightCurveVerySmall
	LeftCurveSmall
	RightCurveSmall
	LeftCurve
	RightCurve
	LeftCurveLarge
	RightCurveLarge
	DiagonalLeftCurveLarge
	DiagonalRightCurveLarge
	SBendLeft
	SBendRight
	StraightSlopeUp
	StraightSlopeDown
	StraightSteepSlopeUp
	StraightSteepSlopeDown
	LeftCurveSmallSlopeUp
	RightCurveSmallSlopeUp
	LeftCurveSmallSlopeDown
	RightCurveSmallSlopeDown
	LeftCurveSmallSteepSlopeUp
	RightCurveSmallSteepSlopeUp
	LeftCurveSmallSteepSlopeDown
	RightCurveSmallSteepSlopeDown
	UnkStraight1
	UnkStraight2
	UnkLeftCurveVerySmall1
	UnkLeftCurveVerySmall2
	UnkRightCurveVerySmall1
	UnkRightCurveVerySmall2
	UnkSBendRight
	UnkSBendLeft
	UnkStraightSteepSlopeUp1
	UnkStraightSteepSlopeUp2
	UnkStraightSteepSlopeDown1
	UnkStraightSteepSlopeDown2
	SBendToDualTrack
	SBendToSingleTrack
	UnkSBendToDualTrack
	UnkSBendToSingleTrack
	Turnaround
	UnkTurnaround
)
type RoadId int

const (
	Straight RoadId = iota
	LeftCurveVerySmall
	RightCurveVerySmall
	LeftCurveSmall
	RightCurveSmall
	StraightSlopeUp
	StraightSlopeDown
	StraightSteepSlopeUp
	StraightSteepSlopeDown
	Turnaround
)
// // For some reason we have Common, Track and Road Trait flags
// // but they are all pretty much the same thing. One day we should
// // look into amalgamating them into one.
type CommonTraitFlags int

const (
	None CommonTraitFlags = 0
	Slope CommonTraitFlags = 1 << 0
	SteepSlope CommonTraitFlags = 1 << 1
	CurveSlope CommonTraitFlags = 1 << 2
	Diagonal CommonTraitFlags = 1 << 3
	VerySmallCurve CommonTraitFlags = 1 << 4
	SmallCurve CommonTraitFlags = 1 << 5
	Curve CommonTraitFlags = 1 << 6
	LargeCurve CommonTraitFlags = 1 << 7
	SBendCurve CommonTraitFlags = 1 << 8
	OneSided CommonTraitFlags = 1 << 9
	StartsAtHalfHeight CommonTraitFlags = 1 << 10
	Junction CommonTraitFlags = 1 << 11
	Unk12 CommonTraitFlags = 1 << 12
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(CommonTraitFlags);
type TrackTraitFlags int

const (
	None TrackTraitFlags = 0
	Diagonal TrackTraitFlags = 1 << 0
	LargeCurve TrackTraitFlags = 1 << 1
	NormalCurve TrackTraitFlags = 1 << 2
	SmallCurve TrackTraitFlags = 1 << 3
	VerySmallCurve TrackTraitFlags = 1 << 4
	Slope TrackTraitFlags = 1 << 5
	SteepSlope TrackTraitFlags = 1 << 6
	OneSided TrackTraitFlags = 1 << 7
	SlopedCurve TrackTraitFlags = 1 << 8
	SBend TrackTraitFlags = 1 << 9
	Junction TrackTraitFlags = 1 << 10
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TrackTraitFlags);
type RoadTraitFlags int

const (
	None RoadTraitFlags = 0
	SmallCurve RoadTraitFlags = 1 << 0
	VerySmallCurve RoadTraitFlags = 1 << 1
	Slope RoadTraitFlags = 1 << 2
	SteepSlope RoadTraitFlags = 1 << 3
	Unk4 RoadTraitFlags = 1 << 4
	Turnaround RoadTraitFlags = 1 << 5
	Junction RoadTraitFlags = 1 << 6
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(RoadTraitFlags);
