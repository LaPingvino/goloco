package ui

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
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
		w.hold.Release()
		w.ScrollX, w.ScrollY = 313, 41
		w.HandleClick(c.px, c.py)
		switch {
		case !c.ok && len(got) != 0:
			t.Errorf("%s: centred on %v", c.name, got)
		case c.ok && (len(got) != 1 || got[0] != want(c.cx, c.cy)):
			t.Errorf("%s: centred on %v, want %v (content %d, %d)", c.name, got, want(c.cx, c.cy), c.cx, c.cy)
		case (w.hold.Part == L.PartView) != c.ok:
			t.Errorf("%s: held %d", c.name, w.hold.Part)
		}
	}

	// The drag: each tick over the view part centres again; off this
	// window (another on top) or on a bar it waits; the release ends it.
	got = nil
	w.hold.Release()
	w.ScrollX, w.ScrollY = 313, 41
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

// TestMapScrollBars: the map's bars as every scroll view's (ScrollView.cpp):
// the window at 100, 50 has its scroll view at 103 … 342 × 94 … 308, the
// horizontal bar its bottom 12 rows, the vertical its right 12 columns. A
// press on the right button steps 3 and repeats while held on it; the
// vertical thumb follows the pointer (verticalFollow: 10 px × 769 / 184);
// the corner's deadzone does nothing; a bar press never centres the view.
func TestMapScrollBars(t *testing.T) {
	w := NewMapWindow(nil, 0, 1)
	w.X, w.Y = 100, 50
	w.MainView = func() MapMainView { return MapMainView{OK: true} }
	centred := 0
	w.OnCentre = func(x, y int32) { centred++ }
	w.ScrollX, w.ScrollY = 100, 100
	w.HandleClick(103+220, 94+208) // hscrollbarButtonRight
	if w.ScrollX != 103 || w.hold.PressedPart() != L.PartHRight {
		t.Fatalf("right button: %d %d", w.ScrollX, w.hold.PressedPart())
	}
	w.ScrollDrag(103+220, 94+208, true, true)
	w.ScrollDrag(103+220, 94+208, false, true)
	if w.ScrollX != 106 || w.hold.Part != L.PartNone {
		t.Fatalf("held, released: %d %d", w.ScrollX, w.hold.Part)
	}
	top, _ := w.bars().VThumb(w.ScrollY)
	w.HandleClick(103+230, 94+top+2) // vscrollbarThumb
	w.ScrollDrag(103+230, 94+top+12, true, true)
	if w.ScrollY != 100+41 || w.hold.PressedPart() != L.PartVThumb {
		t.Fatalf("v thumb: %d %d", w.ScrollY, w.hold.PressedPart())
	}
	w.ScrollDrag(0, 0, false, true)
	w.HandleClick(103+235, 94+210) // the corner
	if w.ScrollX != 106 || w.ScrollY != 141 || w.hold.Part != L.PartNone || centred != 0 {
		t.Fatalf("corner: %d %d %d %d", w.ScrollX, w.ScrollY, w.hold.Part, centred)
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

// TestMapLastAttributes: onClose keeps the size, var_88A / var_88C and the
// flags with hasStoredState (getLastMapWindowAttributes); open takes the size
// back within 350 … screen width, 272 … screen height − 56, and the
// maximised flag and the size to un-maximise to (MapWindow.cpp l. 197, 2409).
func TestMapLastAttributes(t *testing.T) {
	if w, h := MapOpenSize(scenario.LastMapWindow{Width: 500, Height: 400}, 800, 600); w != 350 || h != 272 {
		t.Errorf("never closed: %dx%d", w, h)
	}
	w := NewMapWindow(nil, 0, 1)
	w.W, w.H, w.restoreW, w.restoreH = 640, 580, 400, 300
	w.WinFlags |= WindowFlagMaximised | WindowFlagHasBeenResized
	last := w.LastAttributes()
	want := scenario.LastMapWindow{Flags: 1<<31 | 1<<16 | 1<<15 | 1<<9, Width: 640, Height: 580, Var88A: 400, Var88C: 300}
	if last != want {
		t.Fatalf("onClose: %+v, want %+v", last, want)
	}
	if ow, oh := MapOpenSize(last, 800, 600); ow != 640 || oh != 544 { // 580 kept to 600 − 56
		t.Errorf("open: %dx%d", ow, oh)
	}
	if ow, oh := MapOpenSize(scenario.LastMapWindow{Flags: 1 << 31, Width: 229, Height: 176}, 800, 600); ow != 350 || oh != 272 {
		t.Errorf("open small: %dx%d", ow, oh)
	}
	n := NewMapWindow(nil, 0, 1)
	n.RestoreLast(last)
	if n.restoreW != 400 || n.restoreH != 300 || !n.HasFlags(WindowFlagMaximised) || n.HasFlags(WindowFlagHasBeenResized) {
		t.Errorf("restored: %d %d %#x", n.restoreW, n.restoreH, n.WinFlags)
	}
	n = NewMapWindow(nil, 0, 1)
	n.RestoreLast(scenario.LastMapWindow{Flags: 1 << 31, Var88A: 5})
	if n.HasFlags(WindowFlagMaximised) || n.restoreW != 5 {
		t.Errorf("not maximised: %#x %d", n.WinFlags, n.restoreW)
	}
}
