package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleOrderDownArgs struct {
// VehicleOrderDownArgs() = default;
	// method: explicit VehicleOrderDownArgs(const registers& regs)
// : head(EntityId(regs.di))
// , orderOffset(regs.edx)
}
const VehicleOrderDownArgsCommand any = GameCommand.vehicleOrderDown
// orphan member: EntityId head;
// orphan member: uint32_t orderOffset;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// regs.edx = orderOffset;
// orphan member: return regs;
// func VehicleOrderDown(regs registers) 
