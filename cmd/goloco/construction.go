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
// OpenLoco reference: src/OpenLoco/src/Ui/Windows/Construction/Construction.cpp
// Curve row + slope row of sprite buttons, Rotate/Undo, and a Build action.
// Click the map to place the construction head; pieces then chain from it.

// consPieceNames maps upstream's piece widgets to goloco's consCurve values
// (the old window's sprite↔value pairs) and gradient widgets to consSlope.
var (
	consPieceCurve = map[string]int{
		"left_hand_curve_very_small": -1, "left_hand_curve_small": -2, "left_hand_curve": -3,
		"right_hand_curve": 3, "right_hand_curve_small": 2, "right_hand_curve_very_small": 1,
		"left_hand_curve_large": -4, "straight": 0, "right_hand_curve_large": 4,
	}
	consGradientSlope = map[string]int{
		"steep_slope_down": -2, "slope_down": -1, "level": 0, "slope_up": 1, "steep_slope_up": 2,
	}
)

func nameFor(m map[string]int, v int) string {
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
	if g.r != nil && g.r.ObjMgr != nil {
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
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			f.StationTabIcon = int(g.r.ObjMgr.InterfaceSkin.ImageOffset) + objects.ISToolbarMenuAirport
		}
	}
	if g.buildMode == buildModeDock { // drawDockTabs; StationTab::prepareDraw's caption title_ship_port
		g.consTab, f.Tab = 1, 1
		f.Title = i18n.Label(i18n.TitleShipPort)
		f.Disabled = [4]bool{true, false, true, true}
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			f.StationTabIcon = int(g.r.ObjMgr.InterfaceSkin.ImageOffset) + objects.ISToolbarMenuShipPort
		}
	}
	st.Rotation = g.w.ConstructionHeadState().Rotation
	st.Piece = nameFor(consPieceCurve, g.consCurve)
	if g.consOneShot >= 0 && g.consIsRoad() {
		st.Piece = "turnaround"
	} else if g.consOneShot >= 0 { // the one-shot piece's button is the active one
		switch g.consOneShot {
		case construct.SBendLeft:
			st.Piece = "s_bend_left"
		case construct.SBendRight:
			st.Piece = "s_bend_right"
		default:
			rot := g.w.ConstructionHeadState().Rotation
			st.Piece = "s_bend_dual_track_left"
			if consDualTrackPiece(false, rot) != g.consOneShot {
				st.Piece = "s_bend_dual_track_right"
			}
		}
	}
	st.Gradient = nameFor(consGradientSlope, g.consSlope)
	st.Hover = !g.w.ConstructionHeadState().Active
	st.BridgeImage, st.Company = -1, f.CompanyColour
	// trackCost/roadCost: the selected piece's cost while extending (the
	// ghost's cost upstream).
	if !st.Hover && g.r != nil && g.r.ObjMgr != nil {
		if id := g.consSelectedID(); id >= 0 {
			cmf := &g.w.CurrencyMultiplicationFactor
			om := g.r.ObjMgr
			mods := func(names []string) []*construct.Mod { // trackObj->mods[i]
				out := make([]*construct.Mod, len(names))
				for i, n := range names {
					if e := om.Extra(n); e != nil {
						out[i] = &construct.Mod{Pieces: e.Pieces, BuildCostFactor: e.BuildCostFactor, CostIndex: e.CostIndex}
					}
				}
				return out
			}
			if g.consIsRoad() {
				if r := om.GetRoadObjectByIndex(g.buildObjID); r != nil {
					st.Cost = int64(construct.RoadPieceCost(id, r.BuildCostFactor, r.CostIndex, cmf) +
						construct.ModsCost(true, id, g.consMods, mods(r.ModNames), cmf))
				}
			} else if t := om.GetTrackObjectByIndex(g.buildObjID); t != nil {
				st.Cost = int64(construct.TrackPieceCost(id, t.BuildCostFactor, t.CostIndex, cmf) +
					construct.ModsCost(false, id, g.consMods, mods(t.ModNames), cmf))
			}
		}
	}
	if g.r != nil && g.r.ObjMgr != nil {
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
	if g.windowMgr != nil {
		for _, w := range g.windowMgr.Windows {
			if w.Title == i18n.Label(i18n.Construction) {
				return w.X, w.Y
			}
		}
	}
	sw, sh := g.sw, g.sh
	if sw <= 0 || sh <= 0 {
		sw, sh = defaultWidth, defaultHeight
	}
	if g.windowMgr == nil {
		return ui.AutoWindowPosition(ui.ConstructionW, ui.ConstructionH, sw, sh, nil)
	}
	return g.windowMgr.AutoPosition(ui.ConstructionW, ui.ConstructionH, sw, sh)
}

