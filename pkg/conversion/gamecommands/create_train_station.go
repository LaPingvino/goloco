package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type TrainStationPlacementArgs struct {
// TrainStationPlacementArgs() = default;
	// method: explicit TrainStationPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , trackId(regs.dl & 0x3F)
// , index(regs.dh & 0x3)
// , trackObjectId(regs.bp)
// , type(regs.edi >> 16)
}
const TrainStationPlacementArgsCommand any = GameCommand.createTrainStation
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t trackId;
// orphan member: uint8_t index;
// orphan member: uint8_t trackObjectId;
// orphan member: uint8_t type;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.edi = pos.z | (type << 16);
// regs.bh = rotation;
// regs.dl = trackId;
// regs.dh = index;
// regs.bp = trackObjectId;
// orphan member: return regs;
// func CreateTrainStation(regs registers) 
