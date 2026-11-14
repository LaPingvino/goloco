package main

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/shortcuts"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// shortcut_actions.go — what each keyboard shortcut does (the keys that run
// them: pkg/shortcuts, keyboard.go normalKey), and the cheat words.
//
// OpenLoco reference: src/OpenLoco/src/Input/Shortcuts.cpp (closeTopmostWindow
// 0x004BF089 … openDebugWindow; buildTracksBuildRoadsCommon),
// src/OpenLoco/src/Input/Keyboard.cpp (loc_4BECDE, loc_4BED04, loc_4BED79),
// src/OpenLoco/src/Ui/WindowManager.cpp (closeTopmost, closeAllFloatingWindows),
// src/OpenLoco/src/Ui/ToolManager.cpp (toolCancel).
//
// PARTIAL: no editor mode (the shortcuts it turns off); showMap (no map
// window), sendMessage (no network game), openDebugWindow (no debug window)
// do nothing; rotateConstructionObject's town list (the editor's build
// tabs) and company window (the headquarters ghost) parts.

// shortcutActions is ShortcutManager's list of actions, by shortcut.
var shortcutActions = [shortcuts.Count]func(g *Game){
	shortcuts.CloseTopmostWindow:      (*Game).closeTopmostWindow,
	shortcuts.CloseAllFloatingWindows: (*Game).closeAllFloatingWindows,
	shortcuts.CancelConstructionMode:  (*Game).cancelConstructionMode,
	shortcuts.PauseUnpauseGame:        func(g *Game) { g.isPaused = !g.isPaused }, // GameCommands::PauseGame
	// zoomViewOut / zoomViewIn: viewportZoomOut / In(false) — about the
	// view's centre, not the cursor.
	shortcuts.ZoomViewOut:                    func(g *Game) { g.w.ZoomOut() },
	shortcuts.ZoomViewIn:                     func(g *Game) { g.w.ZoomIn() },
	shortcuts.RotateView:                     func(g *Game) { g.w.RotateView(true) }, // viewportRotateRight
	shortcuts.RotateConstructionObject:       (*Game).rotateConstructionObject,
	shortcuts.ToggleUndergroundView:          toggleView(paint.ViewUnderground),
	shortcuts.ToggleSeeThroughTracks:         toggleView(paint.ViewSeeThroughTracks),
	shortcuts.ToggleSeeThroughRoads:          toggleView(paint.ViewSeeThroughRoads),
	shortcuts.ToggleSeeThroughTrees:          toggleView(paint.ViewSeeThroughTrees),
	shortcuts.ToggleSeeThroughBuildings:      toggleView(paint.ViewSeeThroughBuildings),
	shortcuts.ToggleSeeThroughScenery:        toggleView(paint.ViewSeeThroughScenery),
	shortcuts.ToggleSeeThroughBridges:        toggleView(paint.ViewSeeThroughBridges),
	shortcuts.ToggleHeightMarksOnLand:        toggleView(paint.ViewHeightMarksOnLand),
	shortcuts.ToggleHeightMarksOnTracks:      toggleView(paint.ViewHeightMarksOnTracks),
	shortcuts.ToggleDirArrowsonTracks:        toggleView(paint.ViewOneWayDirectionArrows),
	shortcuts.AdjustLand:                     func(g *Game) { g.openTerraform(ui.TerraformTabLand) },  // Terraform::openAdjustLand
	shortcuts.AdjustWater:                    func(g *Game) { g.openTerraform(ui.TerraformTabWater) }, // openAdjustWater
	shortcuts.PlantTrees:                     func(g *Game) { g.openTerraform(ui.TerraformTabTrees) }, // openPlantTrees
	shortcuts.BulldozeArea:                   func(g *Game) { g.openTerraform(ui.TerraformTabClear) }, // openClearArea
	shortcuts.BuildTracks:                    func(g *Game) { g.buildTracksOrRoads(true) },
	shortcuts.BuildRoads:                     func(g *Game) { g.buildTracksOrRoads(false) },
	shortcuts.BuildAirports:                  func(g *Game) { g.buildPort(0) },
	shortcuts.BuildShipPorts:                 func(g *Game) { g.buildPort(1) },
	shortcuts.BuildNewVehicles:               (*Game).buildNewVehicles,
	shortcuts.ShowVehiclesList:               func(g *Game) { g.openVehicleList(int(g.w.Map.LastVehicleType)) }, // VehicleList::open(controlling, lastVehicleType)
	shortcuts.ShowStationsList:               (*Game).openStationsWindow,                                        // StationList::open(controlling, 0)
	shortcuts.ShowTownsList:                  (*Game).openTownsWindow,
	shortcuts.ShowIndustriesList:             (*Game).openIndustryList,
	shortcuts.ShowCompaniesList:              (*Game).openCompanyList,
	shortcuts.ShowMap:                        func(g *Game) { log.Println("[Game] showMap: the map window is not ported yet") },
	shortcuts.ShowCompanyInformation:         func(g *Game) { c, _ := g.w.ControllingCompany(); g.openCompanyWindow(c) },
	shortcuts.ShowFinances:                   func(g *Game) { c, _ := g.w.ControllingCompany(); g.openCompanyFinances(c) },
	shortcuts.ShowAnnouncementsList:          (*Game).openMessagesWindow,
	shortcuts.ShowOptionsWindow:              (*Game).openOptionsWindow,
	shortcuts.ShowJukeboxWindow:              (*Game).openMusicJukebox,
	shortcuts.Screenshot:                     func(g *Game) { g.triggerScreenshotCountdown(2) }, // makeScreenshot (0x004BF3AB)
	shortcuts.ToggleLastAnnouncement:         (*Game).toggleLastAnnouncement,
	shortcuts.SendMessage:                    func(g *Game) {}, // "if (!SceneManager::isNetworked()) return"
	shortcuts.ConstructionPreviousTab:        func(g *Game) { g.consStepTab(-1) },
	shortcuts.ConstructionNextTab:            func(g *Game) { g.consStepTab(1) },
	shortcuts.ConstructionPreviousTrackPiece: func(g *Game) { g.consStepWidget(false, -1) },
	shortcuts.ConstructionNextTrackPiece:     func(g *Game) { g.consStepWidget(false, 1) },
	shortcuts.ConstructionPreviousSlope:      func(g *Game) { g.consStepWidget(true, -1) },
	shortcuts.ConstructionNextSlope:          func(g *Game) { g.consStepWidget(true, 1) },
	shortcuts.ConstructionBuildAtCurrentPos:  (*Game).consBuildAtCurrentPos,
	shortcuts.ConstructionRemoveAtCurrentPos: (*Game).consRemoveAtCurrentPos,
	shortcuts.ConstructionSelectPosition:     (*Game).consSelectPosition,
	// GameCommands::SetGameSpeed: Normal, FastForward, ExtraFastForward — the
	// time panel's buttons.
	shortcuts.GameSpeedNormal:           func(g *Game) { g.setSpeed(1) },
	shortcuts.GameSpeedFastForward:      func(g *Game) { g.setSpeed(2) },
	shortcuts.GameSpeedExtraFastForward: func(g *Game) { g.setSpeed(4) },
	shortcuts.OpenDebugWindow:           func(g *Game) { log.Println("[Game] openDebugWindow: the debug window is not ported") },
}

