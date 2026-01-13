package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleReverseArgs struct {
// VehicleReverseArgs() = default;
	// method: explicit VehicleReverseArgs(const registers& regs)
// : head(static_cast<EntityId>(regs.dx))
}
const VehicleReverseArgsCommand any = GameCommand.vehicleReverse
// orphan member: EntityId head;
// explicit operator registers() const
// orphan member: registers regs;
// regs.dx = enumValue(head);
// orphan member: return regs;
// func VehicleReverse(regs registers) 
