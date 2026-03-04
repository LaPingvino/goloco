package world

// roadPiece0 holds the image offset for one Style0 road sub-tile piece (single image per rotation).
// img[rotation] = image offset relative to roadObj.Image.
// 0 means "no road-object sprite" (either a global hit-test image or unimplemented) — skip.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintRoadStyle0Data.h, RoadObject.h Style0 namespace
type roadPiece0 struct {
	img [4]uint32
}

// roadPiece1 holds the image offsets for one Style1 road sub-tile piece (3 layers per rotation).
// img[rotation][layer]: layer 0=Ballast, 1=Sleeper, 2=Rail.
// All values are offsets relative to roadObj.Image.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintRoadStyle1Data.h, RoadObject.h Style1 namespace
type roadPiece1 struct {
	img [4][3]uint32
}

// Road ID → TPP array mapping (same order for Style0, Style1, Style2):
//
//	0=Straight, 1=LeftCurveVerySmall, 2=RightCurveVerySmall,
//	3=LeftCurveSmall, 4=RightCurveSmall,
//	5=StraightSlopeUp, 6=StraightSlopeDown,
//	7=StraightSteepSlopeUp, 8=StraightSteepSlopeDown,
//	9=Turnaround

// kRoadPartsStyle0[roadID] = ordered slice of pieces (indexed by sequenceIndex).
// Style0: single image per rotation (no Ballast/Sleeper/Rail layering).
// roadID 0 (Straight) uses global hit-test G1 IDs — not road-object offsets — so all 0 (skip).
var kRoadPartsStyle0 = [10][]roadPiece0{

	// 0: Straight — uses ImageIds::road_hit_test_straight_NE (global G1), NOT roadObj offset
	{{img: [4]uint32{0, 0, 0, 0}}},

	// 1: LeftCurveVerySmall
	// rotateRoadPP(kRightCurveVerySmall0, {1,2,3,0})
	// kRightCurveVerySmall0: NE=36, SE=37, SW=38, NW=39
	{{img: [4]uint32{37, 38, 39, 36}}},

	// 2: RightCurveVerySmall
	{{img: [4]uint32{36, 37, 38, 39}}},

	// 3: LeftCurveSmall (4 pieces)
	// rotateRoadPP(kRightCurveSmall{3,1,2,0}, {1,2,3,0})
	{
		{img: [4]uint32{52, 56, 60, 48}}, // rotateRoadPP(RCS3, {1,2,3,0}): RCS3={48,52,56,60}
		{img: [4]uint32{50, 54, 58, 46}}, // rotateRoadPP(RCS1, {1,2,3,0}): RCS1={46,50,54,58}
		{img: [4]uint32{51, 55, 59, 47}}, // rotateRoadPP(RCS2, {1,2,3,0}): RCS2={47,51,55,59}
		{img: [4]uint32{49, 53, 57, 45}}, // rotateRoadPP(RCS0, {1,2,3,0}): RCS0={45,49,53,57}
	},

	// 4: RightCurveSmall (4 pieces)
	// kRightCurveSmall0NE=45,SE=49,SW=53,NW=57; 1NE=46...; 2NE=47...; 3NE=48...
	{
		{img: [4]uint32{45, 49, 53, 57}},
		{img: [4]uint32{46, 50, 54, 58}},
		{img: [4]uint32{47, 51, 55, 59}},
		{img: [4]uint32{48, 52, 56, 60}},
	},

	// 5: StraightSlopeUp (2 pieces)
	// kStraightSlopeUp0NE=61, SE=63, SW=65, NW=67; piece1: NE=62,SE=64,SW=66,NW=68
	{
		{img: [4]uint32{61, 63, 65, 67}},
		{img: [4]uint32{62, 64, 66, 68}},
	},

	// 6: StraightSlopeDown (2 pieces)
	// rotateRoadPP(kStraightSlopeUp{1,0}, {2,3,0,1})
	{
		{img: [4]uint32{66, 68, 62, 64}},
		{img: [4]uint32{65, 67, 61, 63}},
	},

	// 7: StraightSteepSlopeUp (1 piece)
	// kStraightSteepSlopeUp0NE=69, SE=70, SW=71, NW=72
	{{img: [4]uint32{69, 70, 71, 72}}},

	// 8: StraightSteepSlopeDown (1 piece)
	// rotateRoadPP(kStraightSteepSlopeUp0, {2,3,0,1})
	{{img: [4]uint32{71, 72, 69, 70}}},

	// 9: Turnaround (1 piece)
	// kTurnaround0NE=73, SE=74, SW=75, NW=76
	{{img: [4]uint32{73, 74, 75, 76}}},
}

