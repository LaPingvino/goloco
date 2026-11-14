package main

import (
	"log"
	"time"

	"github.com/LaPingvino/goloco/pkg/paint"

	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// update.go — Game.Update: one frame of input and simulation. The modal
// windows first, then the title screen or the game (tools, windows, the map
// click, the camera, the game ticks).

func (g *Game) Update() error {
	advanceClock()
	g.windowFrame++
	g.diagTrackShip()
	g.w.TitleMode = g.inTitleScreen // SceneManager::isTitleMode() for the viewport's names
	if fstats.on {
		t0 := time.Now()
		defer func() { fstats.addUpdate(time.Since(t0)) }()
	}
	// The script steps before the modal windows, so a playthrough can type
	// into a text input or answer a prompt (it used to stall behind one).
	if g.script != nil {
		g.script.step()
		g.scriptScenario()
		g.scriptExpect()
		g.scriptCamera()
		g.scriptNews()
		if g.script.quit {
			return ebiten.Termination
		}
	}
	g.mouseX, g.mouseY = ebiten.CursorPosition()
	if g.diagMouse != nil {
		g.mouseX, g.mouseY = g.diagMouse[0], g.diagMouse[1]
	}
	if g.script != nil && g.script.hasMouse {
		g.mouseX, g.mouseY = g.script.mouseX, g.script.mouseY
	}
	g.updateKeyboard() // every key: the windows that take keys, then the shortcuts (keyboard.go)
	if g.updateModal() {
		return nil
	}
	// A right click on the load/save list deletes a file (after asking).
	browse := g.browseWin()
	if browse != nil && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if mx, my := ebiten.CursorPosition(); g.wins.FindAt(mx, my) == ui.Win(browse) {
			browse.HandleRightClick(mx, my)
		}
	}
	if g.quitAfterSave {
		return ebiten.Termination
	}
	if browse != nil { // PromptBrowse: mouse-over row, typed filename, Enter/Escape
		browse.HandleHover(g.mouseX, g.mouseY)
		browse.HandleText(nil, false, false, false) // the caret's blink (its keys: keyboard.go browseKeyUp)
	}

	if g.inTitleScreen {
		g.updateTitle()
		return nil
	}
	g.updateGame()
	return nil
}

// updateModal gives the frame's mouse to an open text input or confirmation,
// and reports that one took it (their keys: keyboard.go callKeyUp).
func (g *Game) updateModal() bool {
	if g.textInput != nil { // TextInputWindow: onMouseUp, onUpdate
		t := g.textInput
		t.Update()
		if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
			t.HandleClick(g.mouseX, g.mouseY)
		}
		if t.Closed && g.textInput == t {
			g.textInput = nil
		}
		return true
	}
	if g.prompt != nil { // the modal confirmation takes all input (setCurrentModalType)
		if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			g.prompt.HandleClick(mx, my)
		}
		if g.prompt.Closed {
			g.prompt = nil
		}
		return true
	}
	return false
}

// updateTitle is a frame of the title screen: the title map ticking under its
// camera tour, the windows opened from the menu, and the menu itself.
func (g *Game) updateTitle() {
	// Title screen: advance camera animation and apply to world.
	// The title map is a running game: w.Update is its game tick.
	g.titleFrame++
	g.w.Map.TitleMode = true
	for range g.w.GameTicks(g.mouseMoved()) { // tickLogic(numUpdates): the title is a running game upstream
		g.w.Update() // tickLogic: one game tick
		if g.titleSeq != nil {
			g.titleSeq.Update()
		}
	}
	if g.titleSeq != nil {
		if rot := g.titleSeq.GetRotation(); rot != g.w.Rotation() { // Title.cpp RotateStep
			g.w.SetRotation(rot)
		}
		cx, cy := g.titleSeq.GetCameraPosition()
		g.w.SetCamera(cx, cy)
	}
	g.updateWindows()
	if g.dropdown != nil && g.dropdown.Visible { // a window's dropdown (Options' language, heights …)
		g.dropdown.UpdateHover(g.mouseX, g.mouseY)
	}
	if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
		// An open dropdown is on top: it takes the click (a click outside
		// closes it); then the windows, front-most first.
		if g.dropdown != nil && g.dropdown.Visible {
			if !g.dropdown.HandleClick(g.mouseX, g.mouseY) {
				g.dropdown = nil
			}
		} else if !g.windowsPress() {
			g.handleTitleMenuClick(g.mouseX, g.mouseY)
		}
	}
	g.routeWheelToWindows()
}

