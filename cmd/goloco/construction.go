package main

import (
	"github.com/LaPingvino/goloco/pkg/construct"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"image"

	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// The ported construction window (pkg/ui ConstructionFrame + construction tab)
// for railway track and roads, in place of the improvised SimpleWindow.

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

func (g *Game) newPortedConstructionWindow() *ui.SimpleWindow {
	win := ui.NewSimpleWindow(i18n.Label(i18n.Construction), 10, 40, ui.ConstructionW, ui.ConstructionH)
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
				g.consTab = 1 // switchTab; tabReset turns the placing tool on
				g.consPlaceStation()
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
			speed = L.Velocity(int(b.MaxSpeed))
		}
		items[i] = ui.DropdownItem{Sprite: int(b.ImageOffset), SpriteColour: company,
			Lines: []string{"Max. speed: " + speed, "Height limit: " + L.Height(int(b.MaxHeight))}}
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

// stationArgsAt is getTrainStationPlacementArgsFromCursor with the selected
// station type; ok false off track.
func (g *Game) stationArgsAt(sx, sy int) (sim.TrainStationPlacement, bool) {
	if g.r == nil || g.r.ObjMgr == nil {
		return sim.TrainStationPlacement{}, false
	}
	list := g.r.ObjMgr.AvailableCompatibleStations(g.buildObjID, false, int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.TrainStationPlacement{}, false
	}
	return g.w.TrainStationPlacementAt(sx, sy, uint8(pickSlot(list, g.consStation)))
}

func (g *Game) stationCmd() (*sim.Cmd, sim.G1Raw) {
	company, _ := g.w.ControllingCompany()
	g1, _ := g.r.ObjMgr.G1File.(sim.G1Raw)
	return &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}, g1
}

// stationToolUpdate is StationTab's onToolUpdateTrainStation (0x0049E421 side):
// off track, the ghost goes (onToolUpdateFail); on the same piece it stays;
// else the old ghost goes and a new one is placed (apply | noErrorWindow |
// noPayment | ghost), its cost kept; the catchment shown on the map and the
// ghost's accepted / produced cargo (calcAcceptedCargoTrainStationGhost).
func (g *Game) stationToolUpdate(sx, sy int) {
	if g.buildMode == buildModeAirport {
		g.airportToolUpdate(sx, sy)
		return
	}
	if g.buildMode == buildModeDock {
		g.dockToolUpdate(sx, sy)
		return
	}
	if g.consIsRoad() {
		g.roadStationToolUpdate(sx, sy)
		return
	}
	a, ok := g.stationArgsAt(sx, sy)
	if !ok {
		g.removeStationGhost()
		return
	}
	if gh := g.stationGhost; gh != nil && gh.X == a.X && gh.Y == a.Y && gh.Z == a.Z && gh.Rotation == a.Rotation &&
		gh.TrackID == a.TrackID && gh.Index == a.Index && gh.TrackObjectID == a.TrackObjectID {
		return
	}
	g.removeStationGhost()
	c, g1 := g.stationCmd()
	cost, ok := g.w.Map.CreateTrainStation(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, a, g1)
	g.stationGhostCost = cost
	if ok {
		g.stationGhost = &a
		m := &g.w.Map // constructingStationId and the ghost's catchment cargo
		g.stationGhostStation = m.LastAdjoiningStation
		x, y := m.LastAdjoiningStationX, m.LastAdjoiningStationY
		if x == -1 {
			x, y = a.X, a.Y
		}
		g.stationGhostAccepted, g.stationGhostProduced = m.CalcAcceptedCargoTrainStationGhost(g.stationGhostStation, x, y, 0xFFFFFFFF, g.r.ObjMgr)
		g.w.SetCatchmentDisplay(m.GhostCatchmentDisplay(g.stationGhostStation, x, y, g.r.ObjMgr)) // catchmentArea + setCatchmentDisplay / sub_491BF5 (flag_0)
	}
}

