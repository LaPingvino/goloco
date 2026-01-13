package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleCloneArgs struct {
// VehicleCloneArgs() = default;
	// method: explicit VehicleCloneArgs(const registers& regs)
// : vehicleHeadId(static_cast<EntityId>(regs.ax))
}
const VehicleCloneArgsCommand any = GameCommand.vehicleClone
// orphan member: EntityId vehicleHeadId;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = enumValue(vehicleHeadId);
// orphan member: return regs;
// func CloneVehicle(regs registers) 
