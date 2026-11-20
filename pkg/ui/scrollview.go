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

// drawVScrollView draws ScrollView::draw for a widget with a vertical bar
// only (Scrollbars::vertical) at x, y (target coordinates) of w × h, content
// contentH tall scrolled to offset: drawScrollView. It returns the content
// rectangle drawScroll draws in and whether the bar shows. The arrow glyphs
// are drawn by the caller (drawVScrollArrows) since text goes through ebiten.
func drawVScrollView(rt *gfx8.RenderTarget, x, y, w, h int, c gfx8.InsetColour, contentH, offset int) (cx, cy, cw, ch int, bar bool) {
	return drawVScrollViewPressed(rt, x, y, w, h, c, contentH, offset, "")
}

// vPressed maps the vertical bar's part names (VScrollHold) to ScrollParts.
var vPressed = map[string]L.Part{"up": L.PartVTop, "down": L.PartVBottom, "thumb": L.PartVThumb}

// drawVScrollViewPressed is drawVScrollView with a part drawn pressed
// (vscrollbarUpPressed / DownPressed / ThumbPressed: borderInset).
func drawVScrollViewPressed(rt *gfx8.RenderTarget, x, y, w, h int, c gfx8.InsetColour, contentH, offset int, pressed string) (cx, cy, cw, ch int, bar bool) {
	b := L.Bars{W: w, H: h, ContentH: contentH, VBar: true}
	cx, cy, cw, ch = drawScrollView(rt, x, y, b, c, 0, offset, vPressed[pressed])
	return cx, cy, cw, ch, b.VShown()
}

