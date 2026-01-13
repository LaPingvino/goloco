package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type WallPlacementArgs struct {
// WallPlacementArgs() = default;
	// method: explicit WallPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
// , rotation(regs.dl)
// , type(regs.bh)
// , primaryColour(static_cast<Colour>(regs.dh))
// , secondaryColour(static_cast<Colour>(regs.bp & 0x1F))
// , tertiaryColour(static_cast<Colour>((regs.bp >> 8) & 0x1F))
}
const WallPlacementArgsCommand any = GameCommand.createWall
// World::Pos3 pos;
// orphan member: uint8_t rotation;
// orphan member: uint8_t type;
// orphan member: Colour primaryColour;
// orphan member: Colour secondaryColour;
// orphan member: Colour tertiaryColour; // Note: will not work; render engine does not support tertiary
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.dl = rotation;
// regs.dh = enumValue(primaryColour);
// regs.di = pos.z;
// regs.bp = enumValue(secondaryColour) | (enumValue(tertiaryColour) << 8);
// regs.bh = type;
// orphan member: return regs;
// func CreateWall(regs registers) 
