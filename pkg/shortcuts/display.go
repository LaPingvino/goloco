package shortcuts

import "github.com/LaPingvino/goloco/pkg/i18n"

// What the KeyboardShortcuts and EditKeyboardShortcut windows show of a
// binding: its modifiers (Shortcuts::pushModifierStrings: "CTRL + ",
// "R.CTRL + ", "SHIFT + ", "ALT + ", "R.ALT + " in that order) and its key
// (KeyboardShortcuts.cpp getBindingString: the game's own names for the keys
// it names, else SDL_GetKeyName).

// keyStrings is getBindingString's keysToString, by SDL key name.
var keyStrings = map[string]int{
	"Backspace": i18n.KeyboardBackspace, "Tab": i18n.KeyboardTab, "Return": i18n.KeyboardReturn, "Pause": i18n.KeyboardPause,
	"CapsLock": i18n.KeyboardCaps, "Escape": i18n.KeyboardEscape, "Space": i18n.KeyboardSpacebar, "PageUp": i18n.KeyboardPageup,
	"PageDown": i18n.KeyboardPagedown, "End": i18n.KeyboardEnd, "Home": i18n.KeyboardHome, "Left": i18n.KeyboardLeft,
	"Up": i18n.KeyboardUp, "Right": i18n.KeyboardRight, "Down": i18n.KeyboardDown, "Insert": i18n.KeyboardInsert,
	"Delete": i18n.KeyboardDelete, "Keypad 1": i18n.KeyboardNumpad1, "Keypad 2": i18n.KeyboardNumpad2,
	"Keypad 3": i18n.KeyboardNumpad3, "Keypad 4": i18n.KeyboardNumpad4, "Keypad 5": i18n.KeyboardNumpad5,
	"Keypad 6": i18n.KeyboardNumpad6, "Keypad 7": i18n.KeyboardNumpad7, "Keypad 8": i18n.KeyboardNumpad8,
	"Keypad 9": i18n.KeyboardNumpad9, "Keypad 0": i18n.KeyboardNumpad0, "Keypad /": i18n.KeyboardNumpadDivide,
	"Keypad Enter": i18n.KeyboardNumpadEnter, "Keypad -": i18n.KeyboardNumpadMinus, "Keypad *": i18n.KeyboardNumpadMultiply,
	"Keypad .": i18n.KeyboardNumpadPeriod, "Keypad +": i18n.KeyboardNumpadPlus, "Numlock": i18n.KeyboardNumlock,
	"ScrollLock": i18n.KeyboardScroll, "Menu": i18n.KeyboardMenu,
}

// KeyDisplayName is getBindingString: the key's name as the window shows it.
func KeyDisplayName(code uint32) string {
	name := KeyName(code)
	if id, ok := keyStrings[name]; ok {
		return i18n.Str(id)
	}
	return name
}

// ModifierText is pushModifierStrings' text for the modifiers held.
func ModifierText(m KeyModifier) string {
	s := ""
	for _, x := range [...]struct {
		mod KeyModifier
		id  int
	}{{ModLeftControl, i18n.KeyboardShortcutModifierCtrl}, {ModRightControl, i18n.KeyboardShortcutModifierRightCtrl},
		{ModShift, i18n.KeyboardShortcutModifierShift}, {ModLeftAlt, i18n.KeyboardShortcutModifierAlt},
		{ModRightAlt, i18n.KeyboardShortcutModifierRightAlt}} {
		if m&x.mod != 0 {
			s += i18n.Str(x.id)
		}
	}
	return s
}

// BindingText is keyboard_shortcut_binding: the modifiers and the key; ""
// for an unbound shortcut.
func BindingText(b Binding) string {
	if b.KeyCode == InvalidKeyCode || b.Modifiers == ModInvalid {
		return ""
	}
	return ModifierText(b.Modifiers) + KeyDisplayName(b.KeyCode)
}
