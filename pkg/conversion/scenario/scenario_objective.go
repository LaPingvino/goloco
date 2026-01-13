package scenario

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <cstdint>
// namespace OpenLoco::Scenario
type ObjectiveFlags int

const (
	None ObjectiveFlags = 0
	BeTopCompany ObjectiveFlags = (1 << 0)
	BeWithinTopThreeCompanies ObjectiveFlags = (1 << 1)
	WithinTimeLimit ObjectiveFlags = (1 << 2)
	Flag_3 ObjectiveFlags = (1 << 3)
	Flag_4 ObjectiveFlags = (1 << 4)
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(ObjectiveFlags);
type ObjectiveType int

const (
	CompanyValue ObjectiveType = iota
	VehicleProfit
	PerformanceIndex
	CargoDelivery
)
type Objective struct {
	Type ObjectiveType
	Flags ObjectiveFlags
	CompanyValue uint32
	MonthlyVehicleProfit uint32
	PerformanceIndex uint8
	DeliveredCargoType uint8
	DeliveredCargoAmount uint32
	TimeLimitYears uint8
}
// Objective& getObjective();
type ObjectiveProgress struct {
	TimeLimitUntilYear uint16
	MonthsInChallenge uint16
	CompletedChallengeInMonths uint16
}
// ObjectiveProgress& getObjectiveProgress();
