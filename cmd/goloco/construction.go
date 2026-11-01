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
	f := &ui.ConstructionFrame{X: x, Y: y, Title: g.consObjName(), FrameNo: g.titleFrame}
	var st ui.ConstructionTabState
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
	company, _ := g.w.ControllingCompany()
	if company < len(g.w.CompanyColours) {
		f.CompanyColour = g.w.CompanyColours[company]
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
		f.Draw(screen)
		ui.DrawConstructionTab(screen, f, st)
		// The piece preview in the construct area (goloco's preview drawing;
		// PARTIAL: upstream draws it with the paint code, drawTrack).
		if clip, ok := screen.SubImage(image.Rect(x+4, y+146, x+3+132-1, y+145+100-1)).(*ebiten.Image); ok {
			rot := int(g.w.ConstructionHeadState().Rotation) & 3
			g.w.DrawPiecePreview(clip, float64(x+3+66), float64(y+145+28), 1.0, g.consIsRoad(), g.consDisplayID(), g.buildObjID, rot)
		}
	}
	win.OnWindowClick = func(rx, ry int) {
		if ry >= 15 && ry < 42 && rx >= 3 { // tabs
			if (rx-3)/31 == 1 {
				g.consPlaceStation()
			}
			return
		}
		_, st := g.constructionFrame(win.X, win.Y)
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
