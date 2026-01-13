package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Objects/Object.h"
// namespace OpenLoco::Scenario
// forward: struct Options;
// namespace OpenLoco::S5
type Objective struct {
	Type uint8
	Flags uint8
	CompanyValue uint32
	MonthlyVehicleProfit uint32
	PerformanceIndex uint8
	DeliveredCargoType uint8
	DeliveredCargoAmount uint32
	TimeLimitYears uint8
}
// static_assert(sizeof(Objective) == 0x11);
type Options struct {
	EditorStep uint8
	Difficulty uint8
	ScenarioStartYear uint16
// uint8_t pad_4[2];                     // 0x04
	ScenarioFlags uint16
	MadeAnyChanges uint8
// uint8_t pad_9[1];                     // 0x09
// uint8_t landDistributionPatterns[32]; // 0x0A
// char scenarioName[64];                // 0x2A
// char scenarioDetails[256];            // 0x6A
	ScenarioText ObjectHeader
	NumberOfForests uint16
	MinForestRadius uint8
	MaxForestRadius uint8
	MinForestDensity uint8
	MaxForestDensity uint8
	NumberRandomTrees uint16
	MinAltitudeForTrees uint8
	MaxAltitudeForTrees uint8
	MinLandHeight uint8
	TopographyStyle uint8
	HillDensity uint8
	NumberOfTowns uint8
	MaxTownSize uint8
	NumberOfIndustries uint8
// uint8_t preview[128][128];            // 0x18A
	MaxCompetingCompanies uint8
	CompetitorStartDelay uint8
	Objective Objective
	ObjectiveDeliveredCargo ObjectHeader
	Currency ObjectHeader
// // new fields:
	Generator uint8
	NumTerrainSmoothingPasses uint8
	NumRiverbeds uint8
	MinRiverWidth uint8
	MaxRiverWidth uint8
	RiverbankWidth uint8
	RiverMeanderRate uint8
// std::byte pad_41BD[342];
}
// static_assert(sizeof(Options) == 0x431A);
// OpenLoco::Scenario::Options importOptions(const S5::Options& src);
// S5::Options exportOptions(const OpenLoco::Scenario::Options& src);
