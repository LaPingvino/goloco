package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type AiRoadAndStationPlacementArgs struct {
// AiRoadAndStationPlacementArgs() = default;
	// method: explicit AiRoadAndStationPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , roadObjectId(regs.dl)
// , stationObjectId(regs.dh)
// , stationLength((regs.edi >> 24) & 0xFFU)
// , mods((regs.edi >> 16) & 0xFU)
// , unk1((regs.edx >> 16) & 0xFFU)
// , bridge((regs.edx >> 24) & 0xFFU)
}
const AiRoadAndStationPlacementArgsCommand any = GameCommand.aiCreateRoadAndStation
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t roadObjectId;
// orphan member: uint8_t stationObjectId;
// orphan member: uint8_t stationLength;
// orphan member: uint8_t mods;
// orphan member: uint8_t unk1;
// orphan member: uint8_t bridge;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.bh = rotation;
// regs.edx = (roadObjectId & 0xFFU) | ((stationObjectId & 0xFFU) << 8) | ((unk1 & 0xFFU) << 16) | ((bridge & 0xFFU) << 24);
// regs.edi = (pos.z & 0xFFFFFU) | ((mods & 0xFU) << 16) | ((stationLength & 0xFFU) << 24);
// orphan member: return regs;
// func AiCreateRoadAndStation(regs registers) 
