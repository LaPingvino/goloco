package ui

import (
	"sort"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// The UI-sizing pass for fallback text (LOOP.md N56, a deviation: upstream's
// windows have fixed row heights for its 10 px fonts, and draws CJK as "?").
// A line holding a fallback glyph is fallbackLineH (18) tall, so in a window
// whose rows are 14–16 px apart those rows overlap. Joop decided (2026-11-20)
// to grow the UI rather than shrink the glyphs. A rowStretch does that for a
// window drawn with upstream's fixed layout: it inserts blank lines between
// the window's rows (repeating the line above the gap, which is the panel and
// any group box's sides), so every widget and its text keep their upstream
// positions relative to their row. The rows come from the text the window
// drew in its previous frame; a window with no tall text gets no gaps and
// draws exactly as before.

// rowGap inserts n lines before window line at.
type rowGap struct{ at, n int }

// rowStretch is a window's gaps, sorted by at.
type rowStretch []rowGap

// m maps a window y in the upstream layout to the stretched one.
func (s rowStretch) m(y int) int {
	out := y
	for _, g := range s {
		if y >= g.at {
			out += g.n
		}
	}
	return out
}

// inv maps a stretched window y back: a line in a gap is the line above it.
func (s rowStretch) inv(y int) int {
	off := 0
	for _, g := range s {
		if y < g.at+off {
			break
		}
		if y < g.at+off+g.n {
			return g.at - 1
		}
		off += g.n
	}
	return y - off
}

// total is how many lines the gaps add.
func (s rowStretch) total() int {
	t := 0
	for _, g := range s {
		t += g.n
	}
	return t
}

// draw8 runs an 8-bit draw laid out for the upstream size (h0 lines) on the
// stretched region: the region's lines are gathered into an h0 target, drawn
// on, and spread back with the gaps repeating their line above.
func (s rowStretch) draw8(rt *gfx8.RenderTarget, h0 int, draw func(rt *gfx8.RenderTarget)) {
	if len(s) == 0 {
		draw(rt)
		return
	}
	w := rt.Width
	src := gfx8.NewRenderTarget(w, h0)
	for y := 0; y < h0; y++ {
		if sy := s.m(y); sy < rt.Height {
			copy(src.Bits[y*w:(y+1)*w], rt.Bits[sy*w:(sy+1)*w])
		}
	}
	draw(src)
	for y := 0; y < rt.Height; y++ {
		if sy := s.inv(y); sy >= 0 && sy < h0 {
			copy(rt.Bits[y*w:(y+1)*w], src.Bits[sy*w:(sy+1)*w])
		}
	}
}

// textRows records a window's text rows while it draws and moves its text
// onto the stretched layout (the text functions call mapTextY).
type textRows struct {
	oy   int        // the window's top
	s    rowStretch // this frame's gaps
	rows map[int]bool
}

var activeTextRows *textRows

// mapTextY is the y a text line drawn at y goes to while a stretched window
// draws (y itself otherwise), noting the line's row and whether it is tall.
func mapTextY(y int, str string) int {
	t := activeTextRows
	if t == nil || str == "" {
		return y
	}
	rel := y - t.oy
	t.rows[rel] = t.rows[rel] || hasFallback(str)
	return t.oy + t.s.m(rel)
}

// beginTextRows starts recording a window's text (top oy, stretched by s).
func beginTextRows(oy int, s rowStretch) {
	activeTextRows = &textRows{oy: oy, s: s, rows: map[int]bool{}}
}

// endTextRows stops recording and gives the gaps the drawn text needs in a
// window h0 lines tall whose rows start at line top (above it: the title bar
// and tabs, never stretched). Text lines up to 3 px apart are one row (a
// label and its widget's text); after a tall row the next one starts at
// least fallbackLineH + 1 below it, and the last tall row gets room above
// the window's bottom 5-line margin.
func endTextRows(top, h0 int) rowStretch {
	t := activeTextRows
	activeTextRows = nil
	if t == nil {
		return nil
	}
	var ys []int
	any := false
	for y, tall := range t.rows {
		if y >= top && y < h0 {
			ys = append(ys, y)
			any = any || tall
		}
	}
	if !any {
		return nil
	}
	sort.Ints(ys)
	type row struct {
		top  int
		tall bool
	}
	var rows []row
	for _, y := range ys {
		if n := len(rows); n > 0 && y-rows[n-1].top <= 3 {
			rows[n-1].tall = rows[n-1].tall || t.rows[y]
			continue
		}
		rows = append(rows, row{y, t.rows[y]})
	}
	var s rowStretch
	for i, r := range rows {
		if !r.tall {
			continue
		}
		next, at := h0-5, r.top+fallbackLineH-4 // the bottom margin; the gap below the row's 12 px widgets
		if i+1 < len(rows) {
			next, at = rows[i+1].top, rows[i+1].top-1
		}
		if n := r.top + fallbackLineH + 1 - next; n > 0 {
			s = append(s, rowGap{at, n})
		}
	}
	return s
}
