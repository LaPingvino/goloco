package scenario

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <cstdint>
// namespace OpenLoco::Scenario
type ObjectiveFlags int

const (
func init() {
	None                      ObjectiveFlags = 0
// SKIPPED CONSTRUCTOR: 	BeTopCompany              ObjectiveFlags = (1 << 0)
// SKIPPED CONSTRUCTOR: 	BeWithinTopThreeCompanies ObjectiveFlags = (1 << 1)
// SKIPPED CONSTRUCTOR: 	WithinTimeLimit           ObjectiveFlags = (1 << 2)
// SKIPPED CONSTRUCTOR: 	Flag_3                    ObjectiveFlags = (1 << 3)
// SKIPPED CONSTRUCTOR: 	Flag_4                    ObjectiveFlags = (1 << 4)
}

)

// OPENLOCO_ENABLE_ENUM_OPERATORS(ObjectiveFlags);
type ObjectiveType int

func init() {
	CompanyValue ObjectiveType = iota

	VehicleProfit
	PerformanceIndex
	CargoDelivery

type Objective struct {
	Type                 ObjectiveType
	Flags                ObjectiveFlags
	CompanyValue         uint32
	MonthlyVehicleProfit uint32
	PerformanceIndex     uint8
	DeliveredCargoType   uint8
	DeliveredCargoAmount uint32
	TimeLimitYears       uint8

// Objective& getObjective();
type ObjectiveProgress struct {
	TimeLimitUntilYear         uint16
	MonthsInChallenge          uint16
	CompletedChallengeInMonths uint16

// ObjectiveProgress& getObjectiveProgress();