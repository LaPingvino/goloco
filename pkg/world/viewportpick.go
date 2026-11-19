package world

import (
	"github.com/LaPingvino/goloco/pkg/scene"
)

// viewportpick.go — the viewports the hit tests read (Ui::Viewport as
// ViewportInteraction's getMapCoordinatesFromPos, getItemLeft and rightOver
// take it from WindowManager::findAt's window): the main view's, and the
// Town, Industry, Station and Vehicle windows' (their viewport widgets).
// OpenLoco reference: src/OpenLoco/src/Ui/ViewportInteraction.cpp,
// src/OpenLoco/include/OpenLoco/Viewport.hpp (containsWindowPos,
// windowToViewport, viewportToWindow)

// Viewport is one viewport of a window: its top-left on the screen (the
// window's position + vp->x, y) and the paint view it draws (viewX, viewY,
// its width and height on screen, zoom, flags).
type Viewport struct {
	X, Y int
	View scene.View
}

// Contains is containsWindowPos for screen point (sx, sy).
func (v Viewport) Contains(sx, sy int) bool {
	return sx >= v.X && sx < v.X+v.View.W && sy >= v.Y && sy < v.Y+v.View.H
}

// ToScreen is viewportToWindow + the window's position for a world position:
// where the viewport shows it on the screen.
func (v Viewport) ToScreen(w *World, x, y, z int32) (int, int) {
	vx, vy := w.Map.GameToScreen(x, y, z)
	return v.X + (int(vx)-v.View.X)>>v.View.Zoom, v.Y + (int(vy)-v.View.Y)>>v.View.Zoom
}

// MainViewport is the main window's viewport: the whole screen.
func (w *World) MainViewport() Viewport {
	v := w.mainView()
	v.W, v.H = w.viewSize()
	return Viewport{View: v}
}

// mainViewports is the main window's viewport list (no window over the point).
func (w *World) mainViewports() []Viewport { return []Viewport{w.MainViewport()} }

// windowViewport is a window's viewport at screen (x, y): the view the
// window draws there for that size (the same call its draw makes).
func windowViewport(x, y int, v scene.View, ok bool) (Viewport, bool) {
	if !ok {
		return Viewport{}, false
	}
	return Viewport{X: x, Y: y, View: v}, true
}

// TownViewport is the Town window's viewport at screen (x, y), width × height.
func (w *World) TownViewport(idx, x, y, width, height int) (Viewport, bool) {
	v, ok := w.townView(idx, width, height)
	return windowViewport(x, y, v, ok)
}

// IndustryViewport is the Industry window's viewport.
func (w *World) IndustryViewport(idx, x, y, width, height int) (Viewport, bool) {
	v, ok := w.industryView(idx, width, height)
	return windowViewport(x, y, v, ok)
}

// StationViewport is the Station window's viewport.
func (w *World) StationViewport(id, x, y, width, height int) (Viewport, bool) {
	v, ok := w.stationView(id, width, height)
	return windowViewport(x, y, v, ok)
}

// VehicleViewport is the Vehicle window's viewport (none while the train is
// not placed: the viewport removed).
func (w *World) VehicleViewport(headID uint16, x, y, width, height int) (Viewport, bool) {
	v, ok := w.vehicleView(headID, width, height)
	return windowViewport(x, y, v, ok)
}

// containingViewport is the first of vps that holds screen point (sx, sy)
// (the loop over w->viewports with containsWindowPos); -1 none.
func containingViewport(vps []Viewport, sx, sy int) int {
	for i, vp := range vps {
		if vp.Contains(sx, sy) {
			return i
		}
	}
	return -1
}
