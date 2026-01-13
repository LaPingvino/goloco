package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Vehicle.h"
// namespace OpenLoco
type AirportObjectFlags int

const (
)
// namespace OpenLoco::Vehicles
type VehicleBogie struct {
	VehicleBase // embedded
	ColourScheme ColourScheme
	ObjectSpriteType uint8
	ObjectId uint16
	Var_44 uint16
	AnimationIndex uint8
	Var_47 uint8
	SecondaryCargo VehicleCargo
	TotalCarWeight uint16
	BodyIndex uint8
	CreationDay uint32
	Var_5A uint32
	WheelSlipping uint8
	BreakdownFlags BreakdownFlags
	Var_60 uint8
	Var_61 uint8
	RefundCost uint32
	Reliability uint16
	TimeoutToBreakdown uint16
	BreakdownTimeout uint8
	// method: AirportObjectFlags getCompatibleAirportType();
	// method: bool update();
	// method: void updateSegmentCrashed();
	// method: bool isOnRackRail();
	// method: constexpr bool hasBreakdownFlags(BreakdownFlags flagsToTest) const
// return (breakdownFlags & flagsToTest) != BreakdownFlags::none;
}
const VehicleBogieVehicleThingType any = VehicleEntityType.bogie
// private:
// func UpdateRoll(unkDistance int32) 
// func Collision(collideEntityId EntityId) 
// static_assert(sizeof(VehicleBogie) <= sizeof(Entity));
