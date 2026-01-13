package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type ClearLandArgs struct {
// ClearLandArgs() = default;
	// method: explicit ClearLandArgs(const registers& regs)
// : centre(regs.ax, regs.cx)
// , pointA(regs.edx & 0xFFFF, regs.ebp & 0xFFFF)
// , pointB(regs.edx >> 16, regs.ebp >> 16)
}
const ClearLandArgsCommand any = GameCommand.clearLand
// World::Pos2 centre;
// World::Pos2 pointA;
// World::Pos2 pointB;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = centre.x;
// regs.cx = centre.y;
// regs.edx = (pointB.x << 16) | pointA.x;
// regs.ebp = (pointB.y << 16) | pointA.y;
// orphan member: return regs;
// func ClearLand(regs registers) 
