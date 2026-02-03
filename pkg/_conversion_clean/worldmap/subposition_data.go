package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Types.hpp"
// #include <OpenLoco/Engine/World.hpp>
// #include <span>
// namespace OpenLoco
type Pitch int

// SKIPPED CONSTRUCTOR: const ()

// namespace OpenLoco::World::TrackData
type MoveInfo struct {
	// World::Pos3 loc; // 0x00
	Yaw   uint8
	Pitch Pitch
}

// static_assert(sizeof(MoveInfo) == 0x8);
// []<const MoveInfo> getTrackSubPositon(const uint16 trackAndDirection);
// []<const MoveInfo> getRoadSubPositon(const uint16 trackAndDirection);
// []<const MoveInfo> getRoadPlacementSubPositon(const uint16 trackAndDirection);
