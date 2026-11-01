package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// colourRed is Colour::red ({COLOUR RED} in the negative-cash string).
const colourRed = 27

// PlayerPanelState is what PlayerInfoPanel.cpp draws from.
type PlayerPanelState struct {
	Cash       int64
	Bankrupt   bool
	HoverCash  bool
	SkinColour uint8 // InterfaceSkinObject::playerInfoToolbarColour
}

// DrawPlayerPanel is PlayerInfoPanel.cpp draw: the translucent bottom-left
// window, cash centred on frame.width/2 + 12 in the top row ({CURRENCY48}, red
// when negative, "Bankrupt!" when bankrupt).
// PARTIAL: the owner's face (CompetitorObject images by emotion, recoloured to
// the company colour) and the performance index row are not drawn — goloco has
// no player company record or performance index yet (game logic, item 6).
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
}