// kRoadPartsStyle1[roadID] = ordered slice of pieces (indexed by sequenceIndex).
// Style1: 3 layers (Ballast, Sleeper, Rail) per rotation — like Track Style0.
var kRoadPartsStyle1 = [10][]roadPiece1{

	// 0: Straight (1 piece)
	// kStraight0BallastNE=34,SleeperNE=36,RailNE=38; SE: Ballast=35,Sleeper=37,Rail=39
	{
		{img: [4][3]uint32{
			{34, 36, 38}, // rot0 NE
			{35, 37, 39}, // rot1 SE
			{34, 36, 38}, // rot2 (same as NE)
			{35, 37, 39}, // rot3 (same as SE)
		}},
	},

	// 1: LeftCurveVerySmall (1 piece)
	// rotateRoadPP(kRightCurveVerySmall0, {1,2,3,0})
	// kRightCurveVerySmall0: NE={124,128,132}, SE={125,129,133}, SW={126,130,134}, NW={127,131,135}
	{
		{img: [4][3]uint32{
			{125, 129, 133}, // rot0 (was SE)
			{126, 130, 134}, // rot1 (was SW)
			{127, 131, 135}, // rot2 (was NW)
			{124, 128, 132}, // rot3 (was NE)
		}},
	},

	// 2: RightCurveVerySmall (1 piece)
	{
		{img: [4][3]uint32{
			{124, 128, 132}, // rot0 NE
			{125, 129, 133}, // rot1 SE
			{126, 130, 134}, // rot2 SW
			{127, 131, 135}, // rot3 NW
		}},
	},

	// 3: LeftCurveSmall (4 pieces)
	// rotateRoadPP(kRightCurveSmall{3,1,2,0}, {1,2,3,0})
	// RCS0: NE={40,56,72}, SE={44,60,76}, SW={48,64,80}, NW={52,68,84}
	// RCS1: NE={41,57,73}, SE={45,61,77}, SW={49,65,81}, NW={53,69,85}
	// RCS2: NE={42,58,74}, SE={46,62,78}, SW={50,66,82}, NW={54,70,86}
	// RCS3: NE={43,59,75}, SE={47,63,79}, SW={51,67,83}, NW={55,71,87}
	{
		{img: [4][3]uint32{ // rotateRoadPP(RCS3, {1,2,3,0})
			{47, 63, 79}, // rot0 (was RCS3 SE)
			{51, 67, 83}, // rot1 (was RCS3 SW)
			{55, 71, 87}, // rot2 (was RCS3 NW)
			{43, 59, 75}, // rot3 (was RCS3 NE)
		}},
		{img: [4][3]uint32{ // rotateRoadPP(RCS1, {1,2,3,0})
			{45, 61, 77}, // rot0
			{49, 65, 81}, // rot1
			{53, 69, 85}, // rot2
			{41, 57, 73}, // rot3
		}},
		{img: [4][3]uint32{ // rotateRoadPP(RCS2, {1,2,3,0})
			{46, 62, 78}, // rot0
			{50, 66, 82}, // rot1
			{54, 70, 86}, // rot2
			{42, 58, 74}, // rot3
		}},
		{img: [4][3]uint32{ // rotateRoadPP(RCS0, {1,2,3,0})
			{44, 60, 76}, // rot0
			{48, 64, 80}, // rot1
			{52, 68, 84}, // rot2
			{40, 56, 72}, // rot3
		}},
	},

	// 4: RightCurveSmall (4 pieces)
	{
		{img: [4][3]uint32{ // kRightCurveSmall0
			{40, 56, 72}, // rot0 NE
			{44, 60, 76}, // rot1 SE
			{48, 64, 80}, // rot2 SW
			{52, 68, 84}, // rot3 NW
		}},
		{img: [4][3]uint32{ // kRightCurveSmall1
			{41, 57, 73},
			{45, 61, 77},
			{49, 65, 81},
			{53, 69, 85},
		}},
		{img: [4][3]uint32{ // kRightCurveSmall2
			{42, 58, 74},
			{46, 62, 78},
			{50, 66, 82},
			{54, 70, 86},
		}},
		{img: [4][3]uint32{ // kRightCurveSmall3
			{43, 59, 75},
			{47, 63, 79},
			{51, 67, 83},
			{55, 71, 87},
		}},
	},

	// 5: StraightSlopeUp (2 pieces)
	// kStraightSlopeUp0: NE={88,96,104}, SE={90,98,106}, SW={92,100,108}, NW={94,102,110}
	// kStraightSlopeUp1: NE={89,97,105}, SE={91,99,107}, SW={93,101,109}, NW={95,103,111}
	{
		{img: [4][3]uint32{
			{88, 96, 104},
			{90, 98, 106},
			{92, 100, 108},
			{94, 102, 110},
		}},
		{img: [4][3]uint32{
			{89, 97, 105},
			{91, 99, 107},
			{93, 101, 109},
			{95, 103, 111},
		}},
	},

	// 6: StraightSlopeDown (2 pieces)
	// rotateRoadPP(kStraightSlopeUp{1,0}, {2,3,0,1})
	{
		{img: [4][3]uint32{
			{93, 101, 109}, // rot0 (was SlopeUp1 rot2)
			{95, 103, 111}, // rot1 (was SlopeUp1 rot3)
			{89, 97, 105},  // rot2 (was SlopeUp1 rot0)
			{91, 99, 107},  // rot3 (was SlopeUp1 rot1)
		}},
		{img: [4][3]uint32{
			{92, 100, 108},
			{94, 102, 110},
			{88, 96, 104},
			{90, 98, 106},
		}},
	},

	// 7: StraightSteepSlopeUp (1 piece)
	// NE={112,116,120}, SE={113,117,121}, SW={114,118,122}, NW={115,119,123}
	{
		{img: [4][3]uint32{
			{112, 116, 120},
			{113, 117, 121},
			{114, 118, 122},
			{115, 119, 123},
		}},
	},

	// 8: StraightSteepSlopeDown (1 piece)
	// rotateRoadPP(kStraightSteepSlopeUp0, {2,3,0,1})
	{
		{img: [4][3]uint32{
			{114, 118, 122},
			{115, 119, 123},
			{112, 116, 120},
			{113, 117, 121},
		}},
	},

	// 9: Turnaround (1 piece)
	// NE={136,140,144}, SE={137,141,145}, SW={138,142,146}, NW={139,143,147}
	{
		{img: [4][3]uint32{
			{136, 140, 144},
			{137, 141, 145},
			{138, 142, 146},
			{139, 143, 147},
		}},
	},
}

