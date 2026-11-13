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

// The resizable windows' limits, by tab (each tab's onResize), and their
// layout following the size.
func TestResizableWindows(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	// Town: 223 × 161 within 192 × 161 … 600 × 440; its viewport follows.
	town := NewTownWindow(800, 600, "Town", "village", 100, 0, 1)
	town.WinType = WindowTypeTown
	town.X, town.Y = 100, 100
	s.Open(town)
	if town.W != 223 || town.H != 161 || !town.CanResize() || !town.HasResizeHandle() {
		t.Fatalf("town window %dx%d", town.W, town.H)
	}
	gx, gy := town.X+town.W-2, town.Y+town.H-2
	s.Press(gx, gy)
	s.Held(gx+100, gy+60)
	s.Release(gx+100, gy+60)
	if _, _, r, b := town.viewport(); town.W != 323 || town.H != 221 || r != 323-4 || b != 221-14 {
		t.Errorf("town window resized to %dx%d, viewport to %d,%d", town.W, town.H, r, b)
	}
	// Inside the viewport's part of the corner the press is the viewport's.
	if !town.GripBlocked(town.X+town.W-10, town.Y+town.H-16) || town.GripBlocked(town.X+town.W-2, town.Y+town.H-2) {
		t.Error("the town window's grip: the viewport blocks it, the corner itself doesn't")
	}
	// Industry: Production 299 × 282 … 337, no handle (fixed width); back on
	// the first tab the window is its kWindowSize again (switchTab).
	ind := NewIndustryWindow(800, 600, "Industry", "", 0, 1)
	ind.WinType = WindowTypeIndustry
	ind.Graph = func(i int) (GraphTab, bool) { return GraphTab{Layout: IndustryGraphLayout}, i == 1 || i == 2 }
	ind.X, ind.Y = 0, 434
	s.Open(ind)
	ind.HandleClick(ind.X+34+5, ind.Y+20) // the second tab
	if ind.Tab != 1 || ind.W != 299 || ind.H != 282 || !ind.CanResize() || ind.HasResizeHandle() {
		t.Errorf("production tab: tab %d, %dx%d", ind.Tab, ind.W, ind.H)
	}
	if ind.Y != 600-27-282 { // moveInsideScreenEdges: the window grew upwards
		t.Errorf("production tab at y %d, want %d", ind.Y, 600-27-282)
	}
	ind.HandleClick(ind.X+3+5, ind.Y+20)
	if ind.Tab != 0 || ind.W != 223 || ind.H != 137 {
		t.Errorf("back on the first tab: %dx%d", ind.W, ind.H)
	}
	// Station: the vehicle tabs 400 × 200 … 600 × 800; the size stays as the
	// tabs change (no setSize(kWindowSize) in its switchTab), within limits.
	st := NewStationWindow(800, 600, "Station", "", 0, 1)
	if st.W != 223 || st.H != 136 || st.MinW != 223 || st.MaxW != 600 {
		t.Errorf("station window %dx%d, limits %d…%d", st.W, st.H, st.MinW, st.MaxW)
	}
	st.Tab = 3
	st.PrepareDraw()
	if st.W != 400 || st.H != 200 {
		t.Errorf("vehicle tab %dx%d", st.W, st.H)
	}
	st.Tab = 2
	st.PrepareDraw()
	if st.W != 249 || st.H != 200 {
		t.Errorf("ratings tab %dx%d, want 249x200", st.W, st.H)
	}
	// A list: the scroll view and the status line keep their distance from
	// the right and the bottom.
	list := NewTownList(800, []TownRow{{Name: "A"}, {Name: "B"}}, 0, 1).Resizable(192, 100, 600, 900)
	l0, s0 := list.List, list.StatusY
	list.W, list.H = 500, 300
	list.PrepareDraw()
	if list.List[2] != l0[2]-100 || list.List[3] != l0[3]+103 || list.StatusY != s0+103 {
		t.Errorf("list resized: scroll view %v (was %v), status y %d (was %d)", list.List, l0, list.StatusY, s0)
	}
}
