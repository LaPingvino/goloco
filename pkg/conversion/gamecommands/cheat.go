package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleApplyShuntCheatArgs struct {
// VehicleApplyShuntCheatArgs() = default;
	// method: explicit VehicleApplyShuntCheatArgs(const registers& regs)
// : head(EntityId(regs.cx))
}
const VehicleApplyShuntCheatArgsCommand any = GameCommand.vehicleApplyShuntCheat
// orphan member: EntityId head;
// explicit operator registers() const
// orphan member: registers regs;
// regs.cx = enumValue(head);
// orphan member: return regs;
type ApplyFreeCashCheatArgs struct {
// ApplyFreeCashCheatArgs() = default;
	// method: explicit ApplyFreeCashCheatArgs(const registers&)
}
const ApplyFreeCashCheatArgsCommand any = GameCommand.applyFreeCashCheat
// explicit operator registers() const
// func Registers() return
type CheatCommand int

const (
	AcquireAssets CheatCommand = iota
	AddCash
	ClearLoan
	CompanyRatings
	SwitchCompany
	ToggleBankruptcy
	ToggleJail
	VehicleReliability
	ModifyDate
	CompleteChallenge
)
type GenericCheatArgs struct {
// GenericCheatArgs() = default;
	// method: explicit GenericCheatArgs(const registers& regs)
// : subcommand(static_cast<CheatCommand>(regs.bh))
// , param1(regs.eax)
// , param2(regs.ecx)
// , param3(regs.edx)
}
const GenericCheatArgsCommand any = GameCommand.cheat
// orphan member: CheatCommand subcommand{};
// orphan member: int32_t param1{};
// orphan member: int32_t param2{};
// orphan member: int32_t param3{};
// explicit operator registers() const
// orphan member: registers regs;
// regs.bh = enumValue(subcommand);
// regs.eax = param1;
// regs.ecx = param2;
// regs.edx = param3;
// orphan member: return regs;
// func VehicleShuntCheat(regs registers) 
// func Cheat(regs registers) 
// func FreeCashCheat(regs registers) 