// kRoadPartsStyle2[roadID] = ordered slice of pieces (indexed by sequenceIndex).
// Style2: single image per rotation (same rendering as Style0 but DIFFERENT offsets).
// All values are offsets relative to roadObj.Image.
// Rotation order: [0]=NE, [1]=SE, [2]=SW, [3]=NW.
//
// OpenLoco reference: src/OpenLoco/src/Objects/RoadObject.h  Style2 namespace
// src/OpenLoco/src/Paint/PaintRoad.cpp  Style02::paintRoadPP
var kRoadPartsStyle2 = [10][]roadPiece0{

	// 0: Straight — kStraight0: NE=34, SE=35, SW=85, NW=86
	{{img: [4]uint32{34, 35, 85, 86}}},

	// 1: LeftCurveVerySmall — kLeftCurveVerySmall0: NE=37, SE=38, SW=39, NW=36
	{{img: [4]uint32{37, 38, 39, 36}}},

	// 2: RightCurveVerySmall — kRightCurveVerySmall0: NE=87, SE=88, SW=89, NW=90
	{{img: [4]uint32{87, 88, 89, 90}}},

	// 3: LeftCurveSmall (4 pieces)
	// Piece ordering: seqIdx 0=LCS3, 1=LCS1, 2=LCS2, 3=LCS0 (matches Style0 convention)
	// LCS0: NE=52,SE=56,SW=60,NW=48  LCS1: NE=50,SE=54,SW=58,NW=46
	// LCS2: NE=51,SE=55,SW=59,NW=47  LCS3: NE=49,SE=53,SW=57,NW=45
	{
		{img: [4]uint32{49, 53, 57, 45}}, // seqIdx 0 = LCS3
		{img: [4]uint32{50, 54, 58, 46}}, // seqIdx 1 = LCS1
		{img: [4]uint32{51, 55, 59, 47}}, // seqIdx 2 = LCS2
		{img: [4]uint32{52, 56, 60, 48}}, // seqIdx 3 = LCS0
	},

	// 4: RightCurveSmall (4 pieces)
	// RCS0: NE=96,SE=100,SW=104,NW=108  RCS1: NE=97,SE=101,SW=105,NW=109
	// RCS2: NE=98,SE=102,SW=106,NW=110  RCS3: NE=99,SE=103,SW=107,NW=111
	{
		{img: [4]uint32{96, 100, 104, 108}},  // seqIdx 0 = RCS0
		{img: [4]uint32{97, 101, 105, 109}},  // seqIdx 1 = RCS1
		{img: [4]uint32{98, 102, 106, 110}},  // seqIdx 2 = RCS2
		{img: [4]uint32{99, 103, 107, 111}},  // seqIdx 3 = RCS3
	},

	// 5: StraightSlopeUp (2 pieces)
	// SSU0: NE=61,SE=63,SW=65,NW=67  SSU1: NE=62,SE=64,SW=66,NW=68
	{
		{img: [4]uint32{61, 63, 65, 67}},
		{img: [4]uint32{62, 64, 66, 68}},
	},

	// 6: StraightSlopeDown (2 pieces) — explicit (not rotated from SlopeUp)
	// SSD0: NE=117,SE=119,SW=113,NW=115  SSD1: NE=116,SE=118,SW=112,NW=114
	{
		{img: [4]uint32{117, 119, 113, 115}},
		{img: [4]uint32{116, 118, 112, 114}},
	},

	// 7: StraightSteepSlopeUp (1 piece)
	// NE=69, SE=70, SW=71, NW=72
	{{img: [4]uint32{69, 70, 71, 72}}},

	// 8: StraightSteepSlopeDown (1 piece) — explicit
	// SW=120, NW=121, NE=122, SE=123
	{{img: [4]uint32{122, 123, 120, 121}}},

	// 9: Turnaround (1 piece)
	// NE=73, SE=74, SW=75, NW=76
	{{img: [4]uint32{73, 74, 75, 76}}},
}
