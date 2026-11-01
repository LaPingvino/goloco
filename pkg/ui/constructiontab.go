package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// TrackTraitFlags (World::Track::TrackTraitFlags).
const (
	TraitLargeCurve     = 1 << 1
	TraitNormalCurve    = 1 << 2
	TraitSmallCurve     = 1 << 3
	TraitVerySmallCurve = 1 << 4
	TraitSlope          = 1 << 5
	TraitSteepSlope     = 1 << 6
	TraitOneSided       = 1 << 7
)

// ConstructionButton is one of the construction tab's image buttons
// (ConstructionTab.cpp widgets: position, size, image).
type ConstructionButton struct {
	Name       string
	X, Y, W, H int
	Image      int
}

// constructionButtons are the tab's piece and gradient buttons in widget
// order (widx left_hand_curve_very_small … steep_slope_up), at the positions
// activateSelectedTrackWidgets leaves for a track without very small curves.
var constructionButtons = []ConstructionButton{
	{"left_hand_curve_very_small", 3, 45, 22, 24, 2340},
	{"left_hand_curve_small", 3, 45, 22, 24, 2342},
	{"left_hand_curve", 25, 45, 22, 24, 2344},
	{"left_hand_curve_large", 47, 45, 22, 24, 2346},
	{"right_hand_curve_large", 69, 45, 22, 24, 2347},
	{"right_hand_curve", 91, 45, 22, 24, 2345},
	{"right_hand_curve_small", 113, 45, 22, 24, 2343},
	{"right_hand_curve_very_small", 113, 45, 22, 24, 2341},
	{"s_bend_dual_track_left", 9, 69, 24, 24, 2355},
	{"s_bend_left", 33, 69, 24, 24, 2353},
	{"straight", 57, 69, 24, 24, 2335},
	{"s_bend_right", 81, 69, 24, 24, 2354},
	{"s_bend_dual_track_right", 105, 69, 24, 24, 2356},
	{"steep_slope_down", 9, 96, 24, 24, 2348},
	{"slope_down", 33, 96, 24, 24, 2349},
	{"level", 57, 96, 24, 24, 2350},
	{"slope_up", 81, 96, 24, 24, 2351},
	{"steep_slope_up", 105, 96, 24, 24, 2352},
}

// RoadTraitFlags (World::Track::RoadTraitFlags).
const (
	RoadSmallCurve     = 1 << 0
	RoadVerySmallCurve = 1 << 1
	RoadSlope          = 1 << 2
	RoadSteepSlope     = 1 << 3
	RoadTurnaround     = 1 << 5
)

// RoadButtons is activateSelectedRoadWidgets' layout: only small and very
// small curves (normal and large curves and s-bends stay hidden), straight,
// the turnaround in the first s-bend slot (construction_left_turnaround when
// trafficHandedness is 0, else the right one), slopes by the road's traits.
func RoadButtons(traits uint16, trafficHandedness uint8) []ConstructionButton {
	var out []ConstructionButton
	vs := traits&RoadVerySmallCurve != 0
	for _, b := range constructionButtons {
		show := false
		switch b.Name {
		case "left_hand_curve_very_small", "right_hand_curve_very_small":
			show = vs
		case "left_hand_curve_small", "right_hand_curve_small":
			show = traits&RoadSmallCurve != 0
		case "straight", "level":
			show = true
		case "s_bend_dual_track_left":
			show = traits&RoadTurnaround != 0
			b.Name, b.Image = "turnaround", imgConstructionRightTurnaround
			if trafficHandedness == 0 {
				b.Image = imgConstructionLeftTurnaround
			}
		case "slope_down", "slope_up":
			show = traits&RoadSlope != 0
		case "steep_slope_down", "steep_slope_up":
			show = traits&RoadSteepSlope != 0
		}
		if !show {
			continue
		}
		if vs {
			switch b.Name {
			case "left_hand_curve_small":
				b.X = 25
			case "right_hand_curve_small":
				b.X = 91
			}
		}
		out = append(out, b)
	}
	return out
}

// Buttons is the visible button set for the state (track or road).
func (st ConstructionTabState) Buttons() []ConstructionButton {
	if st.IsRoad {
		return RoadButtons(st.Traits, st.TrafficHandedness)
	}
	return TrackButtons(st.Traits)
}

// ConstructionTabState is what activateSelected{Track,Road}Widgets read.
type ConstructionTabState struct {
	IsRoad            bool
	TrafficHandedness uint8
	Traits            uint16 // the track object's TrackPieces or the road's RoadPieces
	Piece             string // the activated piece widget ("straight", …); "" none
	Gradient          string // the activated gradient widget ("level", …)
	Hover             bool   // constructionHover: placing the first piece
}

