package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type RoadPlacementArgs struct {
	// RoadPlacementArgs() = default;
	// method: explicit RoadPlacementArgs(const registers& regs)
	// : pos(regs.ax, regs.cx, regs.di)
	// , rotation(regs.bh & 0x3)
	// , roadId(regs.dh & 0xF)
	// , mods(regs.edi >> 16)
	// , bridge((regs.edx >> 24) & 0xFF)
	// , roadObjectId(regs.dl)
	// , unkFlags((regs.edx >> 16) & 0xFF)
}

// MALFORMED FIELD: const RoadPlacementArgsCommand any = GameCommand.createRoad

// World::Pos3 pos;
// orphan member: uint8 rotation;
// orphan member: uint8 roadId;
// orphan member: uint8 mods;
// orphan member: uint8 bridge;
// orphan member: uint8 roadObjectId;
// orphan member: uint8 unkFlags;
// explicit operator registers() const
// orphan member: registers regs;
// regs.eax = pos.x;
// regs.cx = pos.y;
// regs.edi = (0xFFFFU & pos.z) | (mods << 16);
// regs.bh = rotation;
// regs.edx = roadObjectId | (roadId << 8) | (unkFlags << 16) | (bridge << 24);
// orphan member: return regs;
// func CreateRoad(regs registers)
