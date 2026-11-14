// Package shortcuts is the keyboard shortcuts' table and their key bindings:
// which key, with which modifiers, runs which shortcut. The actions
// themselves are the game's (cmd/goloco shortcuts.go).
//
// OpenLoco reference: src/OpenLoco/src/Input/Shortcuts.cpp (initialize,
// parseBinding, formatBinding, loadBindings, resetBindings, getBinding,
// setBinding), src/OpenLoco/include/OpenLoco/Input/Shortcuts.h,
// src/OpenLoco/include/OpenLoco/Input.h (KeyModifier),
// src/Engine/src/Input/ShortcutManager.cpp.
//
// Keys are SDL keycodes, as upstream's and its config file's: a printable
// key is its lower-case character, the others SDL's own numbers
// (table_gen.go, with SDL's names for them — "Left Ctrl+S", "PageUp").
package shortcuts

import "strings"

// Shortcut is Input::Shortcut.
type Shortcut uint32

// Def is ShortcutManager's ShortcutEntry without its action.
type Def struct {
	ID             Shortcut
	DisplayName    string // the StringIds name of the shortcut's description
	DisplayID      int    // its StringIds id (the KeyboardShortcuts window's text)
	ConfigName     string
	DefaultBinding string
}

// KeyModifier is Input::KeyModifier.
type KeyModifier uint8

const (
	ModNone         KeyModifier = 0
	ModShift        KeyModifier = 1 << 0
	ModLeftControl  KeyModifier = 1 << 1
	ModUnknown      KeyModifier = 1 << 2
	ModLeftAlt      KeyModifier = 1 << 3
	ModRightControl KeyModifier = 1 << 4
	ModRightAlt     KeyModifier = 1 << 5
	ModCheat        KeyModifier = 1 << 7
	ModInvalid      KeyModifier = 0xFF

	ModControl = ModLeftControl | ModRightControl
	ModAlt     = ModLeftAlt | ModRightAlt
)

// The keycodes the game names itself (SDL_keycode.h).
const (
	InvalidKeyCode uint32 = 0xFFFFFFFF // kInvalidKeyCode

	KeyBackspace uint32 = 8
	KeyTab       uint32 = 9
	KeyReturn    uint32 = 13
	KeyEscape    uint32 = 27
	KeySpace     uint32 = 32
	KeyDelete    uint32 = 127

	scancodeMask = 1 << 30

	KeyInsert  uint32 = scancodeMask | 73
	KeyHome    uint32 = scancodeMask | 74
	KeyEnd     uint32 = scancodeMask | 77
	KeyRight   uint32 = scancodeMask | 79
	KeyLeft    uint32 = scancodeMask | 80
	KeyDown    uint32 = scancodeMask | 81
	KeyUp      uint32 = scancodeMask | 82
	KeyKPEnter uint32 = scancodeMask | 88
	KeyLCtrl   uint32 = scancodeMask | 224
	KeyLShift  uint32 = scancodeMask | 225
	KeyLAlt    uint32 = scancodeMask | 226
	KeyLGUI    uint32 = scancodeMask | 227
	KeyRCtrl   uint32 = scancodeMask | 228
	KeyRShift  uint32 = scancodeMask | 229
	KeyRAlt    uint32 = scancodeMask | 230
	KeyRGUI    uint32 = scancodeMask | 231
)

// KeyFromName is SDL_GetKeyFromName: a single character is its own key
// (letters lower-case); other names are SDL's key names, whatever their
// case. 0 (SDLK_UNKNOWN) for no such key.
func KeyFromName(name string) uint32 {
	if r := []rune(name); len(r) == 1 {
		c := r[0]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		return uint32(c)
	}
	for _, k := range sdlKeys {
		if strings.EqualFold(k.Name, name) {
			return k.Code
		}
	}
	return 0
}

// KeyName is SDL_GetKeyName: SDL's name of a key ("" for none) — a
// printable key's character, letters upper-case.
func KeyName(code uint32) string {
	for _, k := range sdlKeys {
		if k.Code == code {
			return k.Name
		}
	}
	if code >= 0x20 && code < scancodeMask && code != KeyDelete {
		if code >= 'a' && code <= 'z' {
			code -= 'a' - 'A'
		}
		return string(rune(code))
	}
	return ""
}

// KeyFromScancode is the key of an SDL scancode on SDL's default layout
// (0: none).
func KeyFromScancode(sc uint16) uint32 {
	for _, k := range sdlKeys {
		if k.Scancode == sc {
			return k.Code
		}
	}
	return 0
}

