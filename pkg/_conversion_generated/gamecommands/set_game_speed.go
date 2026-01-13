package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type SetGameSpeedArgs struct {
// SetGameSpeedArgs() = default;
	// method: explicit SetGameSpeedArgs(const registers& regs)
// : newSpeed(static_cast<GameSpeed>(regs.edi))
}
const SetGameSpeedArgsCommand any = GameCommand.setGameSpeed
// func SetGameSpeedArgs(speed GameSpeed) explicit
// newSpeed = speed;
// orphan member: GameSpeed newSpeed;
// explicit operator registers() const
// orphan member: registers regs;
// regs.edi = static_cast<std::underlying_type_t<GameSpeed>>(newSpeed);
// orphan member: return regs;
// func SetGameSpeed(regs registers) 
