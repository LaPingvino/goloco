package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// The window stack: WindowManager's list of windows (OpenLoco reference:
// src/OpenLoco/src/Ui/WindowManager.cpp) and the part of the mouse's state
// machine that moves and resizes them (src/OpenLoco/src/Input/MouseInput.cpp
// stateNormalLeft, statePositioningWindow, stateResizing).
//
// Every goloco window type embeds WinBase and so is a Win; the stack holds
// them back to front. A window is found by its type and number (the town,
// station, vehicle head or company it shows — WindowManager::find), so there
// can be one Town window per town and so on.
//
// goloco's windows are plain structs that their openers rebuild (a list
// repopulated, a vehicle window after a command), where upstream keeps one
// Window and changes its fields: Show and Replace put the new struct in the
// old one's slot, keeping its place, its size (within the new limits) and its
// position in the order.
//
// PARTIAL: the toolbar and the bottom panels are not windows of the stack
// (upstream: stickToFront windows); Fixed carries their rectangles for the
// placement. A press on a window's bare panel or frame does not move it
// (upstream: windowPositionBegin there too) — only the caption does. No white
// border flash (WindowFlags::whiteBorderMask), no open sound.

// WindowType is a window's type (Ui/WindowType.h), the half of its identity
// that its number completes.
type WindowType uint8

const (
	WindowTypeConstruction WindowType = iota + 32
	WindowTypeTown
	WindowTypeIndustry
	WindowTypeStation
	WindowTypeTownList
	WindowTypeIndustryList
	WindowTypeStationList
	WindowTypeVehicleList
	WindowTypeVehicle
	WindowTypeBuildVehicle
	WindowTypeCompany
	WindowTypeCompanyList
	WindowTypeOptions
	WindowTypeScenarioSelect
	WindowTypeFileBrowserPrompt
	WindowTypeMusicJukebox
	WindowTypeMusicSelection
	WindowTypeMessages
	WindowTypeNews
)

// WindowFlags are Ui::WindowFlags, those the stack acts on.
type WindowFlags uint32

const (
	WindowFlagStickToBack  WindowFlags = 1 << 0
	WindowFlagStickToFront WindowFlags = 1 << 1
	WindowFlagResizable    WindowFlags = 1 << 2
	WindowFlagNoAutoClose  WindowFlags = 1 << 5
	// WindowFlagIgnoreInFindAt (the news ticker): findAt looks through the
	// window — a press there is for what lies under it — findAtAlt does not.
	WindowFlagIgnoreInFindAt WindowFlags = 1 << 7
)

// Flags the stack keeps itself (MouseInput's stateResizing).
const (
	WindowFlagHasBeenResized WindowFlags = 1 << (16 + iota)
	WindowFlagMaximised
)

// kMaxWindows, kResizeHandleSize (Ui/Widget.h) and the caption widget every
// goloco window has ({1, 1}, {width − 2, 13}; the close button {width − 15,
// 2}, {13, 13} over its right end).
const (
	maxWindows       = 64
	ResizeHandleSize = 18
	captionTop       = 1
	captionBottom    = 13
)

// AnyNumber finds a window of a type whatever its number (WindowManager::
// find(type)).
const AnyNumber = -1

// WinBase is what the stack knows of a window (Ui::Window's type, number,
// x, y, width, height, min / max size and flags). OnClose is its onClose
// event: what goes with the window (a tool, a catchment display).
type WinBase struct {
	WinType                WindowType
	Number                 int
	X, Y, W, H             int
	MinW, MinH, MaxW, MaxH int
	WinFlags               WindowFlags
	OnClose                func()
	// NoCaption: the window has no caption widget (the news): a press on its
	// top edge doesn't move it.
	NoCaption bool

	stack              *WindowStack
	self               Win
	restoreW, restoreH int // var_88A, var_88C: the size before maximising
}

// Win is a window on the stack.
type Win interface {
	Base() *WinBase
	Draw(screen *ebiten.Image)
	// HandleClick is a left press on the window, in screen coordinates (not
	// on its caption or resize grip: the stack takes those; a caption's click
	// arrives when the button comes up without the window having moved).
	HandleClick(x, y int)
}

// Preparer is a window's prepareDraw / onResize: its size brought in line with
// its current tab, before the stack looks at its rectangle.
type Preparer interface{ PrepareDraw() }

