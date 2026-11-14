package shortcuts

import "testing"

// The default bindings as Shortcuts.cpp's initialize gives them, parsed.
func TestDefaults(t *testing.T) {
	b := Load(nil)
	for _, c := range []struct {
		id   Shortcut
		key  uint32
		mods KeyModifier
	}{
		{CloseTopmostWindow, KeyBackspace, ModNone},
		{CloseAllFloatingWindows, KeyBackspace, ModShift},
		{CancelConstructionMode, KeyEscape, ModNone},
		{PauseUnpauseGame, scancodeMask | 72, ModNone},
		{ZoomViewOut, scancodeMask | 75, ModNone},
		{ZoomViewIn, scancodeMask | 78, ModNone},
		{RotateView, KeyReturn, ModNone},
		{RotateConstructionObject, 'z', ModNone},
		{ToggleUndergroundView, '1', ModNone},
		{ToggleDirArrowsonTracks, '0', ModNone},
		{BuildTracks, 't', ModNone},
		{BuildRoads, 'r', ModNone},
		{ShowAnnouncementsList, KeyTab, ModNone},
		{Screenshot, 's', ModLeftControl},
		{ToggleLastAnnouncement, KeySpace, ModNone},
		{SendMessage, scancodeMask | 58, ModNone},
		{OpenDebugWindow, scancodeMask | 67, ModNone},
		{ShowOptionsWindow, InvalidKeyCode, ModInvalid},
		{GameSpeedNormal, InvalidKeyCode, ModInvalid},
	} {
		if got := b.Get(c.id); got != (Binding{c.key, c.mods}) {
			t.Errorf("%s: %+v, want key %d mods %d", Defs[c.id].ConfigName, got, c.key, c.mods)
		}
	}
	if len(b.Config) != int(Count) {
		t.Errorf("config has %d shortcuts, want %d", len(b.Config), Count)
	}
	for i, d := range Defs {
		if d.ID != Shortcut(i) {
			t.Errorf("Defs[%d] is %d", i, d.ID)
		}
		if got := FormatBinding(b.Get(d.ID)); got != d.DefaultBinding { // SDL's names round-trip
			t.Errorf("%s: formats as %q, default %q", d.ConfigName, got, d.DefaultBinding)
		}
	}
}

func TestMatch(t *testing.T) {
	b := Load(nil)
	// plain S is the stations list, Left Ctrl+S the screenshot, Right Ctrl+S nothing
	if id, ok := b.Find('s', ModNone); !ok || id != ShowStationsList {
		t.Errorf("s: %v %v", id, ok)
	}
	if id, ok := b.Find('s', ModLeftControl); !ok || id != Screenshot {
		t.Errorf("Left Ctrl+s: %v %v", id, ok)
	}
	if _, ok := b.Find('s', ModRightControl); ok {
		t.Error("Right Ctrl+s matched")
	}
	if _, ok := b.Find('s', ModShift); ok { // the modifiers must be the binding's exactly
		t.Error("Shift+s matched")
	}
	if id, ok := b.Find('s', ModLeftControl|ModCheat); !ok || id != Screenshot { // the cheat modifier is left out
		t.Errorf("cheat Left Ctrl+s: %v %v", id, ok)
	}
	if id, ok := b.Find(KeyBackspace, ModShift); !ok || id != CloseAllFloatingWindows {
		t.Errorf("Shift+Backspace: %v %v", id, ok)
	}
	if _, ok := b.Find(KeyKPEnter, ModNone); ok {
		t.Error("Keypad Enter matched")
	}
}

func TestSetAndConfig(t *testing.T) {
	b := Load(map[string]string{"pauseUnpauseGame": "P", "showMap": "", "rotateView": "Right Ctrl+Left Alt+Keypad Enter"})
	if got := b.Get(PauseUnpauseGame); got != (Binding{'p', ModNone}) {
		t.Errorf("pause from config: %+v", got)
	}
	if got := b.Get(ShowMap); got != Unbound {
		t.Errorf("an empty binding: %+v", got)
	}
	if got := b.Get(RotateView); got != (Binding{KeyKPEnter, ModRightControl | ModLeftAlt}) {
		t.Errorf("rotateView: %+v", got)
	}
	// both pause and plantTrees are on P: the list's first wins
	if id, _ := b.Find('p', ModNone); id != PauseUnpauseGame {
		t.Errorf("p: %v", id)
	}
	// setBinding takes the key from whoever had it
	b.Set(ShowMap, 'p', ModNone|ModCheat)
	if b.Get(PauseUnpauseGame) != Unbound || b.Get(PlantTrees) != Unbound || b.Get(ShowMap) != (Binding{'p', ModNone}) {
		t.Errorf("after Set: %+v %+v %+v", b.Get(PauseUnpauseGame), b.Get(PlantTrees), b.Get(ShowMap))
	}
	if b.Config["pauseUnpauseGame"] != "" || b.Config["plantTrees"] != "" || b.Config["showMap"] != "P" {
		t.Errorf("config after Set: %q %q %q", b.Config["pauseUnpauseGame"], b.Config["plantTrees"], b.Config["showMap"])
	}
	b.Set(Screenshot, 's', ModRightControl|ModShift)
	if b.Config["makeScreenshot"] != "Left Shift+Right Ctrl+S" {
		t.Errorf("format: %q", b.Config["makeScreenshot"])
	}
	b.Reset()
	if b.Get(PauseUnpauseGame) != (Binding{scancodeMask | 72, ModNone}) || b.Config["showMap"] != "M" {
		t.Errorf("after Reset: %+v %q", b.Get(PauseUnpauseGame), b.Config["showMap"])
	}
}

func TestKeyNames(t *testing.T) {
	for name, code := range map[string]uint32{"Z": 'z', "z": 'z', "1": '1', "Return": KeyReturn, "return": KeyReturn, "Escape": KeyEscape,
		"Left Ctrl": KeyLCtrl, "Right Alt": KeyRAlt, "Insert": KeyInsert, "Up": KeyUp, "Keypad Enter": KeyKPEnter, "No such key": 0} {
		if got := KeyFromName(name); got != code {
			t.Errorf("KeyFromName(%q) = %d, want %d", name, got, code)
		}
	}
	if KeyName('z') != "Z" || KeyName(KeyReturn) != "Return" || KeyName(KeySpace) != "Space" || KeyName(InvalidKeyCode) != "" {
		t.Errorf("KeyName: %q %q %q %q", KeyName('z'), KeyName(KeyReturn), KeyName(KeySpace), KeyName(InvalidKeyCode))
	}
}
