package ui

import (
	"image/color"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// SignalTabW / SignalTabH are the signal tab's window size (SignalTab.cpp
// kWidth, kHeight: 04e7900 / 3d6c318's taller tab with the repeated
// placement).
const SignalTabW, SignalTabH = 156, 222

// SignalTabState is what the signal tab shows (Construction/SignalTab.cpp).
type SignalTabState struct {
	Name, Description string // the selected TrainSignalObject's strings
	Image             int    // its image (TrainSignalObject::image)
	Both              bool   // isSignalBothDirections: which button's tool is on
	Repeated          bool   // repeatedSignalMode: the checkbox
	Step              int    // signalPlacementStepSize
	Cost              int32  // signalCost (0 or less: not shown)
}

// Signal tab widgets (SignalTab.cpp widgets: kSpacing 4, kImageButtonSize 40,
// kStepperWidth 48; the stepper's buttons by makeStepperDecrease /
// IncreaseWidget).
var (
	sigDropdown = [4]int{4, 45, SignalTabW - 8, 12}
	sigBoth     = [4]int{(SignalTabW-8)/2 - 4 - 40, 110, 40, 40}
	sigSingle   = [4]int{(SignalTabW-8)/2 + 4, 110, 40, 40}
	sigRepeat   = [4]int{4, 166, SignalTabW - 8, 12}
	sigStepText = [4]int{4, 182, SignalTabW - 8, 13}
	sigStep     = [4]int{SignalTabW - 4 - 48, 182, 48, 12}
	sigStepDec  = [4]int{SignalTabW - 4 - 48 + 48 - 26, 183, 13, 10}
	sigStepInc  = [4]int{SignalTabW - 4 - 48 + 48 - 13, 183, 12, 10}
)

// DrawSignalTab is Signal::draw (0x0049E501) after the frame: the signal
// dropdown with the selected signal's name, its description wrapped to 130 px
// at (3, 63), the two direction buttons (the one whose tool is on drawn
// pressed, ImageButton activated) and the signal images under them: both
// directions = image at midX − 8 and image + 4 at midX + 8, single = image at
// midX, each at the button's bottom − 4; the separators under the buttons and
// above the cost; the "Repeated placement" checkbox, the step label and the
// stepper (disabled — inset — without repeated placement, prepareDraw); the
// cost centred at the bottom when there is one.
// PARTIAL: the hover look (draw_3).
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
	draw8(screen, f.X, f.Y, SignalTabW, f.Height(), func(rt *gfx8.RenderTarget) {
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
		for _, y := range []int{bottom(sigSingle) + 9, SignalTabH - 12 - 9} { // drawSeparator: drawRectInset(3, y, width − 7, 1)
			gfx8.FillRectInset(rt, 3, y, 3+(SignalTabW-7)-1, y, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset)
		}
		// Checkbox::drawCheckBox: a 10 × 10 box (inset, fillDarker); the
		// stepper: TextBox (inset, fillDarker) and its two Buttons.
		gfx8.FillRectInset(rt, sigRepeat[0], sigRepeat[1], sigRepeat[0]+9, sigRepeat[1]+9, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		gfx8.FillRectInset(rt, sigStep[0], sigStep[1], sigStep[0]+sigStep[2]-1, sigStep[1]+sigStep[3]-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		for _, b := range [][4]int{sigStepDec, sigStepInc} {
			gfx8.FillRectInset(rt, b[0], b[1], b[0]+b[2]-1, b[1]+b[3]-1, gfx8.InsetColour{Map: sec}, 0)
		}
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
	// The checkbox's tick (0xAC) and label, 4 past the box, in the window's
	// colour; the step label, value and stepper buttons — inset while
	// disabled.
	wc, _ := TextColours(int(f.SkinColour))
	if st.Repeated {
		DrawTextBold(screen, string(rune(0xAC)), f.X+sigRepeat[0], f.Y+sigRepeat[1], wc)
	}
	DrawTextBold(screen, i18n.Str(i18n.SignalPlacementRepeat), f.X+sigRepeat[0]+10+4, f.Y+sigRepeat[1], wc)
	text := func(s string, x, y int, c color.Color) {
		if st.Repeated {
			DrawTextBold(screen, s, x, y, c)
		} else {
			DrawTextInset(screen, s, x, y, sec)
		}
	}
	text(i18n.Str(i18n.SignalPlacementStepSize), f.X+sigStepText[0], f.Y+sigStepText[1], wc)
	text(strconv.Itoa(st.Step), f.X+sigStep[0]+1, f.Y+sigStep[1]+1, black) // uint16_raw
	for _, s := range []struct {
		b  [4]int
		id int
	}{{sigStepDec, i18n.StepperMinus}, {sigStepInc, i18n.StepperPlus}} { // Button: its text centred
		t := stripCodes(i18n.Str(s.id))
		tw, _ := MeasureTextBold(t)
		text(t, f.X+(s.b[0]+s.b[0]+s.b[2])/2-tw/2, f.Y+s.b[1], black)
	}
	if st.Cost > 0 { // build_cost "{COLOUR WINDOW_2}Cost: {COLOUR BLACK}{CURRENCY32}", centred at (kWidth / 2, kHeight − kSpacing − 12)
		label, amount := stripCodes(strings.SplitN(i18n.Str(i18n.BuildCost), "{COLOUR BLACK}", 2)[0]), FormatCurrency(int64(st.Cost))
		lw, _ := MeasureTextBold(label)
		tw, _ := MeasureTextBold(amount)
		x := f.X + SignalTabW/2 - (lw+tw)/2
		DrawTextBold(screen, label, x, f.Y+SignalTabH-4-12, pal[gfx8.Shade(sec, 9)])
		DrawTextBold(screen, amount, x+lw, f.Y+SignalTabH-4-12, black)
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

// SignalHit returns "both", "single", "dropdown", "repeat", "stepdec",
// "stepinc" or "" for a click at (rx, ry) relative to the window.
func SignalHit(rx, ry int) string {
	in := func(r [4]int) bool { return rx >= r[0] && rx < r[0]+r[2] && ry >= r[1] && ry < r[1]+r[3] }
	switch {
	case in(sigBoth):
		return "both"
	case in(sigSingle):
		return "single"
	case in(sigDropdown):
		return "dropdown"
	case in(sigRepeat):
		return "repeat"
	case in(sigStepDec):
		return "stepdec"
	case in(sigStepInc):
		return "stepinc"
	}
	return ""
}
