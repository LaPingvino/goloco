package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include "Map/Tile.h"
// #include <set>
// namespace OpenLoco::GameCommands
type LowerRaiseLandMountainArgs struct {
// LowerRaiseLandMountainArgs() = default;
	// method: explicit LowerRaiseLandMountainArgs(const registers& regs)
// : centre(regs.ax, regs.cx)
// , pointA(regs.dx, regs.bp)
// , pointB(regs.edx >> 16, regs.ebp >> 16)
// , adjustment(regs.di)
}
const LowerRaiseLandMountainArgsCommand any = GameCommand.lowerRaiseLandMountain
// World::Pos2 centre;
// World::Pos2 pointA;
// World::Pos2 pointB;
// orphan member: int16_t adjustment;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = centre.x;
// regs.cx = centre.y;
// regs.edx = (pointB.x << 16) | pointA.x;
// regs.ebp = (pointB.y << 16) | pointA.y;
// regs.di = adjustment;
// orphan member: return regs;
// func LowerRaiseLandMountain(regs registers) 
