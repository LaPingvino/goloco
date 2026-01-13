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
const RoadPlacementArgsCommand any = GameCommand.createRoad
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t roadId;
// orphan member: uint8_t mods;
// orphan member: uint8_t bridge;
// orphan member: uint8_t roadObjectId;
// orphan member: uint8_t unkFlags;
// explicit operator registers() const
// orphan member: registers regs;
// regs.eax = pos.x;
// regs.cx = pos.y;
// regs.edi = (0xFFFFU & pos.z) | (mods << 16);
// regs.bh = rotation;
// regs.edx = roadObjectId | (roadId << 8) | (unkFlags << 16) | (bridge << 24);
// orphan member: return regs;
// func CreateRoad(regs registers) 
