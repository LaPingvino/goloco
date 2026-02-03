package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleOrderDeleteArgs struct {
	// VehicleOrderDeleteArgs() = default;
	// method: explicit VehicleOrderDeleteArgs(const registers& regs)
	// : head(EntityId(regs.di))
	// , orderOffset(regs.edx)
}

// MALFORMED FIELD: const VehicleOrderDeleteArgsCommand any = GameCommand.vehicleOrderDelete

// orphan member: EntityId head;
// orphan member: uint32 orderOffset;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// regs.edx = orderOffset;
// orphan member: return regs;
// func VehicleOrderDelete(regs registers)