// executeShortcut is ShortcutManager::execute.
func (g *Game) executeShortcut(id shortcuts.Shortcut) {
	if id < shortcuts.Count && shortcutActions[id] != nil {
		shortcutActions[id](g)
	}
}

// toggleView is toggleUndergroundView … toggleDirArrowsOnTracks: the main
// viewport's flag turned, the view redrawn.
func toggleView(flag paint.ViewFlags) func(g *Game) {
	return func(g *Game) { g.w.ToggleViewFlags(flag) }
}

// closeDropdown is WindowManager::close(WindowType::dropdown).
func (g *Game) closeDropdown() {
	if g.dropdown == nil && g.colourDropdown == nil {
		return
	}
	g.dropdown, g.colourDropdown = nil, nil
	for i := range g.toolbar.Buttons {
		g.toolbar.Buttons[i].Pressed = false
	}
}

// topmostFloating is the front-most window that is neither stickToBack nor
// stickToFront (nil: none).
func (g *Game) topmostFloating() ui.Win {
	for i := len(g.wins.Windows) - 1; i >= 0; i-- {
		if w := g.wins.Windows[i]; !w.Base().HasFlags(ui.WindowFlagStickToBack | ui.WindowFlagStickToFront) {
			return w
		}
	}
	return nil
}

