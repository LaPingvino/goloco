package main

import (
	"time"

	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
	"github.com/hajimehoshi/ebiten/v2"
)

// maptooltip.go — the map tooltip (Ui/Windows/MapToolTip.cpp) and what
// processMouseOver (Input/MouseInput.cpp 0x004CD47A) gives it: over the main
// view without a tool, getItemLeft's item (a vehicle, town, station,
// industry or headquarters: MapTipLeft); otherwise, with the map selection
// on, its size ("{INT16} x {INT16}", map_tooltip_int_x_int).
// MapToolTip::open, once a tick (dispatchUpdateAll): the text shows after
// the mouse has rested 25 ticks (within 5 px; a right press restarts the
// count) — the selection's size at once — and not with ctrl or shift held,
// the right button down or an error window open.
// Then rightOver's text (pkg/world/rightover.go): the track or road to
// modify, the signal, station, tree, wall … to remove, a building's cargo.
// PARTIAL: viewports other than the main view's, a tool's cursor hiding the
// item (onToolCursor's out), the hovered station's highlight.

// mapTipTick is a tick of upstream's 40 Hz loop (Engine::UpdateRateInMs).
const mapTipTick = 25 * time.Millisecond

type mapTipState struct {
	timeout    uint16 // _mapTooltipTimeout
	x, y       int    // _tooltipLocation
	lastTick   time.Time
	tip        world.MapTip // FormatArguments::mapToolTip and _mapTooltipOwner (Text "": StringIds::null)
	open       bool         // the window open
	winX, winY int
}

// toolActive is Input::hasFlag(toolActive): some window's tool on the map.
func (g *Game) toolActive() bool {
	return g.tf.on || g.industryToolOn() || g.pickupOn || (g.routeTool && g.vehicleWin(g.routeHead) != nil) || g.consToolOn()
}

// updateMapTip runs processMouseOver's map tooltip part and, a tick at a
// time, MapToolTip::open.
func (g *Game) updateMapTip() {
	t := &g.mapTip
	now := gameNow()
	if t.lastTick.IsZero() || now.Sub(t.lastTick) > time.Second {
		t.lastTick = now
	}
	ticks := 0
	for now.Sub(t.lastTick) >= mapTipTick {
		t.lastTick = t.lastTick.Add(mapTipTick)
		ticks++
	}
	for ; ticks > 0; ticks-- {
		g.mapTipMouseOver()
		g.mapTipOpen()
	}
}

// mapTipMouseOver is processMouseOver's part for the tooltip: reset, then
// the item under the mouse, or the selection's size and rightOver's text.
func (g *Game) mapTipMouseOver() {
	t := &g.mapTip
	t.tip = world.MapTip{Owner: -1} // MapToolTip::reset
	if g.inTitleScreen {
		return
	}
	mx, my := g.mouseX, g.mouseY
	if !g.overWindow(mx, my) && !g.toolActive() { // the viewport widget, no tool
		if it, ok := g.w.ItemLeftAt(mx, my); ok {
			switch it.Kind {
			case paint.ItemEntity, paint.ItemTownLabel, paint.ItemStationLabel, paint.ItemIndustry, paint.ItemHeadquarterBuilding:
				if tip, ok := g.w.MapTipLeft(it); ok {
					t.tip = tip
				}
				return // skipItem
			}
		}
	}
	if on, _, a := g.w.MapSelection(); on {
		l, d := abs(a[0]-a[2])+1, abs(a[1]-a[3])+1
		t.tip.Text = itoa(l) + " x " + itoa(d)
	}
	// ViewportInteraction::rightOver: what a right click would act on gives
	// its text (over a window there is no viewport under the mouse).
	if !g.overWindow(mx, my) && g.r.ObjMgr != nil {
		if _, tip, hasTip, ok := g.w.RightOver(mx, my, g.rightOverState()); ok && hasTip {
			t.tip = tip
		}
	}
}

// mapTipOpen is MapToolTip::open.
func (g *Game) mapTipOpen() {
	t := &g.mapTip
	t.timeout++
	mx, my := g.mouseX, g.mouseY
	right := g.inMousePressed(ebiten.MouseButtonRight)
	if on, _, _ := g.w.MapSelection(); !on {
		if abs(t.x-mx) > 5 || abs(t.y-my) > 5 || right {
			t.timeout = 0
		}
		t.x, t.y = mx, my
		if t.timeout < 25 || t.tip.Text == "" || right || ebiten.IsKeyPressed(ebiten.KeyControl) ||
			ebiten.IsKeyPressed(ebiten.KeyShift) || g.errorWin != nil {
			t.open = false
			return
		}
	}
	t.open = true
	t.winX, t.winY = ui.MapTipPos(mx, my, g.sw, g.sh)
}

// drawMapTip is the tooltip window's draw (stickToFront: over the windows).
func (g *Game) drawMapTip(screen *ebiten.Image) {
	t := &g.mapTip
	if !t.open || g.inTitleScreen {
		return
	}
	tip := ui.MapTip{Text: t.tip.Text}
	if sk := g.skin(); sk != nil {
		tip.ObjectColour, tip.CargoColour = sk.MapTooltipObjectColour, sk.MapTooltipCargoColour
	}
	if ctrl, _ := g.w.ControllingCompany(); t.tip.Owner >= 0 && t.tip.Owner != ctrl {
		tip.Owned = true
		tip.Face, tip.FaceColour = g.companyFace(t.tip.Owner)
	}
	ui.DrawMapTip(screen, tip, t.winX, t.winY)
}
