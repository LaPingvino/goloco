package paint

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Objects/RoadExtraObject.h"
// #include "Paint.h"
// #include <OpenLoco/Core/Numerics.hpp>
// #include <array>
// #include <optional>
// #include <span>
// namespace OpenLoco::Paint::AdditionStyle1
type RoadAdditionSupport struct {
// []<[]<uint32, 4>, 4> imageIds;
	Height int16
// []<uint8, 4> frequencies;   // Make array
// []<SegmentFlags, 4> segments; // Make array
// constexpr RoadAdditionSupport(
// const []<[]<uint32, 4>, 4>& _imageIds,
// const int16 _height,
// const []<uint8, 4>& _frequencies,
// const []<SegmentFlags, 4>& _segments)
// : imageIds(_imageIds)
// , height(_height)
// , frequencies(_frequencies)
// , segments(_segments)
}
// constexpr RoadAdditionSupport(
// const []<[]<uint32, 4>, 4>& _imageIds,
// const int16 _height,
// const uint8 _frequency,
// const SegmentFlags _segment)
// : imageIds(_imageIds)
// , height(_height)
// , frequencies()
// , segments()
// frequencies[0] = _frequency;
// frequencies[1] = Numerics::rotl4bit(frequencies[0], 2);
// frequencies[2] = frequencies[0];
// frequencies[3] = frequencies[1];
// segments[0] = _segment;
// for (auto i = 1U; i < 4; ++i)
// segments[i] = rotlSegmentFlags(segments[0], i);
type RoadPaintAdditionPiece struct {
// []<uint32, 4> imageIds;
// []<World::Pos3, 4> boundingBoxOffsets;
// []<World::Pos3, 4> boundingBoxSizes;
	IsIsMergeable bool
// std::optional<RoadAdditionSupport> supports;
func init() {
RotationTable1230 = [4]uint8{
RotationTable2301 = [4]uint8{
RotationTable3012 = [4]uint8{

const NullRoadPaintAdditionPiece RoadPaintAdditionPiece = {}
func init() {
NoSupports = std.nullopt // auto

// consteval std::optional<RoadAdditionSupport> rotateRoadPPASupport(const std::optional<RoadAdditionSupport>& reference, const []<uint8, 4>& rotationTable)
// if (!reference.has_value())
// return std::nullopt;
// return RoadAdditionSupport{
// []<[]<uint32, 4>, 4>{
// reference->imageIds[rotationTable[0]],
// reference->imageIds[rotationTable[1]],
// reference->imageIds[rotationTable[2]],
// reference->imageIds[rotationTable[3]],
// },
// reference->height,
// []<uint8, 4>{
// reference->frequencies[rotationTable[0]],
// reference->frequencies[rotationTable[1]],
// reference->frequencies[rotationTable[2]],
// reference->frequencies[rotationTable[3]],
// },
// []<SegmentFlags, 4>{
// reference->segments[rotationTable[0]],
// reference->segments[rotationTable[1]],
// reference->segments[rotationTable[2]],
// reference->segments[rotationTable[3]],
// consteval RoadPaintAdditionPiece rotateRoadPPA(const RoadPaintAdditionPiece& reference, const []<uint8, 4>& rotationTable)
// return RoadPaintAdditionPiece{
// []<uint32, 4>{
// reference.imageIds[rotationTable[0]],
// reference.imageIds[rotationTable[1]],
// reference.imageIds[rotationTable[2]],
// reference.imageIds[rotationTable[3]],
// },
// []<World::Pos3, 4>{
// reference.boundingBoxOffsets[rotationTable[0]],
// reference.boundingBoxOffsets[rotationTable[1]],
// reference.boundingBoxOffsets[rotationTable[2]],
// reference.boundingBoxOffsets[rotationTable[3]],
// },
// []<World::Pos3, 4>{
// reference.boundingBoxSizes[rotationTable[0]],
// reference.boundingBoxSizes[rotationTable[1]],
// reference.boundingBoxSizes[rotationTable[2]],
// reference.boundingBoxSizes[rotationTable[3]],
// },
// reference.isIsMergeable,
// rotateRoadPPASupport(reference.supports, rotationTable)
// using namespace OpenLoco::RoadExtraObj::ImageIds::Style1;
// // 0x00410159, 0x0041020A, 0x00410159, 0x0041020A
// constexpr RoadPaintAdditionPiece kStraightAddition0 = {
// /* ImageIds */ []<uint32, 4>{
// kStraight0NE,
// kStraight0SE,
// kStraight0NE,
// kStraight0SE,
// },
// /* BoundingBoxOffsets */ []<World::Pos3, 4>{
// World::Pos3{ 2, 2, 26 },
// World::Pos3{ 2, 2, 26 },
// World::Pos3{ 2, 2, 26 },
// World::Pos3{ 2, 2, 26 },
// },
// /* BoundingBoxSizes */ []<World::Pos3, 4>{
// World::Pos3{ 28, 28, 1 },
// World::Pos3{ 28, 28, 1 },
// World::Pos3{ 28, 28, 1 },
// World::Pos3{ 28, 28, 1 },
// },
// /* Mergable */ true,
// /* Supports */ RoadAdditionSupport{
// /* ImageIds */ []<[]<uint32, 4>, 4>{
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// },
// /* SupportHeight */ 0,
// /* Frequency */ 2,
// /* Segment */ SegmentFlags::x1y0 | SegmentFlags::x1y2,
// },
func init() {
StraightTPPA = [1]RoadPaintAdditionPiece{

// kStraightAddition0,
// // 0x004102BB, 0x00410302, 0x00410349, 0x00410390
// constexpr RoadPaintAdditionPiece kRightCurveVerySmallAddition0 = {
// /* ImageIds */ []<uint32, 4>{
// kRightCurveVerySmall0NE,
// kRightCurveVerySmall0SE,
// kRightCurveVerySmall0SW,
// kRightCurveVerySmall0NW,
// },
// /* BoundingBoxOffsets */ []<World::Pos3, 4>{
// World::Pos3{ 2, 2, 26 },
// World::Pos3{ 2, 2, 26 },
// World::Pos3{ 2, 2, 26 },
// World::Pos3{ 2, 2, 26 },
// },
// /* BoundingBoxSizes */ []<World::Pos3, 4>{
// World::Pos3{ 28, 28, 1 },
// World::Pos3{ 28, 28, 1 },
// World::Pos3{ 28, 28, 1 },
// World::Pos3{ 28, 28, 1 },
// },
// /* Mergable */ true,
// /* Supports */ kNoSupports,
func init() {
RightCurveVerySmallTPPA = [1]RoadPaintAdditionPiece{

// kRightCurveVerySmallAddition0,
const LeftCurveVerySmallAddition0 RoadPaintAdditionPiece = rotateRoadPPA(kRightCurveVerySmallAddition0, kRotationTable1230)
func init() {
LeftCurveVerySmallTPPA = [1]RoadPaintAdditionPiece{

// kLeftCurveVerySmallAddition0,
// // 0x0041052B, 0x0041068F, 0x004107F3, 0x00410957
// constexpr RoadPaintAdditionPiece kRightCurveSmallAddition0 = {
// /* ImageIds */ []<uint32, 4>{
// kRightCurveSmall0NE,
// kRightCurveSmall0SE,
// kRightCurveSmall0SW,
// kRightCurveSmall0NW,
// },
// /* BoundingBoxOffsets */ []<World::Pos3, 4>{
// World::Pos3{ 2, 6, 26 },
// World::Pos3{ 6, 2, 26 },
// World::Pos3{ 2, 6, 26 },
// World::Pos3{ 6, 2, 26 },
// },
// /* BoundingBoxSizes */ []<World::Pos3, 4>{
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// },
// /* Mergable */ true,
// /* Supports */ RoadAdditionSupport{
// /* ImageIds */ []<[]<uint32, 4>, 4>{
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// },
// /* SupportHeight */ 0,
// /* Frequency */ 0,
// /* Segment */ SegmentFlags::x1y0 | SegmentFlags::x1y2,
// },
const RightCurveSmallAddition1 RoadPaintAdditionPiece = kNullRoadPaintAdditionPiece
const RightCurveSmallAddition2 RoadPaintAdditionPiece = kNullRoadPaintAdditionPiece
// // 0x004105DE, 0x00410742, 0x004108A6, 0x00410A0A
// constexpr RoadPaintAdditionPiece kRightCurveSmallAddition3 = {
// /* ImageIds */ []<uint32, 4>{
// kRightCurveSmall3NE,
// kRightCurveSmall3SE,
// kRightCurveSmall3SW,
// kRightCurveSmall3NW,
// },
// /* BoundingBoxOffsets */ []<World::Pos3, 4>{
// World::Pos3{ 6, 2, 26 },
// World::Pos3{ 2, 6, 26 },
// World::Pos3{ 6, 2, 26 },
// World::Pos3{ 2, 6, 26 },
// },
// /* BoundingBoxSizes */ []<World::Pos3, 4>{
// World::Pos3{ 20, 28, 1 },
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// World::Pos3{ 28, 20, 1 },
// },
// /* Mergable */ true,
// /* Supports */ RoadAdditionSupport{
// /* ImageIds */ []<[]<uint32, 4>, 4>{
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// },
// /* SupportHeight */ 0,
// /* Frequency */ 0,
// /* Segment */ SegmentFlags::x0y1 | SegmentFlags::x2y1,
// },
func init() {
RightCurveSmallTPPA = [4]RoadPaintAdditionPiece{

// kRightCurveSmallAddition0,
// kRightCurveSmallAddition1,
// kRightCurveSmallAddition2,
// kRightCurveSmallAddition3,
const LeftCurveSmallAddition0 RoadPaintAdditionPiece = rotateRoadPPA(kRightCurveSmallAddition3, kRotationTable1230)
const LeftCurveSmallAddition1 RoadPaintAdditionPiece = kNullRoadPaintAdditionPiece
const LeftCurveSmallAddition2 RoadPaintAdditionPiece = kNullRoadPaintAdditionPiece
const LeftCurveSmallAddition3 RoadPaintAdditionPiece = rotateRoadPPA(kRightCurveSmallAddition0, kRotationTable1230)
func init() {
LeftCurveSmallTPPA = [4]RoadPaintAdditionPiece{

// kLeftCurveSmallAddition0,
// kLeftCurveSmallAddition1,
// kLeftCurveSmallAddition2,
// kLeftCurveSmallAddition3,
// // 0x00410AF3, 0x00410C83, 0x00410E13, 0x00410FA3
// constexpr RoadPaintAdditionPiece kStraightSlopeUpAddition0 = {
// /* ImageIds */ []<uint32, 4>{
// kStraightSlopeUp0NE,
// kStraightSlopeUp0SE,
// kStraightSlopeUp0SW,
// kStraightSlopeUp0NW,
// },
// /* BoundingBoxOffsets */ []<World::Pos3, 4>{
// World::Pos3{ 2, 6, 34 },
// World::Pos3{ 6, 2, 34 },
// World::Pos3{ 2, 6, 34 },
// World::Pos3{ 6, 2, 34 },
// },
// /* BoundingBoxSizes */ []<World::Pos3, 4>{
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// },
// /* Mergable */ false,
// /* Supports */ RoadAdditionSupport{
// /* ImageIds */ []<[]<uint32, 4>, 4>{
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// },
// /* SupportHeight */ 4,
// /* Frequency */ 1,
// /* Segment */ SegmentFlags::x1y0 | SegmentFlags::x1y2,
// },
// // 0x00410BBB, 0x00410D4B, 0x00410EDB, 0x0041106B
// constexpr RoadPaintAdditionPiece kStraightSlopeUpAddition1 = {
// /* ImageIds */ []<uint32, 4>{
// kStraightSlopeUp1NE,
// kStraightSlopeUp1SE,
// kStraightSlopeUp1SW,
// kStraightSlopeUp1NW,
// },
// /* BoundingBoxOffsets */ []<World::Pos3, 4>{
// World::Pos3{ 2, 6, 34 },
// World::Pos3{ 6, 2, 34 },
// World::Pos3{ 2, 6, 34 },
// World::Pos3{ 6, 2, 34 },
// },
// /* BoundingBoxSizes */ []<World::Pos3, 4>{
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// },
// /* Mergable */ false,
// /* Supports */ RoadAdditionSupport{
// /* ImageIds */ []<[]<uint32, 4>, 4>{
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// },
// /* SupportHeight */ 12,
// /* Frequency */ 1,
// /* Segment */ SegmentFlags::x1y0 | SegmentFlags::x1y2,
// },
func init() {
StraightSlopeUpTPPA = [2]RoadPaintAdditionPiece{

// kStraightSlopeUpAddition0,
// kStraightSlopeUpAddition1,
const StraightSlopeDownAddition0 RoadPaintAdditionPiece = rotateRoadPPA(kStraightSlopeUpAddition1, kRotationTable2301)
const StraightSlopeDownAddition1 RoadPaintAdditionPiece = rotateRoadPPA(kStraightSlopeUpAddition0, kRotationTable2301)
func init() {
StraightSlopeDownTPPA = [2]RoadPaintAdditionPiece{

// kStraightSlopeDownAddition0,
// kStraightSlopeDownAddition1,
// // 0x00411133, 0x004111FB, 0x004112C3, 0x0041138B
// constexpr RoadPaintAdditionPiece kStraightSteepSlopeUpAddition0 = {
// /* ImageIds */ []<uint32, 4>{
// kStraightSteepSlopeUp0NE,
// kStraightSteepSlopeUp0SE,
// kStraightSteepSlopeUp0SW,
// kStraightSteepSlopeUp0NW,
// },
// /* BoundingBoxOffsets */ []<World::Pos3, 4>{
// World::Pos3{ 2, 6, 34 },
// World::Pos3{ 6, 2, 34 },
// World::Pos3{ 2, 6, 34 },
// World::Pos3{ 6, 2, 34 },
// },
// /* BoundingBoxSizes */ []<World::Pos3, 4>{
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// World::Pos3{ 28, 20, 1 },
// World::Pos3{ 20, 28, 1 },
// },
// /* Mergable */ false,
// /* Supports */ RoadAdditionSupport{
// /* ImageIds */ []<[]<uint32, 4>, 4>{
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// []<uint32, 4>{ kSupportStraight0NE, kSupportConnectorStraight0NE, kSupportStraight0SW, kSupportConnectorStraight0SW },
// []<uint32, 4>{ kSupportStraight0SE, kSupportConnectorStraight0SE, kSupportStraight0NW, kSupportConnectorStraight0NW },
// },
// /* SupportHeight */ 8,
// /* Frequency */ 1,
// /* Segment */ SegmentFlags::x1y0 | SegmentFlags::x1y2,
// },
func init() {
StraightSteepSlopeUpTPPA = [1]RoadPaintAdditionPiece{

// kStraightSteepSlopeUpAddition0,
const StraightSteepSlopeDownAddition0 RoadPaintAdditionPiece = rotateRoadPPA(kStraightSteepSlopeUpAddition0, kRotationTable2301)
func init() {
StraightSteepSlopeDownTPPA = [1]RoadPaintAdditionPiece{

// kStraightSteepSlopeDownAddition0,
// // 0x004103D7, 0x0041041E, 0x00410465, 0x004104AC
// constexpr RoadPaintAdditionPiece kTurnaroundAddition0 = {
// /* ImageIds */ []<uint32, 4>{
// kTurnaround0NE,
// kTurnaround0SE,
// kTurnaround0SW,
// kTurnaround0NW,
// },
// /* BoundingBoxOffsets */ []<World::Pos3, 4>{
// World::Pos3{ 16, 2, 26 },
// World::Pos3{ 2, 2, 26 },
// World::Pos3{ 2, 2, 26 },
// World::Pos3{ 2, 16, 26 },
// },
// /* BoundingBoxSizes */ []<World::Pos3, 4>{
// World::Pos3{ 14, 28, 1 },
// World::Pos3{ 28, 14, 1 },
// World::Pos3{ 14, 28, 1 },
// World::Pos3{ 28, 14, 1 },
// },
// /* Mergable */ true,
// /* Supports */ kNoSupports,
func init() {
TurnaroundTPPA = [1]RoadPaintAdditionPiece{

// kTurnaroundAddition0,
// constexpr []<[]<const RoadPaintAdditionPiece>, 10> kRoadPaintAdditionParts = {
// kStraightTPPA,
// kLeftCurveVerySmallTPPA,
// kRightCurveVerySmallTPPA,
// kLeftCurveSmallTPPA,
// kRightCurveSmallTPPA,
// kStraightSlopeUpTPPA,
// kStraightSlopeDownTPPA,
// kStraightSteepSlopeUpTPPA,
// kStraightSteepSlopeDownTPPA,
// kTurnaroundTPPA,
