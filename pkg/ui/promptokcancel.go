package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/hajimehoshi/ebiten/v2"
)

// PromptOkCancel is PromptOkCancelWindow.cpp: a 280×92 modal centred on the
// screen, translucent muted dark red, a boxed caption, a close button, the
// description centred and wrapped at y 41, and OK / Cancel buttons. Upstream
// blocks in its own loop until it closes; here the game routes all input to
// it while it is open and calls OnOK when OK is chosen.
type PromptOkCancel struct {
	X, Y                   int
	Caption, Text, OKLabel string
	OnOK                   func()
	Closed                 bool
}

const (
	pocW, pocH         = 280, 92
	colourMutedDarkRed = 25
)

// NewPromptOkCancel opens the prompt centred on an sw × sh screen
// (createWindowCentred).
func NewPromptOkCancel(caption, text, okLabel string, sw, sh int, onOK func()) *PromptOkCancel {
	return &PromptOkCancel{X: (sw - pocW) / 2, Y: (sh - pocH) / 2, Caption: caption, Text: text, OKLabel: okLabel, OnOK: onOK}
}

func (p *PromptOkCancel) in(r [4]int, px, py int) bool {
	return px >= p.X+r[0] && px < p.X+r[0]+r[2] && py >= p.Y+r[1] && py < p.Y+r[1]+r[3]
}

var (
	pocClose  = [4]int{267, 2, 11, 11}
	pocOK     = [4]int{20, 77, 100, 12}
	pocCancel = [4]int{160, 77, 100, 12}
)

// HandleClick is onMouseUp: close / Cancel close it, OK closes it and calls
// OnOK. Any click is the prompt's while it is open.
func (p *PromptOkCancel) HandleClick(px, py int) {
	switch {
	case p.in(pocClose, px, py), p.in(pocCancel, px, py):
		p.Closed = true
	case p.in(pocOK, px, py):
		p.Closed = true
		if p.OnOK != nil {
			p.OnOK()
		}
	}
}

// Key is keyUp: Escape is the close button, Enter is OK.
func (p *PromptOkCancel) Key(escape, enter bool) {
	switch {
	case escape:
		p.Closed = true
	case enter:
		p.HandleClick(p.X+pocOK[0], p.Y+pocOK[1])
	}
}

// Draw draws the window: panel, boxed caption (inset + fillDarker, the
// inside darkened with ExtColour unk2E, white outlined text centred), the
// close button (❌), OK and Cancel, and the description.
func (p *PromptOkCancel) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	cmap, err := g1.GetPaletteMap(colourMutedDarkRed)
	if err != nil {
		return
	}
	var trans [3][]uint8
	for k := range trans {
		trans[k], _ = g1.GetPaletteMap(int(paint.TranslucentColour[colourMutedDarkRed][k]))
	}
	c := gfx8.InsetColour{Map: cmap, Translucent: true, Trans: trans}
	m2e, _ := g1.GetPaletteMap(extColourUnk2E)
	draw8(screen, p.X, p.Y, pocW, pocH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 0, pocW-1, pocH-1, c, 0)                                           // Panel
		gfx8.FillRectInset(rt, 1, 1, 1+278-1, 1+13-1, c, gfx8.InsetBorderInset|gfx8.InsetFillDarker) // Caption boxed
		gfx8.FillRectTranslucent(rt, 2, 2, 2+276-1, 2+11-1, m2e)
		for _, b := range [][4]int{pocClose, pocOK, pocCancel} {
			gfx8.FillRectInset(rt, b[0], b[1], b[0]+b[2]-1, b[1]+b[3]-1, c, 0)
		}
		// The close button's cross: the font's ❌ (close_window_cross) is the
		// same image the other windows' close buttons draw.
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, pocClose[0], pocClose[1], gfx8.OpTransparent|gfx8.OpSrc, cmap, nil)
		}
	})
	pal := g1.GetPalette()
	white := pal[gfx8.Shade(mustMap(colourWhite), 9)]
	black := pal[paletteBlack0]
	text := pal[gfx8.Shade(cmap, 9)] // the window colour, opaque
	cw := 278 - 4 - 10
	tw, _ := MeasureTextBold(p.Caption)
	DrawTextOutlined(screen, p.Caption, p.X+1+2+cw/2-tw/2, p.Y+1+1, white, black)
	for _, b := range []struct {
		r [4]int
		s string
	}{{pocOK, p.OKLabel}, {pocCancel, "Cancel"}} {
		bw, _ := MeasureTextBold(b.s)
		DrawTextBold(screen, b.s, p.X+b.r[0]+(b.r[2]-1)/2-bw/2, p.Y+b.r[1], text)
	}
	// wcolour2_stringid, drawStringCentredWrapped at (width/2, 41).
	for i, line := range WrapTextBold(p.Text, pocW) {
		lw, _ := MeasureTextBold(line)
		DrawTextBold(screen, line, p.X+pocW/2-lw/2, p.Y+41+10*i, text)
	}
}

// mustMap is a palette map by colour (nil on error).
func mustMap(colour int) []uint8 {
	m, _ := globalG1Font.g1.GetPaletteMap(colour)
	return m
}
