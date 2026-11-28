package ui

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"
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
// The toolbar and the bottom panels are windows of the stack too, as
// upstream's stickToFront ToolbarTop, CompanyInfoPanel and TimePanel (Gui::
// init; cmd/goloco opens them, BarWindow): the list is drawn, hit-tested and
// walked by the placement in one order, upstream's.
//
// A press on a window's bare panel or frame moves it (upstream:
// windowPositionBegin there too) for the windows that tell their bare panel
// apart (FrameHitter: their widget lists, framehit.go and the Vehicle and
// Build Vehicle layouts). PARTIAL: the others move by the caption alone —
// the Construction window (SimpleWindow, drawn by cmd/goloco), the Town /
// Industry windows on this branch, a list on a tab goloco doesn't open
// (the editor's town build tabs).

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
	WindowTypeTerraform
	WindowTypeKeyboardShortcuts
	WindowTypeEditKeyboardShortcut
	WindowTypeMap
	// WindowType::topToolbar, companyInfoPanel, timePanel: the bars.
	WindowTypeTopToolbar
	WindowTypeCompanyInfoPanel
	WindowTypeTimePanel
	// WindowType::textInput: the rename box (TextInputWindow).
	WindowTypeTextInput
)

// WindowTypeNames are the types by their upstream WindowType names (a
// script's "expect window:NAME").
var WindowTypeNames = map[string]WindowType{
	"construction": WindowTypeConstruction, "town": WindowTypeTown, "industry": WindowTypeIndustry,
	"station": WindowTypeStation, "townList": WindowTypeTownList, "industryList": WindowTypeIndustryList,
	"stationList": WindowTypeStationList, "vehicleList": WindowTypeVehicleList, "vehicle": WindowTypeVehicle,
	"buildVehicle": WindowTypeBuildVehicle, "company": WindowTypeCompany, "companyList": WindowTypeCompanyList,
	"options": WindowTypeOptions, "scenarioSelect": WindowTypeScenarioSelect,
	"fileBrowserPrompt": WindowTypeFileBrowserPrompt, "musicJukebox": WindowTypeMusicJukebox,
	"musicSelection": WindowTypeMusicSelection, "messages": WindowTypeMessages, "news": WindowTypeNews,
	"terraform": WindowTypeTerraform, "keyboardShortcuts": WindowTypeKeyboardShortcuts,
	"editKeyboardShortcut": WindowTypeEditKeyboardShortcut, "map": WindowTypeMap,
	"topToolbar": WindowTypeTopToolbar, "companyInfoPanel": WindowTypeCompanyInfoPanel, "timePanel": WindowTypeTimePanel,
	"textInput": WindowTypeTextInput,
}

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
	// WindowFlagNoBackground (the toolbar, the panels): findAt sees the window
	// only where one of its widgets is (WidgetHitter).
	WindowFlagNoBackground WindowFlags = 1 << 8
	// WindowFlagPlaySoundOnOpen, WindowFlagOpenQuietly: createWindow flashes
	// the new window's border and plays openWindow for a window that sticks
	// neither to the back nor the front unless it opens quietly, and for any
	// window with playSoundOnOpen.
	WindowFlagPlaySoundOnOpen WindowFlags = 1 << 9
	WindowFlagOpenQuietly     WindowFlags = 1 << 10
)

// Flags the stack keeps itself (MouseInput's stateResizing).
const (
	WindowFlagHasBeenResized WindowFlags = 1 << (16 + iota)
	WindowFlagMaximised
)

// upstreamFlagBits are Ui::WindowFlags' bits (Window.h) for goloco's flags,
// for what keeps a window's flags in a save (LastMapWindowAttributes).
var upstreamFlagBits = []struct {
	f   WindowFlags
	bit uint32
}{
	{WindowFlagStickToBack, 1 << 0}, {WindowFlagStickToFront, 1 << 1}, {WindowFlagNoBackground, 1 << 5},
	{WindowFlagIgnoreInFindAt, 1 << 7}, {WindowFlagResizable, 1 << 9}, {WindowFlagNoAutoClose, 1 << 10},
	{WindowFlagPlaySoundOnOpen, 1 << 12}, {WindowFlagOpenQuietly, 1 << 13},
	{WindowFlagHasBeenResized, 1 << 15}, {WindowFlagMaximised, 1 << 16},
}