func (g *Game) newPortedConstructionWindow() *ui.SimpleWindow {
	x, y := g.consWindowOrigin()
	win := ui.NewSimpleWindow(i18n.Label(i18n.Construction), x, y, ui.ConstructionW, ui.ConstructionH)
	win.Ported = true
	win.PortedDraw = func(screen *ebiten.Image, x, y int) {
		f, st := g.constructionFrame(x, y)
		if f.Disabled[g.consTab] { // the tab went away (another track type)
			g.consTab = 0
			f, st = g.constructionFrame(x, y)
		}
		win.Height = f.Height()
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
			rot := int(g.w.ConstructionHeadState().Rotation) & 3
			g.w.DrawPiecePreview(clip, float64(x+3+66), float64(y+145+28), 1.0, g.consIsRoad(), g.consDisplayID(), g.buildObjID, rot)
		}
	}
	// clickable(Construction): close_button tab_construction tab_station tab_signal tab_overhead construct rotate bridge_dropdown station_dropdown signal_dropdown both_directions single_direction checkbox_1 checkbox_2 checkbox_3 checkbox_4 image rotate_90 straight left_hand_curve left_hand_curve_large left_hand_curve_small left_hand_curve_very_small right_hand_curve right_hand_curve_large right_hand_curve_small right_hand_curve_very_small s_bend_left s_bend_right s_bend_dual_track_left s_bend_dual_track_right level slope_up slope_down steep_slope_up steep_slope_down
	// clickable-partial(Construction): remove
	win.OnWindowClick = func(rx, ry int) {
		f, st := g.constructionFrame(win.X, win.Y)
		if tab := f.TabAt(rx, ry); tab >= 0 || (ry >= 15 && ry < 42) { // tabs
			switch tab {
			case 1:
				g.consTab = 1 // switchTab; tabReset turns the placing tool on (the Station tab's tool builds, through createTrainStation)
			case 0, 2, 3:
				g.consTab = tab // switchTab; tabReset picks the both-directions tool
				g.consSigOne = false
			}
			return
		}
		if g.consTab == 3 {
			if i := ui.OverheadHit(rx, ry, g.overheadTabState()); i >= 0 {
				g.consMods ^= 1 << i // lastSelectedMods ^= (1 << checkbox)
			}
			return
		}
		if g.consTab == 1 && g.buildMode == buildModeAirport && rx >= 112 && rx < 112+24 && ry >= 104 && ry < 104+24 {
			g.airportRotation = (g.airportRotation + 1) & 3 // widx::rotate: constructionRotation++, stationCost = failure
			g.stationGhostCost = 0
			return
		}
		if g.consTab == 1 && rx >= 3 && rx < 3+132 && ry >= 45 && ry < 45+12 {
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
		switch hit := ui.ConstructionHit(rx, ry, st); hit {
		case "":
		case "construct":
			if id := g.consSelectedID(); id >= 0 {
				g.consPlace(id)
			}
		case "bridge":
			g.openBridgeDropdown(win.X, win.Y, f.SkinColour, f.CompanyColour)
		case "rotate":
			g.w.RotateConstructionHead()
		case "remove":
			g.consUndo()
		case "copy", "paste": // PARTIAL: blueprints not ported
		case "s_bend_left", "s_bend_right":
			if !g.consIsRoad() { // changeTrackPiece(s_bend_…)
				g.consOneShot = map[string]int{"s_bend_left": construct.SBendLeft, "s_bend_right": construct.SBendRight}[hit]
			}
		case "s_bend_dual_track_left", "s_bend_dual_track_right", "turnaround":
			if g.consIsRoad() {
				g.consOneShot = construct.Turnaround // the road's turnaround button
			} else {
				g.consOneShot = consDualTrackPiece(hit == "s_bend_dual_track_right", g.w.ConstructionHeadState().Rotation)
			}
		default:
			if v, ok := consPieceCurve[hit]; ok {
				g.consCurve = v
				g.consOneShot = -1
			} else if v, ok := consGradientSlope[hit]; ok {
				// Piece and gradient are separate choices (changeTrackPiece); a
				// combination that can't be built gives no piece.
				g.consSlope = v
			}
		}
	}
	return win
}

