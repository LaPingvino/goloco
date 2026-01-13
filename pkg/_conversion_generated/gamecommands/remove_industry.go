package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type IndustryRemovalArgs struct {
// IndustryRemovalArgs() = default;
	// method: explicit IndustryRemovalArgs(const registers& regs)
// : industryId(static_cast<IndustryId>(regs.dl))
}
const IndustryRemovalArgsCommand any = GameCommand.removeIndustry
// orphan member: IndustryId industryId;
// explicit operator registers() const
// orphan member: registers regs;
// regs.dl = enumValue(industryId);
// orphan member: return regs;
// func RemoveIndustry(regs registers) 
