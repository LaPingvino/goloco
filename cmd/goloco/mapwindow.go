package main

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// openMap is MapWindow::open (PORT-PLANS A, step 1a): the one map window,
// brought to the front when open; the skin's windowTitlebarColour and
// windowMapColour; the buffer cleared (clearMap), filled as it updates.
func (g *Game) openMap() {
	if w := g.wins.Find(ui.WindowTypeMap, ui.AnyNumber); w != nil {
		g.wins.BringToFront(w)
		return
	}
	var tc, mc uint8 = 0, 1
	if sk := g.skin(); sk != nil {
		tc, mc = sk.WindowTitlebarColour, sk.WindowMapColour
	}
	w := ui.NewMapWindow(sim.NewMapBuffer(), tc, mc) // clearMap
	w.Rotation = g.w.Map.CurrentRotation
	g.showAuto(w, 0, 0)
}

// mapChatMenu is TimePanel.cpp mapMouseDown (0x0043A67F) outside a network
// game: a one-item menu below the button, the map (its icon by the view's
// rotation); choosing it is mapDropdown's MapWindow::open.
func (g *Game) mapChatMenu(x, y, w, h int) {
	skin := 0
	if sk := g.skin(); sk != nil {
		skin = int(sk.ImageOffset)
	}
	items := []ui.DropdownItem{{MenuSprite: true, Sprite: skin + objects.ISToolbarMenuMapNorth + int(g.w.Map.CurrentRotation&3),
		Text: i18n.Label(i18n.MenuMap), Action: g.openMap}}
	d := ui.NewDropdownShowText(x, y, w, h, ui.DropdownColour, items, 25, g.sw, g.sh)
	d.OpenedByPress(image.Rect(x, y, x+w, y+h))
	d.Highlighted = 0
	g.dropdown = d
}

// updateMap is MapWindow's onUpdate (each frame): a new view rotation
// clears the buffer (var_846), then 80 rows are filled (setMapPixels).
func (g *Game) updateMap(w *ui.MapWindow) {
	m := &g.w.Map
	if m.CurrentRotation != w.Rotation {
		w.Rotation = m.CurrentRotation
		w.Buf.Clear()
	}
	g1 := g.r.G1
	pixel := func(img uint32, i int) uint8 {
		if g1 == nil || int(img) >= len(g1.Elements) || i < 0 || i >= len(g1.Elements[img].Data) {
			return 0
		}
		return g1.Elements[img].Data[i]
	}
	for range 80 {
		m.FillMapRowOverall(w.Buf, int(w.Rotation), g.r.ObjMgr, pixel, 0)
	}
}
