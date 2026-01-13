package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include <array>
// namespace OpenLoco::GameCommands
type RenameTownArgs struct {
// RenameTownArgs() = default;
	// method: explicit RenameTownArgs(const registers& regs)
// : townId(TownId(regs.cx))
// , nameBufferIndex(regs.ax)
// , buffer{}
// std::memcpy(buffer, &regs.edx, 4);
// std::memcpy(buffer + 4, &regs.ebp, 4);
// std::memcpy(buffer + 8, &regs.edi, 4);
}
const RenameTownArgsCommand any = GameCommand.renameTown
// orphan member: TownId townId;
// orphan member: uint8_t nameBufferIndex;
// char buffer[37];
// explicit operator registers() const
// orphan member: registers regs;
// regs.cx = enumValue(townId);
// regs.ax = nameBufferIndex;
var IToOffset = [3]uint8{
// const auto offset = iToOffset[nameBufferIndex];
// std::memcpy(&regs.edx, buffer + offset, 4);
// std::memcpy(&regs.ebp, buffer + offset + 4, 4);
// std::memcpy(&regs.edi, buffer + offset + 8, 4);
// orphan member: return regs;
// func RenameTown(regs registers) 
