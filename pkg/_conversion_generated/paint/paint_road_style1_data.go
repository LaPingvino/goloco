package paint

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Objects/RoadObject.h"
// #include "Paint.h"
// #include <array>
// #include <span>
// namespace OpenLoco::Paint::Style1
type RoadPaintPiece struct {
// std::array<std::array<uint32_t, 3>, 4> imageIndexOffsets;
}
var RotationTable1230 = [4]uint8{
var RotationTable2301 = [4]uint8{
// consteval RoadPaintPiece rotateRoadPP(const RoadPaintPiece& reference, const std::array<uint8_t, 4>& rotationTable)
// // MSVC 14.44.35207 will ICE if we don't create a temporary array here
// const auto intermediateArray = std::array<std::array<uint32_t, 3>, 4>{
// reference.imageIndexOffsets[rotationTable[0]],
// reference.imageIndexOffsets[rotationTable[1]],
// reference.imageIndexOffsets[rotationTable[2]],
// reference.imageIndexOffsets[rotationTable[3]],
// orphan member: return RoadPaintPiece{ intermediateArray };
// // 0x0040D9AD, 0x0040DADE, 0x0040D9AD, 0x0040DADE
// constexpr RoadPaintPiece kStraight0 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraight0BallastNE, RoadObj::ImageIds::Style1::kStraight0SleeperNE, RoadObj::ImageIds::Style1::kStraight0RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraight0BallastSE, RoadObj::ImageIds::Style1::kStraight0SleeperSE, RoadObj::ImageIds::Style1::kStraight0RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraight0BallastNE, RoadObj::ImageIds::Style1::kStraight0SleeperNE, RoadObj::ImageIds::Style1::kStraight0RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraight0BallastSE, RoadObj::ImageIds::Style1::kStraight0SleeperSE, RoadObj::ImageIds::Style1::kStraight0RailSE },
// },
var StraightTPP = [1]RoadPaintPiece{
// kStraight0,
// // 0x0040DC0F, 0x0040DD46, 0x0040DE7D, 0x0040DFB2
// constexpr RoadPaintPiece kRightCurveVerySmall0 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveVerySmall0BallastNE, RoadObj::ImageIds::Style1::kRightCurveVerySmall0SleeperNE, RoadObj::ImageIds::Style1::kRightCurveVerySmall0RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveVerySmall0BallastSE, RoadObj::ImageIds::Style1::kRightCurveVerySmall0SleeperSE, RoadObj::ImageIds::Style1::kRightCurveVerySmall0RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveVerySmall0BallastSW, RoadObj::ImageIds::Style1::kRightCurveVerySmall0SleeperSW, RoadObj::ImageIds::Style1::kRightCurveVerySmall0RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveVerySmall0BallastNW, RoadObj::ImageIds::Style1::kRightCurveVerySmall0SleeperNW, RoadObj::ImageIds::Style1::kRightCurveVerySmall0RailNW },
// },
var RightCurveVerySmallTPP = [1]RoadPaintPiece{
// kRightCurveVerySmall0,
const LeftCurveVerySmall0 RoadPaintPiece = rotateRoadPP(kRightCurveVerySmall0, kRotationTable1230)
var LeftCurveVerySmallTPP = [1]RoadPaintPiece{
// kLeftCurveVerySmall0,
// // 0x0040E545, 0x0040E8E7, 0x0040EC8B, 0x0040F02D
// constexpr RoadPaintPiece kRightCurveSmall0 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall0BallastNE, RoadObj::ImageIds::Style1::kRightCurveSmall0SleeperNE, RoadObj::ImageIds::Style1::kRightCurveSmall0RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall0BallastSE, RoadObj::ImageIds::Style1::kRightCurveSmall0SleeperSE, RoadObj::ImageIds::Style1::kRightCurveSmall0RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall0BallastSW, RoadObj::ImageIds::Style1::kRightCurveSmall0SleeperSW, RoadObj::ImageIds::Style1::kRightCurveSmall0RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall0BallastNW, RoadObj::ImageIds::Style1::kRightCurveSmall0SleeperNW, RoadObj::ImageIds::Style1::kRightCurveSmall0RailNW },
// },
// // 0x0040E646, 0x0040E9E8, 0x0040ED8C, 0x0040F12E
// constexpr RoadPaintPiece kRightCurveSmall1 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall1BallastNE, RoadObj::ImageIds::Style1::kRightCurveSmall1SleeperNE, RoadObj::ImageIds::Style1::kRightCurveSmall1RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall1BallastSE, RoadObj::ImageIds::Style1::kRightCurveSmall1SleeperSE, RoadObj::ImageIds::Style1::kRightCurveSmall1RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall1BallastSW, RoadObj::ImageIds::Style1::kRightCurveSmall1SleeperSW, RoadObj::ImageIds::Style1::kRightCurveSmall1RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall1BallastNW, RoadObj::ImageIds::Style1::kRightCurveSmall1SleeperNW, RoadObj::ImageIds::Style1::kRightCurveSmall1RailNW },
// },
// // 0x0040E715, 0x0040EAB9, 0x0040EE5D, 0x0040F1FF
// constexpr RoadPaintPiece kRightCurveSmall2 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall2BallastNE, RoadObj::ImageIds::Style1::kRightCurveSmall2SleeperNE, RoadObj::ImageIds::Style1::kRightCurveSmall2RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall2BallastSE, RoadObj::ImageIds::Style1::kRightCurveSmall2SleeperSE, RoadObj::ImageIds::Style1::kRightCurveSmall2RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall2BallastSW, RoadObj::ImageIds::Style1::kRightCurveSmall2SleeperSW, RoadObj::ImageIds::Style1::kRightCurveSmall2RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall2BallastNW, RoadObj::ImageIds::Style1::kRightCurveSmall2SleeperNW, RoadObj::ImageIds::Style1::kRightCurveSmall2RailNW },
// },
// // 0x0040E7E6, 0x0040EB8A, 0x0040EF2C, 0x0040F2D0
// constexpr RoadPaintPiece kRightCurveSmall3 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall3BallastNE, RoadObj::ImageIds::Style1::kRightCurveSmall3SleeperNE, RoadObj::ImageIds::Style1::kRightCurveSmall3RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall3BallastSE, RoadObj::ImageIds::Style1::kRightCurveSmall3SleeperSE, RoadObj::ImageIds::Style1::kRightCurveSmall3RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall3BallastSW, RoadObj::ImageIds::Style1::kRightCurveSmall3SleeperSW, RoadObj::ImageIds::Style1::kRightCurveSmall3RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kRightCurveSmall3BallastNW, RoadObj::ImageIds::Style1::kRightCurveSmall3SleeperNW, RoadObj::ImageIds::Style1::kRightCurveSmall3RailNW },
// },
var RightCurveSmallTPP = [4]RoadPaintPiece{
// kRightCurveSmall0,
// kRightCurveSmall1,
// kRightCurveSmall2,
// kRightCurveSmall3,
const LeftCurveSmall0 RoadPaintPiece = rotateRoadPP(kRightCurveSmall3, kRotationTable1230)
const LeftCurveSmall1 RoadPaintPiece = rotateRoadPP(kRightCurveSmall1, kRotationTable1230)
const LeftCurveSmall2 RoadPaintPiece = rotateRoadPP(kRightCurveSmall2, kRotationTable1230)
const LeftCurveSmall3 RoadPaintPiece = rotateRoadPP(kRightCurveSmall0, kRotationTable1230)
var LeftCurveSmallTPP = [4]RoadPaintPiece{
// kLeftCurveSmall0,
// kLeftCurveSmall1,
// kLeftCurveSmall2,
// kLeftCurveSmall3,
// // 0x0040F409, 0x0040F61F, 0x0040F835, 0x0040FA4B
// constexpr RoadPaintPiece kStraightSlopeUp0 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSlopeUp0BallastNE, RoadObj::ImageIds::Style1::kStraightSlopeUp0SleeperNE, RoadObj::ImageIds::Style1::kStraightSlopeUp0RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSlopeUp0BallastSE, RoadObj::ImageIds::Style1::kStraightSlopeUp0SleeperSE, RoadObj::ImageIds::Style1::kStraightSlopeUp0RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSlopeUp0BallastSW, RoadObj::ImageIds::Style1::kStraightSlopeUp0SleeperSW, RoadObj::ImageIds::Style1::kStraightSlopeUp0RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSlopeUp0BallastNW, RoadObj::ImageIds::Style1::kStraightSlopeUp0SleeperNW, RoadObj::ImageIds::Style1::kStraightSlopeUp0RailNW },
// },
// // 0x0040F512, 0x0040F728, 0x0040F93E, 0x0040FB54
// constexpr RoadPaintPiece kStraightSlopeUp1 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSlopeUp1BallastNE, RoadObj::ImageIds::Style1::kStraightSlopeUp1SleeperNE, RoadObj::ImageIds::Style1::kStraightSlopeUp1RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSlopeUp1BallastSE, RoadObj::ImageIds::Style1::kStraightSlopeUp1SleeperSE, RoadObj::ImageIds::Style1::kStraightSlopeUp1RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSlopeUp1BallastSW, RoadObj::ImageIds::Style1::kStraightSlopeUp1SleeperSW, RoadObj::ImageIds::Style1::kStraightSlopeUp1RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSlopeUp1BallastNW, RoadObj::ImageIds::Style1::kStraightSlopeUp1SleeperNW, RoadObj::ImageIds::Style1::kStraightSlopeUp1RailNW },
// },
var StraightSlopeUpTPP = [2]RoadPaintPiece{
// kStraightSlopeUp0,
// kStraightSlopeUp1,
const StraightSlopeDown0 RoadPaintPiece = rotateRoadPP(kStraightSlopeUp1, kRotationTable2301)
const StraightSlopeDown1 RoadPaintPiece = rotateRoadPP(kStraightSlopeUp0, kRotationTable2301)
var StraightSlopeDownTPP = [2]RoadPaintPiece{
// kStraightSlopeDown0,
// kStraightSlopeDown1,
// // 0x0040FC61, 0x0040FD9E, 0x0040FEDB, 0x00410018
// constexpr RoadPaintPiece kStraightSteepSlopeUp0 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0BallastNE, RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0SleeperNE, RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0BallastSE, RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0SleeperSE, RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0BallastSW, RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0SleeperSW, RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0BallastNW, RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0SleeperNW, RoadObj::ImageIds::Style1::kStraightSteepSlopeUp0RailNW },
// },
var StraightSteepSlopeUpTPP = [1]RoadPaintPiece{
// kStraightSteepSlopeUp0,
const StraightSteepSlopeDown0 RoadPaintPiece = rotateRoadPP(kStraightSteepSlopeUp0, kRotationTable2301)
var StraightSteepSlopeDownTPP = [1]RoadPaintPiece{
// kStraightSteepSlopeDown0,
// // 0x0040E0E9, 0x0040E1F3, 0x0040E2FB, 0x0040E403
// constexpr RoadPaintPiece kTurnaround0 = {
// std::array<std::array<uint32_t, 3>, 4>{
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kTurnaround0BallastNE, RoadObj::ImageIds::Style1::kTurnaround0SleeperNE, RoadObj::ImageIds::Style1::kTurnaround0RailNE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kTurnaround0BallastSE, RoadObj::ImageIds::Style1::kTurnaround0SleeperSE, RoadObj::ImageIds::Style1::kTurnaround0RailSE },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kTurnaround0BallastSW, RoadObj::ImageIds::Style1::kTurnaround0SleeperSW, RoadObj::ImageIds::Style1::kTurnaround0RailSW },
// std::array<uint32_t, 3>{ RoadObj::ImageIds::Style1::kTurnaround0BallastNW, RoadObj::ImageIds::Style1::kTurnaround0SleeperNW, RoadObj::ImageIds::Style1::kTurnaround0RailNW },
// },
var TurnaroundTPP = [1]RoadPaintPiece{
// kTurnaround0,
// constexpr std::array<std::span<const RoadPaintPiece>, 10> kRoadPaintParts = {
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
