package world

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "CompanyAi/CompanyAi.h"
// #include "Economy/Currency.h"
// #include "Economy/Expenditures.h"
// #include "Engine/Limits.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/BitSet.hpp>
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <cstddef>
// #include <cstdint>
// #include <limits>
// #include <sfl/static_vector.hpp>
// #include <vector>
// namespace OpenLoco
type CompanyFlags int

const (
	None CompanyFlags = 0
	Unk0 CompanyFlags = (1 << 0)
	Unk1 CompanyFlags = (1 << 1)
	Unk2 CompanyFlags = (1 << 2)
	Sorted CompanyFlags = (1 << 3)
	IncreasedPerformance CompanyFlags = (1 << 4)
	DecreasedPerformance CompanyFlags = (1 << 5)
	ChallengeCompleted CompanyFlags = (1 << 6)
	ChallengeFailed CompanyFlags = (1 << 7)
	ChallengeBeatenByOpponent CompanyFlags = (1 << 8)
	Bankrupt CompanyFlags = (1 << 9)
	AutopayLoan CompanyFlags = (1 << 31)
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(CompanyFlags);
type AiPlaystyleFlags int

const (
	None AiPlaystyleFlags = 0
	Unk0 AiPlaystyleFlags = (1 << 0)
	Unk1 AiPlaystyleFlags = (1 << 1)
	Unk2 AiPlaystyleFlags = (1 << 2)
	Unk3 AiPlaystyleFlags = (1 << 3)
	NoAir AiPlaystyleFlags = (1 << 4)
	NoWater AiPlaystyleFlags = (1 << 5)
	Unk6 AiPlaystyleFlags = (1 << 6)
	Unk7 AiPlaystyleFlags = (1 << 7)
	TownIdSet AiPlaystyleFlags = (1 << 8)
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(AiPlaystyleFlags);
type CorporateRating int

const (
	Platelayer CorporateRating = iota
	Engineer
	TrafficManager
	TransportCoordinator
	RouteSupervisor
	Director
	ChiefExecutive
	Chairman
	President
	Tycoon
)
type ObservationStatus int

const (
	Empty ObservationStatus = iota
	BuildingTrackRoad
	BuildingAirport
	BuildingDock
	CheckingServices
	SurveyingLandscape
)
type Emotion int

const (
	Neutral Emotion = iota
	Happy
	Worried
	Thinking
	Dejected
	Surprised
	Scared
	Angry
	Disgusted
)
type OwnerStatus struct {
// int16_t data[2];
// OwnerStatus()
// data[0] = -1;
// data[1] = 0;
}
// OwnerStatus(EntityId id)
// data[0] = -2;
// data[1] = enumValue(id);
// OwnerStatus(const World::Pos2& pos)
// data[0] = pos.x;
// data[1] = pos.y;
// OwnerStatus(int16_t ax, int16_t cx)
// data[0] = ax;
// data[1] = cx;
// func GetData(res int16) 
// res[0] = data[0];
// res[1] = data[1];
func IsEmpty() bool {
	// bool isEntity() const { return data[0] == -2; }
	// EntityId getEntity() const
	if isEntity() {
	return EntityId(data[1])
	}
	return EntityId.null
	}
	// World::Pos2 getPosition() const
	if isEntity() {
	return {}
	}
	return World.Pos2{ data[0], data[1] }
	}
	// };
	// void formatPerformanceIndex(const int16_t performanceIndex, FormatArguments& args);
const ExpenditureHistoryCapacity int = 16
type Company struct {
}

type CompanyUnk25C0HashTableEntry struct {
	var var_00 uint16
	var var_02 uint16
	var var_04 uint8
	var var_05 uint8
	// Unk25C0HashTableEntry() = default;
	Unk25C0HashTableEntry(World.Pos3 pos, uint8_t trackRoadId, uint8_t direction)
	// constexpr World::Pos3 getPosition() const
	return World.Pos3(var_00, var_02 & 0xFFFE, var_04 * 4 /*World.kSmallZStep*/)
	}
	// constexpr uint8_t getTrackRoadId() const
	return var_05 & 0x3F
	}
	// constexpr uint8_t getDirection() const
	return (var_05 >> 6) & 0x03
	}
	// // When looking up the hash table entry if there is a hash collision check the next entries in
	// // the table as well for matches until no hash collision or match is found
	// constexpr bool hasHashCollision() const
	return (var_02 & (1 << 0)) != 0
	}
	// constexpr uint16_t calculateHash() const;
	// };
	var name StringId
	var ownerName StringId
	var challengeFlags CompanyFlags
	var cash currency48_t
	var currentLoan currency32_t
	var updateCounter uint32
	var performanceIndex int16
	var competitorId uint8
	var ownerEmotion Emotion
	var mainColours ColourScheme
	// ColourScheme vehicleColours[10];                                                // 0x1C
	var customVehicleColoursSet uint32
	var unlockedVehicles any /* BitSet<224> */ 
	var availableVehicles uint16
	var aiPlaystyleFlags AiPlaystyleFlags
	var aiPlaystyleTownId uint8
	var numExpenditureYears uint8
	// currency32_t expenditures[kExpenditureHistoryCapacity][ExpenditureType::Count]; // 0x58
	var startedDate uint32
	var var_49C uint32
	var var_4A0 uint32
	var var_4A4 AiThinkState
	var var_4A5 uint8
	var var_4A6 AiPlaceVehicleState
	var var_4A7 uint8
	// AiThought aiThoughts[kMaxAiThoughts]; // 0x04A8
	var activeThoughtId uint8
	// World::SmallZ headquartersZ;          // 0x2579
	var headquartersX coord_t
	var headquartersY coord_t
	// union
	var activeThoughtRevenueEstimate currency32_t
	var thoughtState2AiStationIdx uint32
	// };
	var var_2582 uint32
	var var_2596 uint32
	var var_259A uint8
	var var_259B uint8
	var var_259C uint8
	var aiPlaceVehicleIndex uint32
	var var_25BE AiThoughtType
	var currentRating CorporateRating
	// Unk25C0HashTableEntry var_25C0[0x1000]; // 0x25C0 Hash table entries
	var var_25C0_length uint16
	var var_85C2 uint8
	var var_85C3 uint8
	// World::Pos2 var_85C4;
	// World::SmallZ var_85C8;
	// World::Pos2 var_85C9;
	// World::SmallZ var_85CD;
	var var_85CE uint8
	var var_85CF uint8
	// World::Pos2 var_85D0;
	// World::SmallZ var_85D4;
	var var_85D5 uint16
	// World::Pos2 var_85D7;
	// World::SmallZ var_85DB;
	var var_85DC uint16
	var var_85DE uint32
	var var_85E2 uint32
	var var_85E6 uint16
	var var_85E8 uint16
	var var_85EA uint32
	var var_85EE uint8
	var var_85EF uint8
	var var_85F0 uint16
	var var_85F2 currency32_t
	var var_85F6 uint16
	var cargoUnitsTotalDelivered uint32
	// uint32_t cargoUnitsDeliveredHistory[120]; // 0x85FC
	// int16_t performanceIndexHistory[120];     // 0x87DC
	var historySize uint16
	// currency48_t companyValueHistory[120];    // 0x88CE
	var vehicleProfit currency48_t
	// uint16_t transportTypeCount[6];           // 0x8BA4
	// uint8_t activeEmotions[9];                // 0x8BB0 duration in days that emotion is active 0 == not active
	var observationStatus ObservationStatus
	var observationTownId TownId
	var observationEntity EntityId
	var observationX int16
	var observationY int16
	var observationObject uint16
	var observationTimeout uint16
	var ownerStatus OwnerStatus
	// uint32_t cargoDelivered[32];              // 0x8BCE;
	var challengeProgress uint8
	var numMonthsInTheRed uint8
	var cargoUnitsTotalDistance uint32
	// uint32_t cargoUnitsDistanceHistory[120];  // 0x8C54
	var jailStatus uint16
	// CompanyId id() const;
	// bool empty() const;
	// bool isVehicleIndexUnlocked(const uint8_t vehicleIndex) const;
	// void recalculateTransportCounts();
	// void clearOwnerStatusForDeletedVehicle(EntityId vehicleId);
	// void updateDaily();
	// void updateDailyLogic();
	// void updateDailyPlayer();
	// void evaluateChallengeProgress();
	// void updateDailyControllingPlayer();
	// void updateMonthlyHeadquarters();
	// void updateMonthly1();
	// void updateLoanAutorepay();
	// void updateQuarterly();
	// void updateVehicleColours();
	// void updateHeadquartersColour();
	// void updateOwnerEmotion();
	// uint8_t getHeadquarterPerformanceVariation() const;
	// bool hashTableContains(const Unk25C0HashTableEntry& entry) const;
	// bool addHashTableEntry(const Unk25C0HashTableEntry& entry);
	// private:
	// void setHeadquartersVariation(const uint8_t variation);
	// void setHeadquartersVariation(const uint8_t variation, const World::TilePos2& pos);
	// uint8_t getNewChallengeProgress() const;
	// };
	// StringId getCorporateRatingAsStringId(CorporateRating rating);
	// constexpr CorporateRating performanceToRating(int16_t performanceIndex);
	// void formatPerformanceIndex(const int16_t performanceIndex, FormatArguments& args);
	// void companyEmotionEvent(CompanyId companyId, Emotion emotion);
	// void companySetObservation(CompanyId id, ObservationStatus status, World::Pos2 pos, EntityId entity, uint16_t object);
	// // This is kMaxRoadObjects + kMaxTrackObjects as tram tracks are roads but are tracks
	// // and vice versa there was capabilities for some unknown track type to be classed as a road
type AvailableTracksAndRoads = any /* sfl::static_vector<uint8_t, Limits::kMaxRoadObjects + Limits::kMaxTrackObjects> */ 
	// AvailableTracksAndRoads companyGetAvailableRailTracks(const CompanyId id);
	// AvailableTracksAndRoads companyGetAvailableRoads(const CompanyId id);
	// void updateYearly(Company& company);
}

type CompanyProfitAndValue struct {
	var vehicleProfit currency48_t
	var companyValue currency48_t
	// };
	// // 0x00437D79
	// ProfitAndValue calculateCompanyValue(const Company& company);
	}