// GripBlocker reports a widget other than the frame and the panel at a
// screen point inside the resize grip's corner: MouseInput starts resizing
// only on those two (stateNormalLeft).
type GripBlocker interface{ GripBlocked(x, y int) bool }

func (b *WinBase) Base() *WinBase { return b }

// Contains reports the screen point on the window.
func (b *WinBase) Contains(x, y int) bool {
	return x >= b.X && y >= b.Y && x < b.X+b.W && y < b.Y+b.H
}

func (b *WinBase) HasFlags(f WindowFlags) bool { return b.WinFlags&f != 0 }

// CanResize is Window::canResize.
func (b *WinBase) CanResize() bool {
	return b.HasFlags(WindowFlagResizable) && (b.MinW != b.MaxW || b.MinH != b.MaxH)
}

// HasResizeHandle is the frame's drawResizeHandle condition (FrameWidget.cpp):
// resizable and neither dimension fixed. The panel's own (PanelWidget.cpp,
// HasPanelResizeHandle) is what shows on a window with a panel.
func (b *WinBase) HasResizeHandle() bool {
	return b.HasFlags(WindowFlagResizable) && b.MinH != b.MaxH && b.MinW != b.MaxW
}

// HasPanelResizeHandle is the panel's drawResizeHandle condition
// (PanelWidget.cpp 0x004CAB8E): resizable and not both dimensions fixed — a
// window that resizes one way only has its handle too (the Messages window,
// the graph tabs).
func (b *WinBase) HasPanelResizeHandle() bool {
	return b.HasFlags(WindowFlagResizable) && !(b.MinH == b.MaxH && b.MinW == b.MaxW)
}

// SetSize is Window::setSize(minSize, maxSize): the limits set and the size
// brought within them; reports a change.
func (b *WinBase) SetSize(minW, minH, maxW, maxH int) bool {
	b.MinW, b.MinH, b.MaxW, b.MaxH = minW, minH, maxW, maxH
	w, h := min(max(b.W, minW), maxW), min(max(b.H, minH), maxH)
	changed := w != b.W || h != b.H
	b.W, b.H = w, h
	return changed
}

// SetFixedSize is Window::setSize(size).
func (b *WinBase) SetFixedSize(w, h int) bool { return b.SetSize(w, h, w, h) }

// ApplyLimits is the window's onResize after it is created or changes tab:
// setSize(min, max) with its own limits, when it has any.
func (b *WinBase) ApplyLimits() {
	if b.MaxW != 0 || b.MaxH != 0 {
		b.SetSize(b.MinW, b.MinH, b.MaxW, b.MaxH)
	}
}

// Close is WindowManager::close(window): its onClose event, then off the
// stack. A window's close button calls it.
func (b *WinBase) Close() {
	if b.stack != nil && b.self != nil {
		b.stack.Close(b.self)
		return
	}
	if b.OnClose != nil {
		b.OnClose()
	}
}

// Stack is the stack the window is on (nil: none).
func (b *WinBase) Stack() *WindowStack { return b.stack }

func (b *WinBase) inGrip(x, y int) bool {
	return x >= b.X+b.W-ResizeHandleSize-1 && y >= b.Y+b.H-ResizeHandleSize-1
}

func (b *WinBase) onCaption(x, y int) bool {
	rx, ry := x-b.X, y-b.Y
	if b.NoCaption {
		return false
	}
	if ry < captionTop || ry > captionBottom || rx < 1 || rx > b.W-2 {
		return false
	}
	return !(rx >= b.W-15 && rx < b.W-2 && ry >= 2 && ry < 15) // the close button
}

type stackDrag int

const (
	dragNone stackDrag = iota
	dragPositioning
	dragResizing
)

// WindowStack is WindowManager's _windows, back to front, and the input
// state's window drag.
type WindowStack struct {
	Windows []Win
	// ScreenW, ScreenH are Ui::width() and Ui::height().
	ScreenW, ScreenH int
	// Fixed are the rectangles of what upstream keeps as stickToFront windows
	// outside this stack (the top toolbar, the player and time panels): they
	// block and anchor a new window's placement.
	Fixed []WinRect

	// The drag: _dragWindowType / _dragWindowNumber, _dragLast, _windowMoved,
	// and where the press was (the caption's click on an unmoved release).
	drag             stackDrag
	dragType         WindowType
	dragNumber       int
	lastX, lastY     int
	pressX, pressY   int
	moved            bool
	captionClickable bool
}

