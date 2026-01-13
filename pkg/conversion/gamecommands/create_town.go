package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type TownPlacementArgs struct {
// TownPlacementArgs() = default;
	// method: explicit TownPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx)
// , size(regs.dl)
}
const TownPlacementArgsCommand any = GameCommand.createTown
// World::Pos2 pos;
// orphan member: uint8_t size;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.edx = size;
// orphan member: return regs;
// func CreateTown(regs registers) 
