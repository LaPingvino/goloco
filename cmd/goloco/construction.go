package main

import (
	"github.com/LaPingvino/goloco/pkg/construct"
	"image"

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
		f.H = ui.StationTabH
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
	st.Piece = nameFor(consPieceCurve, g.consCurve)
	if g.consOneShot >= 0 && !g.consIsRoad() { // the one-shot piece's button is the active one
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
	if g.r != nil && g.r.ObjMgr != nil {
		om := g.r.ObjMgr
		// lastSelectedBridge: the first available (PARTIAL: the scenario's remembered bridges[]).
		if list := om.AvailableCompatibleBridges(g.buildObjID, g.consIsRoad(), int(g.w.CurrentYear)); len(list) > 0 {
			if b := om.GetBridgeObjectByIndex(list[0]); b != nil {
				st.BridgeImage = int(b.ImageOffset)
			}
		}
	}
	return f, st
}

func (g *Game) newPortedConstructionWindow() *ui.SimpleWindow {
	win := ui.NewSimpleWindow("Construction", 10, 40, ui.ConstructionW, ui.ConstructionH)
	win.Ported = true
	win.PortedDraw = func(screen *ebiten.Image, x, y int) {
		f, st := g.constructionFrame(x, y)
		if f.Disabled[g.consTab] { // the tab went away (another track type)
			g.consTab = 0
			f, st = g.constructionFrame(x, y)
		}
		win.Height = f.Height()
		f.Draw(screen)
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
			if !g.consIsRoad() {
				g.consOneShot = consDualTrackPiece(hit == "s_bend_dual_track_right", g.w.ConstructionHeadState().Rotation)
			}
			// PARTIAL: the road turnaround (getRoadPieceId not ported yet).
		default:
			if v, ok := consPieceCurve[hit]; ok {
				g.consCurve = v
				g.consOneShot = -1
				if g.consIsRoad() {
					g.consSlope = 0
				}
			} else if v, ok := consGradientSlope[hit]; ok {
				// Track: piece and gradient are separate choices (changeTrackPiece);
				// a combination that can't be built gives no piece.
				g.consSlope = v
				if v != 0 && g.consIsRoad() {
					g.consCurve = 0
				}
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
	road := g.consIsRoad()
	list := om.AvailableCompatibleStations(g.buildObjID, road, int(g.w.CurrentYear))
	if len(list) == 0 {
		return st
	}
	slot := pickSlot(list, g.consStation)
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
	if station {
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
