package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
// func ChangeCompanyFace(regs registers) 
type ChangeCompanyFaceArgs struct {
// ChangeCompanyFaceArgs() = default;
	// method: explicit ChangeCompanyFaceArgs(const registers& regs)
// : companyId(CompanyId(regs.bh))
// , objHeader()
// uint8_t objData[sizeof(ObjectHeader)]{};
// uint8_t* objPtr = objData;
// std::memcpy(objPtr, &regs.eax, sizeof(uint32_t));
// objPtr += sizeof(uint32_t);
// std::memcpy(objPtr, &regs.ecx, sizeof(uint32_t));
// objPtr += sizeof(uint32_t);
// std::memcpy(objPtr, &regs.edx, sizeof(uint32_t));
// objPtr += sizeof(uint32_t);
// std::memcpy(objPtr, &regs.edi, sizeof(uint32_t));
// std::memcpy(&objHeader, objData, sizeof(objData));
}
const ChangeCompanyFaceArgsCommand any = GameCommand.changeCompanyFace
// orphan member: CompanyId companyId;
// orphan member: ObjectHeader objHeader;
// explicit operator registers() const
// orphan member: registers regs;
// regs.bh = enumValue(companyId); // company id
// uint8_t objData[sizeof(ObjectHeader)]{};
// std::memcpy(objData, &objHeader, sizeof(objData));
// const uint8_t* objPtr = objData;
// std::memcpy(&regs.eax, objPtr, sizeof(uint32_t));
// objPtr += sizeof(uint32_t);
// std::memcpy(&regs.ecx, objPtr, sizeof(uint32_t));
// objPtr += sizeof(uint32_t);
// std::memcpy(&regs.edx, objPtr, sizeof(uint32_t));
// objPtr += sizeof(uint32_t);
// std::memcpy(&regs.edi, objPtr, sizeof(uint32_t));
// orphan member: return regs;
