package ui

import (
	"reflect"
	"testing"
)

// Real windows name their widgets: through the stack a press on Options'
// close button, a tab and a dropdown's button / combo box sounds clickDown at
// the window's x + that widget's midX (Options.cpp _widgets), and its
// release over it clickUp.
func TestOptionsClickSounds(t *testing.T) {
	press := func(rx, ry int, setup func(*OptionsWindow)) []ClickSound {
		s := &WindowStack{ScreenW: 640, ScreenH: 480}
		w := NewOptions(640, 480, 0, 0)
		if setup != nil {
			setup(w)
		}
		s.Open(w)
		s.TakeOpenSound()
		x, y := w.X+rx, w.Y+ry
		s.Press(x, y)
		s.ReleaseButton(x, y, false)
		return s.TakeClickSounds()
	}
	ww, _ := NewOptions(640, 480, 0, 0).size()
	x0 := (640 - ww) / 2
	cases := []struct {
		name   string
		rx, ry int
		setup  func(*OptionsWindow)
		mid    int // window-relative midX
	}{
		{"close button", ww - 10, 8, nil, MidX(ww-15, 13)},
		{"Audio tab", 3 + 31*OptTabAudio + 5, 20, nil, MidX(3+31*OptTabAudio, 31)},
		{"current tab", 3 + 31*OptTabDisplay + 5, 20, nil, MidX(3, 31)},
		{"screen mode button", 235 + 154 - 6, 68, nil, MidX(235+154-12, 11)},
		{"screen mode combo box", 240, 68, nil, MidX(235+154-12, 11)}, // findWidgetAt: a combobox is its button
		{"Misc checkbox", 20, 80, func(w *OptionsWindow) { w.Tab = OptTabMisc }, MidX(10, 400)},
	}
	for _, c := range cases {
		want := []ClickSound{{ClickDown, x0 + c.mid}, {ClickUp, x0 + c.mid}}
		if got := press(c.rx, c.ry, c.setup); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v, want %v", c.name, got, want)
		}
	}
	// The open-jukebox button is disabled without game music: silent.
	if got := press(optOpenJukebox[0]+5, optOpenJukebox[1]+5, func(w *OptionsWindow) { w.Tab, w.GameMusic = OptTabAudio, false }); len(got) != 0 {
		t.Errorf("disabled jukebox button: %v, want none", got)
	}
}

// A SimpleWindow's close button (the Construction window's, …) sounds at its
// midX; a press OnWindowClick does not name is silent.
func TestSimpleWindowCloseClickSound(t *testing.T) {
	s := &WindowStack{ScreenW: 640, ScreenH: 480}
	w := &SimpleWindow{WinBase: WinBase{WinType: WindowTypeConstruction, X: 50, Y: 60, W: 138, H: 276}}
	s.Open(w)
	s.TakeOpenSound()
	s.Press(50+138-10, 60+8)
	if got, want := s.TakeClickSounds(), []ClickSound{{ClickDown, 50 + MidX(138-15, 13)}}; !reflect.DeepEqual(got, want) {
		t.Errorf("close button: %v, want %v", got, want)
	}
	s = &WindowStack{ScreenW: 640, ScreenH: 480}
	w = &SimpleWindow{WinBase: WinBase{WinType: WindowTypeConstruction, X: 50, Y: 60, W: 138, H: 276}}
	w.OnWindowClick = func(int, int) {}
	s.Open(w)
	s.TakeOpenSound()
	s.Press(50+60, 60+100)
	if got := s.TakeClickSounds(); len(got) != 0 {
		t.Errorf("unnamed press: %v, want none", got)
	}
}

// The Miscellaneous tab's competing companies group: its steppers go from
// the scenario's value (-1) up to 14 competitors and 240 months, and each
// change is told; a press there is the widget's, not the frame's.
func TestOptionsCompetitorSteppers(t *testing.T) {
	s := &WindowStack{ScreenW: 640, ScreenH: 480}
	w := NewOptions(640, 480, 0, 0)
	w.Tab, w.Competitors, w.StartDelay = OptTabMisc, -1, 239
	s.Open(w)
	w.PrepareDraw()
	told := 0
	w.OnCompetitors = func() { told++ }
	click := func(rx, ry int) {
		s.Press(w.X+rx, w.Y+ry)
		s.ReleaseButton(w.X+rx, w.Y+ry, false)
	}
	minus, plus := 250+156-26+5, 250+156-13+5
	click(minus, optCompRows[1]+5) // already the scenario's: no change
	click(plus, optCompRows[1]+5)
	click(plus, optCompRows[1]+5)
	click(plus, optCompRows[2]+5)
	click(plus, optCompRows[2]+5) // 240 is the most
	if w.Competitors != 1 || w.StartDelay != 240 || told != 3 {
		t.Errorf("competitors %d, delay %d, told %d", w.Competitors, w.StartDelay, told)
	}
	if w.OnFrame(w.X+plus, w.Y+optCompRows[1]+5) {
		t.Error("the stepper's button is the window's frame")
	}
}

// The loaded-games rows: the setting's dropdown tells OnCompetitors; the
// game's offers "take over now" only with OpenLoco AI companies and an AI,
// and is dead on the title screen.
func TestOptionsLoadedAIs(t *testing.T) {
	s := &WindowStack{ScreenW: 640, ScreenH: 480}
	w := NewOptions(640, 480, 0, 0)
	w.Tab, w.CompetitorAI, w.GameOpenLocoAIs = OptTabMisc, "buslink", 0
	s.Open(w)
	w.PrepareDraw()
	var open *DropdownMenu
	w.OpenDropdown = func(d *DropdownMenu) { open = d }
	told, game := 0, -1
	w.OnCompetitors = func() { told++ }
	w.OnGameAIs = func(c int) { game = c }
	w.HandleClick(w.X+260, w.Y+optCompRows[3]+5)
	if open == nil {
		t.Fatal("no loaded-games dropdown")
	}
	open.OnSelect(1)
	if !w.TakeOverLoaded || told != 1 {
		t.Errorf("take over: %v, told %d", w.TakeOverLoaded, told)
	}
	open = nil
	w.HandleClick(w.X+260, w.Y+optCompRows[4]+5)
	if open == nil || len(open.Items) != 3 || !open.Items[GameAIsTakeOver].Disabled {
		t.Fatalf("the game's dropdown: %+v", open)
	}
	open.OnSelect(GameAIsTakeOver)
	if game != -1 {
		t.Error("a disabled take over now was told")
	}
	open.OnSelect(GameAIsKeep)
	if game != GameAIsKeep || !w.GameKeepAIs {
		t.Errorf("keep: told %d, keep %v", game, w.GameKeepAIs)
	}
	w.TitleMode, open = true, nil
	w.HandleClick(w.X+260, w.Y+optCompRows[4]+5)
	if open != nil {
		t.Error("the game's dropdown opened on the title screen")
	}
}
