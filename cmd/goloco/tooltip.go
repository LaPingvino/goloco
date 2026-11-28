package main

import (
	"time"

	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// Window tooltips: Input/MouseInput.cpp's tooltip part of stateNormalHover
// (0x004C8098), the left press's close (stateNormalLeft) and the owner set
// when the input state goes back to normal after a widget press
// (stateWidgetPressed's leftReleased, dropdownRegisterSelection,
// windowPositionEnd), with ToolTip.cpp's open (common: the box placed from
// the cursor at opening, and kept there) and closeAndReset.
//
// The "owner" is ToolTip's windowType / windowNumber / widgetIndex: the
// widget whose tooltip shows, or the widget last pressed. While the mouse
// stays on the owner nothing opens — after a click on a widget its tooltip
// doesn't come back until the mouse has been on another widget — and an
// open tooltip stays put (moving inside its widget doesn't close it) until
// 8 s are up. Another widget under the mouse: closeAndReset, and the usual
// timing starts again (2 s at rest; at once when a tooltip closed within the
// last second).
//
// Times in ms of the frame clock (gameNow: virtual while a script plays), as
// upstream adds getTimeSinceLastTick.
const (
	tipDelay     = 2000 // bp: the rest before a tooltip opens
	tipQuickOpen = 1000 // tooltipNotShownTicks ≤ 1000: no delay
	tipMoveGrace = 500  // tooltipNotShownTicks < 500: the timeout runs while the mouse moves
	tipLifetime  = 8000 // an open tooltip closes after this (its owner stays)
)

// tipKey is a widget's identity: ToolTip's window type/number + widgetIndex
// (by type and number, as upstream: goloco may rebuild a window's value).
// Windows that say which widget is under the mouse (ui.TooltipWidgets) give
// its index; for the others the tooltip text stands in for the widget
// (two neighbouring widgets with the same text count as one) — PARTIAL.
type tipKey struct {
	untyped ui.Win // a window without a WindowType: itself
	winType ui.WindowType
	number  int
	widget  int
	text    string
}

type tooltipState struct {
	owner   tipKey
	ownerOn bool // windowType != undefined

	timeout   int    // _tooltipTimeout
	notShown  uint16 // _tooltipNotShownTicks: uint16 as upstream (it wraps after ~65 s)
	mouseX    int    // _tooltipCursor
	mouseY    int
	last      time.Time
	started   bool
	pressKey  tipKey // the pressed widget (_pressedWindowType / _pressedWidgetIndex)
	pressOn   bool   // a widget was pressed: its release sets the owner
	pressHeld bool   // the left button went down in State::normal and is held

	showing bool   // the tooltip window is open
	text    string // its text
	x, y    int    // the cursor it opened at (common's cursorX/Y)
}

// tipAt is the widget under the mouse and its tooltip (callTooltip: "" when
// the widget has none).
func (g *Game) tipAt() (tipKey, string) {
	mx, my := g.mouseX, g.mouseY
	w := g.wins.FindAt(mx, my) // the front-most window under the mouse
	if w == nil {
		return tipKey{widget: -1}, ""
	}
	text := g.windowTooltip(w, mx, my)
	b := w.Base()
	var untyped ui.Win
	if b.WinType == 0 {
		untyped = w
	}
	if b.WinType == ui.WindowTypeTimePanel {
		return tipKey{winType: b.WinType, number: b.Number, widget: uilayout.TimePanelWidgetAt(mx, my, g.sw, g.sh)}, text
	}
	if tw, ok := w.(ui.TooltipWidgets); ok {
		return tipKey{untyped: untyped, winType: b.WinType, number: b.Number, widget: tw.TooltipWidgetAt(mx, my)}, text
	}
	return tipKey{untyped: untyped, winType: b.WinType, number: b.Number, widget: -1, text: text}, text
}

// windowTooltip is the tooltip the window gives at the mouse (the toolbar's
// buttons have none upstream).
func (g *Game) windowTooltip(w ui.Win, mx, my int) string {
	switch w := w.(type) {
	case *ui.BarWindow:
		if w.WinType == ui.WindowTypeTimePanel {
			return g.timePanelTooltip(uilayout.TimePanelWidgetAt(mx, my, g.sw, g.sh))
		}
	case *ui.ViewWindow:
		return w.WidgetTooltip(mx, my)
	case *ui.StationWindow:
		return w.WidgetTooltip(mx, my)
	case *ui.TownListWindow:
		if w.Tooltip != nil {
			if tip := w.Tooltip(mx, my); tip != "" {
				return tip
			}
		}
		return w.WidgetTooltip(mx, my)
	case *ui.VehicleWindow:
		return w.WidgetTooltip(mx, my)
	case *ui.MusicJukeboxWindow:
		return w.WidgetTooltip(mx, my)
	case *ui.NewsWindow:
		return w.WidgetTooltip(mx, my)
	case *ui.MessagesWindow:
		return w.WidgetTooltip(mx, my)
	case *ui.TerraformWindow:
		return w.WidgetTooltip(mx, my)
	case *ui.MusicSelectionWindow:
		return w.WidgetTooltip(mx, my)
	case *ui.KeyboardShortcutsWindow:
		return w.WidgetTooltip(mx, my)
	}
	return ""
}

// close is WindowManager::close(WindowType::tooltip): the owner stays.
func (t *tooltipState) close() { t.showing = false }

// closeAndReset is ToolTip::closeAndReset (0x004C87B5).
func (t *tooltipState) closeAndReset() {
	t.showing = false
	t.timeout = 0
	t.ownerOn = false
}

// updateTooltip runs each frame, before the frame's press is handled (the
// key under the mouse is then still the widget pressed).
func (g *Game) updateTooltip() {
	now := gameNow()
	t := &g.tip
	dt := 0
	if t.started {
		dt = int(now.Sub(t.last) / time.Millisecond)
	} else {
		t.notShown = 0xFFFF // initMouse
	}
	t.last, t.started = now, true
	key, text := g.tipAt()
	dropdown := g.dropdown != nil && g.dropdown.Visible || g.colourDropdown != nil && g.colourDropdown.Visible
	press := g.inMouseJustPressed(ebiten.MouseButtonLeft)
	held := g.inMousePressed(ebiten.MouseButtonLeft)
	release := g.inMouseJustReleased(ebiten.MouseButtonLeft)
	t.step(dt, key, text, g.mouseX, g.mouseY, tipInput{
		press: press, held: held, release: release,
		dropdown: dropdown, rightDrag: g.rightDragging(),
		resetPress: g.tipPressResets(),
	})
	// processMouseOver (after handleMouse): the window's cursor event —
	// TimePanel's onCursor sets the timeout to 2000 over the date, so its
	// tooltip opens as soon as the mouse rests there (and, open, never times
	// out).
	if key.winType == ui.WindowTypeTimePanel && key.widget == uilayout.TimePanelWidxDate {
		t.timeout = tipDelay // Ui::ToolTip::setTooltipTimeout(2000)
	}
}

// tipPressResets says a left press here ends in State::reset rather than in
// the widget pressed state: no window (the main view: viewportLeft), a
// window's viewport widget (viewportLeft) or scroll view (scrollLeft) — their
// releases leave no owner.
func (g *Game) tipPressResets() bool {
	x, y := g.mouseX, g.mouseY
	w := g.wins.FindAt(x, y)
	if w == nil || ui.ScrollAreaAt(w, x, y) >= 0 {
		return true
	}
	if vw, ok := w.(ui.ViewportWidgets); ok && vw.ViewportWidgetAt(x, y) {
		return true
	}
	return false
}

type tipInput struct {
	press, held, release bool // the left button this frame
	dropdown             bool // a dropdown is up (State::dropdownActive)
	rightDrag            bool // viewportRight / scrollRight
	resetPress           bool // the press is on a viewport, scroll view or no window
}

// step is one frame of the tooltip's input handling (dt ms since the last).
func (t *tooltipState) step(dt int, key tipKey, text string, x, y int, in tipInput) {
	// WindowManager::tick; the open tooltip's update (0x004C94FF) keeps it 0.
	t.notShown += uint16(dt)
	if t.showing {
		t.notShown = 0
	}
	if in.press && !in.dropdown { // stateNormalLeft: close(tooltip); the press is on this widget
		t.close()
		t.pressKey, t.pressOn, t.pressHeld = key, !in.resetPress, true
	} else if in.press { // a press in a dropdown: its selection's owner is the widget that opened it
		t.close()
	}
	if in.release && t.pressHeld && !in.dropdown {
		// back to State::normal: after a widget press the owner is the pressed
		// widget (timeout 0); after a viewport / scroll press State::reset
		// clears it and takes the cursor.
		t.pressHeld = false
		t.timeout = 0
		if t.pressOn {
			t.owner, t.ownerOn = t.pressKey, true
		} else {
			t.ownerOn = false
			t.mouseX, t.mouseY = x, y
		}
		return
	}
	if in.release && in.dropdown && t.pressHeld { // the press opened a dropdown: dropdownActive
		t.pressHeld = false
	}
	if in.held || in.dropdown || in.rightDrag { // not State::normal: no hover timing
		if in.dropdown && !in.held && t.pressOn {
			// dropdownRegisterSelection follows when the dropdown goes: the
			// opener becomes the owner then.
			t.owner, t.ownerOn, t.timeout = t.pressKey, true, 0
		}
		return
	}
	if t.ownerOn { // stateNormalHover with a tooltip owner
		if key == t.owner {
			t.timeout += dt
			if t.timeout >= tipLifetime {
				t.close()
			}
		} else {
			t.closeAndReset()
		}
		return
	}
	if t.notShown < tipMoveGrace || (x == t.mouseX && y == t.mouseY) {
		t.timeout += dt
		bp := tipDelay
		if t.notShown <= tipQuickOpen {
			bp = 0
		}
		if bp > t.timeout {
			return
		}
		if text != "" { // ToolTip::open: a widget with a tooltip
			t.owner, t.ownerOn = key, true
			t.showing, t.text, t.x, t.y = true, text, x, y
			t.notShown = 0
		}
	}
	t.timeout = 0
	t.mouseX, t.mouseY = x, y
}

// drawTooltip draws the open tooltip where it opened.
func (g *Game) drawTooltip(screen *ebiten.Image) {
	if !g.tip.showing || g.tip.text == "" {
		return
	}
	var c uint8
	if sk := g.skin(); sk != nil {
		c = sk.TooltipColour
	}
	ui.DrawTooltip(screen, g.tip.text, g.tip.x, g.tip.y, c)
}
