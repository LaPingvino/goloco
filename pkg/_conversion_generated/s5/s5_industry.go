package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Engine/World.hpp>
// namespace OpenLoco
// forward: struct Industry;
// namespace OpenLoco::S5
type Industry struct {
	Name uint16
	X coord_t
	Y coord_t
	Flags uint16
	Prng0 uint32
	Prng1 uint32
	ObjectId uint8
	Under_construction uint8
	FoundingYear uint16
	NumTiles uint8
// World::Pos3 tiles[32];                                   // 0x15 bit 15 of z indicates if multiTile (2x2)
	Town uint16
// World::Pos2 tileLoop;                                    // 0xD7
	NumFarmTiles int16
	NumIdleFarmTiles int16
	ProductionRate uint8
	Owner uint8
// uint32_t stationsInRange[32];                            // 0xE1 each bit represents one station
// uint16_t producedCargoStatsStation[2][4];                // 0x161
// uint8_t producedCargoStatsRating[2][4];                  // 0x171
// uint16_t dailyProductionTarget[2];                       // 0x179
// uint16_t dailyProduction[2];                             // 0x17D
// uint16_t outputBuffer[2];                                // 0x181
// uint16_t producedCargoQuantityMonthlyTotal[2];           // 0x185
// uint16_t producedCargoQuantityPreviousMonth[2];          // 0x189
// uint16_t receivedCargoQuantityMonthlyTotal[3];           // 0x18D
// uint16_t receivedCargoQuantityPreviousMonth[3];          // 0x193
// uint16_t receivedCargoQuantityDailyTotal[3];             // 0x199
// uint16_t producedCargoQuantityDeliveredMonthlyTotal[2];  // 0x19F
// uint16_t producedCargoQuantityDeliveredPreviousMonth[2]; // 0x1A3
// uint8_t producedCargoPercentTransportedPreviousMonth[2]; // 0x1A7 (%)
// uint8_t producedCargoMonthlyHistorySize[2];              // 0x1A9 (<= 20 * 12)
// uint8_t producedCargoMonthlyHistory1[20 * 12];           // 0x1AB (20 years, 12 months)
// uint8_t producedCargoMonthlyHistory2[20 * 12];           // 0x29B
// int32_t history_min_production[2];                       // 0x38B
// uint8_t pad_393[0x453 - 0x393];
}
// static_assert(sizeof(Industry) == 0x453);
// S5::Industry exportIndustry(const OpenLoco::Industry& src);
// OpenLoco::Industry importIndustry(const S5::Industry& src);
