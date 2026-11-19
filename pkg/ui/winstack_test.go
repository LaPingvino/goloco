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

// openBars opens the game scene's bars as Gui::init does: the toolbar, the
// company panel, the time panel (a panel's widgets cover it; the toolbar's
// one button at 0–29).
func openBars(s *WindowStack) (top, player, tp *BarWindow) {
	top = NewBarWindow(WindowTypeTopToolbar, 0, 0, s.ScreenW, 28)
	top.WidgetAt = func(x, y int) bool { return x < 30 }
	player = NewBarWindow(WindowTypeCompanyInfoPanel, 0, s.ScreenH-27, 140, 27)
	tp = NewBarWindow(WindowTypeTimePanel, s.ScreenW-140, s.ScreenH-27, 140, 27)
	for _, b := range []*BarWindow{player, tp} {
		b.WidgetAt = func(x, y int) bool { return true }
	}
	for _, b := range []*BarWindow{top, player, tp} {
		s.Open(b)
	}
	return
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
	// New windows open behind a stickToFront window; a press brings a window
	// in front of it too (bringToFront swaps with the front-most window that
	// doesn't stick to the back).
	top := newTestWin(WindowTypeScenarioSelect, 0, 300, 300, 100, 100)
	top.WinFlags = WindowFlagStickToFront
	s.Open(top)
	d := newTestWin(WindowTypeTown, 4, 300, 300, 100, 100)
	s.Open(d)
	if !sameOrder(order(s), 3, 2, 1, 4, 0) {
		t.Errorf("order with a stickToFront window: %v", order(s))
	}
	s.Press(55, 155) // b alone there
	if !sameOrder(order(s), 3, 0, 1, 4, 2) || !s.IsInFront(b) {
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
	if town.OnFrame(town.X+town.W-10, town.Y+town.H-16) || !town.OnFrame(town.X+town.W-2, town.Y+town.H-2) {
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

// The bars on the stack (Gui::init's stickToFront windows): a new window
// opens behind them; pressed, it changes places with the time panel and lies
// over the toolbar and the panels, as in the real OpenLoco (a Town window
// dragged up over the toolbar and down over the company panel covers them:
// WindowManager::bringToFront, render's list order). The toolbar is under
// the mouse only on a button (noBackground).
func TestBarsOnTheStack(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	top, player, tp := openBars(s)
	a := newTestWin(WindowTypeTown, 1, 0, 30, 220, 160)
	s.Open(a)
	if s.Windows[0] != Win(a) || s.Windows[3] != Win(tp) {
		t.Fatalf("a new window opens behind the bars: %v", s.Windows)
	}
	if s.FindAt(10, 10) != Win(top) || s.FindAt(300, 10) != nil || s.FindAt(10, 590) != Win(player) {
		t.Error("FindAt: the toolbar on a button only, the panel everywhere")
	}
	a.Y = 560 // over the company panel, still behind it
	if s.FindAt(10, 590) != Win(player) {
		t.Error("the panel is in front of a window not yet pressed")
	}
	s.Press(100, 565) // the caption: brought to the front, moving
	s.Release(100, 565)
	if s.Windows[3] != Win(a) || s.Windows[0] != Win(tp) || s.FindAt(10, 590) != Win(a) {
		t.Errorf("pressed: swapped with the time panel, over the company panel: %v", s.Windows)
	}
	// The press on the panel brings nothing forward (stickToFront).
	s.Press(700, 590)
	if s.Windows[3] != Win(a) {
		t.Error("a press on the time panel reordered the list")
	}
}

// createWindow: a window that sticks neither way flashes its border for
// three ticks and sounds openWindow at its middle; a stickToFront one is
// quiet unless playSoundOnOpen; Show on an open window flashes it again
// (bringToFront(type, number)) without a sound.
func TestOpenFlashAndSound(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	a := newTestWin(WindowTypeTown, 1, 100, 30, 200, 160)
	s.Open(a)
	if pan, ok := s.TakeOpenSound(); !ok || pan != 200 || !a.WhiteBorder() {
		t.Fatalf("open: sound %v at %d, flash %v", ok, pan, a.WhiteBorder())
	}
	if _, ok := s.TakeOpenSound(); ok {
		t.Error("the sound was taken twice")
	}
	for range 2 {
		s.Tick()
	}
	if !a.WhiteBorder() {
		t.Error("the flash ended after two ticks")
	}
	if s.Tick(); a.WhiteBorder() {
		t.Error("the flash lasts three ticks")
	}
	news := newTestWin(WindowTypeNews, 0, 0, 500, 360, 100)
	news.WinFlags = WindowFlagStickToFront
	s.Open(news)
	if _, ok := s.TakeOpenSound(); ok || news.WhiteBorder() {
		t.Error("a stickToFront window opened with a sound")
	}
	browse := newTestWin(WindowTypeFileBrowserPrompt, 0, 300, 200, 100, 100)
	browse.WinFlags = WindowFlagStickToFront | WindowFlagPlaySoundOnOpen
	s.Open(browse)
	if _, ok := s.TakeOpenSound(); !ok || !browse.WhiteBorder() {
		t.Error("playSoundOnOpen: no sound")
	}
	a2 := newTestWin(WindowTypeTown, 1, 0, 0, 200, 160)
	if s.Show(a2) || !a2.WhiteBorder() {
		t.Error("Show on an open window: no flash")
	}
	if _, ok := s.TakeOpenSound(); ok {
		t.Error("Show on an open window sounded")
	}
	// A tab switch that rebuilds the window (Quietly): neither flash nor sound,
	// a new window included.
	for range 3 {
		s.Tick()
	}
	s.Quietly(func() {
		s.Show(newTestWin(WindowTypeTown, 1, 0, 0, 200, 160))
		s.Show(newTestWin(WindowTypeIndustry, 5, 0, 0, 200, 160))
	})
	if w := s.Find(WindowTypeTown, 1).Base(); w.WhiteBorder() {
		t.Error("a quiet rebuild flashed")
	}
	if _, ok := s.TakeOpenSound(); ok || s.Find(WindowTypeIndustry, 5).Base().WhiteBorder() {
		t.Error("a quiet open sounded or flashed")
	}
	// The browser modal: a press on another window brings it forward instead.
	s.Modal = WindowTypeFileBrowserPrompt
	for range 3 {
		s.Tick()
	}
	if !s.PressModal(150, 100) || !browse.WhiteBorder() || s.PressModal(350, 250) {
		t.Error("modal: a press elsewhere is taken and flashes the browser; on it, it isn't")
	}
	s.Close(browse)
	if s.PressModal(150, 100) || s.Modal != 0 {
		t.Error("modal after the browser closed")
	}
}

// A press on a window's bare panel moves it (stateNormalLeft: panel / frame
// → windowPositionBegin) when the window tells its widgets apart
// (FrameHitter); on a widget the window takes the press. The Build Vehicle
// window resizes (onResize: resizable, 380 × 233 … 700 × 600), by the grip
// only where its list doesn't reach.
func TestDragByFrame(t *testing.T) {
	s := &WindowStack{ScreenW: 1024, ScreenH: 768}
	w := &BuildVehicleWindow{WinBase: WinBase{WinType: WindowTypeBuildVehicle, X: 100, Y: 100, W: 500, H: 400}, NumTrackTabs: 2, Hover: -1}
	w.OnResize()
	s.Open(w)
	s.Press(100+200, 100+60) // the bare panel
	s.Held(100+230, 100+90)
	s.Release(100+230, 100+90)
	if w.X != 130 || w.Y != 130 || w.W != 500 {
		t.Errorf("dragged by the panel to %d,%d", w.X, w.Y)
	}
	s.Press(130+50, 130+150) // the list: not a move
	s.Held(130+80, 130+180)
	s.Release(130+80, 130+180)
	if w.X != 130 || w.Y != 130 {
		t.Errorf("a press on the list moved the window to %d,%d", w.X, w.Y)
	}
	if !w.CanResize() || w.MinW != 380 || w.MaxH != 600 {
		t.Errorf("build vehicle bounds %d…%d", w.MinW, w.MaxH)
	}
	gx, gy := w.X+w.W-2, w.Y+w.H-2 // the corner, right of the list (to width − 187)
	s.Press(gx, gy)
	s.Held(gx+50, gy+50)
	s.Release(gx+50, gy+50)
	if w.W != 550 || w.H != 450 {
		t.Errorf("resized to %dx%d", w.W, w.H)
	}
}

// stateNormalLeft on a Town window without a widget list of its own
// (findWidgetAt over ViewWindow.frameWidgets): the viewport and the status
// bar take a press, the bare panel moves the window — on the Population tab
// the graph is drawn on the bare panel, so a press anywhere on it moves the
// window (checked on the real OpenLoco, 2026-11-19: dragged by the graph).
func TestViewWindowDragByFrame(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	town := NewTownWindow(800, 600, "Town", "Town", 1000, 0, 1)
	town.WinType = WindowTypeTown
	town.X, town.Y = 100, 300
	town.Graph = func(i int) (GraphTab, bool) { return GraphTab{Layout: TownGraphLayout}, i == 1 }
	s.Open(town)
	drag := func(x, y, dx, dy int) (int, int) {
		ox, oy := town.X, town.Y
		s.Press(x, y)
		s.Held(x+dx, y+dy)
		s.Release(x+dx, y+dy)
		return town.X - ox, town.Y - oy
	}
	if dx, dy := drag(town.X+100, town.Y+80, 20, -30); dx != 0 || dy != 0 {
		t.Errorf("a press on the viewport moved the window by %d,%d", dx, dy)
	}
	if dx, dy := drag(town.X+50, town.Y+town.H-6, 20, -30); dx != 0 || dy != 0 {
		t.Errorf("a press on the status bar moved the window by %d,%d", dx, dy)
	}
	if dx, dy := drag(town.X+1, town.Y+80, 20, -30); dx != 20 || dy != -30 {
		t.Errorf("a press on the panel's edge moved the window by %d,%d, want 20,-30", dx, dy)
	}
	town.HandleClick(town.X+34+5, town.Y+20) // the Population tab
	if town.Tab != 1 {
		t.Fatalf("tab %d", town.Tab)
	}
	if dx, dy := drag(town.X+150, town.Y+100, 50, -100); dx != 50 || dy != -100 {
		t.Errorf("a press on the graph moved the window by %d,%d, want 50,-100", dx, dy)
	}
	if dx, dy := drag(town.X+3+31+5, town.Y+20, 50, 50); dx != 0 || dy != 0 {
		t.Errorf("a press on a tab moved the window by %d,%d", dx, dy)
	}
}