func prepare(w Win) {
	if p, ok := w.(Preparer); ok {
		p.PrepareDraw()
	}
}

// Find is WindowManager::find(type, number) (AnyNumber: find(type)): the
// first such window, from the back.
func (s *WindowStack) Find(typ WindowType, number int) Win {
	for _, w := range s.Windows {
		if b := w.Base(); b.WinType == typ && (number == AnyNumber || b.Number == number) {
			return w
		}
	}
	return nil
}

func (s *WindowStack) indexOf(w Win) int {
	for i, o := range s.Windows {
		if o == w {
			return i
		}
	}
	return -1
}

// FindAt is WindowManager::findAt: the front-most window at a screen point.
func (s *WindowStack) FindAt(x, y int) Win {
	for i := len(s.Windows) - 1; i >= 0; i-- {
		w := s.Windows[i]
		prepare(w) // callOnResize
		if b := w.Base(); b.Contains(x, y) && !b.HasFlags(WindowFlagIgnoreInFindAt) {
			return w
		}
	}
	return nil
}

// FindAtAlt is WindowManager::findAtAlt: findAt that sees the windows
// ignoreInFindAt hides from it.
func (s *WindowStack) FindAtAlt(x, y int) Win {
	for i := len(s.Windows) - 1; i >= 0; i-- {
		w := s.Windows[i]
		prepare(w)
		if w.Base().Contains(x, y) {
			return w
		}
	}
	return nil
}

// BringToFront is WindowManager::bringToFront(Window&): the window changes
// places with the front-most one that does not stick to the front or back
// (upstream swaps the two — it does not move the others down), and a window
// almost off the screen's left comes back to x 20.
func (s *WindowStack) BringToFront(w Win) {
	b := w.Base()
	if b.HasFlags(WindowFlagStickToBack | WindowFlagStickToFront) {
		return
	}
	front := -1
	for i := len(s.Windows) - 1; i >= 0; i-- {
		// Upstream's loop skips stickToBack only; its stickToFront windows
		// (toolbars, panels) are swapped with too. goloco's are not on the
		// stack, and its own stickToFront ones are modal prompts that must
		// stay in front, so those are skipped as well.
		if !s.Windows[i].Base().HasFlags(WindowFlagStickToBack | WindowFlagStickToFront) {
			front = i
			break
		}
	}
	if i := s.indexOf(w); i >= 0 && front >= 0 && front != i {
		s.Windows[i], s.Windows[front] = s.Windows[front], s.Windows[i]
	}
	if b.X+b.W < 20 {
		b.X = 20
	}
}

// IsInFront is WindowManager::isInFront: nothing but stickToFront windows
// over it.
func (s *WindowStack) IsInFront(w Win) bool {
	i := s.indexOf(w)
	if i < 0 {
		return false
	}
	for _, o := range s.Windows[i+1:] {
		if !o.Base().HasFlags(WindowFlagStickToFront) {
			return false
		}
	}
	return true
}

// Open is WindowManager::createWindow(type, origin, size, flags, events)'s
// list work for a window already built and placed: at kMaxWindows the first
// window that may go is closed; the new one goes in front of every window but
// the stickToFront ones (behind the others if it sticks to the back).
func (s *WindowStack) Open(w Win) {
	if len(s.Windows) >= maxWindows {
		for _, o := range s.Windows {
			if !o.Base().HasFlags(WindowFlagStickToBack | WindowFlagStickToFront | WindowFlagNoAutoClose) {
				s.Close(o)
				break
			}
		}
	}
	b := w.Base()
	b.stack, b.self = s, w
	dst := 0
	switch {
	case b.HasFlags(WindowFlagStickToBack):
		for i, o := range s.Windows {
			if o.Base().HasFlags(WindowFlagStickToBack) {
				dst = i
			}
		}
	case b.HasFlags(WindowFlagStickToFront):
		dst = len(s.Windows)
	default:
		for i := len(s.Windows); i > 0; i-- {
			if !s.Windows[i-1].Base().HasFlags(WindowFlagStickToFront) {
				dst = i
				break
			}
		}
	}
	s.Windows = append(s.Windows, nil)
	copy(s.Windows[dst+1:], s.Windows[dst:])
	s.Windows[dst] = w
}

