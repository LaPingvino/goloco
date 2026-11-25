//go:build js

package ui

import (
	"math"
	"syscall/js"
)

// canvasSource draws fallback glyphs with the browser's own fonts (the web
// build has no font files): Unifont when the browser has it, else whatever
// font the browser falls back to for the rune, at Unifont's 16 px cell,
// thresholded to 1 bit so it sits in the Loco look.
type canvasSource struct {
	ctx js.Value
}

func findFallbackSource() glyphSource {
	doc := js.Global().Get("document")
	if doc.IsUndefined() || doc.IsNull() {
		return nil
	}
	c := doc.Call("createElement", "canvas")
	c.Set("width", 2*fallbackCell)
	c.Set("height", fallbackCell)
	ctx := c.Call("getContext", "2d", map[string]any{"willReadFrequently": true})
	if ctx.IsNull() || ctx.IsUndefined() {
		return nil
	}
	ctx.Set("font", "16px Unifont, sans-serif")
	ctx.Set("textBaseline", "alphabetic")
	ctx.Set("fillStyle", "#fff")
	return &canvasSource{ctx: ctx}
}

func (s *canvasSource) glyph(r rune) (glyphBits, bool) {
	ch := string(r)
	w := int(math.Round(s.ctx.Call("measureText", ch).Get("width").Float()))
	if w <= 0 {
		return glyphBits{}, false
	}
	w = min(w, 2*fallbackCell)
	s.ctx.Call("clearRect", 0, 0, 2*fallbackCell, fallbackCell)
	s.ctx.Call("fillText", ch, 0, fallbackAscent)
	data := s.ctx.Call("getImageData", 0, 0, w, fallbackCell).Get("data")
	buf := make([]byte, data.Get("length").Int())
	js.CopyBytesToGo(buf, js.Global().Get("Uint8Array").New(data.Get("buffer")))
	g := glyphBits{width: w}
	for y := 0; y < fallbackCell; y++ {
		row := make([]bool, w)
		for x := 0; x < w; x++ {
			row[x] = buf[(y*w+x)*4+3] >= 0x80
		}
		g.rows = append(g.rows, row)
	}
	return g, true
}
