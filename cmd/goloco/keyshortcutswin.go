package main

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/shortcuts"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// keyshortcutswin.go — Options → Controls → Customise Keys: the
// KeyboardShortcuts window and EditKeyboardShortcut (their ui/ ports), the
// bindings written to openloco.yml's shortcuts section (Config::write).

// kSeparatorPositions: a separator follows each of these in the list.
var shortcutSeparatorsAfter = map[shortcuts.Shortcut]bool{
	shortcuts.RotateConstructionObject: true, shortcuts.ToggleDirArrowsonTracks: true, shortcuts.BuildNewVehicles: true,
	shortcuts.ShowCompaniesList: true, shortcuts.ShowJukeboxWindow: true, shortcuts.SendMessage: true,
	shortcuts.ConstructionSelectPosition: true, shortcuts.GameSpeedExtraFastForward: true,
}

// windowOptionsColours are the skin's windowTitlebarColour and windowOptionsColour.
func (g *Game) windowOptionsColours() (uint8, uint8) {
	if sk := g.skin(); sk != nil {
		return sk.WindowTitlebarColour, sk.WindowOptionsColour
	}
	return 0, 0
}

// openKeyboardShortcuts is KeyboardShortcuts::open (bringToFront, else a new
// centred window).
func (g *Game) openKeyboardShortcuts() {
	if w := g.wins.Find(ui.WindowTypeKeyboardShortcuts, ui.AnyNumber); w != nil {
		g.wins.BringToFront(w)
		return
	}
	tc, wc := g.windowOptionsColours()
	w := ui.NewKeyboardShortcuts(g.sw, g.sh, tc, wc)
	w.Rows = func() []ui.ShortcutRow { // repopulateShortcutList
		var rows []ui.ShortcutRow
		for _, def := range shortcuts.Defs {
			rows = append(rows, ui.ShortcutRow{ID: int(def.ID), Name: i18n.Label(def.DisplayID), Binding: shortcuts.BindingText(g.kb.bindings.Get(def.ID))})
			if shortcutSeparatorsAfter[def.ID] {
				rows = append(rows, ui.ShortcutRow{Separator: true})
			}
		}
		return rows
	}
	w.OnRow = func(id int) { g.openEditShortcut(shortcuts.Shortcut(id)) }
	w.OnReset = func() { // resetShortcuts: resetBindings, Config::write
		g.kb.bindings.Reset()
		g.writeShortcutBindings()
	}
	g.wins.Show(w)
}

// openEditShortcut is EditKeyboardShortcut::open: the one editor, centred.
func (g *Game) openEditShortcut(id shortcuts.Shortcut) {
	g.wins.CloseType(ui.WindowTypeEditKeyboardShortcut, ui.AnyNumber)
	tc, wc := g.windowOptionsColours()
	g.wins.Show(ui.NewEditShortcut(g.sw, g.sh, int(id), i18n.Label(shortcuts.Defs[id].DisplayID), tc, wc))
}

// editShortcutWin is the open EditKeyboardShortcut window (nil: none).
func (g *Game) editShortcutWin() *ui.EditShortcutWindow {
	return winOf[*ui.EditShortcutWindow](g, ui.WindowTypeEditKeyboardShortcut, ui.AnyNumber)
}

// editShortcutKey is EditKeyboardShortcut's keyUp → editShortcut: the arrows,
// Num Lock and the system keys are ignored; any other key, with the
// modifiers held, becomes the shortcut's binding (another shortcut on it
// unbound), written, and the editor closes. It takes every key while open.
func (g *Game) editShortcutKey(w *ui.EditShortcutWindow, key keyEvent) {
	switch shortcuts.KeyName(key.keyCode) {
	case "Up", "Down", "Left", "Right", "Numlock", "Left GUI", "Right GUI":
		return
	}
	g.kb.bindings.Set(shortcuts.Shortcut(w.ID), key.keyCode, g.kb.modifier)
	g.writeShortcutBindings()
	w.Close()
}

// writeShortcutBindings is Config::write for the shortcuts section.
func (g *Game) writeShortcutBindings() {
	var kvs [][2]string
	for _, def := range shortcuts.Defs {
		kvs = append(kvs, [2]string{"shortcuts." + def.ConfigName, quoteBinding(g.kb.bindings.Config[def.ConfigName])})
	}
	setConfigValues(kvs)
}

// quoteBinding is a binding as YAML text: quoted (an empty one, and keys like
// "Keypad *", must stay strings).
func quoteBinding(s string) string { return `"` + s + `"` }
