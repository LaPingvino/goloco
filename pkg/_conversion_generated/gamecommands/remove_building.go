package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/Buildings/CreateBuilding.h"
// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type BuildingRemovalArgs struct {
// BuildingRemovalArgs() = default;
	// method: explicit BuildingRemovalArgs(const registers& regs)
// : pos(regs.ax, regs.cx, regs.di)
}
const BuildingRemovalArgsCommand any = GameCommand.removeBuilding
// func BuildingRemovalArgs(place BuildingPlacementArgs) explicit
// : pos(place.pos)
// World::Pos3 pos;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.di = pos.z;
// orphan member: return regs;
// func RemoveBuilding(regs registers) 
