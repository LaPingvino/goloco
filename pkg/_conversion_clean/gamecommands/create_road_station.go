package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type RoadStationPlacementArgs struct {
	// RoadStationPlacementArgs() = default;
	// method: explicit RoadStationPlacementArgs(const registers& regs)
	// : pos(regs.ax, regs.cx, regs.di)
	// , rotation(regs.bh & 0x3)
	// , roadId(regs.dl & 0xF)
	// , index(regs.dh & 0x3)
	// , roadObjectId(regs.bp)
	// , type(regs.edi >> 16)
}

// MALFORMED FIELD: const RoadStationPlacementArgsCommand any = GameCommand.createRoadStation

// World::Pos3 pos;
// orphan member: uint8 rotation;
// orphan member: uint8 roadId;
// orphan member: uint8 index;
// orphan member: uint8 roadObjectId;
// orphan member: uint8 type;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.edi = pos.z | (type << 16);
// regs.bh = rotation;
// regs.dl = roadId;
// regs.dh = index;
// regs.bp = roadObjectId;
// orphan member: return regs;
// func CreateRoadStation(regs registers)