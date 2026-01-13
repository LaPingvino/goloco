package paint

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Graphics/ImageIds.h"
// #include "Objects/RoadObject.h"
// #include "Paint.h"
// #include "PaintRoadCommonData.h"
// #include <OpenLoco/Core/Numerics.hpp>
// #include <array>
// #include <span>
// namespace OpenLoco::Paint::Style0
// // 0x004083B1, 0x004084BC, 0x004083B1, 0x004084BC
// constexpr RoadPaintMergeablePiece kStraight0 = {
// std::array<uint32_t, 4>{
// ImageIds::road_hit_test_straight_NE,
// ImageIds::road_hit_test_straight_SW,
// ImageIds::road_hit_test_straight_NE,
// ImageIds::road_hit_test_straight_SW,
// },
// /* StreetlightHeights */ kNoStreetlights,
// /* IsMultiTileMerge */ std::array<RoadPaintMergeType, 4>{ RoadPaintMergeType::standard, RoadPaintMergeType::standard, RoadPaintMergeType::standard, RoadPaintMergeType::standard },
var StraightTPP = [1]RoadPaintMergeablePiece{
// kStraight0,
// // 0x004087DD, 0x004088E8, 0x004089F3, 0x00408AFC
// constexpr RoadPaintMergeablePiece kRightCurveVerySmall0 = {
// std::array<uint32_t, 4>{
// ImageIds::road_hit_test_very_small_curve_right_NE,
// ImageIds::road_hit_test_very_small_curve_right_SE,
// ImageIds::road_hit_test_very_small_curve_right_SW,
// ImageIds::road_hit_test_very_small_curve_right_NW,
// },
// /* StreetlightHeights */ kNoStreetlights,
// /* IsMultiTileMerge */ std::array<RoadPaintMergeType, 4>{ RoadPaintMergeType::standard, RoadPaintMergeType::standard, RoadPaintMergeType::standard, RoadPaintMergeType::standard },
var RightCurveVerySmallTPP = [1]RoadPaintMergeablePiece{
// kRightCurveVerySmall0,
const LeftCurveVerySmall0 RoadPaintMergeablePiece = rotateRoadPP(kRightCurveVerySmall0, kRotationTable1230)
var LeftCurveVerySmallTPP = [1]RoadPaintMergeablePiece{
// kLeftCurveVerySmall0,
// // 0x0040935D, 0x004096C9, 0x00409A37, 0x00409DA3
// constexpr RoadPaintMergeablePiece kRightCurveSmall0 = {
// std::array<uint32_t, 4>{
// RoadObj::ImageIds::Style0::kRightCurveSmall0NE,
// RoadObj::ImageIds::Style0::kRightCurveSmall0SE,
// RoadObj::ImageIds::Style0::kRightCurveSmall0SW,
// RoadObj::ImageIds::Style0::kRightCurveSmall0NW,
// },
// /* StreetlightHeights */ std::array<int16_t, 4>{
// kNoStreetlight,
// kNoStreetlight,
// kNoStreetlight,
// 0,
// },
// /* IsMultiTileMerge */ kNoRoadPaintMerge,
// // 0x0040948D, 0x004097F9, 0x00409B67, 0x00409ED3
// constexpr RoadPaintMergeablePiece kRightCurveSmall1 = {
// std::array<uint32_t, 4>{
// RoadObj::ImageIds::Style0::kRightCurveSmall1NE,
// RoadObj::ImageIds::Style0::kRightCurveSmall1SE,
// RoadObj::ImageIds::Style0::kRightCurveSmall1SW,
// RoadObj::ImageIds::Style0::kRightCurveSmall1NW,
// },
// /* StreetlightHeights */ kNoStreetlights,
// /* IsMultiTileMerge */ kNoRoadPaintMerge,
// // 0x00409512, 0x00409880, 0x00409BEE, 0x00409F5A
// constexpr RoadPaintMergeablePiece kRightCurveSmall2 = {
// std::array<uint32_t, 4>{
// RoadObj::ImageIds::Style0::kRightCurveSmall2NE,
// RoadObj::ImageIds::Style0::kRightCurveSmall2SE,
// RoadObj::ImageIds::Style0::kRightCurveSmall2SW,
// RoadObj::ImageIds::Style0::kRightCurveSmall2NW,
// },
// /* StreetlightHeights */ kNoStreetlights,
// /* IsMultiTileMerge */ kNoRoadPaintMerge,
// // 0x00409599, 0x00409907, 0x00409C73, 0x00409FE1
// constexpr RoadPaintMergeablePiece kRightCurveSmall3 = {
// std::array<uint32_t, 4>{
// RoadObj::ImageIds::Style0::kRightCurveSmall3NE,
// RoadObj::ImageIds::Style0::kRightCurveSmall3SE,
// RoadObj::ImageIds::Style0::kRightCurveSmall3SW,
// RoadObj::ImageIds::Style0::kRightCurveSmall3NW,
// },
// /* StreetlightHeights */ std::array<int16_t, 4>{
// 0,
// kNoStreetlight,
// kNoStreetlight,
// kNoStreetlight,
// },
// /* IsMultiTileMerge */ kNoRoadPaintMerge,
var RightCurveSmallTPP = [4]RoadPaintMergeablePiece{
// kRightCurveSmall0,
// kRightCurveSmall1,
// kRightCurveSmall2,
// kRightCurveSmall3,
const LeftCurveSmall0 RoadPaintMergeablePiece = rotateRoadPP(kRightCurveSmall3, kRotationTable1230)
const LeftCurveSmall1 RoadPaintMergeablePiece = rotateRoadPP(kRightCurveSmall1, kRotationTable1230)
const LeftCurveSmall2 RoadPaintMergeablePiece = rotateRoadPP(kRightCurveSmall2, kRotationTable1230)
const LeftCurveSmall3 RoadPaintMergeablePiece = rotateRoadPP(kRightCurveSmall0, kRotationTable1230)
var LeftCurveSmallTPP = [4]RoadPaintMergeablePiece{
// kLeftCurveSmall0,
// kLeftCurveSmall1,
// kLeftCurveSmall2,
// kLeftCurveSmall3,
// // 0x0040AF19, 0x0040B27F, 0x0040B5E5, 0x0040B94B
// constexpr RoadPaintMergeablePiece kStraightSlopeUp0 = {
// std::array<uint32_t, 4>{
// RoadObj::ImageIds::Style0::kStraightSlopeUp0NE,
// RoadObj::ImageIds::Style0::kStraightSlopeUp0SE,
// RoadObj::ImageIds::Style0::kStraightSlopeUp0SW,
// RoadObj::ImageIds::Style0::kStraightSlopeUp0NW,
// },
// /* StreetlightHeights */ std::array<int16_t, 4>{
// kNoStreetlight,
// 4,
// kNoStreetlight,
// 4,
// },
// /* IsMultiTileMerge */ kNoRoadPaintMerge,
// // 0x0040B0CA, 0x0040B430, 0x0040B796, 0x0040BAFC
// constexpr RoadPaintMergeablePiece kStraightSlopeUp1 = {
// std::array<uint32_t, 4>{
// RoadObj::ImageIds::Style0::kStraightSlopeUp1NE,
// RoadObj::ImageIds::Style0::kStraightSlopeUp1SE,
// RoadObj::ImageIds::Style0::kStraightSlopeUp1SW,
// RoadObj::ImageIds::Style0::kStraightSlopeUp1NW,
// },
// /* StreetlightHeights */ std::array<int16_t, 4>{
// kNoStreetlight,
// 12,
// kNoStreetlight,
// 12,
// },
// /* IsMultiTileMerge */ kNoRoadPaintMerge,
var StraightSlopeUpTPP = [2]RoadPaintMergeablePiece{
// kStraightSlopeUp0,
// kStraightSlopeUp1,
const StraightSlopeDown0 RoadPaintMergeablePiece = rotateRoadPP(kStraightSlopeUp1, kRotationTable2301)
const StraightSlopeDown1 RoadPaintMergeablePiece = rotateRoadPP(kStraightSlopeUp0, kRotationTable2301)
var StraightSlopeDownTPP = [2]RoadPaintMergeablePiece{
// kStraightSlopeDown0,
// kStraightSlopeDown1,
// // 0x0040CA49, 0x0040CC2E, 0x0040CE13, 0x0040CFF8
// constexpr RoadPaintMergeablePiece kStraightSteepSlopeUp0 = {
// std::array<uint32_t, 4>{
// RoadObj::ImageIds::Style0::kStraightSteepSlopeUp0NE,
// RoadObj::ImageIds::Style0::kStraightSteepSlopeUp0SE,
// RoadObj::ImageIds::Style0::kStraightSteepSlopeUp0SW,
// RoadObj::ImageIds::Style0::kStraightSteepSlopeUp0NW,
// },
// /* StreetlightHeights */ std::array<int16_t, 4>{
// kNoStreetlight,
// 8,
// kNoStreetlight,
// 8,
// },
// /* IsMultiTileMerge */ kNoRoadPaintMerge,
var StraightSteepSlopeUpTPP = [1]RoadPaintMergeablePiece{
// kStraightSteepSlopeUp0,
const StraightSteepSlopeDown0 RoadPaintMergeablePiece = rotateRoadPP(kStraightSteepSlopeUp0, kRotationTable2301)
var StraightSteepSlopeDownTPP = [1]RoadPaintMergeablePiece{
// kStraightSteepSlopeDown0,
// // 0x00409031, 0x004090E8, 0x0040919D, 0x00409252
// constexpr RoadPaintMergeablePiece kTurnaround0 = {
// std::array<uint32_t, 4>{
// RoadObj::ImageIds::Style0::kTurnaround0NE,
// RoadObj::ImageIds::Style0::kTurnaround0SE,
// RoadObj::ImageIds::Style0::kTurnaround0SW,
// RoadObj::ImageIds::Style0::kTurnaround0NW,
// },
// /* StreetlightHeights */ kNoStreetlights,
// /* IsMultiTileMerge */ kNoRoadPaintMerge,
var TurnaroundTPP = [1]RoadPaintMergeablePiece{
// kTurnaround0,
// constexpr std::array<std::span<const RoadPaintMergeablePiece>, 10> kRoadPaintParts = {
// kStraightTPP,
// kLeftCurveVerySmallTPP,
// kRightCurveVerySmallTPP,
// kLeftCurveSmallTPP,
// kRightCurveSmallTPP,
// kStraightSlopeUpTPP,
// kStraightSlopeDownTPP,
// kStraightSteepSlopeUpTPP,
// kStraightSteepSlopeDownTPP,
// kTurnaroundTPP,