// UpstreamFlags is the window's flags in Ui::WindowFlags' bits, the white
// border's count as whiteBorderOne's multiple (whiteBorderMask, 1 << 17 …).
func (b *WinBase) UpstreamFlags() uint32 {
	var u uint32
	for _, m := range upstreamFlagBits {
		if b.WinFlags&m.f != 0 {
			u |= m.bit
		}
	}
	return u | uint32(b.whiteBorder&3)<<17
}

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
	// whiteBorder is WindowFlags::whiteBorderMask's count (3 at the flash's
	// start, one less each WindowManager::tick).
	whiteBorder int
	// pressed: the widget a HandleClick named (Button, PressWidget), in
	// screen coordinates.
	pressed    image.Rectangle
	hasPressed bool
	// pressedHit: where a release still counts as over the pressed widget,
	// when more than its own rectangle (a combobox's button: the combobox
	// too, findWidgetAt); empty: pressed alone.
	pressedHit image.Rectangle
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

// WidgetHitter reports a widget at a screen point: what findAt asks a
// noBackground window (findWidgetAt != kWidgetIndexNull).
type WidgetHitter interface{ HasWidgetAt(x, y int) bool }

// FrameHitter reports that the widget at a screen point is the window's
// frame or panel (findWidgetAt: no other widget there): stateNormalLeft then
// begins moving the window (windowPositionBegin), or resizing it in the grip.
// A window without it moves by its caption alone.
type FrameHitter interface{ OnFrame(x, y int) bool }

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
	// Modal is WindowManager's _currentModalType (0: none): a press on any
	// other window brings the modal one to the front instead.
	Modal WindowType

	// sound is createWindow's openWindow sound, panned to the new window's
	// middle, for the game to play (TakeOpenSound).
	sound    bool
	soundPan int
	// quiet: a window rebuilt in place (a tab switch, Quietly) — no flash
	// and no sound, as upstream's switchTab changes the one window.
	quiet bool

	// The drag: _dragWindowType / _dragWindowNumber, _dragLast, _windowMoved,
	// and where the press was (the caption's click on an unmoved release).
	drag             stackDrag
	dragType         WindowType
	dragNumber       int
	lastX, lastY     int
	pressX, pressY   int
	moved            bool
	captionClickable bool

	// clicks: the click sounds due (TakeClickSounds); pressedBtn: the
	// pressed widget, in screen coordinates, until the release
	// (_pressedWindowType / _pressedWidgetIndex, State::widgetPressed).
	clicks        []ClickSound
	pressedBtn    image.Rectangle
	pressedHit    image.Rectangle // pressedBtn and what counts as it (WinBase.pressedHit)
	hasPressedBtn bool
	// dropdownBtn: a dropdown button whose press opens a dropdown (set by
	// dropdownPress), pressed while that dropdown is up (SetDropdownActive).
	dropdownBtn stackDropdownBtn
}

// stackDropdownBtn is _pressedWindowType / _pressedWidgetIndex kept through
// State::dropdownActive: the window and its button (window coordinates).
type stackDropdownBtn struct {
	owner Win
	r     [4]int
	state uint8
}

const (
	dropdownNone    = iota
	dropdownPending // pressed; the dropdown is to open
	dropdownActive  // the dropdown is up
)

// SetDropdownActive tells the stack whether a dropdown is up, before the
// windows are drawn: a dropdown button pressed since is the dropdown's
// (Input::State::dropdownActive with its pressed widget) until none is up.
func (s *WindowStack) SetDropdownActive(open bool) {
	switch {
	case !open:
		s.dropdownBtn = stackDropdownBtn{}
	case s.dropdownBtn.state == dropdownPending:
		s.dropdownBtn.state = dropdownActive
	}
}

// dropdownPending makes the window's widget r (window coordinates) the
// pressed widget of the dropdown its press opens (dropdownPress's).
func (b *WinBase) dropdownPending(r [4]int) {
	if b.stack != nil {
		b.stack.dropdownBtn = stackDropdownBtn{owner: b.self, r: r, state: dropdownPending}
	}
}

