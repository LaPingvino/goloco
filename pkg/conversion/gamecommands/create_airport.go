package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type AirportPlacementArgs struct {
// AirportPlacementArgs() = default;
	// method: explicit AirportPlacementArgs(const registers regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh)
// , type(regs.dl)
}
const AirportPlacementArgsCommand any = GameCommand.createAirport
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t type;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// regs.bh = rotation;
// regs.dl = type;
// orphan member: return regs;
// func CreateAirport(regs registers) 
