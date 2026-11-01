package uilayout

import "testing"

func TestDropdownPlace(t *testing.T) {
	const w, h, iw = 100, 50, 96
	if x, y := DropdownPlace(10, 0, 30, 28, iw, w, h, 640, 480); x != 10 || y != 28 {
		t.Errorf("below: got %d,%d", x, y)
	}
	if _, y := DropdownPlace(10, 440, 30, 28, iw, w, h, 640, 480); y != 440-h {
		t.Errorf("above: y=%d", y)
	}
	if x, y := DropdownPlace(10, 200, 30, 28, iw, w, 470, 640, 480); y != 0 || x != 10+30+iw {
		t.Errorf("beside: got %d,%d", x, y)
	}
	if x, _ := DropdownPlace(600, 0, 30, 28, iw, w, h, 640, 480); x != 540 {
		t.Errorf("clamp: x=%d", x)
	}
	// beside would leave the screen on the right → to the left of the widget
	if x, y := DropdownPlace(560, 200, 30, 28, iw, w, 470, 640, 480); y != 0 || x != 560-(iw+4) {
		t.Errorf("beside left: got %d,%d", x, y)
	}
}
