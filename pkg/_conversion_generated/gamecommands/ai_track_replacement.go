package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type AiTrackReplacementArgs struct {
// AiTrackReplacementArgs() = default;
	// method: explicit AiTrackReplacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , trackId(regs.dl)
// , sequenceIndex(regs.dh)
// , trackObjectId(regs.bp)
}
const AiTrackReplacementArgsCommand any = GameCommand.aiTrackReplacement
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t trackId;
// orphan member: uint8_t sequenceIndex;
// orphan member: uint8_t trackObjectId;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// regs.bh = rotation;
// regs.dl = trackId;
// regs.dh = sequenceIndex;
// regs.bp = trackObjectId;
// orphan member: return regs;
// func AiTrackReplacement(regs registers) 
