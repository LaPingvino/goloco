package main

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"os"
	"strings"
)

// windows.go — the window stack (ui.WindowStack, g.wins) as the game uses it:
// finding a window by its type and number, sending it the mouse, the per-frame
// work of each window type, and the error window.

// winOf is WindowManager::find(type, number) as the window's own type (nil
// when there is none).
func winOf[T ui.Win](g *Game, typ ui.WindowType, number int) T {
	w, _ := g.wins.Find(typ, number).(T)
	return w
}

// The windows there is one of.
func (g *Game) optionsWin() *ui.OptionsWindow {
	return winOf[*ui.OptionsWindow](g, ui.WindowTypeOptions, ui.AnyNumber)
}
func (g *Game) browseWin() *ui.PromptBrowseWindow {
	return winOf[*ui.PromptBrowseWindow](g, ui.WindowTypeFileBrowserPrompt, ui.AnyNumber)
}
func (g *Game) scenarioWin() *ui.ScenarioSelectWindow {
	return winOf[*ui.ScenarioSelectWindow](g, ui.WindowTypeScenarioSelect, ui.AnyNumber)
}
func (g *Game) buildVehicleWin() *ui.BuildVehicleWindow {
	return winOf[*ui.BuildVehicleWindow](g, ui.WindowTypeBuildVehicle, ui.AnyNumber)
}
func (g *Game) companyListWin() *ui.CompanyListWindow {
	return winOf[*ui.CompanyListWindow](g, ui.WindowTypeCompanyList, ui.AnyNumber)
}
func (g *Game) jukeboxWin() *ui.MusicJukeboxWindow {
	return winOf[*ui.MusicJukeboxWindow](g, ui.WindowTypeMusicJukebox, ui.AnyNumber)
}
func (g *Game) consWin() *ui.SimpleWindow {
	return winOf[*ui.SimpleWindow](g, ui.WindowTypeConstruction, ui.AnyNumber)
}

// The windows there is one of per company, vehicle, station, town, industry.
func (g *Game) companyWin(company int) *ui.CompanyFinancesWindow {
	return winOf[*ui.CompanyFinancesWindow](g, ui.WindowTypeCompany, company)
}
func (g *Game) vehicleWin(head uint16) *ui.VehicleWindow {
	return winOf[*ui.VehicleWindow](g, ui.WindowTypeVehicle, int(head))
}
func (g *Game) stationWin(id int) *ui.StationWindow {
	return winOf[*ui.StationWindow](g, ui.WindowTypeStation, id)
}

// showAuto is a window's open() with WindowManager::createWindow(type, size,
// flags, events): one of its type and number being open, that one comes to
// the front (replaced by w); else w opens where upstream places a window of
// createW × createH (0: w's own size) — the size its create() gives it — and
// then takes its tab's size (onResize; switchTab's moveInsideScreenEdges).
func (g *Game) showAuto(w ui.Win, createW, createH int) {
	b := w.Base()
	if g.wins.Find(b.WinType, b.Number) != nil {
		g.wins.Show(w)
		return
	}
	g.syncScreen()
	bw, bh := b.W, b.H
	if createW > 0 {
		b.W, b.H = createW, createH
	}
	g.wins.Place(w)
	b.W, b.H = bw, bh
	b.ApplyLimits()
	g.wins.Show(w)
	if createW > 0 {
		g.wins.MoveInsideScreenEdges(w)
	}
}

// showCentred is the same for createWindowCentred.
func (g *Game) showCentred(w ui.Win) {
	b := w.Base()
	if g.wins.Find(b.WinType, b.Number) == nil {
		g.syncScreen()
		g.wins.PlaceCentred(w)
	}
	g.wins.Show(w)
}

// showList is TownList::open / IndustryList::open: the list at its own
// origin (the screen's right, y 30), the windows under it moved down
// (moveOtherWindowsDown).
func (g *Game) showList(w ui.Win) {
	if g.wins.Show(w) {
		g.syncScreen()
		g.wins.MoveOtherWindowsDown(w)
	}
}

// frontWin is the front-most window of a type (nil: none open).
func frontWin[T ui.Win](g *Game) T {
	var zero T
	for i := len(g.wins.Windows) - 1; i >= 0; i-- {
		if w, ok := g.wins.Windows[i].(T); ok {
			return w
		}
	}
	return zero
}

// frontVehicleWin is the front-most Vehicle window (the diagnostics', which
// open one and work on it).
func (g *Game) frontVehicleWin() *ui.VehicleWindow { return frontWin[*ui.VehicleWindow](g) }

// frontListWin is the front-most list window (TownList, IndustryList,
// StationList, VehicleList: one Go type).
func (g *Game) frontListWin() *ui.TownListWindow { return frontWin[*ui.TownListWindow](g) }