// updateGame is a frame of a running game.
func (g *Game) updateGame() {
	g.syncScreen() // the fixed windows known to the stack before a press reorders it
	g.diagKeepMoneyEffect()
	g.updateConstructionTools()
	g.updateHeld()
	if g.inMouseJustPressed(ebiten.MouseButtonLeft) {
		g.gameClick()
	}
	if g.inMouseJustReleased(ebiten.MouseButtonLeft) && g.dropdown != nil && g.dropdown.Visible {
		// the release of the press that opened a toolbar menu — after the press:
		// a quick click brings both in one frame (ui.DropdownMenu
		// HandleOpeningRelease): on its button, the highlighted item
		if g.dropdown.HandleOpeningRelease(g.mouseX, g.mouseY) {
			g.dropdown = nil
		}
	}
	if g.inMouseJustReleased(ebiten.MouseButtonLeft) && g.colourDropdown != nil && g.colourDropdown.Visible {
		// held from the widget to an item and released there: the item is taken
		if g.colourDropdown.HandleRelease(g.mouseX, g.mouseY) {
			g.colourDropdown = nil
		}
	}
	if g.inMouseJustReleased(ebiten.MouseButtonLeft) {
		for i := range g.toolbar.Buttons {
			g.toolbar.Buttons[i].Pressed = false
		}
		if g.viewportLeft {
			g.viewportLeftRelease()
		}
		if g.stDrag.pressed && g.stationToolOn() {
			g.stationToolRelease(g.mouseX, g.mouseY) // onToolUp: single or the dragged area
		}
		if g.sigToolDown {
			g.sigToolDown = false
			if g.signalToolOn() && !g.overWindow(g.mouseX, g.mouseY) {
				g.signalToolUp(g.mouseX, g.mouseY)
			}
		}
	}

	g.updateCamera()
	g.syncHeightMarkerOffset()
	g.updateMusic()
	// Update world (skipped when paused; repeated per speed multiplier)
	n := g.w.GameTicks(g.mouseMoved())
	g.updateNews(n)  // sub_431695: MessageManager's update and the news windows', per interface update
	if !g.isPaused { // tickLogic(numUpdates × the speed)
		for i := 0; i < n*g.speedMult; i++ {
			g.w.Map.TitleMode = false
			g.w.Update() // tickLogic: one game tick
			if g.inGame {
				g.checkObjective()
			}
		}
	}
	if g.saveMsgFrames > 0 {
		g.saveMsgFrames--
	}
}

// diagKeepMoneyEffect is GOLOCO_DIAG_BUILDVEHICLE=moneyeffect's every frame.
func (g *Game) diagKeepMoneyEffect() {
	if g.diagMoney && len(g.w.Map.MoneyEffects) == 0 { // DIAGNOSTIC ONLY (moneyeffect)
		g.w.Map.TitleMode = false
		if px, py, ok := g.w.SurfaceOrWaterAt(g.sw/2, g.sh/2); ok {
			t := g.w.Map.TileAt(int(px)>>5, int(py)>>5)
			g.w.Map.SpendMoneyEffect(px, py, int32(t.BaseZ)*4+28, g.w.Map.PlayerCompanies[0], -26447)
			if g.titleFrame%300 == 0 {
				log.Printf("[Diag] moneyeffect: at %d,%d: %d effects, lists %v", px, py, len(g.w.Map.MoneyEffects), g.w.Map.EntityLists != nil)
			}
		}
	}
}

