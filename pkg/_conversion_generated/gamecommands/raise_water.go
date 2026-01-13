package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type RaiseWaterArgs struct {
// RaiseWaterArgs() = default;
	// method: explicit RaiseWaterArgs(const registers& regs)
// : pointA(regs.ax, regs.cx)
// , pointB(regs.di, regs.bp)
}
const RaiseWaterArgsCommand any = GameCommand.raiseWater
// World::Pos2 pointA;
// World::Pos2 pointB;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pointA.x;
// regs.cx = pointA.y;
// regs.di = pointB.x;
// regs.bp = pointB.y;
// orphan member: return regs;
// func RaiseWater(regs registers) 
