package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include "Map/MapSelection.h"
// #include "Map/Tile.h"
// #include "Map/TileClearance.h"
// namespace OpenLoco::GameCommands
type LowerLandArgs struct {
// LowerLandArgs() = default;
	// method: explicit LowerLandArgs(const registers& regs)
// : centre(regs.ax, regs.cx)
// , pointA(regs.dx, regs.bp)
// , pointB(regs.edx >> 16, regs.ebp >> 16)
// , corner(static_cast<World::MapSelectionType>(regs.di))
}
const LowerLandArgsCommand any = GameCommand.lowerLand
// World::Pos2 centre;
// World::Pos2 pointA;
// World::Pos2 pointB;
// World::MapSelectionType corner;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = centre.x;
// regs.cx = centre.y;
// regs.edx = (pointB.x << 16) | pointA.x;
// regs.ebp = (pointB.y << 16) | pointA.y;
// regs.di = enumValue(corner);
// orphan member: return regs;
// func LowerLand(args LowerLandArgs, removedBuildings World::TileClearance::RemovedBuildings, flags uint8) uint32
// func LowerLand(regs registers) 
