package world

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/scene"
)

// VehicleViewImage is the Vehicle window's Main viewport (Main::createViewport):
// zoom 0 (ZoomLevel::full), following the train's first car body — veh2 when it
// has no cars — width × height pixels; nil when the train is not placed (the
// viewport removed) or the renderer is missing. It has its own scene renderer,
// so the main view's kept frame is untouched. It is handed back 8-bit, as
// upstream's viewport paints into the frame, so the window can draw over it
// (the centre-view button) before it reaches the screen. PARTIAL: rotation 0
// only.
func (w *World) VehicleView(headID uint16, width, height int) *gfx8.RenderTarget {
	if w.renderer == nil || w.renderer.G1 == nil || w.renderer.ObjMgr == nil || width <= 0 || height <= 0 {
		return nil
	}
	m := &w.Map
	tr, err := m.TrainOf(headID)
	if err != nil || m.Entities[tr.Head].TileX == -1 {
		return nil
	}
	target := tr.Veh2
	if len(tr.Cars) > 0 && len(tr.Cars[0]) > 0 {
		target = tr.Cars[0][0].Body
	}
	e := m.Entities[target]
	x, y, z := int(e.PosX), int(e.PosY), int(e.PosZ)
	vx, vy := y-x, (x+y)>>1-z // gameToScreen, rotation 0
	if w.vehicleScene == nil {
		w.vehicleScene = scene.New(w.renderer.G1, w.renderer.ObjMgr)
	}
	v := scene.View{X: vx - width/2, Y: vy - height/2, W: width, H: height, Zoom: 0, Ticks: m.ScenarioTicks}
	return w.vehicleScene.Render(m, v)
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
