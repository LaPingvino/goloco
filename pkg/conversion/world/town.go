package world

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Company.h"
// #include "LabelFrame.h"
// #include "Map/Tile.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Core/Prng.h>
// #include <limits>
// #include <optional>
// namespace OpenLoco
const MinCompanyRating int32 = -1000
const MaxCompanyRating int32 = 1000
type TownFlags int

const (
	None TownFlags = 0
	Sorted TownFlags = 1 << 0
	RatingAdjusted TownFlags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TownFlags);
type TownGrowFlags int

const (
	None TownGrowFlags = 0
	BuildInitialRoad TownGrowFlags = 1 << 0
	RoadUpdate TownGrowFlags = 1 << 1
	NeutralRoadTakeover TownGrowFlags = 1 << 2
	AllowRoadExpansion TownGrowFlags = 1 << 3
	AllowRoadBranching TownGrowFlags = 1 << 4
	ConstructBuildings TownGrowFlags = 1 << 5
	BuildImmediately TownGrowFlags = 1 << 6
	AlwaysUpdateBuildings TownGrowFlags = 1 << 7
	All TownGrowFlags = buildInitialRoad | roadUpdate | neutralRoadTakeover | allowRoadExpansion | allowRoadBranching | constructBuildings | buildImmediately | alwaysUpdateBuildings
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(TownGrowFlags);
type TownSize int

const (
	Hamlet TownSize = iota
	Village
	Town
	City
	Metropolis
)
// namespace Gfx
// forward: class DrawingContext;
// forward: struct RenderTarget;
type Town struct {
	Name StringId
	X coord_t
	Y coord_t
	Flags TownFlags
	LabelFrame LabelFrame
// Core::Prng prng;              // 0x28
	Population uint32
	PopulationCapacity uint32
	NumBuildings int16
// int16_t companyRatings[15];   // 0x3A
	CompaniesWithRating uint16
	Size TownSize
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
	// method: bool empty() const;
	// method: TownId id() const;
	// method: void update();
	// method: void drawLabel(Gfx::DrawingContext& drawingCtx, const Gfx::RenderTarget& rt);
	// method: void updateLabel();
	// method: void updateMonthly();
	// method: void adjustCompanyRating(CompanyId cid, int amount);
	// method: void recalculateSize();
	// method: void grow(TownGrowFlags growFlags);
	// method: StringId getTownSizeString() const;
}