// dropdownPressed reports the window's dropdown button r as the pressed
// widget of the dropdown that is up (Window::draw's pressedWidget).
func (b *WinBase) dropdownPressed(r [4]int) bool {
	if b.stack == nil {
		return false
	}
	d := b.stack.dropdownBtn
	return d.state == dropdownActive && d.owner == b.self && d.r == r
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

// FindAt is WindowManager::findAt: the front-most window at a screen point
// (not an ignoreInFindAt one; a noBackground one only on a widget).
func (s *WindowStack) FindAt(x, y int) Win { return s.findAt(x, y, false) }

// FindAtAlt is WindowManager::findAtAlt: findAt that sees the windows
// ignoreInFindAt hides from it.
func (s *WindowStack) FindAtAlt(x, y int) Win { return s.findAt(x, y, true) }

func (s *WindowStack) findAt(x, y int, alt bool) Win {
	for i := len(s.Windows) - 1; i >= 0; i-- {
		w := s.Windows[i]
		prepare(w) // callOnResize
		b := w.Base()
		if !b.Contains(x, y) || !alt && b.HasFlags(WindowFlagIgnoreInFindAt) {
			continue
		}
		if b.HasFlags(WindowFlagNoBackground) {
			if h, ok := w.(WidgetHitter); !ok || !h.HasWidgetAt(x, y) {
				continue
			}
		}
		return w
	}
	return nil
}

// BringToFront is WindowManager::bringToFront(Window&): the window changes
// places with the front-most one that does not stick to the back — a
// toolbar or panel, the news window, a stickToFront prompt — unless that is
// the list's first (upstream swaps the two: it does not move the others
// down), and a window almost off the screen's left comes back to x 20. So a
// window pressed comes in front of the panels and the toolbar, and the panel
// it changed places with lies where the window was.
func (s *WindowStack) BringToFront(w Win) {
	b := w.Base()
	if b.HasFlags(WindowFlagStickToBack | WindowFlagStickToFront) {
		return
	}
	front := -1
	for i := len(s.Windows) - 1; i >= 0; i-- {
		if !s.Windows[i].Base().HasFlags(WindowFlagStickToBack) {
			front = i
			break
		}
	}
	if i := s.indexOf(w); i >= 0 && front > 0 && front != i {
		s.Windows[i], s.Windows[front] = s.Windows[front], s.Windows[i]
	}
	if b.X+b.W < 20 {
		b.X = 20
	}
}

// BringToFrontType is WindowManager::bringToFront(type, number): the window
// flashes its border (whiteBorderMask) and comes to the front; nil: none open.
func (s *WindowStack) BringToFrontType(typ WindowType, number int) Win {
	w := s.Find(typ, number)
	if w == nil {
		return nil
	}
	w.Base().whiteBorder = 3
	s.BringToFront(w)
	return w
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
// the stickToFront ones (behind the others if it sticks to the back). A
// window that sticks neither way, or has playSoundOnOpen, flashes its border
// and sounds openWindow at its middle (TakeOpenSound) unless it opens quietly.
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
	if !s.quiet && (b.HasFlags(WindowFlagPlaySoundOnOpen) || !b.HasFlags(WindowFlagStickToBack|WindowFlagStickToFront|WindowFlagOpenQuietly)) {
		b.whiteBorder = 3
		s.sound, s.soundPan = true, b.X+b.W/2
	}
	s.Windows = append(s.Windows, nil)
	copy(s.Windows[dst+1:], s.Windows[dst:])
	s.Windows[dst] = w
	s.KeepOnScreen(w)
}

// RepanOpenSound is for a window that createWindow made at another width
// than the one it opened with (its tab's onResize came after): the sound
// due is panned to x, the middle of the size it was created at.
func (s *WindowStack) RepanOpenSound(x int) {
	if s.sound {
		s.soundPan = x
	}
}

// Quietly runs f, a window's tab switch that goloco does by rebuilding the
// window (its opener on another tab): Show and Open within it neither flash
// nor sound — upstream's switchTab changes the window it has.
func (s *WindowStack) Quietly(f func()) {
	q := s.quiet
	s.quiet = true
	defer func() { s.quiet = q }()
	f()
}

// TakeOpenSound reports an openWindow sound due since the last call and
// where to pan it (createWindow's origin.x + size.width / 2).
func (s *WindowStack) TakeOpenSound() (pan int, ok bool) {
	pan, ok = s.soundPan, s.sound
	s.sound = false
	return
}

// Tick is WindowManager::tick's border flash: each flashing window's count
// one less.
func (s *WindowStack) Tick() {
	for _, w := range s.Windows {
		if b := w.Base(); b.whiteBorder > 0 {
			b.whiteBorder--
		}
	}
}

// WhiteBorder reports the window's border flashing (whiteBorderMask set).
func (b *WinBase) WhiteBorder() bool { return b.whiteBorder > 0 }

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
	b.whiteBorder = ob.whiteBorder
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
		if s.quiet {
			s.BringToFront(w)
		} else {
			s.BringToFrontType(b.WinType, b.Number)
		}
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
	s.Modal = 0
}