// updateConstructionTools moves what the construction window's tools show on
// the map with the cursor — the Station tab's ghost, the hover tile, the
// track ghost, the head's arrow.
func (g *Game) updateConstructionTools() {
	// The station tab's ghost follows the cursor (onToolUpdateTrainStation),
	// and goes once the tab or the mode does.
	if g.diagHoldGhost {
		// DIAGNOSTIC ONLY: the ghost stays for the screenshot
	} else if g.stationToolOn() {
		if g.stDrag.pressed && g.inMousePressed(ebiten.MouseButtonLeft) {
			g.stationToolDrag(g.mouseX, g.mouseY) // onToolDrag
		}
		if !g.stDrag.dragging { // onToolUpdate: no ghost while dragging
			g.stationToolUpdate(g.mouseX, g.mouseY)
		}
	} else if g.stationGhost != nil || g.roadStationGhost != nil || g.dockGhost != nil || g.airportGhost != nil {
		g.removeStationGhost()
	}
	if g.signalToolOn() { // the Signal tab's ghost (onToolUpdate)
		g.signalToolUpdate(g.mouseX, g.mouseY)
	} else if g.w.SignalGhostShown() {
		g.w.RemoveSignalGhost(g.playerCmd())
		g.consSigCost = -1
	}

	// Build mode: update hover tile every frame (track and road; the dock
	// window has no construction head)
	if g.consHasHead() {
		g.hoverTileX, g.hoverTileY = g.tileUnderCursor(g.mouseX, g.mouseY)
	}
	g.updateConstruction()
}

// updateHeld is what follows the mouse without a click: menu and toolbar
// hover, the pickup tool's ghost, and the windows' scrollbars, held buttons,
// dragged cars and moved frames.
func (g *Game) updateHeld() {
	// Update dropdown hover if visible
	if g.dropdown != nil && g.dropdown.Visible && g.dropdown.LeftBehind(g.mouseX, g.mouseY) {
		g.dropdown = nil // a hover-opened menu the mouse left (Dropdown unk1)
		for i := range g.toolbar.Buttons {
			g.toolbar.Buttons[i].Pressed = false
		}
	}
	if g.dropdown != nil && g.dropdown.Visible {
		g.dropdown.UpdateHover(g.mouseX, g.mouseY)
	} else if g.colourDropdown != nil && g.colourDropdown.Visible {
		g.colourDropdown.UpdateHover(g.mouseX, g.mouseY)
	} else {
		g.toolbar.UpdateHover(g.mouseX, g.mouseY)
		// ToolbarTop.cpp onMouseHover: with Config::toolbarAutoMenu (10c1e49) —
		// in the normal input state (no menu open) a button under the mouse
		// opens its menu.
		if btnIdx := g.toolbar.HandleClick(g.mouseX, g.mouseY); btnIdx >= 0 && !g.inTitleScreen && !g.noAutoMenu &&
			g.textInput == nil && g.prompt == nil && !g.inMousePressed(ebiten.MouseButtonLeft) {
			g.toolbar.Buttons[btnIdx].Pressed = true
			g.handleToolbarButton(btnIdx)
			if g.dropdown != nil {
				g.dropdown.HoverOpened()
			}
		}
	}

	if g.windowFrame%60 == 0 { // the open lists repopulated about once a second
		g.refreshList()
	}
	g.updateWindows()
	g.updateTooltip()                           // a window's tooltip under the resting mouse
	g.updateMapTip()                            // processMouseOver's map tooltip and MapToolTip::open
	if g.industryToolOn() && !g.inTitleScreen { // the Fund New Industries tool: its ghost under the cursor
		if g.overWindow(g.mouseX, g.mouseY) {
			g.w.SetMapSelection(false, 0, 0, 0, 0)
			g.removeIndustryGhost()
		} else {
			g.industryToolUpdate(g.mouseX, g.mouseY)
		}
	}
	g.updateTerraform()                 // the Terraform window's tool: selection, ghost, costs, drag
	if g.pickupOn && !g.inTitleScreen { // pickupToolUpdate: the ghost under the cursor
		if g.overWindow(g.mouseX, g.mouseY) {
			g.pickupToolAbort() // no ghost under a window, whatever the vehicle
		} else {
			g.pickupToolUpdate(g.mouseX, g.mouseY)
		}
	}
}

