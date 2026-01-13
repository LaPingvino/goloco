package economy

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Types.hpp"
// namespace OpenLoco
type ExpenditureType int

const (
	TrainIncome ExpenditureType = iota
	TrainRunningCosts
	BusIncome
	BusRunningCosts
	TruckIncome
	TruckRunningCosts
	TramIncome
	TramRunningCosts
	AircraftIncome
	AircraftRunningCosts
	ShipIncome
	ShipRunningCosts
	Construction
	VehiclePurchases
	VehicleDisposals
	LoanInterest
	Miscellaneous
	Count
)
