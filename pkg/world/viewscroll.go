package world

// The main view's glide: Window::moveWindowToLocation (Window.cpp l. 544)
// sets the main window's savedView and WindowFlags::scrollingToLocation;
// each render pass Window::viewportsUpdatePosition (l. 262, the
// scrollingToLocation branch at l. 328) moves the view an eighth of the way
// there, rounded up, until it arrives. Every "centre on" that goes through
// Window::viewportCentreOnTile glides this way: the map window's click and
// drag (moveMainViewToMapView), Window::viewportCentreMain (the town,
// industry, station and vehicle windows' centre buttons). What sets the
// view outright stays a jump: the rotation (viewportRotateRight / Left),
// viewportFromSavedView, the title's tour (Title.cpp l. 190 clears the flag),
// a focused view following its vehicle, and goloco's script and diagnostics
// cameras.
//
// OpenLoco reference: src/OpenLoco/src/Ui/Window.cpp.

// glide is the main window's scrollingToLocation and, while it is set, its
// savedViewX / savedViewY (the view's top-left to glide to, viewport
// coordinates). Unset, goloco's camera is the saved view itself.
type glide struct {
	on   bool
	x, y int
}

// moveWindowToLocation is Window::moveWindowToLocation (0x004C6801) on the
// main window: not while it follows an entity (viewportTargetSprite set); the
// saved view set to the position and scrollingToLocation on.
func (w *World) moveWindowToLocation(x, y int) {
	if w.following {
		return
	}
	w.glide = glide{on: true, x: x, y: y}
	w.externalCamera = false
}

// viewportCentreOnTile is Window::viewportCentreOnTile (0x004C6827) on the
// main window: the underground flag by the position's height
// (centreUnderground), then moveWindowToLocation to the position's
// gameToScreen less half the view (viewWidth / 2, viewHeight / 2).
func (w *World) viewportCentreOnTile(x, y, z int32) {
	w.centreUnderground(x, y, z)
	vx, vy := w.gameToView(int(x), int(y), int(z))
	sw, sh := w.viewSize()
	w.moveWindowToLocation(vx-(sw<<w.zoom)/2, vy-(sh<<w.zoom)/2)
}

// viewportCentreMain is Window::viewportCentreMain (l. 601) for a saved view
// on a position: the main view unfocused (Main::viewportFocusOnEntity with
// no entity), then viewportCentreOnTile.
func (w *World) viewportCentreMain(x, y, z int32) {
	w.following = false
	w.viewportCentreOnTile(x, y, z)
}

// glideStep is viewportsUpdatePosition's scrollingToLocation branch (l. 328)
// for one axis: the distance to the saved view, an eighth of it rounded up
// (its sign kept), and whether it was 0.
func glideStep(view, saved int) (int, bool) {
	d := saved - view
	flipped := d < 0
	if flipped {
		d = -d
	}
	d = (d + 7) / 8 // ceil(centreX / 8.0)
	if flipped {
		d = -d
	}
	return view + d, d == 0
}

// ViewportsUpdatePosition is Window::viewportsUpdatePosition (0x004C6456)
// for the main view, once a render pass (SoftwareDrawingEngine::render's
// WindowManager::updateViewports; at the default fixed update, one a tick):
// a view following an entity is followVehicle's; otherwise with
// scrollingToLocation the view steps towards the saved view, the flag
// clearing on the pass that finds it there (both steps 0).
func (w *World) ViewportsUpdatePosition() {
	if w.following || !w.glide.on {
		return
	}
	x, xDone := glideStep(int(w.camX), w.glide.x)
	y, yDone := glideStep(int(w.camY), w.glide.y)
	if xDone && yDone {
		w.glide.on = false
	}
	w.camX, w.camY = float64(x), float64(y) // viewportMove
}

// Gliding reports scrollingToLocation (for scripts and tests).
func (w *World) Gliding() bool { return w.glide.on }

// StopGlide clears scrollingToLocation: viewportDragBegin (MouseInput.cpp
// l. 1478) on the main view.
func (w *World) StopGlide() { w.glide.on = false }