// gameClick sends a left press to what lies under it, topmost first: an open
// dropdown, the windows (the stack, front to back), then the HUD and the map.
func (g *Game) gameClick() {
	// Check dropdown first (it's on top: a window's dropdown lies over the
	// window that opened it)
	if g.dropdown != nil && g.dropdown.Visible {
		if g.dropdown.PressOnOrigin(g.mouseX, g.mouseY) { // its own button: the release chooses
			return
		}
		if !g.dropdown.HandleClick(g.mouseX, g.mouseY) {
			// Click outside dropdown - close it
			g.dropdown = nil
		}
	} else if g.colourDropdown != nil && g.colourDropdown.Visible {
		g.colourDropdown.HandleClick(g.mouseX, g.mouseY) // a colour picked, or closed
		g.colourDropdown = nil
	} else if !g.windowsPress() { // the windows, front-most first (stateNormalLeft)
		g.hudOrMapClick()
	}
}

// hudOrMapClick is a left press no window took: the bottom panels, the
// toolbar, else the active tool on the map.
func (g *Game) hudOrMapClick() {
	if sp := uilayout.TimePanelSpeedAt(g.mouseX, g.mouseY, g.sw, g.sh); sp >= 0 && !g.inTitleScreen {
		// TimePanel speed buttons: pause toggles; the others set the speed.
		// clickable(TimePanel.cpp): pause_btn normal_speed_btn fast_forward_btn extra_fast_forward_btn
		switch sp {
		case 0:
			g.isPaused = !g.isPaused
		case 1:
			g.setSpeed(1)
		case 2:
			g.setSpeed(2)
		case 3:
			g.setSpeed(4)
		}
	} else if (uilayout.InTimePanel(g.mouseX, g.mouseY, g.sw, g.sh) || uilayout.InPlayerPanel(g.mouseX, g.mouseY, g.sw, g.sh)) && !g.inTitleScreen {
		// the bottom-bar windows swallow the click; the time panel's date is
		// MessageWindow::open (TimePanel.cpp onMouseUp date_btn); the player panel's
		// company value is companyValueMouseUp: the Company window's Finances
		// clickable(TimePanel.cpp): date_btn
		tx, ty := uilayout.TimePanelPos(g.sw, g.sh)
		if d := uilayout.TimePanelDate; inRect(g.mouseX, g.mouseY, tx+d[0], ty+d[1], d[2], d[3]) {
			g.openMessagesWindow()
		}
		px, py := uilayout.PlayerPanelPos(g.sw, g.sh)
		if r := uilayout.PlayerPanelFace; uilayout.InPlayerPanel(g.mouseX, g.mouseY, g.sw, g.sh) &&
			inRect(g.mouseX, g.mouseY, px+r[0], py+r[1], r[2], r[3]) {
			g.playerCompanyDropdown(px+r[0], py+r[1], r[2], r[3]) // playerMouseDown
		}
		if r := uilayout.PlayerPanelCompanyVal; uilayout.InPlayerPanel(g.mouseX, g.mouseY, g.sw, g.sh) &&
			inRect(g.mouseX, g.mouseY, px+r[0], py+r[1], r[2], r[3]) {
			company, _ := g.w.ControllingCompany()
			g.openCompanyFinances(company)
		}
	} else if btnIdx := g.toolbar.HandleClick(g.mouseX, g.mouseY); btnIdx >= 0 {
		g.toolbar.Buttons[btnIdx].Pressed = true
		g.handleToolbarButton(btnIdx)
	} else if g.tf.on && !g.inTitleScreen {
		g.terraformToolDown(g.mouseX, g.mouseY) // the Terraform window's tool
	} else if g.industryToolOn() && !g.inTitleScreen {
		g.industryToolDown(g.mouseX, g.mouseY) // the Fund New Industries tool
	} else if g.pickupOn && !g.inTitleScreen {
		g.pickupToolDown(g.mouseX, g.mouseY) // the Vehicle window's placing tool
	} else if g.routeTool && g.vehicleWin(g.routeHead) != nil && !g.inTitleScreen {
		g.routeToolDown(g.mouseX, g.mouseY) // the Route tab's tool: a stop at the station clicked
	} else if g.stationToolOn() {
		g.stationToolDown(g.mouseX, g.mouseY) // the Station tab's tool: onToolDown (the build comes with the release, stationdrag.go)
	} else if g.signalToolOn() {
		g.sigToolDown = true // the Signal tab's tool acts on the release (onToolUp)
	} else if g.consHasHead() && g.consTab == 0 && g.w.Cons.Hover && g.hoverTileX >= 0 && g.r.ObjMgr != nil {
		// onToolUp → onToolUpSingle + constructionLoop: the click starts
		// building — the selected piece at the tile, as the Construct button
		// would; refused, the error window shows and the window hovers on.
		road := g.consIsRoad()
		c := g.playerCmd()
		tx, ty, jh := g.consTarget(road)
		if _, built, refused := g.w.ConstructionToolUp(c, tx, ty, jh, road, uint8(g.buildObjID)); built {
			g.errorWin = nil // the loop's WindowManager::close(WindowType::error)
		} else if refused {
			g.openTrackBuildError(c)
		}
	} else if (g.buildMode == buildModeNone || g.consHasHead() && g.consTab == 0) && !g.inTitleScreen {
		// (while building on from a position the construction window has no
		// tool on the map — toolCancel — so a click is the viewport's.)
		// MouseInput.cpp: a press on the main view without a tool begins
		// State::viewportLeft; its release acts (viewportLeftRelease).
		g.viewportLeft = true
	}
}

