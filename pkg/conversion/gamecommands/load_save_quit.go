package gamecommands

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// namespace OpenLoco::GameCommands
type LoadSaveQuitGameArgs struct {
type SaveMode int

const (
	PromptSave SaveMode = iota
	CloseSavePrompt
	DontSave
)
// LoadSaveQuitGameArgs() = default;
	// method: explicit LoadSaveQuitGameArgs(const registers& regs)
// : loadQuitMode(static_cast<LoadOrQuitMode>(regs.di))
// , saveMode(static_cast<SaveMode>(regs.dl))
}
const LoadSaveQuitGameArgsCommand any = GameCommand.loadSaveQuitGame
// orphan member: LoadOrQuitMode loadQuitMode;
// orphan member: SaveMode saveMode;
// explicit operator registers() const
// orphan member: registers regs;
// regs.di = enumValue(loadQuitMode); // [ 0 = load game, 1 = return to title screen, 2 = quit to desktop ]
// regs.dl = enumValue(saveMode);     // [ 0 = prompt save, 1 = close save prompt, 2 = don't save ]
// orphan member: return regs;
// func LoadSaveQuit(regs registers) 
