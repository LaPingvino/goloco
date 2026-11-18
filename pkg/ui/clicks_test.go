package ui

import (
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// buttonWin is a window with one button at (10, 20), 30 × 12 that its
// HandleClick names (Button), unless disabled.
type buttonWin struct {
	WinBase
	disabled bool
	closeOnPress bool
}

func (w *buttonWin) Draw(*ebiten.Image) {}
func (w *buttonWin) HandleClick(x, y int) {
	if !w.disabled && w.Button([4]int{10, 20, 30, 12}, x, y) && w.closeOnPress {
		w.Close()
	}
}

// MouseInput's clicks: clickDown at window x + midX on the press of a
// widget not disabled; clickUp on the release over it (also when goloco's
// window closed on the press: upstream's is still there); over no window
// (the main window's viewport) at its midX; none with the press's dropdown
// up; an unmoved caption release sounds clickPress at (w − 1) / 2.
func TestClickSounds(t *testing.T) {
	newStack := func() (*WindowStack, *buttonWin) {
		s := &WindowStack{ScreenW: 640, ScreenH: 480}
		w := &buttonWin{WinBase: WinBase{WinType: WindowTypeTown, X: 100, Y: 50, W: 200, H: 100}}
		s.Open(w)
		s.TakeOpenSound()
		return s, w
	}
	mid := 100 + MidX(10, 30) // 124
	if mid != 124 {
		t.Fatalf("midX %d, want 124 ((10 + 39) / 2 = 24)", mid)
	}

	s, _ := newStack()
	s.Press(115, 75)
	s.ReleaseButton(116, 76, false)
	if got, want := s.TakeClickSounds(), []ClickSound{{ClickDown, 124}, {ClickUp, 124}}; !reflect.DeepEqual(got, want) {
		t.Errorf("press and release on the button: %v, want %v", got, want)
	}

	s, _ = newStack()
	s.Press(115, 75)
	s.ReleaseButton(500, 400, false) // over no window: the main viewport's midX (639 / 2)
	if got, want := s.TakeClickSounds(), []ClickSound{{ClickDown, 124}, {ClickUp, 319}}; !reflect.DeepEqual(got, want) {
		t.Errorf("release over the map: %v, want %v", got, want)
	}

	s, _ = newStack()
	s.Press(115, 75)
	s.ReleaseButton(115, 75, true)
	if got, want := s.TakeClickSounds(), []ClickSound{{ClickDown, 124}}; !reflect.DeepEqual(got, want) {
		t.Errorf("release with the dropdown up: %v, want %v", got, want)
	}

	s, w := newStack()
	w.closeOnPress = true
	s.Press(115, 75)
	s.ReleaseButton(115, 75, false)
	if got, want := s.TakeClickSounds(), []ClickSound{{ClickDown, 124}, {ClickUp, 124}}; !reflect.DeepEqual(got, want) {
		t.Errorf("a button closing its window: %v, want %v", got, want)
	}

	s, w = newStack()
	w.disabled = true
	s.Press(115, 75)
	s.ReleaseButton(115, 75, false)
	if got := s.TakeClickSounds(); len(got) != 0 {
		t.Errorf("disabled button: %v, want none", got)
	}

	s, _ = newStack()
	s.Press(200, 90) // on the window, no widget it names
	s.ReleaseButton(200, 90, false)
	if got := s.TakeClickSounds(); len(got) != 0 {
		t.Errorf("no widget: %v, want none", got)
	}

	s, _ = newStack()
	s.Press(150, 55) // the caption, released unmoved
	s.Release(150, 55)
	if got, want := s.TakeClickSounds(), []ClickSound{{ClickPress, 100 + 99}}; !reflect.DeepEqual(got, want) {
		t.Errorf("caption click: %v, want %v", got, want)
	}

	s, _ = newStack()
	s.Press(150, 55) // the caption, dragged
	s.Held(170, 65)
	s.Release(170, 65)
	if got := s.TakeClickSounds(); len(got) != 0 {
		t.Errorf("caption drag: %v, want none", got)
	}
}
