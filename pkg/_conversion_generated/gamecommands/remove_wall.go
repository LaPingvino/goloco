package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type WallRemovalArgs struct {
// WallRemovalArgs() = default;
	// method: explicit WallRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.dh * World::kSmallZStep)
// , rotation(regs.dl)
}
const WallRemovalArgsCommand any = GameCommand.removeWall
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.dh = pos.z / World::kSmallZStep;
// regs.dl = rotation;
// orphan member: return regs;
// func RemoveWall(regs registers) 
