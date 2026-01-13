package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "S5/S5LabelFrame.h"
// #include <OpenLoco/Engine/World.hpp>
// namespace OpenLoco
// forward: struct Town;
// namespace OpenLoco::S5
type Town struct {
	Name uint16
	X coord_t
	Y coord_t
	Flags uint16
	LabelFrame LabelFrame
	Prng0 uint32
	Prng1 uint32
	Population uint32
	PopulationCapacity uint32
	NumBuildings int16
// int16_t companyRatings[15];   // 0x3A
	CompaniesWithRating uint16
	Size uint8
	HistorySize uint8
// uint8_t history[20 * 12];     // 0x5C (20 years, 12 months)
	HistoryMinPopulation int32
// uint8_t var_150[8];
// uint16_t monthlyCargoDelivered[32]; // 0x158
	CargoInfluenceFlags uint32
// uint16_t var_19C[2][2];
	BuildSpeed uint8
	NumberOfAirports uint8
	NumStations uint16
	Var_1A8 uint32
// uint8_t pad_1AC[0x270 - 0x1AC];
}
// S5::Town exportTown(const OpenLoco::Town& src);
// OpenLoco::Town importTown(const S5::Town& src);
