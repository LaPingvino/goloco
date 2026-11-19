package ui

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// TestMapCentreOnViewPoint: MapWindow.cpp centerOnViewPoint (l. 2452) on the
// 350 × 272 window (scroll view 240 × 215: the map's visible part 229 × 204,
// contentWidth 769 − 229 = 540, contentHeight 769 − 204 = 565 the most).
func TestMapCentreOnViewPoint(t *testing.T) {
	cases := []struct {
		name   string
		view   MapMainView
		sx, sy int
	}{
		// vpCentre ((400 + 1000) / 32, (300 + 2000) / 16) = (43, 143)
		{"rotation 0", MapMainView{1000, 2000, 800, 600, 0, true}, 43 + 384 - 114, 143 + 0 - 102},
		{"rotation 1: x past the content", MapMainView{1000, 2000, 800, 600, 1, true}, 540, 143 + 384 - 102},
		{"rotation 2: y past the content", MapMainView{1000, 2000, 800, 600, 2, true}, 43 + 384 - 114, 565},
		{"rotation 3: 43 + 0 − 114 < 0", MapMainView{1000, 2000, 800, 600, 3, true}, 0, 143 + 384 - 102},
		// (400 − 20000) / 32 truncates to −612: −612 + 384 − 114 < 0
		{"left of the map", MapMainView{-20000, 2000, 800, 600, 0, true}, 0, 41},
	}
	for _, c := range cases {
		w := NewMapWindow(nil, 0, 1)
		v := c.view
		w.MainView = func() MapMainView { return v }
		w.CentreOnViewPoint()
		if w.ScrollX != c.sx || w.ScrollY != c.sy {
			t.Errorf("%s: offset (%d, %d), want (%d, %d)", c.name, w.ScrollX, w.ScrollY, c.sx, c.sy)
		}
	}
	w := NewMapWindow(nil, 0, 1)
	w.ScrollX, w.ScrollY = 7, 9
	w.MainView = func() MapMainView { return MapMainView{} } // no main viewport
	w.CentreOnViewPoint()
	if w.ScrollX != 7 || w.ScrollY != 9 {
		t.Errorf("without a main viewport the offset moved: (%d, %d)", w.ScrollX, w.ScrollY)
	}
}

// TestMapClickCentres: getPart's view part (the widget less its top / left
// border and the bars) and scrollMouseDown / scrollMouseDrag's
// mapWindowPosToLocation at the view's rotation.
func TestMapClickCentres(t *testing.T) {
	w := NewMapWindow(nil, 0, 1)
	w.X, w.Y = 100, 50 // the scroll view: 103 … 342 × 94 … 308
	w.ScrollX, w.ScrollY = 313, 41
	rot := uint8(1)
	w.MainView = func() MapMainView { return MapMainView{Rotation: rot, OK: true} }
	var got [][2]int32
	w.OnCentre = func(x, y int32) { got = append(got, [2]int32{x, y}) }
	want := func(cx, cy int) [2]int32 {
		x, y := sim.MapWindowPosToLocation(cx, cy, rot)
		return [2]int32{x, y}
	}
	cases := []struct {
		name   string
		px, py int
		ok     bool
		cx, cy int
	}{
		{"the first content pixel", 104, 95, true, 313, 41},
		{"the border", 103, 95, false, 0, 0},
		{"the last column before the vertical bar", 330, 150, true, 330 - 104 + 313, 150 - 95 + 41},
		{"the vertical bar", 331, 150, false, 0, 0},
		{"the last row before the horizontal bar", 200, 296, true, 200 - 104 + 313, 296 - 95 + 41},
		{"the horizontal bar", 200, 297, false, 0, 0},
		{"the key beside the map", 360, 150, false, 0, 0},
	}
	for _, c := range cases {
		got = nil
		w.viewHeld = false
		w.HandleClick(c.px, c.py)
		switch {
		case !c.ok && len(got) != 0:
			t.Errorf("%s: centred on %v", c.name, got)
		case c.ok && (len(got) != 1 || got[0] != want(c.cx, c.cy)):
			t.Errorf("%s: centred on %v, want %v (content %d, %d)", c.name, got, want(c.cx, c.cy), c.cx, c.cy)
		case w.viewHeld != c.ok:
			t.Errorf("%s: held %v", c.name, w.viewHeld)
		}
	}

	// The drag: each tick over the view part centres again; off this
	// window (another on top) or on a bar it waits; the release ends it.
	got = nil
	w.HandleClick(150, 150)
	w.ScrollDrag(160, 150, true, true)
	w.ScrollDrag(170, 150, true, false)
	w.ScrollDrag(331, 150, true, true)
	w.ScrollDrag(180, 150, false, true)
	w.ScrollDrag(190, 150, true, true) // a hold no longer
	if len(got) != 2 || got[1] != want(160-104+313, 150-95+41) {
		t.Errorf("drag centred on %v", got)
	}
}

// TestMapFillRectCrossHatch: drawRectImpl's crossHatching — every other pixel,
// the phase flipped per row and by how far the rectangle is clipped at the
// target's left (crossPattern ^= leftX).
func TestMapFillRectCrossHatch(t *testing.T) {
	row := func(left int) string {
		rt := gfx8.NewRenderTarget(6, 2)
		mapTarget{rt: rt, W: 6, H: 2}.fillRect(left, 0, 5, 1, 1, true)
		s := ""
		for i, b := range rt.Bits {
			if i == 6 {
				s += "/"
			}
			s += string(rune('0' + b))
		}
		return s
	}
	if got := row(0); got != "101010/010101" {
		t.Errorf("unclipped: %s", got)
	}
	if got := row(-3); got != "010101/101010" { // −3: odd, the phase flips
		t.Errorf("clipped by 3: %s", got)
	}
	if got := row(-2); got != "101010/010101" {
		t.Errorf("clipped by 2: %s", got)
	}
}
