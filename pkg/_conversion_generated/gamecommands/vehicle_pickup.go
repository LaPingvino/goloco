package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehiclePickupArgs struct {
// VehiclePickupArgs() = default;
	// method: explicit VehiclePickupArgs(const registers& regs)
// : head(static_cast<EntityId>(regs.di))
}
const VehiclePickupArgsCommand any = GameCommand.vehiclePickup
// orphan member: EntityId head;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// orphan member: return regs;
// func VehiclePickup(regs registers) 
