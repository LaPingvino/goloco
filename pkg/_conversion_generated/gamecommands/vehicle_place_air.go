package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleAirPlacementArgs struct {
// VehicleAirPlacementArgs() = default;
	// method: explicit VehicleAirPlacementArgs(const registers& regs)
// : stationId(StationId(regs.bp))
// , airportNode(regs.dl)
// , head(EntityId(regs.di))
// , convertGhost((regs.ebx & 0xFFFF0000) == 0xFFFF0000)
}
const VehicleAirPlacementArgsCommand any = GameCommand.vehiclePlaceAir
// orphan member: StationId stationId;
// orphan member: uint8_t airportNode;
// orphan member: EntityId head;
// bool convertGhost = false;
// explicit operator registers() const
// orphan member: registers regs;
// regs.bp = enumValue(stationId);
// regs.di = enumValue(head);
// regs.dl = airportNode;
// regs.ebx = convertGhost ? 0xFFFF0000 : 0;
// orphan member: return regs;
// func VehiclePlaceAir(regs registers) 
