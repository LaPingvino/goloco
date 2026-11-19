package ui

import "image"

// The widget click sounds: MouseInput.cpp's clickDown, clickUp and
// clickPress (OpenLoco reference: src/OpenLoco/src/Input/MouseInput.cpp).
//
//   - stateNormalLeft, a press on a widget that is not a caption, panel,
//     frame, viewport or scroll view and not disabled (Window::isEnabled is
//     !isDisabled since the enabledWidgets mask went): clickDown at the
//     window's x + the widget's midX, and the state is widgetPressed;
//   - stateWidgetPressed, the button coming up: clickUp at the x of the
//     window under the pointer + the midX of its widget there;
//   - statePositioningWindow, the button coming up on a window that never
//     moved: clickPress at the pressed caption / frame / panel's midX.
//
// goloco's windows find their widgets in HandleClick by position, so a
// window names the widget a press hit with WinBase.Button; the stack turns
// that into the sounds and the game plays them (TakeClickSounds).
//
// PARTIAL: a window's press on a widget its HandleClick does not look at (a
// label, the bare panel of a window without FrameHitter) is silent, where
// upstream sounds clickDown for every widget not disabled; the release off
// the pressed widget over another window (or another widget of it) is
// silent — upstream sounds clickUp at the midX of the widget under the
// pointer, which needs that window's widget table; the release of a press
// that opened a dropdown, over the dropdown (dropdownRegisterSelection's path:
// clickUp at the pressed window's x + the dropdown widget's midX).

// ClickSoundID is the Audio::SoundId of a click: what the game maps to its
// css1 sounds.
type ClickSoundID uint8

const (
	ClickDown  ClickSoundID = iota // SoundId::clickDown
	ClickUp                        // SoundId::clickUp
	ClickPress                     // SoundId::clickPress
)

// ClickSound is a click for the game to play on the ui channel: X is the
// screen x it pans to (window x + widget midX).
type ClickSound struct {
	ID ClickSoundID
	X  int
}

// MidX is Widget::midX for a widget at left of the given width: (left +
// right) / 2, right = left + width − 1 (makeWidget).
func MidX(left, width int) int { return (left + left + width - 1) / 2 }

// Button reports the screen point (x, y) on the window's widget r (left, top,
// width, height, in window coordinates) and, when it is, makes it the
// press's widget (stateNormalLeft's default case: _pressedWidgetIndex): the
// stack sounds clickDown at its midX, and clickUp when the button comes up
// over it. A window calls it for each widget it acts on, after its disabled
// check (a disabled widget is silent).
func (b *WinBase) Button(r [4]int, x, y int) bool {
	rx, ry := x-b.X, y-b.Y
	if rx < r[0] || ry < r[1] || rx >= r[0]+r[2] || ry >= r[1]+r[3] {
		return false
	}
	b.PressWidget(r)
	return true
}

// PressWidget makes the window's widget r (left, top, width, height, window
// coordinates) the press's widget: what Button does on a hit, for a window
// that found the widget its own way.
func (b *WinBase) PressWidget(r [4]int) {
	b.pressed = image.Rect(b.X+r[0], b.Y+r[1], b.X+r[0]+r[2], b.Y+r[1]+r[3])
	b.pressedHit = image.Rectangle{}
	b.hasPressed = true
}

// takePressed reports the widget a HandleClick named, in screen
// coordinates, and forgets it.
func (b *WinBase) takePressed() (image.Rectangle, bool) {
	r, ok := b.pressed, b.hasPressed
	b.hasPressed = false
	return r, ok
}

// pressHandled is HandleClick with stateNormalLeft's widget press: a widget
// the window named sounds clickDown and is the pressed one until the release
// (ReleaseButton).
func (s *WindowStack) pressHandled(w Win, x, y int) {
	b := w.Base()
	b.hasPressed = false
	w.HandleClick(x, y)
	if r, ok := b.takePressed(); ok {
		s.clicks = append(s.clicks, ClickSound{ClickDown, MidX(r.Min.X, r.Dx())})
		s.pressedBtn, s.hasPressedBtn = r, true
		s.pressedHit = r.Union(b.pressedHit)
	}
}

// ReleaseButton is stateWidgetPressed's release (MouseButton::leftReleased)
// after a widget press: over the pressed widget, clickUp at its midX (the
// window is still there upstream, even when goloco's acted on the press and
// closed); over no window of the stack — the main window, whose one widget
// is the viewport over the whole screen — clickUp at its midX. With the
// dropdown the press opened still up (dropdownOpen), the release on the
// widget only leaves it open (Dropdown flag unk1: no sound); elsewhere it
// goes to the dropdown.
func (s *WindowStack) ReleaseButton(x, y int, dropdownOpen bool) {
	if !s.hasPressedBtn {
		return
	}
	r := s.pressedBtn
	s.hasPressedBtn = false
	if dropdownOpen {
		return
	}
	switch {
	case image.Pt(x, y).In(r) || image.Pt(x, y).In(s.pressedHit):
		s.clicks = append(s.clicks, ClickSound{ClickUp, MidX(r.Min.X, r.Dx())})
	case s.FindAt(x, y) == nil:
		s.clicks = append(s.clicks, ClickSound{ClickUp, MidX(0, s.ScreenW)})
	}
}

// WidgetPressed reports a widget press waiting for its release
// (Input::State::widgetPressed).
func (s *WindowStack) WidgetPressed() bool { return s.hasPressedBtn }

// TakeClickSounds returns the clicks due since the last call.
func (s *WindowStack) TakeClickSounds() []ClickSound {
	c := s.clicks
	s.clicks = nil
	return c
}