// Replace puts a rebuilt window in the slot of the one it replaces: the same
// place on the screen and in the order, the old size within the new one's
// limits (a tab's onResize: setSize(min, max)) — the new one's own size if
// either has no limits (a tab of one fixed size) —, the stack's resize flags.
// The old window's onClose does not run: it is the same window.
func (s *WindowStack) Replace(old, w Win) {
	i := s.indexOf(old)
	if i < 0 || old == w {
		return
	}
	ob, b := old.Base(), w.Base()
	b.X, b.Y = ob.X, ob.Y
	if (b.MaxW != 0 || b.MaxH != 0) && (ob.MaxW != 0 || ob.MaxH != 0) {
		b.W, b.H = ob.W, ob.H
	}
	b.ApplyLimits()
	b.WinFlags |= ob.WinFlags & (WindowFlagHasBeenResized | WindowFlagMaximised)
	b.restoreW, b.restoreH = ob.restoreW, ob.restoreH
	b.stack, b.self = s, w
	ob.stack, ob.self = nil, nil
	s.Windows[i] = w
}

// Show opens a window, or — one of its type and number being open — is that
// window's open(): WindowManager::bringToFront(type, number), the open one
// replaced by the new struct where it stands. It reports a new window, which
// the caller has placed (Place, PlaceCentred or its own origin).
func (s *WindowStack) Show(w Win) bool {
	b := w.Base()
	if old := s.Find(b.WinType, b.Number); old != nil {
		s.Replace(old, w)
		s.BringToFront(w)
		return false
	}
	s.Open(w)
	return true
}

// Refresh replaces the open window of w's type and number where it stands
// and reports that there was one; the order is left alone.
func (s *WindowStack) Refresh(w Win) bool {
	b := w.Base()
	old := s.Find(b.WinType, b.Number)
	if old == nil {
		return false
	}
	s.Replace(old, w)
	return true
}

// Close is WindowManager::close(Window*): the window's onClose, then — if it
// is still there — off the list.
func (s *WindowStack) Close(w Win) {
	if w == nil || s.indexOf(w) < 0 {
		return
	}
	b := w.Base()
	if b.OnClose != nil {
		b.OnClose()
	}
	if i := s.indexOf(w); i >= 0 {
		s.Windows = append(s.Windows[:i], s.Windows[i+1:]...)
	}
	b.stack, b.self = nil, nil
}

// CloseType is WindowManager::close(type) (number AnyNumber) or close(type,
// id).
func (s *WindowStack) CloseType(typ WindowType, number int) {
	for {
		w := s.Find(typ, number)
		if w == nil {
			return
		}
		s.Close(w)
	}
}

// CloseAll is WindowManager::closeAllFloatingWindows, and the stickToFront
// windows of this stack with them (a new game, back to the title).
func (s *WindowStack) CloseAll() {
	for len(s.Windows) > 0 {
		n := len(s.Windows)
		s.Close(s.Windows[n-1])
		if len(s.Windows) == n { // never: Close removes what it is given
			s.Windows = s.Windows[:n-1]
		}
	}
	s.drag = dragNone
}

// rects are the windows as createWindow's placement sees them, in upstream's
// order: the stack's back to front, then the fixed ones (the toolbar and the
// panels, made when the game starts), then the stack's stickToFront windows —
// the news window, the ticker, a prompt —, which are made after those and
// so lie in front of them. The news thus blocks and anchors a new window as
// any other window does.
func (s *WindowStack) rects() []WinRect {
	var out, front []WinRect
	for _, w := range s.Windows {
		prepare(w)
		b := w.Base()
		r := WinRect{X: b.X, Y: b.Y, W: b.W, H: b.H, StickToBack: b.HasFlags(WindowFlagStickToBack)}
		if b.HasFlags(WindowFlagStickToFront) {
			front = append(front, r)
		} else {
			out = append(out, r)
		}
	}
	return append(append(out, s.Fixed...), front...)
}

// Place is WindowManager::createWindow(type, size, flags, events)'s
// placement: the window's origin set where upstream would open one of its
// size among the open windows (AutoWindowPosition).
func (s *WindowStack) Place(w Win) {
	b := w.Base()
	b.X, b.Y = AutoWindowPosition(b.W, b.H, s.ScreenW, s.ScreenH, s.rects())
}

