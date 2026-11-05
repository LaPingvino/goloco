package world

import (
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/scene"
)

// RouteStopAt is Vehicle.cpp's Route::getRouteInteractionFromCursor for the
// stop-at orders: the station a click at screen point (sx, sy) would add to a
// route — through a train station (its track the controlling company's), a
// road station, an airport or dock (its element the company's), or a track or
// road piece with a station on it — when the station itself is the
// company's (stationLabelAdjustedInteraction). AI-allocated and ghost elements
// don't count. ok is false where a click adds no stop.
// PARTIAL: route-through waypoints on plain track / road / water (the
// OrderRouteWaypoint orders), the station label as a target, the water-mode
// flags, and the tooltip / hovered-station highlight are not ported yet.
func (w *World) RouteStopAt(sx, sy int) (station int, ok bool) {
	if w.scene == nil && w.renderer != nil && w.renderer.G1 != nil && w.renderer.ObjMgr != nil {
		w.scene = scene.New(w.renderer.G1, w.renderer.ObjMgr) // as drawScene's first frame makes it
	}
	if w.scene == nil {
		return 0, false
	}
	m := &w.Map
	v := scene.View{X: int(w.camX), Y: int(w.camY), Zoom: w.zoom}
	ps := w.scene.InteractionAt(m, v, sx, sy, func(it paint.InteractionItem) bool {
		switch it {
		case paint.ItemTrack, paint.ItemRoad, paint.ItemTrainStation, paint.ItemRoadStation, paint.ItemAirport, paint.ItemDock:
			return true
		}
		return false
	})
	if ps == nil {
		return 0, false
	}
	t := m.TileAt(int(ps.MapX)/32, int(ps.MapY)/32)
	if t == nil {
		return 0, false
	}
	me := m.ControllingCompany
	stationOn := func(z uint8) *scenario.StationElement { // the station element at a track / road piece's height
		for i := range t.Stations {
			if t.Stations[i].BaseZ == z {
				return &t.Stations[i]
			}
		}
		return nil
	}
	var el *scenario.StationElement
	switch ps.Type {
	case paint.ItemTrack: // trackAdjustedInteraction
		tk, _ := ps.Item.(*scenario.TrackElement)
		if tk == nil || tk.AiAllocated || tk.Ghost || tk.Owner != me || !tk.HasStation {
			return 0, false
		}
		el = stationOn(tk.BaseZ)
	case paint.ItemRoad: // roadAdjustedInteraction: any owner's road
		rd, _ := ps.Item.(*scenario.RoadElement)
		if rd == nil || rd.AiAllocated || rd.Ghost || !rd.HasStationElement {
			return 0, false
		}
		el = stationOn(rd.BaseZ)
	case paint.ItemTrainStation: // trainStationAdjustedInteraction: the track under it the company's
		el, _ = ps.Item.(*scenario.StationElement)
		if el == nil {
			return 0, false
		}
		var trackMine bool
		for _, tk := range t.Tracks {
			if tk.BaseZ == el.BaseZ && tk.HasStation {
				trackMine = tk.Owner == me
			}
		}
		if !trackMine {
			return 0, false
		}
	case paint.ItemAirport, paint.ItemDock: // dockAirportAdjustedInteraction
		el, _ = ps.Item.(*scenario.StationElement)
		if el == nil || el.Owner != me {
			return 0, false
		}
	case paint.ItemRoadStation:
		el, _ = ps.Item.(*scenario.StationElement)
	}
	// stationAdjustedInteraction, then stationLabelAdjustedInteraction
	if el == nil || el.AiAllocated || el.Ghost {
		return 0, false
	}
	id := int(el.StationID)
	if id >= len(m.StationOwners) || m.StationOwners[id] != me {
		return 0, false
	}
	return id, true
}

// SceneReady reports whether the main view has been painted by the scene
// renderer (hit-testing needs it).
func (w *World) SceneReady() bool { return w.scene != nil }
