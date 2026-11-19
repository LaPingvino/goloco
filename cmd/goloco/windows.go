package main

import (
	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/shortcuts"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/LaPingvino/goloco/pkg/world"
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
func (g *Game) textInputWin() *ui.TextInputWindow {
	return winOf[*ui.TextInputWindow](g, ui.WindowTypeTextInput, ui.AnyNumber)
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
		g.wins.RepanOpenSound(b.X + createW/2) // createWindow's sound: its origin and size
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
	if g.prompt != nil || (g.dropdown != nil && g.dropdown.Visible) {
		return true
	}
	for _, w := range g.wins.Windows {
		if !isNews(w) && !ui.IsBar(w) { // news comes by itself: the world keeps its pace under it
			return true
		}
	}
	return false
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
	return g.wins.FindAt(x, y) != nil
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
	case *ui.TerraformWindow:
		w.HandleWheel(dy)
	case *ui.MusicSelectionWindow:
		w.HandleWheel(dy)
	case *ui.KeyboardShortcutsWindow:
		w.HandleWheel(dy)
	case *ui.MessagesWindow:
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
		case *ui.MapWindow:
			g.updateMap(w)                         // onUpdate
			w.ScrollDrag(mx, my, left, win == top) // stateScrollLeft: a held press on the map
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
			w.MouseOver(hx, hy) // onScrollMouseOver: the hovered row's vehicle
			w.Drag(mx, my, left)
		case *ui.TerraformWindow:
			w.Update(mx, my, win == top, left, ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight), g.sh)
		case *ui.MusicJukeboxWindow:
			w.MouseOver(hx, hy) // the control button under the mouse
		case *ui.NewsWindow:
			w.MouseOver(hx, hy) // a subject's button under the mouse
		case *ui.MessagesWindow:
			w.MouseOver(hx, hy) // scrollMouseOver: the message under the mouse
			w.Drag(mx, my, left)
		case *ui.KeyboardShortcutsWindow:
			w.MouseOver(hx, hy) // onScrollMouseOver
			w.Drag(mx, my, left)
		case *ui.TextInputWindow: // onUpdate: the caret's blink
			w.Update()
		case *ui.EditShortcutWindow: // onUpdate: the modifiers held
			w.Modifiers = shortcuts.ModifierText(g.kb.modifier &^ shortcuts.ModCheat)
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
	ui.CloseTextInputWithoutCaller(&g.wins) // dispatchUpdateAll: TextInput::sub_4CE6FF after the windows' onUpdate
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
	case *ui.NewsWindow:
		g.fillNews(w)
	case *ui.NewsTicker:
		g.fillTicker(w)
	case *ui.MessagesWindow:
		for i, v := range g.newsSettings {
			w.Settings[i] = int(v)
		}
		w.PlaySounds = g.playNewsSounds
		if skin >= 0 {
			w.SkinImage = skin
		}
	case *ui.MusicJukeboxWindow:
		w.State = g.musicJukeboxState()
	case *ui.TerraformWindow:
		g.prepareTerraformWindow(w)
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

// drawWindows draws the stack, back to front — the toolbar, the panels, the
// news among the others in the list's order — and a dragged car over it
// (DragVehiclePart: stickToFront).
func (g *Game) drawWindows(screen *ebiten.Image) {
	g.wins.SetDropdownActive(g.dropdown != nil && g.dropdown.Visible || g.colourDropdown != nil && g.colourDropdown.Visible)
	g.wins.Draw(screen, g.prepareWindow)
	for _, win := range g.wins.Windows {
		if w, ok := win.(*ui.VehicleWindow); ok {
			w.DrawDrag(screen)
		}
	}
}

// syncBars keeps the game scene's bars on the stack: Gui::init opens
// ToolbarTop, CompanyInfoPanel and TimePanel (in that order) when the scene
// has none — a new scene, its windows closed — and resizeGameScene puts them
// at the screen's edges; the title screen has none.
func (g *Game) syncBars() {
	bars := [3]ui.WindowType{ui.WindowTypeTopToolbar, ui.WindowTypeCompanyInfoPanel, ui.WindowTypeTimePanel}
	if g.inTitleScreen {
		for _, t := range bars {
			g.wins.CloseType(t, ui.AnyNumber)
		}
		return
	}
	if g.wins.Find(ui.WindowTypeTopToolbar, ui.AnyNumber) == nil {
		top := ui.NewBarWindow(ui.WindowTypeTopToolbar, 0, 0, g.sw, 28)
		top.DrawBar = g.drawToolbar
		top.ClickBar = g.toolbarPress
		top.WidgetAt = func(x, y int) bool { return g.toolbar.HandleClick(x, y) >= 0 }
		g.wins.Open(top)
		player := ui.NewBarWindow(ui.WindowTypeCompanyInfoPanel, 0, g.sh-27, uilayout.PlayerPanelW, 27)
		player.DrawBar = g.drawPlayerPanel
		player.ClickBar = g.playerPanelPress
		player.WidgetAt = func(x, y int) bool { return true } // its outer frame: 140 × 29
		g.wins.Open(player)
		tp := ui.NewBarWindow(ui.WindowTypeTimePanel, g.sw-140, g.sh-27, 140, 27)
		tp.DrawBar = g.drawTimePanel
		tp.ClickBar = g.timePanelPress
		tp.WidgetAt = func(x, y int) bool { return true }
		g.wins.Open(tp)
	}
	w := max(g.sw, 640)
	if b := g.wins.Find(ui.WindowTypeTopToolbar, ui.AnyNumber); b != nil {
		b.Base().X, b.Base().Y, b.Base().W = 0, 0, w
	}
	if b := g.wins.Find(ui.WindowTypeCompanyInfoPanel, ui.AnyNumber); b != nil {
		b.Base().X, b.Base().Y = 0, g.sh-b.Base().H
	}
	if b := g.wins.Find(ui.WindowTypeTimePanel, ui.AnyNumber); b != nil {
		b.Base().X, b.Base().Y = w-b.Base().W, g.sh-b.Base().H
	}
}

// openError is Error::open at the mouse, with the error sound panned to the
// window's centre (createErrorWindow).
func (g *Game) openError(title, message string) {
	g.openErrorQuiet(title, message)
	if e := g.errorWin; e != nil {
		g.playSound(audio.SoundError, audio.ChannelUI, e.W/2+e.X)
	}
}

// openErrorQuiet is Error::openQuiet: the error window without its sound.
func (g *Game) openErrorQuiet(title, message string) {
	g.errorWin = ui.OpenError(title, message, g.mouseX, g.mouseY, g.sw, g.sh) // the frame's mouse (a script's, when one plays)
}

// openCmdError is GameCommands::doCommand's failure for the controlling
// company (loc_4314EA): the error window with the title and the command's
// reason (Error::open) — or, when the thing belongs to another company
// (kErrorBelongsToOther), Error::openWithCompetitor: that company's face
// beside "<track / road / station> belongs to <company>", or "Belongs to
// <company>" for a thing without an element.
func (g *Game) openCmdError(title string, c *sim.Cmd) { g.cmdError(title, c, true) }

// openCmdErrorQuiet is openCmdError with the error sound off
// (GameCommands::setErrorSound(false): Error::openQuiet; openWithCompetitor
// sounds regardless).
func (g *Game) openCmdErrorQuiet(title string, c *sim.Cmd) { g.cmdError(title, c, false) }

func (g *Game) cmdError(title string, c *sim.Cmd, sound bool) {
	if c.ErrorText != sim.ErrorBelongsToOther {
		if sound {
			g.openError(title, cmdErrorText(c))
		} else {
			g.openErrorQuiet(title, cmdErrorText(c))
		}
		return
	}
	m, om := &g.w.Map, g.r.ObjMgr
	company := m.CompanyName(int(c.ErrorCompany))
	msg := ""
	switch c.ErrorKind { // the advanced errors: the element's name
	case sim.ErrorKindTrack:
		if om != nil && int(c.ErrorObject) < len(om.TrackObjects) && om.TrackObjects[c.ErrorObject] != nil {
			msg = ui.FormatErrorArgs(i18n.Label(i18n.ErrorReasonStringidBelongsTo), om.TrackObjects[c.ErrorObject].Name, company)
		}
	case sim.ErrorKindRoad:
		if om != nil && int(c.ErrorObject) < len(om.RoadObjects) && om.RoadObjects[c.ErrorObject] != nil {
			msg = ui.FormatErrorArgs(i18n.Label(i18n.ErrorReasonStringidBelongsTo), om.RoadObjects[c.ErrorObject].Name, company)
		}
	case sim.ErrorKindStation:
		if int(c.ErrorObject) < len(m.StationNames) {
			msg = ui.FormatErrorArgs(i18n.Label(i18n.ErrorReasonStringidBelongsTo), m.StationName(int(c.ErrorObject)), company)
		}
	}
	if msg == "" { // ErrorElementKind::none, or no such object: the fallback
		msg = ui.FormatErrorArgs(i18n.Label(i18n.ErrorReasonBelongsTo), company)
	}
	face, colour := g.companyFace(int(c.ErrorCompany))
	if face > 0 {
		face++ // Error::draw: images[ownerEmotion] + 1
	}
	jailed := int(c.ErrorCompany) < len(m.Companies) && m.Companies[c.ErrorCompany].JailStatus != 0
	g.errorWin = ui.OpenErrorWithCompetitor(title, msg, face, colour, jailed, g.mouseX, g.mouseY, g.sw, g.sh)
	if e := g.errorWin; e != nil {
		g.playSound(audio.SoundError, audio.ChannelUI, e.W/2+e.X)
	}
}

// cmdErrorText is the error window's second line: the command's reason
// (setErrorText, with its name argument), "" when it set none.
func cmdErrorText(c *sim.Cmd) string {
	if c.ErrorText == 0 || c.ErrorText == sim.ErrorBelongsToOther {
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

// viewportsAt is what ViewportInteraction reads at screen (x, y):
// WindowManager::findAt's window and its viewports (Window::viewports, [0]
// first) — the main view's when no window is there (the main window is the
// bottom of upstream's stack). win is nil for the main window. The Town,
// Industry, Station and Vehicle windows' viewports are built with the same
// world view their draw uses for that size, so a pick sees what is drawn.
// PARTIAL: the news window's viewports (its subject buttons lie over them)
// and the Company window's headquarters view (not drawn yet).
func (g *Game) viewportsAt(x, y int) (vps []world.Viewport, win ui.Win) {
	win = g.wins.FindAt(x, y)
	if win == nil {
		return []world.Viewport{g.w.MainViewport()}, nil
	}
	vw, ok := win.(ui.ViewportWidgets)
	if !ok {
		return nil, win
	}
	r, ok := vw.ViewportRect()
	if !ok {
		return nil, win
	}
	b := win.Base()
	px, py, pw, ph := r.Min.X, r.Min.Y, r.Dx(), r.Dy()
	var vp world.Viewport
	switch b.WinType {
	case ui.WindowTypeTown:
		vp, ok = g.w.TownViewport(b.Number, px, py, pw, ph)
	case ui.WindowTypeIndustry:
		ok = false
		for i := range g.w.Map.Industries { // the window's number is the industry's id
			if in := &g.w.Map.Industries[i]; int(in.ID) == b.Number && in.Name != 0xFFFF {
				vp, ok = g.w.IndustryViewport(i, px, py, pw, ph)
				break
			}
		}
	case ui.WindowTypeStation:
		vp, ok = g.w.StationViewport(b.Number, px, py, pw, ph)
	case ui.WindowTypeVehicle:
		vp, ok = g.w.VehicleViewport(uint16(b.Number), px, py, pw, ph)
	default:
		ok = false
	}
	if !ok {
		return nil, win
	}
	return []world.Viewport{vp}, win
}

// viewportWidgetAt is findWidgetAt's widget type at (x, y) being
// WidgetType::viewport: the main view where no window is, else the window's
// viewport widget (ViewportWidgets).
func (g *Game) viewportWidgetAt(x, y int) bool {
	win := g.wins.FindAt(x, y)
	if win == nil {
		return true
	}
	vw, ok := win.(ui.ViewportWidgets)
	return ok && vw.ViewportWidgetAt(x, y)
}
