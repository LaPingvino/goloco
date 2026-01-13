package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type SignalRemovalArgs struct {
	// SignalRemovalArgs() = default;
	// method: explicit SignalRemovalArgs(const registers& regs)
	// : pos(regs.ax, regs.cx, regs.di)
	// , rotation(regs.bh & 0x3)
	// , trackId(regs.dl & 0x3F)
	// , index(regs.dh & 0xF)
	// , trackObjType(regs.bp & 0xF)
	// , flags(regs.edi >> 16)
}

// MALFORMED FIELD: const SignalRemovalArgsCommand any = GameCommand.removeSignal

// World::Pos3 pos;
// orphan member: uint8 rotation;
// orphan member: uint8 trackId;
// orphan member: uint8 index;
// orphan member: uint8 trackObjType;
// orphan member: uint16 flags;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.edi = pos.z | (flags << 16);
// regs.bh = rotation;
// regs.dl = trackId;
// regs.dh = index;
// regs.bp = trackObjType;
// orphan member: return regs;
// func RemoveSignal(regs registers)