package ui

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
)

// testColourMap is a palette map whose every index is itself, so a drawn
// pixel names the map index (the shade) it came from.
func testColourMap() []uint8 {
	m := make([]uint8, 256)
	for i := range m {
		m[i] = uint8(i)
	}
	return m
}

// GroupBox::draw since 588abac: the border inside the widget's w × h — the
// right edge's dark line in column w − 1, the bottom's in row h − 1, nothing
// beyond (the old border lay one pixel outside).
func TestGroupBoxInsideItsSize(t *testing.T) {
	m := testColourMap()
	s4, s7 := gfx8.Shade(m, 4), gfx8.Shade(m, 7)
	const w, h = 30, 20
	rt := gfx8.NewRenderTarget(w+4, h+4)
	groupBox(rt, 0, 0, w, h, m, "")
	at := func(x, y int) uint8 { return rt.Bits[y*rt.Width+x] }
	for y := 4; y < h; y++ {
		if at(w, y) != 0 {
			t.Fatalf("column %d (outside) drawn at row %d", w, y)
		}
	}
	for x := 0; x < w; x++ {
		if at(x, h) != 0 {
			t.Fatalf("row %d (outside) drawn at column %d", h, x)
		}
	}
	if at(w-1, 10) != s7 || at(w-2, 10) != s4 {
		t.Errorf("right edge: %d %d, want %d %d", at(w-2, 10), at(w-1, 10), s4, s7)
	}
	if at(10, h-1) != s7 || at(10, h-2) != s4 {
		t.Errorf("bottom edge: %d %d, want %d %d", at(10, h-2), at(10, h-1), s4, s7)
	}
	// The left inner line: rows 6 … 4 + (h − 4) − 3.
	if at(1, h-3) != s7 || at(1, h-2) == s7 {
		t.Errorf("left inner line ends at row %d: %d, row %d: %d", h-3, at(1, h-3), h-2, at(1, h-2))
	}
}

// ScrollViewWidget's drawVScroll: the thumb at scrollPos + vThumbTop, one
// margin below the widget's top — scrolled to the top it starts right under
// the up button and leaves the button's last row alone.
func TestVScrollThumbUnderTheButton(t *testing.T) {
	m := testColourMap()
	c := gfx8.InsetColour{Map: m}
	const w, h, content = 100, 100, 400
	draw := func(offset int) *gfx8.RenderTarget {
		rt := gfx8.NewRenderTarget(w, h)
		drawVScrollView(rt, 0, 0, w, h, c, content, offset)
		return rt
	}
	top, bottom := draw(0), draw(content-h)
	sx := w - L.ScrollbarSize - L.ScrollbarMargin
	row := func(rt *gfx8.RenderTarget, y int) []uint8 { return rt.Bits[y*w+sx : y*w+sx+L.ScrollbarSize] }
	btnLast := L.ScrollbarMargin + L.ScrollbarSize - 1 // 11
	if string(row(top, btnLast)) != string(row(bottom, btnLast)) {
		t.Errorf("the up button's last row is overdrawn by the thumb: %v vs %v", row(top, btnLast), row(bottom, btnLast))
	}
	if string(row(top, btnLast+1)) == string(row(bottom, btnLast+1)) {
		t.Errorf("no thumb under the up button at the top: %v", row(top, btnLast+1))
	}
}

// dropdownWidgets' button is drawn pressed while the dropdown its press
// opened is up (Window::draw's pressedWidget in State::dropdownActive); a
// press on the combobox is the button's (Window::findWidgetAt).
func TestDropdownButtonPressedWhileOpen(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	w := NewMusicJukebox(800, 600, 0, 1)
	opened := 0
	w.OpenDropdown = func(*DropdownMenu) { opened++ }
	s.Open(w)
	btn := dropdownButton(mjPlaylist)
	for _, p := range [][2]int{{w.X + btn[0] + 3, w.Y + btn[1] + 3}, {w.X + mjPlaylist[0] + 20, w.Y + mjPlaylist[1] + 5}} {
		s.Press(p[0], p[1])
		if opened == 0 {
			t.Fatalf("press at %v opened no dropdown", p)
		}
		if r := s.pressedBtn; r.Min.X != w.X+btn[0] || r.Min.Y != w.Y+btn[1] || r.Dx() != 11 || r.Dy() != 10 {
			t.Errorf("press at %v: pressed widget %v, want the button", p, r)
		}
		if w.dropdownPressed(btn) {
			t.Errorf("pressed before the dropdown is up")
		}
		s.SetDropdownActive(true)
		if !w.dropdownPressed(btn) || w.dropdownPressed(dropdownButton(mjPlaying)) {
			t.Errorf("press at %v: playlist button pressed %v, playing button %v", p, w.dropdownPressed(btn), w.dropdownPressed(dropdownButton(mjPlaying)))
		}
		s.ReleaseButton(p[0], p[1], true)
		if !w.dropdownPressed(btn) {
			t.Errorf("the release lets go of the open dropdown's button")
		}
		s.SetDropdownActive(false)
		if w.dropdownPressed(btn) {
			t.Errorf("still pressed with the dropdown closed")
		}
		s.SetDropdownActive(true) // another dropdown, not opened by this button
		if w.dropdownPressed(btn) {
			t.Errorf("pressed for a dropdown it didn't open")
		}
		s.SetDropdownActive(false)
	}
}

// Music Selection opens where OpenLoco opens it (createWindow 0x004C9C68)
// when the windows are the same: a click on the Jukebox swapped it with the
// time panel (bringToFront), so the time panel is tried first and the window
// goes left of it, bottom-aligned (298, 362). A news window up (goloco's tour
// paused after a speed record, OpenLoco's before it) leaves no free spot:
// the fallback (0, 30) — goloco's shot, not a placement bug.
func TestMusicSelectionPlacement(t *testing.T) {
	for _, news := range []bool{false, true} {
		s := &WindowStack{ScreenW: 800, ScreenH: 600}
		s.Open(NewBarWindow(WindowTypeTopToolbar, 0, 0, 800, 28))
		s.Open(NewBarWindow(WindowTypeCompanyInfoPanel, 0, 573, 140, 27))
		s.Open(NewBarWindow(WindowTypeTimePanel, 660, 573, 140, 27))
		if news {
			n := NewNewsWindow(true, false, false, 800, 600)
			n.WinFlags |= WindowFlagStickToFront
			s.Open(n)
		}
		j := NewMusicJukebox(800, 600, 0, 1)
		s.Open(j)
		s.Press(j.X+100, j.Y+60) // the panel: the Jukebox to the front
		sel := NewMusicSelection(nil, 0, 1)
		s.Place(sel)
		want := [2]int{298, 362}
		if news {
			want = [2]int{0, 30}
		}
		if sel.X != want[0] || sel.Y != want[1] {
			t.Errorf("news %v: at %d,%d, want %v", news, sel.X, sel.Y, want)
		}
	}
}
