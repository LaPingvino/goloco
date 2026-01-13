package scenario

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Objects/Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <cstddef>
// #include <cstdint>
// namespace OpenLoco::Scenario
// forward: struct ObjectiveProgress;
// namespace OpenLoco::ScenarioManager
type ScenarioIndexFlags int

const (
	None ScenarioIndexFlags = 0
	Flag_0 ScenarioIndexFlags = 1 << 0
	Completed ScenarioIndexFlags = 1 << 1
	HasPreviewImage ScenarioIndexFlags = 1 << 2
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(ScenarioIndexFlags);
type ScenarioIndexEntry struct {
// char filename[0x100];           // 0x000
	Category uint8
	NumCompetingCompanies uint8
	CompetingCompanyDelay uint8
// uint8_t pad_103[0x120 - 0x103]; // 0x103
	StartYear uint16
	CompletedMonths uint16
// char scenarioName[0x40];        // 0x124
// char description[0x100];        // 0x164
	Flags ScenarioIndexFlags
// char highscoreName[0x100];      // 0x268
// char objective[0x100];          // 0x368
	Currency ObjectHeader
// uint8_t preview[128][128];      // 0x478
	// method: constexpr bool hasFlag(ScenarioIndexFlags flag) const
// return (flags & flag) != ScenarioIndexFlags::none;
}
// static_assert(offsetof(ScenarioIndexEntry, category) == 0x100);
// static_assert(offsetof(ScenarioIndexEntry, flags) == 0x264);
// static_assert(sizeof(ScenarioIndexEntry) == 0x4478);
// func HasScenariosForCategory(category uint8) bool
// func HasScenarioInCategory(category uint8, scenario ScenarioIndexEntry) bool
// func GetScenarioCountByCategory(category uint8) uint16
// ScenarioIndexEntry* getNthScenarioFromCategory(uint8_t category, uint8_t index);
// func LoadIndex(false bool forceReload =) 
// func SaveNewScore(progress Scenario::ObjectiveProgress, companyId CompanyId) 
// // 0x00525F5E
// func GetScenarioTicks() uint32
// func SetScenarioTicks(ticks uint32) 
// // 0x00525F64
// func GetScenarioTicks2() uint32
// func SetScenarioTicks2(ticks uint32) 