// closeTopmostWindow is WindowManager::closeTopmost (0x004BF089): an open
// dropdown goes, and the front-most window that doesn't stick to the back or
// the front (the news, the error window, a text input stay).
func (g *Game) closeTopmostWindow() {
	g.closeDropdown()
	if w := g.topmostFloating(); w != nil {
		g.wins.Close(w)
	}
}

// closeAllFloatingWindows is WindowManager::closeAllFloatingWindows
// (0x004BF0B6): the same for every such window.
func (g *Game) closeAllFloatingWindows() {
	g.closeDropdown()
	for n := len(g.wins.Windows); n >= 0; n-- { // each Close takes one away (a window's onClose may take more)
		w := g.topmostFloating()
		if w == nil {
			return
		}
		g.wins.Close(w)
	}
}

// cancelConstructionMode is 0x004BF0BC: an error window closes; else the
// Vehicle window's tool goes by a tab change (cancelVehicleTools); else the
// tool on the map is cancelled.
func (g *Game) cancelConstructionMode() {
	if g.errorWin != nil {
		g.errorWin = nil
		return
	}
	if !g.cancelVehicleTools() {
		g.toolCancel()
	}
}

// cancelVehicleTools is Vehicle::cancelVehicleTools (0x004B944A): with a
// Vehicle window's tool active, that window on its Main tab goes to Details
// and on its Route tab to Main (the tab change cancels the tool); on another
// tab nothing here.
func (g *Game) cancelVehicleTools() bool {
	head, ok := g.vehicleToolHead()
	if !ok {
		return false
	}
	w := g.vehicleWin(head)
	if w == nil || w.OnTab == nil {
		return false
	}
	switch w.Tab {
	case 0:
		w.OnTab(1)
		return true
	case 4:
		w.OnTab(0)
		return true
	}
	return false
}

// vehicleToolHead is ToolManager::isToolActive(WindowType::vehicle) with
// getToolWindowNumber: the vehicle whose window has the tool.
func (g *Game) vehicleToolHead() (uint16, bool) {
	switch {
	case g.pickupOn:
		return g.pickupHead, true
	case g.routeTool:
		return g.routeHead, true
	}
	return 0, false
}

// toolCancel is ToolManager::toolCancel for whichever window has the tool:
// its onToolAbort, and what its onUpdate does once the tool is gone.
//   - the Vehicle window's pickup tool: the ghost removed, the window redrawn;
//     its Route tool: off;
//   - the Fund New Industries tool: its ghost removed;
//   - the Terraform window closes (Terraform's Common::onUpdate);
//   - the construction window, hovering on its Construction tab, closes
//     (Construction::onUpdate: constructionHover without the tool); building
//     on from a position it has no tool, and stays; on its other tabs it
//     stays without its tool, the ghosts removed (Common::onUpdate).
func (g *Game) toolCancel() {
	switch {
	case g.pickupOn:
		head := g.pickupHead
		g.pickupToolAbort()
		g.pickupOn = false
		g.refreshVehicleWindow(head, 1)
	case g.routeTool:
		g.routeTool = false
	case g.industryToolOn():
		g.industryToolAbort()
	case g.tf.on:
		g.closeTerraform()
	case g.consToolOn():
		if g.consHasHead() && g.consTab == 0 {
			g.wins.CloseType(ui.WindowTypeConstruction, ui.AnyNumber)
			return
		}
		g.consToolOff = true
		g.removeStationGhost()
		g.w.SetMapSelection(false, 0, 0, 0, 0)
	}
}