// removeStationGhost is removeConstructionGhosts' station part: the ghost
// removed as it was placed. Only with a station ghost showing
// (GhostVisibilityFlags::station) is the catchment display cleared — so the
// Station window's own catchment (showStationCatchment) stays.
func (g *Game) removeStationGhost() {
	if g.stationGhost != nil || g.roadStationGhost != nil || g.dockGhost != nil || g.airportGhost != nil {
		g.w.SetCatchmentDisplay(nil) // sub_491BC6 + resetMapSelectionFlag(catchmentArea)
		g.catchmentStation = -1      // the catchment shown is the ghost's from here
	}
	if rg := g.roadStationGhost; rg != nil { // stationGhostType & (1 << 7): RoadStationRemovalArgs
		g.roadStationGhost = nil
		c, g1 := g.stationCmd()
		g.w.Map.RemoveRoadStation(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost,
			sim.RoadStationRemoval{X: rg.X, Y: rg.Y, Z: rg.Z, Rotation: rg.Rotation, RoadID: rg.RoadID, Index: rg.Index, RoadObjectID: rg.RoadObjectID}, g1)
	}
	gh := g.stationGhost
	if gh == nil {
		return
	}
	g.stationGhost = nil
	c, g1 := g.stationCmd()
	g.w.Map.RemoveTrainStation(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost,
		sim.TrainStationRemoval{X: gh.X, Y: gh.Y, Z: gh.Z, Rotation: gh.Rotation, TrackID: gh.TrackID, Index: gh.Index, TrackObjectID: gh.TrackObjectID}, g1)
}

// stationToolUp is StationTab's onToolUpTrainStation (0x004A5390): the
// placement from the track under the cursor (getTrainStationPlacementArgsFromCursor),
// refused on another track type (wrong_type_of_track_road), then
// CreateTrainStation with apply, after removeConstructionGhosts; a failure
// opens the error window titled cant_build_pop3_string with the station's
// name and the command's reason (when the command sets one). PARTIAL: the
// construct sound.
func (g *Game) stationToolUp(sx, sy int) {
	if g.buildMode == buildModeAirport {
		g.airportToolUp(sx, sy)
		return
	}
	if g.buildMode == buildModeDock {
		g.dockToolUp(sx, sy)
		return
	}
	if g.consIsRoad() {
		g.roadStationToolUp(sx, sy)
		return
	}
	g.removeStationGhost()
	a, ok := g.stationArgsAt(sx, sy)
	if !ok {
		return
	}
	title := ""
	om := g.r.ObjMgr
	if list := om.AvailableCompatibleStations(g.buildObjID, false, int(g.w.CurrentYear)); len(list) > 0 {
		if id := pickSlot(list, g.consStation); id >= 0 && id < len(om.TrainStationObjects) && om.TrainStationObjects[id] != nil {
			title = ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), om.TrainStationObjects[id].Name)
		}
	}
	if int(a.TrackObjectID) != g.buildObjID {
		g.openError("", i18n.Label(i18n.WrongTypeOfTrackRoad))
		return
	}
	c, g1 := g.stationCmd()
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return g.w.Map.CreateTrainStation(c, f, a, g1) }); !ok { // GameCommands::doCommand: checked, then charged
		g.openError(title, cmdErrorText(c))
	}
}

// openError is Error::open at the mouse.
func (g *Game) openError(title, message string) {
	mx, my := ebiten.CursorPosition()
	g.errorWin = ui.OpenError(title, message, mx, my, g.sw, g.sh)
}

// roadStationArgsAt is getRoadStationPlacementArgsFromCursor with the
// selected stop type.
func (g *Game) roadStationArgsAt(sx, sy int) (sim.RoadStationPlacement, int, bool) {
	if g.r == nil || g.r.ObjMgr == nil {
		return sim.RoadStationPlacement{}, -1, false
	}
	list := g.r.ObjMgr.AvailableCompatibleStations(g.buildObjID, true, int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.RoadStationPlacement{}, -1, false
	}
	typ := pickSlot(list, g.consStation)
	a, ok := g.w.RoadStationPlacementAt(sx, sy, uint8(typ))
	return a, typ, ok
}

// roadStationToolUpdate is StationTab's onToolUpdateRoadStation: as the train
// one, with the road stop ghost; the catchment's cargo filtered by the stop
// (a passenger stop: its cargo only; a freight bay: all but that).
func (g *Game) roadStationToolUpdate(sx, sy int) {
	a, typ, ok := g.roadStationArgsAt(sx, sy)
	if !ok {
		g.removeStationGhost()
		return
	}
	if gh := g.roadStationGhost; gh != nil && gh.X == a.X && gh.Y == a.Y && gh.Z == a.Z && gh.Rotation == a.Rotation &&
		gh.RoadID == a.RoadID && gh.Index == a.Index && gh.RoadObjectID == a.RoadObjectID {
		return
	}
	g.removeStationGhost()
	c, g1 := g.stationCmd()
	cost, ok := g.w.Map.CreateRoadStation(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, a, g1)
	g.stationGhostCost = cost
	if !ok {
		return
	}
	g.roadStationGhost = &a
	m := &g.w.Map
	g.stationGhostStation = m.LastAdjoiningStation
	x, y := m.LastAdjoiningStationX, m.LastAdjoiningStationY
	if x == -1 {
		x, y = a.X, a.Y
	}
	g.w.SetCatchmentDisplay(m.GhostCatchmentDisplay(g.stationGhostStation, x, y, g.r.ObjMgr))
	g.stationGhostAccepted, g.stationGhostProduced = m.CalcAcceptedCargoTrainStationGhost(g.stationGhostStation, x, y, sim.RoadStationCargoFilter(g.r.ObjMgr, typ), g.r.ObjMgr)
}

