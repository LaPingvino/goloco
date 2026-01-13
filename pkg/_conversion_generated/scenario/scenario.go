package scenario

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Core/FileSystem.hpp>
// #include <OpenLoco/Engine/World.hpp>
// #include <cstdint>
// namespace OpenLoco
type MonthId int

const (
// forward: class FormatArguments;
// forward: struct ClimateObject;
)
// namespace OpenLoco::ObjectManager
type SelectedObjectsFlags int

const (
)
// namespace OpenLoco::Scenario
// forward: struct Objective;
// forward: struct ObjectiveProgress;
type Season int

const (
	Autumn Season = 0
	Winter Season = 1
	Spring Season = 2
	Summer Season = 3
)
// // NB: kMinYear has been changed to 1800 in OpenLoco; Locomotion uses 1900.
const MinYear uint16 = 1800
const MaxYear uint16 = 2100
const MinObjectiveYearLimit uint8 = 2
const MaxObjectiveYearLimit uint8 = 100
const MinObjectiveCompanyValue uint32 = 100000
const MaxObjectiveCompanyValue uint32 = 200000000
const MinObjectiveMonthlyProfitFromVehicles uint32 = 1000
const MaxObjectiveMonthlyProfitFromVehicles uint32 = 1000000
const MinObjectivePerformanceIndex uint32 = 10
const MaxObjectivePerformanceIndex uint32 = 100
const MinObjectiveDeliveredCargo uint32 = 100
const MaxObjectiveDeliveredCargo uint32 = 200000000
const MinCompetingCompanies uint8 = 0
const MaxCompetingCompanies uint8 = 14
const MinCompetitorStartDelay uint8 = 0
const MaxCompetitorStartDelay uint8 = 240
const MinStartLoanUnits uint16 = 50
const MaxStartLoanUnits uint16 = 1250
const MinLoanSizeUnits uint16 = 50
const MaxLoanSizeUnits uint16 = 5000
const MinLoanInterestUnits uint16 = 5
const MaxLoanInterestUnits uint16 = 25
const MinSeaLevel uint8 = 0
const MaxSeaLevel uint8 = 28
const MinNumRiverbeds uint8 = 0
const MaxNumRiverbeds uint8 = 4
const MinMinRiverWidth uint8 = 2
const MaxMinRiverWidth uint8 = 20
const MinMaxRiverWidth uint8 = 2
const MaxMaxRiverWidth uint8 = 30
const MinRiverbankWidth uint8 = 0
const MaxRiverbankWidth uint8 = 10
const MinRiverMeanderRate uint8 = 0
const MaxRiverMeanderRate uint8 = 20
const MinBaseLandHeight uint8 = 0
const MaxBaseLandHeight uint8 = 15
const MinHillDensity uint8 = 0
const MaxHillDensity uint8 = 100
const MinNumForests uint16 = 0
const MaxNumForests uint16 = 990
const MinForestRadius uint8 = 4
const MaxForestRadius uint8 = 40
const MinForestDensity uint8 = 1
const MaxForestDensity uint8 = 7
const MinNumTrees uint16 = 0
const MaxNumTrees uint16 = 20000
const MinAltitudeTrees uint8 = 0
const MaxAltitudeTrees uint8 = 40
// func NextSeason(season Season) Season
// func InitialiseSnowLine() 
// func UpdateSnowLine(currentDayOfYear int32) 
// // 0x00525FB4
// World::SmallZ getCurrentSnowLine();
// func SetCurrentSnowLine(snowline World::SmallZ) 
// // 0x00525FB5
// func GetCurrentSeason() Season
// func SetCurrentSeason(season Season) 
// func UpdateSeason(currentDayOfYear int32, climateObj ClimateObject) 
// func Reset() 
// func Sub_4748D4() 
// func Sub_4969E0(al uint8) 
// func EraseLandscape() 
// func GenerateLandscape() 
// func InitialiseDate(year uint16) 
// func InitialiseDate(year uint16, month OpenLoco::MonthId, day uint8) 
// /**
// * Loads the given scenario file, but does not initialise any game state.
// */
// func Load(path fs::path) bool
// /**
// * Loads the given scenario file and resets the game state for starting a new scenario.
// */
// func LoadAndStart(path fs::path) bool
// /**
// * Resets the game state (e.g. companies, year, money etc.) for starting a new scenario.
// */
// [[noreturn]] void start();
// func FormatChallengeArguments(objective Objective, progress ObjectiveProgress, args FormatArguments) 
// func Sub_46115C() 
// func LoadPreferredCurrencyAlways() 
// func LoadPreferredCurrencyNewGame() 
// func DrawScenarioMiniMapImage() 
