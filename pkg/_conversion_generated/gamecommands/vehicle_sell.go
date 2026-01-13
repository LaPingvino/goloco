package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleSellArgs struct {
// VehicleSellArgs() = default;
	// method: explicit VehicleSellArgs(const registers& regs)
// : car(static_cast<EntityId>(regs.dx))
}
const VehicleSellArgsCommand any = GameCommand.vehicleSell
// orphan member: EntityId car;
// explicit operator registers() const
// orphan member: registers regs;
// regs.dx = enumValue(car);
// orphan member: return regs;
// func SellVehicle(regs registers) 
