package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehiclePlacementArgs struct {
// VehiclePlacementArgs() = default;
	// method: explicit VehiclePlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.dx * World::kSmallZStep)
// , trackAndDirection(regs.bp)
// , trackProgress(regs.ebx >> 16)
// , head(EntityId(regs.di))
// , convertGhost((regs.ebx & 0xFFFF0000) == 0xFFFF0000)
}
const VehiclePlacementArgsCommand any = GameCommand.vehiclePlace
// World::Pos3 pos;
// orphan member: uint16_t trackAndDirection;
// orphan member: uint16_t trackProgress;
// orphan member: EntityId head;
// bool convertGhost = false;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ebp = trackAndDirection;
// regs.di = enumValue(head);
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.dx = pos.z / World::kSmallZStep;
// regs.ebx = convertGhost ? 0xFFFF0000 : (trackProgress << 16);
// orphan member: return regs;
// func VehiclePlace(regs registers) 
