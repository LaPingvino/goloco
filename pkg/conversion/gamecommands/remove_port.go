package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type PortRemovalArgs struct {
// PortRemovalArgs() = default;
	// method: explicit PortRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
}
const PortRemovalArgsCommand any = GameCommand.removePort
// World::Pos3 pos;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// orphan member: return regs;
// func RemovePort(regs registers) 
