package world

import (
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// LeftItem is what a left click on the main view points at
// (ViewportInteraction::getItemLeft, as MouseInput.cpp's viewport click uses
// it): a vehicle (its head), a town label, a station label or an industry.
type LeftItem struct {
	Kind paint.InteractionItem // ItemEntity, ItemTownLabel, ItemStationLabel, ItemIndustry, ItemHeadquarterBuilding
	ID   int                   // the head's entity id, the town, station, industry or company
}

// interaction is an InteractionArg as getMapCoordinatesFromPos returns it:
// the paint struct hit (nil for a label) or the label's town / station.
type interaction struct {
	Type paint.InteractionItem
	ps   *paint.Struct
	ID   int
}

// mapCoordinatesAt is ViewportInteraction::getMapCoordinatesFromPos for the
// main view at screen point (sx, sy): the topmost element accept takes
// (getNormalInteractionInfo), then the station labels up to
// stationNamesMinScale, then the town labels, the last match of each winning
// (the labels too only when accept takes them).
func (w *World) mapCoordinatesAt(sx, sy int, accept func(paint.InteractionItem) bool) interaction {
	var in interaction
	v := w.mainView()
	if ps := w.scene.InteractionAt(&w.Map, v, sx, sy, accept); ps != nil {
		in = interaction{Type: ps.Type, ps: ps}
	}
	// the 1×1 render target's getDrawableRect at this zoom
	z := max(0, min(3, w.zoom))
	mask := ^0 << z
	x, y := (int(w.camX)+sx<<z)&mask, (int(w.camY)+sy<<z)&mask
	l, t := x>>z, y>>z
	r, b := l+(1>>z), t+(1>>z)
	// the labels a viewport hides (hideStationNames / hideTownNames) aren't found
	if !w.viewFlags.Has(paint.ViewHideStationNames) && w.zoom <= stationNamesMinScale && accept(paint.ItemStationLabel) { // getStationNameInteractionInfo
		for id := range w.Map.StationNames {
			st := &w.Map.StationNames[id]
			if st.Name == 0xFFFF || st.Flags&(1<<5) != 0 || !st.LabelFrame.Contains(l, t, r, b, z) {
				continue
			}
			in = interaction{Type: paint.ItemStationLabel, ID: id}
		}
	}
	if !w.viewFlags.Has(paint.ViewHideTownNames) && accept(paint.ItemTownLabel) {
		for i := range w.Map.Towns { // getTownNameInteractionInfo
			if tn := &w.Map.Towns[i]; tn.Name != 0xFFFF && tn.LabelFrame.Contains(l, t, r, b, z) {
				in = interaction{Type: paint.ItemTownLabel, ID: i}
			}
		}
	}
	return in
}

// headquartersCompany is the company whose headquarters' first tile the
// building element painted at (mapX, mapY) belongs to (getHeadquarterArguments,
// rightOverHeadquarters).
func (w *World) headquartersCompany(el *scenario.BuildingElement, mapX, mapY int) (int, bool) {
	off := [4][2]int{{0, 0}, {0, 32}, {32, 32}, {32, 0}}[el.SequenceIndex&3] // World::kOffsets
	fx, fy := int16(mapX-off[0]), int16(mapY-off[1])
	for id := range w.Map.Companies {
		if co := &w.Map.Companies[id]; co.HeadquartersX == fx && co.HeadquartersY == fy && co.HeadquartersZ == el.BaseZ {
			return id, true
		}
	}
	return 0, false
}

// ItemLeftAt is getItemLeft at screen point (sx, sy): first only vehicles and
// the labels, then — when that found no vehicle — vehicles, track, road,
// stations, industries and the labels (mapCoordinatesAt). A track, road or station element
// of a station counts as that station's label (_track / _road /
// getStationArguments). ok is false where the click opens nothing, and in the title.
// A headquarters building gives its company; the nearest vehicle part when
// nothing was hit (nearestVehicleAt).
func (w *World) ItemLeftAt(sx, sy int) (LeftItem, bool) {
	if w.TitleMode || !w.ensureScene() {
		return LeftItem{}, false
	}
	pass := func(accept func(paint.InteractionItem) bool, industries bool) (LeftItem, bool) {
		var it LeftItem
		found := false
		in := w.mapCoordinatesAt(sx, sy, accept)
		ps := in.ps
		switch in.Type {
		case paint.ItemStationLabel, paint.ItemTownLabel:
			it, found = LeftItem{Kind: in.Type, ID: in.ID}, true
		case paint.ItemEntity:
			it = LeftItem{Kind: paint.ItemEntity, ID: -1} // an effect: getVehicleArguments fails
			if e, ok := ps.Item.(*scenario.VehicleEntity); ok {
				it, found = LeftItem{Kind: paint.ItemEntity, ID: int(e.HeadID)}, true
			}
		case paint.ItemIndustry:
			if el, ok := ps.Item.(*scenario.IndustryElement); ok && industries { // ghosts paint as noInteraction
				it, found = LeftItem{Kind: paint.ItemIndustry, ID: int(el.IndustryID)}, true
			}
		case paint.ItemHeadquarterBuilding: // the company whose headquarters' first tile it is
			if el, ok := ps.Item.(*scenario.BuildingElement); ok {
				if id, ok := w.headquartersCompany(el, ps.MapX, ps.MapY); ok {
					it, found = LeftItem{Kind: paint.ItemHeadquarterBuilding, ID: id}, true
				}
			}
		case paint.ItemTrack, paint.ItemRoad, paint.ItemTrainStation, paint.ItemRoadStation, paint.ItemAirport, paint.ItemDock:
			// _track / _road / getStationArguments: a station's element counts as its station label
			it = LeftItem{Kind: ps.Type}
			if id, ok := w.stationOfElement(ps); ok {
				it, found = LeftItem{Kind: paint.ItemStationLabel, ID: id}, true
			}
		}
		return it, found
	}
	// the first pass's result stands when it is an entity (a vehicle or not)
	if it, ok := pass(func(i paint.InteractionItem) bool {
		return i == paint.ItemEntity || i == paint.ItemTownLabel || i == paint.ItemStationLabel
	}, false); it.Kind == paint.ItemEntity {
		return it, ok
	}
	if it, ok := pass(func(i paint.InteractionItem) bool {
		switch i {
		case paint.ItemEntity, paint.ItemTrack, paint.ItemRoad, paint.ItemTrainStation, paint.ItemRoadStation,
			paint.ItemAirport, paint.ItemDock, paint.ItemIndustry, paint.ItemHeadquarterBuilding, paint.ItemTownLabel, paint.ItemStationLabel:
			return true
		}
		return false
	}, true); ok {
		return it, true
	}
	return w.nearestVehicleAt(sx, sy)
}

// vehiclesMinScale is Config::vehiclesMinScale (default 2).
const vehiclesMinScale = 2

// nearestVehicleAt is getItemLeft's fallback when nothing was hit: up to
// vehiclesMinScale, the vehicle part (each train's veh2, then its cars'
// bogies and bodies) whose sprite box holds the point (windowToViewport) and
// whose box centre is nearest by Manhattan distance, if within 32.
func (w *World) nearestVehicleAt(sx, sy int) (LeftItem, bool) {
	if w.zoom > vehiclesMinScale {
		return LeftItem{}, false
	}
	tx, ty := int32(w.camX)+int32(sx)<<w.zoom, int32(w.camY)+int32(sy)<<w.zoom
	trains, err := w.Map.Trains()
	if err != nil {
		return LeftItem{}, false
	}
	best, head := uint32(1<<32-1), -1
	check := func(i, h int) { // checkAndSetNearestVehicle / vehicleDistanceFromLocation
		e := &w.Map.Entities[i]
		if e.SpriteLeft == -0x8000 { // Location::null
			return
		}
		l, t, r, b := int32(e.SpriteLeft), int32(e.SpriteTop), int32(e.SpriteRight), int32(e.SpriteBottom)
		if !(l < tx && t < ty && r >= tx && b >= ty) {
			return
		}
		dx, dy := tx-(r+l)/2, ty-(t+b)/2
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if d := uint32(dx + dy); d < best {
			best, head = d, h
		}
	}
	for _, tr := range trains {
		check(tr.Veh2, tr.Head)
		for _, car := range tr.Cars {
			for _, c := range car {
				check(c.Front, tr.Head)
				check(c.Back, tr.Head)
				check(c.Body, tr.Head)
			}
		}
	}
	if best > 32 || head < 0 {
		return LeftItem{}, false
	}
	return LeftItem{Kind: paint.ItemEntity, ID: int(w.Map.Entities[head].HeadID)}, true
}

// stationOfElement is _track / _road / getStationArguments(interaction) for
// the element ps painted: a track with a station element right after it, a
// road with one after its matching road (same height, rotation, piece and
// object), or a station element itself — not AI-allocated (the track's and
// road's station) nor a ghost.
func (w *World) stationOfElement(ps *paint.Struct) (int, bool) {
	t := w.Map.TileAt(int(ps.MapX)>>5, int(ps.MapY)>>5)
	if t == nil {
		return 0, false
	}
	station := func(st *scenario.StationElement, viaTrackOrRoad bool) (int, bool) {
		if (viaTrackOrRoad && st.AiAllocated) || st.Ghost {
			return 0, false
		}
		return int(st.StationID), true
	}
	switch el := ps.Item.(type) {
	case *scenario.StationElement:
		return station(el, false)
	case *scenario.TrackElement:
		if !el.HasStation {
			return 0, false
		}
		for k, r := range t.Order {
			if r.Kind == scenario.KindTrack && &t.Tracks[r.Index] == el {
				if k+1 < len(t.Order) && t.Order[k+1].Kind == scenario.KindStation {
					return station(&t.Stations[t.Order[k+1].Index], true)
				}
				return 0, false
			}
		}
	case *scenario.RoadElement:
		if !el.HasStationElement {
			return 0, false
		}
		foundRoad := false
		for _, r := range t.Order {
			if foundRoad {
				if r.Kind == scenario.KindStation {
					return station(&t.Stations[r.Index], true)
				}
				continue
			}
			if r.Kind == scenario.KindRoad {
				if o := &t.Roads[r.Index]; o.BaseZ == el.BaseZ && o.Rotation == el.Rotation && o.RoadID == el.RoadID && o.RoadObjectID == el.RoadObjectID {
					foundRoad = true
				}
			}
		}
	}
	return 0, false
}

// TrackUnder is getMapCoordinatesFromPos for tracks only (SignalTab.cpp
// getSignalPlacementArgsFromCursor): the track element at screen point
// (sx, sy) and its tile (world units); ok false where there is none.
func (w *World) TrackUnder(sx, sy int) (el scenario.TrackElement, x, y int32, ok bool) {
	if w.TitleMode || !w.ensureScene() {
		return el, 0, 0, false
	}
	in := w.mapCoordinatesAt(sx, sy, func(i paint.InteractionItem) bool { return i == paint.ItemTrack })
	if in.Type != paint.ItemTrack || in.ps == nil {
		return el, 0, 0, false
	}
	e, isTrack := in.ps.Item.(*scenario.TrackElement)
	if !isTrack {
		return el, 0, 0, false
	}
	return *e, int32(in.ps.MapX), int32(in.ps.MapY), true
}
