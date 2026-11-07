package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// FillRectCrossHatch is drawRectImpl with RectFlags::crossHatching: every
// other pixel, the pattern alternating per row (starting phase 0, as for a
// rectangle not clipped at its top-left).
func fillRectCrossHatch(rt *gfx8.RenderTarget, l, t, r, b int, c uint8) {
	cross := uint32(0)
	for y := t; y <= b; y++ {
		p := cross>>1 | cross<<31 // std::rotr(crossPattern, 1)
		for x := l; x <= r; x++ {
			p ^= 0x80000000
			if p&0x80000000 != 0 && x >= 0 && x < rt.Width && y >= 0 && y < rt.Height {
				rt.Bits[y*rt.Width+x] = c
			}
		}
		cross ^= 1
	}
}

// drawVScrollView draws ScrollView::draw's frame (inset, fillDarker) for a
// widget at x, y (target coordinates) of w × h, and — when the content is
// taller — drawVScroll: the hatched track with its lines, the up/down buttons
// and the thumb. It returns the content rectangle (1 px inset, minus the bar).
// PARTIAL: pressed states of the buttons/thumb; the arrow glyphs are drawn by
// the caller (drawVScrollArrows) since text goes through ebiten.
func drawVScrollView(rt *gfx8.RenderTarget, x, y, w, h int, c gfx8.InsetColour, contentH, offset int) (cx, cy, cw, ch int, bar bool) {
	return drawVScrollViewPressed(rt, x, y, w, h, c, contentH, offset, "")
}

// drawVScrollViewPressed is drawVScrollView with a part drawn pressed
// (vscrollbarUpPressed / DownPressed / ThumbPressed: borderInset).
func drawVScrollViewPressed(rt *gfx8.RenderTarget, x, y, w, h int, c gfx8.InsetColour, contentH, offset int, pressed string) (cx, cy, cw, ch int, bar bool) {
	pf := func(part string) int {
		if part == pressed {
			return gfx8.InsetBorderInset
		}
		return 0
	}
	gfx8.FillRectInset(rt, x, y, x+w-1, y+h-1, c, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
	cx, cy, cw, ch = x+L.ScrollbarMargin, y+L.ScrollbarMargin, w-2*L.ScrollbarMargin, h-2*L.ScrollbarMargin
	if contentH <= h {
		return cx, cy, cw, ch, false
	}
	sx, sy := x+w-L.ScrollbarSize-L.ScrollbarMargin, y+L.ScrollbarMargin
	sh := h - 2*L.ScrollbarMargin
	s3, s7 := gfx8.Shade(c.Map, 3), gfx8.Shade(c.Map, 7)
	gfx8.FillRect(rt, sx, sy, sx+L.ScrollbarSize-1, sy+sh-1, s7)
	fillRectCrossHatch(rt, sx, sy, sx+L.ScrollbarSize-1, sy+sh-1, s3)
	for _, ln := range [][2]int{{2, int(s3)}, {3, int(s7)}, {7, int(s3)}, {8, int(s7)}} {
		gfx8.FillRect(rt, sx+ln[0], sy, sx+ln[0], sy+sh-1, uint8(ln[1]))
	}
	gfx8.FillRectInset(rt, sx, sy, sx+L.ScrollbarSize-1, sy+L.ScrollbarSize-1, c, pf("up"))
	gfx8.FillRectInset(rt, sx, sy+sh-L.ScrollbarSize, sx+L.ScrollbarSize-1, sy+sh-1, c, pf("down"))
	top, bot := L.VThumb(h, contentH, offset)
	gfx8.FillRectInset(rt, sx, y+top, sx+L.ScrollbarSize-1, y+top+(bot-top-L.ScrollbarMargin)-1, c, pf("thumb"))
	return cx, cy, cw - L.ScrollbarSize, ch, true
}

// drawVScrollArrows draws the up/down arrow glyphs (font characters 0xA0 and
// 0xAA, black) on a scroll view's buttons; x, y, w, h are screen coordinates
// of the widget.
func drawVScrollArrows(screen *ebiten.Image, x, y, w, h int) {
	if globalG1Font == nil {
		return
	}
	pal := globalG1Font.g1.GetPalette()
	sx, sy := x+w-L.ScrollbarSize-L.ScrollbarMargin, y+L.ScrollbarMargin
	sh := h - 2*L.ScrollbarMargin
	DrawTextBold(screen, string(rune(0xA0)), sx+1, sy-1, pal[paletteBlack0])
	DrawTextBold(screen, string(rune(0xAA)), sx+L.ScrollbarMargin, sy+sh-L.ScrollbarSize, pal[paletteBlack0])
}
