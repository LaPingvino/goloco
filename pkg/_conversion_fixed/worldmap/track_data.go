package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Map/QuarterTile.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <array>
// #include <cstddef>
// #include <cstdlib>
// #include <span>
// namespace OpenLoco::World::Track
type TrackTraitFlags int

const (
type RoadTraitFlags int

type CommonTraitFlags int

)
// namespace OpenLoco::World::TrackData
type ConnectionsByRotation = [4]uint8
type PreviewTrackFlags int

func init() {
	None PreviewTrackFlags = 0
	Unk0 PreviewTrackFlags = 1 << 0
	Unk1 PreviewTrackFlags = 1 << 1
	Unk2 PreviewTrackFlags = 1 << 2
	Unk3 PreviewTrackFlags = 1 << 3
	Unk4 PreviewTrackFlags = 1 << 4
	Unused PreviewTrackFlags = 1 << 6
	Diagonal PreviewTrackFlags = 1 << 7
}

// OPENLOCO_ENABLE_ENUM_OPERATORS(PreviewTrackFlags);
type PreviewTrack struct {
	Index uint8
	X int16
	Y int16
	Z int16
	ClearZ uint8
	SubTileClearance QuarterTile
	Flags PreviewTrackFlags
	ConnectFlags ConnectionsByRotation
	// method: constexpr bool hasFlags(PreviewTrackFlags flagsToTest) const
// return (flags & flagsToTest) != PreviewTrackFlags::none;
// // Pos is difference from the next first tile and the track first tile
type TrackCoordinates struct {
	RotationBegin uint8
	RotationEnd uint8
// World::Pos3 pos;       // 0x02
// const []<const PreviewTrack> getTrackPiece(size_t trackId);
// const []<const PreviewTrack> getRoadPiece(size_t trackId);
// const TrackCoordinates& getUnkTrack(uint16 trackAndDirection);
// const TrackCoordinates& getUnkRoad(uint16 trackAndDirection);
type TrackMiscData struct {
	CostFactor uint16
// Track::CommonTraitFlags flags;          // 0x004F8764
	ReverseTrackId uint8
	ReverseRotation uint8
	SignalHeightOffsetLeft uint8
	SignalHeightOffsetRight uint8
// Track::TrackTraitFlags compatibleFlags; // 0x004F891C
	CurveSpeedFraction uint16
	UnkWeighting uint32
	SparkDirection bool
type RoadMiscData struct {
// Track::CommonTraitFlags flags;         // 0x004F7284
	ReverseRoadId uint8
	ReverseLane uint8
// Track::RoadTraitFlags compatibleFlags; // 0x004F72E8
// const TrackMiscData& getTrackMiscData(size_t trackId);
// const RoadMiscData& getRoadMiscData(size_t roadId);
type RoadUnkNextTo struct {
// World::Pos3 pos;  // 0x00 (was 3x int8)
	Rotation uint8
// []<const RoadUnkNextTo> getRoadUnkNextTo(uint16 trackAndDirection);
// // 0x004F7358
// // Returns the lane occupation mask for a given road trackAndDirection value.
// // The index is (trackAndDirection._data >> 2), which encodes:
// // - roadId (track piece identifier)
// // - isBackwards (direction of travel)
// // - isOvertakeLane (which lane the vehicle is in)
// // - isChangingLane (whether the vehicle is in the process of changing lanes)
// // The returned value's upper nibble (>> 4) contains a 2-bit lane occupation mask.
// func GetRoadOccupationMask(index uint8) uint8
// // 0x004F865C
// // The index is (trackAndDirection._track.data >> 2), which encodes:
// // - reversed (bit 0)
// // - track id (bit 1-6)
// func GetCurvatureDegree(index uint8) int8