// PlaceCentred is createWindowCentred's origin.
func (s *WindowStack) PlaceCentred(w Win) {
	b := w.Base()
	b.X, b.Y = s.ScreenW/2-b.W/2, max(28, s.ScreenH/2-b.H/2)
}

// MoveInsideScreenEdges is Window::moveInsideScreenEdges (a window's
// switchTab ends with it): a window past the screen's right or left edge,
// past height − 27 or above 28 comes back in.
func (s *WindowStack) MoveInsideScreenEdges(w Win) {
	prepare(w)
	b := w.Base()
	dx, dy := 0, 0
	if over := b.X + b.W - s.ScreenW; over > 0 {
		dx -= over
	}
	if b.X < 0 {
		dx -= b.X
	}
	if over := b.Y + b.H - (s.ScreenH - 27); over > 0 {
		dy -= over
	}
	if b.Y-28 < 0 {
		dy -= b.Y - 28
	}
	b.X += dx
	b.Y += dy
}

// MoveInsideScreenEdges is the stack's, for the window itself (nothing off
// a stack).
func (b *WinBase) MoveInsideScreenEdges() {
	if b.stack != nil && b.self != nil {
		b.stack.MoveInsideScreenEdges(b.self)
	}
}

// MoveOtherWindowsDown is WindowManager::moveOtherWindowsDown: the windows
// under self (not the sticking ones) move to 3 px below it, when it ends more
// than 80 px above the screen's bottom.
func (s *WindowStack) MoveOtherWindowsDown(self Win) {
	sb := self.Base()
	left, right, top, bottom := sb.X, sb.X+sb.W, sb.Y, sb.Y+sb.H
	for _, w := range s.Windows {
		b := w.Base()
		if w == self || b.HasFlags(WindowFlagStickToBack|WindowFlagStickToFront) {
			continue
		}
		if b.X >= right || b.X+b.W <= left || b.Y >= bottom || b.Y+b.H <= top {
			continue
		}
		if bottom < s.ScreenH-80 {
			b.Y += bottom + 3 - b.Y
		}
	}
}

// Resize is the screen's new size: WindowManager::relocateWindows — a window
// whose origin is within 10 px of the new right or bottom edge, or past it,
// moves to (8, 36), the next to (16, 44) …
func (s *WindowStack) Resize(sw, sh int) {
	if sw == s.ScreenW && sh == s.ScreenH {
		return
	}
	s.ScreenW, s.ScreenH = sw, sh
	loc := 8
	for _, w := range s.Windows {
		b := w.Base()
		extendsX, extendsY := b.X+10 >= sw, b.Y+10 >= sh
		if b.HasFlags(WindowFlagStickToBack | WindowFlagStickToFront) {
			extendsY = b.Y+10-27 >= sh // toolbars are 27px high
		}
		if extendsX || extendsY {
			b.X, b.Y = loc, loc+28
			loc += 8
		}
	}
}

// Dragging reports a window being moved or resized (Input::State::
// positioningWindow / resizing).
func (s *WindowStack) Dragging() bool { return s.drag != dragNone }

// Press is stateNormalLeft for a left press: the window under it comes to
// the front; on its caption the window starts moving, on a resizable
// window's grip resizing, else the window takes the press. It reports a
// window under the press.
func (s *WindowStack) Press(x, y int) bool {
	w := s.FindAt(x, y)
	if w == nil {
		return false
	}
	s.BringToFront(w)
	b := w.Base()
	begin := func(d stackDrag) {
		s.drag, s.dragType, s.dragNumber = d, b.WinType, b.Number
		s.lastX, s.lastY, s.pressX, s.pressY, s.moved = x, y, x, y, false
	}
	switch {
	case b.onCaption(x, y): // WidgetType::caption: windowPositionBegin
		begin(dragPositioning)
	case b.CanResize() && b.inGrip(x, y) && !gripBlocked(w, x, y): // panel / frame: windowResizeBegin
		begin(dragResizing)
		b.WinFlags &^= WindowFlagHasBeenResized
	default:
		w.HandleClick(x, y)
	}
	return true
}

func gripBlocked(w Win, x, y int) bool {
	g, ok := w.(GripBlocker)
	return ok && g.GripBlocked(x, y)
}

