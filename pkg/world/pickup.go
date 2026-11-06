package world

import (
	"github.com/LaPingvino/goloco/pkg/objects"
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

// RoadPlacementAt is RailPlacementAt for road vehicles —
// getVehicleRoadPlacementArgsFromCursor: the road element under the cursor
// (getMapCoordinatesFromPos filtered to road and tram), its placement nearest
// the cursor. PARTIAL: rotation 0 only.
func (w *World) RoadPlacementAt(head uint16, sx, sy int) (sim.VehiclePlacement, bool) {
	if w.scene == nil {
		return sim.VehiclePlacement{}, false
	}
	v := scene.View{X: int(w.camX), Y: int(w.camY), Zoom: w.zoom}
	ps := w.scene.InteractionAt(&w.Map, v, sx, sy, func(it paint.InteractionItem) bool { return it == paint.ItemRoad })
	if ps == nil {
		return sim.VehiclePlacement{}, false
	}
	re, ok := ps.Item.(*scenario.RoadElement)
	if !ok {
		return sim.VehiclePlacement{}, false
	}
	toScreen := func(x, y, z int32) (int, int) { // gameToScreen (rotation 0), then viewportToScreen
		vx, vy := int(y-x), int((y+x)>>1-z)
		return (vx - v.X) >> v.Zoom, (vy - v.Y) >> v.Zoom
	}
	return w.Map.RoadPlacementAtCursor(head, int32(ps.MapX), int32(ps.MapY), re, 0, sx, sy, toScreen)
}

// PlacementAt is pickupToolUpdate / pickupToolDown's choice by the vehicle's
// transport mode: rail or road (air and water: not yet — false).
func (w *World) PlacementAt(head uint16, sx, sy int) (sim.VehiclePlacement, bool) {
	tr, err := w.Map.TrainOf(head)
	if err != nil {
		return sim.VehiclePlacement{}, false
	}
	switch objects.TransportMode(w.Map.Entities[tr.Head].Mode) {
	case objects.TransportModeRail:
		return w.RailPlacementAt(head, sx, sy)
	case objects.TransportModeRoad:
		return w.RoadPlacementAt(head, sx, sy)
	}
	return sim.VehiclePlacement{}, false
}
