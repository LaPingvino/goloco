package ui

import "testing"

// The news window's widgets as initViewport lays them out, and its slide-in.
func TestNewsWindowLayout(t *testing.T) {
	w := NewNewsWindow(true, false, true, 800, 600)
	if w.X != 220 || w.Y != 595 || w.W != 360 || w.H != 159 {
		t.Fatalf("a new paper: %d,%d %d×%d", w.X, w.Y, w.W, w.H)
	}
	if l, tp, r, b := w.viewport(0); l != 6 || tp != 86 || r != 351 || b != 149 {
		t.Errorf("viewport 1 alone: %d %d %d %d", l, tp, r, b)
	}
	if l, tp, r, b := w.button(0); l != 4 || tp != 84 || r != 355 || b != 158 {
		t.Errorf("button 1 alone: %d %d %d %d", l, tp, r, b)
	}
	w.HasSecond = true
	if l, _, r, _ := w.viewport(0); l != 6 || r != 174 {
		t.Errorf("viewport 1 beside a second: %d–%d", l, r)
	}
	if l, _, r, _ := w.viewport(1); l != 188 || r != 356 {
		t.Errorf("viewport 2: %d–%d", l, r)
	}
	// 4 px an update until it all shows.
	n := 0
	for ; w.Y > 600-159 && n < 100; n++ {
		w.Update(800, 600)
	}
	if n != 39 || w.Y != 441 {
		t.Errorf("slid in after %d updates, y %d", n, w.Y)
	}
	// Old news, or a faster game: at once; the advice panel is 117 high.
	if o := NewNewsWindow(false, true, true, 800, 600); o.Y != 600-117 || o.H != 117 {
		t.Errorf("old advice: y %d h %d", o.Y, o.H)
	}
	if o := NewNewsWindow(true, false, false, 800, 600); o.Y != 441 {
		t.Errorf("at speed: y %d", o.Y)
	}
}

// The subject buttons: the second over the first where they overlap, only
// selectable ones, and the close button.
func TestNewsWindowClicks(t *testing.T) {
	w := NewNewsWindow(true, true, true, 800, 600)
	w.HasSecond = true
	w.Subjects[0].Selectable, w.Subjects[1].Selectable = true, true
	var got []int
	closed := 0
	w.OnSubject = func(i int) { got = append(got, i) }
	w.OnDismiss = func() { closed++ }
	w.HandleClick(w.X+100, w.Y+120)
	w.HandleClick(w.X+300, w.Y+120)
	w.HandleClick(w.X+100, w.Y+40) // the text: nothing
	w.HandleClick(w.X+350, w.Y+8)
	if len(got) != 2 || got[0] != 0 || got[1] != 1 || closed != 1 {
		t.Errorf("subjects %v, closed %d", got, closed)
	}
	w.Subjects[1].Selectable = false
	got = nil
	w.HandleClick(w.X+300, w.Y+120)
	if len(got) != 0 {
		t.Errorf("a hidden button took the click: %v", got)
	}
}

// The ticker: findAt looks through it, findAtAlt doesn't; it stays where it
// was made until the screen changes; the mouse over it holds the writing 12
// updates.
func TestNewsTicker(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	tk := NewNewsTicker(800, 600)
	s.Open(tk)
	if tk.X != 662 || tk.Y != 575 {
		t.Fatalf("made at %d,%d", tk.X, tk.Y)
	}
	if s.FindAt(700, 590) != nil || s.FindAtAlt(700, 590) != Win(tk) {
		t.Error("findAt saw the ticker, or findAtAlt didn't")
	}
	if s.Press(700, 590) {
		t.Error("a press on the ticker went to a window")
	}
	for range 4 {
		tk.Update(800, 600, false, false)
	}
	if tk.numCharsToDisplay != 8 || tk.Y != 575 {
		t.Errorf("after 4 updates: %d quarter characters, y %d", tk.numCharsToDisplay, tk.Y)
	}
	tk.Update(800, 600, true, false) // the mouse on it: 12 updates of nothing (the first counts down at once)
	for range 11 {
		tk.Update(800, 600, false, false)
	}
	if tk.numCharsToDisplay != 10 || tk.var852 != 0 {
		t.Errorf("after the mouse's pause: %d, var852 %d", tk.numCharsToDisplay, tk.var852)
	}
	tk.Update(800, 600, false, true) // paused: nothing moves
	if tk.numCharsToDisplay != 10 {
		t.Errorf("paused: %d", tk.numCharsToDisplay)
	}
	tk.Update(1024, 768, false, false) // onResize
	if tk.X != 1024-138 || tk.Y != 768-26 {
		t.Errorf("after a resize: %d,%d", tk.X, tk.Y)
	}
	if got := tickerText("a\x05b"); got != "a  b" {
		t.Errorf("tickerText %q", got)
	}
}

