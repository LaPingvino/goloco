package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type BuildingPlacementArgs struct {
// BuildingPlacementArgs() = default;
	// method: explicit BuildingPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.bh & 0x3)
// , type(regs.dl)
// , variation(regs.dh)
// , colour(static_cast<Colour>((regs.edi >> 16) & 0x1F))
// , buildImmediately(regs.bh & 0x80)
}
const BuildingPlacementArgsCommand any = GameCommand.createBuilding
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t type;
// orphan member: uint8_t variation;
// orphan member: Colour colour;
// bool buildImmediately = false; // No scaffolding required (editor mode)
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.edi = pos.z | (enumValue(colour) << 16);
// regs.dl = type;
// regs.dh = variation;
// regs.bh = rotation | (buildImmediately ? 0x80 : 0);
// orphan member: return regs;
// func CreateBuilding(regs registers) 
