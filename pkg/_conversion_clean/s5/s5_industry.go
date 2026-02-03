package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Engine/World.hpp>
// namespace OpenLoco
// forward: struct Industry;
// namespace OpenLoco::S5
type Industry struct {
	Name               uint16
	X                  coord_t
	Y                  coord_t
	Flags              uint16
	Prng0              uint32
	Prng1              uint32
	ObjectId           uint8
	Under_construction uint8
	FoundingYear       uint16
	NumTiles           uint8
	// World::Pos3 tiles[32];                                   // 0x15 bit 15 of z indicates if multiTile (2x2)
	Town uint16
	// World::Pos2 tileLoop;                                    // 0xD7
	NumFarmTiles     int16
	NumIdleFarmTiles int16
	ProductionRate   uint8
	Owner            uint8
	// uint32 stationsInRange[32];                            // 0xE1 each bit represents one station
	// uint16 producedCargoStatsStation[2][4];                // 0x161
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
	// uint8 pad_393[0x453 - 0x393];
}

// static_assert(sizeof(Industry) == 0x453);
// S5::Industry exportIndustry(const OpenLoco::Industry& src);
// OpenLoco::Industry importIndustry(const S5::Industry& src);
