package scenario

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "EditorController.h"
// #include "Objects/Object.h"
// #include "Scenario/ScenarioObjective.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <cstddef>
// namespace OpenLoco::Scenario
type ScenarioFlags int

const (
	None ScenarioFlags = 0
	LandscapeGenerationDone ScenarioFlags = (1 << 0)
	HillsEdgeOfMap ScenarioFlags = (1 << 1)
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(ScenarioFlags);
type TopographyStyle int

const (
	FlatLand TopographyStyle = iota
	SmallHills
	Mountains
	HalfMountainsHills
	HalfMountainsFlat
)
type LandGeneratorType int

const (
	Original LandGeneratorType = iota
	Simplex
	PngHeightMap
)
type LandDistributionPattern int

const (
	Everywhere LandDistributionPattern = iota
	Nowhere
	FarFromWater
	NearWater
	OnMountains
	FarFromMountains
	InSmallRandomAreas
	InLargeRandomAreas
	AroundCliffs
)
type Options struct {
// EditorController::Step editorStep;                    // 0x00
	Difficulty uint8
	ScenarioStartYear uint16
	ScenarioFlags ScenarioFlags
	MadeAnyChanges uint8
// LandDistributionPattern landDistributionPatterns[32]; // 0x0A
// char scenarioName[64];                                // 0x2A
// char scenarioDetails[256];                            // 0x6A
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
	TopographyStyle TopographyStyle
	HillDensity uint8
	NumberOfTowns uint8
	MaxTownSize uint8
	NumberOfIndustries uint8
// uint8_t preview[128][128];                            // 0x18A
	MaxCompetingCompanies uint8
	CompetitorStartDelay uint8
// Scenario::Objective objective;                        // 0x418C
	ObjectiveDeliveredCargo ObjectHeader
	Currency ObjectHeader
// // new fields:
	Generator LandGeneratorType
	NumTerrainSmoothingPasses uint8
	NumRiverbeds uint8
	MinRiverWidth uint8
	MaxRiverWidth uint8
	RiverbankWidth uint8
	RiverMeanderRate uint8
}
// Options& getOptions();
