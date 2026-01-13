package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type SignalPlacementArgs struct {
// SignalPlacementArgs() = default;
	// method: explicit SignalPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , trackId(regs.dl & 0x3F)
// , index(regs.dh)
// , type((regs.edi >> 16) & 0xFF)
// , trackObjType(regs.ebp & 0xFF)
// , sides((regs.edi >> 16) & 0xC000)
}
const SignalPlacementArgsCommand any = GameCommand.createSignal
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t trackId;
// orphan member: uint8_t index;
// orphan member: uint8_t type;
// orphan member: uint8_t trackObjType;
// orphan member: uint16_t sides;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.bh = rotation;
// regs.dl = trackId;
// regs.dh = index;
// regs.edi = pos.z | (type << 16) | ((sides & 0xC000) << 16);
// regs.ebp = trackObjType;
// orphan member: return regs;
// func CreateSignal(regs registers) 
