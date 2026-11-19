package sim

import (
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The map window's other tabs (OpenLoco reference:
// src/OpenLoco/src/Ui/Windows/MapWindow.cpp; PORT-PLANS A step d): the pixel
// fills setMapPixelsVehicles (l. 447), setMapPixelsIndustries (l. 549),
// setMapPixelsRoutes (l. 671) and setMapPixelsOwnership (l. 806), the colours
// open assigns (assignIndustryColours l. 2307, assignRouteColours l. 2356),
// and drawScroll's vehicles and routes (drawVehiclesOnMap l. 1913,
// countVehiclesOnMap l. 1886).

// The map window's tabs (currentTab: widx::tabOverall … tabOwnership).
const (
	MapTabOverall = iota
	MapTabVehicles
	MapTabIndustries
	MapTabRoutes
	MapTabOwnership
)

// More of Colour.h's PaletteIndex the map uses.
const (
	palBlack1           = 0x0B
	palBlack4           = 0x0E
	palBlack8           = 0x12
	palMutedOliveGreen9 = 0x1F
	palMutedDarkYellow2 = 0x24
	palMutedDarkYellow7 = 0x29
	palYellow7          = 0x35
	palYellowA          = 0x38
	palMutedDarkRed2    = 0x3C
	palMutedDarkRed5    = 0x3F
	palMutedDarkRed9    = 0x43
	palMutedGrassGreen5 = 0x4B
	palMutedGrassGreenA = 0x50
	palMutedAvocado6    = 0x58
	palGreen8           = 0x66
	palGreen9           = 0x67
	palMutedOrange7     = 0x71
	palBlue3            = 0x85
	palBlue7            = 0x89
	palBlue9            = 0x8B
	palPurple3          = 0x9D
	palPurple7          = 0xA1
	palPurple8          = 0xA2
	palPurple9          = 0xA3
	palRed6             = 0xAC
	palRed7             = 0xAD
	palOrange6          = 0xB8
	palOrange9          = 0xBB
	palOrangeA          = 0xBC
	palMutedDarkTeal5   = 0xC3
	palMutedDarkTeal8   = 0xC6
	palPink6            = 0xD0
	palPink9            = 0xD3
	palBrown5           = 0xDB
	palBrown8           = 0xDE
)

// MapIndustryColours is kIndustryColours (0x004FB464): the colours an
// industry type (and a route type) can be given.
var MapIndustryColours = [31]uint8{
	palBlack0, palBlack4, palBlackB, palMutedOliveGreen9,
	palMutedDarkYellow7, palYellow7, palYellowA, palMutedDarkRed5,
	palMutedDarkRed9, palMutedGrassGreen5, palMutedGrassGreenA, palMutedAvocado6,
	palGreen8, palMutedOrange7, palMutedPurple7, palBlue3,
	palBlue7, palPurple3, palPurple7, palPurple9,
	palRed6, palOrange6, palOrange9, palMutedDarkTeal5,
	palMutedDarkTeal8, palPink6, palPink9, palBrown5,
	palBrown8, palMutedDarkYellow2, palBlack8,
}

// MapVehicleTypeColours is vehicleTypeColours (0x004FDD62), by VehicleType:
// train, bus, truck, tram, aircraft, ship.
var MapVehicleTypeColours = [6]uint8{palRed7, palGreen9, palPurple8, palOrangeA, palBlackB, palOrange6}

// MapFlashColour is kFlashColours (0x004FDC5C).
func MapFlashColour(c uint8) uint8 { return mapFlashColour(c) }

// The route list's special entries (_routeToObjectIdMap).
const (
	MapRouteAircraft = 0xFE
	MapRouteShips    = 0xFD
	MapRouteEnd      = 0xFF
)

// MapColours is the window's colour tables, statics upstream that open
// refills (assignIndustryColours, assignRouteColours) and that persist
// between opens: _assignedIndustryColours (0x00F253CE, an index into
// kIndustryColours per industry object), _routeToObjectIdMap (0x00F253DF: a
// track id, a road id | 0x80, then aircraft 0xFE, ships 0xFD and the end
// 0xFF), _routeColours (0x00F253F2), _trackColours (0x00F25404) and
// _roadColours (0x00F2540C).
type MapColours struct {
	Industry        [16]uint8
	RouteToObjectID [19]uint8
	Route           [19]uint8
	Track, Road     [8]uint8
}

// MapFill is what setMapPixels reads besides the map: the objects, a G1
// element's bytes (getG1Element(image)->offset[i]), Colours::getShade,
// _flashingItems and the colour tables.
type MapFill struct {
	OM       *objects.ObjectManager
	Pixel    func(image uint32, i int) uint8
	Shade    func(colour uint8, shade int) uint8
	Flashing uint32
	Colours  *MapColours
}

// checkIndustryColours is checkIndustryColours (0x0046D0C3): the colours of
// kIndustryColours within 2 of colour taken out of the mask.
func checkIndustryColours(colour uint8, mask uint32) uint32 {
	for i, ic := range MapIndustryColours {
		diff := int(ic) - int(colour)
		if ic < colour {
			diff = -diff
		}
		if diff <= 2 {
			mask &^= 1 << i
		}
	}
	return mask
}

// mapSurfaceColoursMask is both assign functions' start: the colours too close
// to water's and every loaded land object's two map pixels left out.
func mapSurfaceColoursMask(om *objects.ObjectManager, pixel func(uint32, int) uint8) uint32 {
	avail := uint32(0x7FFFFFFF)
	if om.WaterObj != nil {
		img := om.WaterObj.ImageOffset + 40 // WaterObject::mapPixelImage
		avail = checkIndustryColours(pixel(img, 0), avail)
		avail = checkIndustryColours(pixel(img, 1), avail)
	}
	for i := range om.LandObjects { // ObjectManager::forEachLoaded<LandObject>
		if lo := om.GetLandObjectByIndex(i); lo != nil {
			avail = checkIndustryColours(pixel(lo.MapPixelImage, 0), avail)
			avail = checkIndustryColours(pixel(lo.MapPixelImage, 1), avail)
		}
	}
	return avail
}

// AssignIndustryColours is assignIndustryColours (0x0046CFF0): each loaded
// industry object its preferred map colour (IndustryObject::mapColour) while
// still free — not near water, land, buildings (mutedDarkRed2), roads
// (black2) or black0 — then the rest the lowest free one.
func (c *MapColours) AssignIndustryColours(om *objects.ObjectManager, pixel func(uint32, int) uint8) {
	avail := mapSurfaceColoursMask(om, pixel)
	avail = checkIndustryColours(palMutedDarkRed2, avail)
	avail = checkIndustryColours(palBlack2, avail)
	avail = checkIndustryColours(palBlack0, avail)
	for i := range c.Industry {
		c.Industry[i] = 0xFF
	}
	n := min(len(om.IndustryObjects), len(c.Industry))
	for i := range n { // forEachLoaded<IndustryObject>: the preferred colour
		if io := om.GetIndustryObjectByIndex(i); io != nil && io.MapColour < 32 && avail&(1<<io.MapColour) != 0 {
			c.Industry[i] = io.MapColour
			avail &^= 1 << io.MapColour
		}
	}
	for i := range n { // then an alternative
		if io := om.GetIndustryObjectByIndex(i); io == nil || c.Industry[i] != 0xFF {
			continue
		}
		free := max(0, bitScanForward(avail))
		avail &^= 1 << free
		c.Industry[i] = uint8(free)
	}
}

// AssignRouteColours is assignRouteColours (0x0046CED0): the controlling
// company's available rail tracks (companyGetAvailableRailTracks) and roads
// (companyGetAvailableRoads) each the lowest free colour — not near water,
// land, buildings, stations (orange8), aircraft (pink9), ships (blue9),
// black0 or blackB — in the route list and the track / road colour tables;
// then aircraft (0xD3) and ships (0x8B), and the list's end.
func (m *Map) AssignRouteColours(c *MapColours, om *objects.ObjectManager, pixel func(uint32, int) uint8) {
	avail := mapSurfaceColoursMask(om, pixel)
	for _, col := range []uint8{palMutedDarkRed2, palOrange8, palPink9, palBlue9, palBlack0, palBlackB} {
		avail = checkIndustryColours(col, avail)
	}
	tracks := m.companyAvailableRailTracks(m.ControllingCompany, om)
	roads := m.companyAvailableRoads(m.ControllingCompany, om)
	i := 0
	assign := func(id uint8) {
		if i+2 >= len(c.RouteToObjectID) { // upstream's 19 entries: 16 types, aircraft, ships, the end
			return
		}
		c.RouteToObjectID[i] = id
		free := max(0, bitScanForward(avail))
		avail &^= 1 << free
		col := MapIndustryColours[free]
		c.Route[i] = col
		if id&(1<<7) != 0 {
			if int(id&^(1<<7)) < len(c.Road) {
				c.Road[id&^(1<<7)] = col
			}
		} else if int(id) < len(c.Track) {
			c.Track[id] = col
		}
		i++
	}
	for _, t := range tracks {
		assign(t)
	}
	for _, r := range roads {
		assign(r)
	}
	c.RouteToObjectID[i] = MapRouteAircraft
	c.RouteToObjectID[i+1] = MapRouteShips
	c.RouteToObjectID[i+2] = MapRouteEnd
	c.Route[i] = 0xD3
	c.Route[i+1] = 0x8B
}

// mapSurfacePixels is the surface's two colours as the routes and ownership
// tabs read them (and the first for the vehicles and industries tabs): the
// land object's map pixels 0 and 1 (not by height), or water's.
func mapSurfacePixels(t *Tile, f *MapFill) (uint8, uint8) {
	if t.WaterLevel == 0 {
		if lo := f.OM.GetLandObjectByIndex(int(t.TerrainIndex)); lo != nil {
			return f.Pixel(lo.MapPixelImage, 0), f.Pixel(lo.MapPixelImage, 1)
		}
		return 0, 0
	}
	if f.OM.WaterObj != nil {
		img := f.OM.WaterObj.ImageOffset + 40
		return f.Pixel(img, 0), f.Pixel(img, 1)
	}
	return 0, 0
}

// MapPixelsTab is the body of setMapPixelsVehicles, setMapPixelsIndustries,
// setMapPixelsRoutes or setMapPixelsOwnership (by tab) for the tile at
// (tx, ty): its two colours and their flashing pair. ok is false on the
// map's edge, which upstream leaves untouched.
func (m *Map) MapPixelsTab(tab, tx, ty int, f *MapFill) (c0, c1, f0, f1 uint8, ok bool) {
	if tx <= 0 || ty <= 0 || tx >= m.Width-1 || ty >= m.Height-1 || f.OM == nil {
		return 0, 0, 0, 0, false
	}
	t := m.TileAt(tx, ty)
	if t == nil {
		return 0, 0, 0, 0, false
	}
	all := func(c uint8) { c0, c1, f0, f1 = c, c, c, c }
	haveTrackOrRoad := false // ch
	for _, r := range t.Order {
		ghost, ai := r.Flags&elemFlagGhost != 0, r.Flags&elemFlagAiAllocated != 0
		switch tab {
		case MapTabVehicles: // setMapPixelsVehicles (0x0046C873)
			switch r.Kind {
			case scenario.KindSurface:
				p, _ := mapSurfacePixels(t, f)
				all(p)
			case scenario.KindTrack, scenario.KindStation, scenario.KindRoad:
				if !ghost && !ai {
					all(palBlack2)
				}
			case scenario.KindBuilding, scenario.KindIndustry:
				if !ghost {
					all(palMutedDarkRed2)
				}
			}
		case MapTabIndustries: // setMapPixelsIndustries (0x0046C9A8)
			switch r.Kind {
			case scenario.KindSurface:
				p, _ := mapSurfacePixels(t, f)
				all(p)
				if t.IsIndustrial {
					if ind := m.industryByID(t.Variation); ind != nil && int(ind.ObjectID) < len(f.Colours.Industry) {
						c0 = MapIndustryColours[f.Colours.Industry[ind.ObjectID]%31]
						f0 = c0
						if f.Flashing&(1<<ind.ObjectID) != 0 {
							f0 = palBlack0
						}
					}
				}
			case scenario.KindBuilding:
				if !ghost { // vanilla omitted the ghost check
					all(palMutedDarkRed2)
				}
			case scenario.KindIndustry:
				if ghost || int(r.Index) >= len(t.Industries) {
					continue
				}
				if ind := m.industryByID(t.Industries[r.Index].IndustryID); ind != nil && int(ind.ObjectID) < len(f.Colours.Industry) {
					all(MapIndustryColours[f.Colours.Industry[ind.ObjectID]%31])
					if f.Flashing&(1<<ind.ObjectID) != 0 {
						f0, f1 = palBlack0, palBlack0
					}
				}
			case scenario.KindTrack, scenario.KindStation, scenario.KindRoad:
				if !ghost && !ai {
					all(palBlack2)
				}
			}
		case MapTabRoutes: // setMapPixelsRoutes (0x0046CB68)
			switch r.Kind {
			case scenario.KindSurface:
				p0, p1 := mapSurfacePixels(t, f)
				c0, f0 = p0, p0
				if !haveTrackOrRoad {
					c1, f1 = p1, p1
				}
			case scenario.KindBuilding, scenario.KindIndustry:
				if !ghost {
					all(palMutedDarkRed2)
				}
			case scenario.KindTrack:
				if ghost || ai || int(r.Index) >= len(t.Tracks) {
					continue
				}
				id := t.Tracks[r.Index].TrackObjectID
				c0 = f.Colours.Track[id&7]
				f0 = c0
				if first := bitScanForward(f.Flashing); first != -1 && first < len(f.Colours.RouteToObjectID) && f.Colours.RouteToObjectID[first] == id {
					f0 = mapFlashColour(f0)
				}
				f1, c1 = f0, c0
				haveTrackOrRoad = true
			case scenario.KindStation:
				if !ghost && !ai {
					all(palOrange8)
				}
			case scenario.KindRoad:
				if ghost || ai || int(r.Index) >= len(t.Roads) {
					continue
				}
				id := t.Roads[r.Index].RoadObjectID
				c0 = f.Colours.Road[id&7]
				f0 = c0
				if first := bitScanForward(f.Flashing); first != -1 && first < len(f.Colours.RouteToObjectID) && f.Colours.RouteToObjectID[first] == id|1<<7 {
					f0 = mapFlashColour(f0)
				}
				c1, f1 = c0, f0
				haveTrackOrRoad = true
			}
		case MapTabOwnership: // setMapPixelsOwnership (0x0046CD31)
			switch r.Kind {
			case scenario.KindSurface:
				p0, p1 := mapSurfacePixels(t, f)
				c0, f0 = p0, p0
				if !haveTrackOrRoad {
					c1, f1 = p1, p1
				}
			case scenario.KindTrack, scenario.KindStation, scenario.KindRoad:
				if ghost || ai {
					continue
				}
				owner := uint8(companyIDNull)
				switch {
				case r.Kind == scenario.KindStation && int(r.Index) < len(t.Stations):
					if id := int(t.Stations[r.Index].StationID); id < len(m.StationNames) {
						owner = m.StationNames[id].Owner
					}
				case r.Kind == scenario.KindTrack && int(r.Index) < len(t.Tracks):
					owner = t.Tracks[r.Index].Owner
				case r.Kind == scenario.KindRoad && int(r.Index) < len(t.Roads):
					owner = t.Roads[r.Index].Owner
				}
				if owner != companyIDNeutral {
					var colour uint8
					if int(owner) < len(m.CompanyColours) {
						colour = m.CompanyColours[owner] // CompanyManager::getCompanyColour
					}
					all(f.Shade(colour, 5))
					if f.Flashing&(1<<owner) != 0 {
						f0 = mapFlashColour(c0)
						f1 = f0
					}
					haveTrackOrRoad = true
					continue
				}
				all(palBlack1) // [[fallthrough]] to the building case: not a ghost here
			case scenario.KindBuilding, scenario.KindIndustry:
				if !ghost { // vanilla omitted the ghost check
					all(palBlack1)
				}
			}
		default:
			return 0, 0, 0, 0, false
		}
	}
	return c0, c1, f0, f1, true
}

// MapMark is one of drawScroll's vehicle or route marks on the map buffer:
// a pixel (drawVehicleOnMap's fillRect) or a line (drawRouteLine's
// drawLine) from (X0, Y0) to (X1, Y1).
type MapMark struct {
	Line           bool
	X0, Y0, X1, Y1 int
	Colour         uint8
}

// MapVehicles is countVehiclesOnMap (0x0046BFAD, on the vehicles tab) and
// drawVehiclesOnMap (0x0046BE6E): for every train in VehicleList order that
// is not a ghost and placed, each car's components as a pixel in
// getVehicleColour's colour, then on the routes tab its route
// (drawRoutesOnMap). counted is whether counts was recounted. tab is the
// window's, flashing _flashingItems, frame mapFrameNumber, rotation the
// view's; shade is Colours::getShade.
func (m *Map) MapVehicles(tab int, c *MapColours, flashing uint32, frame uint16, rotation uint8, shade func(uint8, int) uint8) (marks []MapMark, counts [6]uint16, counted bool) {
	trains, _ := m.Trains()
	placed := func(tr Train) (*scenario.VehicleEntity, bool) {
		h := &m.Entities[tr.Head]
		return h, h.Flags38&flags38IsGhost == 0 && h.PosX != locationNull
	}
	if tab == MapTabVehicles { // countVehiclesOnMap
		counted = true
		for _, tr := range trains {
			if h, ok := placed(tr); ok && h.VehicleType < 6 {
				counts[h.VehicleType]++
			}
		}
	}
	flash := mapFlashColour
	for _, tr := range trains {
		h, ok := placed(tr)
		if !ok {
			continue
		}
		for _, car := range tr.Cars {
			if len(car) == 0 {
				continue
			}
			colour := uint8(palBlackB) // getVehicleColour (0x0046C426)
			if tab == MapTabOwnership || tab == MapTabVehicles {
				owner := m.Entities[car[0].Front].Owner // car.front->owner
				var cc uint8
				if int(owner) < len(m.CompanyColours) {
					cc = m.CompanyColours[owner]
				}
				colour = shade(cc, 7)
				if tab == MapTabVehicles && h.VehicleType < 6 {
					colour = MapVehicleTypeColours[h.VehicleType]
				}
				if (tab == MapTabOwnership && owner < 32 && flashing&(1<<owner) != 0) ||
					(tab == MapTabVehicles && h.VehicleType < 32 && flashing&(1<<h.VehicleType) != 0) {
					if frame&(1<<2) != 0 {
						colour = flash(colour)
					}
				}
			}
			for _, comp := range car { // applyToComponents: front, back, body
				for _, id := range [3]int{comp.Front, comp.Back, comp.Body} {
					e := &m.Entities[id]
					if e.PosX == locationNull { // drawVehicleOnMap
						continue
					}
					x, y := LocationToMapWindowPos(int32(e.PosX), int32(e.PosY), rotation)
					marks = append(marks, MapMark{X0: x, Y0: y, X1: x, Y1: y, Colour: colour})
				}
			}
		}
		if tab == MapTabRoutes {
			marks = m.mapRoute(marks, h, c, flashing, frame, rotation)
		}
	}
	return marks, counts, counted
}

// mapRoute is drawRoutesOnMap (0x0046C18D) with getRouteColour: an aircraft's
// (colour 211) or ship's (139) orders' stations joined by lines, the last
// back to the first; flashing while its key row is the first hovered one.
func (m *Map) mapRoute(marks []MapMark, h *scenario.VehicleEntity, c *MapColours, flashing uint32, frame uint16, rotation uint8) []MapMark {
	var colour, special uint8
	switch h.VehicleType {
	case 4: // VehicleType::aircraft
		colour, special = 211, MapRouteAircraft
	case vehicleTypeShip:
		colour, special = 139, MapRouteShips
	default:
		return marks
	}
	if i := bitScanForward(flashing); i != -1 && i < len(c.RouteToObjectID) && c.RouteToObjectID[i] == special && frame&(1<<2) != 0 {
		colour = mapFlashColour(colour)
	}
	var start, end [2]int
	haveStart, haveEnd := false, false
	for _, o := range m.OrderRing(h.OrderTableOffset, 0) { // OrderRingView(orderTableOffset)
		if !o.HasFlags(OrderHasStation) {
			continue
		}
		st := o.Station()
		if st >= len(m.StationNames) {
			continue
		}
		s := &m.StationNames[st]
		x, y := LocationToMapWindowPos(int32(s.X), int32(s.Y), rotation) // drawRouteLine (0x0046C294)
		if haveEnd {
			marks = append(marks, MapMark{Line: true, X0: end[0], Y0: end[1], X1: x, Y1: y, Colour: colour})
		}
		end, haveEnd = [2]int{x, y}, true
		if !haveStart {
			start, haveStart = end, true
		}
	}
	if !haveStart || !haveEnd {
		return marks
	}
	return append(marks, MapMark{Line: true, X0: start[0], Y0: start[1], X1: end[0], Y1: end[1], Colour: colour})
}

// MapVehicleStatus is formatVehicleString (0x0046D81F): with no key row
// hovered the total ("{INT16} vehicle(s)"), else the hovered type's count
// ("{INT16} train(s)" …), from countVehiclesOnMap's counts.
func MapVehicleStatus(counts [6]uint16, hover uint32) string {
	i := bitScanForward(hover)
	var n uint16
	id := i18n.StatusNumVehiclesPlural
	if i == -1 || i >= 6 {
		for _, c := range counts {
			n += c
		}
		if n == 1 {
			id = i18n.StatusNumVehiclesSingular
		}
	} else {
		n = counts[i]
		id = i18n.NumTrainsPlural + i
		if n == 1 {
			id = i18n.NumTrainsSingular + i
		}
	}
	return strings.Replace(i18n.Label(id), "{INT16}", strconv.Itoa(int(n)), 1)
}

// MapIndustryStatus is formatIndustryString (0x0046D87C): with no key row
// hovered the number of industries; else the hovered object's count in its
// singular or plural name, and in brackets what it requires ("Require …"
// and " to produce …") or produces ("Produce …").
func (m *Map) MapIndustryStatus(hover uint32, om *objects.ObjectManager) string {
	i := bitScanForward(hover)
	if i == -1 {
		n := 0
		for k := range m.Industries {
			if m.Industries[k].Name != 0xFFFF {
				n++
			}
		}
		id := i18n.StatusNumIndustriesPlural
		if n == 1 {
			id = i18n.StatusNumIndustriesSingular
		}
		return strings.Replace(i18n.Label(id), "{INT16}", strconv.Itoa(n), 1)
	}
	n := 0
	for k := range m.Industries {
		if m.Industries[k].Name != 0xFFFF && int(m.Industries[k].ObjectID) == i {
			n++
		}
	}
	io := om.GetIndustryObjectByIndex(i)
	if io == nil {
		return ""
	}
	raw := io.RawStrings[7] // namePlural
	if n == 1 {
		raw = io.RawStrings[6] // nameSingular
	}
	var sb strings.Builder
	for _, c := range raw { // formatString with the count
		switch {
		case c == 0x7F: // int16_grouped
			sb.WriteString(groupedInt(int32(int16(n))))
		case c >= 0x20 && c < 0x7F:
			sb.WriteByte(c)
		}
	}
	s := sb.String() + " ("
	produced, required := om.IndustryCargoIDs(io)
	produces := produced[0] != 0xFF || produced[1] != 0xFF
	if required[0] != 0xFF || required[1] != 0xFF || required[2] != 0xFF { // requiresCargo
		s += i18n.Label(i18n.IndustryRequire) + cargoList(om, required[:], io.Flags&objects.IndustryFlagRequiresAllCargo == 0)
		if produces {
			s += i18n.Label(i18n.CargoToProduce) + cargoList(om, produced[:], false)
		}
	} else if produces {
		s += i18n.Label(i18n.IndustryProduce) + cargoList(om, produced[:], false)
	}
	return s + ")"
}
