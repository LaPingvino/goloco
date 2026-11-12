package main

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// windows.go — what the open windows share: whether any is open or under the
// cursor, the mouse wheel's routing, and the error window.

// uiOpen reports a window, menu or prompt open: the world may then be drawn
// less often so the UI stays responsive (LOOP item 0).
func (g *Game) uiOpen() bool {
	return g.prompt != nil || g.options != nil || g.townList != nil || g.browse != nil || g.scenarioSelect != nil ||
		(g.dropdown != nil && g.dropdown.Visible) || (g.windowMgr != nil && len(g.windowMgr.Windows) > 0)
}

// overWindow reports whether a screen point is on one of goloco's windows or
// bars, where the tool does not act (the viewport is under none of them).
func (g *Game) overWindow(x, y int) bool {
	return g.vehicleWin != nil && g.vehicleWin.Contains(x, y) ||
		g.buildVehicle != nil && g.buildVehicle.Contains(x, y) ||
		g.companyFinances != nil && g.companyFinances.Contains(x, y) ||
		g.townList != nil && g.townList.Contains(x, y) ||
		g.options != nil && g.options.Contains(x, y) ||
		g.toolbar.HandleClick(x, y) >= 0 ||
		uilayout.InTimePanel(x, y, g.sw, g.sh) || uilayout.InPlayerPanel(x, y, g.sw, g.sh)
}

// routeWheelToWindows sends mouse-wheel events to any visible window under the
// cursor that has an OnScrollWheel handler.  Returns true if consumed.
func (g *Game) routeWheelToWindows() bool {
	_, dy := g.inWheel()
	if dy == 0 {
		return false
	}
	if g.scenarioSelect != nil && g.scenarioSelect.Contains(g.mouseX, g.mouseY) {
		g.scenarioSelect.HandleWheel(dy)
		return true
	}
	if g.browse != nil && g.browse.Contains(g.mouseX, g.mouseY) {
		g.browse.HandleWheel(dy)
		return true
	}
	if g.townList != nil && g.townList.Contains(g.mouseX, g.mouseY) {
		g.townList.HandleWheel(dy)
		return true
	}
	for i := len(g.windowMgr.Windows) - 1; i >= 0; i-- {
		w := g.windowMgr.Windows[i]
		if w.Visible && w.OnScrollWheel != nil &&
			g.mouseX >= w.X && g.mouseX < w.X+w.Width &&
			g.mouseY >= w.Y && g.mouseY < w.Y+w.Height {
			w.OnScrollWheel(dy)
			return true
		}
	}
	if g.buildVehicle != nil && g.buildVehicle.HandleWheel(g.mouseX, g.mouseY, dy) {
		return true
	}
	return false
}

// openError is Error::open at the mouse.
func (g *Game) openError(title, message string) {
	mx, my := ebiten.CursorPosition()
	g.errorWin = ui.OpenError(title, message, mx, my, g.sw, g.sh)
}

// cmdErrorText is the error window's second line: the command's reason
// (setErrorText, with its name argument), "" when it set none.
func cmdErrorText(c *sim.Cmd) string {
	if c.ErrorText == 0 {
		return ""
	}
	msg := i18n.Label(c.ErrorText)
	if c.ErrorArg != "" {
		msg = ui.FormatErrorArgs(msg, c.ErrorArg)
	}
	return msg
}
