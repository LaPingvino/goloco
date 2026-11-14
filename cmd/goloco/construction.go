package main

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
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

// consUnusedAtRotation is getUnusedPiecesRotation (0x0049DB71): the piece
// and gradient widgets a track's construction rotation disables — beside a
// dual track's other track (4–7, 8–11) the curves but the very small ones,
// one s-bend and the slopes; on the diagonal (12–15) every curve but the
// large ones, the s-bends and every gradient.
func consUnusedAtRotation(widget string, rotation uint8) bool {
	in := func(names ...string) bool {
		for _, n := range names {
			if n == widget {
				return true
			}
		}
		return false
	}
	switch {
	case rotation >= 12:
		return in("left_hand_curve_very_small", "left_hand_curve_small", "left_hand_curve", "right_hand_curve",
			"right_hand_curve_very_small", "right_hand_curve_small", "s_bend_dual_track_left", "s_bend_left",
			"s_bend_right", "s_bend_dual_track_right", "steep_slope_down", "slope_down", "slope_up", "steep_slope_up")
	case rotation >= 4:
		sBend := "s_bend_left"
		if rotation >= 8 {
			sBend = "s_bend_right"
		}
		return in("left_hand_curve_small", "left_hand_curve", "left_hand_curve_large", "right_hand_curve_large",
			"right_hand_curve", "right_hand_curve_small", sBend, "slope_down", "slope_up")
	}
	return false
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
		f.H = ui.SignalTabH
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
	if g.consWin() != nil {
		g.consResetView() // the window it replaces closes first (its onClose)
	}
	g.consShowView()
	if g.r.ObjMgr != nil {
		g.w.EndConstruction(g.playerCmd())
	}
	g.w.ResetConstruction()
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
		win.H = f.Height()
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
		ui.DrawConstructionTab(screen, f, st)
		// The piece preview in the construct area (goloco's preview drawing;
		// PARTIAL: upstream draws it with the paint code, drawTrack).
		if clip, ok := screen.SubImage(image.Rect(x+4, y+146, x+3+132-1, y+145+100-1)).(*ebiten.Image); ok {
			id, rot := -1, int(g.w.Cons.Rotation)&3
			if p, ok := g.consPiece(); ok {
				id, rot = int(p.ID), int(p.Rotation)&3
			}
			// drawTrack / drawRoad paint with options.rotation = getCurrentRotation():
			// the piece shows as the view shows it.
			g.w.DrawPiecePreview(clip, float64(x+3+66), float64(y+145+28), 1.0, g.consIsRoad(), id, g.buildObjID, (rot+int(g.w.Rotation()))&3)
		}
	}
	// clickable(Construction): close_button tab_construction tab_station tab_signal tab_overhead construct rotate bridge_dropdown station_dropdown signal_dropdown both_directions single_direction checkbox_1 checkbox_2 checkbox_3 checkbox_4 image rotate_90 straight left_hand_curve left_hand_curve_large left_hand_curve_small left_hand_curve_very_small right_hand_curve right_hand_curve_large right_hand_curve_small right_hand_curve_very_small s_bend_left s_bend_right s_bend_dual_track_left s_bend_dual_track_right level slope_up slope_down steep_slope_up steep_slope_down remove
	win.OnWindowClick = func(rx, ry int) {
		f, st := g.constructionFrame(win.X, win.Y)
		if tab := f.TabAt(rx, ry); tab >= 0 || (ry >= 15 && ry < 42) { // tabs
			g.consSwitchTab(tab)
			return
		}
		if g.consTab == 3 {
			if i := ui.OverheadHit(rx, ry, g.overheadTabState()); i >= 0 {
				g.consMods ^= 1 << i // lastSelectedMods ^= (1 << checkbox)
			}
			return
		}
		if g.consTab == 1 && g.buildMode == buildModeAirport && inRect(rx, ry, 112, 104, 24, 24) {
			g.consStationRotate()
			return
		}
		if g.consTab == 1 && inRect(rx, ry, 3, 45, 132, 12) {
			g.openConsDropdown(win.X, win.Y, f.SkinColour, true)
			return
		}
		if g.consTab == 2 {
			switch ui.SignalHit(rx, ry) {
			case "dropdown":
				g.openConsDropdown(win.X, win.Y, f.SkinColour, false)
			case "both":
				g.consSigOne = false
			case "single":
				g.consSigOne = true
				// PARTIAL: placing signals (the tool, CreateSignal), the dropdown list
			}
			return
		}
		hit := ui.ConstructionHit(rx, ry, st)
		if hit == "bridge" {
			g.openBridgeDropdown(win.X, win.Y, f.SkinColour, f.CompanyColour)
			return
		}
		g.consWidget(hit)
	}
	return win
}

