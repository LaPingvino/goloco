package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleSpeedControlArgs struct {
// VehicleSpeedControlArgs() = default;
	// method: explicit VehicleSpeedControlArgs(const registers& regs)
// : head(static_cast<EntityId>(regs.cx))
// , speed(regs.dx)
}
const VehicleSpeedControlArgsCommand any = GameCommand.vehicleSpeedControl
// orphan member: EntityId head;
// orphan member: int16_t speed;
// explicit operator registers() const
// orphan member: registers regs;
// regs.cx = enumValue(head);
// regs.dx = speed;
// orphan member: return regs;
// func VehicleSpeedControl(regs registers) 