// The news window among the windows a new one is placed by: it blocks its
// place at the screen's bottom, counted after the panels; it has no caption
// to move it by and stays in front.
func TestNewsWindowInStack(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	s.Fixed = []WinRect{{X: 0, Y: 0, W: 800, H: 28}, {X: 0, Y: 573, W: 140, H: 27}, {X: 660, Y: 573, W: 140, H: 27}}
	news := NewNewsWindow(true, true, true, 800, 600)
	s.Open(news)
	a := newTestWin(WindowTypeTown, 1, 0, 0, 300, 200)
	s.Open(a)
	if s.Windows[len(s.Windows)-1] != Win(news) {
		t.Error("a new window went in front of the news")
	}
	r := s.rects()
	if len(r) != 5 || r[0].W != 300 || r[1].H != 28 || r[4].W != 360 {
		t.Errorf("placement order: %+v", r)
	}
	// Three corners taken; the bottom right one would lie over the news
	// (x 400–799 against 220–579): the window goes beside the first instead.
	for i, p := range [][2]int{{0, 30}, {400, 30}, {0, 371}} {
		s.Open(newTestWin(WindowTypeTown, 2+i, p[0], p[1], 398, 200)) // 398: a window may not touch another (x > its right edge)
	}
	a.X, a.Y, a.W, a.H = -500, 30, 10, 10 // out of the way
	b := newTestWin(WindowTypeTown, 9, 0, 0, 400, 200)
	s.Place(b)
	if b.X == 400 && b.Y == 600-200-29 {
		t.Errorf("placed over the news window at %d,%d", b.X, b.Y)
	}
	s.CloseType(WindowTypeNews, AnyNumber)
	s.Place(b)
	if b.X != 400 || b.Y != 600-200-29 {
		t.Errorf("without the news: %d,%d, want the bottom right corner", b.X, b.Y)
	}
	// No caption: a press on its top edge is a click, not a move.
	news = NewNewsWindow(true, true, true, 800, 600)
	s.Open(news)
	s.Press(news.X+100, news.Y+5)
	if s.Dragging() {
		t.Error("the news window is being moved")
	}
}

