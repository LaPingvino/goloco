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
type TrackExtraObject struct {
	Name StringId
	// World::Track::TrackTraitFlags trackPieces; // 0x02
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

// MALFORMED FIELD: const TrackExtraObjectObjectType any = ObjectType.trackExtra

// static_assert(sizeof(TrackExtraObject) == 0x12);
// namespace TrackExtraObj::ImageIds
// // Note: Style imageIds are relative to 0x0A so you need to +8 to get its
// // real id relative to object at rest
// namespace Style0
// MALFORMED FIELD: const Straight0NE uint32 = 0
// MALFORMED FIELD: const Straight0SE uint32 = 1
// MALFORMED FIELD: const Straight0SW uint32 = 2
// MALFORMED FIELD: const Straight0NW uint32 = 3
// MALFORMED FIELD: const RightCurveSmall0NE uint32 = 4
// MALFORMED FIELD: const RightCurveSmall1NE uint32 = 5
// MALFORMED FIELD: const RightCurveSmall2NE uint32 = 6
// MALFORMED FIELD: const RightCurveSmall3NE uint32 = 7
// MALFORMED FIELD: const RightCurveSmall0SE uint32 = 8
// MALFORMED FIELD: const RightCurveSmall1SE uint32 = 9
// MALFORMED FIELD: const RightCurveSmall2SE uint32 = 10
// MALFORMED FIELD: const RightCurveSmall3SE uint32 = 11
// MALFORMED FIELD: const RightCurveSmall0SW uint32 = 12
// MALFORMED FIELD: const RightCurveSmall1SW uint32 = 13
// MALFORMED FIELD: const RightCurveSmall2SW uint32 = 14
// MALFORMED FIELD: const RightCurveSmall3SW uint32 = 15
// MALFORMED FIELD: const RightCurveSmall0NW uint32 = 16
// MALFORMED FIELD: const RightCurveSmall1NW uint32 = 17
// MALFORMED FIELD: const RightCurveSmall2NW uint32 = 18
// MALFORMED FIELD: const RightCurveSmall3NW uint32 = 19
// MALFORMED FIELD: const RightCurve0NE uint32 = 20
// MALFORMED FIELD: const RightCurve1NE uint32 = 21
// MALFORMED FIELD: const RightCurve2NE uint32 = 22
// MALFORMED FIELD: const RightCurve3NE uint32 = 23
// MALFORMED FIELD: const RightCurve4NE uint32 = 24
// MALFORMED FIELD: const RightCurve0SE uint32 = 25
// MALFORMED FIELD: const RightCurve1SE uint32 = 26
// MALFORMED FIELD: const RightCurve2SE uint32 = 27
// MALFORMED FIELD: const RightCurve3SE uint32 = 28
// MALFORMED FIELD: const RightCurve4SE uint32 = 29
// MALFORMED FIELD: const RightCurve0SW uint32 = 30
// MALFORMED FIELD: const RightCurve1SW uint32 = 31
// MALFORMED FIELD: const RightCurve2SW uint32 = 32
// MALFORMED FIELD: const RightCurve3SW uint32 = 33
// MALFORMED FIELD: const RightCurve4SW uint32 = 34
// MALFORMED FIELD: const RightCurve0NW uint32 = 35
// MALFORMED FIELD: const RightCurve1NW uint32 = 36
// MALFORMED FIELD: const RightCurve2NW uint32 = 37
// MALFORMED FIELD: const RightCurve3NW uint32 = 38
// MALFORMED FIELD: const RightCurve4NW uint32 = 39
// MALFORMED FIELD: const SBendLeft0NE uint32 = 40
// MALFORMED FIELD: const SBendLeft1NE uint32 = 41
// MALFORMED FIELD: const SBendLeft2NE uint32 = 42
// MALFORMED FIELD: const SBendLeft3NE uint32 = 43
// MALFORMED FIELD: const SBendLeft0SE uint32 = 44
// MALFORMED FIELD: const SBendLeft1SE uint32 = 45
// MALFORMED FIELD: const SBendLeft2SE uint32 = 46
// MALFORMED FIELD: const SBendLeft3SE uint32 = 47
// MALFORMED FIELD: const SBendLeft3SW uint32 = 48
// MALFORMED FIELD: const SBendLeft2SW uint32 = 49
// MALFORMED FIELD: const SBendLeft1SW uint32 = 50
// MALFORMED FIELD: const SBendLeft0SW uint32 = 51
// MALFORMED FIELD: const SBendLeft3NW uint32 = 52
// MALFORMED FIELD: const SBendLeft2NW uint32 = 53
// MALFORMED FIELD: const SBendLeft1NW uint32 = 54
// MALFORMED FIELD: const SBendLeft0NW uint32 = 55
// MALFORMED FIELD: const SBendRight0NE uint32 = 56
// MALFORMED FIELD: const SBendRight1NE uint32 = 57
// MALFORMED FIELD: const SBendRight2NE uint32 = 58
// MALFORMED FIELD: const SBendRight3NE uint32 = 59
// MALFORMED FIELD: const SBendRight0SE uint32 = 60
// MALFORMED FIELD: const SBendRight1SE uint32 = 61
// MALFORMED FIELD: const SBendRight2SE uint32 = 62
// MALFORMED FIELD: const SBendRight3SE uint32 = 63
// MALFORMED FIELD: const SBendRight3SW uint32 = 64
// MALFORMED FIELD: const SBendRight2SW uint32 = 65
// MALFORMED FIELD: const SBendRight1SW uint32 = 66
// MALFORMED FIELD: const SBendRight0SW uint32 = 67
// MALFORMED FIELD: const SBendRight3NW uint32 = 68
// MALFORMED FIELD: const SBendRight2NW uint32 = 69
// MALFORMED FIELD: const SBendRight1NW uint32 = 70
// MALFORMED FIELD: const SBendRight0NW uint32 = 71
// MALFORMED FIELD: const StraightSlopeUp0NE uint32 = 72
// MALFORMED FIELD: const StraightSlopeUp1NE uint32 = 73
// MALFORMED FIELD: const StraightSlopeUp0SE uint32 = 74
// MALFORMED FIELD: const StraightSlopeUp1SE uint32 = 75
// MALFORMED FIELD: const StraightSlopeUp0SW uint32 = 76
// MALFORMED FIELD: const StraightSlopeUp1SW uint32 = 77
// MALFORMED FIELD: const StraightSlopeUp0NW uint32 = 78
// MALFORMED FIELD: const StraightSlopeUp1NW uint32 = 79
// MALFORMED FIELD: const StraightSteepSlopeUp0NE uint32 = 80
// MALFORMED FIELD: const StraightSteepSlopeUp0SE uint32 = 81
// MALFORMED FIELD: const StraightSteepSlopeUp0SW uint32 = 82
// MALFORMED FIELD: const StraightSteepSlopeUp0NW uint32 = 83
// MALFORMED FIELD: const RightCurveSmallSlopeUp0NE uint32 = 84
// MALFORMED FIELD: const RightCurveSmallSlopeUp1NE uint32 = 85
// MALFORMED FIELD: const RightCurveSmallSlopeUp2NE uint32 = 86
// MALFORMED FIELD: const RightCurveSmallSlopeUp3NE uint32 = 87
// MALFORMED FIELD: const RightCurveSmallSlopeUp0SE uint32 = 88
// MALFORMED FIELD: const RightCurveSmallSlopeUp1SE uint32 = 89
// MALFORMED FIELD: const RightCurveSmallSlopeUp2SE uint32 = 90
// MALFORMED FIELD: const RightCurveSmallSlopeUp3SE uint32 = 91
// MALFORMED FIELD: const RightCurveSmallSlopeUp0SW uint32 = 92
// MALFORMED FIELD: const RightCurveSmallSlopeUp1SW uint32 = 93
// MALFORMED FIELD: const RightCurveSmallSlopeUp2SW uint32 = 94
// MALFORMED FIELD: const RightCurveSmallSlopeUp3SW uint32 = 95
// MALFORMED FIELD: const RightCurveSmallSlopeUp0NW uint32 = 96
// MALFORMED FIELD: const RightCurveSmallSlopeUp1NW uint32 = 97
// MALFORMED FIELD: const RightCurveSmallSlopeUp2NW uint32 = 98
// MALFORMED FIELD: const RightCurveSmallSlopeUp3NW uint32 = 99
// MALFORMED FIELD: const RightCurveSmallSlopeDown0NE uint32 = 100
// MALFORMED FIELD: const RightCurveSmallSlopeDown1NE uint32 = 101
// MALFORMED FIELD: const RightCurveSmallSlopeDown2NE uint32 = 102
// MALFORMED FIELD: const RightCurveSmallSlopeDown3NE uint32 = 103
// MALFORMED FIELD: const RightCurveSmallSlopeDown0SE uint32 = 104
// MALFORMED FIELD: const RightCurveSmallSlopeDown1SE uint32 = 105
// MALFORMED FIELD: const RightCurveSmallSlopeDown2SE uint32 = 106
// MALFORMED FIELD: const RightCurveSmallSlopeDown3SE uint32 = 107
// MALFORMED FIELD: const RightCurveSmallSlopeDown0SW uint32 = 108
// MALFORMED FIELD: const RightCurveSmallSlopeDown1SW uint32 = 109
// MALFORMED FIELD: const RightCurveSmallSlopeDown2SW uint32 = 110
// MALFORMED FIELD: const RightCurveSmallSlopeDown3SW uint32 = 111
// MALFORMED FIELD: const RightCurveSmallSlopeDown0NW uint32 = 112
// MALFORMED FIELD: const RightCurveSmallSlopeDown1NW uint32 = 113
// MALFORMED FIELD: const RightCurveSmallSlopeDown2NW uint32 = 114
// MALFORMED FIELD: const RightCurveSmallSlopeDown3NW uint32 = 115
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp0NE uint32 = 116
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp1NE uint32 = 117
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp2NE uint32 = 118
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp3NE uint32 = 119
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp0SE uint32 = 120
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp1SE uint32 = 121
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp2SE uint32 = 122
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp3SE uint32 = 123
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp0SW uint32 = 124
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp1SW uint32 = 125
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp2SW uint32 = 126
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp3SW uint32 = 127
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp0NW uint32 = 128
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp1NW uint32 = 129
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp2NW uint32 = 130
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp3NW uint32 = 131
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown0NE uint32 = 132
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown1NE uint32 = 133
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown2NE uint32 = 134
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown3NE uint32 = 135
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown0SE uint32 = 136
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown1SE uint32 = 137
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown2SE uint32 = 138
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown3SE uint32 = 139
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown0SW uint32 = 140
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown1SW uint32 = 141
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown2SW uint32 = 142
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown3SW uint32 = 143
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown0NW uint32 = 144
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown1NW uint32 = 145
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown2NW uint32 = 146
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown3NW uint32 = 147
// MALFORMED FIELD: const RightCurveLarge0NE uint32 = 148
// MALFORMED FIELD: const RightCurveLarge1NE uint32 = 149
// MALFORMED FIELD: const RightCurveLarge2NE uint32 = 150
// MALFORMED FIELD: const RightCurveLarge3NE uint32 = 151
// MALFORMED FIELD: const RightCurveLarge4NE uint32 = 152
// MALFORMED FIELD: const RightCurveLarge0SE uint32 = 153
// MALFORMED FIELD: const RightCurveLarge1SE uint32 = 154
// MALFORMED FIELD: const RightCurveLarge2SE uint32 = 155
// MALFORMED FIELD: const RightCurveLarge3SE uint32 = 156
// MALFORMED FIELD: const RightCurveLarge4SE uint32 = 157
// MALFORMED FIELD: const RightCurveLarge0SW uint32 = 158
// MALFORMED FIELD: const RightCurveLarge1SW uint32 = 159
// MALFORMED FIELD: const RightCurveLarge2SW uint32 = 160
// MALFORMED FIELD: const RightCurveLarge3SW uint32 = 161
// MALFORMED FIELD: const RightCurveLarge4SW uint32 = 162
// MALFORMED FIELD: const RightCurveLarge0NW uint32 = 163
// MALFORMED FIELD: const RightCurveLarge1NW uint32 = 164
// MALFORMED FIELD: const RightCurveLarge2NW uint32 = 165
// MALFORMED FIELD: const RightCurveLarge3NW uint32 = 166
// MALFORMED FIELD: const RightCurveLarge4NW uint32 = 167
// MALFORMED FIELD: const LeftCurveLarge0NE uint32 = 168
// MALFORMED FIELD: const LeftCurveLarge1NE uint32 = 169
// MALFORMED FIELD: const LeftCurveLarge2NE uint32 = 170
// MALFORMED FIELD: const LeftCurveLarge3NE uint32 = 171
// MALFORMED FIELD: const LeftCurveLarge4NE uint32 = 172
// MALFORMED FIELD: const LeftCurveLarge0SE uint32 = 173
// MALFORMED FIELD: const LeftCurveLarge1SE uint32 = 174
// MALFORMED FIELD: const LeftCurveLarge2SE uint32 = 175
// MALFORMED FIELD: const LeftCurveLarge3SE uint32 = 176
// MALFORMED FIELD: const LeftCurveLarge4SE uint32 = 177
// MALFORMED FIELD: const LeftCurveLarge0SW uint32 = 178
// MALFORMED FIELD: const LeftCurveLarge1SW uint32 = 179
// MALFORMED FIELD: const LeftCurveLarge2SW uint32 = 180
// MALFORMED FIELD: const LeftCurveLarge3SW uint32 = 181
// MALFORMED FIELD: const LeftCurveLarge4SW uint32 = 182
// MALFORMED FIELD: const LeftCurveLarge0NW uint32 = 183
// MALFORMED FIELD: const LeftCurveLarge1NW uint32 = 184
// MALFORMED FIELD: const LeftCurveLarge2NW uint32 = 185
// MALFORMED FIELD: const LeftCurveLarge3NW uint32 = 186
// MALFORMED FIELD: const LeftCurveLarge4NW uint32 = 187
// MALFORMED FIELD: const Diagonal0NE uint32 = 188
// MALFORMED FIELD: const Diagonal2NE uint32 = 189
// MALFORMED FIELD: const Diagonal1NE uint32 = 190
// MALFORMED FIELD: const Diagonal3NE uint32 = 191
// MALFORMED FIELD: const Diagonal0SE uint32 = 192
// MALFORMED FIELD: const Diagonal2SE uint32 = 193
// MALFORMED FIELD: const Diagonal1SE uint32 = 194
// MALFORMED FIELD: const Diagonal3SE uint32 = 195
// MALFORMED FIELD: const Diagonal0SW uint32 = 196
// MALFORMED FIELD: const Diagonal2SW uint32 = 197
// MALFORMED FIELD: const Diagonal1SW uint32 = 198
// MALFORMED FIELD: const Diagonal3SW uint32 = 199
// MALFORMED FIELD: const Diagonal0NW uint32 = 200
// MALFORMED FIELD: const Diagonal2NW uint32 = 201
// MALFORMED FIELD: const Diagonal1NW uint32 = 202
// MALFORMED FIELD: const Diagonal3NW uint32 = 203
// MALFORMED FIELD: const RightCurveVerySmall0NE uint32 = 204
// MALFORMED FIELD: const RightCurveVerySmall0SE uint32 = 205
// MALFORMED FIELD: const RightCurveVerySmall0SW uint32 = 206
// MALFORMED FIELD: const RightCurveVerySmall0NW uint32 = 207

// namespace Style1
// MALFORMED FIELD: const RightCurveSmall0NE uint32 = 6
// MALFORMED FIELD: const RightCurveSmall3SE uint32 = 9
// MALFORMED FIELD: const RightCurveSmall0SW uint32 = 10
// MALFORMED FIELD: const RightCurveSmall3SW uint32 = 11
// MALFORMED FIELD: const RightCurveSmall0NW uint32 = 12
// MALFORMED FIELD: const RightCurveSmall3NW uint32 = 13
// MALFORMED FIELD: const RightCurve1NE uint32 = 14
// MALFORMED FIELD: const RightCurve1SW uint32 = 15
// MALFORMED FIELD: const RightCurve1SE uint32 = 16
// MALFORMED FIELD: const RightCurve1NW uint32 = 17
// MALFORMED FIELD: const RightCurveVerySmall0NE uint32 = 18
// MALFORMED FIELD: const RightCurveVerySmall0SE uint32 = 19
// MALFORMED FIELD: const RightCurveVerySmall0SW uint32 = 20
// MALFORMED FIELD: const RightCurveVerySmall0NW uint32 = 21
// MALFORMED FIELD: const Diagonal0NW uint32 = 22
// MALFORMED FIELD: const Diagonal0NE uint32 = 23
// MALFORMED FIELD: const Diagonal0SE uint32 = 24
// MALFORMED FIELD: const Diagonal0SW uint32 = 25
// MALFORMED FIELD: const StraightSlopeUp0NE uint32 = 50
// MALFORMED FIELD: const StraightSlopeUp0SE uint32 = 51
// MALFORMED FIELD: const StraightSlopeUp0SW uint32 = 52
// MALFORMED FIELD: const StraightSlopeUp0NW uint32 = 53
// MALFORMED FIELD: const StraightSlopeUp1NE uint32 = 54
// MALFORMED FIELD: const StraightSlopeUp1SE uint32 = 55
// MALFORMED FIELD: const StraightSlopeUp1SW uint32 = 56
// MALFORMED FIELD: const StraightSlopeUp1NW uint32 = 57
// MALFORMED FIELD: const StraightSteepSlopeUp0NE uint32 = 58
// MALFORMED FIELD: const StraightSteepSlopeUp0SE uint32 = 59
// MALFORMED FIELD: const StraightSteepSlopeUp0SW uint32 = 60
// MALFORMED FIELD: const StraightSteepSlopeUp0NW uint32 = 61
// MALFORMED FIELD: const RightCurveSmallSlopeUp0NE uint32 = 62
// MALFORMED FIELD: const RightCurveSmallSlopeUp3NE uint32 = 63
// MALFORMED FIELD: const RightCurveSmallSlopeUp0SE uint32 = 64
// MALFORMED FIELD: const RightCurveSmallSlopeUp3SE uint32 = 65
// MALFORMED FIELD: const RightCurveSmallSlopeUp0SW uint32 = 66
// MALFORMED FIELD: const RightCurveSmallSlopeUp3SW uint32 = 67
// MALFORMED FIELD: const RightCurveSmallSlopeUp0NW uint32 = 68
// MALFORMED FIELD: const RightCurveSmallSlopeUp3NW uint32 = 69
// MALFORMED FIELD: const RightCurveSmallSlopeDown0NE uint32 = 70
// MALFORMED FIELD: const RightCurveSmallSlopeDown3NE uint32 = 71
// MALFORMED FIELD: const RightCurveSmallSlopeDown0SE uint32 = 72
// MALFORMED FIELD: const RightCurveSmallSlopeDown3SE uint32 = 73
// MALFORMED FIELD: const RightCurveSmallSlopeDown0SW uint32 = 74
// MALFORMED FIELD: const RightCurveSmallSlopeDown3SW uint32 = 75
// MALFORMED FIELD: const RightCurveSmallSlopeDown0NW uint32 = 76
// MALFORMED FIELD: const RightCurveSmallSlopeDown3NW uint32 = 77
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp0NE uint32 = 78
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp3NE uint32 = 79
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp0SE uint32 = 80
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp3SE uint32 = 81
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp0SW uint32 = 82
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp3SW uint32 = 83
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp0NW uint32 = 84
// MALFORMED FIELD: const RightCurveSmallSteepSlopeUp3NW uint32 = 85
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown0NE uint32 = 86
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown3NE uint32 = 87
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown0SE uint32 = 88
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown3SE uint32 = 89
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown0SW uint32 = 90
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown3SW uint32 = 91
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown0NW uint32 = 92
// MALFORMED FIELD: const RightCurveSmallSteepSlopeDown3NW uint32 = 93
// MALFORMED FIELD: const SupportStraight0SE uint32 = 94
// MALFORMED FIELD: const SupportConnectorStraight0SE uint32 = 95
// MALFORMED FIELD: const SupportStraight0SW uint32 = 96
// MALFORMED FIELD: const SupportConnectorStraight0SW uint32 = 97
// MALFORMED FIELD: const SupportStraight0NW uint32 = 98
// MALFORMED FIELD: const SupportConnectorStraight0NW uint32 = 99
// MALFORMED FIELD: const SupportStraight0NE uint32 = 100
// MALFORMED FIELD: const SupportConnectorStraight0NE uint32 = 101
// MALFORMED FIELD: const SupportRightCurve1NE uint32 = 102
// MALFORMED FIELD: const SupportConnectorRightCurve1NE uint32 = 103
// MALFORMED FIELD: const SupportRightCurve3NE uint32 = 104
// MALFORMED FIELD: const SupportConnectorRightCurve3NE uint32 = 105
// MALFORMED FIELD: const SupportRightCurve1SE uint32 = 106
// MALFORMED FIELD: const SupportConnectorRightCurve1SE uint32 = 107
// MALFORMED FIELD: const SupportRightCurve3SE uint32 = 108
// MALFORMED FIELD: const SupportConnectorRightCurve3SE uint32 = 109
// MALFORMED FIELD: const SupportRightCurve1SW uint32 = 110
// MALFORMED FIELD: const SupportConnectorRightCurve1SW uint32 = 111
// MALFORMED FIELD: const SupportRightCurve3SW uint32 = 112
// MALFORMED FIELD: const SupportConnectorRightCurve3SW uint32 = 113
// MALFORMED FIELD: const SupportRightCurve1NW uint32 = 114
// MALFORMED FIELD: const SupportConnectorRightCurve1NW uint32 = 115
// MALFORMED FIELD: const SupportRightCurve3NW uint32 = 116
// MALFORMED FIELD: const SupportConnectorRightCurve3NW uint32 = 117
