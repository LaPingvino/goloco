package ui

// Where a new window opens when it asks for no position: WindowManager::
// createWindow(type, size, flags, events) (OpenLoco reference:
// src/Ui/WindowManager.cpp, 0x004C9C68) with windowFitsWithinSpace,
// windowFitsOnScreen (0x004C9BA2) and createWindowOnScreen's clamp (0x004C9F27).

// WinRect is an open window's frame; StickToBack ones (the main view) don't
// block a position.
type WinRect struct {
	X, Y, W, H  int
	StickToBack bool
}

// AutoWindowPosition is the origin upstream gives a new window of w × h on a
// screen of screenW × screenH with the windows open: (0, 30), then the three
// other corners, then beside / above / below each window where nothing
// overlaps, then where it merely stays mostly on screen, else (0, 30) stepped
// by 5 while a window sits exactly there; clamped onto the screen.
func AutoWindowPosition(w, h, screenW, screenH int, wins []WinRect) (int, int) {
	fits := func(x, y int) bool { // windowFitsWithinSpace
		// Vanilla quirk kept: the bottom edge is checked against the screen's WIDTH.
		if x < 0 || y < 28 || x+w > screenW || y+h > screenW {
			return false
		}
		for _, o := range wins {
			if o.StickToBack {
				continue
			}
			if x+w <= o.X || x > o.X+o.W || y+h <= o.Y || y >= o.Y+o.H {
				continue
			}
			return false
		}
		return true
	}
	onScreen := func(x, y int) bool { // windowFitsOnScreen
		if x < -(w/4) || x > screenW-w/2 || y < 28 || y > screenH-h/4 {
			return false
		}
		return fits(x, y)
	}
	clamp := func(x, y int) (int, int) { // createWindowOnScreen
		x = min(max(x, 0), max(0, screenW-w))
		y = min(max(y, 28), max(28, screenH-h))
		return x, y
	}
	for _, p := range [][2]int{{0, 30}, {screenW - w, 30}, {0, screenH - h - 29}, {screenW - w, screenH - h - 29}} {
		if fits(p[0], p[1]) {
			return clamp(p[0], p[1])
		}
	}
	for _, o := range wins {
		if o.StickToBack {
			continue
		}
		for _, p := range [][2]int{
			{o.X + o.W + 2, o.Y}, {o.X - w - 2, o.Y}, {o.X, o.Y + o.H + 2}, {o.X, o.Y - h - 2},
			{o.X + o.W + 2, o.Y + o.H - h}, {o.X - w - 2, o.Y + o.H - h}, {o.X + o.W - w, o.Y - h - 2},
		} {
			if fits(p[0], p[1]) {
				return clamp(p[0], p[1])
			}
		}
	}
	for _, o := range wins {
		if o.StickToBack {
			continue
		}
		for _, p := range [][2]int{{o.X + o.W + 2, o.Y}, {o.X - w - 2, o.Y}, {o.X, o.Y + o.H + 2}, {o.X, o.Y - h - 2}} {
			if onScreen(p[0], p[1]) {
				return clamp(p[0], p[1])
			}
		}
	}
	x, y := 0, 30
	for again := true; again; {
		again = false
		for _, o := range wins {
			if o.X == x && o.Y == y {
				x, y, again = x+5, y+5, true
				break
			}
		}
	}
	return clamp(x, y)
}

// AutoPosition places a new window as upstream would among the manager's
// visible windows (none of goloco's SimpleWindows sticks to the back).
func (wm *SimpleWindowManager) AutoPosition(w, h, screenW, screenH int) (int, int) {
	var wins []WinRect
	for _, o := range wm.Windows {
		if o.Visible {
			wins = append(wins, WinRect{X: o.X, Y: o.Y, W: o.Width, H: o.Height})
		}
	}
	return AutoWindowPosition(w, h, screenW, screenH, wins)
}
