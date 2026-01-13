package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include <array>
// namespace OpenLoco::GameCommands
type ChangeCompanyOwnerNameArgs struct {
// ChangeCompanyOwnerNameArgs() = default;
	// method: explicit ChangeCompanyOwnerNameArgs(const registers& regs)
// : companyId(CompanyId(regs.cx))
// , bufferIndex(regs.ax)
// memcpy(newName, &regs.edx, 4);
// memcpy(newName + 4, &regs.ebp, 4);
// memcpy(newName + 8, &regs.edi, 4);
}
const ChangeCompanyOwnerNameArgsCommand any = GameCommand.changeCompanyOwnerName
// orphan member: CompanyId companyId;
// orphan member: uint16_t bufferIndex;
// char newName[37];
// explicit operator registers() const
// orphan member: registers regs;
// regs.cl = enumValue(companyId);
// regs.ax = bufferIndex; // [ 0, 1, 2]
var IToOffset = [3]uint8{
// const auto offset = iToOffset[bufferIndex];
// std::memcpy(&regs.edx, newName + offset, 4);
// std::memcpy(&regs.ebp, newName + offset + 4, 4);
// std::memcpy(&regs.edi, newName + offset + 8, 4);
// orphan member: return regs;
// func ChangeCompanyOwnerName(regs registers) 
