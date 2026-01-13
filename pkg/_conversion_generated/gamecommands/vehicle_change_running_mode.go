package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleChangeRunningModeArgs struct {
type Mode int

const (
	StopVehicle Mode = iota
	StartVehicle
	ToggleLocalExpress
	DriveManually
)
// VehicleChangeRunningModeArgs() = default;
	// method: explicit VehicleChangeRunningModeArgs(const registers& regs)
// : head(static_cast<EntityId>(regs.dx))
// , mode(static_cast<Mode>(regs.bh))
}
const VehicleChangeRunningModeArgsCommand any = GameCommand.vehicleChangeRunningMode
// orphan member: EntityId head;
// orphan member: Mode mode;
// explicit operator registers() const
// orphan member: registers regs;
// regs.dx = enumValue(head);
// regs.bh = enumValue(mode);
// orphan member: return regs;
// func VehicleChangeRunningMode(regs registers) 
