package world

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/scene"
)

// TownView is the Town window's viewport (TownWindow.cpp initViewport): the
// town's centre at its land height, zoom level 2 (ZoomLevel::quarter),
// rotation 0, width × height screen pixels. nil: no such town.
func (w *World) TownView(idx, width, height int) *gfx8.RenderTarget {
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || width <= 0 || height <= 0 {
		return nil
	}
	m := &w.Map
	if idx < 0 || idx >= len(m.Towns) {
		return nil
	}
	t := &m.Towns[idx]
	x, y := int(t.X), int(t.Y)
	z := 0
	if tl := m.TileAt(x>>5, y>>5); tl != nil {
		z = int(tl.BaseZ) * 4 // getHeight(…).landHeight
	}
	const zoom = 2
	vx, vy := y-x, (x+y)>>1-z // gameToScreen, rotation 0
	if w.townScene == nil {
		w.townScene = scene.New(w.renderer.G1, w.renderer.ObjMgr)
	}
	v := scene.View{X: vx - (width<<zoom)/2, Y: vy - (height<<zoom)/2, W: width, H: height, Zoom: zoom, Ticks: m.ScenarioTicks}
	return w.townScene.Render(m, v)
}

// CentreOnTown is the Town window's centre_on_viewport (viewportCentreMain):
// the main view moved to the town.
func (w *World) CentreOnTown(idx int) bool {
	m := &w.Map
	if idx < 0 || idx >= len(m.Towns) {
		return false
	}
	t := &m.Towns[idx]
	w.SetCamera(float64(t.X)/32, float64(t.Y)/32)
	w.externalCamera = false
	return true
}

// StationView renders the Station window's viewport (Station::initViewport):
// centred on the station's x, y, z at zoom half (the station names hidden —
// the scene draws none).
func (w *World) StationView(id, width, height int) *gfx8.RenderTarget {
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || width <= 0 || height <= 0 {
		return nil
	}
	m := &w.Map
	if id < 0 || id >= len(m.StationNames) {
		return nil
	}
	s := &m.StationNames[id]
	x, y, z := int(s.X), int(s.Y), int(s.Z)
	const zoom = 1
	vx, vy := y-x, (x+y)>>1-z // gameToScreen, rotation 0
	if w.townScene == nil {
		w.townScene = scene.New(w.renderer.G1, w.renderer.ObjMgr)
	}
	v := scene.View{X: vx - (width<<zoom)/2, Y: vy - (height<<zoom)/2, W: width, H: height, Zoom: zoom, Ticks: m.ScenarioTicks}
	return w.townScene.Render(m, v)
}

// CentreOnStation is the Station window's centre_on_viewport
// (viewportCentreMain): the main view moves to the station.
func (w *World) CentreOnStation(id int) bool {
	m := &w.Map
	if id < 0 || id >= len(m.StationNames) {
		return false
	}
	s := &m.StationNames[id]
	w.SetCamera(float64(s.X)/32, float64(s.Y)/32)
	w.externalCamera = false
	return true
}
