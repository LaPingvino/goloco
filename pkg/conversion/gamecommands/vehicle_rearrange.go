package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleRearrangeArgs struct {
// VehicleRearrangeArgs() = default;
	// method: explicit VehicleRearrangeArgs(const registers& regs)
// : source(static_cast<EntityId>(regs.dx))
// , dest(static_cast<EntityId>(regs.di))
}
const VehicleRearrangeArgsCommand any = GameCommand.vehicleRearrange
// orphan member: EntityId source;
// orphan member: EntityId dest;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(dest);
// regs.dx = enumValue(source);
// orphan member: return regs;
// func VehicleRearrange(regs registers) 