// rects are the windows as createWindow's placement sees them: the list,
// back to front — the toolbar, the panels and the news window among the
// others where creation and bringToFront's swaps have put them.
func (s *WindowStack) rects() []WinRect {
	var out []WinRect
	for _, w := range s.Windows {
		prepare(w)
		b := w.Base()
		out = append(out, WinRect{X: b.X, Y: b.Y, W: b.W, H: b.H, StickToBack: b.HasFlags(WindowFlagStickToBack)})
	}
	return out
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
	for _, w := range s.Windows {
		s.KeepOnScreen(w)
	}
}

// KeepOnScreen is goloco's small-screen deviation (N57; upstream has no
// such step): a window (not the bars or the news) is brought wholly onto the
// screen, within moveInsideScreenEdges' bounds (Window.cpp: x in [0,
// width − w], y in [28, height − 27 − h]) — the toolbar above and the bottom
// panels (the time panel's pause, the company's) stay uncovered. One larger than
// the screen that may be resized is first shrunk to fit (down to its minimum
// size, its onResize laying it out again); one that still does not fit keeps
// its top left — title bar and caption — on screen, the rest reached by
// dragging it by its panel. Open and Resize call it: upstream's windows open
// where createWindow puts them and only relocateWindows' origin test moves
// them on a resize, so on a phone (or a turned or shrunk browser viewport)
// a window could open or end with its buttons past the edge, out of reach.
func (s *WindowStack) KeepOnScreen(w Win) {
	b := w.Base()
	if b.HasFlags(WindowFlagStickToBack) || b.HasFlags(WindowFlagStickToFront) && b.NoCaption || s.ScreenW <= 0 || s.ScreenH <= 0 {
		return // the bars, the news and its ticker: placed by their own
	}
	prepare(w)
	if b.HasFlags(WindowFlagResizable) {
		nw, nh := b.W, b.H
		if nw > s.ScreenW {
			nw = max(s.ScreenW, b.MinW)
		}
		if nh > s.ScreenH-28-27 {
			nh = max(s.ScreenH-28-27, b.MinH)
		}
		if nw != b.W || nh != b.H {
			b.W, b.H = nw, nh
			b.WinFlags |= WindowFlagHasBeenResized
			prepare(w) // callOnResize, callPrepareDraw
		}
	}
	b.X = min(max(b.X, 0), max(0, s.ScreenW-b.W))
	b.Y = min(max(b.Y, 28), max(28, s.ScreenH-27-b.H))
}

// Dragging reports a window being moved or resized (Input::State::
// positioningWindow / resizing).
func (s *WindowStack) Dragging() bool { return s.drag != dragNone }

// Resizing is Input::State::resizing.
func (s *WindowStack) Resizing() bool { return s.drag == dragResizing }

// ResizeGripAt is processMouseOver's diagonalArrows test (MouseInput.cpp
// 0x004CD47A): over a resizable window whose size isn't fixed both ways, on
// its panel or frame (the widget there), in the 19 px corner — where a press
// begins resizing (Press).
func (s *WindowStack) ResizeGripAt(x, y int) bool {
	w := s.FindAt(x, y)
	if w == nil {
		return false
	}
	b := w.Base()
	frame, knows := onFrame(w, x, y)
	return b.CanResize() && b.inGrip(x, y) && (frame || !knows && !gripBlocked(w, x, y))
}

