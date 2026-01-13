package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type IndustryPlacementArgs struct {
// IndustryPlacementArgs() = default;
	// method: explicit IndustryPlacementArgs(const registers& regs)
// : pos(regs.ax, regs.cx)
// , type(regs.dl & 0x7F)
// , buildImmediately(regs.dl & 0x80)
// , srand0(regs.ebp)
// , srand1(regs.edi)
}
const IndustryPlacementArgsCommand any = GameCommand.createIndustry
// World::Pos2 pos;
// orphan member: uint8_t type;
// bool buildImmediately = false; // No scaffolding required (editor mode)
// orphan member: uint32_t srand0;
// orphan member: uint32_t srand1;
// explicit operator registers() const
// orphan member: registers regs;
// regs.ax = pos.x;
// regs.cx = pos.y;
// regs.dl = type | (buildImmediately ? 0x80 : 0);
// regs.ebp = srand0;
// regs.edi = srand1;
// regs.esi = enumValue(command); // Vanilla bug? Investigate when doing createIndustry
// orphan member: return regs;
// func CreateIndustry(regs registers) 
