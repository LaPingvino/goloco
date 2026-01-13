package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Economy/Currency.h"
// #include <OpenLoco/Engine/World.hpp>
// namespace OpenLoco
// forward: struct Company;
// namespace CompanyManager
// forward: struct Records;
// namespace OpenLoco::S5
type AiThought struct {
}

type AiThoughtStation struct {
	Id uint16
	Var_02 uint8
	Rotation uint8
// World::Pos2 pos;  // 0x4
	BaseZ uint8
	Var_9 uint8
	Var_A uint8
	Var_B uint8
	Var_C uint8
// uint8_t pad_D[0xE - 0xD];
}
// static_assert(sizeof(Station) == 0xE);
	Type uint8
	DestinationA uint8
	DestinationB uint8
	NumStations uint8
	StationLength uint8
	Pad_05 uint8
// Station stations[4];  // 0x06 0x4AE Will lists stations created that vehicles will route to
	TrackObjId uint8
	RackRailType uint8
	Mods uint16
	CargoType uint8
	Var_43 uint8
	NumVehicles uint8
	Var_45 uint8
// uint16_t var_46[16];  // 0x4EF array of uint16_t object id
// uint16_t vehicles[8]; // 0x66 0x50E see also numVehicles for current size
	Var_76 currency32_t
// uint8_t pad_7A[0x7C - 0x7A];
	Var_7C currency32_t
	Var_80 currency32_t
	Var_84 currency32_t
	Var_88 uint8
	StationObjId uint8
	SignalObjId uint8
	PurchaseFlags uint8
}
type Company struct {
}

type CompanyHashTableEntry struct {
	Var_00 uint16
	Var_02 uint16
	Var_04 uint8
	Var_05 uint8
}
	Name uint16
	OwnerName uint16
	ChallengeFlags uint32
	Cash currency48_t
	CurrentLoan currency32_t
	UpdateCounter uint32
	PerformanceIndex int16
	CompetitorId uint8
	OwnerEmotion uint8
// uint8_t mainColours[2];            // 0x1A
// uint8_t vehicleColours[10][2];     // 0x1C
	CustomVehicleColoursSet uint32
// uint32_t unlockedVehicles[7];      // 0x34
	AvailableVehicles uint16
	AiPlaystyleFlags uint32
	AiPlaystyleTownId uint8
	NumExpenditureYears uint8
// currency32_t expenditures[16][17]; // 0x58
	StartedDate uint32
	Var_49C uint32
	Var_4A0 uint32
	Var_4A4 uint8
	Var_4A5 uint8
	Var_4A6 uint8
	Var_4A7 uint8
// AiThought aiThoughts[60];                  // 0x04A8
	ActiveThoughtId uint8
// World::SmallZ headquartersZ;               // 0x2579
	HeadquartersX coord_t
	HeadquartersY coord_t
	ActiveThoughtRevenueEstimate currency32_t
	Var_2582 uint32
// uint8_t pad_2586[0x2596 - 0x2586];
	Var_2596 uint32
	Var_259A uint8
	Var_259B uint8
	Var_259C uint8
	Pad_259D uint8
	AiPlaceVehicleIndex uint32
// uint8_t pad_25A2[0x25BE - 0x25A2];
	Var_25BE uint8
	CurrentRating uint8
// HashTableEntry var_25C0[0x1000]; // 0x25C0 Hash table entries
	Var_25C0_length uint16
	Var_85C2 uint8
	Var_85C3 uint8
// World::Pos2 var_85C4;
// World::SmallZ var_85C8;
// World::Pos2 var_85C9;
// World::SmallZ var_85CD;
	Var_85CE uint8
	Var_85CF uint8
// World::Pos2 var_85D0;
// World::SmallZ var_85D4;
	Var_85D5 uint16
// World::Pos2 var_85D7;
// World::SmallZ var_85DB;
	Var_85DC uint16
	Var_85DE uint32
	Var_85E2 uint32
	Var_85E6 uint16
	Var_85E8 uint16
	Var_85EA uint32
	Var_85EE uint8
	Var_85EF uint8
	Var_85F0 uint16
	Var_85F2 currency32_t
	Var_85F6 uint16
	CargoUnitsTotalDelivered uint32
// uint32_t cargoUnitsDeliveredHistory[120]; // 0x85FC
// int16_t performanceIndexHistory[120];     // 0x87DC
	HistorySize uint16
// currency48_t companyValueHistory[120];    // 0x88CE
	VehicleProfit currency48_t
// uint16_t transportTypeCount[6];           // 0x8BA4
// uint8_t activeEmotions[9];                // 0x8BB0 duration in days that emotion is active 0 == not active
	ObservationStatus uint8
	ObservationTownId uint16
	ObservationEntity uint16
	ObservationX int16
	ObservationY int16
	ObservationObject uint16
	ObservationTimeout uint16
