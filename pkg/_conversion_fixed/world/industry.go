package world

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Engine/Limits.h"
// #include "Map/Tile.h"
// #include "Map/TileLoop.hpp"
// #include "Types.hpp"
// #include <OpenLoco/Core/BitSet.hpp>
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Core/Prng.h>
// #include <limits>
// #include <span>
// namespace OpenLoco
// forward: struct IndustryObject;
// // TODO: Move this to a different header shared with buildings
type Unk4F9274 struct {
	// World::Pos2 pos;
	Index uint8
}

// const []<const Unk4F9274> getBuildingTileOffsets(bool type);
type IndustryFlags int

const (
func init() {
	None        IndustryFlags = 0
	IsGhost     IndustryFlags = 1 << 0
	Sorted      IndustryFlags = 1 << 1
	ClosingDown IndustryFlags = 1 << 2
	Flag_04     IndustryFlags = 1 << 3

)

// OPENLOCO_ENABLE_ENUM_OPERATORS(IndustryFlags);
type Industry struct {
	Name  StringId
	X     coord_t
	Y     coord_t
	Flags IndustryFlags
	// Core::Prng prng;                                         // 0x08
	ObjectId           uint8
	Under_construction uint8
	FoundingYear       uint16
	NumTiles           uint8
	// World::Pos3 tiles[32];                                   // 0x15 bit 15 of z indicates if multiTile (2x2)
	Town TownId
	// World::TileLoop tileLoop;                                // 0xD7
	NumFarmTiles     int16
	NumIdleFarmTiles int16
	ProductionRate   uint8
	Owner            CompanyId
	StationsInRange  any /* BitSet<Limits::kMaxStations> */
	// StationId producedCargoStatsStation[2][4];               // 0x161
	// uint8 producedCargoStatsRating[2][4];                  // 0x171
	// uint16 dailyProductionTarget[2];                       // 0x179
	// uint16 dailyProduction[2];                             // 0x17D
	// uint16 outputBuffer[2];                                // 0x181
	// uint16 producedCargoQuantityMonthlyTotal[2];           // 0x185
	// uint16 producedCargoQuantityPreviousMonth[2];          // 0x189
	// uint16 receivedCargoQuantityMonthlyTotal[3];           // 0x18D
	// uint16 receivedCargoQuantityPreviousMonth[3];          // 0x193
	// uint16 receivedCargoQuantityDailyTotal[3];             // 0x199
	// uint16 producedCargoQuantityDeliveredMonthlyTotal[2];  // 0x19F
	// uint16 producedCargoQuantityDeliveredPreviousMonth[2]; // 0x1A3
	// uint8 producedCargoPercentTransportedPreviousMonth[2]; // 0x1A7 (%)
	// uint8 producedCargoMonthlyHistorySize[2];              // 0x1A9 (<= 20 * 12)
	// uint8 producedCargoMonthlyHistory1[20 * 12];           // 0x1AB (20 years, 12 months)
	// uint8 producedCargoMonthlyHistory2[20 * 12];           // 0x29B
	// int32 history_min_production[2];                       // 0x38B
	// method: IndustryId id() const;
	// const IndustryObject* getObject() const;
	// method: bool empty() const;
	// method: bool canReceiveCargo() const;
	// method: bool canProduceCargo() const;
	// method: void getStatusString(const char* buffer);
	// method: void update();
	// method: void updateDaily();
	// method: void updateMonthly();
	// method: bool isMonthlyProductionUp();
	// method: bool isMonthlyProductionDown();
	// method: bool isMonthlyProductionClosing();
	// method: void isFarmTileProducing(const World::Pos2& pos);
	// method: void calculateFarmProduction();
	// method: void expandGrounds(const World::Pos2& pos, uint8 primaryWallType, uint8 wallEntranceType, uint8 growthStage, uint8 updateTimer);
	// method: void createMapAnimations();
	// method: void updateProducedCargoStats();
	// method: constexpr bool hasFlags(IndustryFlags flagsToTest) const
	// return (flags & flagsToTest) != IndustryFlags::none;

// func ClaimSurfaceForIndustry(pos World::TilePos2, industryId IndustryId, growthStage uint8, updateTimer uint8) bool
