package uilayout

import "testing"

func TestVThumb(t *testing.T) {
	// 356 px list, 30 rows of 24 px = 720: thumb ≈ half the bar, at the top.
	top, bot := VThumb(356, 720, 0)
	if top != 11 || bot-top < 160 || bot-top > 170 {
		t.Errorf("top of list: %d..%d", top, bot)
	}
	// scrolled to the end: thumb ends at the bottom button.
	_, bot = VThumb(356, 720, 720-356)
	if bot != 356-11-1 {
		t.Errorf("end of list: bottom %d", bot)
	}
	// a very long list keeps the minimum thumb.
	top, bot = VThumb(356, 100000, 0)
	if bot-top != MinThumbSize {
		t.Errorf("min thumb: %d", bot-top)
	}
}

// The vertical bar's parts and moves (ScrollView.cpp).
func TestVScrollInput(t *testing.T) {
	const h, content = 100, 400
	if p := VScrollPart(5, h, content, 0); p != "up" {
		t.Errorf("top: %s", p)
	}
	if p := VScrollPart(95, h, content, 0); p != "down" {
		t.Errorf("bottom: %s", p)
	}
	top, _ := VThumb(h, content, 0)
	if p := VScrollPart(top+1, h, content, 0); p != "thumb" {
		t.Errorf("thumb: %s", p)
	}
	if p := VScrollPart(80, h, content, 0); p != "trackBottom" {
		t.Errorf("below the thumb: %s", p)
	}
	if VScrollClick("down", h, content, 0) != 3 || VScrollClick("up", h, content, 2) != 0 {
		t.Error("buttons move 3")
	}
	if VScrollClick("trackBottom", h, content, 0) != 98 || VScrollClick("trackBottom", h, content, 300) != 302 { // height() − 2; contentHeight − (height() − 2)
		t.Errorf("page: %d %d", VScrollClick("trackBottom", h, content, 0), VScrollClick("trackBottom", h, content, 300))
	}
	if got := VScrollDrag(h, content, 0, 10); got != 10*400/80 { // verticalFollow: height() − 2 × kThumbSize
		t.Errorf("drag: %d", got)
	}
}

// The horizontal bar (Finances: 499 wide, 4 years = 4 × 128 + 1).
func TestHScrollInput(t *testing.T) {
	const w, content = 499, 4*128 + 1
	if p := HScrollPart(3, w, content, 0); p != "left" {
		t.Errorf("left: %s", p)
	}
	if p := HScrollPart(495, w, content, 0); p != "right" {
		t.Errorf("right: %s", p)
	}
	if got := HScrollLatest(w, content); got != 513-497 {
		t.Errorf("latest: %d", got)
	}
	if HScrollClick("right", w, content, 15) != 16 || HScrollClick("trackLeft", w, content, 16) != 0 {
		t.Error("clamps")
	}
	l, _ := HThumb(w, content, 16)
	if p := HScrollPart(l+1, w, content, 16); p != "thumb" {
		t.Errorf("thumb: %s", p)
	}
}

// A held down button repeats 3 a tick while on it; off it, unpressed and
// still; the release ends it; a held thumb follows the pointer.
func TestVScrollHold(t *testing.T) {
	const h, content = 100, 400
	var s VScrollHold
	off := s.Begin(95, 595, h, content, 0)
	if off != 3 || s.PressedPart() != "down" {
		t.Fatalf("begin: %d %q", off, s.PressedPart())
	}
	off = s.Continue(true, true, 95, 595, h, content, off)
	off = s.Continue(true, true, 95, 595, h, content, off)
	if off != 9 {
		t.Fatalf("held: %d", off)
	}
	off = s.Continue(true, false, 95, 595, h, content, off)
	if off != 9 || s.PressedPart() != "" {
		t.Fatalf("off the bar: %d %q", off, s.PressedPart())
	}
	off = s.Continue(false, true, 95, 595, h, content, off)
	if s.Part != "" || s.Continue(true, true, 95, 595, h, content, off) != 9 {
		t.Fatal("not released")
	}
	top, _ := VThumb(h, content, 0)
	off = s.Begin(top+1, 500, h, content, 0)
	off = s.Continue(true, true, 0, 510, h, content, off)
	if off != 10*400/80 || s.PressedPart() != "thumb" {
		t.Fatalf("thumb: %d %q", off, s.PressedPart())
	}
}

