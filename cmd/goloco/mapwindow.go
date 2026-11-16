package main

import (
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// openMap is MapWindow::open (PORT-PLANS A, step 1a): the one map window,
// brought to the front when open; the skin's windowTitlebarColour and
// windowMapColour; the buffer cleared (clearMap) and, for now, filled whole
// at the view's rotation — upstream fills a row an update (step 1b).
func (g *Game) openMap() {
	if w := g.wins.Find(ui.WindowTypeMap, ui.AnyNumber); w != nil {
		g.wins.BringToFront(w)
		return
	}
	var tc, mc uint8 = 0, 1
	if sk := g.skin(); sk != nil {
		tc, mc = sk.WindowTitlebarColour, sk.WindowMapColour
	}
	buf := sim.NewMapBuffer()
	m, om, g1 := &g.w.Map, g.r.ObjMgr, g.r.G1
	pixel := func(img uint32, i int) uint8 {
		if g1 == nil || int(img) >= len(g1.Elements) || i < 0 || i >= len(g1.Elements[img].Data) {
			return 0
		}
		return g1.Elements[img].Data[i]
	}
	for range sim.MapRenderedWidth/2 + 1 {
		m.FillMapRowOverall(buf, int(m.CurrentRotation), om, pixel, 0)
	}
	g.showAuto(ui.NewMapWindow(buf, tc, mc), 0, 0)
}