// consSwitchTab is a click on one of the construction window's tabs
// (Common::switchTab): the tab, its tool on the map again.
func (g *Game) consSwitchTab(tab int) {
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
func (g *Game) consWidget(hit string) {
	if !g.consIsRoad() && consUnusedAtRotation(hit, g.w.Cons.Rotation) {
		return // a disabled widget (PARTIAL: it isn't drawn greyed yet)
	}
	switch hit {
	case "":
	case "construct":
		g.consConstruct()
	case "rotate": // rotate_90: the rotate arrows while hovering, the red cross while building
		if !g.w.Cons.Hover { // back to a new position: the overground view (viewportSetVisibility)
			g.consOverground()
		}
		g.w.ConstructionRotate90(g.playerCmd())
	case "remove":
		g.consRemove()
	case "copy", "paste": // PARTIAL: blueprints not ported
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
	st := ui.SignalTabState{Both: !g.consSigOne, Image: -1}
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
		}
	}
	g.dropdown = d
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
	d.OnSelect = func(i int) {
		if i >= 0 && i < len(list) {
			g.consBridge = list[i]
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
		if g.w.TrackGhost != nil || g.w.RoadGhost != nil || g.w.ConsArrow.On {
			g.w.EndConstruction(g.playerCmd())
		}
		g.w.SyncConstructionArrow(false)
		return
	}
	road, obj := g.consIsRoad(), uint8(g.buildObjID)
	// cState.lastSelectedBridge and lastSelectedMods: the window's choices, for every piece built
	g.w.Cons.Bridge, g.w.Cons.Mods = 0, g.consMods
	if list := g.r.ObjMgr.AvailableCompatibleBridges(g.buildObjID, road, int(g.w.CurrentYear)); len(list) > 0 {
		g.w.Cons.Bridge = uint8(pickSlot(list, g.consBridge)) + 1
	}
	c := g.playerCmd()
	if !g.w.Cons.Hover {
		g.w.UpdateConstruction(c, road, obj)
	} else if g.hoverTileX >= 0 && !g.overWindow(g.mouseX, g.mouseY) {
		tx, ty, jh := g.consTarget(road)
		g.w.ConstructionToolUpdate(c, tx, ty, jh, road, obj)
	}
	g.w.SyncConstructionArrow(true)
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

// stationToolOn: a construction window is on its Station tab (every mode has
// one), so the map takes the station tool.
func (g *Game) stationToolOn() bool {
	return g.buildMode != buildModeNone && g.consTab == 1 && !g.inTitleScreen && !g.consToolOff
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
	title := i18n.Label(i18n.CantBuildPop3String)
	id := g.buildObjID
	if g.consIsRoad() {
		if id >= 0 && id < len(g.r.ObjMgr.RoadObjects) && g.r.ObjMgr.RoadObjects[id] != nil {
			title = ui.FormatErrorArgs(title, g.r.ObjMgr.RoadObjects[id].Name)
		}
	} else if id >= 0 && id < len(g.r.ObjMgr.TrackObjects) && g.r.ObjMgr.TrackObjects[id] != nil {
		title = ui.FormatErrorArgs(title, g.r.ObjMgr.TrackObjects[id].Name)
	}
	g.openError(title, cmdErrorText(c))
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