// rotateConstructionObject is 0x004BF148: the first of these that can turn
// something — the Vehicle window's placing tool (Vehicle::rotate 0x004B949C:
// the direction a vehicle is placed in), the Terraform window's tree
// (Terraform::rotate), the construction window (Construction::rotate
// 0x004A5D48: hovering on the Construction tab, rotate_90 and the ghosts
// removed; on the Station tab, its rotate button when it shows — the
// airport's).
// PARTIAL: TownList::rotate (the editor's build tabs), CompanyWindow::rotate
// (the headquarters ghost).
func (g *Game) rotateConstructionObject() {
	if g.pickupOn {
		g.w.Map.PickupDirection ^= 1
		return
	}
	if w := g.terraformWin(); w != nil && w.Rotate() {
		return
	}
	if g.consWin() != nil && g.buildMode != buildModeNone {
		switch {
		case g.consHasHead() && g.consTab == 0 && g.w.Cons.Hover && g.r.ObjMgr != nil:
			g.consWidget("rotate")
			g.w.RemoveConstructionGhosts(g.playerCmd())
		case g.consTab == 1 && g.buildMode == buildModeAirport:
			g.consStationRotate()
		}
	}
}

// buildTracksOrRoads is buildTracks (0x004BF232) / buildRoads (0x004BF24F)
// through buildTracksBuildRoadsCommon: the construction window for the
// default railroad or road type; when it is open on that very type already,
// for the next of the company's available types, round again (upstream
// 98de125).
func (g *Game) buildTracksOrRoads(rail bool) {
	om, m := g.r.ObjMgr, &g.w.Map
	if om == nil {
		return
	}
	ctrl, _ := g.w.ControllingCompany()
	id, available := m.LastRoadOption, m.CompanyAvailableRoads(uint8(ctrl), om)
	if rail {
		id, available = m.LastRailroadOption, m.CompanyAvailableRailTracks(uint8(ctrl), om)
	}
	if id == 0xFF || len(available) == 0 {
		return
	}
	if cur, ok := g.consCurrentTrackType(); ok && cur == id {
		next := available[0]
		for i, t := range available {
			if t == id && i+1 < len(available) {
				next = available[i+1]
			}
		}
		id = next
	}
	g.openConstructionWithFlags(id)
}

// consCurrentTrackType is Construction::getCurrentTrackType with the window
// open: cState.trackType — the track object, or the road object with bit 7.
// The airport and dock windows have none of either (upstream's trackType is
// then the flags' low byte, 0).
func (g *Game) consCurrentTrackType() (uint8, bool) {
	if g.consWin() == nil {
		return 0, false
	}
	switch g.buildMode {
	case buildModeTrack:
		return uint8(g.buildObjID), true
	case buildModeRoad:
		return uint8(g.buildObjID) | 1<<7, true
	case buildModeAirport, buildModeDock:
		return 0, true
	}
	return 0, false
}

// buildPort is buildAirports (0x004BF276, option 0) / buildShipPorts
// (0x004BF295, option 1): the airport or dock construction window, when the
// game has an airport or a dock to build.
func (g *Game) buildPort(option int) {
	m := &g.w.Map
	m.UpdateAvailableAirportAndDockOptions(g.r.ObjMgr) // as the toolbar's port menu does before it reads them
	if option == 0 && m.LastAirport == 0xFF || option == 1 && m.LastShipPort == 0xFF {
		return
	}
	g.openPortConstruction(option)
}

// buildNewVehicles is 0x004BF2B4: BuildVehicle::openByType(defaultBuildVehicleType).
func (g *Game) buildNewVehicles() {
	if t := g.w.Map.LastBuildVehiclesOption; t != 0xFF {
		g.openBuildVehicle(int(t))
	}
}

// toggleLastAnnouncement is 0x004BF3B3: the news window closed, or — none
// open — the last message shown again (NewsWindow::openLastMessage
// 0x0042AC27).
func (g *Game) toggleLastAnnouncement() {
	if g.newsWin() != nil {
		g.closeNews()
		return
	}
	m := &g.w.Map
	m.CloseNewsWindow = g.closeNews
	last := len(m.Messages) - 1
	g.closeNews()
	if m.SelectMessage(uint16(max(last, 0))); last >= 0 {
		g.openNews(uint16(last))
	}
}

// consShortcutWin is the construction shortcuts' WindowManager::find(
// WindowType::construction): the window is open.
func (g *Game) consShortcutWin() bool {
	return g.consWin() != nil && g.buildMode != buildModeNone
}

