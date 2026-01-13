package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Vehicle.h"
// namespace OpenLoco::Vehicles
type Flags48 int

const (
	None Flags48 = 0
	PassSignal Flags48 = 1 << 0
	ExpressMode Flags48 = 1 << 1
	Flag2 Flags48 = 1 << 2 // cargo related?
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(Flags48);
type IncomeStats struct {
	Day int32
// uint8_t cargoTypes[4];
// uint16_t cargoQtys[4];
// uint16_t cargoDistances[4];
// uint8_t cargoAges[4];
// currency32_t cargoProfits[4];
	// method: void beginNewIncome();
	// method: bool addToStats(uint8_t cargoType, uint16_t cargoQty, uint16_t cargoDist, uint8_t cargoAge, currency32_t profit);
}
type Vehicle1 struct {
	VehicleBase // embedded
	Var_3C int32
	TargetSpeed Speed16
	TimeAtSignal uint16
	Var_48 Flags48
	Var_49 uint8
	DayCreated uint32
	Var_4E uint16
	Var_50 uint16
	Var_52 uint8
	LastIncome IncomeStats
	// method: bool update();
	// method: bool updateRoad();
	// method: bool updateRail();
	// method: UpdateMotionResult updateRoadMotion(int32_t distance);
}
const Vehicle1VehicleThingType any = VehicleEntityType.vehicle_1
// static_assert(sizeof(Vehicle1) <= sizeof(Entity));
