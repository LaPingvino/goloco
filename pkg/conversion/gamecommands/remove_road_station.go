package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type RoadStationRemovalArgs struct {
// RoadStationRemovalArgs() = default;
	// method: explicit RoadStationRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , roadId(regs.dl & 0xF)
// , index(regs.dh & 0x3)
// , roadObjectId(regs.bp & 0xF)
}
const RoadStationRemovalArgsCommand any = GameCommand.removeRoadStation
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t roadId;
// orphan member: uint8_t index;
// orphan member: uint8_t roadObjectId;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// regs.bh = rotation;
// regs.dl = roadId;
// regs.dh = index;
// regs.bp = roadObjectId;
// orphan member: return regs;
// func RemoveRoadStation(regs registers) 
