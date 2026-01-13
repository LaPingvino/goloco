package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehiclePassSignalArgs struct {
// VehiclePassSignalArgs() = default;
	// method: explicit VehiclePassSignalArgs(const registers& regs)
// : head(static_cast<EntityId>(regs.di))
}
const VehiclePassSignalArgsCommand any = GameCommand.vehiclePassSignal
// orphan member: EntityId head;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// orphan member: return regs;
// func VehiclePassSignal(regs registers) 
