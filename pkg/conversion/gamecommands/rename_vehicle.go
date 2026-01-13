package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include <array>
// namespace OpenLoco::GameCommands
type VehicleRenameArgs struct {
// VehicleRenameArgs() = default;
	// method: explicit VehicleRenameArgs(const registers& regs)
// : head(static_cast<EntityId>(regs.cx))
// , buffer{}
// , i(regs.ax)
// // Copies it into the first 12 bytes not into the specific slot as per i
// std::memcpy(buffer, &regs.edx, 4);
// std::memcpy(buffer + 4, &regs.ebp, 4);
// std::memcpy(buffer + 8, &regs.edi, 4);
}
const VehicleRenameArgsCommand any = GameCommand.vehicleRename
// orphan member: EntityId head;
// char buffer[37];
// orphan member: uint16_t i;
// explicit operator registers() const
// orphan member: registers regs;
// regs.cx = enumValue(head);
// regs.ax = i;
var IToOffset = [3]uint8{
// const auto offset = iToOffset[i];
// std::memcpy(&regs.edx, buffer + offset, 4);
// std::memcpy(&regs.ebp, buffer + offset + 4, 4);
// std::memcpy(&regs.edi, buffer + offset + 8, 4);
// orphan member: return regs;
// func RenameVehicle(regs registers) 
