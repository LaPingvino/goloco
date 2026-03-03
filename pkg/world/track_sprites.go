package world

// trackPiece holds the image offsets for one sequence sub-tile of a track piece.
// img[rotation][layer]: layer 0=Ballast, 1=Sleeper, 2=Rail.
// For non-mergeable (slope) pieces, only layer 0 is valid; layers 1 and 2 are 0.
// All values are offsets relative to trackObj.Image (add to get actual G1 sprite ID).
// 0 means "no sprite / skip".
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrackData.h, TrackObject.h Style0 namespace
type trackPiece struct {
	img          [4][3]uint32
	nonMergeable bool // if true, only img[rot][0] is valid
}

// kTrackParts[trackID] = ordered slice of pieces (indexed by sequenceIndex).
// TrackID map (kTrackPaintParts order in PaintTrackData.h):
//
//	0=Straight, 1=Diagonal, 2=LeftCurveVerySmall, 3=RightCurveVerySmall,
//	4=LeftCurveSmall, 5=RightCurveSmall, 6=LeftCurve, 7=RightCurve,
//	8=LeftCurveLarge(TODO), 9=RightCurveLarge(TODO),
//	10=DiagonalLeftCurveLarge(TODO), 11=DiagonalRightCurveLarge(TODO),
//	12=SBendLeft, 13=SBendRight,
//	14=StraightSlopeUp, 15=StraightSlopeDown,
//	16=StraightSteepSlopeUp, 17=StraightSteepSlopeDown,
//	18=LeftCurveSmallSlopeUp(TODO), 19=RightCurveSmallSlopeUp,
//	20-25=TODO
var kTrackParts = [26][]trackPiece{

	// 0: Straight (1 piece, mergeable)
	// kStraight0: rot 0,2 → NE sprites; rot 1,3 → SE sprites
	{
		{img: [4][3]uint32{
			{18, 20, 22}, // rot0 NE
			{19, 21, 23}, // rot1 SE
			{18, 20, 22}, // rot2 (same as NE)
			{19, 21, 23}, // rot3 (same as SE)
		}},
	},

	// 1: Diagonal (4 pieces, mergeable)
	// kDiagonal0,1,2,3 from PaintTrackData.h
	{
		{img: [4][3]uint32{
			{328, 336, 344}, // rot0
			{332, 340, 348}, // rot1
			{331, 339, 347}, // rot2
			{335, 343, 351}, // rot3
		}},
		{img: [4][3]uint32{
			{330, 338, 346}, // rot0
			{334, 342, 350}, // rot1
			{329, 337, 345}, // rot2
			{333, 341, 349}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kDiagonal1, {2,3,0,1})
			{329, 337, 345}, // rot0
			{333, 341, 349}, // rot1
			{330, 338, 346}, // rot2
			{334, 342, 350}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kDiagonal0, {2,3,0,1})
			{331, 339, 347}, // rot0
			{335, 343, 351}, // rot1
			{328, 336, 344}, // rot2
			{332, 340, 348}, // rot3
		}},
	},

	// 2: LeftCurveVerySmall (1 piece, mergeable)
	// rotateTrackPP(kRightCurveVerySmall0, {1,2,3,0})
	{
		{img: [4][3]uint32{
			{401, 405, 409}, // rot0
			{402, 406, 410}, // rot1
			{403, 407, 411}, // rot2
			{400, 404, 408}, // rot3
		}},
	},

	// 3: RightCurveVerySmall (1 piece, mergeable)
	{
		{img: [4][3]uint32{
			{400, 404, 408}, // rot0 NE
			{401, 405, 409}, // rot1 SE
			{402, 406, 410}, // rot2 SW
			{403, 407, 411}, // rot3 NW
		}},
	},

	// 4: LeftCurveSmall (4 pieces, mergeable)
	// rotateTrackPP(kRightCurveSmall{3,1,2,0}, {1,2,3,0})
	{
		{img: [4][3]uint32{ // rotateTrackPP(RCS3, {1,2,3,0})
			{31, 47, 63}, // rot0
			{35, 51, 67}, // rot1
			{39, 55, 71}, // rot2
			{27, 43, 59}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(RCS1, {1,2,3,0})
			{29, 45, 61}, // rot0
			{33, 49, 65}, // rot1
			{37, 53, 69}, // rot2
			{25, 41, 57}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(RCS2, {1,2,3,0})
			{30, 46, 62}, // rot0
			{34, 50, 66}, // rot1
			{38, 54, 70}, // rot2
			{26, 42, 58}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(RCS0, {1,2,3,0})
			{28, 44, 60}, // rot0
			{32, 48, 64}, // rot1
			{36, 52, 68}, // rot2
			{24, 40, 56}, // rot3
		}},
	},

	// 5: RightCurveSmall (4 pieces, mergeable)
	{
		{img: [4][3]uint32{ // kRightCurveSmall0
			{24, 40, 56}, // rot0 NE
			{28, 44, 60}, // rot1 SE
			{32, 48, 64}, // rot2 SW
			{36, 52, 68}, // rot3 NW
		}},
		{img: [4][3]uint32{ // kRightCurveSmall1
			{25, 41, 57}, // rot0
			{29, 45, 61}, // rot1
			{33, 49, 65}, // rot2
			{37, 53, 69}, // rot3
		}},
		{img: [4][3]uint32{ // kRightCurveSmall2
			{26, 42, 58}, // rot0
			{30, 46, 62}, // rot1
			{34, 50, 66}, // rot2
			{38, 54, 70}, // rot3
		}},
		{img: [4][3]uint32{ // kRightCurveSmall3
			{27, 43, 59}, // rot0
			{31, 47, 63}, // rot1
			{35, 51, 67}, // rot2
			{39, 55, 71}, // rot3
		}},
	},

	// 6: LeftCurve (5 pieces, mergeable)
	// rotateTrackPP(kRightCurve{4,3,2,1,0}, {1,2,3,0})
	{
		{img: [4][3]uint32{ // rotateTrackPP(RC4, {1,2,3,0})
			{145, 165, 185}, // rot0
			{150, 170, 190}, // rot1
			{155, 175, 195}, // rot2
			{140, 160, 180}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(RC3, {1,2,3,0})
			{144, 164, 184}, // rot0
			{149, 169, 189}, // rot1
			{154, 174, 194}, // rot2
			{139, 159, 179}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(RC2, {1,2,3,0})
			{143, 163, 183}, // rot0
			{148, 168, 188}, // rot1
			{153, 173, 193}, // rot2
			{138, 158, 178}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(RC1, {1,2,3,0})
			{142, 162, 182}, // rot0
			{147, 167, 187}, // rot1
			{152, 172, 192}, // rot2
			{137, 157, 177}, // rot3
		}},
		{img: [4][3]uint32{ // rotateTrackPP(RC0, {1,2,3,0})
			{141, 161, 181}, // rot0
			{146, 166, 186}, // rot1
			{151, 171, 191}, // rot2
			{136, 156, 176}, // rot3
		}},
	},

	// 7: RightCurve (5 pieces, mergeable)
	{
		{img: [4][3]uint32{ // kRightCurve0
			{136, 156, 176}, // rot0 NE
			{141, 161, 181}, // rot1 SE
			{146, 166, 186}, // rot2 SW
			{151, 171, 191}, // rot3 NW
		}},
		{img: [4][3]uint32{ // kRightCurve1
			{137, 157, 177}, // rot0
			{142, 162, 182}, // rot1
			{147, 167, 187}, // rot2
			{152, 172, 192}, // rot3
		}},
		{img: [4][3]uint32{ // kRightCurve2
			{138, 158, 178}, // rot0
			{143, 163, 183}, // rot1
			{148, 168, 188}, // rot2
			{153, 173, 193}, // rot3
		}},
		{img: [4][3]uint32{ // kRightCurve3
			{139, 159, 179}, // rot0
			{144, 164, 184}, // rot1
			{149, 169, 189}, // rot2
			{154, 174, 194}, // rot3
		}},
		{img: [4][3]uint32{ // kRightCurve4
			{140, 160, 180}, // rot0
			{145, 165, 185}, // rot1
			{150, 170, 190}, // rot2
			{155, 175, 195}, // rot3
		}},
	},

	// 8: LeftCurveLarge (5 pieces, mergeable)
	// kLeftCurveLarge0-4: Ballast NE=228-232/SE=233-237/SW=238-242/NW=243-247
	//                      Sleeper NE=268-272/SE=273-277/SW=278-282/NW=283-287
	//                      Rail    NE=308-312/SE=313-317/SW=318-322/NW=323-327
	// OpenLoco reference: src/OpenLoco/src/Objects/TrackObject.h Style0 namespace
	{
		{img: [4][3]uint32{
			{228, 268, 308}, // rot0 NE
			{233, 273, 313}, // rot1 SE
			{238, 278, 318}, // rot2 SW
			{243, 283, 323}, // rot3 NW
		}},
		{img: [4][3]uint32{
			{229, 269, 309},
			{234, 274, 314},
			{239, 279, 319},
			{244, 284, 324},
		}},
		{img: [4][3]uint32{
			{230, 270, 310},
			{235, 275, 315},
			{240, 280, 320},
			{245, 285, 325},
		}},
		{img: [4][3]uint32{
			{231, 271, 311},
			{236, 276, 316},
			{241, 281, 321},
			{246, 286, 326},
		}},
		{img: [4][3]uint32{
			{232, 272, 312},
			{237, 277, 317},
			{242, 282, 322},
			{247, 287, 327},
		}},
	},

	// 9: RightCurveLarge (5 pieces, mergeable)
	// kRightCurveLarge0-4: Ballast NE=208-212/SE=213-217/SW=218-222/NW=223-227
	//                       Sleeper NE=248-252/SE=253-257/SW=258-262/NW=263-267
	//                       Rail    NE=288-292/SE=293-297/SW=298-302/NW=303-307
	{
		{img: [4][3]uint32{
			{208, 248, 288}, // rot0 NE
			{213, 253, 293}, // rot1 SE
			{218, 258, 298}, // rot2 SW
			{223, 263, 303}, // rot3 NW
		}},
		{img: [4][3]uint32{
			{209, 249, 289},
			{214, 254, 294},
			{219, 259, 299},
			{224, 264, 304},
		}},
		{img: [4][3]uint32{
			{210, 250, 290},
			{215, 255, 295},
			{220, 260, 300},
			{225, 265, 305},
		}},
		{img: [4][3]uint32{
			{211, 251, 291},
			{216, 256, 296},
			{221, 261, 301},
			{226, 266, 306},
		}},
		{img: [4][3]uint32{
			{212, 252, 292},
			{217, 257, 297},
			{222, 262, 302},
			{227, 267, 307},
		}},
	},

	// 10: DiagonalLeftCurveLarge (5 pieces, mergeable)
	// = rotateTrackPP(kRightCurveLarge{4,2,3,1,0}, kRotationTable2301={2,3,0,1})
	// result.rot[i] = source.rot[table[i]]
	// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrackData.h kDiagonalLeftCurveLargeTPP
	{
		{img: [4][3]uint32{ // rotateTrackPP(kRCL4, {2,3,0,1})
			{222, 262, 302}, // rot0 = RCL4.rot2
			{227, 267, 307}, // rot1 = RCL4.rot3
			{212, 252, 292}, // rot2 = RCL4.rot0
			{217, 257, 297}, // rot3 = RCL4.rot1
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kRCL2, {2,3,0,1})
			{220, 260, 300},
			{225, 265, 305},
			{210, 250, 290},
			{215, 255, 295},
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kRCL3, {2,3,0,1})
			{221, 261, 301},
			{226, 266, 306},
			{211, 251, 291},
			{216, 256, 296},
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kRCL1, {2,3,0,1})
			{219, 259, 299},
			{224, 264, 304},
			{209, 249, 289},
			{214, 254, 294},
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kRCL0, {2,3,0,1})
			{218, 258, 298},
			{223, 263, 303},
			{208, 248, 288},
			{213, 253, 293},
		}},
	},

	// 11: DiagonalRightCurveLarge (5 pieces, mergeable)
	// = rotateTrackPP(kLeftCurveLarge{4,2,3,1,0}, kRotationTable3012={3,0,1,2})
	// result.rot[i] = source.rot[table[i]]
	// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrackData.h kDiagonalRightCurveLargeTPP
	{
		{img: [4][3]uint32{ // rotateTrackPP(kLCL4, {3,0,1,2})
			{247, 287, 327}, // rot0 = LCL4.rot3
			{232, 272, 312}, // rot1 = LCL4.rot0
			{237, 277, 317}, // rot2 = LCL4.rot1
			{242, 282, 322}, // rot3 = LCL4.rot2
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kLCL2, {3,0,1,2})
			{245, 285, 325},
			{230, 270, 310},
			{235, 275, 315},
			{240, 280, 320},
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kLCL3, {3,0,1,2})
			{246, 286, 326},
			{231, 271, 311},
			{236, 276, 316},
			{241, 281, 321},
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kLCL1, {3,0,1,2})
			{244, 284, 324},
			{229, 269, 309},
			{234, 274, 314},
			{239, 279, 319},
		}},
		{img: [4][3]uint32{ // rotateTrackPP(kLCL0, {3,0,1,2})
			{243, 283, 323},
			{228, 268, 308},
			{233, 273, 313},
			{238, 278, 318},
		}},
	},

	// 12: SBendLeft (4 pieces, mergeable)
	{
		{img: [4][3]uint32{ // kSBendLeft0
			{352, 368, 384}, // rot0 NE
			{356, 372, 388}, // rot1 SE
			{355, 371, 387}, // rot2 (uses SBendLeft3 images)
			{359, 375, 391}, // rot3
		}},
		{img: [4][3]uint32{ // kSBendLeft1
			{353, 369, 385}, // rot0
			{357, 373, 389}, // rot1
			{354, 370, 386}, // rot2 (uses SBendLeft2 images)
			{358, 374, 390}, // rot3
		}},
		{img: [4][3]uint32{ // kSBendLeft2 = rotateTrackPP(kSBendLeft1, {2,3,0,1})
			{354, 370, 386}, // rot0
			{358, 374, 390}, // rot1
			{353, 369, 385}, // rot2
			{357, 373, 389}, // rot3
		}},
		{img: [4][3]uint32{ // kSBendLeft3 = rotateTrackPP(kSBendLeft0, {2,3,0,1})
			{355, 371, 387}, // rot0
			{359, 375, 391}, // rot1
			{352, 368, 384}, // rot2
			{356, 372, 388}, // rot3
		}},
	},

	// 13: SBendRight (4 pieces, mergeable)
	{
		{img: [4][3]uint32{ // kSBendRight0
			{360, 376, 392}, // rot0 NE
			{364, 380, 396}, // rot1 SE
			{363, 379, 395}, // rot2 (uses SBendRight3 images)
			{367, 383, 399}, // rot3
		}},
		{img: [4][3]uint32{ // kSBendRight1
			{361, 377, 393}, // rot0
			{365, 381, 397}, // rot1
			{362, 378, 394}, // rot2 (uses SBendRight2 images)
			{366, 382, 398}, // rot3
		}},
		{img: [4][3]uint32{ // kSBendRight2 = rotateTrackPP(kSBendRight1, {2,3,0,1})
			{362, 378, 394}, // rot0
			{366, 382, 398}, // rot1
			{361, 377, 393}, // rot2
			{365, 381, 397}, // rot3
		}},
		{img: [4][3]uint32{ // kSBendRight3 = rotateTrackPP(kSBendRight0, {2,3,0,1})
			{363, 379, 395}, // rot0
			{367, 383, 399}, // rot1
			{360, 376, 392}, // rot2
			{364, 380, 396}, // rot3
		}},
	},

	// 14: StraightSlopeUp (2 pieces, non-mergeable)
	// kStraightSlopeUp0NE=196, SE=198, SW=200, NW=202
	{
		{img: [4][3]uint32{{196}, {198}, {200}, {202}}, nonMergeable: true},
		{img: [4][3]uint32{{197}, {199}, {201}, {203}}, nonMergeable: true},
	},

	// 15: StraightSlopeDown (2 pieces, non-mergeable)
	// rotateTrackPP(kStraightSlopeUp{1,0}, {2,3,0,1})
	{
		{img: [4][3]uint32{{201}, {203}, {197}, {199}}, nonMergeable: true},
		{img: [4][3]uint32{{200}, {202}, {196}, {198}}, nonMergeable: true},
	},

	// 16: StraightSteepSlopeUp (1 piece, non-mergeable)
	// kStraightSteepSlopeUp0NE=204, SE=205, SW=206, NW=207
	{
		{img: [4][3]uint32{{204}, {205}, {206}, {207}}, nonMergeable: true},
	},

	// 17: StraightSteepSlopeDown (1 piece, non-mergeable)
	// rotateTrackPP(kStraightSteepSlopeUp0, {2,3,0,1})
	{
		{img: [4][3]uint32{{206}, {207}, {204}, {205}}, nonMergeable: true},
	},

	// 18: LeftCurveSmallSlopeUp (4 pieces, non-mergeable)
	// = rotateTrackPP(kRightCurveSmallSlopeDown{3,1,2,0}, kRotationTable1230={1,2,3,0})
	// result.rot[i] = source.rot[table[i]]; source uses kRCSDown0-3 NE=88-91/SE=92-95/SW=96-99/NW=100-103
	// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrackData.h kLeftCurveSmallSlopeUpTPP
	{
		{img: [4][3]uint32{{95}, {99}, {103}, {91}}, nonMergeable: true},  // rotateTrackPP(kRCSDown3, 1230)
		{img: [4][3]uint32{{93}, {97}, {101}, {89}}, nonMergeable: true},  // rotateTrackPP(kRCSDown1, 1230)
		{img: [4][3]uint32{{94}, {98}, {102}, {90}}, nonMergeable: true},  // rotateTrackPP(kRCSDown2, 1230)
		{img: [4][3]uint32{{92}, {96}, {100}, {88}}, nonMergeable: true},  // rotateTrackPP(kRCSDown0, 1230)
	},

	// 19: RightCurveSmallSlopeUp (4 pieces, non-mergeable)
	// kRightCurveSmallSlopeUp: NE=72-75, SE=76-79, SW=80-83, NW=84-87
	{
		{img: [4][3]uint32{{72}, {76}, {80}, {84}}, nonMergeable: true},
		{img: [4][3]uint32{{73}, {77}, {81}, {85}}, nonMergeable: true},
		{img: [4][3]uint32{{74}, {78}, {82}, {86}}, nonMergeable: true},
		{img: [4][3]uint32{{75}, {79}, {83}, {87}}, nonMergeable: true},
	},

	// 20: LeftCurveSmallSlopeDown (4 pieces, non-mergeable)
	// = rotateTrackPP(kRightCurveSmallSlopeUp{3,1,2,0}, kRotationTable1230={1,2,3,0})
	// source uses kRCSUp0-3 NE=72-75/SE=76-79/SW=80-83/NW=84-87
	// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrackData.h kLeftCurveSmallSlopeDownTPP
	{
		{img: [4][3]uint32{{79}, {83}, {87}, {75}}, nonMergeable: true},  // rotateTrackPP(kRCSUp3, 1230)
		{img: [4][3]uint32{{77}, {81}, {85}, {73}}, nonMergeable: true},  // rotateTrackPP(kRCSUp1, 1230)
		{img: [4][3]uint32{{78}, {82}, {86}, {74}}, nonMergeable: true},  // rotateTrackPP(kRCSUp2, 1230)
		{img: [4][3]uint32{{76}, {80}, {84}, {72}}, nonMergeable: true},  // rotateTrackPP(kRCSUp0, 1230)
	},

	// 21: RightCurveSmallSlopeDown (4 pieces, non-mergeable)
	// kRightCurveSmallSlopeDown: NE=88-91/SE=92-95/SW=96-99/NW=100-103
	// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrackData.h kRightCurveSmallSlopeDownTPP
	{
		{img: [4][3]uint32{{88}, {92}, {96}, {100}}, nonMergeable: true},
		{img: [4][3]uint32{{89}, {93}, {97}, {101}}, nonMergeable: true},
		{img: [4][3]uint32{{90}, {94}, {98}, {102}}, nonMergeable: true},
		{img: [4][3]uint32{{91}, {95}, {99}, {103}}, nonMergeable: true},
	},

	// 22: LeftCurveSmallSteepSlopeUp (4 pieces, non-mergeable)
	// = rotateTrackPP(kRightCurveSmallSteepSlopeDown{3,1,2,0}, kRotationTable1230={1,2,3,0})
	// source uses kRCSteepDown0-3 NE=120-123/SE=124-127/SW=128-131/NW=132-135
	// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrackData.h kLeftCurveSmallSteepSlopeUpTPP
	{
		{img: [4][3]uint32{{127}, {131}, {135}, {123}}, nonMergeable: true}, // rotateTrackPP(kRCSteepDown3, 1230)
		{img: [4][3]uint32{{125}, {129}, {133}, {121}}, nonMergeable: true}, // rotateTrackPP(kRCSteepDown1, 1230)
		{img: [4][3]uint32{{126}, {130}, {134}, {122}}, nonMergeable: true}, // rotateTrackPP(kRCSteepDown2, 1230)
		{img: [4][3]uint32{{124}, {128}, {132}, {120}}, nonMergeable: true}, // rotateTrackPP(kRCSteepDown0, 1230)
	},

	// 23: RightCurveSmallSteepSlopeUp (4 pieces, non-mergeable)
	// kRightCurveSmallSteepSlopeUp: NE=104-107, SE=108-111, SW=112-115, NW=116-119
	{
		{img: [4][3]uint32{{104}, {108}, {112}, {116}}, nonMergeable: true},
		{img: [4][3]uint32{{105}, {109}, {113}, {117}}, nonMergeable: true},
		{img: [4][3]uint32{{106}, {110}, {114}, {118}}, nonMergeable: true},
		{img: [4][3]uint32{{107}, {111}, {115}, {119}}, nonMergeable: true},
	},

	// 24: LeftCurveSmallSteepSlopeDown (4 pieces, non-mergeable)
	// = rotateTrackPP(kRightCurveSmallSteepSlopeUp{3,1,2,0}, kRotationTable1230={1,2,3,0})
	// source uses kRCSteepUp0-3 NE=104-107/SE=108-111/SW=112-115/NW=116-119
	// OpenLoco reference: src/OpenLoco/src/Paint/PaintTrackData.h kLeftCurveSmallSteepSlopeDownTPP
	{
		{img: [4][3]uint32{{111}, {115}, {119}, {107}}, nonMergeable: true}, // rotateTrackPP(kRCSteepUp3, 1230)
		{img: [4][3]uint32{{109}, {113}, {117}, {105}}, nonMergeable: true}, // rotateTrackPP(kRCSteepUp1, 1230)
		{img: [4][3]uint32{{110}, {114}, {118}, {106}}, nonMergeable: true}, // rotateTrackPP(kRCSteepUp2, 1230)
		{img: [4][3]uint32{{108}, {112}, {116}, {104}}, nonMergeable: true}, // rotateTrackPP(kRCSteepUp0, 1230)
	},

	// 25: RightCurveSmallSteepSlopeDown (4 pieces, non-mergeable)
	// kRightCurveSmallSteepSlopeDown: NE=120-123, SE=124-127, SW=128-131, NW=132-135
	{
		{img: [4][3]uint32{{120}, {124}, {128}, {132}}, nonMergeable: true},
		{img: [4][3]uint32{{121}, {125}, {129}, {133}}, nonMergeable: true},
		{img: [4][3]uint32{{122}, {126}, {130}, {134}}, nonMergeable: true},
		{img: [4][3]uint32{{123}, {127}, {131}, {135}}, nonMergeable: true},
	},
}