// TrackButtons is activateSelectedTrackWidgets' layout for a track: the
// visible piece and gradient buttons with their positions.
func TrackButtons(traits uint16) []ConstructionButton {
	var out []ConstructionButton
	vs := traits&TraitVerySmallCurve != 0
	for _, b := range constructionButtons {
		show := true
		switch b.Name {
		case "left_hand_curve_very_small", "right_hand_curve_very_small":
			show = vs
		case "left_hand_curve_large", "right_hand_curve_large":
			show = traits&TraitLargeCurve != 0
		case "left_hand_curve", "right_hand_curve":
			show = traits&TraitNormalCurve != 0
		case "left_hand_curve_small", "right_hand_curve_small":
			show = traits&TraitSmallCurve != 0
		case "s_bend_dual_track_left", "s_bend_dual_track_right":
			show = traits&TraitOneSided != 0
		case "slope_down", "slope_up":
			show = traits&TraitSlope != 0
		case "steep_slope_down", "steep_slope_up":
			show = traits&TraitSteepSlope != 0
		}
		if !show {
			continue
		}
		if vs { // very small curves take the outer slots, the others move in
			switch b.Name {
			case "left_hand_curve_small":
				b.X = 25
			case "right_hand_curve_small":
				b.X = 91
			case "left_hand_curve":
				b.X = 47
			case "right_hand_curve":
				b.X = 69
			}
		}
		out = append(out, b)
	}
	return out
}

// DrawConstructionTab draws the construction tab's buttons inside the frame f
// (ImageButton::draw: an activated button gets an inset background in the
// secondary colour, every visible button its image recoloured to that colour).
// PARTIAL: the bridge dropdown, the construct/preview area (Wt3Widget with the
// track piece drawn by the paint code), remove/rotate/copy/paste, hover.
func DrawConstructionTab(screen *ebiten.Image, f *ConstructionFrame, st ConstructionTabState) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(f.SkinColour))
	if err != nil {
		return
	}
	draw8(screen, f.X, f.Y, ConstructionW, ConstructionH, func(rt *gfx8.RenderTarget) {
		// The construct area (Wt3Widget 3,145 132×100): a raised inset in the
		// window colour; the piece preview is drawn into it by the caller.
		gfx8.FillRectInset(rt, 3, 145, 3+132-1, 145+100-1, gfx8.InsetColour{Map: sec}, 0)
		// Bottom row: Copy, Paste (Buttons), Remove (only while extending),
		// Rotate (rotate_object while placing the first piece, else
		// construction_new_position).
		gfx8.FillRectInset(rt, 3, 248, 3+33-1, 248+24-1, gfx8.InsetColour{Map: sec}, 0)
		gfx8.FillRectInset(rt, 36, 248, 36+37-1, 248+24-1, gfx8.InsetColour{Map: sec}, 0)
		icons := [][3]int{{111, 248, imgConstructionNewPosition}}
		if st.Hover {
			icons[0][2] = imgRotateObject
		} else {
			icons = append(icons, [3]int{72, 248, imgConstructionRemove})
		}
		for _, ic := range icons {
			if sp, err := gfx8.Decode(&g1.Elements[ic[2]]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, ic[0], ic[1], gfx8.OpFor(sp, true, false), sec, nil)
			}
		}
		for _, b := range st.Buttons() {
			if b.Name == st.Piece || b.Name == st.Gradient {
				gfx8.FillRectInset(rt, b.X, b.Y, b.X+b.W-1, b.Y+b.H-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset)
			}
			if sp, err := gfx8.Decode(&g1.Elements[b.Image]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, b.X, b.Y, gfx8.OpFor(sp, true, false), sec, nil)
			}
		}
	})
	// Button labels (Label::draw centred, opaque window colour).
	pal := g1.GetPalette()
	text := pal[gfx8.Shade(sec, 9)]
	// drawCostString: "Build this..." centred at the construct widget's midX,
	// its bottom - 23, while extending (not while placing the first piece).
	// PARTIAL: the "Cost: …" line 11 px below (goloco has no construction cost
	// yet: the CreateTrack/CreateRoad game commands' cost isn't ported).
	if !st.Hover {
		const midX, y = (3 + 3 + 132 - 1) / 2, 145 + 100 - 1 - 23
		bw, _ := MeasureTextBold("Build this...")
		DrawTextBold(screen, "Build this...", f.X+midX-bw/2, f.Y+y, text)
	}
	for _, l := range []struct {
		x, w int
		s    string
	}{{3, 33, "Copy"}, {36, 37, "Paste"}} {
		tw, _ := MeasureTextBold(l.s)
		DrawTextBold(screen, l.s, f.X+l.x+(l.w-1)/2-tw/2, f.Y+248+6, text)
	}
}

// Construction tab image ids (ImageIds.h).
const (
	imgConstructionRemove      = 2361
	imgConstructionNewPosition = 2362
	imgRotateObject            = 2365

	imgConstructionRightTurnaround = 2359
	imgConstructionLeftTurnaround  = 2360
)

// ConstructionHit returns the construction tab widget under (rx, ry) relative
// to the window: a piece/gradient button name, "remove", "rotate", "construct",
// "copy", "paste", or "" (tabs and the close button are the caller's).
func ConstructionHit(rx, ry int, st ConstructionTabState) string {
	in := func(x, y, w, h int) bool { return rx >= x && rx < x+w && ry >= y && ry < y+h }
	for _, b := range st.Buttons() {
		if in(b.X, b.Y, b.W, b.H) {
			return b.Name
		}
	}
	switch {
	case in(3, 145, 132, 100):
		return "construct"
	case in(111, 248, 24, 24):
		return "rotate"
	case !st.Hover && in(72, 248, 39, 24):
		return "remove"
	case in(3, 248, 33, 24):
		return "copy"
	case in(36, 248, 37, 24):
		return "paste"
	}
	return ""
}
