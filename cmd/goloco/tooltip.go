package main

import (
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
)

// Window tooltips (Input/MouseInput.cpp's tooltip timing, ToolTip.cpp): a
// window's tooltip under the mouse opens once the mouse has rested 2 s (at
// once within 1 s of the last one closing), and closes after 8 s — staying
// closed until the mouse moves — or when the mouse moves. Times in updates
// (60 a second).
// PARTIAL: the toolbar's tooltips still show at once (pkg/ui toolbar.go).
const (
	tipDelay     = 120 // 2000 ms
	tipQuickOpen = 60  // tooltipNotShownTicks ≤ 1000: no delay
	tipLifetime  = 480 // 8000 ms
)

type tooltipState struct {
	text        string
	x, y        int
	rest, shown int
	showing     bool
	expired     bool // closed by its lifetime: not again until the mouse moves
	closedAt    int  // windowFrame when the last tooltip closed
	everClosed  bool
}

// windowTooltip is the tooltip the window under the mouse gives there.
func (g *Game) windowTooltip() string {
	if w := g.townList; w != nil && w.Tooltip != nil && w.Contains(g.mouseX, g.mouseY) {
		return w.Tooltip(g.mouseX, g.mouseY)
	}
	return ""
}

func (t *tooltipState) close(frame int) {
	if t.showing {
		t.closedAt, t.everClosed = frame, true
	}
	t.showing = false
}

// updateTooltip runs the timing each update.
func (g *Game) updateTooltip() {
	t := &g.tip
	text := g.windowTooltip()
	if text != t.text || g.mouseX != t.x || g.mouseY != t.y { // moved, or another tooltip
		t.close(g.windowFrame)
		t.text, t.x, t.y, t.rest, t.shown, t.expired = text, g.mouseX, g.mouseY, 0, 0, false
		t.showing = text != "" && t.everClosed && g.windowFrame-t.closedAt <= tipQuickOpen
		return
	}
	if t.text == "" || t.expired {
		return
	}
	if !t.showing {
		if t.rest++; t.rest >= tipDelay {
			t.showing = true
		}
		return
	}
	if t.shown++; t.shown >= tipLifetime {
		t.close(g.windowFrame)
		t.expired = true
	}
}

// drawTooltip draws the open window tooltip at the mouse.
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
