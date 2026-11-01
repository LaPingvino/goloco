package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/hajimehoshi/ebiten/v2"
)

// Window chrome — ports of Ui/Widgets/FrameWidget.cpp (default style:
// background) and CaptionWidget.cpp drawSimple (whiteText).

// ImageIds::frame_background_image / _alt and PaletteIndex::black0 (from the
// compiler against OpenLoco b93e074).
const (
	imgFrameBackground    = 2322
	imgFrameBackgroundAlt = 2323
	paletteBlack0         = 10
)

// DrawWindowFrame is Frame::draw with WindowFrameStyle::background: the frame
// background image recoloured to the window colour, repeated from the right
// across the window width and clipped to the top 41 px, then a 1-px shadow
// (shade 3, or shade 1 with WindowFlags::lighterFrame) down the right edge.
// PARTIAL: drawResizeHandle (resizable windows) is not drawn.
// FrameStyle is Config windowFrameStyle: 0 background (the gradient image,
// default), 1 solid, 2 transparent (Options' Rendering tab sets it).
var FrameStyle int

func DrawWindowFrame(screen *ebiten.Image, x, y, w, h int, colour uint8, lighterFrame bool) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	cmap, err := g1.GetPaletteMap(int(colour))
	if err != nil {
		return
	}
	id := imgFrameBackgroundAlt
	if lighterFrame {
		id = imgFrameBackground
	}
	sp, err := gfx8.Decode(&g1.Elements[id])
	if err != nil {
		return
	}
	hh := min(41, h)
	draw8(screen, x, y, w, hh, func(rt *gfx8.RenderTarget) {
		// Frame::draw by Config windowFrameStyle: the background image
		// (drawBackground), a solid inset (drawSolid) or a translucent one
		// (drawTransparent). goloco draws the frame's top 41 px here (the rest
		// is under the window's panel).
		// PARTIAL: with the transparent style upstream also makes the whole
		// window translucent (Window::isTranslucent).
		switch FrameStyle {
		case 1:
			gfx8.FillRectInset(rt, 0, 0, w-1, hh-1, gfx8.InsetColour{Map: cmap}, 0)
		case 2:
			var trans [3][]uint8
			if int(colour) < len(paint.TranslucentColour) {
				for k := range trans {
					trans[k], _ = g1.GetPaletteMap(int(paint.TranslucentColour[colour][k]))
				}
			}
			gfx8.FillRectInset(rt, 0, 0, w-1, hh-1, gfx8.InsetColour{Map: cmap, Translucent: true, Trans: trans}, 0)
		default:
			passes := (w + sp.Width - 1) / sp.Width
			for i := passes; i >= 0; i-- {
				gfx8.DrawSpriteNoise(rt, sp, i*(sp.Width-1), 0, gfx8.OpTransparent|gfx8.OpSrc, cmap, nil)
			}
		}
		shade := gfx8.Shade(cmap, 3)
		if lighterFrame {
			shade = gfx8.Shade(cmap, 1)
		}
		gfx8.FillRect(rt, w-1, 0, w-1, 40, shade)
	})
}

// DrawWindowCaption is Caption::drawSimple with Style::whiteText for a caption widget
// at x, y of width w: medium bold, clipped to (w - 4 - 14) - 8, centred on
// 2 + (w - 18) / 2, 1 px down, white text with the outline style (black0).
// PARTIAL: clipping truncates without upstream's "..." (TextRenderer::clipString).
func DrawWindowCaption(screen *ebiten.Image, x, y, w int, text string) {
	DrawWindowCaptionColour(screen, x, y, w, text, colourWhite)
}

// DrawWindowCaptionColour is Caption::drawSimple for a given text colour:
// Style::whiteText is colourWhite, Style::colourText the window's primary
// colour (ControlCodes::windowColour1). Outline style (black0) either way.
func DrawWindowCaptionColour(screen *ebiten.Image, x, y, w int, text string, colour int) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	wm, err := g1.GetPaletteMap(colour)
	if err != nil {
		return
	}
	width := w - 4 - 14
	for len(text) > 0 {
		if tw, _ := MeasureTextBold(text); tw <= width-8 {
			break
		}
		text = text[:len(text)-1]
	}
	sw, _ := MeasureTextBold(text)
	px := x + 2 + width/2 - (sw-1)/2
	DrawTextOutlined(screen, text, px, y+1, pal[gfx8.Shade(wm, 9)], pal[paletteBlack0])
}