// DragFollow: verticalDragFollow / horizontalDragFollow's arithmetic, worked
// by hand from ScrollView.cpp.
func TestDragFollow(t *testing.T) {
	for _, c := range []struct {
		bar, other                   bool
		size, content, offset, delta int
		want                         int
	}{
		{true, false, 200, 721, 0, 10, 40},   // track 178: 7210 / 178
		{true, false, 200, 721, 30, -10, 0},  // −40 from 30: kept at 0
		{true, false, 200, 721, 0, 300, 523}, // to the end: 721 − 198
		{true, false, 200, 721, 100, -1, 96}, // −721 / 178 = −4 (towards zero)
		{true, true, 499, 513, 0, 50, 27},    // the other bar: track 466, end 513 − 486
		{false, false, 200, 721, 5, 10, 5},   // no bar: nothing moves
		{true, false, 200, 101, 0, 50, 0},    // content within the frame: stays at 0
		{true, false, 22, 721, 7, 10, 7},     // a track of 0: nothing moves
	} {
		if got := DragFollow(c.bar, c.other, c.size, c.content, c.offset, c.delta); got != c.want {
			t.Errorf("DragFollow(%v, %v, %d, %d, %d, %d) = %d, want %d", c.bar, c.other, c.size, c.content, c.offset, c.delta, got, c.want)
		}
	}
}

// The map window's scroll view at 350 × 272 (MapWindow.cpp: the widget
// {3, 44} to width − 108, height − 14 = 240 × 215; content 769 × 769, both
// bars), worked by hand from ScrollView.cpp.
var mapBars = Bars{W: 240, H: 215, ContentW: 769, ContentH: 769, HBar: true, VBar: true}

// calculateThumbSizeAndPosition with the other bar: 229 × 207 / 769 → 61 wide,
// 146 to travel; 204 × 182 / 769 → 48 tall.
func TestBarsThumbs(t *testing.T) {
	if l, r := mapBars.HThumb(0); l != 11 || r != 72 {
		t.Errorf("h thumb at 0: %d..%d", l, r)
	}
	if l, r := mapBars.HThumb(540); l != 157 || r != 217 { // the end clamped to 229 − 11 − 1
		t.Errorf("h thumb at 540: %d..%d", l, r)
	}
	if tp, b := mapBars.VThumb(0); tp != 11 || b != 59 {
		t.Errorf("v thumb at 0: %d..%d", tp, b)
	}
}

// getPart: the horizontal bar takes the bottom 12 rows (the corner too), its
// right button 10 px before the corner's deadzone; the vertical bar the
// right 12 columns above it; the view starts 1 px in.
func TestBarsPart(t *testing.T) {
	for _, c := range []struct {
		x, y int
		want Part
	}{
		{5, 208, PartHLeft}, {10, 203, PartHLeft}, {11, 208, PartHThumb}, {72, 208, PartHThumb},
		{73, 208, PartHTrackRight}, {217, 208, PartHTrackRight}, {218, 208, PartHRight},
		{227, 208, PartHRight}, {228, 208, PartNone}, {239, 214, PartNone},
		{230, 5, PartVTop}, {228, 10, PartVTop}, {230, 11, PartVThumb}, {230, 59, PartVThumb},
		{230, 60, PartVTrackBottom}, {230, 193, PartVBottom}, {230, 202, PartVBottom},
		{0, 50, PartNone}, {50, 0, PartNone}, {227, 202, PartView},
	} {
		if p, _, _ := mapBars.Part(c.x, c.y, 0, 0); p != c.want {
			t.Errorf("Part(%d, %d) = %d, want %d", c.x, c.y, p, c.want)
		}
	}
	if p, cx, cy := mapBars.Part(50, 60, 460, 200); p != PartView || cx != 509 || cy != 259 {
		t.Errorf("view point: %d %d,%d", p, cx, cy)
	}
	// Without the vertical bar the horizontal one runs to the widget's end.
	h := Bars{W: 240, H: 215, ContentW: 769, ContentH: 100, HBar: true}
	if p, _, _ := h.Part(229, 208, 0, 0); p != PartHRight {
		t.Errorf("no v bar: %d", p)
	}
}

