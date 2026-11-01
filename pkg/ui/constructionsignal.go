package ui

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// SignalTabH is the signal tab's window height (Construction::Signal
// makeCommonWidgets(138, 167)).
const SignalTabH = 167

// SignalTabState is what the signal tab shows (Construction/SignalTab.cpp).
type SignalTabState struct {
	Name, Description string // the selected TrainSignalObject's strings
	Image             int    // its image (TrainSignalObject::image)
	Both              bool   // isSignalBothDirections: which button's tool is on
}

// Signal tab widgets (SignalTab.cpp widgets).
var (
	sigDropdown = [4]int{3, 45, 132, 12}
	sigBoth     = [4]int{27, 110, 40, 40}
	sigSingle   = [4]int{71, 110, 40, 40}
)

// DrawSignalTab is Signal::draw (0x0049E501) after the frame: the signal
// dropdown with the selected signal's name, its description wrapped to 130 px
// at (3, 63), the two direction buttons (the one whose tool is on drawn
// pressed, ImageButton activated) and the signal images under them: both
// directions = image at midX − 8 and image + 4 at midX + 8, single = image at
// midX, each at the button's bottom − 4.
// PARTIAL: the cost line (no signal cost: CreateSignal isn't ported), the
// dropdown list, the hover look (draw_3).
func DrawSignalTab(screen *ebiten.Image, f *ConstructionFrame, st SignalTabState) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sec, err := g1.GetPaletteMap(int(f.SkinColour))
	if err != nil {
		return
	}
	pal := g1.GetPalette()
	black := pal[paletteBlack0]
	draw8(screen, f.X, f.Y, ConstructionW, f.Height(), func(rt *gfx8.RenderTarget) {
		optDropdown(rt, sigDropdown[0], sigDropdown[1], sigDropdown[2], sec)
		on := sigSingle
		if st.Both {
			on = sigBoth
		}
		gfx8.FillRectInset(rt, on[0], on[1], on[0]+on[2]-1, on[1]+on[3]-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset)
		img := func(id, x, y int) {
			if id >= 0 && id < len(g1.Elements) {
				if sp, err := gfx8.Decode(&g1.Elements[id]); err == nil {
					gfx8.DrawSpriteNoise(rt, sp, x, y, gfx8.OpFor(sp, false, false), nil, nil)
				}
			}
		}
		mid := func(r [4]int) int { return (r[0] + r[0] + r[2] - 1) / 2 }
		bottom := func(r [4]int) int { return r[1] + r[3] - 1 }
		img(st.Image, mid(sigBoth)-8, bottom(sigBoth)-4)
		img(st.Image+4, mid(sigBoth)+8, bottom(sigBoth)-4)
		img(st.Image, mid(sigSingle), bottom(sigSingle)-4)
	})
	// The dropdown's text (drawDropdownContent) and its ▼ button.
	DrawTextBold(screen, st.Name, f.X+sigDropdown[0]+1, f.Y+sigDropdown[1]+1, black)
	arrow := string(rune(170))
	aw, _ := MeasureTextBold(arrow)
	bx := sigDropdown[0] + sigDropdown[2] - 12
	DrawTextBold(screen, arrow, f.X+bx+(11-1)/2-aw/2, f.Y+sigDropdown[1]+1, black)
	// signal_black: the description, drawStringLeftWrapped(130).
	for i, line := range WrapTextBold(st.Description, 130) {
		DrawTextBold(screen, line, f.X+3, f.Y+63+10*i, black)
	}
}

// WrapTextBold breaks str into lines no wider than width in the bold font,
// at spaces (drawStringLeftWrapped's word wrap), and at ControlCodes::newline
// (5), which object strings use ("Colourlight Signal,\x05American style").
func WrapTextBold(str string, width int) []string {
	var lines []string
	for _, part := range strings.Split(str, "\x05") {
		lines = append(lines, wrapWords(part, width)...)
	}
	return lines
}

func wrapWords(str string, width int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(str) {
		try := word
		if line != "" {
			try = line + " " + word
		}
		if w, _ := MeasureTextBold(try); w > width && line != "" {
			lines = append(lines, line)
			line = word
			continue
		}
		line = try
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// SignalHit returns "both", "single", "dropdown" or "" for a click at (rx, ry)
// relative to the window.
func SignalHit(rx, ry int) string {
	in := func(r [4]int) bool { return rx >= r[0] && rx < r[0]+r[2] && ry >= r[1] && ry < r[1]+r[3] }
	switch {
	case in(sigBoth):
		return "both"
	case in(sigSingle):
		return "single"
	case in(sigDropdown):
		return "dropdown"
	}
	return ""
}
