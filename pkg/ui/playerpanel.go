package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"strconv"
)

// colourRed is Colour::red ({COLOUR RED} in the negative-cash string).
const colourRed = 27

// PlayerPanelState is what PlayerInfoPanel.cpp draws from.
type PlayerPanelState struct {
	Cash       int64
	Bankrupt   bool
	HoverCash  bool
	SkinColour uint8 // InterfaceSkinObject::playerInfoToolbarColour
	// The controlling company's owner image (competitor images[ownerEmotion],
	// 0 = none) in its primary colour, performanceIndex, and the trend from
	// its challengeFlags (+1 increased, −1 decreased).
	Face        int
	FaceColour  uint8
	Performance int16
	Trend       int
	HoverPerf   bool
}

// int16OneDecimal is formatShortWithOneDecimal ({INT16_1DP}): value / 10,
// ".", value % 10 (C++ truncation; the C locale, so no grouping).
func int16OneDecimal(v int16) string { return strconv.Itoa(int(v/10)) + "." + strconv.Itoa(int(v%10)) }

// DrawPlayerPanel is PlayerInfoPanel.cpp draw: the translucent bottom-left
// window, cash centred on frame.width/2 + 12 in the top row ({CURRENCY48}, red
// when negative, "Bankrupt!" when bankrupt).
// Then the owner's image at (frame + 2, frame + 2) and the performance row
// centred at frame top + 14 ("{INT16_1DP}%", or with nine spaces and the
// ▲/▼ sprite 2324/2325 when the company's performance went up/down).
func DrawPlayerPanel(screen *ebiten.Image, s PlayerPanelState) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	wx, wy := L.PlayerPanelPos(sw, sh)
	cmap, err := g1.GetPaletteMap(int(s.SkinColour))
	if err != nil {
		return
	}
	var trans [3][]uint8
	for k := range trans {
		trans[k], _ = g1.GetPaletteMap(int(paintTranslucent(s.SkinColour, k)))
	}
	tc := gfx8.InsetColour{Map: cmap, Translucent: true, Trans: trans}
	m34, _ := g1.GetPaletteMap(extColourUnk34)
	draw8(screen, wx, wy, L.PlayerPanelW, L.PlayerPanelH, func(rt *gfx8.RenderTarget) {
		o := L.TimePanelOuter
		gfx8.FillRectTranslucent(rt, o[0], o[1], o[0]+o[2]-1, o[1]+o[3]-1, m34)
		for _, r := range [][4]int{L.TimePanelOuter, L.TimePanelInner} {
			gfx8.FillRectInset(rt, r[0], r[1], r[0]+r[2]-1, r[1]+r[3]-1, tc, 0)
		}
		gfx8.FillRectInset(rt, o[0]+1, o[1]+1, o[0]+o[2]-2, o[1]+o[3]-2, tc, gfx8.InsetBorderInset|gfx8.InsetFillNone)
		if s.Face > 0 && s.Face < len(g1.Elements) { // the owner's image, recoloured
			if sp, err := gfx8.Decode(&g1.Elements[s.Face]); err == nil {
				m, _ := g1.GetPaletteMap(int(s.FaceColour))
				gfx8.DrawSpriteNoise(rt, sp, o[0]+2, o[1]+2, gfx8.OpFor(sp, m != nil, false), m, nil)
			}
		}
	})
	pal := g1.GetPalette()
	text, textMap := L.Currency48(s.Cash), cmap
	if s.Bankrupt {
		text = "Bankrupt!"
	}
	if s.Bankrupt || s.Cash < 0 {
		if rm, err := g1.GetPaletteMap(colourRed); err == nil {
			textMap = rm
		}
	}
	c := pal[gfx8.Shade(textMap, 9)]
	if s.HoverCash && !s.Bankrupt && s.Cash >= 0 {
		if wm, err := g1.GetPaletteMap(colourWhite); err == nil {
			c = pal[gfx8.Shade(wm, 9)]
		}
	}
	x := wx + L.TimePanelOuter[2]/2 + 12
	tw, _ := MeasureTextBold(text)
	DrawTextBold(screen, text, x-tw/2, wy+L.TimePanelOuter[1]+2, c)
	if s.Face == 0 { // no company record (the title's preview panel)
		return
	}
	perf := int16OneDecimal(s.Performance) + "%"
	sprite := 0
	switch {
	case s.Trend > 0: // player_info_performance_increase
		perf, sprite = "         "+perf+" ", 2324
	case s.Trend < 0:
		perf, sprite = "         "+perf+" ", 2325
	}
	pc := pal[gfx8.Shade(cmap, 9)] // the window's primary colour, opaque
	if s.HoverPerf {
		if wm, err := g1.GetPaletteMap(colourWhite); err == nil {
			pc = pal[gfx8.Shade(wm, 9)]
		}
	}
	pw, _ := MeasureTextBold(perf)
	sw2 := 0
	if sprite > 0 && sprite < len(g1.Elements) {
		sw2 = int(g1.Elements[sprite].Width)
	}
	px, py := x-(pw+sw2)/2, wy+L.TimePanelOuter[1]+14 // drawStringCentred: the whole string incl. the sprite
	DrawTextBold(screen, perf, px, py, pc)
	if sprite > 0 {
		draw8(screen, px+pw, py, sw2+2, 12, func(rt *gfx8.RenderTarget) { drawG1At(rt, sprite, 0, 0) })
	}
}
