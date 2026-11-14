package world

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// rightover.go — ViewportInteraction::rightOver (0x004CDB2B) and its
// rightOver… functions: what a right click on the main view would act on,
// and the map tooltip's text for it.
// OpenLoco reference: src/OpenLoco/src/Ui/ViewportInteraction.cpp

// RightItem is rightOver's InteractionArg: the item's type as the rightOver
// functions leave it (a signal or station off its tab is its track or road;
// a building outside a construction mode is buildingInfo), the tile element
// (Track, Road, Signal, Station, Tree, Wall, Building: pointers into the
// tile), the tile's position and the paint struct's mod id (a track or road
// mod's index; a signal's side, 0 left, 1 right). ID is a label's town or
// station.
type RightItem struct {
	Type  paint.InteractionItem
	Tile  *sim.Tile
	Track *scenario.TrackElement
	Road  *scenario.RoadElement
	// Signal's and Station's track or road (the element before them) is in
	// Track / Road.
	Signal   *scenario.SignalElement
	Station  *scenario.StationElement
	Tree     *scenario.TreeElement
	Wall     *scenario.WallElement
	Building *scenario.BuildingElement
	X, Y     int32
	ModID    uint8
	ID       int
}

// RightOverState is what rightOver reads from the windows and the input.
type RightOverState struct {
	// ConstructionOpen: WindowManager::find(WindowType::construction); the
	// tab flags are Construction::isStationTabOpen / isSignalTabOpen /
	// isOverheadTabOpen.
	ConstructionOpen                   bool
	StationTab, SignalTab, OverheadTab bool
	// RepeatedSignals is cState.repeatedSignalMode (the Signal tab's
	// Repeated placement): the signal's text says "Signals block".
	RepeatedSignals bool
	// ConstructTool is Input's toolActive with flag6: a tool that builds
	// (the terraform tools, the industry tool, the Construction tab's).
	ConstructTool bool
	// TipTimeout is MapToolTip::getTooltipTimeout: the track, road and
	// building texts wait for 45.
	TipTimeout uint16
	// SeeThrough… are the viewport's flags: what they hide is not found.
	// PARTIAL: goloco's view has none of them yet (always false).
	SeeThroughBuildings, SeeThroughTrees, SeeThroughScenery, SeeThroughTracks, SeeThroughRoads bool
}

// tipFormat is formatString for a map tooltip: the string's {STRINGID}s
// filled from args in order, its newlines and {COLOUR WINDOW_3} as ui.MapTip
// reads them.
func tipFormat(id int, args ...string) string {
	s := i18n.Str(id)
	for _, a := range args {
		s = strings.Replace(s, "{STRINGID}", a, 1)
	}
	s = strings.ReplaceAll(s, "{NEWLINE}", tipNewline)
	return strings.ReplaceAll(s, "{COLOUR WINDOW_3}", string(ui.MapTipColour3))
}

// orderPos is the place of an element (a pointer into the tile) in the
// tile's order, -1 none.
func orderPos(t *sim.Tile, item any) int {
	for k, r := range t.Order {
		i := int(r.Index)
		switch el := item.(type) {
		case *scenario.TrackElement:
			if r.Kind == scenario.KindTrack && &t.Tracks[i] == el {
				return k
			}
		case *scenario.RoadElement:
			if r.Kind == scenario.KindRoad && &t.Roads[i] == el {
				return k
			}
		case *scenario.SignalElement:
			if r.Kind == scenario.KindSignal && &t.Signals[i] == el {
				return k
			}
		case *scenario.StationElement:
			if r.Kind == scenario.KindStation && &t.Stations[i] == el {
				return k
			}
		}
	}
	return -1
}

