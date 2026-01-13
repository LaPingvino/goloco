package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type VehicleOrderInsertArgs struct {
// VehicleOrderInsertArgs() = default;
	// method: explicit VehicleOrderInsertArgs(const registers& regs)
// : head(EntityId(regs.di))
// , orderOffset(regs.dx)
// , rawOrder((static_cast<uint64_t>(regs.cx) << 32ULL) | static_cast<uint32_t>(regs.eax))
}
const VehicleOrderInsertArgsCommand any = GameCommand.vehicleOrderInsert
// orphan member: EntityId head;
// orphan member: uint32_t orderOffset;
// orphan member: uint64_t rawOrder;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(head);
// regs.dx = orderOffset;
// regs.eax = rawOrder & 0xFFFFFFFF;
// regs.cx = rawOrder >> 32;
// orphan member: return regs;
// func VehicleOrderInsert(regs registers) 