// Binding is KeyboardBinding: unbound by default.
type Binding struct {
	KeyCode   uint32
	Modifiers KeyModifier
}

// Unbound is KeyboardBinding{}.
var Unbound = Binding{InvalidKeyCode, ModInvalid}

const bindingDelimiter = "+" // kBindingDelimiter

// ParseBinding is parseBinding: "Left Shift+Backspace" — the modifiers'
// keys, then the key; "" is unbound.
func ParseBinding(s string) Binding {
	if s == "" {
		return Unbound
	}
	res := Binding{InvalidKeyCode, ModNone}
	for {
		token, rest, found := strings.Cut(s, bindingDelimiter)
		if !found {
			break
		}
		s = rest
		switch KeyFromName(token) {
		case KeyLShift, KeyRShift:
			res.Modifiers |= ModShift
		case KeyLCtrl:
			res.Modifiers |= ModLeftControl
		case KeyRCtrl:
			res.Modifiers |= ModRightControl
		case KeyLAlt:
			res.Modifiers |= ModLeftAlt
		case KeyRAlt:
			res.Modifiers |= ModRightAlt
		case KeyLGUI, KeyRGUI:
			res.Modifiers |= ModUnknown
		}
	}
	res.KeyCode = KeyFromName(s)
	return res
}

// FormatBinding is formatBinding: the config file's text of a binding.
func FormatBinding(b Binding) string {
	if b.KeyCode == InvalidKeyCode || b.Modifiers == ModInvalid {
		return ""
	}
	var name string
	for _, m := range [...]struct {
		mod KeyModifier
		key uint32
	}{{ModShift, KeyLShift}, {ModLeftControl, KeyLCtrl}, {ModRightControl, KeyRCtrl},
		{ModLeftAlt, KeyLAlt}, {ModRightAlt, KeyRAlt}, {ModUnknown, KeyLGUI}} {
		if b.Modifiers&m.mod != 0 {
			name += KeyName(m.key) + bindingDelimiter
		}
	}
	return name + KeyName(b.KeyCode)
}

// Bindings are the shortcuts' keys (_bindings) with the config's shortcuts
// section they are read from and written to (Config::shortcuts: config name
// → binding text).
type Bindings struct {
	keys   [Count]Binding
	Config map[string]string
}

// Load is loadBindings: each shortcut's binding from the config, a shortcut
// the config doesn't name taking (and getting there) its default.
func Load(config map[string]string) *Bindings {
	b := &Bindings{Config: map[string]string{}}
	for k, v := range config {
		b.Config[k] = v
	}
	for _, def := range Defs {
		text, ok := b.Config[def.ConfigName]
		if !ok {
			text = def.DefaultBinding
			b.Config[def.ConfigName] = text
		}
		b.keys[def.ID] = ParseBinding(text)
	}
	return b
}

// Reset is resetBindings: every shortcut back on its default (the caller
// writes the config).
func (b *Bindings) Reset() {
	for _, def := range Defs {
		b.Config[def.ConfigName] = def.DefaultBinding
		b.keys[def.ID] = ParseBinding(def.DefaultBinding)
	}
}

// Get is getBinding.
func (b *Bindings) Get(id Shortcut) Binding {
	if id >= Count {
		return Unbound
	}
	return b.keys[id]
}

// Set is setBinding: the shortcut on that key; a shortcut that had the key
// loses it (the caller writes the config).
func (b *Bindings) Set(id Shortcut, keyCode uint32, modifiers KeyModifier) {
	modifiers &^= ModCheat
	for _, def := range Defs {
		k := &b.keys[def.ID]
		switch {
		case def.ID == id:
			*k = Binding{keyCode, modifiers}
		case k.KeyCode == keyCode && k.Modifiers == modifiers:
			*k = Unbound // "Unbind any shortcut that is already using this keycode."
		default:
			continue
		}
		b.Config[def.ConfigName] = FormatBinding(*k)
	}
}

// Match is Keyboard.cpp's tryShortcut's test: the shortcut's key with
// exactly its modifiers (the cheat modifier aside).
func (b *Bindings) Match(id Shortcut, keyCode uint32, modifiers KeyModifier) bool {
	k := b.Get(id)
	return k.KeyCode == keyCode && k.Modifiers == modifiers&^ModCheat
}

// Find is normalKey's walk of ShortcutManager's list: the first shortcut,
// in the enum's order, bound to the key with those modifiers.
func (b *Bindings) Find(keyCode uint32, modifiers KeyModifier) (Shortcut, bool) {
	for _, def := range Defs {
		if b.Match(def.ID, keyCode, modifiers) {
			return def.ID, true
		}
	}
	return 0, false
}
