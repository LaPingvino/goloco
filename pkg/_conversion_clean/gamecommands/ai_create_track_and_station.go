package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type AiTrackAndStationPlacementArgs struct {
	// AiTrackAndStationPlacementArgs() = default;
	// method: explicit AiTrackAndStationPlacementArgs(const registers& regs)
	// : pos(regs.ax, regs.cx, regs.di)
	// , rotation(regs.bh & 0x3)
	// , trackObjectId(regs.dl)
	// , stationObjectId(regs.dh)
	// , stationLength((regs.edi >> 24) & 0xFFU)
	// , mods((regs.edi >> 16) & 0xFU)
	// , unk1((regs.edx >> 16) & 0xFFU)
	// , bridge((regs.edx >> 24) & 0xFFU)
}

// MALFORMED FIELD: const AiTrackAndStationPlacementArgsCommand any = GameCommand.aiCreateTrackAndStation

// World::Pos3 pos;
// orphan member: uint8 rotation;
// orphan member: uint8 trackObjectId;
// orphan member: uint8 stationObjectId;
// orphan member: uint8 stationLength;
// orphan member: uint8 mods;
// orphan member: uint8 unk1;
// orphan member: uint8 bridge;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.bh = rotation;
// regs.edx = (trackObjectId & 0xFFU) | ((stationObjectId & 0xFFU) << 8) | ((unk1 & 0xFFU) << 16) | ((bridge & 0xFFU) << 24);
// regs.edi = (pos.z & 0xFFFFFU) | ((mods & 0xFU) << 16) | ((stationLength & 0xFFU) << 24);
// orphan member: return regs;
// func AiCreateTrackAndStation(regs registers)
