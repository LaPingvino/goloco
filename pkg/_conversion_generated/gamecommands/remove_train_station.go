package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type TrainStationRemovalArgs struct {
// TrainStationRemovalArgs() = default;
	// method: explicit TrainStationRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , trackId(regs.dl & 0x3F)
// , index(regs.dh & 0xF)
// , type(regs.bp & 0xF)
}
const TrainStationRemovalArgsCommand any = GameCommand.removeTrainStation
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t trackId;
// orphan member: uint8_t index;
// orphan member: uint8_t type;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// regs.bh = rotation;
// regs.dl = trackId;
// regs.dh = index;
// regs.bp = type;
// orphan member: return regs;
// func RemoveTrainStation(regs registers) 
