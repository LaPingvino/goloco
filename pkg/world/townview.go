package world

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

// TownView is the Town window's viewport (TownWindow.cpp initViewport): the
// town's centre at its land height, zoom level 2 (ZoomLevel::quarter), at
// the view rotation, width × height screen pixels. nil: no such town.
func (w *World) TownView(idx, width, height int) *gfx8.RenderTarget {
	return w.renderView(w.townView(idx, width, height))
}

func (w *World) townView(idx, width, height int) (scene.View, bool) {
	if idx < 0 || idx >= len(w.Map.Towns) {
		return scene.View{}, false
	}
	t := &w.Map.Towns[idx]
	return w.placeView(int(t.X), int(t.Y), width, height)
}

// TownViewLabels draws the Town window viewport's names (station and town:
// its viewport hides neither) on dst, the screen clipped to the viewport.
func (w *World) TownViewLabels(idx int, dst *ebiten.Image) {
	w.viewLabels(dst, func(width, height int) (scene.View, bool) { return w.townView(idx, width, height) })
}

// IndustryView is the Industry window's viewport (IndustryWindow.cpp
// initViewport): as the town's, on the industry's x, y. nil: no such industry.
func (w *World) IndustryView(idx, width, height int) *gfx8.RenderTarget {
	return w.renderView(w.industryView(idx, width, height))
}

func (w *World) industryView(idx, width, height int) (scene.View, bool) {
	if idx < 0 || idx >= len(w.Map.Industries) {
		return scene.View{}, false
	}
	in := &w.Map.Industries[idx]
	return w.placeView(int(in.X), int(in.Y), width, height)
}

// IndustryViewLabels draws the Industry window viewport's names (both kinds).
func (w *World) IndustryViewLabels(idx int, dst *ebiten.Image) {
	w.viewLabels(dst, func(width, height int) (scene.View, bool) { return w.industryView(idx, width, height) })
}

// renderView renders a window's view with the windows' own scene renderer
// (the main view's kept frame untouched); nil when there is no view.
func (w *World) renderView(v scene.View, ok bool) *gfx8.RenderTarget {
	if !ok {
		return nil
	}
	if w.townScene == nil {
		w.townScene = scene.New(w.renderer.G1, w.renderer.ObjMgr)
	}
	return w.townScene.Render(&w.Map, v)
}

// viewLabels draws the names of the window view that view gives for dst's
// size on dst (Viewport::paint's station / town names, by the view's flags),
// the view origin on the zoom's grid as the render's.
func (w *World) viewLabels(dst *ebiten.Image, view func(width, height int) (scene.View, bool)) {
	b := dst.Bounds()
	v, ok := view(b.Dx(), b.Dy())
	if !ok || w.TitleMode {
		return
	}
	z := max(0, min(3, v.Zoom))
	w.drawNames(dst, v.X>>z, v.Y>>z, z, v.Flags)
}

// placeView renders width × height screen pixels centred on x, y at the
// tile's land height (getHeight(…).landHeight), zoom quarter.
func (w *World) placeView(x, y, width, height int) (scene.View, bool) {
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || width <= 0 || height <= 0 {
		return scene.View{}, false
	}
	m := &w.Map
	z := 0
	if tl := m.TileAt(x>>5, y>>5); tl != nil {
		z = int(tl.BaseZ) * 4
	}
	const zoom = 2
	vx, vy := w.gameToView(x, y, z)
	return scene.View{X: vx - (width<<zoom)/2, Y: vy - (height<<zoom)/2, W: width, H: height, Zoom: zoom, Ticks: m.ScenarioTicks}, true
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
	w.SetViewFlags(paint.ViewUnderground, false) // viewportCentreOnTile: a place on the ground ends the underground view
	return true
}

// CentreOnIndustry is the Industry window's centre_on_viewport
// (viewportCentreMain): the main view moved to the industry.
func (w *World) CentreOnIndustry(idx int) bool {
	if idx < 0 || idx >= len(w.Map.Industries) {
		return false
	}
	in := &w.Map.Industries[idx]
	w.SetCamera(float64(in.X)/32, float64(in.Y)/32)
	w.externalCamera = false
	w.SetViewFlags(paint.ViewUnderground, false) // viewportCentreOnTile, as the town's
	return true
}

// StationView renders the Station window's viewport (Station::initViewport):
// centred on the station's x, y, z at zoom half (the station names hidden:
// StationViewLabels draws only the towns').
func (w *World) StationView(id, width, height int) *gfx8.RenderTarget {
	return w.renderView(w.stationView(id, width, height))
}

// StationViewLabels draws the Station window viewport's town names (its
// viewport hides the station names: hideStationNames).
func (w *World) StationViewLabels(id int, dst *ebiten.Image) {
	w.viewLabels(dst, func(width, height int) (scene.View, bool) { return w.stationView(id, width, height) })
}

func (w *World) stationView(id, width, height int) (scene.View, bool) {
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || width <= 0 || height <= 0 {
		return scene.View{}, false
	}
	m := &w.Map
	if id < 0 || id >= len(m.StationNames) {
		return scene.View{}, false
	}
	s := &m.StationNames[id]
	x, y, z := int(s.X), int(s.Y), int(s.Z)
	const zoom = 1
	vx, vy := w.gameToView(x, y, z)
	// "Remove station names from viewport" (StationWindow.cpp)
	return scene.View{X: vx - (width<<zoom)/2, Y: vy - (height<<zoom)/2, W: width, H: height, Zoom: zoom, Ticks: m.ScenarioTicks, Flags: paint.ViewHideStationNames}, true
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
	w.centreUnderground(int32(s.X), int32(s.Y), int32(s.Z)) // viewportCentreOnTile(station->x, y, z)
	return true
}
