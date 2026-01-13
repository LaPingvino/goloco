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
// namespace Gfx
// forward: class DrawingContext;
type TrackObjectFlags int

const (
	None TrackObjectFlags = 0
	HasRackRail TrackObjectFlags = 1 << 0
	NoSlipSurface TrackObjectFlags = 1 << 1
	IsRoad TrackObjectFlags = 1 << 2
	Unk_04 TrackObjectFlags = 1 << 4
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TrackObjectFlags);
type TrackObject struct {
	Name StringId
// World::Track::TrackTraitFlags trackPieces;        // 0x02
// World::Track::TrackTraitFlags stationTrackPieces; // 0x04
	Var_06 uint8
	NumCompatible uint8
	NumMods uint8
	NumSignals uint8
// uint8_t mods[4];           // 0x0A
	Signals uint16
	CompatibleTracks uint16
	CompatibleRoads uint16
	BuildCostFactor int16
	SellCostFactor int16
	TunnelCostFactor int16
	CostIndex uint8
	Tunnel uint8
	CurveSpeed Speed16
	Image uint32
	Flags TrackObjectFlags
	NumBridges uint8
// uint8_t bridges[7];        // 0x25
	NumStations uint8
// uint8_t stations[7];       // 0x2D
	DisplayOffset uint8
	Pad_35 uint8
	// method: void drawPreviewImage(Gfx::DrawingContext& drawingCtx, const int16_t x, const int16_t y) const;
	// method: bool validate() const;
	// method: void load(const LoadedObjectHandle& handle, std::span<const std::byte> data, ObjectManager::DependentObjects* dependencies);
	// method: void unload();
	// method: constexpr bool hasFlags(TrackObjectFlags flagsToTest) const
// return (flags & flagsToTest) != TrackObjectFlags::none;
}
const TrackObjectObjectType any = ObjectType.track
// func HasTraitFlags(flagsToTest World::Track::TrackTraitFlags) bool
// return (trackPieces & flagsToTest) != World::Track::TrackTraitFlags::none;
// static_assert(sizeof(TrackObject) == 0x36);
// namespace TrackObj::ImageIds
const UiPreviewImage0 uint32 = 0
const UiPreviewImage1 uint32 = 1
const UiPreviewImage2 uint32 = 2
const UiPreviewImage3 uint32 = 3
const UiPreviewImage4 uint32 = 4
const UiPreviewImage5 uint32 = 5
const UiPreviewImage6 uint32 = 6
const UiPreviewImage7 uint32 = 7
const UiPreviewImage8 uint32 = 8
const UiPreviewImage9 uint32 = 9
const UiPreviewImage10 uint32 = 10
const UiPreviewImage11 uint32 = 11
const UiPreviewImage12 uint32 = 12
const UiPreviewImage13 uint32 = 13
const UiPreviewImage14 uint32 = 14
const UiPreviewImage15 uint32 = 15
const UiPickupFromTrack uint32 = 16
const UiPlaceOnTrack uint32 = 17
// // Assumes rotational symmetry
// // k{TrackId}{sequenceIndex}[type]{direction}
// // type = Ballast, Sleeper, Rail
// // type only used for mergeable track ids
// namespace Style0
const Straight0BallastNE uint32 = 18
const Straight0BallastSE uint32 = 19
const Straight0SleeperNE uint32 = 20
const Straight0SleeperSE uint32 = 21
const Straight0RailNE uint32 = 22
const Straight0RailSE uint32 = 23
const RightCurveSmall0BallastNE uint32 = 24
const RightCurveSmall1BallastNE uint32 = 25
const RightCurveSmall2BallastNE uint32 = 26
const RightCurveSmall3BallastNE uint32 = 27
const RightCurveSmall0BallastSE uint32 = 28
const RightCurveSmall1BallastSE uint32 = 29
const RightCurveSmall2BallastSE uint32 = 30
const RightCurveSmall3BallastSE uint32 = 31
const RightCurveSmall0BallastSW uint32 = 32
const RightCurveSmall1BallastSW uint32 = 33
const RightCurveSmall2BallastSW uint32 = 34
const RightCurveSmall3BallastSW uint32 = 35
const RightCurveSmall0BallastNW uint32 = 36
const RightCurveSmall1BallastNW uint32 = 37
const RightCurveSmall2BallastNW uint32 = 38
const RightCurveSmall3BallastNW uint32 = 39
const RightCurveSmall0SleeperNE uint32 = 40
const RightCurveSmall1SleeperNE uint32 = 41
const RightCurveSmall2SleeperNE uint32 = 42
const RightCurveSmall3SleeperNE uint32 = 43
const RightCurveSmall0SleeperSE uint32 = 44
const RightCurveSmall1SleeperSE uint32 = 45
const RightCurveSmall2SleeperSE uint32 = 46
const RightCurveSmall3SleeperSE uint32 = 47
const RightCurveSmall0SleeperSW uint32 = 48
const RightCurveSmall1SleeperSW uint32 = 49
const RightCurveSmall2SleeperSW uint32 = 50
const RightCurveSmall3SleeperSW uint32 = 51
const RightCurveSmall0SleeperNW uint32 = 52
const RightCurveSmall1SleeperNW uint32 = 53
const RightCurveSmall2SleeperNW uint32 = 54
const RightCurveSmall3SleeperNW uint32 = 55
const RightCurveSmall0RailNE uint32 = 56
const RightCurveSmall1RailNE uint32 = 57
const RightCurveSmall2RailNE uint32 = 58
const RightCurveSmall3RailNE uint32 = 59
const RightCurveSmall0RailSE uint32 = 60
const RightCurveSmall1RailSE uint32 = 61
const RightCurveSmall2RailSE uint32 = 62
const RightCurveSmall3RailSE uint32 = 63
const RightCurveSmall0RailSW uint32 = 64
const RightCurveSmall1RailSW uint32 = 65
const RightCurveSmall2RailSW uint32 = 66
const RightCurveSmall3RailSW uint32 = 67
const RightCurveSmall0RailNW uint32 = 68
const RightCurveSmall1RailNW uint32 = 69
const RightCurveSmall2RailNW uint32 = 70
const RightCurveSmall3RailNW uint32 = 71
const RightCurveSmallSlopeUp0NE uint32 = 72
const RightCurveSmallSlopeUp1NE uint32 = 73
const RightCurveSmallSlopeUp2NE uint32 = 74
const RightCurveSmallSlopeUp3NE uint32 = 75
const RightCurveSmallSlopeUp0SE uint32 = 76
const RightCurveSmallSlopeUp1SE uint32 = 77
const RightCurveSmallSlopeUp2SE uint32 = 78
const RightCurveSmallSlopeUp3SE uint32 = 79
const RightCurveSmallSlopeUp0SW uint32 = 80
const RightCurveSmallSlopeUp1SW uint32 = 81
const RightCurveSmallSlopeUp2SW uint32 = 82
const RightCurveSmallSlopeUp3SW uint32 = 83
const RightCurveSmallSlopeUp0NW uint32 = 84
const RightCurveSmallSlopeUp1NW uint32 = 85
const RightCurveSmallSlopeUp2NW uint32 = 86
const RightCurveSmallSlopeUp3NW uint32 = 87
const RightCurveSmallSlopeDown0NE uint32 = 88
const RightCurveSmallSlopeDown1NE uint32 = 89
const RightCurveSmallSlopeDown2NE uint32 = 90
const RightCurveSmallSlopeDown3NE uint32 = 91
const RightCurveSmallSlopeDown0SE uint32 = 92
const RightCurveSmallSlopeDown1SE uint32 = 93
const RightCurveSmallSlopeDown2SE uint32 = 94
const RightCurveSmallSlopeDown3SE uint32 = 95
const RightCurveSmallSlopeDown0SW uint32 = 96
const RightCurveSmallSlopeDown1SW uint32 = 97
const RightCurveSmallSlopeDown2SW uint32 = 98
const RightCurveSmallSlopeDown3SW uint32 = 99
const RightCurveSmallSlopeDown0NW uint32 = 100
const RightCurveSmallSlopeDown1NW uint32 = 101
const RightCurveSmallSlopeDown2NW uint32 = 102
const RightCurveSmallSlopeDown3NW uint32 = 103
const RightCurveSmallSteepSlopeUp0NE uint32 = 104
const RightCurveSmallSteepSlopeUp1NE uint32 = 105
const RightCurveSmallSteepSlopeUp2NE uint32 = 106
const RightCurveSmallSteepSlopeUp3NE uint32 = 107
const RightCurveSmallSteepSlopeUp0SE uint32 = 108
const RightCurveSmallSteepSlopeUp1SE uint32 = 109
const RightCurveSmallSteepSlopeUp2SE uint32 = 110
const RightCurveSmallSteepSlopeUp3SE uint32 = 111
const RightCurveSmallSteepSlopeUp0SW uint32 = 112
const RightCurveSmallSteepSlopeUp1SW uint32 = 113
const RightCurveSmallSteepSlopeUp2SW uint32 = 114
const RightCurveSmallSteepSlopeUp3SW uint32 = 115
const RightCurveSmallSteepSlopeUp0NW uint32 = 116
const RightCurveSmallSteepSlopeUp1NW uint32 = 117
const RightCurveSmallSteepSlopeUp2NW uint32 = 118
const RightCurveSmallSteepSlopeUp3NW uint32 = 119
const RightCurveSmallSteepSlopeDown0NE uint32 = 120
const RightCurveSmallSteepSlopeDown1NE uint32 = 121
const RightCurveSmallSteepSlopeDown2NE uint32 = 122
const RightCurveSmallSteepSlopeDown3NE uint32 = 123
const RightCurveSmallSteepSlopeDown0SE uint32 = 124
const RightCurveSmallSteepSlopeDown1SE uint32 = 125
const RightCurveSmallSteepSlopeDown2SE uint32 = 126
const RightCurveSmallSteepSlopeDown3SE uint32 = 127
const RightCurveSmallSteepSlopeDown0SW uint32 = 128
const RightCurveSmallSteepSlopeDown1SW uint32 = 129
const RightCurveSmallSteepSlopeDown2SW uint32 = 130
const RightCurveSmallSteepSlopeDown3SW uint32 = 131
const RightCurveSmallSteepSlopeDown0NW uint32 = 132
const RightCurveSmallSteepSlopeDown1NW uint32 = 133
const RightCurveSmallSteepSlopeDown2NW uint32 = 134
const RightCurveSmallSteepSlopeDown3NW uint32 = 135
const RightCurve0BallastNE uint32 = 136
const RightCurve1BallastNE uint32 = 137
const RightCurve2BallastNE uint32 = 138
const RightCurve3BallastNE uint32 = 139
const RightCurve4BallastNE uint32 = 140
const RightCurve0BallastSE uint32 = 141
const RightCurve1BallastSE uint32 = 142
const RightCurve2BallastSE uint32 = 143
const RightCurve3BallastSE uint32 = 144
const RightCurve4BallastSE uint32 = 145
const RightCurve0BallastSW uint32 = 146
const RightCurve1BallastSW uint32 = 147
const RightCurve2BallastSW uint32 = 148
const RightCurve3BallastSW uint32 = 149
const RightCurve4BallastSW uint32 = 150
const RightCurve0BallastNW uint32 = 151
const RightCurve1BallastNW uint32 = 152
const RightCurve2BallastNW uint32 = 153
const RightCurve3BallastNW uint32 = 154
const RightCurve4BallastNW uint32 = 155
const RightCurve0SleeperNE uint32 = 156
const RightCurve1SleeperNE uint32 = 157
const RightCurve2SleeperNE uint32 = 158
const RightCurve3SleeperNE uint32 = 159
const RightCurve4SleeperNE uint32 = 160
const RightCurve0SleeperSE uint32 = 161
const RightCurve1SleeperSE uint32 = 162
const RightCurve2SleeperSE uint32 = 163
const RightCurve3SleeperSE uint32 = 164
const RightCurve4SleeperSE uint32 = 165
const RightCurve0SleeperSW uint32 = 166
const RightCurve1SleeperSW uint32 = 167
const RightCurve2SleeperSW uint32 = 168
const RightCurve3SleeperSW uint32 = 169
const RightCurve4SleeperSW uint32 = 170
const RightCurve0SleeperNW uint32 = 171
const RightCurve1SleeperNW uint32 = 172
const RightCurve2SleeperNW uint32 = 173
const RightCurve3SleeperNW uint32 = 174
const RightCurve4SleeperNW uint32 = 175
const RightCurve0RailNE uint32 = 176
const RightCurve1RailNE uint32 = 177
const RightCurve2RailNE uint32 = 178
const RightCurve3RailNE uint32 = 179
const RightCurve4RailNE uint32 = 180
const RightCurve0RailSE uint32 = 181
const RightCurve1RailSE uint32 = 182
const RightCurve2RailSE uint32 = 183
const RightCurve3RailSE uint32 = 184
const RightCurve4RailSE uint32 = 185
const RightCurve0RailSW uint32 = 186
const RightCurve1RailSW uint32 = 187
const RightCurve2RailSW uint32 = 188
const RightCurve3RailSW uint32 = 189
const RightCurve4RailSW uint32 = 190
const RightCurve0RailNW uint32 = 191
const RightCurve1RailNW uint32 = 192
const RightCurve2RailNW uint32 = 193
const RightCurve3RailNW uint32 = 194
const RightCurve4RailNW uint32 = 195
const StraightSlopeUp0NE uint32 = 196
const StraightSlopeUp1NE uint32 = 197
const StraightSlopeUp0SE uint32 = 198
const StraightSlopeUp1SE uint32 = 199
const StraightSlopeUp0SW uint32 = 200
const StraightSlopeUp1SW uint32 = 201
const StraightSlopeUp0NW uint32 = 202
const StraightSlopeUp1NW uint32 = 203
const StraightSteepSlopeUp0NE uint32 = 204
const StraightSteepSlopeUp0SE uint32 = 205
const StraightSteepSlopeUp0SW uint32 = 206
const StraightSteepSlopeUp0NW uint32 = 207
const RightCurveLarge0BallastNE uint32 = 208
const RightCurveLarge1BallastNE uint32 = 209
const RightCurveLarge2BallastNE uint32 = 210
const RightCurveLarge3BallastNE uint32 = 211
const RightCurveLarge4BallastNE uint32 = 212
const RightCurveLarge0BallastSE uint32 = 213
const RightCurveLarge1BallastSE uint32 = 214
const RightCurveLarge2BallastSE uint32 = 215
const RightCurveLarge3BallastSE uint32 = 216
const RightCurveLarge4BallastSE uint32 = 217
const RightCurveLarge0BallastSW uint32 = 218
const RightCurveLarge1BallastSW uint32 = 219
const RightCurveLarge2BallastSW uint32 = 220
const RightCurveLarge3BallastSW uint32 = 221
const RightCurveLarge4BallastSW uint32 = 222
const RightCurveLarge0BallastNW uint32 = 223
const RightCurveLarge1BallastNW uint32 = 224
const RightCurveLarge2BallastNW uint32 = 225
const RightCurveLarge3BallastNW uint32 = 226
const RightCurveLarge4BallastNW uint32 = 227
const LeftCurveLarge0BallastNE uint32 = 228
const LeftCurveLarge1BallastNE uint32 = 229
const LeftCurveLarge2BallastNE uint32 = 230
const LeftCurveLarge3BallastNE uint32 = 231
const LeftCurveLarge4BallastNE uint32 = 232
const LeftCurveLarge0BallastSE uint32 = 233
const LeftCurveLarge1BallastSE uint32 = 234
const LeftCurveLarge2BallastSE uint32 = 235
const LeftCurveLarge3BallastSE uint32 = 236
const LeftCurveLarge4BallastSE uint32 = 237
const LeftCurveLarge0BallastSW uint32 = 238
const LeftCurveLarge1BallastSW uint32 = 239
const LeftCurveLarge2BallastSW uint32 = 240
const LeftCurveLarge3BallastSW uint32 = 241
const LeftCurveLarge4BallastSW uint32 = 242
const LeftCurveLarge0BallastNW uint32 = 243
const LeftCurveLarge1BallastNW uint32 = 244
const LeftCurveLarge2BallastNW uint32 = 245
const LeftCurveLarge3BallastNW uint32 = 246
const LeftCurveLarge4BallastNW uint32 = 247
const RightCurveLarge0SleeperNE uint32 = 248
const RightCurveLarge1SleeperNE uint32 = 249
const RightCurveLarge2SleeperNE uint32 = 250
const RightCurveLarge3SleeperNE uint32 = 251
const RightCurveLarge4SleeperNE uint32 = 252
const RightCurveLarge0SleeperSE uint32 = 253
const RightCurveLarge1SleeperSE uint32 = 254
const RightCurveLarge2SleeperSE uint32 = 255
const RightCurveLarge3SleeperSE uint32 = 256
const RightCurveLarge4SleeperSE uint32 = 257
const RightCurveLarge0SleeperSW uint32 = 258
const RightCurveLarge1SleeperSW uint32 = 259
const RightCurveLarge2SleeperSW uint32 = 260
const RightCurveLarge3SleeperSW uint32 = 261
const RightCurveLarge4SleeperSW uint32 = 262
const RightCurveLarge0SleeperNW uint32 = 263
const RightCurveLarge1SleeperNW uint32 = 264
const RightCurveLarge2SleeperNW uint32 = 265
const RightCurveLarge3SleeperNW uint32 = 266
const RightCurveLarge4SleeperNW uint32 = 267
const LeftCurveLarge0SleeperNE uint32 = 268
const LeftCurveLarge1SleeperNE uint32 = 269
const LeftCurveLarge2SleeperNE uint32 = 270
const LeftCurveLarge3SleeperNE uint32 = 271
const LeftCurveLarge4SleeperNE uint32 = 272
const LeftCurveLarge0SleeperSE uint32 = 273
const LeftCurveLarge1SleeperSE uint32 = 274
const LeftCurveLarge2SleeperSE uint32 = 275
const LeftCurveLarge3SleeperSE uint32 = 276
const LeftCurveLarge4SleeperSE uint32 = 277
const LeftCurveLarge0SleeperSW uint32 = 278
const LeftCurveLarge1SleeperSW uint32 = 279
const LeftCurveLarge2SleeperSW uint32 = 280
const LeftCurveLarge3SleeperSW uint32 = 281
const LeftCurveLarge4SleeperSW uint32 = 282
const LeftCurveLarge0SleeperNW uint32 = 283
const LeftCurveLarge1SleeperNW uint32 = 284
const LeftCurveLarge2SleeperNW uint32 = 285
const LeftCurveLarge3SleeperNW uint32 = 286
const LeftCurveLarge4SleeperNW uint32 = 287
const RightCurveLarge0RailNE uint32 = 288
const RightCurveLarge1RailNE uint32 = 289
const RightCurveLarge2RailNE uint32 = 290
const RightCurveLarge3RailNE uint32 = 291
const RightCurveLarge4RailNE uint32 = 292
const RightCurveLarge0RailSE uint32 = 293
const RightCurveLarge1RailSE uint32 = 294
const RightCurveLarge2RailSE uint32 = 295
const RightCurveLarge3RailSE uint32 = 296
const RightCurveLarge4RailSE uint32 = 297
const RightCurveLarge0RailSW uint32 = 298
const RightCurveLarge1RailSW uint32 = 299
const RightCurveLarge2RailSW uint32 = 300
const RightCurveLarge3RailSW uint32 = 301
const RightCurveLarge4RailSW uint32 = 302
const RightCurveLarge0RailNW uint32 = 303
const RightCurveLarge1RailNW uint32 = 304
const RightCurveLarge2RailNW uint32 = 305
const RightCurveLarge3RailNW uint32 = 306
const RightCurveLarge4RailNW uint32 = 307
const LeftCurveLarge0RailNE uint32 = 308
const LeftCurveLarge1RailNE uint32 = 309
const LeftCurveLarge2RailNE uint32 = 310
const LeftCurveLarge3RailNE uint32 = 311
const LeftCurveLarge4RailNE uint32 = 312
const LeftCurveLarge0RailSE uint32 = 313
const LeftCurveLarge1RailSE uint32 = 314
const LeftCurveLarge2RailSE uint32 = 315
const LeftCurveLarge3RailSE uint32 = 316
const LeftCurveLarge4RailSE uint32 = 317
const LeftCurveLarge0RailSW uint32 = 318
const LeftCurveLarge1RailSW uint32 = 319
const LeftCurveLarge2RailSW uint32 = 320
const LeftCurveLarge3RailSW uint32 = 321
const LeftCurveLarge4RailSW uint32 = 322
const LeftCurveLarge0RailNW uint32 = 323
const LeftCurveLarge1RailNW uint32 = 324
const LeftCurveLarge2RailNW uint32 = 325
const LeftCurveLarge3RailNW uint32 = 326
const LeftCurveLarge4RailNW uint32 = 327
const Diagonal0BallastE uint32 = 328
const Diagonal2BallastE uint32 = 329
const Diagonal1BallastE uint32 = 330
const Diagonal3BallastE uint32 = 331
const Diagonal0BallastS uint32 = 332
const Diagonal2BallastS uint32 = 333
const Diagonal1BallastS uint32 = 334
const Diagonal3BallastS uint32 = 335
const Diagonal0SleeperE uint32 = 336
const Diagonal2SleeperE uint32 = 337
const Diagonal1SleeperE uint32 = 338
const Diagonal3SleeperE uint32 = 339
const Diagonal0SleeperS uint32 = 340
const Diagonal2SleeperS uint32 = 341
const Diagonal1SleeperS uint32 = 342
const Diagonal3SleeperS uint32 = 343
const Diagonal0RailE uint32 = 344
const Diagonal2RailE uint32 = 345
const Diagonal1RailE uint32 = 346
const Diagonal3RailE uint32 = 347
const Diagonal0RailS uint32 = 348
const Diagonal2RailS uint32 = 349
const Diagonal1RailS uint32 = 350
const Diagonal3RailS uint32 = 351
const SBendLeft0BallastNE uint32 = 352
const SBendLeft1BallastNE uint32 = 353
const SBendLeft2BallastNE uint32 = 354
const SBendLeft3BallastNE uint32 = 355
const SBendLeft0BallastSE uint32 = 356
const SBendLeft1BallastSE uint32 = 357
const SBendLeft2BallastSE uint32 = 358
const SBendLeft3BallastSE uint32 = 359
const SBendRight0BallastNE uint32 = 360
const SBendRight1BallastNE uint32 = 361
const SBendRight2BallastNE uint32 = 362
const SBendRight3BallastNE uint32 = 363
const SBendRight0BallastSE uint32 = 364
const SBendRight1BallastSE uint32 = 365
const SBendRight2BallastSE uint32 = 366
const SBendRight3BallastSE uint32 = 367
const SBendLeft0SleeperNE uint32 = 368
const SBendLeft1SleeperNE uint32 = 369
const SBendLeft2SleeperNE uint32 = 370
const SBendLeft3SleeperNE uint32 = 371
const SBendLeft0SleeperSE uint32 = 372
const SBendLeft1SleeperSE uint32 = 373
const SBendLeft2SleeperSE uint32 = 374
const SBendLeft3SleeperSE uint32 = 375
const SBendRight0SleeperNE uint32 = 376
const SBendRight1SleeperNE uint32 = 377
const SBendRight2SleeperNE uint32 = 378
const SBendRight3SleeperNE uint32 = 379
const SBendRight0SleeperSE uint32 = 380
const SBendRight1SleeperSE uint32 = 381
const SBendRight2SleeperSE uint32 = 382
const SBendRight3SleeperSE uint32 = 383
const SBendLeft0RailNE uint32 = 384
const SBendLeft1RailNE uint32 = 385
const SBendLeft2RailNE uint32 = 386
const SBendLeft3RailNE uint32 = 387
const SBendLeft0RailSE uint32 = 388
const SBendLeft1RailSE uint32 = 389
const SBendLeft2RailSE uint32 = 390
const SBendLeft3RailSE uint32 = 391
const SBendRight0RailNE uint32 = 392
const SBendRight1RailNE uint32 = 393
const SBendRight2RailNE uint32 = 394
const SBendRight3RailNE uint32 = 395
const SBendRight0RailSE uint32 = 396
const SBendRight1RailSE uint32 = 397
const SBendRight2RailSE uint32 = 398
const SBendRight3RailSE uint32 = 399
const RightCurveVerySmall0BallastNE uint32 = 400
const RightCurveVerySmall0BallastSE uint32 = 401
const RightCurveVerySmall0BallastSW uint32 = 402
const RightCurveVerySmall0BallastNW uint32 = 403
const RightCurveVerySmall0SleeperNE uint32 = 404
const RightCurveVerySmall0SleeperSE uint32 = 405
const RightCurveVerySmall0SleeperSW uint32 = 406
const RightCurveVerySmall0SleeperNW uint32 = 407
const RightCurveVerySmall0RailNE uint32 = 408
const RightCurveVerySmall0RailSE uint32 = 409
const RightCurveVerySmall0RailSW uint32 = 410
const RightCurveVerySmall0RailNW uint32 = 411
