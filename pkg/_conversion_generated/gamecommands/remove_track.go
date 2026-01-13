package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type TrackRemovalArgs struct {
// TrackRemovalArgs() = default;
	// method: explicit TrackRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , trackId(regs.dl & 0x3F)
// , index(regs.dh)
// , trackObjectId(regs.ebp)
}
const TrackRemovalArgsCommand any = GameCommand.removeTrack
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t trackId;
// orphan member: uint8_t index;
// orphan member: uint8_t trackObjectId;
// explicit operator registers() const
// orphan member: registers regs;
// regs.eax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// regs.bh = rotation;
// regs.dl = trackId;
// regs.dh = index;
// regs.ebp = trackObjectId;
// orphan member: return regs;
// func RemoveTrack(regs registers) 