// Held is statePositioningWindow / stateResizing while the button is down
// (MouseButton::released: no button event).
func (s *WindowStack) Held(x, y int) {
	w := s.dragged()
	if w == nil {
		return
	}
	switch s.drag {
	case dragPositioning:
		s.position(w, x, y)
	case dragResizing:
		s.resize(w, x, y)
	}
}

// Release is the button coming up (leftReleased): the last move; a caption
// that never moved gets its click (callOnMouseUp: a rename); a resize grip
// that was only clicked maximises the window, or restores it.
func (s *WindowStack) Release(x, y int) {
	w := s.dragged()
	d := s.drag
	s.drag = dragNone
	if w == nil {
		return
	}
	b := w.Base()
	switch d {
	case dragPositioning:
		s.position(w, x, y)
		if !s.moved {
			w.HandleClick(s.pressX, s.pressY)
		}
	case dragResizing:
		switch {
		case b.HasFlags(WindowFlagHasBeenResized): // do not toggle maximised if the window has been resized
		case b.HasFlags(WindowFlagMaximised): // un-maximise: the previous size
			x, y = b.restoreW-b.W+s.lastX, b.restoreH-b.H+s.lastY
			b.WinFlags &^= WindowFlagMaximised
		default: // maximise, the previous size kept
			b.restoreW, b.restoreH = b.W, b.H
			x, y = s.lastX-b.X-b.W+s.ScreenW, s.lastY-b.Y-b.H+s.ScreenH-27
			b.WinFlags |= WindowFlagMaximised
			if y >= s.ScreenH-2 {
				s.lastX, s.lastY = x, y
				return
			}
			s.resizeBy(w, x-s.lastX, y-s.lastY)
			s.lastX, s.lastY = x, y
			return
		}
		s.resize(w, x, y)
	}
}

func (s *WindowStack) dragged() Win {
	if s.drag == dragNone {
		return nil
	}
	w := s.Find(s.dragType, s.dragNumber)
	if w == nil {
		s.drag = dragNone // state(State::reset)
	}
	return w
}

// position is statePositioningWindow's move: the pointer's y kept between 29
// and the screen's height − 29, the window moved by the pointer's.
func (s *WindowStack) position(w Win, x, y int) {
	y = min(max(y, 29), s.ScreenH-29)
	if dx, dy := x-s.lastX, y-s.lastY; dx != 0 || dy != 0 {
		b := w.Base()
		b.X += dx
		b.Y += dy
		s.moved = true
	}
	s.lastX, s.lastY = x, y
}

// resize is stateResizing's notMaximised branch.
func (s *WindowStack) resize(w Win, x, y int) {
	if y >= s.ScreenH-2 {
		s.lastX, s.lastY = x, y
		return
	}
	dx, dy := x-s.lastX, y-s.lastY
	s.lastX, s.lastY = x, y
	if dx == 0 && dy == 0 {
		return
	}
	w.Base().WinFlags &^= WindowFlagMaximised
	s.resizeBy(w, dx, dy)
}

func (s *WindowStack) resizeBy(w Win, dx, dy int) {
	b := w.Base()
	b.W = min(max(b.W+dx, b.MinW), b.MaxW)
	b.H = min(max(b.H+dy, b.MinH), b.MaxH)
	b.WinFlags |= WindowFlagHasBeenResized
	prepare(w) // callOnResize, callPrepareDraw
}

// Draw draws the windows back to front. A window that a later one overlaps
// is marked foreign for draw8's screen copy: its text and viewport are drawn
// past that copy, and the window over it must read them back.
func (s *WindowStack) Draw(screen *ebiten.Image, before func(Win)) {
	s.DrawWhere(screen, before, nil)
}

// DrawWhere is Draw for the windows keep says yes to (nil: all): the game
// draws the news over the bottom panels, which are not on this stack.
func (s *WindowStack) DrawWhere(screen *ebiten.Image, before func(Win), keep func(Win) bool) {
	for i, w := range s.Windows {
		if keep != nil && !keep(w) {
			continue
		}
		prepare(w)
		if before != nil {
			before(w)
		}
		w.Draw(screen)
		b := w.Base()
		r := image.Rect(b.X, b.Y, b.X+b.W, b.Y+b.H)
		for _, o := range s.Windows[i+1:] {
			ob := o.Base()
			if r.Overlaps(image.Rect(ob.X, ob.Y, ob.X+ob.W, ob.Y+ob.H)) {
				MarkForeign(r)
				break
			}
		}
	}
}
