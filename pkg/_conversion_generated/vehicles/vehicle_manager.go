package vehicles

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Entities/EntityManager.h"
// #include "Types.hpp"
// namespace OpenLoco
// forward: struct Company;
// namespace OpenLoco::Vehicles
// forward: struct VehicleHead;
// forward: struct Car;
// forward: struct TrackAndDirection;
// namespace OpenLoco::VehicleManager
type VehicleList = any /* EntityManager::EntityList<EntityManager::EntityListIterator<Vehicles::VehicleHead>, EntityManager::EntityListType::vehicleHead> */ 
// func Update() 
// func UpdateMonthly() 
// func UpdateDaily() 
// func DetermineAvailableVehicles(company Company) 
// func DetermineAvailableVehicleTypes(company Company) uint16
// func ResetIfHeadingForStation(stationId StationId) 
// func DeleteTrain(head Vehicles::VehicleHead) 
// func DeleteCar(car Vehicles::Car) 
type PlaceDownResult int

const (
	Okay PlaceDownResult = iota
	Unk0
	Unk1
)
// func PlaceDownVehicle(head Vehicles::VehicleHead, x coord_t, y coord_t, baseZ uint8, trackAndDirection Vehicles::TrackAndDirection, initialSubPosition uint16) PlaceDownResult
