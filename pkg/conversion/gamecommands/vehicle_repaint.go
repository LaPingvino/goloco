package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <array>
// namespace OpenLoco::GameCommands
type QuadraColour = [4]ColourScheme
const BodyColour uint8 = 0
const FrontBogieColour uint8 = 1
const BackBogieColour uint8 = 2
type VehicleRepaintFlags int

const (
	None VehicleRepaintFlags = 0
	BodyColour VehicleRepaintFlags = (1 << 0)
	FrontBogieColour VehicleRepaintFlags = (1 << 1)
	BackBogieColour VehicleRepaintFlags = (1 << 2)
	ApplyToEntireCar VehicleRepaintFlags = (1 << 3)
	ApplyToEntireTrain VehicleRepaintFlags = (1 << 4)
	PaintFromVehicleUi VehicleRepaintFlags = bodyColour | frontBogieColour | backBogieColour | applyToEntireCar
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(VehicleRepaintFlags);
type VehicleRepaintArgs struct {
// VehicleRepaintArgs() = default;
	// method: explicit VehicleRepaintArgs(const registers& regs)
// : head(EntityId(regs.ebp))
// , colours{ ColourScheme(regs.cx), ColourScheme(regs.ecx >> 16), ColourScheme(regs.dx), ColourScheme(regs.edx >> 16) }
// , paintFlags(VehicleRepaintFlags(regs.ax))
}
const VehicleRepaintArgsCommand any = GameCommand.vehicleRepaint
// func HasRepaintFlags(flags VehicleRepaintFlags) bool
// return (paintFlags & flags) != VehicleRepaintFlags::none;
// func SetColours(colour ColourScheme) 
// if (hasRepaintFlags(VehicleRepaintFlags::bodyColour))
// colours[kBodyColour] = colour;
// if (hasRepaintFlags(VehicleRepaintFlags::frontBogieColour))
// colours[kFrontBogieColour] = colour;
// if (hasRepaintFlags(VehicleRepaintFlags::backBogieColour))
// colours[kBackBogieColour] = colour;
// orphan member: EntityId head;
// orphan member: QuadraColour colours;
// orphan member: VehicleRepaintFlags paintFlags;
// func Convert(colour ColourScheme) uint16
// func EnumValue(colour.primary) return
// explicit operator registers() const
// orphan member: registers regs;
// regs.ebp = enumValue(head);
// regs.ecx = convert(colours[0]) | (convert(colours[1]) << 16);
// regs.edx = convert(colours[2]) | (convert(colours[3]) << 16);
// regs.ax = enumValue(paintFlags);
// orphan member: return regs;
// func VehicleRepaint(regs registers) 
