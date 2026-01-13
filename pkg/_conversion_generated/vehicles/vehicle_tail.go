package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Vehicle.h"
// namespace OpenLoco::Vehicles
type VehicleTail struct {
	VehicleBase // embedded
	Sound VehicleSound
	TrainDanglingTimeout uint16
	// method: bool update();
}
const VehicleTailVehicleThingType any = VehicleEntityType.tail
// static_assert(sizeof(VehicleTail) <= sizeof(Entity));
