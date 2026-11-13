package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// testWin is a bare window: its clicks counted.
type testWin struct {
	WinBase
	clicks int
}

func (w *testWin) Draw(*ebiten.Image)   {}
func (w *testWin) HandleClick(x, y int) { w.clicks++ }

func newTestWin(typ WindowType, number, x, y, w, h int) *testWin {
	return &testWin{WinBase: WinBase{WinType: typ, Number: number, X: x, Y: y, W: w, H: h}}
}

func order(s *WindowStack) []int {
	var out []int
	for _, w := range s.Windows {
		out = append(out, w.Base().Number)
	}
	return out
}

func sameOrder(a []int, b ...int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Several windows of one type, found by number; a press brings the window
// under it to the front (WindowManager::bringToFront: it changes places with
// the front-most one) and reaches it alone.
func TestWindowStackOrder(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	a, b, c := newTestWin(WindowTypeTown, 1, 0, 30, 200, 100), newTestWin(WindowTypeTown, 2, 50, 60, 200, 100), newTestWin(WindowTypeTown, 3, 100, 90, 200, 100)
	for _, w := range []*testWin{a, b, c} {
		s.Open(w)
	}
	if s.Find(WindowTypeTown, 2) != Win(b) || s.Find(WindowTypeTown, 9) != nil || s.Find(WindowTypeIndustry, 1) != nil {
		t.Fatal("Find by type and number")
	}
	if s.FindAt(120, 100) != Win(c) || s.FindAt(60, 70) != Win(b) || s.FindAt(10, 40) != Win(a) || s.FindAt(500, 500) != nil {
		t.Fatal("FindAt: the front-most window at the point")
	}
	if !s.Press(10, 50) || a.clicks != 1 || b.clicks+c.clicks != 0 {
		t.Fatalf("a press on the back window: clicks %d %d %d", a.clicks, b.clicks, c.clicks)
	}
	if !sameOrder(order(s), 3, 2, 1) { // a swapped with c, as upstream's std::swap
		t.Errorf("order after the press: %v", order(s))
	}
	if !s.IsInFront(a) || s.IsInFront(b) {
		t.Error("IsInFront")
	}
	if s.Press(700, 500) {
		t.Error("a press on no window was taken")
	}
	// A stickToFront window stays in front; new windows open behind it.
	top := newTestWin(WindowTypeScenarioSelect, 0, 300, 300, 100, 100)
	top.WinFlags = WindowFlagStickToFront
	s.Open(top)
	d := newTestWin(WindowTypeTown, 4, 300, 300, 100, 100)
	s.Open(d)
	if !sameOrder(order(s), 3, 2, 1, 4, 0) {
		t.Errorf("order with a stickToFront window: %v", order(s))
	}
	s.Press(55, 155) // b alone there
	if !sameOrder(order(s), 3, 4, 1, 2, 0) || !s.IsInFront(b) {
		t.Errorf("order after bringing 2 forward: %v", order(s))
	}
}

// Show on an open window's type and number replaces it where it stands and
// brings it forward; Refresh leaves the order; Close runs onClose once.
func TestWindowStackShowRefreshClose(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	a, b := newTestWin(WindowTypeVehicle, 7, 10, 40, 265, 177), newTestWin(WindowTypeVehicle, 8, 300, 40, 265, 177)
	closed := 0
	a.OnClose = func() { closed++ }
	s.Open(a)
	s.Open(b)
	a.X, a.Y = 33, 44
	a2 := newTestWin(WindowTypeVehicle, 7, 0, 0, 265, 177)
	a2.MinW, a2.MinH, a2.MaxW, a2.MaxH = 192, 182, 400, 440 // another tab's limits: the size kept within them
	a2.OnClose = a.OnClose
	if s.Refresh(a2); a2.X != 33 || a2.Y != 44 || a2.H != 182 || !sameOrder(order(s), 7, 8) || s.Find(WindowTypeVehicle, 7) != Win(a2) {
		t.Errorf("Refresh: at %d,%d %dx%d, order %v", a2.X, a2.Y, a2.W, a2.H, order(s))
	}
	if closed != 0 || a.Stack() != nil {
		t.Error("a replaced window closed, or kept its stack")
	}
	a3 := newTestWin(WindowTypeVehicle, 7, 0, 0, 265, 177)
	a3.OnClose = a.OnClose
	if isNew := s.Show(a3); isNew || a3.X != 33 || a3.H != 177 || !sameOrder(order(s), 8, 7) {
		t.Errorf("Show on an open window: new %v, at %d, h %d, order %v", isNew, a3.X, a3.H, order(s))
	}
	if s.Refresh(newTestWin(WindowTypeVehicle, 9, 0, 0, 10, 10)) || len(s.Windows) != 2 {
		t.Error("Refresh opened a window")
	}
	a3.Close()
	if closed != 1 || len(s.Windows) != 1 || s.Find(WindowTypeVehicle, 7) != nil {
		t.Errorf("Close: closed %d, %d windows", closed, len(s.Windows))
	}
	s.CloseAll()
	if len(s.Windows) != 0 {
		t.Error("CloseAll")
	}
}

// stateResizing: the grip (the 19 px corner) of a resizable window resizes
// within its limits; a click on it maximises (the screen less 27), another
// restores; a window with both sizes fixed has no grip.
func TestWindowStackResize(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	w := newTestWin(WindowTypeTown, 1, 100, 100, 223, 161)
	w.WinFlags = WindowFlagResizable
	w.SetSize(192, 161, 600, 440)
	s.Open(w)
	if !w.CanResize() || !w.HasResizeHandle() {
		t.Fatal("not resizable")
	}
	s.Press(100+223-5, 100+161-5)
	if !s.Dragging() || w.clicks != 0 {
		t.Fatal("a press on the grip doesn't start resizing")
	}
	s.Held(100+223-5+50, 100+161-5+30)
	if w.W != 273 || w.H != 191 {
		t.Errorf("resized to %dx%d, want 273x191", w.W, w.H)
	}
	s.Held(0, 0) // far past the minimum
	s.Release(0, 0)
	if w.W != 192 || w.H != 161 || s.Dragging() {
		t.Errorf("clamped to %dx%d, want the minimum 192x161", w.W, w.H)
	}
	gx, gy := w.X+w.W-5, w.Y+w.H-5
	s.Press(gx, gy)
	s.Release(gx, gy) // a click: maximise — as wide and high as the limits and the screen allow
	if w.W != 600 || w.H != 440 || !w.HasFlags(WindowFlagMaximised) {
		t.Errorf("maximised to %dx%d flags %#x", w.W, w.H, w.WinFlags)
	}
	gx, gy = w.X+w.W-5, w.Y+w.H-5
	s.Press(gx, gy)
	s.Release(gx, gy)
	if w.W != 192 || w.H != 161 || w.HasFlags(WindowFlagMaximised) {
		t.Errorf("restored to %dx%d", w.W, w.H)
	}
	// Fixed in one direction: resizes (canResize), draws no handle.
	w.SetSize(299, 172, 299, 327)
	if !w.CanResize() || w.HasResizeHandle() {
		t.Error("a window fixed in width: canResize yes, handle no")
	}
	w.SetFixedSize(340, 208)
	s.Press(w.X+w.W-5, w.Y+w.H-5)
	if s.Dragging() || w.clicks != 1 {
		t.Error("a fixed-size window resized")
	}
}

// relocateWindows and moveOtherWindowsDown.
func TestWindowStackRelocate(t *testing.T) {
	s := &WindowStack{ScreenW: 1024, ScreenH: 768}
	a, b := newTestWin(WindowTypeTown, 1, 900, 100, 100, 100), newTestWin(WindowTypeTown, 2, 100, 700, 100, 100)
	c := newTestWin(WindowTypeTown, 3, 100, 100, 100, 100)
	s.Open(a)
	s.Open(b)
	s.Open(c)
	s.Resize(800, 600)
	if a.X != 8 || a.Y != 36 || b.X != 16 || b.Y != 44 || c.X != 100 {
		t.Errorf("relocated to %d,%d and %d,%d; c at %d", a.X, a.Y, b.X, b.Y, c.X)
	}
	list := newTestWin(WindowTypeTownList, 0, 0, 30, 600, 197)
	s.Open(list)
	s.MoveOtherWindowsDown(list)
	if a.Y != 230 || c.Y != 230 {
		t.Errorf("moved down to %d and %d, want 230", a.Y, c.Y)
	}
}
