package world

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Engine/Limits.h"
// #include "Station.h"
// #include <OpenLoco/Core/LocoFixedVector.hpp>
// #include <array>
// #include <cstddef>
// #include <span>
// namespace OpenLoco::StationManager
// // If it exceeds this distance, it will not be considered a nearby station
const MaxStationNearbyDistance int16 = 64
// func Reset() 
// func Stations() any /* FixedVector<Station, Limits::kMaxStations> */ 
// Station* get(StationId id);
// func Update() 
// func UpdateLabels() 
// func UpdateDaily() 
// func GenerateNewStationName(stationId StationId, townId TownId, position World::Pos3, mode uint8) StringId
// func ZeroUnused() 
// func DeliverCargoToNearbyStations(cargoType uint8, cargoQty uint8, pos World::Pos2, size World::TilePos2) uint16
// func DeliverCargoToStations(stations any /* std::span<StationId> */ , cargoType uint8, cargoQty uint8) uint16
// func ExceedsStationSize(station Station, pos World::Pos3) bool
// func AllocateNewStation(pos World::Pos3, owner CompanyId, mode uint8) StationId
// func DeallocateStation(stationId StationId) 
type NearbyStation struct {
	Id StationId
	IsPhysicallyAttached bool
}
// func FindNearbyStation(pos World::Pos3, companyId CompanyId) NearbyStation
// // Subfunction of findNearbyStation (For create airport)
// func FindNearbyEmptyStation(pos World::Pos3, companyId CompanyId, currentMinDistanceStation int16) StationId
