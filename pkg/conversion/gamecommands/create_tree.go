package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type TreePlacementArgs struct {
// TreePlacementArgs() = default;
	// method: explicit TreePlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx)
// , rotation(regs.di & 0x3)
// , type(regs.bh)
// , quadrant(regs.dl)
// , colour(static_cast<Colour>(regs.dh & 0x1F))
// , buildImmediately(regs.di & 0x8000)
// , requiresFullClearance(regs.di & 0x4000)
}
const TreePlacementArgsCommand any = GameCommand.createTree
// World::Pos2 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t type;
// orphan member: uint8_t quadrant;
// orphan member: Colour colour;
// bool buildImmediately = false;
// bool requiresFullClearance = false;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.dl = quadrant;
// regs.dh = enumValue(colour);
// regs.di = rotation | (buildImmediately ? 0x8000 : 0) | (requiresFullClearance ? 0x4000 : 0);
// regs.bh = type;
// orphan member: return regs;
// func CreateTree(regs registers) 
