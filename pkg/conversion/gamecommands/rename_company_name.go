package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include <array>
// namespace OpenLoco::GameCommands
type ChangeCompanyNameArgs struct {
// ChangeCompanyNameArgs() = default;
	// method: explicit ChangeCompanyNameArgs(const registers& regs)
// : companyId(CompanyId(regs.cx))
// , bufferIndex(regs.ax)
// memcpy(buffer, &regs.edx, 4);
// memcpy(buffer + 4, &regs.ebp, 4);
// memcpy(buffer + 8, &regs.edi, 4);
}
const ChangeCompanyNameArgsCommand any = GameCommand.changeCompanyName
// orphan member: CompanyId companyId;
// orphan member: uint16_t bufferIndex;
// char buffer[37];
// explicit operator registers() const
// orphan member: registers regs;
// regs.cl = enumValue(companyId);
// regs.ax = bufferIndex; // [ 0, 1, 2]
var IToOffset = [3]uint8{
// const auto offset = iToOffset[bufferIndex];
// std::memcpy(&regs.edx, buffer + offset, 4);
// std::memcpy(&regs.ebp, buffer + offset + 4, 4);
// std::memcpy(&regs.edi, buffer + offset + 8, 4);
// orphan member: return regs;
// func ChangeCompanyName(regs registers) 
