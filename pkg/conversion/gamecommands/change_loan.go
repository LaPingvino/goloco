package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type ChangeLoanArgs struct {
// ChangeLoanArgs() = default;
	// method: explicit ChangeLoanArgs(const registers& regs)
// : newLoan(regs.edx)
}
const ChangeLoanArgsCommand any = GameCommand.changeLoan
// orphan member: currency32_t newLoan;
// explicit operator registers() const
// orphan member: registers regs;
// regs.edx = newLoan;
// orphan member: return regs;
// func ChangeLoan(regs registers) 