// signalTabState is the signal tab for the track being built: the selected
// signal is the first available one (lastSelectedSignal; PARTIAL: the
// scenario's remembered choice, Scenario::getConstruction().signals).
func (g *Game) signalTabState() ui.SignalTabState {
	st := ui.SignalTabState{Both: !g.consSigOne, Image: -1}
	if g.r == nil || g.r.ObjMgr == nil {
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
	if g.r == nil || g.r.ObjMgr == nil {
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
	if g.r == nil || g.r.ObjMgr == nil {
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
	if g.r == nil || g.r.ObjMgr == nil {
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
	if g.r == nil || g.r.ObjMgr == nil {
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

// updateTrackGhost is the Construction tab's tool update while positioning
// (onToolUpdate → onToolUpdateSingle for track): with the track window on
// its Construction tab and no head placed yet, the selected piece in the
// construction rotation (getTrackPieceId) stands as a ghost under the
// cursor; otherwise none. PARTIAL: road (the road ghost), building off
// existing track (tryMakeTrackJunctionAtLoc), the construction arrow.
func (g *Game) updateTrackGhost() {
	if g.w == nil || g.r == nil || g.r.ObjMgr == nil {
		return
	}
	// updateConstruction for a road: the selected road piece as a ghost at the head.
	if g.buildMode == buildModeRoad && g.consTab == 0 && !g.inTitleScreen && g.w.ConstructionHeadState().Active {
		c, _ := g.stationCmd()
		g.w.UpdateHeadRoadGhost(c, g.consSelectedRoadID(), uint8(g.buildObjID))
	} else if g.w.RoadGhost != nil {
		c, _ := g.stationCmd()
		g.w.RemoveRoadGhost(c)
	}
	if g.buildMode == buildModeTrack && g.consTab == 0 && !g.inTitleScreen && g.w.ConstructionHeadState().Active {
		// updateConstruction: the selected piece as a ghost at the head.
		c, _ := g.stationCmd()
		g.w.UpdateHeadTrackGhost(c, g.consSelectedTrackID(), uint8(g.buildObjID))
		return
	}
	if g.buildMode != buildModeTrack || g.consTab != 0 || g.inTitleScreen || g.hoverTileX < 0 {
		g.removeTrackGhost()
		return
	}
	p, ok := construct.GetTrackPieceID(g.consTrackPiece(), consTrackGradient(g.consSlope), g.w.ConstructionHeadState().Rotation)
	if !ok {
		g.removeTrackGhost()
		return
	}
	bridge := uint8(0xFF)
	if list := g.r.ObjMgr.AvailableCompatibleBridges(g.buildObjID, false, int(g.w.CurrentYear)); len(list) > 0 {
		bridge = uint8(pickSlot(list, g.consBridge))
	}
	c, _ := g.stationCmd()
	g.w.UpdateTrackGhost(c, g.hoverTileX, g.hoverTileY, p.ID, p.Rotation, uint8(g.buildObjID), bridge)
}

// removeTrackGhost is removeConstructionGhosts' track part (removeTrackGhosts).
func (g *Game) removeTrackGhost() {
	if g.w == nil || g.w.TrackGhost == nil {
		return
	}
	c, _ := g.stationCmd()
	g.w.RemoveTrackGhost(c)
}

// consSelectedTrackID resolves the current curve/slope selection to the first
// placeable track id at the head, or -1.
func (g *Game) consSelectedTrackID() int {
	// getTrackPieceId (pkg/construct) for the chosen piece and gradient at
	// the head's constructionRotation.
	rot := g.w.ConstructionHeadState().Rotation
	p, ok := construct.GetTrackPieceID(g.consTrackPiece(), consTrackGradient(g.consSlope), rot)
	if !ok || !g.w.CanPlaceTrackAtHead(int(p.ID)) {
		return -1
	}
	return int(p.ID)
}

// consTrackPiece is lastSelectedTrackPiece: a one-shot piece if chosen, else
// the curve buttons' piece.
func (g *Game) consTrackPiece() uint8 {
	if g.consOneShot >= 0 {
		return uint8(g.consOneShot)
	}
	switch g.consCurve {
	case -1:
		return construct.LeftCurveVerySmall
	case 1:
		return construct.RightCurveVerySmall
	case -2:
		return construct.LeftCurveSmall
	case 2:
		return construct.RightCurveSmall
	case -3:
		return construct.LeftCurve
	case 3:
		return construct.RightCurve
	case -4:
		return construct.LeftCurveLarge
	case 4:
		return construct.RightCurveLarge
	}
	return construct.Straight
}

// consTrackGradient is lastSelectedTrackGradient for goloco's slope choice.
func consTrackGradient(slope int) uint8 {
	switch slope {
	case -2:
		return construct.SteepSlopeDown
	case -1:
		return construct.SlopeDown
	case 1:
		return construct.SlopeUp
	case 2:
		return construct.SteepSlopeUp
	}
	return construct.Level
}

// consDualTrackPiece is the piece the left or right dual-track button
// chooses at this heading (construct.DualTrackButtons).
func consDualTrackPiece(right bool, rot uint8) int {
	_, _, l, r := construct.DualTrackButtons(rot)
	if right {
		return int(r)
	}
	return int(l)
}

// consSelectedRoadID is getRoadPieceId (pkg/construct) for the chosen piece
// and gradient at the head, if it can be placed there; else -1.
func (g *Game) consSelectedRoadID() int {
	rot := g.w.ConstructionHeadState().Rotation
	p, ok := construct.GetRoadPieceID(g.consTrackPiece(), consTrackGradient(g.consSlope), rot)
	if !ok || !g.w.CanPlaceRoadAtHead(int(p.ID)) {
		return -1
	}
	return int(p.ID)
}

func (g *Game) consIsRoad() bool { return g.buildMode == buildModeRoad }

func (g *Game) consSelectedID() int {
	if g.consIsRoad() {
		return g.consSelectedRoadID()
	}
	return g.consSelectedTrackID()
}

// consDisplayID returns a piece id to show in the preview panel for the current
// curve/slope selection, regardless of whether it is placeable at the head (so
// the panel always shows the piece the buttons describe). Falls back to the
// first candidate, or straight.
func (g *Game) consDisplayID() int {
	if id := g.consSelectedID(); id >= 0 {
		return id
	}
	// The piece getTrackPieceId/getRoadPieceId gives; none (−1) when it
	// can't be built at this heading.
	rot := g.w.ConstructionHeadState().Rotation
	get := construct.GetTrackPieceID
	if g.consIsRoad() {
		get = construct.GetRoadPieceID
	}
	if p, ok := get(g.consTrackPiece(), consTrackGradient(g.consSlope), rot); ok {
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
	msg := ""
	if c.ErrorText != 0 {
		msg = i18n.Label(c.ErrorText)
		if c.ErrorArg != "" {
			msg = ui.FormatErrorArgs(msg, c.ErrorArg)
		}
	}
	g.openError(title, msg)
}

func (g *Game) consPlace(id int) {
	var built bool
	if g.consIsRoad() && g.r != nil && g.r.ObjMgr != nil {
		// constructRoad: the createRoad game command (ghost taken up first), checked and charged.
		c, _ := g.stationCmd()
		if _, built = g.w.BuildRoadAtHead(c, id, uint8(g.buildObjID)); !built {
			g.openTrackBuildError(c)
		}
	} else if g.consIsRoad() {
		built = g.w.PlaceRoadAtHead(id, uint8(g.buildObjID))
	} else if g.r != nil && g.r.ObjMgr != nil { // the createTrack game command: checked and charged (PL.1 e)
		c, _ := g.stationCmd()
		if _, built = g.w.BuildTrackAtHead(c, id, uint8(g.buildObjID)); !built {
			g.openTrackBuildError(c)
		}
	} else {
		built = g.w.PlaceTrackAtHead(id, uint8(g.buildObjID))
	}
	if built && g.consOneShot >= 0 {
		// constructTrack/constructRoad: one-shot pieces (≥ 9) go back to straight.
		g.consOneShot, g.consCurve = -1, 0
	}
}

func (g *Game) consUndo() {
	if g.consIsRoad() {
		g.w.UndoLastRoad()
	} else {
		g.w.UndoLastTrack()
	}
}

// consObjectNames returns the object display names for the current mode,
// indexed by object slot (empty string for unused slots).
func (g *Game) consObjectNames() []string {
	var names []string
	if g.r == nil || g.r.ObjMgr == nil {
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
