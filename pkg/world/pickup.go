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

// SurfaceOrWaterAt is ViewportInteraction's getSurfaceOrWaterLocFromUi
// (main view, rotation 0): the surface or water under the screen point
// (getMapCoordinatesFromPos filtered to both), then five passes of
// viewportCoordToMapCoord at the land height (or the water's, on water),
// clamped to the hit tile, and the tile corner of the result.
func (w *World) SurfaceOrWaterAt(sx, sy int) (int32, int32, bool) {
	if w.scene == nil {
		return 0, 0, false
	}
	v := scene.View{X: int(w.camX), Y: int(w.camY), Zoom: w.zoom}
	ps := w.scene.InteractionAt(&w.Map, v, sx, sy, func(it paint.InteractionItem) bool {
		return it == paint.ItemSurface || it == paint.ItemWater
	})
	if ps == nil {
		return 0, 0, false
	}
	minX, minY := int32(ps.MapX), int32(ps.MapY)
	maxX, maxY := minX+31, minY+31
	var water int16
	if ps.Type == paint.ItemWater {
		_, water = w.Map.GetHeight(minX, minY)
	}
	mx, my := minX+16, minY+16
	vx, vy := int32(sx<<v.Zoom+v.X), int32(sy<<v.Zoom+v.Y) // screenToViewport
	for i := 0; i < 5; i++ {
		z := int32(water)
		if ps.Type != paint.ItemWater {
			land, _ := w.Map.GetHeight(mx, my)
			z = int32(land)
		}
		mx, my = vy-vx>>1+z, vy+vx>>1+z // viewportCoordToMapCoord, rotation 0
		mx, my = min(max(mx, minX), maxX), min(max(my, minY), maxY)
	}
	return mx &^ 0x1F, my &^ 0x1F, true
}

// WaterPlacementAt is getVehicleWaterPlacementArgsFromCursor (0x00427595):
// around the surface or water under the cursor, the 8×8 tiles from 4 tiles
// up-left searched for docks (sequence 0) — the one whose boat position (the
// dock object's, rotated, + 32, 32 from its tile) is nearest the cursor's
// tile centre by Manhattan distance; its tile and base height.
func (w *World) WaterPlacementAt(head uint16, sx, sy int) (sim.VehicleWaterPlacement, bool) {
	px, py, ok := w.SurfaceOrWaterAt(sx, sy)
	if !ok || w.renderer == nil || w.renderer.ObjMgr == nil {
		return sim.VehicleWaterPlacement{}, false
	}
	docks := w.renderer.ObjMgr.DockObjects
	cx, cy := px+16, py+16
	best, found := int32(1<<31-1), false
	var out sim.VehicleWaterPlacement
	for i := int32(0); i < 8; i++ {
		for j := int32(0); j < 8; j++ {
			lx, ly := px-4*32+i*32, py-4*32+j*32
			t := w.Map.TileAt(int(lx>>5), int(ly>>5))
			if lx < 0 || ly < 0 || t == nil {
				continue
			}
			for _, el := range t.Stations {
				if el.StationType != 3 || el.SequenceIndex != 0 || int(el.ObjectID) >= len(docks) || docks[el.ObjectID] == nil {
					continue
				}
				d := docks[el.ObjectID]
				ox, oy := sim.RotateOffset(d.BoatPosition[0], d.BoatPosition[1], el.Rotation)
				bx, by := lx+32+int32(ox), ly+32+int32(oy)
				if dist := abs32(bx-cx) + abs32(by-cy); dist < best {
					best, found = dist, true
					out = sim.VehicleWaterPlacement{X: lx, Y: ly, Z: int32(el.BaseZ) * 4, Head: head}
				}
			}
		}
	}
	return out, found
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

// stationFlag5 / stationFlagAirport are StationFlags::flag_5 / flag_6.
const (
	stationFlag5       = 1 << 5
	stationFlagAirport = 1 << 6
)

// AirPlacementAt is getVehicleAirPlacementArgsFromCursor (main view,
// rotation 0): the airport under the cursor — or, over the ground or water,
// an airport station whose start is within 5 tiles of it (upstream's loop
// kept: the last station element it looked at stands) —, then of its
// terminal nodes (last first) the one nearest the cursor on screen by
// Manhattan distance.
func (w *World) AirPlacementAt(head uint16, sx, sy int) (sim.VehicleAirPlacement, bool) {
	if w.scene == nil || w.renderer == nil || w.renderer.ObjMgr == nil {
		return sim.VehicleAirPlacement{}, false
	}
	om := w.renderer.ObjMgr
	v := scene.View{X: int(w.camX), Y: int(w.camY), Zoom: w.zoom}
	var el *scenario.StationElement
	if ps := w.scene.InteractionAt(&w.Map, v, sx, sy, func(it paint.InteractionItem) bool { return it == paint.ItemAirport }); ps != nil {
		el, _ = ps.Item.(*scenario.StationElement)
	} else {
		ps := w.scene.InteractionAt(&w.Map, v, sx, sy, func(it paint.InteractionItem) bool {
			return it == paint.ItemSurface || it == paint.ItemWater
		})
		if ps == nil {
			return sim.VehicleAirPlacement{}, false
		}
		found := false
		for id := range w.Map.StationNames {
			st := &w.Map.StationNames[id]
			if st.Flags&stationFlagAirport == 0 || st.Flags&stationFlag5 != 0 {
				continue
			}
			if abs32(int32(ps.MapX)-int32(st.AirportStartPos[0])) > 5*32 || abs32(int32(ps.MapY)-int32(st.AirportStartPos[1])) > 5*32 {
				continue
			}
			t := w.Map.TileAt(int(st.AirportStartPos[0])>>5, int(st.AirportStartPos[1])>>5)
			if t == nil {
				continue
			}
			for _, r := range t.Order { // el.as<StationElement>(): nil for the other kinds
				el = nil
				if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) {
					continue
				}
				el = &t.Stations[r.Index]
				if el.AiAllocated || el.Ghost || el.StationType != 2 {
					continue
				}
				found = true
				break
			}
		}
		if !found {
			return sim.VehicleAirPlacement{}, false
		}
	}
	if el == nil || el.AiAllocated || el.Ghost || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil {
		return sim.VehicleAirPlacement{}, false
	}
	out := sim.VehicleAirPlacement{StationID: el.StationID, Head: head}
	nodes := om.AirportObjects[el.ObjectID].MovementNodes
	best := int32(1<<31 - 1)
	for n := len(nodes) - 1; n >= 0; n-- {
		if nodes[n].Flags&objects.AirportNodeTerminal == 0 {
			continue
		}
		x, y, z, ok := w.Map.AirportMovementNodeLoc(om, out.StationID, uint8(n))
		if !ok {
			continue
		}
		gx, gy := paint.GameToScreen(paint.Pos3{X: int(x), Y: int(y), Z: int(z)}, 0)
		ux, uy := (gx-v.X)>>v.Zoom, (gy-v.Y)>>v.Zoom // viewportToScreen
		if d := abs32(int32(ux-sx)) + abs32(int32(uy-sy)); d < best {
			best, out.AirportNode = d, uint8(n)
		}
	}
	if best == 1<<31-1 {
		return sim.VehicleAirPlacement{}, false
	}
	return out, true
}