// uint16_t ownerStatus[2];                  // 0x8BC6
// uint8_t pad_8BCA[0x8BCE - 0x8BCA];
// uint32_t cargoDelivered[32];             // 0x8BCE;
	ChallengeProgress uint8
	NumMonthsInTheRed uint8
	CargoUnitsTotalDistance uint32
// uint32_t cargoUnitsDistanceHistory[120]; // 0x8C54
	JailStatus uint16
// uint8_t pad_8E36[0x8FA8 - 0x8E36];
}
// static_assert(sizeof(Company) == 0x8FA8);
type CompanyType2 struct {
}

type CompanyType2HashTableEntry struct {
	Var_00 uint16
	Var_02 uint16
	Var_04 uint8
	Var_05 uint8
}
	Name uint16
	OwnerName uint16
	ChallengeFlags uint32
	Cash currency48_t
	CurrentLoan currency32_t
	UpdateCounter uint32
	PerformanceIndex int16
	CompetitorId uint8
	OwnerEmotion uint8
// uint8_t mainColours[2];            // 0x1A
// uint8_t vehicleColours[10][2];     // 0x1C
	CustomVehicleColoursSet uint32
// uint32_t unlockedVehicles[7];      // 0x34
	AvailableVehicles uint16
	AiPlaystyleFlags uint32
	AiPlaystyleTownId uint8
	NumExpenditureYears uint8
// currency32_t expenditures[16][17]; // 0x58
	StartedDate uint32
	Var_49C uint32
	Var_4A0 uint32
	Var_4A4 uint8
	Var_4A5 uint8
	Var_4A6 uint8
	Var_4A7 uint8
// AiThought aiThoughts[60];                  // 0x04A8
	ActiveThoughtId uint8
// World::SmallZ headquartersZ;               // 0x2579
	HeadquartersX coord_t
	HeadquartersY coord_t
	ActiveThoughtRevenueEstimate currency32_t
	Var_2582 uint32
// uint8_t pad_2586[0x2596 - 0x2586];
	Var_2596 uint32
	Var_259A uint8
	Var_259B uint8
	Var_259C uint8
	Pad_259D uint8
	AiPlaceVehicleIndex uint32
// uint8_t pad_25A2[0x25BE - 0x25A2];
	Var_25BE uint8
	CurrentRating uint8
// HashTableEntry var_25C0[0x1000]; // 0x25C0 Hash table entries
	Var_25C0_length uint16
	Var_85C2 uint8
	Var_85C3 uint8
// World::Pos2 var_85C4;
// World::SmallZ var_85C8;
// World::Pos2 var_85C9;
// World::SmallZ var_85CD;
	Var_85CE uint8
	Var_85CF uint8
// World::Pos2 var_85D0;
// World::SmallZ var_85D4;
	Var_85D5 uint16
// World::Pos2 var_85D7;
// World::SmallZ var_85DB;
	Var_85DC uint16
	Var_85DE uint32
	Var_85E2 uint32
	Var_85E6 uint16
	Var_85E8 uint16
	Var_85EA uint32
	Var_85EE uint8
	Var_85EF uint8
	Var_85F0 uint16
	Var_85F2 currency32_t
	Var_85F6 uint16
	CargoUnitsTotalDelivered uint32
// uint32_t cargoUnitsDeliveredHistory[120]; // 0x85FC
// int16_t performanceIndexHistory[120];     // 0x87DC
	HistorySize uint16
// currency48_t companyValueHistory[120];    // 0x88CE
	VehicleProfit currency48_t
// uint16_t transportTypeCount[6];           // 0x8BA4
// uint8_t activeEmotions[9];                // 0x8BB0 duration in days that emotion is active 0 == not active
	ObservationStatus uint8
	ObservationTownId uint16
	ObservationEntity uint16
	ObservationX int16
	ObservationY int16
	ObservationObject uint16
	ObservationTimeout uint16
// uint16_t ownerStatus[2];                  // 0x8BC6
// uint8_t pad_8BCA[0x8BCE - 0x8BCA];
// uint32_t cargoDelivered[32];      // 0x8BCE;
	ChallengeProgress uint8
	NumMonthsInTheRed uint8
	CargoUnitsTotalDistance uint32
	JailStatus uint16
// uint8_t pad_8E36[0x8DC8 - 0x8C56];
}
// static_assert(sizeof(CompanyType2) == 0x8DC8);
type Records struct {
// int16_t speed[3];   // 0x000436 (0x0052624E)
// uint8_t company[3]; // 0x00043C (0x00526254)
	Pad_43A uint8
// uint32_t date[3]; // 0x000440 (0x00526258)
}
// S5::Company exportCompany(const OpenLoco::Company& src);
// S5::Company importCompanyType2(const S5::CompanyType2& src);
// S5::Records exportRecords(const OpenLoco::CompanyManager::Records& src);
// OpenLoco::CompanyManager::Records importRecords(const S5::Records& src);
// OpenLoco::Company importCompany(const S5::Company& src);
