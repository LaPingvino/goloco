package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleOrderSkipArgs struct {
// VehicleOrderSkipArgs() = default;
	// method: explicit VehicleOrderSkipArgs(const registers& regs)
// : head(EntityId(regs.di))
}
const VehicleOrderSkipArgsCommand any = GameCommand.vehicleOrderSkip
// orphan member: EntityId head;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// orphan member: return regs;
// func VehicleOrderSkip(regs registers) 
