package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleOrderReverseArgs struct {
// VehicleOrderReverseArgs() = default;
	// method: explicit VehicleOrderReverseArgs(const registers& regs)
// : head(EntityId(regs.di))
}
const VehicleOrderReverseArgsCommand any = GameCommand.vehicleOrderReverse
// orphan member: EntityId head;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// orphan member: return regs;
// func VehicleOrderReverse(regs registers) 
