package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type HeadquarterRemovalArgs struct {
// HeadquarterRemovalArgs() = default;
	// method: explicit HeadquarterRemovalArgs(const World::Pos3& place)
// : pos(place)
}
const HeadquarterRemovalArgsCommand any = GameCommand.removeCompanyHeadquarters
// func HeadquarterRemovalArgs(regs registers) explicit
// : pos(regs.ax, regs.cx, regs.di)
// World::Pos3 pos;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// orphan member: return regs;
// func RemoveCompanyHeadquarters(regs registers) 