// viewportLeftRelease is stateViewportLeft's leftReleased without a tool
// (0x4C73C2): released over the main view (no window there), getItemLeft at
// the release point opens the vehicle's, town's, station's or industry's
// window.
func (g *Game) viewportLeftRelease() {
	g.viewportLeft = false
	if g.overWindow(g.mouseX, g.mouseY) {
		return // released over another window: _dragWindowType differs
	}
	if it, ok := g.w.ItemLeftAt(g.mouseX, g.mouseY); ok {
		switch it.Kind {
		case paint.ItemEntity:
			g.openVehicleWindow(uint16(it.ID), 0)
		case paint.ItemTownLabel:
			g.openTownWindow(it.ID)
		case paint.ItemStationLabel:
			g.openStationWindow(it.ID)
		case paint.ItemIndustry:
			g.openIndustryWindow(it.ID)
		case paint.ItemHeadquarterBuilding:
			g.openCompanyWindow(it.ID)
		}
	}
}

// updateCamera moves the main view by the mouse: right-drag and the wheel
// (unless a window takes it). The keys' and the screen edge's scrolling:
// keyboard.go keyScroll, edgeScroll.
func (g *Game) updateCamera() {
	g.updateViewportRight() // the right button: a drag pans, a click acts (rightclick.go)

	if _, dy := g.inWheel(); dy != 0 {
		if g.inMousePressed(ebiten.MouseButtonRight) {
			// MouseInput.cpp processMouseWheel: with the right button held the
			// wheel turns the view, before any window — right on the zoom-out
			// direction (its wheel > 0: towards the player, dy < 0 here), left
			// on the zoom-in one.
			g.w.RotateView(dy < 0)
		} else if !g.routeWheelToWindows() {
			if dy > 0 {
				g.w.ZoomInAt(g.zoomAnchor())
			} else {
				g.w.ZoomOutAt(g.zoomAnchor())
			}
		}
	}

}

// setSpeed runs the game at mult ticks per tick, unpaused.
func (g *Game) setSpeed(mult int) {
	g.speedMult, g.isPaused = mult, false
}

// diagTrackShip is the shipsail / planefly diagnostics' every frame: the view
// follows diagShip, its state logged.
func (g *Game) diagTrackShip() {
	if g.diagShip != 0 {
		if tr, err := g.w.Map.TrainOf(g.diagShip); err == nil {
			g.diagShipFrames++
			if g.diagFollow && g.diagShipFrames%15 == 0 {
				g.w.CentreOnVehicle(g.diagShip)
			}
			if g.diagShipFrames%120 == 0 {
				h, v2 := g.w.Map.Entities[tr.Head], g.w.Map.Entities[tr.Veh2]
				log.Printf("[Diag] shipsail frame %d: ship at %d,%d yaw %d, %.0f mph, status %d, station %d", g.diagShipFrames, v2.PosX, v2.PosY, v2.SpriteYaw, float64(v2.CurrentSpeed)/65536, h.Status, h.StationID)
			}
		}
	}
}

// zoomAnchor is where the wheel zooms: the cursor (Config::zoomToCursor), else
// the view's centre.
func (g *Game) zoomAnchor() (int, int) {
	if g.noZoomToCursor {
		return g.sw / 2, g.sh / 2
	}
	return g.mouseX, g.mouseY
}
