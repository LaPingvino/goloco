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
	Kind paint.InteractionItem // ItemEntity, ItemTownLabel, ItemStationLabel, ItemIndustry
	ID   int                   // the head's entity id, the town, station or industry
}

// ItemLeftAt is getItemLeft at screen point (sx, sy): first only vehicles and
// the labels, then — when that found no vehicle — vehicles, track, road,
// stations, industries and the labels (getMapCoordinatesFromPos: the topmost
// element, then the station labels up to stationNamesMinScale, then the town
// labels, the last match of each winning). Track, road and station elements
// count as their station (_track / _road / getStationArguments), which a click
// does not open. ok is false where the click opens nothing, and in the title.
// PARTIAL: headquarters (the HQ building's interaction isn't painted yet) and
// a window's own viewport (getItemLeft's fallback). The game acts on the
// press; upstream on the release of a click that didn't drag (viewportLeft).
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
	return pass(func(i paint.InteractionItem) bool {
		switch i {
		case paint.ItemEntity, paint.ItemTrack, paint.ItemRoad, paint.ItemTrainStation, paint.ItemRoadStation,
			paint.ItemAirport, paint.ItemDock, paint.ItemIndustry, paint.ItemHeadquarterBuilding:
			return true
		}
		return false
	}, true)
}
