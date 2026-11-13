package world

import (
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/scene"
)

// LeftItem is what a left click on the main view points at
// (ViewportInteraction::getItemLeft, as MouseInput.cpp's viewport click uses
// it): a vehicle (its head), a town label, a station label or an industry.
type LeftItem struct {
	Kind paint.InteractionItem // ItemEntity, ItemTownLabel, ItemStationLabel, ItemIndustry, ItemHeadquarterBuilding
	ID   int                   // the head's entity id, the town, station, industry or company
}

// ItemLeftAt is getItemLeft at screen point (sx, sy): first only vehicles and
// the labels, then — when that found no vehicle — vehicles, track, road,
// stations, industries and the labels (getMapCoordinatesFromPos: the topmost
// element, then the station labels up to stationNamesMinScale, then the town
// labels, the last match of each winning). Track, road and station elements
// count as their station (_track / _road / getStationArguments), which a click
// does not open. ok is false where the click opens nothing, and in the title.
// A headquarters building gives its company; the nearest vehicle part when
// nothing was hit (nearestVehicleAt).
func (w *World) ItemLeftAt(sx, sy int) (LeftItem, bool) {
	if w.TitleMode || !w.ensureScene() {
		return LeftItem{}, false
	}
	pass := func(accept func(paint.InteractionItem) bool, industries bool) (LeftItem, bool) {
		var it LeftItem
		found := false
		v := scene.View{X: int(w.camX), Y: int(w.camY), Zoom: w.zoom}
		if ps := w.scene.InteractionAt(&w.Map, v, sx, sy, accept); ps != nil {
			switch ps.Type {
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
					off := [4][2]int{{0, 0}, {0, 32}, {32, 32}, {32, 0}}[el.SequenceIndex&3] // World::kOffsets
					fx, fy := int16(ps.MapX-off[0]), int16(ps.MapY-off[1])
					for id := range w.Map.Companies {
						if co := &w.Map.Companies[id]; co.HeadquartersX == fx && co.HeadquartersY == fy && co.HeadquartersZ == el.BaseZ {
							it, found = LeftItem{Kind: paint.ItemHeadquarterBuilding, ID: id}, true
							break
						}
					}
				}
			default: // track, road, station elements: no window from a click
				it, found = LeftItem{Kind: ps.Type}, false
			}
		}
		// the 1×1 render target's getDrawableRect at this zoom
		z := max(0, min(3, w.zoom))
		mask := ^0 << z
		x, y := (int(w.camX)+sx<<z)&mask, (int(w.camY)+sy<<z)&mask
		l, t := x>>z, y>>z
		r, b := l+(1>>z), t+(1>>z)
		if w.zoom <= stationNamesMinScale { // getStationNameInteractionInfo
			for id := range w.Map.StationNames {
				st := &w.Map.StationNames[id]
				if st.Name == 0xFFFF || st.Flags&(1<<5) != 0 || !st.LabelFrame.Contains(l, t, r, b, z) {
					continue
				}
				it, found = LeftItem{Kind: paint.ItemStationLabel, ID: id}, true
			}
		}
		for i := range w.Map.Towns { // getTownNameInteractionInfo
			if tn := &w.Map.Towns[i]; tn.Name != 0xFFFF && tn.LabelFrame.Contains(l, t, r, b, z) {
				it, found = LeftItem{Kind: paint.ItemTownLabel, ID: i}, true
			}
		}
		return it, found
	}
	// the first pass's result stands when it is an entity (a vehicle or not)
	if it, ok := pass(func(i paint.InteractionItem) bool { return i == paint.ItemEntity }, false); it.Kind == paint.ItemEntity {
		return it, ok
	}
	if it, ok := pass(func(i paint.InteractionItem) bool {
		switch i {
		case paint.ItemEntity, paint.ItemTrack, paint.ItemRoad, paint.ItemTrainStation, paint.ItemRoadStation,
			paint.ItemAirport, paint.ItemDock, paint.ItemIndustry, paint.ItemHeadquarterBuilding:
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
