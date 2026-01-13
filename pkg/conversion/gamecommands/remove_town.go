package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type TownRemovalArgs struct {
// TownRemovalArgs() = default;
	// method: explicit TownRemovalArgs(const registers& regs)
// : townId(TownId(regs.edi))
}
const TownRemovalArgsCommand any = GameCommand.removeTown
// orphan member: TownId townId;
// explicit operator registers() const
// orphan member: registers regs;
// regs.edi = enumValue(townId);
// orphan member: return regs;
// func RemoveTown(regs registers) 
