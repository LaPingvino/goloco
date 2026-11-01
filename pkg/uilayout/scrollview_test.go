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
