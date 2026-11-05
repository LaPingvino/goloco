package world

import (
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// RailPlacementAt is Vehicle.cpp's getVehicleRailPlacementArgsFromCursor for
// the main view: the track element under screen point (sx, sy) (getMapCoordinatesFromPos
// filtered to track), then the placement on it nearest the cursor
// (sim.RailPlacementAtCursor), with gameToScreen + viewportToScreen of this
// view. ok is false where there is no track. PARTIAL: rotation 0 only (goloco's
// view does not rotate yet).
func (w *World) RailPlacementAt(head uint16, sx, sy int) (sim.VehiclePlacement, bool) {
	if w.scene == nil {
		return sim.VehiclePlacement{}, false
	}
	v := scene.View{X: int(w.camX), Y: int(w.camY), Zoom: w.zoom}
	ps := w.scene.InteractionAt(&w.Map, v, sx, sy, func(it paint.InteractionItem) bool { return it == paint.ItemTrack })
	if ps == nil {
		return sim.VehiclePlacement{}, false
	}
	te, ok := ps.Item.(*scenario.TrackElement)
	if !ok {
		return sim.VehiclePlacement{}, false
	}
	toScreen := func(x, y, z int32) (int, int) { // gameToScreen (rotation 0), then viewportToScreen
		vx, vy := int(y-x), int((y+x)>>1-z)
		return (vx - v.X) >> v.Zoom, (vy - v.Y) >> v.Zoom
	}
	return w.Map.RailPlacementAtCursor(head, int32(ps.MapX), int32(ps.MapY), te, 0, sx, sy, toScreen)
}
