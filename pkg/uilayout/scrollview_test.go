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
