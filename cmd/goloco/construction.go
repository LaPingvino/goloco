package main

import (
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
	if g.consTab == 2 {
		f.H = ui.SignalTabH
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
	st.Gradient = nameFor(consGradientSlope, g.consSlope)
	st.Hover = !g.w.ConstructionHeadState().Active
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
		if g.consTab == 2 {
			ui.DrawSignalTab(screen, f, g.signalTabState())
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
				g.consPlaceStation()
			case 0, 2:
				g.consTab = tab // switchTab; tabReset picks the both-directions tool
				g.consSigOne = false
			}
			return
		}
		if g.consTab == 2 {
			switch ui.SignalHit(rx, ry) {
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
		default:
			if v, ok := consPieceCurve[hit]; ok {
				g.consCurve = v
				g.consSlope = 0
			} else if v, ok := consGradientSlope[hit]; ok {
				g.consSlope = v
				if v != 0 {
					g.consCurve = 0
				}
			}
			// PARTIAL: s-bends and the road turnaround (no goloco piece mapping yet).
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
	if s := om.GetTrainSignalObjectByIndex(list[0]); s != nil {
		st.Name, st.Description, st.Image = s.Name, s.Description, int(s.ImageOffset)
	}
	return st
}