// roadStationToolUp is StationTab's onToolUpRoadStation (0x004A548F): after
// removeConstructionGhosts, CreateRoadStation with apply on the road under
// the cursor; a failure opens the error window titled cant_build_pop3_string
// with the stop's name. PARTIAL: the construct sound, the command's own
// error text.
func (g *Game) roadStationToolUp(sx, sy int) {
	g.removeStationGhost()
	a, typ, ok := g.roadStationArgsAt(sx, sy)
	if !ok {
		return
	}
	om := g.r.ObjMgr
	title := ""
	if typ >= 0 && typ < len(om.RoadStationObjects) && om.RoadStationObjects[typ] != nil {
		title = ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), om.RoadStationObjects[typ].Name)
	}
	c, g1 := g.stationCmd()
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return g.w.Map.CreateRoadStation(c, f, a, g1) }); !ok { // GameCommands::doCommand: checked, then charged
		g.openError(title, "")
	}
}

// ghostTabInfo is the Station tab's ghost part: cState.stationCost, the
// station it joins (or New Station) and the catchment's cargo lists.
func (g *Game) ghostTabInfo(st *ui.StationTabState) {
	om := g.r.ObjMgr
	st.Cost = int64(g.stationGhostCost) // cState.stationCost while a ghost stands
	st.Ghost = true
	st.GhostName = i18n.Label(i18n.NewStation)
	if id := g.stationGhostStation; id >= 0 && id < len(g.w.Map.StationNames) {
		st.GhostName = g.w.Map.StationName(id)
	}
	lines := func(bits uint32) []ui.CargoLine {
		var out []ui.CargoLine
		for i := 0; i < 32; i++ {
			if bits&(1<<i) == 0 || i >= len(om.CargoObjects) || om.CargoObjects[i] == nil {
				continue
			}
			c := om.CargoObjects[i]
			out = append(out, ui.CargoLine{Name: c.Name, Image: int(c.ImageOffset)})
		}
		return out
	}
	st.Accepted, st.Produced = lines(g.stationGhostAccepted), lines(g.stationGhostProduced)
}

// dockArgsAt is getDockPlacementArgsFromCursor: the surface or water under
// the cursor (getSurfaceOrWaterLocFromUi), the dock placed by
// getDockPlacementArgs with the selected dock.
func (g *Game) dockArgsAt(sx, sy int) (sim.PortPlacement, int, bool) {
	if g.r == nil || g.r.ObjMgr == nil {
		return sim.PortPlacement{}, -1, false
	}
	list := g.r.ObjMgr.AvailableDocks(int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.PortPlacement{}, -1, false
	}
	px, py, ok := g.w.SurfaceOrWaterAt(sx, sy)
	if !ok {
		return sim.PortPlacement{}, -1, false
	}
	typ := pickSlot(list, g.consStation)
	a, ok := g.w.Map.DockPlacementAt(px, py, uint8(typ), g.r.ObjMgr)
	return a, typ, ok
}

// dockToolUpdate is StationTab's onToolUpdateDock: the ghost dock follows
// the cursor (createPort with apply | noErrorWindow | noPayment | ghost),
// its cost, the catchment shown (sub_491D20) and its cargo
// (calcAcceptedCargoDockGhost). PARTIAL: the construction arrow and the
// selected 2×2 tiles (map selection).
func (g *Game) dockToolUpdate(sx, sy int) {
	a, _, ok := g.dockArgsAt(sx, sy)
	if !ok {
		g.removeStationGhost()
		return
	}
	if gh := g.dockGhost; gh != nil && gh.X == a.X && gh.Y == a.Y && gh.Z == a.Z && gh.Rotation == a.Rotation && gh.Type == a.Type {
		return
	}
	g.removeStationGhost()
	c, g1 := g.stationCmd()
	cost, ok := g.w.Map.CreatePort(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, a, g1)
	g.stationGhostCost = cost
	if !ok {
		return
	}
	g.dockGhost = &a
	m := &g.w.Map
	g.stationGhostStation = m.LastAdjoiningStation
	x, y := m.LastAdjoiningStationX, m.LastAdjoiningStationY
	if x == -1 {
		x, y = a.X, a.Y
	}
	g.w.SetCatchmentDisplay(m.GhostDockCatchmentDisplay(g.stationGhostStation, x, y, g.r.ObjMgr))
	g.stationGhostAccepted, g.stationGhostProduced = m.CalcAcceptedCargoDockGhost(g.stationGhostStation, x, y, 0xFFFFFFFF, g.r.ObjMgr)
}

