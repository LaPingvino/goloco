package main

import (
	"time"

	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// Window tooltips (Input/MouseInput.cpp's tooltip timing, ToolTip.cpp): a
// tooltip under the mouse opens once the mouse has rested 2 s (at once within
// 1 s of the last one closing), and closes after 8 s — staying closed until
// the mouse moves — or when the mouse moves. Real time (gameNow: virtual while a script plays), as upstream adds
// getTimeSinceLastTick.
const (
	tipDelay     = 2000 * time.Millisecond
	tipQuickOpen = 1000 * time.Millisecond // tooltipNotShownTicks ≤ 1000: no delay
	tipLifetime  = 8000 * time.Millisecond
)

type tooltipState struct {
	text     string
	x, y     int
	since    time.Time // the mouse at rest here since
	shownAt  time.Time
	showing  bool
	expired  bool      // closed by its lifetime: not again until the mouse moves
	closedAt time.Time // when the last tooltip closed (zero: never)
}

// windowTooltip is the tooltip the window under the mouse gives there (the
// toolbar's buttons have none upstream).
func (g *Game) windowTooltip() string {
	mx, my := g.mouseX, g.mouseY
	switch w := g.wins.FindAt(mx, my).(type) { // the front-most window under the mouse
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
		return w.Tooltip()
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

func (t *tooltipState) close(now time.Time) {
	if t.showing {
		t.closedAt = now
	}
	t.showing = false
}

// updateTooltip runs the timing each update.
func (g *Game) updateTooltip() {
	t, now := &g.tip, gameNow()
	text := g.windowTooltip()
	if text != t.text || g.mouseX != t.x || g.mouseY != t.y { // moved, or another tooltip
		t.close(now)
		t.text, t.x, t.y, t.since, t.expired = text, g.mouseX, g.mouseY, now, false
		if text != "" && !t.closedAt.IsZero() && now.Sub(t.closedAt) <= tipQuickOpen {
			t.showing, t.shownAt = true, now
		}
		return
	}
	switch {
	case t.text == "" || t.expired:
	case !t.showing:
		if now.Sub(t.since) >= tipDelay {
			t.showing, t.shownAt = true, now
		}
	case now.Sub(t.shownAt) >= tipLifetime:
		t.close(now)
		t.expired = true
	}
}

// drawTooltip draws the open tooltip at the mouse.
func (g *Game) drawTooltip(screen *ebiten.Image) {
	if !g.tip.showing || g.tip.text == "" {
		return
	}
	var c uint8
	if sk := g.skin(); sk != nil {
		c = sk.TooltipColour
	}
	ui.DrawTooltip(screen, g.tip.text, g.mouseX, g.mouseY, c)
}
