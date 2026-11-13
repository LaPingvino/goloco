package world

import (
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The view's rotation. Upstream has one for every viewport
// (WindowManager::getCurrentRotation); it is the map's CurrentRotation
// (pkg/sim viewrotation.go), which the scene renderer paints by.

// Rotation is WindowManager::getCurrentRotation: 0–3.
func (w *World) Rotation() uint8 { return w.Map.CurrentRotation & 3 }

// gameToView is World::gameToScreen at the current rotation: a world position
// in viewport coordinates.
func (w *World) gameToView(x, y, z int) (int, int) {
	vx, vy := w.Map.GameToScreen(int32(x), int32(y), int32(z))
	return int(vx), int(vy)
}

// viewSize is the main viewport's width and height on screen (800 × 600
// before the first draw).
func (w *World) viewSize() (int, int) {
	if w.viewW > 0 && w.viewH > 0 {
		return w.viewW, w.viewH
	}
	return 800, 600
}

// centreOn is Window::viewportCentreOnTile's move (Viewport::centre2dCoordinates):
// the main view with the world position at its centre.
func (w *World) centreOn(x, y, z int) {
	w.savedCentre = nil
	vx, vy := w.gameToView(x, y, z)
	sw, sh := w.viewSize()
	w.camX, w.camY = float64(vx-(sw<<w.zoom)/2), float64(vy-(sh<<w.zoom)/2)
}

// surfaceLocAt is ViewportInteraction's getSurfaceLocFromUi (water false) and
// the search of getSurfaceOrWaterLocFromUi (water true) for the main view:
// the surface (or water) under the screen point (getMapCoordinatesFromPos),
// then five passes of viewportCoordToMapCoord at the land height (the water's,
// on water) clamped to the hit tile.
func (w *World) surfaceLocAt(sx, sy int, water bool) (int32, int32, bool) {
	if w.scene == nil {
		return 0, 0, false
	}
	v := scene.View{X: int(w.camX), Y: int(w.camY), Zoom: w.zoom}
	ps := w.scene.InteractionAt(&w.Map, v, sx, sy, func(it paint.InteractionItem) bool {
		return it == paint.ItemSurface || water && it == paint.ItemWater
	})
	if ps == nil {
		return 0, 0, false
	}
	minX, minY := int32(ps.MapX), int32(ps.MapY)
	maxX, maxY := minX+31, minY+31
	var waterHeight int16
	if ps.Type == paint.ItemWater {
		_, waterHeight = w.Map.GetHeight(minX, minY)
	}
	mx, my := minX+16, minY+16
	vx, vy := int32(sx<<v.Zoom+v.X), int32(sy<<v.Zoom+v.Y) // screenToViewport
	for i := 0; i < 5; i++ {
		z := int32(waterHeight)
		if ps.Type != paint.ItemWater {
			land, _ := w.Map.GetHeight(mx, my)
			z = int32(land)
		}
		mx, my = w.Map.ViewportCoordToMapCoord(vx, vy, z)
		mx, my = min(max(mx, minX), maxX), min(max(my, minY), maxY)
	}
	return mx, my, true
}

// centreMapPosition is Viewport::getCentreMapPosition: the map position the
// view's centre shows, found by six passes of viewportCoordToMapCoord at the
// land height.
func (w *World) centreMapPosition() (int32, int32) {
	sw, sh := w.viewSize()
	cx, cy := int32(int(w.camX)+(sw<<w.zoom)/2), int32(int(w.camY)+(sh<<w.zoom)/2)
	x, y := w.Map.ViewportCoordToMapCoord(cx, cy, 0)
	for i := 0; i < 6; i++ {
		land, _ := w.Map.GetHeight(x, y)
		x, y = w.Map.ViewportCoordToMapCoord(cx, cy, int32(land))
	}
	return x, y
}

// RotateView is Window::viewportRotateRight / viewportRotateLeft on the main
// window with what its callers do after (TownManager / StationManager::
// updateLabels): the ground at the view's centre (getSurfaceLocFromUi, else
// getCentreMapPosition) stays at the centre while the rotation steps by one;
// everything is repainted, the windows' viewports follow (they share the
// rotation), and every entity's sprite bounds and every name's frame are
// recomputed (EntityManager::updateSpatialIndex, updateLabels).
func (w *World) RotateView(right bool) {
	sw, sh := w.viewSize()
	x, y, ok := w.surfaceLocAt(sw/2, sh/2, false)
	if !ok {
		x, y = w.centreMapPosition()
	}
	land, _ := w.Map.GetHeight(x, y)
	step := uint8(1)
	if !right {
		step = 3
	}
	w.setRotation((w.Rotation() + step) & 3)
	w.centreOn(int(x), int(y), int(land))
}

// SetRotation sets the rotation outright (Viewport::setRotation with the
// updates a rotation needs), the view's position kept: a saved view's, the
// diagnostics'.
func (w *World) SetRotation(rotation uint8) { w.setRotation(rotation & 3) }

func (w *World) setRotation(rotation uint8) {
	w.Map.CurrentRotation = rotation
	w.dirtyView, w.lastView = scene.View{}, scene.View{} // Window::invalidate: all of the view repainted
	w.Map.UpdateSpatialIndex()
	w.UpdateLabels()
}

// UpdateLabels is TownManager::updateLabels and StationManager::updateLabels
// (after a rotation, a load).
func (w *World) UpdateLabels() {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
	if g1, ok := w.renderer.ObjMgr.G1File.(sim.G1Raw); ok {
		w.Map.UpdateLabels(g1)
	}
}

// SetSavedView is Window::viewportFromSavedView for the main view: the saved
// centre (viewport coordinates), zoom and rotation of a save. A world not
// drawn yet doesn't know its view's size: the centre is kept and placed again
// by the first Draw.
func (w *World) SetSavedView(x, y int, zoom, rotation uint8) {
	w.zoom = int(min(zoom, 3))
	w.setRotation(rotation & 3)
	w.savedCentre = &[2]int{x, y}
	w.placeSavedCentre()
	if w.viewW > 0 {
		w.savedCentre = nil
	}
}

// placeSavedCentre puts the view's centre on the saved view's.
func (w *World) placeSavedCentre() {
	sw, sh := w.viewSize()
	w.camX, w.camY = float64(w.savedCentre[0]-(sw<<w.zoom)/2), float64(w.savedCentre[1]-(sh<<w.zoom)/2)
}