// uiOpen reports a window, menu or prompt open: the world may then be drawn
// less often so the UI stays responsive (LOOP item 0).
func (g *Game) uiOpen() bool {
	return g.prompt != nil || (g.dropdown != nil && g.dropdown.Visible) || len(g.wins.Windows) > 0
}

// windowColours are a window's title bar and body colours from the interface
// skin (0 and 1 without one): the body in the player-window colour for a
// company's windows (vehicles, stations lists, finances …), else the plain
// window colour. An owned window's title bar then takes its company's colour.
func (g *Game) windowColours(player bool) (tc, wc uint8) {
	sk := g.skin()
	if sk == nil {
		return 0, 1
	}
	if player {
		return sk.WindowTitlebarColour, sk.WindowPlayerColor
	}
	return sk.WindowTitlebarColour, sk.WindowColour
}

// overWindow reports whether a screen point is on one of goloco's windows or
// bars, where the tool does not act (the viewport is under none of them).
func (g *Game) overWindow(x, y int) bool {
	return g.wins.FindAt(x, y) != nil ||
		g.toolbar.HandleClick(x, y) >= 0 ||
		uilayout.InTimePanel(x, y, g.sw, g.sh) || uilayout.InPlayerPanel(x, y, g.sw, g.sh)
}

// routeWheelToWindows is WindowManager::wheelInput for the window under the
// cursor: its list scrolls. Returns true if a window is there (the main view
// then does not zoom). PARTIAL: upstream scrolls only a scroll view with a
// visible bar, else any window that has one.
func (g *Game) routeWheelToWindows() bool {
	_, dy := g.inWheel()
	if dy == 0 {
		return false
	}
	switch w := g.wins.FindAt(g.mouseX, g.mouseY).(type) {
	case nil:
		return false
	case *ui.ScenarioSelectWindow:
		w.HandleWheel(dy)
	case *ui.PromptBrowseWindow:
		w.HandleWheel(dy)
	case *ui.StationWindow: // a vehicle tab's list
		w.HandleWheel(dy)
	case *ui.TownListWindow:
		w.HandleWheel(dy)
	case *ui.SimpleWindow:
		if w.OnScrollWheel != nil {
			w.OnScrollWheel(dy)
		}
	case *ui.BuildVehicleWindow:
		w.HandleWheel(g.mouseX, g.mouseY, dy)
	case *ui.MusicSelectionWindow:
		w.HandleWheel(dy)
	}
	return true
}

// farAway is a screen point on no window: what a window that is not under
// the mouse is given as the mouse's place.
const farAway = -1 << 20

// updateWindows is the windows' share of a frame without a click: the window
// being moved or resized (statePositioningWindow, stateResizing), then each
// window's own — the row under the mouse (the window under it only: findAt),
// held scrollbars and buttons, a dragged car.
func (g *Game) updateWindows() {
	left := g.inMousePressed(ebiten.MouseButtonLeft)
	mx, my := g.mouseX, g.mouseY
	if g.wins.Dragging() {
		if left {
			g.wins.Held(mx, my)
		} else {
			g.wins.Release(mx, my)
		}
	}
	var top ui.Win
	if !(g.dropdown != nil && g.dropdown.Visible) && !g.wins.Dragging() {
		top = g.wins.FindAt(mx, my)
	}
	g.updateRouteTool()
	for _, win := range append([]ui.Win(nil), g.wins.Windows...) {
		hx, hy := farAway, farAway
		if win == top {
			hx, hy = mx, my
		}
		switch w := win.(type) {
		case *ui.TownListWindow:
			w.MouseOver(hx, hy)  // onScrollMouseOver: the row (or grid cell) under the mouse
			w.Drag(mx, my, left) // the list's scrollbar thumb follows the mouse while held
		case *ui.StationWindow:
			w.MouseOver(hx, hy) // a vehicle tab's row under the mouse
			w.Drag(mx, my, left)
		case *ui.PromptBrowseWindow:
			w.Drag(mx, my, left)
		case *ui.ScenarioSelectWindow:
			w.Drag(mx, my, left)
			w.MouseOver(hx, hy)
		case *ui.CompanyListWindow:
			w.ScrollDrag(mx, my, left)
			w.MouseOver(hx, hy)
			w.Update(mx, my, win == top, ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)) // findAt: the window under the mouse
		case *ui.BuildVehicleWindow:
			w.Drag(mx, my, left)
		case *ui.MusicJukeboxWindow:
			w.MouseOver(hx, hy) // the control button under the mouse
		case *ui.MusicSelectionWindow:
			if g.jukeboxWin() == nil { // onUpdate: it goes with the jukebox window
				w.Close()
				break
			}
			w.MouseOver(hx, hy)
			w.Drag(mx, my, left)
		case *ui.CompanyFinancesWindow:
			w.UpdateHeld(mx, my, left)
			w.UpdateDrag(mx, my, left)
		case *ui.VehicleWindow:
			w.ScrollDrag(mx, my, left)
			w.SpeedHeld(mx, my, left)
			if os.Getenv("GOLOCO_DIAG_DRAG_HOLD") != "" {
				break
			}
			if w.Dragging { // DragVehiclePart: follows the mouse, dropped on release (the diag's hold: neither)
				w.MouseOver(mx, my)
				w.DragTo(mx, my)
				if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
					w.DragEnd(mx, my)
				}
			} else {
				w.MouseOver(hx, hy)
			}
		}
	}
}