// dockToolUp is StationTab's onToolUpDock (0x004A55AB): after
// removeConstructionGhosts, createPort with apply; a failure opens the error
// window titled cant_build_pop3_string with the dock's name.
func (g *Game) dockToolUp(sx, sy int) {
	g.removeStationGhost()
	a, typ, ok := g.dockArgsAt(sx, sy)
	if !ok {
		return
	}
	om := g.r.ObjMgr
	title := ""
	if typ >= 0 && typ < len(om.DockObjects) && om.DockObjects[typ] != nil {
		title = ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), om.DockObjects[typ].Name)
	}
	c, g1 := g.stationCmd()
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return g.w.Map.CreatePort(c, f, a, g1) }); !ok { // GameCommands::doCommand: checked, then charged
		g.openError(title, "")
	}
}

// airportArgsAt is getAirportPlacementArgsFromCursor: the surface or water
// under the cursor, the selected airport at the window's rotation.
func (g *Game) airportArgsAt(sx, sy int) (sim.AirportPlacement, int, bool) {
	if g.r == nil || g.r.ObjMgr == nil {
		return sim.AirportPlacement{}, -1, false
	}
	list := g.r.ObjMgr.AvailableAirports(int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.AirportPlacement{}, -1, false
	}
	px, py, ok := g.w.SurfaceOrWaterAt(sx, sy)
	if !ok {
		return sim.AirportPlacement{}, -1, false
	}
	typ := pickSlot(list, g.consStation)
	a, ok := g.w.Map.AirportPlacementAt(px, py, uint8(typ), g.airportRotation, g.r.ObjMgr)
	return a, typ, ok
}

// airportToolUpdate is StationTab's onToolUpdateAirport: the ghost airport
// follows the cursor, its cost, the catchment (sub_491C6F) and its cargo
// (calcAcceptedCargoAirportGhost). PARTIAL: the construction arrow and the
// selected tiles of its extents.
func (g *Game) airportToolUpdate(sx, sy int) {
	a, typ, ok := g.airportArgsAt(sx, sy)
	if !ok {
		g.removeStationGhost()
		return
	}
	if gh := g.airportGhost; gh != nil && *gh == a {
		return
	}
	g.removeStationGhost()
	c, g1 := g.stationCmd()
	cost, ok := g.w.Map.CreateAirport(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, a, g1)
	g.stationGhostCost = cost
	if !ok {
		return
	}
	g.airportGhost = &a
	m := &g.w.Map
	g.stationGhostStation = m.LastAdjoiningStation
	x, y := m.LastAdjoiningStationX, m.LastAdjoiningStationY
	if x == -1 {
		x, y = a.X, a.Y
	}
	ap := g.r.ObjMgr.AirportObjects[typ]
	g.w.SetCatchmentDisplay(m.GhostAirportCatchmentDisplay(g.stationGhostStation, ap, x, y, a.Rotation, g.r.ObjMgr))
	g.stationGhostAccepted, g.stationGhostProduced = m.CalcAcceptedCargoAirportGhost(g.stationGhostStation, ap, x, y, a.Rotation, 0xFFFFFFFF, g.r.ObjMgr)
}

// airportToolUp is StationTab's onToolUpAirport (0x004A5550): after
// removeConstructionGhosts, createAirport with apply; a failure opens the
// error window titled cant_build_pop3_string with the airport's name.
func (g *Game) airportToolUp(sx, sy int) {
	g.removeStationGhost()
	a, typ, ok := g.airportArgsAt(sx, sy)
	if !ok {
		return
	}
	om := g.r.ObjMgr
	title := ""
	if typ >= 0 && typ < len(om.AirportObjects) && om.AirportObjects[typ] != nil {
		title = ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), om.AirportObjects[typ].Name)
	}
	c, g1 := g.stationCmd()
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return g.w.Map.CreateAirport(c, f, a, g1) }); !ok { // GameCommands::doCommand: checked, then charged
		g.openError(title, "")
	}
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
