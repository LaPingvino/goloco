package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include "World/Company.h"
// namespace OpenLoco::GameCommands
type UpdateOwnerStatusArgs struct {
// UpdateOwnerStatusArgs() = default;
	// method: explicit UpdateOwnerStatusArgs(const registers& regs)
// : ownerStatus(regs.ax, regs.cx)
}
const UpdateOwnerStatusArgsCommand any = GameCommand.updateOwnerStatus
// orphan member: OwnerStatus ownerStatus;
// explicit operator registers() const
// orphan member: registers regs;
// int16_t res[2];
// ownerStatus.getData(res);
// regs.ax = res[0];
// regs.cx = res[1];
// orphan member: return regs;
// func UpdateOwnerStatus(regs registers) 