// windowsPress is a left press for the windows (stateNormalLeft): the one
// under it comes to the front and takes it. It reports that one did.
func (g *Game) windowsPress() bool { return g.wins.Press(g.mouseX, g.mouseY) }

// prepareWindow is what a window is given before it draws: the frame
// counter, the skin, a company's figures as they change.
func (g *Game) prepareWindow(win ui.Win) {
	skin := -1
	if sk := g.skin(); sk != nil {
		skin = int(sk.ImageOffset)
	}
	switch w := win.(type) {
	case *ui.MusicJukeboxWindow:
		w.State = g.musicJukeboxState()
	case *ui.OptionsWindow:
		w.TitleMusic, w.GameMusic = g.playTitleMusic, g.audioCfg.PlayJukeboxMusic // the Audio menu changes them too
	case *ui.VehicleWindow:
		if tr, err := g.w.Map.TrainOf(uint16(w.Number)); err == nil { // the owner's colour, as each draw sets it
			if o := int(g.w.Map.Entities[tr.Head].Owner); o < len(g.w.CompanyColours) {
				w.TitleColour = g.w.CompanyColours[o]
			}
		}
		w.FrameNo = g.windowFrame
		if skin >= 0 {
			w.SkinImage = skin
		}
	case *ui.StationWindow:
		w.FrameNo = g.windowFrame
	case *ui.CompanyListWindow:
		w.FrameNo = g.windowFrame
		w.CurrentMonths = int(g.w.Map.CurrentYear)*12 + int(g.w.Map.CurrentMonth)
		if skin >= 0 {
			w.SkinImage = skin
		}
	case *ui.CompanyFinancesWindow:
		id := w.CompanyID
		w.Data = g.companyFinancesData(id) // the figures as they change
		w.Title = g.w.Map.CompanyName(id)  // the caption's company->name
		if id >= 0 && id < len(g.w.CompanyColours) {
			w.TitleColour = g.w.CompanyColours[id] // WindowManager's draw: an owned window's primary is its company's colour
		}
		w.CargoLines = g.companyCargoLines(id)
		w.Challenge = g.companyChallenge(id)
		w.Status = g.companyStatus(id)
		w.Details = g.companyDetails(id)
		w.Colours = g.companyColours(id)
		w.FrameNo = g.windowFrame
		if skin >= 0 {
			w.SkinImage = skin
		}
	}
}

// drawWindows draws the stack, back to front, and a dragged car over it
// (DragVehiclePart: stickToFront).
func (g *Game) drawWindows(screen *ebiten.Image) {
	g.wins.Draw(screen, g.prepareWindow)
	for _, win := range g.wins.Windows {
		if w, ok := win.(*ui.VehicleWindow); ok {
			w.DrawDrag(screen)
		}
	}
}

// fixedWindows are upstream's stickToFront windows goloco keeps outside the
// stack — ToolbarTop (0, 0, width × 28), PlayerInfoPanel and TimePanel (140 ×
// 27 in the bottom corners), in Gui::init's order — as a new window's
// placement sees them; none on the title screen (its windows are centred).
func (g *Game) fixedWindows() []ui.WinRect {
	if g.inTitleScreen {
		return nil
	}
	return []ui.WinRect{{X: 0, Y: 0, W: g.sw, H: 28}, {X: 0, Y: g.sh - 27, W: 140, H: 27}, {X: g.sw - 140, Y: g.sh - 27, W: 140, H: 27}}
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
	if strings.Contains(msg, "{CURRENCY32}") { // FormatArguments::common(payment)
		msg = strings.Replace(msg, "{CURRENCY32}", uilayout.Currency48(int64(c.ErrorCost)), 1)
	}
	return msg
}

// inRect reports the point (px, py) inside the w×h rectangle at (x, y).
func inRect(px, py, x, y, w, h int) bool {
	return px >= x && px < x+w && py >= y && py < y+h
}
