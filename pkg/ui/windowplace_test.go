package ui

import "testing"

func TestAutoWindowPosition(t *testing.T) {
	const sw, sh = 800, 600
	cases := []struct {
		name string
		wins []WinRect
		x, y int
	}{
		{"empty screen: top left under the toolbar", nil, 0, 30},
		{"the main view sticks to the back", []WinRect{{X: 0, Y: 0, W: 800, H: 600, StickToBack: true}}, 0, 30},
		{"top left taken: top right", []WinRect{{X: 0, Y: 30, W: 200, H: 150}}, 800 - 138, 30},
		{"both top corners: bottom left", []WinRect{{X: 0, Y: 30, W: 200, H: 150}, {X: 600, Y: 30, W: 200, H: 150}}, 0, 600 - 220 - 29},
	}
	for _, c := range cases {
		if x, y := AutoWindowPosition(138, 220, sw, sh, c.wins); x != c.x || y != c.y {
			t.Errorf("%s: (%d, %d), want (%d, %d)", c.name, x, y, c.x, c.y)
		}
	}
	// All four corners taken: beside the first window.
	all := []WinRect{{X: 0, Y: 30, W: 200, H: 300}, {X: 600, Y: 30, W: 200, H: 300}, {X: 0, Y: 340, W: 200, H: 260}, {X: 600, Y: 340, W: 200, H: 260}}
	if x, y := AutoWindowPosition(138, 220, sw, sh, all); x != 202 || y != 30 {
		t.Errorf("beside the first: (%d, %d), want (202, 30)", x, y)
	}
}