// drawScrollView is ScrollView::draw (ScrollViewWidget.cpp, 0x004CAB58) for
// the widget at x, y (target coordinates) with bars b scrolled to offX, offY,
// pressed the part drawn pressed: the inset, darker frame, the horizontal bar
// (drawHScroll) when it shows, then the vertical (drawVScroll). It returns
// the rectangle drawScroll draws in: 1 px inset, less the bars shown, and the
// crop (cropSize) one margin short at the right and the bottom, so the
// column and row inside the right and bottom border keep the frame's fill.
func drawScrollView(rt *gfx8.RenderTarget, x, y int, b L.Bars, c gfx8.InsetColour, offX, offY int, pressed L.Part) (cx, cy, cw, ch int) {
	gfx8.FillRectInset(rt, x, y, x+b.W-1, y+b.H-1, c, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
	cx, cy = x+L.ScrollbarMargin, y+L.ScrollbarMargin
	cw, ch = b.W-2*L.ScrollbarMargin, b.H-2*L.ScrollbarMargin
	if b.HShown() {
		drawHScroll(rt, x, y, b, c, offX, pressed)
		ch -= L.ScrollbarSize
	}
	if b.VShown() {
		drawVScroll(rt, x, y, b, c, offY, pressed)
		cw -= L.ScrollbarSize
	}
	return cx, cy, cw - L.ScrollbarMargin, ch - L.ScrollbarMargin
}

// drawBarTrack is the bar's track: shade 7, hatched shade 3, and its lines
// at 2, 3, 7 and 8 across it (along the bar).
func drawBarTrack(rt *gfx8.RenderTarget, l, t, r, bt int, horizontal bool, c gfx8.InsetColour) {
	s3, s7 := gfx8.Shade(c.Map, 3), gfx8.Shade(c.Map, 7)
	gfx8.FillRect(rt, l, t, r, bt, s7)
	fillRectCrossHatch(rt, l, t, r, bt, s3)
	for _, ln := range [][2]int{{2, int(s3)}, {3, int(s7)}, {7, int(s3)}, {8, int(s7)}} {
		if horizontal {
			gfx8.FillRect(rt, l, t+ln[0], r, t+ln[0], uint8(ln[1]))
		} else {
			gfx8.FillRect(rt, l+ln[0], t, l+ln[0], bt, uint8(ln[1]))
		}
	}
}

// insetIf is a part's RectInsetFlags: borderInset when it is the pressed one.
func insetIf(p, pressed L.Part) int {
	if p == pressed {
		return gfx8.InsetBorderInset
	}
	return 0
}

// hScrollRect is drawHScroll's scrollPos and scrollSize (widget-relative):
// one margin in from the left, at the bottom; the width less a bar and a
// margin when the vertical bar is flagged.
func hScrollRect(b L.Bars) (sx, sy, sw int) {
	sx, sy, sw = L.ScrollbarMargin, b.H-L.ScrollbarSize-L.ScrollbarMargin, b.W-2*L.ScrollbarMargin
	if b.VBar {
		sw -= L.ScrollbarSize + L.ScrollbarMargin
	}
	return sx, sy, sw
}

// vScrollRect is drawVScroll's scrollPos and scrollSize: at the right, one
// margin from the top; the height less a bar (no margin) when the horizontal
// bar is flagged.
func vScrollRect(b L.Bars) (sx, sy, sh int) {
	sx, sy, sh = b.W-L.ScrollbarSize-L.ScrollbarMargin, L.ScrollbarMargin, b.H-2*L.ScrollbarMargin
	if b.HBar {
		sh -= L.ScrollbarSize
	}
	return sx, sy, sh
}

// drawHScroll is ScrollViewWidget.cpp drawHScroll: the track, the left and
// right buttons, and the thumb (scrollPos + hThumbLeft, hThumbRight −
// hThumbLeft − 1 wide) for the widget at x, y.
func drawHScroll(rt *gfx8.RenderTarget, x, y int, b L.Bars, c gfx8.InsetColour, offX int, pressed L.Part) {
	sx, sy, sw := hScrollRect(b)
	sx, sy = x+sx, y+sy
	const sz = L.ScrollbarSize
	drawBarTrack(rt, sx, sy, sx+sw-1, sy+sz-1, true, c)
	gfx8.FillRectInset(rt, sx, sy, sx+sz-1, sy+sz-1, c, insetIf(L.PartHLeft, pressed))
	rx := sx + sw - sz + L.ScrollbarMargin
	gfx8.FillRectInset(rt, rx, sy, rx+sz-1, sy+sz-1, c, insetIf(L.PartHRight, pressed))
	l, r := b.HThumb(offX)
	gfx8.FillRectInset(rt, sx+l, sy, sx+l+(r-l-L.ScrollbarMargin)-1, sy+sz-1, c, insetIf(L.PartHThumb, pressed))
}

// drawVScroll is ScrollViewWidget.cpp drawVScroll: the track, the up and
// down buttons, and the thumb (scrollPos + vThumbTop, vThumbBottom −
// vThumbTop − 1 tall: vThumbTop counts from the widget's top, scrollPos is
// one margin below it, as getPart's hit test does not).
func drawVScroll(rt *gfx8.RenderTarget, x, y int, b L.Bars, c gfx8.InsetColour, offY int, pressed L.Part) {
	sx, sy, sh := vScrollRect(b)
	sx, sy = x+sx, y+sy
	const sz = L.ScrollbarSize
	drawBarTrack(rt, sx, sy, sx+sz-1, sy+sh-1, false, c)
	gfx8.FillRectInset(rt, sx, sy, sx+sz-1, sy+sz-1, c, insetIf(L.PartVTop, pressed))
	gfx8.FillRectInset(rt, sx, sy+sh-sz, sx+sz-1, sy+sh-1, c, insetIf(L.PartVBottom, pressed))
	top, bot := b.VThumb(offY)
	gfx8.FillRectInset(rt, sx, sy+top, sx+sz-1, sy+top+(bot-top-L.ScrollbarMargin)-1, c, insetIf(L.PartVThumb, pressed))
}

// drawVScrollArrows draws the up/down arrow glyphs on a vertical-only scroll
// view's buttons; x, y, w, h are screen coordinates of the widget.
func drawVScrollArrows(screen *ebiten.Image, x, y, w, h int) {
	drawScrollArrows(screen, x, y, L.Bars{W: w, H: h, ContentH: h + 1, VBar: true})
}

// drawScrollArrows draws drawHScroll's and drawVScroll's arrow strings in
// black on the bars shown: "\x90\xBE" left (scrollPos + kArrowOffset 2),
// "\x90\xAF" right, "\x90\xA0" up (scrollPos + (1, −1)), "\x90\xAA" down;
// x, y are the widget's screen coordinates.
func drawScrollArrows(screen *ebiten.Image, x, y int, b L.Bars) {
	if globalG1Font == nil {
		return
	}
	black := globalG1Font.g1.GetPalette()[paletteBlack0]
	const sz, arrowOffset = L.ScrollbarSize, 2
	if b.HShown() {
		sx, sy, sw := hScrollRect(b)
		DrawTextBold(screen, string(rune(0xBE)), x+sx+arrowOffset, y+sy, black)
		DrawTextBold(screen, string(rune(0xAF)), x+sx+sw-sz+L.ScrollbarMargin+arrowOffset, y+sy, black)
	}
	if b.VShown() {
		sx, sy, sh := vScrollRect(b)
		DrawTextBold(screen, string(rune(0xA0)), x+sx+1, y+sy-1, black)
		DrawTextBold(screen, string(rune(0xAA)), x+sx+L.ScrollbarMargin, y+sy+sh-sz, black)
	}
}