// prevTrackRoad is tileElement->prev() of an element as a track or road
// element (nil: it is neither, or there is none).
func prevTrackRoad(t *sim.Tile, item any) (*scenario.TrackElement, *scenario.RoadElement) {
	k := orderPos(t, item)
	if k < 1 {
		return nil, nil
	}
	switch r := t.Order[k-1]; r.Kind {
	case scenario.KindTrack:
		return &t.Tracks[r.Index], nil
	case scenario.KindRoad:
		return nil, &t.Roads[r.Index]
	}
	return nil, nil
}

// RightOver is ViewportInteraction::rightOver at screen point (sx, sy): the
// item a right click would act on (ok false: none — InteractionArg{}), and,
// when hasTip, the map tooltip it sets (FormatArguments::mapToolTip, with
// setOwner's company; otherwise the tooltip stays as it was).
// PARTIAL: viewports other than the main view's.
func (w *World) RightOver(sx, sy int, st RightOverState) (item RightItem, tip MapTip, hasTip, ok bool) {
	if w.TitleMode || !w.ensureScene() {
		return RightItem{}, MapTip{}, false, false
	}
	// "Interaction types to exclude by default": the surface and the water,
	// then what the view's see-through flags hide.
	accept := func(i paint.InteractionItem) bool {
		switch i {
		case paint.ItemSurface, paint.ItemWater:
			return false
		case paint.ItemBuilding, paint.ItemHeadquarterBuilding, paint.ItemIndustry:
			return !st.SeeThroughBuildings
		case paint.ItemTree:
			return !st.SeeThroughTrees
		case paint.ItemWall:
			return !st.SeeThroughScenery
		case paint.ItemTrack, paint.ItemTrackExtra, paint.ItemSignal:
			return !st.SeeThroughTracks
		case paint.ItemRoad, paint.ItemRoadExtra:
			return !st.SeeThroughRoads
		}
		return true
	}
	in := w.mapCoordinatesAt(sx, sy, accept)
	item = RightItem{Type: in.Type, ID: in.ID}
	if ps := in.ps; ps != nil {
		item.X, item.Y, item.ModID = int32(ps.MapX), int32(ps.MapY), ps.ModID
		item.Tile = w.Map.TileAt(ps.MapX>>5, ps.MapY>>5)
		switch el := ps.Item.(type) {
		case *scenario.TrackElement:
			item.Track = el
		case *scenario.RoadElement:
			item.Road = el
		case *scenario.SignalElement:
			item.Signal = el
		case *scenario.StationElement:
			item.Station = el
		case *scenario.TreeElement:
			item.Tree = el
		case *scenario.WallElement:
			item.Wall = el
		case *scenario.BuildingElement:
			item.Building = el
		}
		if item.Tile != nil && (item.Signal != nil || item.Station != nil) {
			item.Track, item.Road = prevTrackRoad(item.Tile, ps.Item)
		}
	}
	r := rightOver{w: w, st: st, item: &item}
	switch in.Type {
	case paint.ItemTrack:
		ok = r.track()
	case paint.ItemEntity: // "No interaction"
	case paint.ItemTrackExtra:
		ok = r.trackExtra()
	case paint.ItemSignal:
		ok = r.signal()
	case paint.ItemTrainStation:
		ok = r.trainStation()
	case paint.ItemRoadStation:
		ok = r.roadStation()
	case paint.ItemAirport:
		ok = r.airportOrDock(i18n.QuoteStringQuote)
	case paint.ItemDock:
		ok = r.airportOrDock(i18n.QuoteStringQuote2)
	case paint.ItemRoad:
		ok = r.road()
	case paint.ItemRoadExtra:
		ok = r.roadExtra()
	default:
		if !st.ConstructTool && !st.ConstructionOpen {
			if in.Type == paint.ItemBuilding {
				ok = r.building()
			}
			break
		}
		switch in.Type { // 0x4CDBC5
		case paint.ItemTree:
			ok = r.tree()
		case paint.ItemWall:
			ok = r.wall()
		case paint.ItemBuilding:
			ok = r.buildingConstruct()
		case paint.ItemHeadquarterBuilding:
			ok = r.headquarters()
		default:
			ok = true
		}
	}
	if !ok {
		return RightItem{}, MapTip{}, false, false
	}
	return item, r.tip, r.hasTip, true
}

