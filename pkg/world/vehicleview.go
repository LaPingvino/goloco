package world

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/hajimehoshi/ebiten/v2"
)

// VehicleViewImage is the Vehicle window's Main viewport (Main::createViewport):
// zoom 0 (ZoomLevel::full), following the train's first car body — veh2 when it
// has no cars — width × height pixels; nil when the train is not placed (the
// viewport removed) or the renderer is missing. It has its own scene renderer,
// so the main view's kept frame is untouched. It is handed back 8-bit, as
// upstream's viewport paints into the frame, so the window can draw over it
// (the centre-view button) before it reaches the screen.
func (w *World) VehicleView(headID uint16, width, height int) *gfx8.RenderTarget {
	v, ok := w.vehicleView(headID, width, height)
	if !ok {
		return nil
	}
	if w.vehicleScene == nil {
		w.vehicleScene = scene.New(w.renderer.G1, w.renderer.ObjMgr)
	}
	return w.vehicleScene.Render(&w.Map, v)
}

// VehicleViewLabels draws the Vehicle window viewport's station and town
// names on dst, the screen clipped to the viewport.
func (w *World) VehicleViewLabels(headID uint16, dst *ebiten.Image) {
	w.viewLabels(dst, func(width, height int) (scene.View, bool) { return w.vehicleView(headID, width, height) })
}

func (w *World) vehicleView(headID uint16, width, height int) (scene.View, bool) {
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || width <= 0 || height <= 0 {
		return scene.View{}, false
	}
	m := &w.Map
	tr, err := m.TrainOf(headID)
	if err != nil || m.Entities[tr.Head].TileX == -1 {
		return scene.View{}, false
	}
	target := tr.Veh2
	if len(tr.Cars) > 0 && len(tr.Cars[0]) > 0 {
		target = tr.Cars[0][0].Body
	}
	e := m.Entities[target]
	x, y, z := int(e.PosX), int(e.PosY), int(e.PosZ)
	vx, vy := w.gameToView(x, y, z)
	return scene.View{X: vx - width/2, Y: vy - height/2, W: width, H: height, Zoom: 0, Ticks: m.ScenarioTicks}, true
}

// CentreOnVehicle is Main::onCentreViewportControl: the main view moved to
// show the train's viewport target (its first car's body, else veh2). false
// when the train is not placed.
func (w *World) CentreOnVehicle(headID uint16) bool {
	m := &w.Map
	tr, err := m.TrainOf(headID)
	if err != nil || m.Entities[tr.Head].TileX == -1 {
		return false
	}
	target := tr.Veh2
	if len(tr.Cars) > 0 && len(tr.Cars[0]) > 0 {
		target = tr.Cars[0][0].Body
	}
	e := m.Entities[target]
	w.SetCamera(float64(e.PosX)/32, float64(e.PosY)/32)
	w.externalCamera = false
	return true
}

// FocusOnVehicle is the "focus main view on this" item: the main view follows
// the train (viewportFocusOnEntity) until the player moves the view.
func (w *World) FocusOnVehicle(headID uint16) {
	if w.CentreOnVehicle(headID) {
		w.followHead, w.following = headID, true
	}
}

// followVehicle keeps a focused view on its train, each frame; a train taken
// off the map ends the focus.
func (w *World) followVehicle() {
	if w.following && !w.CentreOnVehicle(w.followHead) {
		w.following = false
	}
}
