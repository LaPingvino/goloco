package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type ChangeLandMaterialArgs struct {
// ChangeLandMaterialArgs() = default;
	// method: explicit ChangeLandMaterialArgs(const registers& regs)
// : pointA(regs.ax, regs.cx)
// , pointB(regs.di, regs.bp)
// , landType(regs.dl)
}
const ChangeLandMaterialArgsCommand any = GameCommand.changeLandMaterial
// World::Pos2 pointA;
// World::Pos2 pointB;
// orphan member: uint8_t landType;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pointA.x;
// regs.cx = pointA.y;
// regs.di = pointB.x;
// regs.bp = pointB.y;
// regs.dl = landType;
// orphan member: return regs;
// func ChangeLandMaterial(regs registers) 
