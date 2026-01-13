package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type AirportRemovalArgs struct {
// AirportRemovalArgs() = default;
	// method: explicit AirportRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
}
const AirportRemovalArgsCommand any = GameCommand.removeAirport
// // Note: pos.z must be a floored BigZ
// World::Pos3 pos;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// orphan member: return regs;
// func RemoveAirport(regs registers) 
