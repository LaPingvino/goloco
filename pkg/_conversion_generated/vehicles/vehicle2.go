package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Vehicle.h"
// namespace OpenLoco::Vehicles
type MotorState int

const (
	Stopped MotorState = 0
	Accelerating MotorState = 1
	Coasting MotorState = 2
	Braking MotorState = 3
	StoppedOnIncline MotorState = 4
	AirplaneAtTaxiSpeed MotorState = 5
)
type Flags73 int

const (
	None Flags73 = 0
	IsBrokenDown Flags73 = 1 << 0
	IsStillPowered Flags73 = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(Flags73);
type Vehicle2 struct {
	VehicleBase // embedded
	Sound VehicleSound
	Var_4F int8
	TotalPower uint16
	TotalWeight uint16
	MaxSpeed Speed16
	CurrentSpeed Speed32
	MotorState MotorState
	BrakeLightTimeout uint8
	RackRailMaxSpeed Speed16
	CurMonthRevenue currency32_t
// currency32_t profit[4];       // last 4 months net profit
	Reliability uint8
	Var_73 Flags73
	// method: bool has73Flags(Flags73 flagsToTest) const;
	// method: bool update();
	// method: bool sub_4A9F20();
	// method: currency32_t totalRecentProfit() const
// return profit[0] + profit[1] + profit[2] + profit[3];
}
const Vehicle2VehicleThingType any = VehicleEntityType.vehicle_2
// static_assert(sizeof(Vehicle2) <= sizeof(Entity));
