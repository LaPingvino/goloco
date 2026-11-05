package world

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/scene"
)

// VehicleViewImage is the Vehicle window's Main viewport (Main::createViewport):
// zoom 0 (ZoomLevel::full), following the train's first car body — veh2 when it
// has no cars — width × height pixels; nil when the train is not placed (the
// viewport removed) or the renderer is missing. It has its own scene renderer,
// so the main view's kept frame is untouched. PARTIAL: rotation 0 only.
func (w *World) VehicleViewImage(headID uint16, width, height int) *ebiten.Image {
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
	rgba := w.vehicleScene.ToRGBA(w.vehicleScene.Render(m, v))
	if w.vehicleViewImg == nil || w.vehicleViewImg.Bounds().Dx() != width || w.vehicleViewImg.Bounds().Dy() != height {
		w.vehicleViewImg = ebiten.NewImage(width, height)
	}
	w.vehicleViewImg.WritePixels(rgba.Pix)
	return w.vehicleViewImg
}
