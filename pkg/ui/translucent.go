package ui

import (
	"image"
	"image/color"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// draw8 runs 8-bit drawing (gfx8: translucent remaps, inset rects…) on a
// screen region: the region's pixels are read back to palette indices, drawn
// on as upstream draws into its 8-bit frame, and written back. Exact as long as
// the region holds palette colours (the CPU renderer's frame and the UI do).
// Pixels that are not a palette colour take the nearest one.
func draw8(screen *ebiten.Image, x, y, w, h int, draw func(rt *gfx8.RenderTarget)) {
	if globalG1Font == nil || w <= 0 || h <= 0 {
		return
	}
	b := screen.Bounds()
	r := image.Rect(x, y, x+w, y+h).Intersect(b)
	if r.Empty() {
		return
	}
	pal := globalG1Font.g1.GetPalette()
	inv := paletteIndex(pal)
	buf := make([]byte, 4*r.Dx()*r.Dy())
	screen.SubImage(r).(*ebiten.Image).ReadPixels(buf)
	rt := gfx8.NewRenderTarget(w, h)
	ox, oy := r.Min.X-x, r.Min.Y-y
	for j := 0; j < r.Dy(); j++ {
		for i := 0; i < r.Dx(); i++ {
			p := buf[4*(j*r.Dx()+i):]
			rt.Bits[(oy+j)*w+ox+i] = inv.lookup(pal, p[0], p[1], p[2])
		}
	}
	draw(rt)
	img := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for j := 0; j < r.Dy(); j++ {
		for i := 0; i < r.Dx(); i++ {
			c := pal[rt.Bits[(oy+j)*w+ox+i]]
			img.SetRGBA(i, j, color.RGBA{c.R, c.G, c.B, 255})
		}
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	screen.DrawImage(draw8Texture(r.Dx(), r.Dy(), img.Pix), op)
}

// draw8Textures keeps one texture per size for draw8: it runs many times a
// frame (panels, toolbar, windows), and a new texture per call — even freed
// with Deallocate — let the browser's GPU process grow ~1 GB/s in gameplay
// until it was killed (ticks 259–261, measured).
var draw8Textures = map[[2]int]*ebiten.Image{}

func draw8Texture(w, h int, pix []byte) *ebiten.Image {
	k := [2]int{w, h}
	t := draw8Textures[k]
	if t == nil {
		t = ebiten.NewImage(w, h)
		draw8Textures[k] = t
	}
	t.WritePixels(pix)
	return t
}

type palInverse map[[3]uint8]uint8

var palInv palInverse

// paletteIndex is the RGB → palette index map (the first index for a colour;
// index 0 is skipped, it is transparent).
func paletteIndex(pal []color.RGBA) palInverse {
	if palInv == nil {
		palInv = palInverse{}
		for i := len(pal) - 1; i >= 1; i-- {
			c := pal[i]
			palInv[[3]uint8{c.R, c.G, c.B}] = uint8(i)
		}
	}
	return palInv
}

func (m palInverse) lookup(pal []color.RGBA, r, g, b uint8) uint8 {
	if i, ok := m[[3]uint8{r, g, b}]; ok {
		return i
	}
	best, bd := uint8(1), 1<<30
	for i := 1; i < len(pal); i++ {
		c := pal[i]
		d := sq(int(c.R)-int(r)) + sq(int(c.G)-int(g)) + sq(int(c.B)-int(b))
		if d < bd {
			best, bd = uint8(i), d
		}
	}
	m[[3]uint8{r, g, b}] = best
	return best
}

func sq(v int) int { return v * v }

// ExtColour values the UI's translucent rectangles use.
const (
	extColourUnk2D = 0x2D
	extColourUnk74 = 0x74 // + a colour: Colours::getGlass
)

// DrawTooltip is ToolTip.cpp common + draw: text (medium bold, max 196 px
// wide, word-wrapped) in a translucent box 26 px below the cursor, centred on
// it, kept on screen. tooltipColour is the interface skin's tooltipColour.
// PARTIAL: wrapping breaks at spaces with pkg/ui text measurement (upstream's
// TextRenderer::wrapString is not ported) and lines are drawn with DrawTextBold.
func DrawTooltip(screen *ebiten.Image, text string, cursorX, cursorY int, tooltipColour uint8) {
	if globalG1Font == nil || text == "" {
		return
	}
	lines := wrapTooltip(text, 196)
	ww := 0
	for _, l := range lines {
		if w, _ := MeasureTextBold(l); w > ww {
			ww = w
		}
	}
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	width := ww + 3
	height := len(lines)*10 + 4
	maxY := sh - height
	y := cursorY + 26
	if y > maxY {
		y -= height + 40
	}
	y = min(max(y, 22), maxY)
	x := 0
	if width <= sw {
		x = min(max(cursorX-width/2, 0), sw-width)
	}
	g1 := globalG1Font.g1
	m2d, _ := g1.GetPaletteMap(extColourUnk2D)
	mGlass, _ := g1.GetPaletteMap(extColourUnk74 + int(tooltipColour))
	m2e, _ := g1.GetPaletteMap(0x2E)
	draw8(screen, x, y, width, height, func(rt *gfx8.RenderTarget) {
		tr := func(l, t, w, h int, m []uint8) { gfx8.FillRectTranslucent(rt, l, t, l+w-1, t+h-1, m) }
		tr(1, 1, width-2, height-2, m2d)
		tr(1, 1, width-2, height-2, mGlass)
		tr(0, 2, 1, height-4, m2e)
		tr(width-1, 2, 1, height-4, m2e)
		tr(2, height-1, width-4, 1, m2e)
		tr(2, 0, width-4, 1, m2e)
		tr(1, 1, 1, 1, m2e)
		tr(width-2, 1, 1, 1, m2e)
		tr(1, height-2, 1, 1, m2e)
		tr(width-2, height-2, 1, 1, m2e)
	})
	// drawStringCentredRaw at ((width+1)/2 + x - 1, y + 1), Colour::black.
	cx := (width+1)/2 + x - 1
	for i, l := range lines {
		lw, _ := MeasureTextBold(l)
		DrawTextBold(screen, l, cx-lw/2, y+1+i*10, color.Black)
	}
}

func wrapTooltip(text string, maxW int) []string {
	if i := strings.IndexByte(text, '\n'); i >= 0 { // {NEWLINE}
		return append(wrapTooltip(text[:i], maxW), wrapTooltip(text[i+1:], maxW)...)
	}
	if w, _ := MeasureTextBold(text); w <= maxW {
		return []string{text}
	}
	var lines []string
	cur := ""
	for _, word := range splitWords(text) {
		try := word
		if cur != "" {
			try = cur + " " + word
		}
		if w, _ := MeasureTextBold(try); w > maxW && cur != "" {
			lines = append(lines, cur)
			cur = word
		} else {
			cur = try
		}
	}
	return append(lines, cur)
}

func splitWords(s string) []string {
	var out []string
	w := ""
	for _, r := range s {
		if r == ' ' {
			if w != "" {
				out = append(out, w)
			}
			w = ""
			continue
		}
		w += string(r)
	}
	if w != "" {
		out = append(out, w)
	}
	return out
}
