package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type ChangeCompanyColourSchemeArgs struct {
// ChangeCompanyColourSchemeArgs() = default;
	// method: explicit ChangeCompanyColourSchemeArgs(const registers& regs)
// : companyId(CompanyId(regs.dl))
// , isPrimary()
// , value(regs.al)
// , colourType(regs.cl)
// , setColourMode(regs.dh)
// if (!setColourMode)
// isPrimary = regs.ah == 0;
}
const ChangeCompanyColourSchemeArgsCommand any = GameCommand.changeCompanyColourScheme
// orphan member: CompanyId companyId;
// orphan member: bool isPrimary;
// orphan member: uint8_t value;
// orphan member: uint8_t colourType;
// orphan member: bool setColourMode;
// explicit operator registers() const
// orphan member: registers regs;
// regs.cl = colourType;           // vehicle type or main
// regs.dh = setColourMode;        // [ 0, 1 ] -- 0 = set colour, 1 = toggle enabled/disabled;
// regs.dl = enumValue(companyId); // company id
// if (!setColourMode)
// // cl is divided by 2 when used
// regs.ah = isPrimary ? 1 : 0; // [ 0, 1 ] -- primary or secondary palette
// regs.al = value;             // new colour
// func If(setColourMode) else
// regs.al = value; // [ 0, 1 ] -- off or on
// orphan member: return regs;
// func ChangeCompanyColour(regs registers) 
