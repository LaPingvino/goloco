package ui

import (
	"image"
	"image/color"
	"log"
	"os"
	"strings"
	"time"

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
	t0 := time.Now()
	defer func() { Draw8Stats.Calls++; Draw8Stats.Total += time.Since(t0) }()
	rt := gfx8.NewRenderTarget(w, h)
	ox, oy := r.Min.X-x, r.Min.Y-y
	if screen8.usable(r) { // the indices under it, from the CPU copy of the screen: no GPU readback
		for j := 0; j < r.Dy(); j++ {
			copy(rt.Bits[(oy+j)*w+ox:(oy+j)*w+ox+r.Dx()], screen8.bits[(r.Min.Y+j)*screen8.w+r.Min.X:])
		}
		if draw8Check { // GOLOCO_DRAW8_CHECK: compare with the readback
			buf := make([]byte, 4*r.Dx()*r.Dy())
			screen.SubImage(r).(*ebiten.Image).ReadPixels(buf)
			diff := 0
			for j := 0; j < r.Dy(); j++ {
				for i := 0; i < r.Dx(); i++ {
					p := buf[4*(j*r.Dx()+i):]
					if pal[rt.Bits[(oy+j)*w+ox+i]] != pal[inv.lookup(pal, p[0], p[1], p[2])] {
						diff++
					}
				}
			}
			Draw8Stats.CheckPixels += r.Dx() * r.Dy()
			Draw8Stats.CheckDiff += diff
		}
	} else {
		buf := make([]byte, 4*r.Dx()*r.Dy())
		screen.SubImage(r).(*ebiten.Image).ReadPixels(buf)
		Draw8Stats.Readback += time.Since(t0)
		Draw8Stats.Readbacks++
		for j := 0; j < r.Dy(); j++ {
			for i := 0; i < r.Dx(); i++ {
				p := buf[4*(j*r.Dx()+i):]
				rt.Bits[(oy+j)*w+ox+i] = inv.lookup(pal, p[0], p[1], p[2])
			}
		}
	}
	draw(rt)
	screen8.store(r, rt, ox, oy)
	img := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for j := 0; j < r.Dy(); j++ {
		row := img.Pix[j*img.Stride:]
		for i := 0; i < r.Dx(); i++ {
			c := pal[rt.Bits[(oy+j)*w+ox+i]]
			row[4*i], row[4*i+1], row[4*i+2], row[4*i+3] = c.R, c.G, c.B, 255
		}
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	screen.DrawImage(draw8Texture(r.Dx(), r.Dy(), img.Pix), op)
}

// Draw8Stats counts draw8's calls and time (frame stats, LOOP item 0): the
// GPU readback is what is left of the UI's cost.
var Draw8Stats struct {
	Calls, Readbacks int
	// CheckPixels / CheckDiff: GOLOCO_DRAW8_CHECK — pixels compared with a
	// readback, and how many differed in colour.
	CheckPixels, CheckDiff int
	Readback, Total        time.Duration
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

// palInverse is the RGB → palette index table: a flat entry per 24-bit
// colour (16 MB, made once), 0 = not looked up yet (index 0, transparent, is
// never an answer). It was a map keyed by [3]uint8 — a hash per pixel, a tenth
// of the CPU with a window open (LOOP item 0).
type palInverse []uint8

var palInv palInverse

// paletteIndex is the table, with the palette's own colours filled in (the
// first index for a colour; index 0 skipped).
func paletteIndex(pal []color.RGBA) palInverse {
	if palInv == nil {
		palInv = make(palInverse, 1<<24)
		for i := len(pal) - 1; i >= 1; i-- {
			c := pal[i]
			palInv[rgbKey(c.R, c.G, c.B)] = uint8(i)
		}
	}
	return palInv
}

func rgbKey(r, g, b uint8) int { return int(r)<<16 | int(g)<<8 | int(b) }

// lookup is the palette index of a colour: its own, or the nearest (squared
// RGB distance, the lowest index on a tie), remembered.
func (m palInverse) lookup(pal []color.RGBA, r, g, b uint8) uint8 {
	k := rgbKey(r, g, b)
	if i := m[k]; i != 0 {
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
	m[k] = best
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

// screen8 is a CPU copy of the screen as palette indices, so draw8 can take
// what lies under it without reading the GPU back (that readback was nearly
// all of the UI's cost, LOOP item 0): the world's own 8-bit frame
// (BeginFrame8, from pkg/scene), updated by every draw8. Whatever else is drawn
// on the screen meanwhile (labels, the toolbar, the title menu, old-style
// windows, construction ghosts) marks its rectangle foreign (MarkForeign); a
// draw8 over a foreign rectangle, or without a world frame this frame, reads
// the GPU back as before.
type screen8Copy struct {
	bits    []uint8
	w, h    int
	valid   bool
	foreign []image.Rectangle
}

var screen8 screen8Copy

var draw8Check = os.Getenv("GOLOCO_DRAW8_CHECK") != ""

// BeginFrame8 starts the frame's copy from the world's 8-bit frame (w×h
// palette indices at the screen's origin).
func BeginFrame8(bits []uint8, w, h int) {
	if len(screen8.bits) != len(bits) {
		screen8.bits = make([]uint8, len(bits))
	}
	copy(screen8.bits, bits)
	screen8.w, screen8.h, screen8.valid = w, h, true
	screen8.foreign = screen8.foreign[:0]
}

// InvalidateFrame8 drops the copy (a frame without the 8-bit world).
func InvalidateFrame8() { screen8.valid = false; screen8.foreign = screen8.foreign[:0] }

// MarkForeign records a rectangle drawn on the screen other than by draw8.
func MarkForeign(r image.Rectangle) {
	if screen8.valid && !r.Empty() {
		screen8.foreign = append(screen8.foreign, r)
	}
}

func (s *screen8Copy) usable(r image.Rectangle) bool {
	if !s.valid || !r.In(image.Rect(0, 0, s.w, s.h)) {
		if draw8Check {
			log.Printf("[draw8] %v: no copy (valid %v, %dx%d)", r, s.valid, s.w, s.h)
		}
		return false
	}
	for _, f := range s.foreign {
		if f.Overlaps(r) {
			if draw8Check {
				log.Printf("[draw8] %v: foreign %v", r, f)
			}
			return false
		}
	}
	return true
}

// store keeps what draw8 drew in the copy (later draw8s see it).
func (s *screen8Copy) store(r image.Rectangle, rt *gfx8.RenderTarget, ox, oy int) {
	if !s.valid || !r.In(image.Rect(0, 0, s.w, s.h)) {
		return
	}
	for j := 0; j < r.Dy(); j++ {
		copy(s.bits[(r.Min.Y+j)*s.w+r.Min.X:(r.Min.Y+j)*s.w+r.Max.X], rt.Bits[(oy+j)*rt.Width+ox:])
	}
}

// DrawRect8 is draw8 for other packages: draw paints the 8-bit pixels of the
// screen rectangle (x, y, w, h), starting from what is under it.
func DrawRect8(screen *ebiten.Image, x, y, w, h int, draw func(rt *gfx8.RenderTarget)) {
	draw8(screen, x, y, w, h, draw)
}