// The buttons step 3, the tracks a page (the widget − 2 − the other bar),
// both kept within contentWidth − that page; a thumb dragged moves the
// content by delta × content / (widget − 20 − the other bar).
func TestBarsClickFollow(t *testing.T) {
	if x, _ := mapBars.Click(PartHRight, 541, 0); x != 542 { // 769 − 227
		t.Errorf("right at the end: %d", x)
	}
	if x, _ := mapBars.Click(PartHTrackRight, 0, 0); x != 227 {
		t.Errorf("track right: %d", x)
	}
	if _, y := mapBars.Click(PartVTrackBottom, 0, 0); y != 202 {
		t.Errorf("track bottom: %d", y)
	}
	if x, y := mapBars.Click(PartHLeft, 2, 7); x != 0 || y != 7 {
		t.Errorf("left: %d %d", x, y)
	}
	if got := mapBars.HFollow(0, 10); got != 36 { // 7690 / 209
		t.Errorf("h follow: %d", got)
	}
	if got := mapBars.VFollow(0, 10); got != 41 { // 7690 / 184
		t.Errorf("v follow: %d", got)
	}
	if got := mapBars.VFollow(560, 100); got != 567 { // 769 − 202
		t.Errorf("v follow clamp: %d", got)
	}
}

// A held button repeats on its part, shows unpressed off it; a held thumb
// follows the pointer while it is on the widget (off it only the pressed
// look clears); the view part reports its drag.
func TestScrollHold(t *testing.T) {
	var s ScrollHold
	x, y, view, _, _ := s.Begin(mapBars, 220, 208, 223, 252, 0, 0)
	if x != 3 || y != 0 || view || s.PressedPart() != PartHRight {
		t.Fatalf("begin: %d %d %v %d", x, y, view, s.PressedPart())
	}
	x, _, _, _, _ = s.Continue(mapBars, true, 220, 208, 223, 252, x, y)
	if x != 6 {
		t.Fatalf("held: %d", x)
	}
	x, _, _, _, _ = s.Continue(mapBars, true, 100, 208, 103, 252, x, y)
	if x != 6 || s.PressedPart() != PartNone {
		t.Fatalf("off the button: %d %d", x, s.PressedPart())
	}
	x, _, _, _, _ = s.Continue(mapBars, true, 220, 208, 223, 252, x, y)
	if x != 9 || s.PressedPart() != PartHRight {
		t.Fatalf("back on it: %d", x)
	}
	s.Release()
	if s.Part != PartNone {
		t.Fatal("not released")
	}
	x, _, _, _, _ = s.Begin(mapBars, 40, 208, 100, 252, 0, 0)
	if x != 0 || s.PressedPart() != PartNone {
		t.Fatalf("thumb begin: %d %d", x, s.PressedPart())
	}
	x, _, _, _, _ = s.Continue(mapBars, true, 50, 208, 110, 252, x, 0)
	if x != 36 || s.PressedPart() != PartHThumb {
		t.Fatalf("thumb: %d %d", x, s.PressedPart())
	}
	x, _, _, _, _ = s.Continue(mapBars, false, 50, 300, 110, 400, x, 0)
	if x != 36 || s.PressedPart() != PartNone {
		t.Fatalf("thumb off the widget: %d %d", x, s.PressedPart())
	}
	x, _, _, _, _ = s.Continue(mapBars, true, 60, 208, 120, 252, x, 0)
	if x != 72 {
		t.Fatalf("thumb back: %d", x)
	}
	s.Release()
	_, _, view, cx, cy := s.Begin(mapBars, 50, 60, 53, 104, 460, 200)
	if !view || cx != 509 || cy != 259 {
		t.Fatalf("view: %v %d %d", view, cx, cy)
	}
	if _, _, view, cx, _ = s.Continue(mapBars, true, 60, 60, 63, 104, 460, 200); !view || cx != 519 {
		t.Fatalf("view drag: %v %d", view, cx)
	}
	if _, _, view, _, _ = s.Continue(mapBars, true, 230, 60, 233, 104, 460, 200); view {
		t.Fatal("drag onto the bar still drags the view")
	}
}

// getPart's ends of the vertical bar alone: the up button is the top 11 rows,
// the down button the 10 above the bottom border row, the thumb's last row
// still the thumb (ScrollView.cpp l. 237–268).
func TestVScrollPartEdges(t *testing.T) {
	const h, content = 100, 400
	if p := VScrollPart(11, h, content, 0); p != "thumb" {
		t.Errorf("row 11: %s", p)
	}
	if p := VScrollPart(89, h, content, 0); p != "down" {
		t.Errorf("row 89: %s", p)
	}
	if p := VScrollPart(88, h, content, 0); p != "trackBottom" {
		t.Errorf("row 88: %s", p)
	}
	_, bot := VThumb(h, content, 0)
	if p := VScrollPart(bot, h, content, 0); p != "thumb" {
		t.Errorf("thumb's last row %d: %s", bot, p)
	}
}
