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
	if VScrollClick("trackBottom", h, content, 0) != 97 || VScrollClick("trackBottom", h, content, 300) != 303 {
		t.Errorf("page: %d %d", VScrollClick("trackBottom", h, content, 0), VScrollClick("trackBottom", h, content, 300))
	}
	if got := VScrollDrag(h, content, 0, 10); got != 10*400/77 {
		t.Errorf("drag: %d", got)
	}
}
