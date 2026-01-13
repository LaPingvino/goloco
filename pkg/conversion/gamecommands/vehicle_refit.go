package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleRefitArgs struct {
// VehicleRefitArgs() = default;
	// method: explicit VehicleRefitArgs(const registers& regs)
// : head(static_cast<EntityId>(regs.di))
// , cargoType(regs.dl)
}
const VehicleRefitArgsCommand any = GameCommand.vehicleRefit
// orphan member: EntityId head;
// orphan member: uint8_t cargoType;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// regs.dl = cargoType;
// orphan member: return regs;
// func VehicleRefit(regs registers) 
