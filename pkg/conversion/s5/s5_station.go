package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "S5LabelFrame.h"
// #include <OpenLoco/Engine/World.hpp>
// #include <cstdint>
// namespace OpenLoco
// forward: struct Station;
// namespace OpenLoco::S5
type StationCargoStats struct {
	Quantity uint16
	Origin uint16
	Flags uint8
	Age uint8
	Rating uint8
	EnrouteAge uint8
	VehicleSpeed int16
	VehicleAge uint8
	IndustryId uint8
	DensityPerTile uint8
}
// static_assert(sizeof(StationCargoStats) == 0xD);
type Station struct {
	Name uint16
	X coord_t
	Y coord_t
	Z coord_t
	LabelFrame LabelFrame
	Owner uint8
	NoTilesTimeout uint8
	Flags uint16
	Town uint16
// StationCargoStats cargoStats[32]; // 0x2E
	StationTileSize uint16
// World::Pos3 stationTiles[80];     // 0x1D0 Note: z coordinate also contains rotation so always floor
	Var_3B0 uint8
	Var_3B1 uint8
	Var_3B2 uint8
	AirportRotation uint8
// World::Pos3 airportStartPos;           // 0x3B4
	AirportMovementOccupiedEdges uint32
// uint8_t pad_3BE[0x3D2 - 0x3BE];
}
// static_assert(sizeof(Station) == 0x3D2);
// S5::Station exportStation(const OpenLoco::Station& src);
// OpenLoco::Station importStation(const S5::Station& src);
