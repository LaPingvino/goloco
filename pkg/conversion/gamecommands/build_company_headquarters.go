package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type HeadquarterPlacementArgs struct {
// HeadquarterPlacementArgs() = default;
	// method: explicit HeadquarterPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , type(regs.dl)
// , buildImmediately(regs.bh & 0x80)
}
const HeadquarterPlacementArgsCommand any = GameCommand.buildCompanyHeadquarters
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t type;
// bool buildImmediately = false; // No scaffolding required (editor mode)
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// regs.dx = type;
// regs.bh = rotation | (buildImmediately ? 0x80 : 0);
// orphan member: return regs;
// func BuildCompanyHeadquarters(regs registers) 
