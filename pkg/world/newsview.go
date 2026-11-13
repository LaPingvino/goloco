package world

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/hajimehoshi/ebiten/v2"
)

// The news window's viewports (OpenLoco reference: src/OpenLoco/src/Ui/
// Windows/News/News.cpp getView): an industry or a town at its land height
// at zoom half, a station at its own height and a vehicle (its first car's
// body, else veh2) at zoom full. PARTIAL: rotation 0 only; no message type
// has a location for a subject.

// NewsView renders subject slot's (0, 1) viewport on the item of a message
// argument type; nil when there is none (the item gone, a vehicle not
// placed).
func (w *World) NewsView(slot int, argType uint8, id uint16, width, height int) *gfx8.RenderTarget {
	v, ok := w.newsView(argType, id, width, height)
	if !ok || slot < 0 || slot >= len(w.newsScene) {
		return nil
	}
	if w.newsScene[slot] == nil {
		w.newsScene[slot] = scene.New(w.renderer.G1, w.renderer.ObjMgr)
	}
	return w.newsScene[slot].Render(&w.Map, v)
}

// NewsViewLabels draws the viewport's station and town names on dst.
func (w *World) NewsViewLabels(argType uint8, id uint16, dst *ebiten.Image) {
	w.viewLabels(dst, true, func(width, height int) (scene.View, bool) { return w.newsView(argType, id, width, height) })
}

// HasNewsView reports that the item has a viewport (getView's view not empty).
func (w *World) HasNewsView(argType uint8, id uint16) bool {
	_, ok := w.newsView(argType, id, 1, 1)
	return ok
}

func (w *World) newsView(argType uint8, id uint16, width, height int) (scene.View, bool) {
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || width <= 0 || height <= 0 {
		return scene.View{}, false
	}
	m := &w.Map
	at := func(x, y, z, zoom int) (scene.View, bool) {
		vx, vy := y-x, (x+y)>>1-z // gameToScreen, rotation 0
		return scene.View{X: vx - (width<<zoom)/2, Y: vy - (height<<zoom)/2, W: width, H: height, Zoom: zoom, Ticks: m.ScenarioTicks}, true
	}
	land := func(x, y int) int { // getHeight(…).landHeight
		if tl := m.TileAt(x>>5, y>>5); tl != nil {
			return int(tl.BaseZ) * 4
		}
		return 0
	}
	switch argType {
	case sim.MsgArgIndustry:
		for i := range m.Industries {
			if in := &m.Industries[i]; uint16(in.ID) == id && in.Name != 0xFFFF {
				return at(int(in.X), int(in.Y), land(int(in.X), int(in.Y)), 1)
			}
		}
	case sim.MsgArgStation:
		if int(id) < len(m.StationNames) {
			s := &m.StationNames[id]
			return at(int(s.X), int(s.Y), int(s.Z), 0)
		}
	case sim.MsgArgTown:
		if t := m.TownByID(id); t != nil {
			return at(int(t.X), int(t.Y), land(int(t.X), int(t.Y)), 1)
		}
	case sim.MsgArgVehicle:
		return w.vehicleView(id, width, height)
	}
	return scene.View{}, false
}
