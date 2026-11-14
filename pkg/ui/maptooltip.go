package ui

import (
	"image/color"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// The map tooltip (Ui/Windows/MapToolTip.cpp): the text under the mouse over
// the map — what getItemLeft / rightOver found there, or the map selection's
// size. A window 240 × 55 without background: its text outlined
// (outlined_wcolour2_stringid) and centred, wrapped, in the interface skin's
// mapTooltipObjectColour (WINDOW_2), switching to mapTooltipCargoColour after
// MapTipColour3 ({COLOUR WINDOW_3}). Another company's item: the owner's face
// on black, left of the text.

// MapTipColour3 in a map tooltip's text is {COLOUR WINDOW_3}: the rest in the
// cargo colour.
const MapTipColour3 = '\x0E'

// Map tooltip window size (MapToolTip::open).
const (
	MapTipW = 240
	MapTipH = 55
)

// MapTipPos is open's placement for the mouse at (mx, my) on a sw × sh
// screen: 15 below the cursor (above it, 19 more, where that would leave
// the screen), centred on it within the screen.
func MapTipPos(mx, my, sw, sh int) (x, y int) {
	y = my + 15
	if y > sh-MapTipH {
		y -= MapTipH + 19
	}
	if MapTipW <= sw {
		x = max(0, min(sw-MapTipW, mx-MapTipW/2))
	}
	return x, y
}

// MapTip is what MapToolTip::draw shows: Text (newlines as controlNewline,
// MapTipColour3 as the colour switch) and, for another company's item,
// Face (its competitor image, 0: none) in FaceColour.
type MapTip struct {
	Text                      string
	Owned                     bool // another company's: the face drawn
	Face                      int
	FaceColour                uint8
	ObjectColour, CargoColour uint8 // the skin's mapTooltipObjectColour / mapTooltipCargoColour
}

// DrawMapTip is MapToolTip::draw for the window at (x, y).
func DrawMapTip(screen *ebiten.Image, t MapTip, x, y int) {
	if globalG1Font == nil || t.Text == "" {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	objMap, err1 := g1.GetPaletteMap(int(t.ObjectColour))
	cargoMap, err2 := g1.GetPaletteMap(int(t.CargoColour))
	if err1 != nil || err2 != nil {
		return
	}
	colours := [2]color.Color{pal[gfx8.Shade(objMap, 9)], pal[gfx8.Shade(cargoMap, 9)]} // getShade(hue, 9)
	outline := pal[paletteBlack0]
	if !t.Owned {
		drawMapTipText(screen, t.Text, x+MapTipW/2, y+MapTipH/2-5, MapTipW, colours, outline)
		return
	}
	left := drawMapTipText(screen, t.Text, x+MapTipW/2+13, y+MapTipH/2-5, MapTipW-28, colours, outline) - 28
	top := y + MapTipH/2 - 13
	draw8(screen, left, top, 26, 26, func(rt *gfx8.RenderTarget) {
		gfx8.FillRect(rt, 0, 0, 25, 25, paletteBlack0)
		if t.Face > 0 && t.Face < len(g1.Elements) {
			if sp, err := gfx8.Decode(&g1.Elements[t.Face]); err == nil {
				m, _ := g1.GetPaletteMap(int(t.FaceColour))
				gfx8.DrawSpriteNoise(rt, sp, 1, 1, gfx8.OpFor(sp, m != nil, false), m, nil)
			}
		}
	})
}

// drawMapTipText is drawStringCentredWrapped in medium bold, outlined, its
// colour switching at MapTipColour3 (lasting over the lines after): the
// lines centred on x, the block on y; the left x of the widest line.
func drawMapTipText(screen *ebiten.Image, text string, x, y, width int, colours [2]color.Color, outline color.Color) int {
	const font = G1FontMediumBold
	lines := WrapString(text, font, width)
	lh := g1FontHeight(font)
	y -= (lh / 2) * (len(lines) - 1)
	cur, maxW := 0, 0
	for _, l := range lines {
		plain := strings.ReplaceAll(l, string(MapTipColour3), "")
		lw := MeasureTextFont(plain, font)
		maxW = max(maxW, lw)
		px := x - lw/2
		for i, seg := range strings.Split(l, string(MapTipColour3)) {
			if i > 0 {
				cur = 1
			}
			px += DrawTextOutlinedFont(screen, seg, font, px, y, colours[cur], outline)
		}
		y += lh
	}
	return x - maxW/2
}
