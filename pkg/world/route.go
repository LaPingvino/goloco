package world

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/scene"
	"github.com/LaPingvino/goloco/pkg/sim"
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
	h, ok := w.RouteAt(sx, sy, uint8(objects.TransportModeRail))
	if !ok || h.Waypoint != nil {
		return 0, false
	}
	return h.Station, true
}

// RouteHit is what a click with the Route tool means: a station to stop at,
// or (Waypoint set) a route-through waypoint order on plain track or road.
type RouteHit struct {
	Station  int
	Waypoint []byte // sim.OrderRouteWaypointRaw, for track / road without a station
}

// RouteAt is Route::getRouteInteractionFromCursor and onToolDown's choice
// for a vehicle of the given transport mode: trackAdjustedInteraction (the
// company's own track; its station if it has one, else a waypoint — not for
// ships or aircraft), roadAdjustedInteraction (any owner's road; its station,
// else a waypoint — not for aircraft), the station elements as before. The
// waypoint is at the piece's first tile, its height less the part's z
// (0x004B5160 / 0x004B5223). PARTIAL: water waypoints (ships) and the station
// label hit; the waypoint sound.
func (w *World) RouteAt(sx, sy int, mode uint8) (RouteHit, bool) {
	if w.scene == nil && w.renderer != nil && w.renderer.G1 != nil && w.renderer.ObjMgr != nil {
		w.scene = scene.New(w.renderer.G1, w.renderer.ObjMgr)
		w.scene.WaveFrame = func(tx, ty int) uint8 { return uint8(w.Waves[sim.WaveIndex(tx, ty)].Frame) } // as drawScene's first frame makes it
	}
	if w.scene == nil {
		return RouteHit{}, false
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
		return RouteHit{}, false
	}
	t := m.TileAt(int(ps.MapX)/32, int(ps.MapY)/32)
	if t == nil {
		return RouteHit{}, false
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
	usable := func(el *scenario.StationElement) bool { return el != nil && !el.AiAllocated && !el.Ghost }
	// waypoint: the piece's first tile and height (TrackData's part offsets)
	waypoint := func(parts []sim.PreviewTrack, seq, rot, id uint8, baseZ uint8) (RouteHit, bool) {
		if int(seq) >= len(parts) {
			return RouteHit{}, false
		}
		p := parts[seq]
		ox, oy := rotateVec(int(p.X), int(p.Y), rot)
		fx, fy := int(ps.MapX)/32*32-ox, int(ps.MapY)/32*32-oy
		height := int(baseZ)*4 - int(p.Z)
		return RouteHit{Station: -1, Waypoint: sim.OrderRouteWaypointRaw(fx/32, fy/32, uint8(height/8), rot, id)}, true
	}
	var el *scenario.StationElement
	switch ps.Type {
	case paint.ItemTrack: // trackAdjustedInteraction
		tk, _ := ps.Item.(*scenario.TrackElement)
		if tk == nil || tk.AiAllocated || tk.Ghost || tk.Owner != me {
			return RouteHit{}, false
		}
		if tk.HasStation {
			if st := stationOn(tk.BaseZ); usable(st) {
				el = st // trainStationAdjustedInteraction: the track is the company's
				break
			}
		}
		if mode == uint8(objects.TransportModeWater) || mode == uint8(objects.TransportModeAir) {
			return RouteHit{}, false
		}
		if int(tk.TrackID) >= len(sim.TrackPieces) {
			return RouteHit{}, false
		}
		return waypoint(sim.TrackPieces[tk.TrackID], tk.SequenceIndex, tk.Rotation, tk.TrackID, tk.BaseZ)
	case paint.ItemRoad: // roadAdjustedInteraction: any owner's road
		rd, _ := ps.Item.(*scenario.RoadElement)
		if rd == nil || rd.AiAllocated || rd.Ghost {
			return RouteHit{}, false
		}
		if rd.HasStationElement {
			if st := stationOn(rd.BaseZ); usable(st) {
				el = st
				break
			}
		}
		if mode == uint8(objects.TransportModeAir) {
			return RouteHit{}, false
		}
		if int(rd.RoadID) >= len(sim.RoadPieces) {
			return RouteHit{}, false
		}
		return waypoint(sim.RoadPieces[rd.RoadID], rd.SequenceIndex, rd.Rotation, rd.RoadID, rd.BaseZ)
	case paint.ItemTrainStation: // trainStationAdjustedInteraction: the track under it the company's
		el, _ = ps.Item.(*scenario.StationElement)
		if el == nil {
			return RouteHit{}, false
		}
		var trackMine bool
		for _, tk := range t.Tracks {
			if tk.BaseZ == el.BaseZ && tk.HasStation {
				trackMine = tk.Owner == me
			}
		}
		if !trackMine {
			return RouteHit{}, false
		}
	case paint.ItemAirport, paint.ItemDock: // dockAirportAdjustedInteraction
		el, _ = ps.Item.(*scenario.StationElement)
		if el == nil || el.Owner != me {
			return RouteHit{}, false
		}
	case paint.ItemRoadStation:
		el, _ = ps.Item.(*scenario.StationElement)
	}
	// stationAdjustedInteraction, then stationLabelAdjustedInteraction
	if !usable(el) {
		return RouteHit{}, false
	}
	id := int(el.StationID)
	if id >= len(m.StationOwners) || m.StationOwners[id] != me {
		return RouteHit{}, false
	}
	return RouteHit{Station: id}, true
}

// rotateVec is Math::Vector::rotate.
func rotateVec(x, y int, dir uint8) (int, int) {
	switch dir & 3 {
	case 1:
		return y, -x
	case 2:
		return -x, -y
	case 3:
		return -y, x
	}
	return x, y
}

// SceneReady reports whether the main view has been painted by the scene
// renderer (hit-testing needs it).
func (w *World) SceneReady() bool { return w.scene != nil }