// rightOver carries one RightOver call through its rightOver… functions.
type rightOver struct {
	w      *World
	st     RightOverState
	item   *RightItem
	tip    MapTip
	hasTip bool
}

func (r *rightOver) set(text string, owner int) {
	r.tip, r.hasTip = MapTip{Text: text, Owner: owner}, true
}

func (r *rightOver) om() *objects.ObjectManager { return r.w.renderer.ObjMgr }

func (r *rightOver) controlling() int {
	c, _ := r.w.ControllingCompany()
	return c
}

// track is rightOverTrack (0x004CE1D4): the track's name with "Right-Click
// to Modify", another company's with "Owned by" and its face.
func (r *rightOver) track() bool {
	r.item.Type = paint.ItemTrack
	t := r.item.Track
	if t == nil || t.Ghost {
		return false
	}
	if r.st.TipTimeout < 45 {
		return true
	}
	name := ""
	if o := r.om().GetTrackObjectByIndex(int(t.TrackObjectID)); o != nil {
		name = o.Name
	}
	if int(t.Owner) == r.controlling() {
		r.set(tipFormat(i18n.StringidRightClickToModify, name), -1)
	} else {
		r.set(tipFormat(i18n.StringOwnedByString, name, r.w.Map.CompanyName(int(t.Owner))), int(t.Owner))
	}
	return true
}

// modName is the name of a track or road object's mod i.
func (r *rightOver) modName(mods []string, i uint8) string {
	if int(i) < len(mods) {
		return r.om().ExtraName(mods[i])
	}
	return ""
}

// trackExtra is rightOverTrackExtra (0x004CE18F): on the Overhead tab, a mod
// the track has is "Right-Click to Remove"; otherwise the track's.
func (r *rightOver) trackExtra() bool {
	if !r.st.OverheadTab {
		return r.track()
	}
	t := r.item.Track
	if t == nil || t.Ghost {
		return false
	}
	if t.Mods&(1<<r.item.ModID) == 0 {
		return r.track()
	}
	var mods []string
	if o := r.om().GetTrackObjectByIndex(int(t.TrackObjectID)); o != nil {
		mods = o.ModNames
	}
	r.set(tipFormat(i18n.StringidRightClickToRemove, r.modName(mods, r.item.ModID)), -1)
	return true
}

// signal is rightOverSignal (0x004CDBEA): off the Signal tab the track's;
// on it, the player's own signal is "Signal — Right-Click to Remove".
func (r *rightOver) signal() bool {
	s, t := r.item.Signal, r.item.Track
	if s == nil || t == nil || s.Ghost {
		return false
	}
	if !r.st.SignalTab {
		return r.track()
	}
	if int(t.Owner) != r.controlling() {
		return false
	}
	caption := i18n.CaptSignal
	if r.st.RepeatedSignals {
		caption = i18n.CaptSignalsBlock
	}
	r.set(tipFormat(i18n.StringidRightClickToRemove, i18n.Str(caption)), -1)
	return true
}

// trainStation is rightOverTrainStation (0x004CDD8C).
func (r *rightOver) trainStation() bool {
	s, t := r.item.Station, r.item.Track
	if s == nil || t == nil || s.Ghost {
		return false
	}
	if !r.st.StationTab {
		return r.track()
	}
	if int(t.Owner) != r.controlling() {
		return false
	}
	r.set(tipFormat(i18n.StringidRightClickToRemove, tipFormat(i18n.StringStationPlatform, r.w.Map.StationName(int(s.StationID)))), -1)
	return true
}

