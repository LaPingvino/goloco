package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type RoadRemovalArgs struct {
// RoadRemovalArgs() = default;
	// method: explicit RoadRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , roadId(regs.dl & 0xF)
// , sequenceIndex(regs.dh & 0x3)
// , objectId(regs.bp & 0xF)
}
const RoadRemovalArgsCommand any = GameCommand.removeRoad
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t roadId;
// orphan member: uint8_t sequenceIndex;
// orphan member: uint8_t objectId;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// regs.bh = rotation;
// regs.dl = roadId;
// regs.dh = sequenceIndex;
// regs.bp = objectId;
// orphan member: return regs;
// func RemoveRoad(regs registers) 
