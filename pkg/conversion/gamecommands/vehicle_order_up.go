package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleOrderUpArgs struct {
// VehicleOrderUpArgs() = default;
	// method: explicit VehicleOrderUpArgs(const registers& regs)
// : head(EntityId(regs.di))
// , orderOffset(regs.edx)
}
const VehicleOrderUpArgsCommand any = GameCommand.vehicleOrderUp
// orphan member: EntityId head;
// orphan member: uint32_t orderOffset;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// regs.edx = orderOffset;
// orphan member: return regs;
// func VehicleOrderUp(regs registers) 