// road is rightOverRoad (0x004CE2C1): as the track's; a neutral road too is
// the player's to modify.
func (r *rightOver) road() bool {
	r.item.Type = paint.ItemRoad
	rd := r.item.Road
	if rd == nil || rd.Ghost {
		return false
	}
	if r.st.TipTimeout < 45 {
		return true
	}
	name := ""
	if o := r.om().GetRoadObjectByIndex(int(rd.RoadObjectID)); o != nil {
		name = o.Name
	}
	if int(rd.Owner) == r.controlling() || rd.Owner == sim.CompanyNeutral {
		r.set(tipFormat(i18n.StringidRightClickToModify, name), -1)
	} else {
		r.set(tipFormat(i18n.StringOwnedByString, name, r.w.Map.CompanyName(int(rd.Owner))), int(rd.Owner))
	}
	return true
}

// roadExtra is rightOverRoadExtra (0x004CE271).
func (r *rightOver) roadExtra() bool {
	if !r.st.OverheadTab {
		return r.road()
	}
	rd := r.item.Road
	if rd == nil || rd.Ghost {
		return false
	}
	if rd.Mods&(1<<r.item.ModID) == 0 {
		return r.road()
	}
	var mods []string
	if o := r.om().GetRoadObjectByIndex(int(rd.RoadObjectID)); o != nil {
		mods = o.ModNames
	}
	r.set(tipFormat(i18n.StringidRightClickToRemove, r.modName(mods, r.item.ModID)), -1)
	return true
}

// roadStation is rightOverRoadStation (0x004CDDF2): the station's owner
// decides, not the road's.
func (r *rightOver) roadStation() bool {
	s, rd := r.item.Station, r.item.Road
	if s == nil || rd == nil || s.Ghost {
		return false
	}
	if !r.st.StationTab {
		return r.road()
	}
	owner, ok := r.w.Map.StationOwner(int(s.StationID))
	if !ok || int(owner) != r.controlling() {
		return false
	}
	r.set(tipFormat(i18n.StringidRightClickToRemove, tipFormat(i18n.StringStationBuildingBusStop, r.w.Map.StationName(int(s.StationID)))), -1)
	return true
}

// airportOrDock is rightOverAirport (0x004CDC26) / rightOverDock
// (0x004CDCD9): the player's own, to remove on the Station tab, else to
// modify (the click opens the airport / dock window).
func (r *rightOver) airportOrDock(quote int) bool {
	s := r.item.Station
	if s == nil || s.Ghost || int(s.Owner) != r.controlling() {
		return false
	}
	id := i18n.StringidRightClickToModify
	if r.st.StationTab {
		id = i18n.StringidRightClickToRemove
	}
	r.set(tipFormat(id, tipFormat(quote, r.w.Map.StationName(int(s.StationID)))), -1)
	return true
}

// tree is rightOverTree (0x004CDE58).
func (r *rightOver) tree() bool {
	if r.item.Tree == nil {
		return false
	}
	name := ""
	if o := r.om().GetTreeObjectByIndex(int(r.item.Tree.TreeObjectID)); o != nil {
		name = o.Name
	}
	r.set(tipFormat(i18n.StringidRightClickToRemove, name), -1)
	return true
}

// wall is rightOverWall (0x004CE0D8).
func (r *rightOver) wall() bool {
	if r.item.Wall == nil {
		return false
	}
	name := ""
	if o := r.om().GetWallObjectByIndex(int(r.item.Wall.WallObjectID)); o != nil {
		name = o.Name
	}
	r.set(tipFormat(i18n.StringidRightClickToRemove, name), -1)
	return true
}