// Press is stateNormalLeft for a left press: the window under it comes to
// the front; on its caption the window starts moving, on a resizable
// window's grip resizing, on its bare frame or panel (FrameHitter) moving,
// else the window takes the press. It reports a window under the press.
//
// With a modal window open (Modal), a press on any other window brings the
// modal one to the front instead (MouseInput's handleMouse: the press goes
// no further; the error sound is the caller's): see PressModal.
func (s *WindowStack) Press(x, y int) bool {
	s.hasPressedBtn = false // a new press: the last one's widget is done (a modal may have taken its release)
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
	frame, knows := onFrame(w, x, y)
	bare := frame || !knows && !gripBlocked(w, x, y) // without FrameHitter, the grip's corner by GripBlocker
	switch {
	case b.onCaption(x, y): // WidgetType::caption: windowPositionBegin
		begin(dragPositioning)
	case b.CanResize() && b.inGrip(x, y) && bare: // panel / frame: windowResizeBegin
		begin(dragResizing)
		b.WinFlags &^= WindowFlagHasBeenResized
	case frame: // panel / frame elsewhere: windowPositionBegin
		begin(dragPositioning)
	default:
		s.pressHandled(w, x, y)
	}
	return true
}

// PressModal is handleMouse's modal check for a left press at a screen
// point: with a modal window open, a press anywhere but on it brings it to
// the front (bringToFront(modalType): its border flashes) and is taken —
// the caller sounds the error. It reports that the press went to it.
func (s *WindowStack) PressModal(x, y int) bool {
	if s.Modal != 0 && s.Find(s.Modal, AnyNumber) == nil {
		s.Modal = 0 // the modal window has gone (setCurrentModalType(undefined))
	}
	if s.Modal == 0 {
		return false
	}
	if w := s.FindAt(x, y); w != nil && w.Base().WinType == s.Modal {
		return false
	}
	s.BringToFrontType(s.Modal, AnyNumber)
	return true
}

// onFrame is FrameHitter's answer, and whether the window gives one.
func onFrame(w Win, x, y int) (frame, knows bool) {
	f, ok := w.(FrameHitter)
	return ok && f.OnFrame(x, y), ok
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
			// MouseInput statePositioningWindow: a window that never moved
			// sounds clickPress at the pressed caption / frame / panel's
			// midX — each spans the window: (1 + w − 2) / 2 = (w − 1) / 2.
			s.clicks = append(s.clicks, ClickSound{ClickPress, b.X + MidX(0, b.W)})
		}
		if !s.moved && b.onCaption(s.pressX, s.pressY) {
			w.HandleClick(s.pressX, s.pressY)
			b.hasPressed = false // the caption's own click: the clickPress above
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

// Draw draws the windows back to front (WindowManager::render: the list's
// order). A window that a later one overlaps is marked foreign for draw8's
// screen copy: its text and viewport are drawn past that copy, and the window
// over it must read them back. A flashing window gets its white border
// (Window::draw: fillRectInset(0, 0, width − 1, height − 1, white, fillNone)).
func (s *WindowStack) Draw(screen *ebiten.Image, before func(Win)) {
	for i, w := range s.Windows {
		prepare(w)
		if before != nil {
			before(w)
		}
		w.Draw(screen)
		b := w.Base()
		if b.whiteBorder > 0 {
			drawWhiteBorder(screen, b.X, b.Y, b.W, b.H)
		}
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

// drawWhiteBorder is the flash: an outset border in Colour::white's shades.
func drawWhiteBorder(screen *ebiten.Image, x, y, w, h int) {
	if globalG1Font == nil {
		return
	}
	m, err := globalG1Font.g1.GetPaletteMap(colourWhite)
	if err != nil {
		return
	}
	// The four edges only, each a draw8 of its own: the window's inside
	// (text drawn past draw8's screen copy) is left as it is. The edges are
	// all overwritten; their old pixels are read back (foreign).
	MarkForeign(image.Rect(x, y, x+w, y+h))
	for _, e := range [4][4]int{{0, 0, w, 1}, {0, h - 1, w, 1}, {0, 0, 1, h}, {w - 1, 0, 1, h}} {
		ox, oy := e[0], e[1]
		draw8(screen, x+ox, y+oy, e[2], e[3], func(rt *gfx8.RenderTarget) {
			gfx8.FillRectInset(rt, -ox, -oy, w-1-ox, h-1-oy, gfx8.InsetColour{Map: m}, gfx8.InsetFillNone)
		})
	}
}
