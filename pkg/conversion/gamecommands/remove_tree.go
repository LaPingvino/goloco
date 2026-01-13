package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type TreeRemovalArgs struct {
// TreeRemovalArgs() = default;
	// method: explicit TreeRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.dl * World::kSmallZStep)
// , type(regs.dh)
// , elementType(regs.bh)
}
const TreeRemovalArgsCommand any = GameCommand.removeTree
// World::Pos3 pos;
// orphan member: uint8_t type;
// orphan member: uint8_t elementType;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.dl = pos.z / World::kSmallZStep;
// regs.dh = type;
// regs.bh = elementType;
// orphan member: return regs;
// func RemoveTree(regs registers) 
