package main

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/LaPingvino/goloco/pkg/world"
	"github.com/hajimehoshi/ebiten/v2"
)

// The View menu and its shortcuts: the main viewport's flags (pkg/world
// viewflags.go), which the paint session reads (pkg/paint viewflags.go).
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/ToolbarTop.cpp
// viewMenuMouseDown (0x0043ADF6) / viewMenuDropdown (0x0043AF37),
// src/OpenLoco/src/Input/Shortcuts.cpp toggleUndergroundView … toggleDirArrowsOnTracks.

// viewMenuEntry is one of the menu's entries: its text, the flag it toggles,
// whether the tick shows the flag clear (the two "names displayed" entries
// toggle a hide flag), and its shortcut's default key (Shortcuts.cpp: "1" …
// "0"; none for the names).
type viewMenuEntry struct {
	label    int
	flag     paint.ViewFlags
	inverted bool
	key      ebiten.Key
	sepAfter bool
}

var viewMenuEntries = [...]viewMenuEntry{
	{label: i18n.MenuUndergroundView, flag: paint.ViewUnderground, key: ebiten.KeyDigit1},
	{label: i18n.MenuSeeThroughTracks, flag: paint.ViewSeeThroughTracks, key: ebiten.KeyDigit2},
	{label: i18n.MenuSeeThroughRoads, flag: paint.ViewSeeThroughRoads, key: ebiten.KeyDigit3},
	{label: i18n.MenuSeeThroughTrees, flag: paint.ViewSeeThroughTrees, key: ebiten.KeyDigit4},
	{label: i18n.MenuSeeThroughBuildings, flag: paint.ViewSeeThroughBuildings, key: ebiten.KeyDigit5},
	{label: i18n.MenuSeeThroughScenery, flag: paint.ViewSeeThroughScenery, key: ebiten.KeyDigit7},
	{label: i18n.MenuSeeThroughBridges, flag: paint.ViewSeeThroughBridges, key: ebiten.KeyDigit6, sepAfter: true},
	{label: i18n.MenuHeightMarksOnLand, flag: paint.ViewHeightMarksOnLand, key: ebiten.KeyDigit8},
	{label: i18n.MenuHeightMarksOnTracksRoads, flag: paint.ViewHeightMarksOnTracks, key: ebiten.KeyDigit9},
	{label: i18n.MenuOneWayDirectionArrows, flag: paint.ViewOneWayDirectionArrows, key: ebiten.KeyDigit0, sepAfter: true},
	{label: i18n.MenuTownNamesDisplayed, flag: paint.ViewHideTownNames, inverted: true, key: -1},
	{label: i18n.MenuStationNamesDisplayed, flag: paint.ViewHideStationNames, inverted: true, key: -1},
}

// viewMenu is viewMenuMouseDown: the twelve entries in their three groups,
// each ticked by the main viewport's flag (the names: ticked while not
// hidden), the first highlighted; choosing one is viewMenuDropdown — the
// flag toggled, the view redrawn.
func (g *Game) viewMenu(btn *ui.ToolbarButton) {
	flags := g.w.ViewFlags()
	var items []ui.DropdownItem
	for _, e := range viewMenuEntries {
		items = append(items, ui.DropdownItem{Text: i18n.Label(e.label), Checkmark: true, Selected: flags.Has(e.flag) != e.inverted,
			Action: func() { g.w.ToggleViewFlags(e.flag) }})
		if e.sepAfter {
			items = append(items, ui.DropdownItem{Separator: true})
		}
	}
	d := g.toolbarDropdown(btn, items)
	d.Highlighted = 0
	g.dropdown = d
}

// updateViewShortcuts takes the view toggles' keys (their Shortcuts.cpp
// defaults, 1 … 0). PARTIAL: the keys are fixed (no keyboard shortcuts
// window yet).
func (g *Game) updateViewShortcuts() {
	if b := g.browseWin(); b != nil && b.Save { // the focused text box takes the keys
		return
	}
	for _, e := range viewMenuEntries {
		if e.key >= 0 && g.inKeyJustPressed(e.key) {
			g.w.ToggleViewFlags(e.flag)
		}
	}
}

// syncHeightMarkerOffset keeps Config::heightMarkerOffset by the Options
// window's Regional settings, as its dropdowns do (Options.cpp
// distanceSpeedDropdown / heightsLabelsDropdown: 0 with heights as units,
// else 1 imperial, 2 metric), the screen redrawn on a change.
func (g *Game) syncHeightMarkerOffset() {
	off := 1
	switch {
	case uilayout.HeightUnits:
		off = 0
	case uilayout.Metric:
		off = 2
	}
	if off != paint.HeightMarkerOffset {
		paint.HeightMarkerOffset = off
		g.w.InvalidateView()
	}
}

// consShowView is the construction window opening (Construction's
// createConstructionWindow: Main::showDirectionArrows, showGridlines);
// consResetView its onClose: WindowManager::viewportSetVisibility(reset) —
// the underground view, the see-through modes and the height marks off —
// then hideDirectionArrows and hideGridlines.
// PARTIAL: Config::gridlinesOnLandscape (the Options checkbox) keeping the
// gridlines on.
func (g *Game) consShowView() {
	g.w.SetViewFlags(paint.ViewOneWayDirectionArrows|paint.ViewGridlines, true)
}

func (g *Game) consResetView() {
	g.w.ViewportSetVisibility(world.VisibilityReset)
	g.w.SetViewFlags(paint.ViewOneWayDirectionArrows|paint.ViewGridlines, false)
}

// consOverground is viewportSetVisibility(overgroundView): the construction
// tool taken up again at a new position.
func (g *Game) consOverground() { g.w.ViewportSetVisibility(world.VisibilityOvergroundView) }
