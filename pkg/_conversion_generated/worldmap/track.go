package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Types.hpp"
// #include <OpenLoco/Engine/World.hpp>
// #include <sfl/static_vector.hpp>
// #include <utility>
// namespace OpenLoco::World::Track
// namespace AdditionalTaDFlags
const BasicTaDMask uint16 = 0b0000'0001'1111'1111
const BasicRaDMask uint16 = 0b0000'0000'0111'1111
const BridgeMask uint16 = 0b0000'1110'0000'0000
const IsOvertaking uint16 = (1 << 7)
const IsChangingLane uint16 = (1 << 8)
const HasBridge uint16 = (1 << 12)
const HasMods uint16 = (1 << 13)
const HasSignal uint16 = (1 << 15)
const BasicTaDWithSignalMask uint16 = basicTaDMask | hasSignal
const BasicRaDWithSignalMask uint16 = basicRaDMask | hasSignal
type RoadConnections struct {
// sfl::static_vector<uint16_t, 16> connections;
// StationId stationId = StationId::null; // 0x01135FAE
// uint8_t stationObjectId = 0U;          // 0x01136087
// uint8_t roadObjectId = 0xFFU;          // 0x0112C2ED (I wouldn't trust this to be correct which connection!)
}
// // requiredMods : 0x0113601A
// // queryMods : 0x0113601B
// //
// // if set requiredMods must exist on connections to be added to connection list
// //  e.g. if required mods is `0b10` then tracks with mods `0b10` and `0b11` can connect
// //  but `0b00` and `0b01` cannot
// // if requiredMods == 0 then mods ignored
// //
// // queryMods sets AdditionalTaDFlags::hasMods of connection if connection has the queryMods
// func GetRoadConnections(nextTrackPos World::Pos3, nextRotation uint8, company CompanyId, roadObjectId uint8, requiredMods uint8, queryMods uint8) RoadConnections
// func GetRoadConnectionsOneWay(nextTrackPos World::Pos3, nextRotation uint8, company CompanyId, roadObjectId uint8, requiredMods uint8, queryMods uint8) RoadConnections
// func GetRoadConnectionsAiAllocated(nextTrackPos World::Pos3, nextRotation uint8, company CompanyId, roadObjectId uint8, requiredMods uint8, queryMods uint8) RoadConnections
type ConnectionEnd struct {
// World::Pos3 nextPos;
	NextRotation uint8
}
// func GetRoadConnectionEnd(pos World::Pos3, trackAndDirection uint16) ConnectionEnd
type TrackConnections struct {
// sfl::static_vector<uint16_t, 16> connections;
// bool hasLevelCrossing = false;         // 0x0113607D
// StationId stationId = StationId::null; // 0x01135FAE
}
// // requiredMods : 0x0113601A
// // queryMods : 0x0113601B
// //
// // if set requiredMods must exist on connections to be added to connection list
// //  e.g. if required mods is `0b10` then tracks with mods `0b10` and `0b11` can connect
// //  but `0b00` and `0b01` cannot
// // if requiredMods == 0 then mods ignored
// //
// // queryMods sets AdditionalTaDFlags::hasMods of connection if connection has the queryMods
// func GetTrackConnections(nextTrackPos World::Pos3, nextRotation uint8, company CompanyId, trackObjectId uint8, requiredMods uint8, queryMods uint8) TrackConnections
// func GetTrackConnectionsAi(nextTrackPos World::Pos3, nextRotation uint8, company CompanyId, trackObjectId uint8, requiredMods uint8, queryMods uint8) TrackConnections
// func GetTrackConnectionEnd(pos World::Pos3, trackAndDirection uint16) ConnectionEnd
