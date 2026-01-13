package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleWaterPlacementArgs struct {
// VehicleWaterPlacementArgs() = default;
	// method: explicit VehicleWaterPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.dx)
// , head(EntityId(regs.di))
// , convertGhost((regs.ebx & 0xFFFF0000) == 0xFFFF0000)
}
const VehicleWaterPlacementArgsCommand any = GameCommand.vehiclePlaceWater
// World::Pos3 pos;
// orphan member: EntityId head;
// bool convertGhost = false;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.dx = pos.z;
// regs.di = enumValue(head);
// regs.ebx = convertGhost ? 0xFFFF0000 : 0;
// orphan member: return regs;
// func VehiclePlaceWater(regs registers) 
