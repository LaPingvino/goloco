package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleCreateArgs struct {
// VehicleCreateArgs() = default;
	// method: explicit VehicleCreateArgs(const registers& regs)
// : vehicleId(static_cast<EntityId>(regs.di))
// , vehicleType(regs.dx)
}
const VehicleCreateArgsCommand any = GameCommand.vehicleCreate
// orphan member: EntityId vehicleId; // Optional id representing where it will attach
// orphan member: uint16_t vehicleType;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(vehicleId);
// regs.edx = vehicleType;
// orphan member: return regs;
// func CreateVehicle(regs registers) 