// buildingConstruct is rightOverBuildingConstruct (0x004CDE78), in a
// constructing mode: the building's name, "Right-Click to Remove" unless it
// is indestructible (the bare name then). PARTIAL: the editor and sandbox
// modes, where every building can go.
func (r *rightOver) buildingConstruct() bool {
	b := r.item.Building
	if b == nil {
		return false
	}
	o := r.om().GetBuildingObjectByIndex(int(b.ObjectID))
	if o == nil {
		return false
	}
	if o.HasFlags(objects.BuildingFlagIndestructible) {
		r.set(o.Name, -1)
	} else {
		r.set(tipFormat(i18n.StringidRightClickToRemove, o.Name), -1)
	}
	return true
}

// headquarters is rightOverHeadquarters (0x004CE107).
func (r *rightOver) headquarters() bool {
	b := r.item.Building
	if b == nil {
		return false
	}
	id, ok := r.w.headquartersCompany(b, int(r.item.X), int(r.item.Y))
	if !ok {
		return false
	}
	r.set(tipFormat(i18n.StringidRightClickToRemove, tipFormat(i18n.StringidHeadquarters, r.w.Map.CompanyName(id))), -1)
	return true
}

// building is rightOverBuilding (0x004CDEB8), outside a constructing mode:
// the item becomes buildingInfo (a click does nothing with it); the text is
// the building's name with "Under construction", or what it produces and
// accepts (kQuantityToString: the eighths of a unit it takes).
func (r *rightOver) building() bool {
	r.item.Type = paint.ItemBuildingInfo
	if r.st.TipTimeout < 45 {
		return true
	}
	b := r.item.Building
	if b == nil {
		return false
	}
	om := r.om()
	o := om.GetBuildingObjectByIndex(int(b.ObjectID))
	if o == nil {
		return false
	}
	cargoName := func(id uint8) string {
		if int(id) < len(om.CargoObjects) && om.CargoObjects[id] != nil {
			return om.CargoObjects[id].Name
		}
		return ""
	}
	var sb strings.Builder
	sb.WriteString(o.Name)
	if !b.IsConstructed {
		sb.WriteString(tipFormat(i18n.UnderConstruction))
	} else {
		produced, consumed := om.ProducedCargoIDs(o), om.ConsumedCargoIDs(o)
		if o.ProducedQuantity[0] != 0 || o.ProducedQuantity[1] != 0 {
			sb.WriteString(tipFormat(i18n.Produces))
			comma := false
			for i := 0; i < 2; i++ {
				if produced[i] != 0xFF && o.ProducedQuantity[i] != 0 {
					if comma {
						sb.WriteString(tipFormat(i18n.Comma))
					}
					comma = true
					sb.WriteString(cargoName(produced[i]))
				}
			}
		}
		if o.ProducedCargoQty[0] != 0 || o.ProducedCargoQty[1] != 0 || o.ConsumedCargoQty[0] != 0 || o.ConsumedCargoQty[1] != 0 {
			sb.WriteString(tipFormat(i18n.Accepts))
			comma := false
			accepts := func(ids, qty [2]uint8) {
				for i := 0; i < 2; i++ {
					if ids[i] != 0xFF && qty[i] != 0 {
						if comma {
							sb.WriteString(tipFormat(i18n.Comma))
						}
						if qty[i] < 8 {
							sb.WriteString(tipFormat(i18n.QuantityEigth + int(qty[i]) - 1))
						}
						comma = true
						sb.WriteString(cargoName(ids[i]))
					}
				}
			}
			accepts(produced, o.ProducedCargoQty)
			accepts(consumed, o.ConsumedCargoQty)
		}
	}
	r.set(sb.String(), -1)
	return true
}

// WorldToScreen is where the main view shows a world position: gameToScreen,
// then Viewport::viewportToWindow.
func (w *World) WorldToScreen(x, y, z int32) (int, int) {
	vx, vy := w.Map.GameToScreen(x, y, z)
	return (int(vx) - int(w.camX)) >> w.zoom, (int(vy) - int(w.camY)) >> w.zoom
}

// CameraPos is the main view's position (the viewport's view x / y).
func (w *World) CameraPos() (int, int) { return int(w.camX), int(w.camY) }