// consStepTab is Common::previousTab / nextTab: the tab before or after the
// current one among those the window shows, round again.
func (g *Game) consStepTab(dir int) {
	w := g.consWin()
	if !g.consShortcutWin() {
		return
	}
	f, _ := g.constructionFrame(w.X, w.Y)
	for i, tab := 0, g.consTab; i < 4; i++ {
		tab = (tab + dir + 4) % 4
		if !f.Disabled[tab] {
			if tab != g.consTab {
				g.consSwitchTab(tab)
			}
			return
		}
	}
}

// consStepWidget is previousTrackPiece / nextTrackPiece (the piece widgets,
// left_hand_curve_very_small … s_bend_dual_track_right) and previousSlope /
// nextSlope (steep_slope_down … steep_slope_up): from the activated widget
// of the range, the one before or after it that is shown and enabled, round
// again (Window::prevAvailableWidgetInRange / nextAvailableWidgetInRange) —
// pressed as the mouse would.
func (g *Game) consStepWidget(gradient bool, dir int) {
	w := g.consWin()
	if !g.consShortcutWin() || !g.consHasHead() || g.consTab != 0 {
		return
	}
	_, st := g.constructionFrame(w.X, w.Y)
	active := st.Piece
	if gradient {
		active = st.Gradient
	}
	var names []string
	at := -1
	for _, b := range st.Buttons() {
		if _, isGradient := consGradientWidget[b.Name]; isGradient != gradient {
			continue
		}
		if b.Name == active {
			at = len(names)
		} else if !g.consIsRoad() && consUnusedAtRotation(b.Name, g.w.Cons.Rotation) {
			continue
		}
		names = append(names, b.Name)
	}
	if at < 0 || len(names) < 2 { // firstActivatedWidgetInRange: none
		return
	}
	g.consWidget(names[(at+dir+len(names))%len(names)])
}

// consBuildAtCurrentPos is Construction::buildAtCurrentPos: the Construct
// button, while building on from a position.
func (g *Game) consBuildAtCurrentPos() {
	if g.consShortcutWin() && g.consHasHead() && g.consTab == 0 && !g.w.Cons.Hover {
		g.consWidget("construct")
	}
}

// consRemoveAtCurrentPos is Construction::removeAtCurrentPos: the remove
// button.
func (g *Game) consRemoveAtCurrentPos() {
	if g.consShortcutWin() && g.consHasHead() && g.consTab == 0 {
		g.consWidget("remove")
	}
}

// consSelectPosition is Construction::selectPosition: rotate_90 while
// building on — back to choosing a position.
func (g *Game) consSelectPosition() {
	if g.consShortcutWin() && g.consHasHead() && g.consTab == 0 && !g.w.Cons.Hover {
		g.consWidget("rotate")
	}
}

// The cheat words (Keyboard.cpp kCheats). PARTIAL: their click sound
// (Audio::playSound clickPress).

// cheatDriver is DRIVER (loc_4BECDE): SceneManager's driverCheatEnabled —
// the Vehicle window's manual control of trains.
func (g *Game) cheatDriver() { g.driverCheat = true }

// cheatShunt is SHUNT (loc_4BED04): with DRIVER on, the front-most Vehicle
// window of one of the controlling company's trains gets the shunt cheat
// (VehicleApplyShuntCheat).
func (g *Game) cheatShunt() {
	if !g.driverCheat {
		return
	}
	m := &g.w.Map
	ctrl, _ := g.w.ControllingCompany()
	for i := len(g.wins.Windows) - 1; i >= 0; i-- {
		w, ok := g.wins.Windows[i].(*ui.VehicleWindow)
		if !ok {
			continue
		}
		head := uint16(w.Base().Number)
		tr, err := m.TrainOf(head)
		if err != nil {
			continue
		}
		if h := &m.Entities[tr.Head]; int(h.Owner) != ctrl || h.Mode != 0 { // TransportMode::rail
			continue
		}
		m.VehicleShuntCheat(g.playerCmd(), sim.CmdApply, head)
		return
	}
}

// cheatFreeCash is FREECASH (loc_4BED79): ApplyFreeCashCheat.
func (g *Game) cheatFreeCash() {
	g.w.Map.FreeCashCheat(g.playerCmd(), sim.CmdApply)
}
