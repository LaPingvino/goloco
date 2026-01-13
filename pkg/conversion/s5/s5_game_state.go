package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Limits.h"
// #include "S5Animation.h"
// #include "S5Company.h"
// #include "S5Entity.h"
// #include "S5Industry.h"
// #include "S5LabelFrame.h"
// #include "S5Message.h"
// #include "S5Station.h"
// #include "S5Town.h"
// #include "S5Wave.h"
// #include <memory>
// namespace OpenLoco
// forward: struct GameState;
// namespace OpenLoco::S5
type Construction struct {
// uint8_t signals[8];       // 0x00015A (0x00525F72)
// uint8_t bridges[8];       // 0x000162 (0x00525F7A)
// uint8_t trainStations[8]; // 0x00016A (0x00525F82)
// uint8_t trackMods[8];     // 0x000172 (0x00525F8A)
// uint8_t var_17A[8];       // 0x00017A (0x00525F92) NOTE: not used always filled with 0xFF
// uint8_t roadStations[8];  // 0x000182 (0x00525F9A)
// uint8_t roadMods[8];      // 0x00018A (0x00525FA2)
}
// static_assert(sizeof(Construction) == 0x38);
// // We are only splitting this up to make handling the GameState2 struct easier
type GeneralState struct {
// uint32_t rng[2];                                    // 0x000000 (0x00525E18)
// uint32_t unkRng[2];                                 // 0x000008 (0x00525E20)
	Flags uint32
	CurrentDay uint32
	DayCounter uint16
	CurrentYear uint16
	CurrentMonth uint8
	CurrentDayOfMonth uint8
	SavedViewX int16
	SavedViewY int16
	SavedViewZoom uint8
	SavedViewRotation uint8
// uint8_t playerCompanies[2];                         // 0x000024 (0x00525E3C)
// uint16_t entityListHeads[Limits::kNumEntityLists];  // 0x000026 (0x00525E3E)
// uint16_t entityListCounts[Limits::kNumEntityLists]; // 0x000034 (0x00525E4C)
// uint8_t pad_0042[0x046 - 0x042];                    // 0x000042
// uint32_t currencyMultiplicationFactor[32];          // 0x000046 (0x00525E5E)
// uint32_t unusedCurrencyMultiplicationFactor[32];    // 0x0000C6 (0x00525EDE)
	ScenarioTicks uint32
	Var_014A uint16
	ScenarioTicks2 uint32
	MagicNumber uint32
	NumMapAnimations uint16
// World::Pos2 tileUpdateStartLocation;                // 0x000156 (0x00525F6E)
	ScenarioConstruction Construction
	LastRailroadOption uint8
	LastRoadOption uint8
	LastAirport uint8
	LastShipPort uint8
	TrafficHandedness bool
	LastVehicleType uint8
	PickupDirection uint8
	LastTreeOption uint8
	SeaLevel uint16
	CurrentSnowLine uint8
	CurrentSeason uint8
	LastLandOption uint8
	MaxCompetingCompanies uint8
	OrderTableLength uint32
	RoadObjectIdIsAnyRoadTypeCompatible uint32
	RoadObjectIdIsUsableByAllCompanies uint32
	CurrentDefaultLevelCrossingType uint8
	LastTrackTypeOption uint8
	LoanInterestRate uint8
	LastIndustryOption uint8
	LastBuildingOption uint8
	LastMiscBuildingOption uint8
	LastWallOption uint8
	ProduceAICompanyTimeout uint8
// uint32_t tickStartPrngState[2];                     // 0x0001B4 (0x00525FCC)
// char scenarioFileName[256];                         // 0x0001BC (0x00525FD4)
// char scenarioName[64];                              // 0x0002BC (0x005260D4)
// char scenarioDetails[256];                          // 0x0002FC (0x00526114)
	CompetitorStartDelay uint8
	PreferredAIIntelligence uint8
	PreferredAIAggressiveness uint8
	PreferredAICompetitiveness uint8
	StartingLoanSize uint16
	MaxLoanSize uint16
// uint32_t multiplayerPrng[2];                        // 0x000404 (0x0052621C)
	MultiplayerChecksumA uint32
	MultiplayerChecksumB uint32
	LastBuildVehiclesOption uint8
	NumberOfIndustries uint8
	VehiclePreviewRotationFrame uint16
	ObjectiveType uint8
	ObjectiveFlags uint8
	ObjectiveCompanyValue uint32
	ObjectiveMonthlyVehicleProfit uint32
	ObjectivePerformanceIndex uint8
	ObjectiveDeliveredCargoType uint8
	ObjectiveDeliveredCargoAmount uint32
	ObjectiveTimeLimitYears uint8
	ObjectiveTimeLimitUntilYear uint16
	ObjectiveMonthsInChallenge uint16
	ObjectiveCompletedChallengeInMonths uint16
	IndustryFlags uint8
	ForbiddenVehiclesPlayers uint16
	ForbiddenVehiclesCompetitors uint16
	FixFlags uint16
	CompanyRecords Records
	Var_44C uint32
	Var_450 uint32
	Var_454 uint32
	Var_458 uint32
	Var_45C uint32
	Var_460 uint32
	Var_464 uint32
	Var_468 uint32
	LastMapWindowFlags uint32
// uint16_t lastMapWindowSize[2];                      // 0x000470 (0x00526288)
	LastMapWindowVar88A uint16
	LastMapWindowVar88C uint16
	Var_478 uint32
// uint8_t pad_047C[0x13B6 - 0x47C];                   // 0x00047C
	NumMessages uint16
	ActiveMessageIndex uint16
// Message messages[S5::Limits::kMaxMessages];         // 0x0013BA (0x005271D2)
// uint8_t pad_B95A[0xB95C - 0xB95A];                  // 0x00B95A
	Var_B95C uint8
// uint8_t pad_B95D[0xB960 - 0xB95D];                  // 0x00B95D
	Var_B960 uint8
	Pad_B961 uint8
	Var_B962 uint8
	Pad_B963 uint8
	Var_B964 uint8
	Pad_B965 uint8
	Var_B966 uint8
	Pad_B967 uint8
	CurrentRainLevel uint8
// uint8_t pad_B969[0xB96C - 0xB969];                  // 0x00B969
}
// static_assert(sizeof(GeneralState) == 0x00B96C);
type GameState struct {
	General GeneralState
// Company companies[S5::Limits::kMaxCompanies];                                    // 0x00B96C (0x00531784)
// Town towns[S5::Limits::kMaxTowns];                                               // 0x092444 (0x005B825C)
// Industry industries[S5::Limits::kMaxIndustries];                                 // 0x09E744 (0x005C455C)
// Station stations[S5::Limits::kMaxStations];                                      // 0x0C10C4 (0x005E6EDC)
// Entity entities[S5::Limits::kMaxEntities];                                       // 0x1B58C4 (0x006DB6DC)
// Animation animations[S5::Limits::kMaxAnimations];                                // 0x4268C4 (0x0094C6DC)
// Wave waves[S5::Limits::kMaxWaves];                                               // 0x4328C4 (0x009586DC)
// char userStrings[S5::Limits::kMaxUserStrings][32];                               // 0x432A44 (0x0095885C)
// uint16_t routings[S5::Limits::kMaxVehicles][S5::Limits::kMaxRoutingsPerVehicle]; // 0x442A44 (0x0096885C)
// uint8_t orders[S5::Limits::kMaxOrders];                                          // 0x461E44 (0x00987C5C)
}
// static_assert(sizeof(GameState) == 0x4A0644);
type GameStateType2 struct {
	General GeneralState
// CompanyType2 companies[S5::Limits::kMaxCompanies];                               // 0x00B96C
// Town towns[S5::Limits::kMaxTowns];                                               // 0x090824
// Industry industries[S5::Limits::kMaxIndustries];                                 // 0x09CB24
// Station stations[S5::Limits::kMaxStations];                                      // 0x0BF4A4
// Entity entities[S5::Limits::kMaxEntities];                                       // 0x1B3CA4
// Animation animations[S5::Limits::kMaxAnimations];                                // 0x424CA4
// Wave waves[S5::Limits::kMaxWaves];                                               // 0x430CA4
// char userStrings[S5::Limits::kMaxUserStrings][32];                               // 0x430E24
// uint16_t routings[S5::Limits::kMaxVehicles][S5::Limits::kMaxRoutingsPerVehicle]; // 0x440E24
// uint8_t orders[S5::Limits::kMaxOrders];                                          // 0x460224
}
// static_assert(sizeof(GameStateType2) == 0x49EA24);
// std::unique_ptr<S5::GameState> exportGameState(const OpenLoco::GameState& src);
// std::unique_ptr<S5::GameState> importGameStateType2(const S5::GameStateType2& src);
// std::unique_ptr<OpenLoco::GameState> importGameState(const S5::GameState& src);
