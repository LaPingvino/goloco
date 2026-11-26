package main

import (
	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/shortcuts"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// construction.go — the ported construction window (pkg/ui ConstructionFrame +
// construction tab) for railway track and roads: the frame and tab states, the
// piece / gradient selection, the track-vs-road dispatch, and the head's ghost.
// The Station tab's map tools are in stationtool.go.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/ConstructionTab.cpp
// The state itself (position, rotation, hovering, piece and gradient) is
// sim.State.Cons (pkg/sim/construction.go).

// consPieceWidget maps upstream's piece widgets to lastSelectedTrackPiece
// (trackPieceWidgets; the dual-track pair and the turnaround go by the
// rotation, constructionFrame), consGradientWidget its gradient widgets to
// lastSelectedTrackGradient.
var (
	consPieceWidget = map[string]uint8{
		"left_hand_curve_very_small": construct.LeftCurveVerySmall, "left_hand_curve_small": construct.LeftCurveSmall,
		"left_hand_curve": construct.LeftCurve, "left_hand_curve_large": construct.LeftCurveLarge,
		"right_hand_curve_very_small": construct.RightCurveVerySmall, "right_hand_curve_small": construct.RightCurveSmall,
		"right_hand_curve": construct.RightCurve, "right_hand_curve_large": construct.RightCurveLarge,
		"straight": construct.Straight, "s_bend_left": construct.SBendLeft, "s_bend_right": construct.SBendRight,
	}
	consGradientWidget = map[string]uint8{
		"steep_slope_down": construct.SteepSlopeDown, "slope_down": construct.SlopeDown, "level": construct.Level,
		"slope_up": construct.SlopeUp, "steep_slope_up": construct.SteepSlopeUp,
	}
)

// consDisabled is the Construction tab's disabledWidgets (setDisabledWidgets,
// ui.ConstructionTabState.Disabled) for a widget by name.
func (g *Game) consDisabled(name string) bool {
	_, st := g.constructionFrame(0, 0)
	return st.Disabled()[name]
}

func nameFor(m map[string]uint8, v uint8) string {
	for k, x := range m {
		if x == v {
			return k
		}
	}
	return ""
}

func (g *Game) constructionFrame(x, y int) (*ui.ConstructionFrame, ui.ConstructionTabState) {
	f := &ui.ConstructionFrame{X: x, Y: y, Title: g.consObjName(), FrameNo: g.titleFrame, Tab: g.consTab}
	switch g.consTab {
	case 1:
		f.H = ui.StationTabHeight(g.stationTabState(0))
	case 2:
		f.W, f.H = ui.SignalTabW, ui.SignalTabH
	case 3:
		f.H = ui.OverheadTabH
	}
	var st ui.ConstructionTabState
	company, _ := g.w.ControllingCompany()
	if g.r.ObjMgr != nil {
		if sk := g.r.ObjMgr.InterfaceSkin; sk != nil {
			f.SkinColour = sk.WindowConstructionColour
		}
		if g.consIsRoad() {
			st.IsRoad, st.TrafficHandedness, f.TabFrames = true, g.w.TrafficHandedness, 32
			if r := g.r.ObjMgr.GetRoadObjectByIndex(g.buildObjID); r != nil {
				f.TrackImage, st.Traits = int(r.Image), r.RoadPieces
			}
		} else if t := g.r.ObjMgr.GetTrackObjectByIndex(g.buildObjID); t != nil {
			f.TrackImage, st.Traits, f.TabFrames = int(t.Image), t.TrackPieces, 16
		}
	}
	if company < len(g.w.CompanyColours) {
		f.CompanyColour = g.w.CompanyColours[company]
	}
	// setDisabledWidgets: no signal (roads have none), no station or no mod
	// available in this year hides that tab.
	if om := g.r.ObjMgr; om != nil {
		year := int(g.w.CurrentYear)
		road := g.consIsRoad()
		f.Disabled[2] = road || len(om.AvailableCompatibleSignals(g.buildObjID, year)) == 0
		f.Disabled[1] = len(om.AvailableCompatibleStations(g.buildObjID, road, year)) == 0
		unlocked := func(v int) bool {
			if company < len(g.w.UnlockedVehicles) && v < 224 {
				return g.w.UnlockedVehicles[company][v/32]&(1<<(v%32)) != 0
			}
			return false
		}
		none := [4]uint8{0xFF, 0xFF, 0xFF, 0xFF}
		if road {
			f.Disabled[3] = om.AvailableCompatibleRoadMods(g.buildObjID, unlocked) == none
		} else {
			f.Disabled[3] = om.AvailableCompatibleMods(g.buildObjID, unlocked) == none
			// drawTrackTabs' content: the selected station type, the selected
			// signal and the available extras (modList).
			tt := &ui.TrackTabIcons{}
			if list := om.AvailableCompatibleStations(g.buildObjID, false, year); len(list) > 0 {
				if st := om.GetTrainStationObjectByIndex(pickSlot(list, g.consStation)); st != nil {
					tt.StationImage, tt.StationRecolourable = int(st.ImageOffset), st.Flags&1 != 0
				}
			}
			if list := om.AvailableCompatibleSignals(g.buildObjID, year); len(list) > 0 {
				if sg := om.GetTrainSignalObjectByIndex(pickSlot(list, g.consSignal)); sg != nil {
					tt.SignalImage, tt.SignalNumFrames, tt.SignalAnimSpeed = int(sg.ImageOffset), int(sg.NumFrames), int(sg.AnimSpeed)
				}
			}
			for _, m := range om.AvailableCompatibleMods(g.buildObjID, unlocked) {
				if m == 0xFF || int(m) >= len(om.TrackExtraOrder) {
					continue
				}
				if ex := om.Extra(om.TrackExtraOrder[m]); ex != nil && ex.Image >= 8 {
					tt.ExtraImages = append(tt.ExtraImages, int(ex.Image)-8) // var_0E: the image table's start
				}
			}
			if g.buildMode == buildModeTrack {
				f.Track = tt
			}
		}
	}
	if g.buildMode == buildModeAirport { // drawAirportTabs; caption title_airport
		g.consTab, f.Tab = 1, 1
		f.Title = i18n.Label(i18n.TitleAirport)
		f.Disabled = [4]bool{true, false, true, true}
		if sk := g.skin(); sk != nil {
			f.StationTabIcon = int(sk.ImageOffset) + objects.ISToolbarMenuAirport
		}
	}
	if g.buildMode == buildModeDock { // drawDockTabs; StationTab::prepareDraw's caption title_ship_port
		g.consTab, f.Tab = 1, 1
		f.Title = i18n.Label(i18n.TitleShipPort)
		f.Disabled = [4]bool{true, false, true, true}
		if sk := g.skin(); sk != nil {
			f.StationTabIcon = int(sk.ImageOffset) + objects.ISToolbarMenuShipPort
		}
	}
	cs := g.w.Cons
	st.Rotation = cs.Rotation
	st.Piece = nameFor(consPieceWidget, cs.Piece)
	if cs.Piece >= construct.SBendToDualTrack { // the pair of buttons whose pieces go by the rotation
		switch {
		case g.consIsRoad():
			st.Piece = "turnaround"
		case consDualTrackPiece(false, cs.Rotation) == cs.Piece:
			st.Piece = "s_bend_dual_track_left"
		default:
			st.Piece = "s_bend_dual_track_right"
		}
	}
	st.Gradient = nameFor(consGradientWidget, cs.Gradient)
	st.Hover = cs.Hover
	st.Tool = g.consTool
	st.GhostBridge = g.w.GhostNeedsBridge()
	st.BridgeImage, st.Company = -1, f.CompanyColour
	// drawCostString's trackCost / roadCost: the ghost's cost (hovering or
	// building), or the piece's just built.
	st.Cost = int64(cs.Cost)
	if g.r.ObjMgr != nil {
		om := g.r.ObjMgr
		// lastSelectedBridge: the first available (PARTIAL: the scenario's remembered bridges[]).
		if list := om.AvailableCompatibleBridges(g.buildObjID, g.consIsRoad(), int(g.w.CurrentYear)); len(list) > 0 {
			if b := om.GetBridgeObjectByIndex(pickSlot(list, g.consBridge)); b != nil {
				st.BridgeImage = int(b.ImageOffset)
			}
		}
	}
	return f, st
}

// consWindowOrigin is where the track / road window opens: Construction::
// openWithFlags creates it with no position, so WindowManager::createWindow
// places it (ui.AutoWindowPosition: (0, 30) under the toolbar on an empty
// screen; before, goloco put it at (10, 40)). An open one keeps its place.
func (g *Game) consWindowOrigin() (int, int) {
	if w := g.consWin(); w != nil {
		return w.X, w.Y
	}
	g.syncScreen()
	win := ui.NewSimpleWindow(ui.WindowTypeConstruction, 0, 0, ui.ConstructionW, ui.ConstructionH)
	g.wins.Place(win)
	return win.X, win.Y
}

func (g *Game) newPortedConstructionWindow() *ui.SimpleWindow {
	// Every construction window opens hovering in rotation 0, straight and
	// level (Common's open, then trackWindow's rotate_90).
	g.closeTerraform() // the construction tool's toolSet
	g.consToolOff = false
	g.consTool, g.bpDrag.pressed, g.bpDrag.dragging = consToolConstruct, false, false
	if g.consWin() != nil {
		g.consResetView() // the window it replaces closes first (its onClose)
	}
	g.consShowView()
	if g.r.ObjMgr != nil {
		g.w.EndConstruction(g.playerCmd())
	}
	g.w.ResetConstruction()
	g.w.Cons.Road = g.consIsRoad()
	g.w.ActivateSelection() // trackWindow's activateSelectedConstructionWidgets
	g.w.SwitchRoadType = g.consSwitchRoadType
	// trackWindow's init: signalPlacementStepSize = kDefaultSignalPlacementStepSize, repeatedSignalMode off.
	g.consSigStep, g.consSigRepeat, g.consSigCost = 4, false, -1
	x, y := g.consWindowOrigin()
	win := ui.NewSimpleWindow(ui.WindowTypeConstruction, x, y, ui.ConstructionW, ui.ConstructionH)
	win.WinFlags = ui.WindowFlagNoAutoClose // createConstructionWindow: lighterFrame | noAutoClose
	// The close button ends the construction (Construction's onClose: the
	// tool cancelled, the ghosts removed) — it used to hide the window and
	// leave the mode on.
	win.OnClose = g.endConstruction
	win.PortedDraw = func(screen *ebiten.Image, x, y int) {
		f, st := g.constructionFrame(x, y)
		if f.Disabled[g.consTab] { // the tab went away (another track type)
			g.consTab = 0
			f, st = g.constructionFrame(x, y)
		}
		win.W, win.H = f.Width(), f.Height()
		f.Draw(screen)
		defer f.DrawCaption(screen) // after the tab (its draw8 rewrites the window's pixels)
		switch g.consTab {
		case 1:
			ui.DrawStationTab(screen, f, g.stationTabState(f.CompanyColour))
			return
		case 2:
			ui.DrawSignalTab(screen, f, g.signalTabState())
			return
		case 3:
			ui.DrawOverheadTab(screen, f, g.overheadTabState())
			return
		}
		// draw: the piece preview in the construct widget, painted as the
		// map is (drawTrackCost / drawRoadCost → drawTrack / drawRoad), in
		// the piece's direction with the arrow, the window's mods.
		if p, ok := g.consPiece(); ok {
			road, obj, mods := g.consIsRoad(), g.buildObjID, g.w.Cons.Mods
			st.Preview = func(rt *gfx8.RenderTarget) {
				g.w.PaintPiecePreview(rt, road, int(p.ID), obj, p.Rotation, mods)
			}
		}
		ui.DrawConstructionTab(screen, f, st)
	}
	// clickable(Construction): close_button tab_construction tab_station tab_signal tab_overhead construct rotate bridge_dropdown station_dropdown signal_dropdown both_directions single_direction checkbox_1 checkbox_2 checkbox_3 checkbox_4 image rotate_90 straight left_hand_curve left_hand_curve_large left_hand_curve_small left_hand_curve_very_small right_hand_curve right_hand_curve_large right_hand_curve_small right_hand_curve_very_small s_bend_left s_bend_right s_bend_dual_track_left s_bend_dual_track_right level slope_up slope_down steep_slope_up steep_slope_down remove
	win.OnWindowClick = func(rx, ry int) {
		f, st := g.constructionFrame(win.X, win.Y)
		// press: the widget's clickDown (MouseInput; ui clicks.go)
		press := win.PressWidget
		if tab := f.TabAt(rx, ry); tab >= 0 || (ry >= 15 && ry < 42) { // tabs
			if tab >= 0 {
				press([4]int{f.TabX()[tab], 15, 31, 27})
			}
			g.consSwitchTab(tab)
			return
		}
		if g.consTab == 3 {
			if i := ui.OverheadHit(rx, ry, g.overheadTabState()); i >= 0 {
				press([4]int{3, 45 + 12*i, 132, 12}) // checkbox_1 …
				g.consMods ^= 1 << i                 // lastSelectedMods ^= (1 << checkbox)
			}
			return
		}
		if g.consTab == 1 && g.buildMode == buildModeAirport && inRect(rx, ry, 112, 104, 24, 24) {
			press([4]int{112, 104, 24, 24})
			g.consStationRotate()
			return
		}
		if g.consTab == 1 && inRect(rx, ry, 3, 45, 132, 12) {
			press(consDropdownWidget([4]int{3, 45, 132, 12}, rx, ry)) // station / station_dropdown
			g.openConsDropdown(win.X, win.Y, f.SkinColour, true)
			return
		}
		if g.consTab == 2 {
			// SignalTab::widgets (kWidth 156, kSpacing 4, kImageButtonSize 40,
			// the stepper 48 wide at kWidth − 52, 182).
			const sw = ui.SignalTabW
			switch hit := ui.SignalHit(rx, ry); hit {
			case "dropdown":
				press(consDropdownWidget([4]int{4, 45, sw - 8, 12}, rx, ry))
				g.openConsDropdown(win.X, win.Y, f.SkinColour, false)
			case "both":
				press([4]int{(sw-8)/2 - 4 - 40, 110, 40, 40})
				g.consSigOne = false
			case "single":
				press([4]int{(sw-8)/2 + 4, 110, 40, 40})
				g.consSigOne = true
			case "repeat": // onMouseUp auto_mode
				press([4]int{4, 166, sw - 8, 12})
				g.consSigRepeat = !g.consSigRepeat
			case "stepdec":
				if g.consSigRepeat { // disabled without repeated placement
					press([4]int{sw - 4 - 26, 183, 13, 10})
					g.consSigStep = max(1, g.consSigStep-1)
				}
			case "stepinc":
				if g.consSigRepeat {
					press([4]int{sw - 4 - 13, 183, 12, 10})
					g.consSigStep = min(64, g.consSigStep+1) // kMaxSignalPlacementStepSize
				}
			}
			return
		}
		hit := ui.ConstructionHit(rx, ry, st)
		if hit == "bridge" {
			press(consDropdownWidget([4]int{40, 123, 58, 20}, rx, ry)) // bridge / bridge_dropdown
			g.openBridgeDropdown(win.X, win.Y, f.SkinColour, f.CompanyColour)
			return
		}
		if r, ok := consTabWidget(hit, st); ok && !g.consDisabled(hit) && !(hit == "construct" && st.PieceUnbuildable) {
			press(r)
		}
		g.consWidget(hit)
	}
	return win
}

// consToolCancelled: the window's tool was cancelled on the tab it shows
// (consToolOff; the Construction tab's tool is never off with the window
// open — cancelled while hovering, the window closes).
func (g *Game) consToolCancelled() bool { return g.consToolOff && g.consTab != 0 }

// consSwitchTab is a click on one of the construction window's tabs
// (Common::switchTab): the tab, its tool on the map again.
func (g *Game) consSwitchTab(tab int) {
	if tab >= 0 && tab <= 3 && g.consTool != consToolConstruct { // switchTab's toolCancel: the copy / paste tool's onToolAbort
		if g.consTool == consToolPaste {
			g.pasteToolAbort()
		}
		g.consTool = consToolConstruct
	}
	switch tab {
	case 1:
		g.consTab = 1 // switchTab; tabReset turns the placing tool on (the Station tab's tool builds, through createTrainStation)
	case 0, 2, 3:
		g.consTab = tab // switchTab; tabReset picks the both-directions tool
		g.consSigOne = false
	default:
		return
	}
	g.consToolOff = false
}

// consStationRotate is the Station tab's rotate button (the airport's):
// constructionRotation++, stationCost = failure.
func (g *Game) consStationRotate() {
	g.airportRotation = (g.airportRotation + 1) & 3
	g.stationGhostCost = 0
}

// consWidget is the Construction tab's onMouseUp / onMouseDown for a widget
// by its name (ui.ConstructionHit's; the keyboard's construction shortcuts
// call the same): a disabled widget does nothing.
// consDropdownWidget is the widget of dropdownWidgets(r) at window point
// (rx, ry): the 11 × 10 button at (x + w − 12, y + 1), else the combo box.
func consDropdownWidget(r [4]int, rx, ry int) [4]int {
	if b := [4]int{r[0] + r[2] - 12, r[1] + 1, 11, 10}; inRect(rx, ry, b[0], b[1], b[2], b[3]) {
		return b
	}
	return r
}

// consTabWidget is the rectangle of ConstructionHit's widget (ConstructionTab
// _widgets: the piece and gradient buttons where activateSelected…Widgets
// leaves them, construct, remove, rotate_90, copy, paste).
func consTabWidget(hit string, st ui.ConstructionTabState) ([4]int, bool) {
	for _, b := range st.Buttons() {
		if b.Name == hit {
			return [4]int{b.X, b.Y, b.W, b.H}, true
		}
	}
	r, ok := map[string][4]int{"construct": {3, 145, 132, 100}, "remove": {72, 248, 39, 24}, "rotate": {111, 248, 24, 24},
		"copy": {3, 248, 33, 24}, "paste": {36, 248, 37, 24}}[hit]
	return r, ok
}

func (g *Game) consWidget(hit string) {
	if g.consDisabled(hit) {
		return // a disabled widget (setDisabledWidgets)
	}
	switch hit {
	case "":
	case "construct":
		g.consConstruct()
	case "rotate": // rotate_90: the rotate arrows while hovering, the red cross while building
		if g.w.Cons.Hover {
			if bp := g.w.Blueprint(); bp != nil { // the blueprint turns with it (rotateBlueprint)
				g.w.RemoveConstructionGhosts(g.playerCmd())
				bp.Rotate(1)
			}
		} else { // back to a new position: the overground view (viewportSetVisibility), the construct tool
			g.consOverground()
			g.consTool = consToolConstruct
		}
		g.w.ConstructionRotate90(g.playerCmd())
	case "remove":
		g.consRemove()
	case "copy":
		g.consCopy()
	case "paste":
		g.consPaste()
	case "s_bend_dual_track_left", "s_bend_dual_track_right", "turnaround":
		piece := uint8(construct.Turnaround) // the road's turnaround button
		if !g.consIsRoad() {
			piece = consDualTrackPiece(hit == "s_bend_dual_track_right", g.w.Cons.Rotation)
		}
		g.w.ChangeTrackPiece(g.playerCmd(), piece, false)
	default:
		// changeTrackPiece: piece and gradient are separate choices; a
		// combination that can't be built gives no piece.
		if v, ok := consPieceWidget[hit]; ok {
			if !g.consIsRoad() || v < construct.SBendLeft { // roads have no s-bends
				g.w.ChangeTrackPiece(g.playerCmd(), v, false)
			}
		} else if v, ok := consGradientWidget[hit]; ok {
			g.w.ChangeTrackPiece(g.playerCmd(), v, true)
		}
	}
}

// signalTabState is the signal tab for the track being built: the selected
// signal is the first available one (lastSelectedSignal; PARTIAL: the
// scenario's remembered choice, Scenario::getConstruction().signals).
func (g *Game) signalTabState() ui.SignalTabState {
	st := ui.SignalTabState{Both: !g.consSigOne, Image: -1, Repeated: g.consSigRepeat, Step: g.consSigStep, Cost: g.consSigCost}
	if g.r.ObjMgr == nil {
		return st
	}
	om := g.r.ObjMgr
	list := om.AvailableCompatibleSignals(g.buildObjID, int(g.w.CurrentYear))
	if len(list) == 0 {
		return st
	}
	if s := om.GetTrainSignalObjectByIndex(pickSlot(list, g.consSignal)); s != nil {
		st.Name, st.Description, st.Image = s.Name, s.Description, int(s.ImageOffset)
	}
	return st
}

// overheadTabState is the overhead tab: the mods available to the company
// (getAvailableCompatibleMods: modList) with their names, and the ticked ones.
func (g *Game) overheadTabState() ui.OverheadTabState {
	st := ui.OverheadTabState{Selected: g.consMods}
	if g.r.ObjMgr == nil {
		return st
	}
	om := g.r.ObjMgr
	company, _ := g.w.ControllingCompany()
	unlocked := func(v int) bool {
		if company < len(g.w.UnlockedVehicles) && v < 224 {
			return g.w.UnlockedVehicles[company][v/32]&(1<<(v%32)) != 0
		}
		return false
	}
	var mods [4]uint8
	order := om.TrackExtraOrder
	if g.consIsRoad() {
		mods, order = om.AvailableCompatibleRoadMods(g.buildObjID, unlocked), om.RoadExtraOrder
	} else {
		mods = om.AvailableCompatibleMods(g.buildObjID, unlocked)
	}
	for i, m := range mods {
		if int(m) < len(order) {
			st.Names[i] = om.ExtraName(order[m])
		}
	}
	return st
}

// stationTabState is the station tab: the first available station type for
// this track or road (lastSelectedStationType; PARTIAL: the scenario's
// remembered choice) and the company's colour.
func (g *Game) stationTabState(company uint8) ui.StationTabState {
	st := ui.StationTabState{Image: -1, Company: company}
	if g.r.ObjMgr == nil {
		return st
	}
	om := g.r.ObjMgr
	if g.buildMode == buildModeAirport { // the airport's name and image, the rotate button
		if g.airportGhost != nil {
			g.ghostTabInfo(&st)
		}
		st.Rotatable = true
		if list := om.AvailableAirports(int(g.w.CurrentYear)); len(list) > 0 {
			if a := om.AirportObjects[pickSlot(list, g.consStation)]; a != nil {
				st.Name, st.Image, st.Recolourable, st.NoOverlay = a.Name, int(a.Image), true, true
			}
		}
		return st
	}
	if g.buildMode == buildModeDock { // the dock's name and image (DockObject image, no windows overlay)
		if g.dockGhost != nil {
			g.ghostTabInfo(&st)
		}
		if list := om.AvailableDocks(int(g.w.CurrentYear)); len(list) > 0 {
			if d := om.DockObjects[pickSlot(list, g.consStation)]; d != nil {
				st.Name, st.Image, st.Recolourable, st.NoOverlay = d.Name, int(d.Image), true, true
			}
		}
		return st
	}
	road := g.consIsRoad()
	list := om.AvailableCompatibleStations(g.buildObjID, road, int(g.w.CurrentYear))
	if len(list) == 0 {
		return st
	}
	slot := pickSlot(list, g.consStation)
	if (g.stationGhost != nil && !road) || (g.roadStationGhost != nil && road) {
		g.ghostTabInfo(&st)
	}
	if road {
		if s := om.GetRoadStationObjectByIndex(slot); s != nil {
			st.Name, st.Image, st.Recolourable = s.Name, int(s.ImageOffset), s.Flags&1 != 0
		}
	} else if s := om.GetTrainStationObjectByIndex(slot); s != nil {
		st.Name, st.Image, st.Recolourable = s.Name, int(s.ImageOffset), s.Flags&1 != 0
	}
	return st
}

// pickSlot is the selected slot if it is in the list, else the list's first.
func pickSlot(list []int, sel int) int {
	for _, s := range list {
		if s == sel {
			return s
		}
	}
	return list[0]
}

// openConsDropdown is the station or signal tab's onMouseDown on its
// dropdown: Dropdown::show at the combobox (3, 45; width() + 2 = 133,
// height() = 11) in the window's secondary colour, an item per available
// type, the selected one highlighted; choosing sets lastSelected….
// PARTIAL: Scenario::getConstruction()'s remembered choice isn't written.
func (g *Game) openConsDropdown(wx, wy int, colour uint8, station bool) {
	if g.r.ObjMgr == nil {
		return
	}
	om := g.r.ObjMgr
	year := int(g.w.CurrentYear)
	var list []int
	name := func(int) string { return "" }
	sel := &g.consSignal
	if station && g.buildMode == buildModeAirport { // AddStationsToDropdown<AirportObject>
		list, sel = om.AvailableAirports(year), &g.consStation
		name = func(s int) string {
			if s < len(om.AirportObjects) && om.AirportObjects[s] != nil {
				return om.AirportObjects[s].Name
			}
			return ""
		}
	} else if station && g.buildMode == buildModeDock { // AddStationsToDropdown<DockObject>
		list, sel = om.AvailableDocks(year), &g.consStation
		name = func(s int) string {
			if s < len(om.DockObjects) && om.DockObjects[s] != nil {
				return om.DockObjects[s].Name
			}
			return ""
		}
	} else if station {
		road := g.consIsRoad()
		list, sel = om.AvailableCompatibleStations(g.buildObjID, road, year), &g.consStation
		name = func(s int) string {
			if road {
				if o := om.GetRoadStationObjectByIndex(s); o != nil {
					return o.Name
				}
			} else if o := om.GetTrainStationObjectByIndex(s); o != nil {
				return o.Name
			}
			return ""
		}
	} else {
		list = om.AvailableCompatibleSignals(g.buildObjID, year)
		name = func(s int) string {
			if o := om.GetTrainSignalObjectByIndex(s); o != nil {
				return o.Name
			}
			return ""
		}
	}
	if len(list) == 0 {
		return
	}
	cur := pickSlot(list, *sel)
	items := make([]ui.DropdownItem, len(list))
	hl := -1
	for i, s := range list {
		items[i] = ui.DropdownItem{Text: name(s)}
		if s == cur {
			hl = i
		}
	}
	d := ui.NewDropdownShow(wx+3, wy+45, 133, 11, colour, items, g.sw, g.sh)
	d.Highlighted = hl
	d.OnSelect = func(i int) {
		if i >= 0 && i < len(list) {
			*sel = list[i]
			g.consRemember(station, list[i])
		}
	}
	g.dropdown = d
}

// consRemember is the Station and Signal tabs' onDropdown beyond
// lastSelected…: the choice remembered — the airport's and dock's in the
// game state (lastAirport, lastShipPort), the others per track / road object
// in Scenario::getConstruction (roadStations, trainStations, signals).
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/StationTab.cpp
// onDropdown, SignalTab.cpp onDropdown
func (g *Game) consRemember(station bool, slot int) {
	m, obj, v := &g.w.Map, g.buildObjID&7, uint8(slot)
	switch {
	case !station:
		m.Construction.Signals[obj] = v
	case g.buildMode == buildModeAirport:
		m.LastAirport = v
	case g.buildMode == buildModeDock:
		m.LastShipPort = v
	case g.consIsRoad():
		m.Construction.RoadStations[obj] = v
	default:
		m.Construction.TrainStations[obj] = v
	}
}

// consRecall is createTrackConstructionWindow / createRoadConstructionWindow's
// (and openAtTrack / openAtRoad's) choices for the object: the signal (track
// only), station and bridge Scenario::getConstruction remembers for it, else
// (0xFF) the first available (-1 here). Bridges is one list for track and
// road objects alike. PARTIAL: the remembered mods (trackMods / roadMods,
// initialiseDefaultTrackRoadMods) — consMods keeps the window's own.
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/Common.cpp
func (g *Game) consRecall(road bool, obj int) {
	c := &g.w.Map.Construction
	slot := func(v uint8) int {
		if v == 0xFF {
			return -1
		}
		return int(v)
	}
	obj &= 7
	if road {
		g.consSignal = -1 // lastSelectedSignal = 0xFF
		g.consStation = slot(c.RoadStations[obj])
	} else {
		g.consSignal = slot(c.Signals[obj])
		g.consStation = slot(c.TrainStations[obj])
	}
	g.consBridge = slot(c.Bridges[obj])
}

// openBridgeDropdown is bridgeDropdown (ConstructionTab.cpp): Dropdown::show
// at the bridge box (40, 123), 155 wide, items 22 high, in the window's
// colour: each bridge's image (company colour) with "Max. speed:" and
// "Height limit:" (dropdown_bridge_stats), the current one highlighted.
// Units as the Regional tab's Distance and Speed sets them (uilayout).
func (g *Game) openBridgeDropdown(wx, wy int, colour, company uint8) {
	if g.r.ObjMgr == nil {
		return
	}
	om := g.r.ObjMgr
	list := om.AvailableCompatibleBridges(g.buildObjID, g.consIsRoad(), int(g.w.CurrentYear))
	if len(list) == 0 {
		return
	}
	cur := pickSlot(list, g.consBridge)
	items := make([]ui.DropdownItem, len(list))
	hl := -1
	for i, s := range list {
		b := om.GetBridgeObjectByIndex(s)
		if b == nil {
			continue
		}
		speed := "Unlimited"
		if b.MaxSpeed != 0xFFFF { // kSpeed16Null
			speed = uilayout.Velocity(int(b.MaxSpeed))
		}
		items[i] = ui.DropdownItem{Sprite: int(b.ImageOffset), SpriteColour: company,
			Lines: []string{"Max. speed: " + speed, "Height limit: " + uilayout.Height(int(b.MaxHeight))}}
		if s == cur {
			hl = i
		}
	}
	d := ui.NewDropdownShowItemHeight(wx+40, wy+123, 155, 19, colour, items, 22, g.sw, g.sh)
	d.Highlighted = hl
	d.OnSelect = func(i int) { // onDropdown: the bridge, the ghosts removed
		if i >= 0 && i < len(list) {
			g.consBridge = list[i]
			g.w.Map.Construction.Bridges[g.buildObjID&7] = uint8(list[i]) // the TODO's & ~(1 << 7): road objects share it
			g.w.ChangeBridge(g.playerCmd(), uint8(list[i]))
		}
	}
	g.dropdown = d
}

// updateConstruction is the Construction tab's part of each update for a
// track or road window: the window's bridge and mods into the state; while
// hovering, the tool's update with the cursor over the map (onToolUpdate →
// onToolUpdateSingle: the ghost and arrow under the cursor); while building,
// updateConstruction (the ghost and the blinking arrow where the next piece
// starts). Off the tab or without the window: the ghosts and the arrow go.
func (g *Game) updateConstruction() {
	if g.r.ObjMgr == nil {
		return
	}
	if !g.consHasHead() || g.consTab != 0 || g.inTitleScreen {
		if g.stationToolSelects() { // the airport / dock tool's selection and arrow (stationtool.go)
			g.w.SyncConstructionArrow(true)
			return
		}
		if g.w.TrackGhost != nil || g.w.RoadGhost != nil || g.w.BlueprintGhostShown() || g.w.ConsArrow.On || g.w.ConsSelection.On {
			g.w.EndConstruction(g.playerCmd())
		}
		g.w.SyncConstructionArrow(false)
		return
	}
	road, obj := g.consIsRoad(), uint8(g.buildObjID)
	g.w.Cons.Road = road
	// cState.lastSelectedBridge, bridgeList and lastSelectedMods: the
	// window's choices, for every piece built
	g.w.Cons.Bridge, g.w.Cons.Mods, g.w.Cons.BridgeList = 0, g.consMods, [9]uint8{}
	if list := g.r.ObjMgr.AvailableCompatibleBridges(g.buildObjID, road, int(g.w.CurrentYear)); len(list) > 0 {
		g.w.Cons.Bridge = uint8(pickSlot(list, g.consBridge)) + 1
		for i, b := range list[:min(len(list), 8)] { // copyToLegacyList
			g.w.Cons.BridgeList[i] = uint8(b) + 1
		}
	}
	// The ghost may choose the bridge (placeTrackGhost's retry, sub_4A193B):
	// lastSelectedBridge is the state's, the window shows it.
	defer func() {
		if g.w.Cons.Bridge != 0 {
			g.consBridge = int(g.w.Cons.Bridge) - 1
		}
	}()
	c := g.playerCmd()
	if g.bpToolOn() { // the copy / paste tool's update and drag (blueprint.go)
		if g.bpDrag.pressed && g.inMousePressed(ebiten.MouseButtonLeft) {
			g.bpToolDrag(g.mouseX, g.mouseY)
		} else if !g.overWindow(g.mouseX, g.mouseY) {
			g.bpToolUpdate(g.mouseX, g.mouseY)
		}
		g.w.SyncConstructionArrow(true)
		return
	}
	if !g.w.Cons.Hover {
		g.w.UpdateConstruction(c, road, obj)
	} else if g.hoverTileX >= 0 && !g.overWindow(g.mouseX, g.mouseY) {
		tx, ty, jh := g.consTarget(road)
		g.w.ConstructionToolUpdate(c, tx, ty, jh, road, obj, g.hasKeyModifier(shortcuts.ModShift))
	}
	g.w.SyncConstructionArrow(true)
}

// consSwitchRoadType is Common::sub_4A3A50 (the road type constructRoad's
// retry switched to, sim.State.SwitchRoadType): the window's road object, the
// toolbar's last road (setTrackOptions), its station and bridge — the
// scenario's remembered ones, else the first available (consRecall) — and
// its mods (PARTIAL: roadMods not remembered, so none); into the state at
// once, for the retry.
func (g *Game) consSwitchRoadType(obj uint8) {
	g.buildObjID = int(obj)
	if g.r.ObjMgr == nil {
		return
	}
	g.w.Map.SetTrackOptions(obj|1<<7, g.r.ObjMgr)
	g.consRecall(true, int(obj))
	g.consSignal, g.consMods = -1, 0
	g.w.Cons.Bridge, g.w.Cons.Mods = 0, 0
	if list := g.r.ObjMgr.AvailableCompatibleBridges(int(obj), true, int(g.w.CurrentYear)); len(list) > 0 {
		g.w.Cons.Bridge = uint8(pickSlot(list, g.consBridge)) + 1
	}
}

// endConstruction is the track / road window closing (Common::onClose): the
// tool cancelled, the ghosts and the arrow removed, the mode off.
func (g *Game) endConstruction() {
	if g.r.ObjMgr != nil {
		g.w.EndConstruction(g.playerCmd())
	}
	g.w.ResetConstruction()
	g.w.SyncConstructionArrow(false)
	g.consResetView()
	g.buildMode = buildModeNone
	g.hoverTileX, g.hoverTileY = -1, -1
}

// consPiece is getTrackPieceId / getRoadPieceId (pkg/construct) for the
// selected piece and gradient in the construction rotation; ok false when
// that can't be built.
func (g *Game) consPiece() (construct.TrackPieceID, bool) {
	cs := g.w.Cons
	if g.consIsRoad() {
		return construct.GetRoadPieceID(cs.Piece, cs.Gradient, cs.Rotation)
	}
	return construct.GetTrackPieceID(cs.Piece, cs.Gradient, cs.Rotation)
}

// consDualTrackPiece is the piece the left or right dual-track button
// chooses at this heading (construct.DualTrackButtons).
func consDualTrackPiece(right bool, rot uint8) uint8 {
	_, _, l, r := construct.DualTrackButtons(rot)
	if right {
		return r
	}
	return l
}

func (g *Game) consIsRoad() bool { return g.buildMode == buildModeRoad }

// consHasHead: the track and road windows build from a construction head (the
// dock and airport windows have none).
func (g *Game) consHasHead() bool {
	return g.buildMode == buildModeTrack || g.buildMode == buildModeRoad
}

// stationToolSelects: the Station tab's tool is the airport's or the dock's,
// whose tool update owns the map selection and the construction arrow
// (onToolUpdateAirport / onToolUpdateDock); a train or road station's
// doesn't touch them.
func (g *Game) stationToolSelects() bool {
	return g.stationToolOn() && (g.buildMode == buildModeAirport || g.buildMode == buildModeDock)
}

// stationToolOn: a construction window is on its Station tab (every mode has
// one), so the map takes the station tool.
func (g *Game) stationToolOn() bool {
	return g.buildMode != buildModeNone && g.consTab == 1 && !g.inTitleScreen && !g.consToolCancelled()
}

// consSelectedID is the selected piece's track or road id (consPiece), -1 none.
func (g *Game) consSelectedID() int {
	if p, ok := g.consPiece(); ok {
		return int(p.ID)
	}
	return -1
}

// openTrackBuildError is doCommand's failure for createTrack / createRoad:
// Error::open(cant_build_pop3_string with the track's or road's name, the
// command's reason).
func (g *Game) openTrackBuildError(c *sim.Cmd) {
	g.openCmdError(g.trackBuildErrorTitle(), c)
}

// openTrackBuildErrorQuiet is openTrackBuildError with the error sound off
// (GameCommands::setErrorSound(false): Error::openQuiet).
func (g *Game) openTrackBuildErrorQuiet(c *sim.Cmd) {
	g.openCmdErrorQuiet(g.trackBuildErrorTitle(), c)
}

// trackBuildErrorTitle is cant_build_pop3_string with the track's or road's name.
func (g *Game) trackBuildErrorTitle() string {
	title := i18n.Label(i18n.CantBuildPop3String)
	id := g.buildObjID
	if g.consIsRoad() {
		if id >= 0 && id < len(g.r.ObjMgr.RoadObjects) && g.r.ObjMgr.RoadObjects[id] != nil {
			title = ui.FormatErrorArgs(title, g.r.ObjMgr.RoadObjects[id].Name)
		}
	} else if id >= 0 && id < len(g.r.ObjMgr.TrackObjects) && g.r.ObjMgr.TrackObjects[id] != nil {
		title = ui.FormatErrorArgs(title, g.r.ObjMgr.TrackObjects[id].Name)
	}
	return title
}

// consConstruct is the Construct button — the piece preview —
// (constructTrackOrRoad): the selected piece built where the state stands,
// the state moved past it; refused, the error window. Upstream doesn't ask
// whether the window hovers: a click on the preview then tries the piece
// where the ghost's position is, 16 below the ground ("Raise or lower land
// first", checked on OpenLoco 26.08), and hovers on.
func (g *Game) consConstruct() {
	if g.r.ObjMgr == nil {
		return
	}
	if _, ok := g.consPiece(); !ok { // constructTrack: no piece, nothing
		return
	}
	c := g.playerCmd()
	var built bool
	if g.consIsRoad() {
		_, built = g.w.ConstructRoad(c, uint8(g.buildObjID))
	} else {
		_, built = g.w.ConstructTrack(c, uint8(g.buildObjID))
	}
	if !built {
		g.openTrackBuildError(c)
	}
}

// consRemove is the bulldozer button (removeTrack 0x004A0121): the piece
// behind the state taken up, the state moved back over it; refused, the
// error window (upstream's title is cant_build_pop3_string here too).
func (g *Game) consRemove() {
	if g.r.ObjMgr == nil {
		return
	}
	c, g1 := g.stationCmd()
	var refused bool
	if g.consIsRoad() {
		_, refused = g.w.ConstructionRemoveRoad(c, uint8(g.buildObjID), g1)
	} else {
		_, refused = g.w.ConstructionRemoveTrack(c, uint8(g.buildObjID), g1)
	}
	if refused {
		g.openTrackBuildError(c)
	}
}

// consObjectNames returns the object display names for the current mode,
// indexed by object slot (empty string for unused slots).
func (g *Game) consObjectNames() []string {
	var names []string
	if g.r.ObjMgr == nil {
		return names
	}
	if g.consIsRoad() {
		for _, o := range g.r.ObjMgr.RoadObjects {
			if o != nil {
				names = append(names, o.Name)
			} else {
				names = append(names, "")
			}
		}
	} else {
		for _, o := range g.r.ObjMgr.TrackObjects {
			if o != nil {
				names = append(names, o.Name)
			} else {
				names = append(names, "")
			}
		}
	}
	return names
}

// consObjName is the current object slot's display name, or "(none)".
func (g *Game) consObjName() string {
	names := g.consObjectNames()
	if g.buildObjID >= 0 && g.buildObjID < len(names) && names[g.buildObjID] != "" {
		return names[g.buildObjID]
	}
	return "(none)"
}

// consCycleType advances g.buildObjID to the next populated object slot.
func (g *Game) consCycleType(dir int) {
	names := g.consObjectNames()
	n := len(names)
	if n == 0 {
		return
	}
	for k := 0; k < n; k++ {
		g.buildObjID = ((g.buildObjID+dir)%n + n) % n
		if names[g.buildObjID] != "" {
			return
		}
	}
}

// consTarget is where the construction tool points: on existing track (or
// road) under the cursor, that tile at that piece's height
// (tryMakeTrackJunctionAtLoc / tryMakeRoadJunctionAtLoc: a junction); else
// the hovered tile on the ground (junction -1).
func (g *Game) consTarget(road bool) (tx, ty int, junction int32) {
	if jx, jy, h, ok := g.w.JunctionAt(g.mouseX, g.mouseY, road); ok {
		return jx, jy, h
	}
	return g.hoverTileX, g.hoverTileY, -1
}