// The Messages window: its list, a row's click, the tabs' sizes, the
// panel's resize handle on a window that only grows in height.
func TestMessagesWindow(t *testing.T) {
	w := NewMessagesWindow(800, 3, 4)
	if w.X != 434 || w.Y != 29 || w.W != 366 || w.H != 217 {
		t.Fatalf("opened at %d,%d %d×%d", w.X, w.Y, w.W, w.H)
	}
	rows := make([]MessageRow, 7)
	w.Rows = func() []MessageRow { return rows }
	w.ShowMessages()
	if l := w.list(); l != [4]int{3, 45, 360, 159} || w.scroll != 7*39-159 {
		t.Errorf("list %v, scrolled to %d", l, w.scroll)
	}
	if !w.CanResize() || w.HasResizeHandle() || !w.HasPanelResizeHandle() {
		t.Error("resizable in height only: no frame handle, the panel's")
	}
	got := -1
	w.OnRow = func(i int) { got = i }
	w.HandleClick(w.X+100, w.Y+46+(6*39-w.scroll)+5) // in the last row
	if got != 6 {
		t.Errorf("row %d, want 6", got)
	}
	w.MouseOver(w.X+100, w.Y+50)
	if w.hover != (w.scroll+4)/39 {
		t.Errorf("hover %d", w.hover)
	}
	w.MouseOver(w.X+100, w.Y+300)
	if w.hover != -1 {
		t.Errorf("hover off the list: %d", w.hover)
	}
	w.HandleClick(w.X+40, w.Y+20) // the settings tab
	w.PrepareDraw()
	if w.Tab != 1 || w.W != 366 || w.H != 155 || w.CanResize() {
		t.Errorf("settings: tab %d %d×%d", w.Tab, w.W, w.H)
	}
	var d *DropdownMenu
	set := [2]int{-1, -1}
	w.OpenDropdown = func(m *DropdownMenu) { d = m }
	w.OnSetting = func(c, v int) { set = [2]int{c, v} }
	w.Settings = [6]int{2, 2, 2, 2, 2, 2}
	w.HandleClick(w.X+350, w.Y+47+15*4+5) // general news
	if d == nil || len(d.Items) != 3 || !d.Items[2].Selected {
		t.Fatalf("dropdown %+v", d)
	}
	d.OnSelect(1)
	if set != [2]int{4, 1} {
		t.Errorf("setting %v, want general news → ticker", set)
	}
	toggled := false
	w.OnPlaySounds = func() { toggled = true }
	w.HandleClick(w.X+10, w.Y+140)
	if !toggled {
		t.Error("the checkbox")
	}
	w.HandleClick(w.X+10, w.Y+20) // back: the smallest size, the first message
	if w.Tab != 0 || w.H != 217 || w.scroll != 0 {
		t.Errorf("back on messages: tab %d h %d scroll %d", w.Tab, w.H, w.scroll)
	}
}

// Upstream's list order, which placement walks: bringToFront swaps a window
// with the front-most one of the list — a panel, or the news window. The
// sequence of scripts/olref/tours/news.txt, where the real OpenLoco opens
// the Town window at (0, 439): its corners are taken (the news window blocks
// the bottom ones), the time panel, the toolbar and the player panel come
// before the other windows in the list by then, and "below the player
// panel" passes windowFitsWithinSpace (which checks the bottom against the
// screen's width) and is clamped onto the screen.
func TestPlacementOrderWithNews(t *testing.T) {
	s := &WindowStack{ScreenW: 800, ScreenH: 600}
	s.Fixed = []WinRect{{X: 0, Y: 0, W: 800, H: 28}, {X: 0, Y: 573, W: 140, H: 27}, {X: 660, Y: 573, W: 140, H: 27}}
	msg := NewMessagesWindow(800, 0, 0)
	s.Open(msg)
	news := func() {
		s.CloseType(WindowTypeNews, AnyNumber)
		s.Open(NewNewsWindow(true, true, true, 800, 600))
	}
	s.Press(msg.X+100, msg.Y+100) // a row: the window to the front (swapped with the time panel), its news
	news()
	s.Press(msg.X+100, msg.Y+100) // another row: swapped with the news window
	news()
	ind := newTestWin(WindowTypeIndustry, 1, 0, 0, 223, 137)
	s.Place(ind) // the news' subject button: the Industry window
	if ind.X != 0 || ind.Y != 30 {
		t.Fatalf("the industry window at %d,%d, want 0,30", ind.X, ind.Y)
	}
	s.Open(ind)
	s.Press(msg.X+100, msg.Y+100)
	news()
	var kinds []int
	for _, e := range s.order {
		switch {
		case e.win == nil:
			kinds = append(kinds, e.fixed)
		case e.win == Win(msg):
			kinds = append(kinds, 10)
		case e.win == Win(ind):
			kinds = append(kinds, 11)
		default:
			kinds = append(kinds, 12)
		}
	}
	if !sameOrder(kinds, 2, 0, 1, 11, 10, 12) {
		t.Errorf("list order %v, want time panel, toolbar, player panel, industry, messages, news", kinds)
	}
	town := newTestWin(WindowTypeTown, 2, 0, 0, 223, 161)
	s.Place(town)
	if town.X != 0 || town.Y != 439 {
		t.Errorf("the town window at %d,%d, want 0,439 as OpenLoco", town.X, town.Y)
	}
	// A new scene: the list starts over.
	s.CloseAll()
	if r := s.rects(); len(r) != 3 || r[0].H != 28 {
		t.Errorf("after CloseAll: %+v", r)
	}
}
