package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type TrackPlacementArgs struct {
// TrackPlacementArgs() = default;
	// method: explicit TrackPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , trackId(regs.dh & 0x3F)
// , mods((regs.edi >> 16) & 0xFF)
// , unkFlags((regs.edx >> 20) & 0xF)
// , bridge((regs.edx >> 24) & 0xFF)
// , trackObjectId(regs.dl)
// , unk(regs.edi & 0x800000)
}
const TrackPlacementArgsCommand any = GameCommand.createTrack
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t trackId;
// orphan member: uint8_t mods;
// orphan member: uint8_t unkFlags;
// orphan member: uint8_t bridge;
// orphan member: uint8_t trackObjectId;
// orphan member: bool unk;
// explicit operator registers() const
// orphan member: registers regs;
// regs.eax = pos.x;
// regs.cx = pos.y;
// regs.edi = (0xFFFF & pos.z) | (mods << 16) | (unk ? 0x800000 : 0);
// regs.bh = rotation;
// regs.edx = trackObjectId | (trackId << 8) | (unkFlags << 20) | (bridge << 24);
// orphan member: return regs;
// func CreateTrack(regs registers) 
