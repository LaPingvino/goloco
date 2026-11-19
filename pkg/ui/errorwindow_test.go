package ui

import "testing"

// createErrorWindow's size and place: the text's width (70 … 250) + 2 × 4 by
// its lines × 10 + 8; openWithCompetitor's 86 wider and 42 higher, both before
// the window is placed round the mouse (−width / 2, +26, clamped).
func TestErrorWindowSize(t *testing.T) {
	fixedFont(t)
	e := OpenError("abcdefghijklmno", "", 400, 300, 800, 600) // 15 × 6 = 90 wide
	if e.W != 98 || e.H != 18 || e.X != 400-49 || e.Y != 326 {
		t.Errorf("plain: %d,%d %d×%d", e.X, e.Y, e.W, e.H)
	}
	c := OpenErrorWithCompetitor("abcdefghijklmno", "", 1, 0, false, 400, 300, 800, 600)
	if c.W != 98+86 || c.H != 18+42 || c.X != 400-(98+86)/2 || c.Y != 326 || c.Face != 1 {
		t.Errorf("competitor: %d,%d %d×%d", c.X, c.Y, c.W, c.H)
	}
	c = OpenErrorWithCompetitor("abc", "", 1, 0, false, 790, 590, 800, 600) // clamped: x ≤ 800 − w − 40, y ≤ 600 − h − 40
	if c.X != 800-c.W-40 || c.Y != 600-c.H-40 {
		t.Errorf("clamped: %d,%d %d×%d", c.X, c.Y, c.W, c.H)
	}
}
