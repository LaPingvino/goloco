package ui

import (
	"image"
	"image/color"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/hajimehoshi/ebiten/v2"
)

// The map window's tabs, graph keys and overlays (OpenLoco reference:
// src/OpenLoco/src/Ui/Windows/MapWindow.cpp; PORT-PLANS A step d): drawTabs
// (0x0046D0E0), drawGraphKey* (l. 1337–1526), setHoverItemTab (l. 1027), the
// status bar (draw l. 1688, formatVehicleString / formatIndustryString), and
// drawScroll's vehicles, routes and town names.

// MapKeyRow is a graph key row as drawGraphKey* draws it: its _flashingItems
// bit (the row's index, an industry object's or a company's id), the
// swatch's colour and the name.
type MapKeyRow struct {
	Bit    int
	Colour uint8
	Name   string
}

// MapTownName is a town for drawTownNames: its position (world units) and
// formatted name.
type MapTownName struct {
	X, Y int32
	Name string
}

// mapLegendWidth / mapLegendItemHeight are legendWidth and legendItemHeight.
const (
	mapLegendWidth      = 100
	mapLegendItemHeight = 10
)

// mapLegendLengths is legendLengths: the rows setHoverItemTab measures the key
// by (kOverallGraphKeySize, the vehicle types, the industry objects, —, the
// companies); the routes tab counts its list instead.
var mapLegendLengths = [5]int{6, 6, 16, 0, 15}

// MapOverallKey is drawGraphKeyOverall's rows (0x0046D273): towns,
// industries, roads, railways, stations, vegetation.
func MapOverallKey() []MapKeyRow {
	colours := [6]uint8{0x41, 0x7D, 0x0C, 0x11, 0xBA, 0x64} // mutedDarkRed7, mutedPurple7, black2, black7, orange8, green6
	names := [6]int{i18n.MapKeyTowns, i18n.MapKeyIndustries, i18n.MapKeyRoads, i18n.MapKeyRailways, i18n.MapKeyStations, i18n.MapKeyVegetation}
	rows := make([]MapKeyRow, 6)
	for i := range rows {
		rows[i] = MapKeyRow{Bit: i, Colour: colours[i], Name: i18n.Label(names[i])}
	}
	return rows
}

// MapVehiclesKey is drawGraphKeyVehicles' rows (0x0046D379): the six vehicle
// types (forbid_trains …) in vehicleTypeColours.
func MapVehiclesKey() []MapKeyRow {
	rows := make([]MapKeyRow, 6)
	for i := range rows {
		rows[i] = MapKeyRow{Bit: i, Colour: sim.MapVehicleTypeColours[i], Name: i18n.Label(i18n.ForbidTrains + i)}
	}
	return rows
}

// MapIndustriesKey is drawGraphKeyIndustries' rows (0x0046D47F): each loaded
// industry object, its assigned colour and name.
func MapIndustriesKey(om *objects.ObjectManager, c *sim.MapColours) []MapKeyRow {
	var rows []MapKeyRow
	for i := range c.Industry {
		io := om.GetIndustryObjectByIndex(i)
		if io == nil {
			continue
		}
		rows = append(rows, MapKeyRow{Bit: i, Colour: sim.MapIndustryColours[c.Industry[i]%31], Name: io.Name})
	}
	return rows
}

// MapRoutesKey is drawGraphKeyRoutes' rows (0x0046D5A4): the route list up to
// its end — a track's or road's (id | 0x80) object name, Aircraft Routes
// (0xFE), Ship Routes (0xFD) — in _routeColours.
func MapRoutesKey(om *objects.ObjectManager, c *sim.MapColours) []MapKeyRow {
	var rows []MapKeyRow
	for i := 0; i < len(c.RouteToObjectID) && c.RouteToObjectID[i] != sim.MapRouteEnd; i++ {
		id := c.RouteToObjectID[i]
		name := i18n.Label(i18n.MapRoutesAircraft)
		if id != sim.MapRouteAircraft {
			name = i18n.Label(i18n.MapRoutesShips)
			if id != sim.MapRouteShips {
				if id&(1<<7) != 0 {
					if ro := om.GetRoadObjectByIndex(int(id &^ (1 << 7))); ro != nil {
						name = ro.Name
					}
				} else if to := om.GetTrackObjectByIndex(int(id)); to != nil {
					name = to.Name
				}
			}
		}
		rows = append(rows, MapKeyRow{Bit: i, Colour: c.Route[i], Name: name})
	}
	return rows
}

// MapHoverItem is setHoverItemTab's row search (l. 1027) with setHoverItem's
// result (0x0046D34D): the key row under a point (cx, cy) relative to the
// key's top left, as a _flashingItems bit — the first row for a point above
// the key, none past its last row or off its 100 px width. rows are the
// tab's drawn rows (empty industry and company slots left out).
func MapHoverItem(tab, cx, cy int, rows []MapKeyRow) uint32 {
	if cx < 0 || cx > mapLegendWidth {
		return 0
	}
	cy = int(int16(cy))
	if tab != 3 && !(tab >= 0 && tab < len(mapLegendLengths) && cy < mapLegendLengths[tab]*mapLegendItemHeight) {
		return 0
	}
	y := cy
	for _, r := range rows {
		y -= mapLegendItemHeight
		if y < 0 {
			return 1 << r.Bit
		}
	}
	return 0
}

// drawMapLine is drawLine (SoftwareDrawingContext.cpp 0x00452DA4) on the
// scroll's target: off the target (Rect::fromLTRB(a, b).intersects, as
// upstream: an unordered rectangle) nothing; else Bresenham, a steep line a
// pixel at a time, others in horizontal runs (drawHorizontalLine: at least a
// pixel).
func (t mapTarget) drawMapLine(ax, ay, bx, by int, colour uint8) {
	if bx <= t.X || by <= t.Y || ax >= t.X+t.W || ay >= t.Y+t.H { // intersects(rt.getUiRect())
		return
	}
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	hline := func(x, y, length int) { // drawHorizontalLine (0x00452D04)
		ox, oy := x-t.X, y-t.Y
		if oy < 0 || oy >= t.H {
			return
		}
		if length == 0 {
			length++
		}
		if ox < 0 {
			length += ox
			if length <= 0 {
				return
			}
			ox = 0
		}
		if ox+length-t.W > 0 {
			length -= ox + length - t.W
			if length <= 0 {
				return
			}
		}
		for i := range length {
			sx, sy := t.OrgX+ox+i, t.OrgY+oy
			if sx >= 0 && sy >= 0 && sx < t.rt.Width && sy < t.rt.Height {
				t.rt.Bits[sy*t.rt.Width+sx] = colour
			}
		}
	}
	steep := abs(ay-by) > abs(ax-bx)
	if steep {
		by, ax = ax, by
		ay, bx = bx, ay
	}
	if ax > bx {
		ax, bx = bx, ax
		by, ay = ay, by
	}
	dx, dy := bx-ax, abs(by-ay)
	e := dx / 2
	ys := -1
	if ay < by {
		ys = 1
	}
	y := ay
	for x, xs, n := ax, ax, 1; x < bx; x, n = x+1, n+1 {
		if steep {
			hline(y, x, 1)
		}
		e -= dy
		if e < 0 {
			if !steep {
				hline(xs, y, n)
			}
			xs, n = x+1, 0
			y += ys
			e += dx
		}
		if x+1 == bx && !steep {
			hline(xs, y, n)
		}
	}
}

// mapTabIcons is drawTabs' images (0x0046D0E0) by tab: the map by the view's
// rotation; the train animated (frameNo / 2, 8 frames) while its tab is
// open; the industries; the routes animated (frameNo / 16, 4 frames); the
// companies.
func (w *MapWindow) mapTabIcons() [5]int {
	skin := w.SkinImage
	vehicle := skin + objects.ISVehicleTrainFrame0
	if w.Tab == sim.MapTabVehicles {
		vehicle += int(w.TabFrame/2) % 8
	}
	routes := skin + mapTabRoutesFrame0
	if w.Tab == sim.MapTabRoutes {
		routes += int(w.TabFrame/16) % 4
	}
	return [5]int{skin + objects.ISToolbarMenuMapNorth + int(w.Rotation&3), vehicle,
		skin + objects.ISToolbarMenuIndustries, routes, skin + mapTabCompanies}
}

// InterfaceSkin::ImageIds tab_routes_frame_0 and tab_companies.
const (
	mapTabRoutesFrame0 = 192
	mapTabCompanies    = 174
)

// mapCaptions is prepareDraw's captionText.
var mapCaptions = [5]int{i18n.TitleMap, i18n.TitleMapVehicles, i18n.TitleMapIndustries, i18n.TitleMapRoutes, i18n.TitleMapCompanies}

// drawTabs8 is drawTabs: each tab (leftAlignTabs: 31 apart from x 3) with
// its image, the vehicles' recoloured to the player's colour.
func (w *MapWindow) drawTabs8(rt *gfx8.RenderTarget, sec []uint8) {
	g1 := globalG1Font.g1
	icons := w.mapTabIcons()
	for i, x := range tabLayout(5, nil) {
		var iconMap []uint8
		if i == sim.MapTabVehicles { // Gfx::recolour(imageId, colour)
			iconMap, _ = g1.GetPaletteMap(int(w.VehicleColour))
		}
		icon := icons[i]
		if icon >= len(g1.Elements) {
			icon = -1
		}
		drawTab(rt, x, 15, i == w.Tab, sec, icon, iconMap)
	}
}

// drawKeySwatches is the graph key's swatches (drawRect(x, y + 3, 5, 5)),
// hidden while the row is hovered on mapFrameNumber's bit-2 frames; it
// returns the y below the key.
func (w *MapWindow) drawKeySwatches(rt *gfx8.RenderTarget) int {
	x, y := w.W-104, 44
	for _, r := range w.keyRows {
		if w.Hover&(1<<r.Bit) == 0 || w.FrameNo&(1<<2) == 0 {
			gfx8.FillRect(rt, x, y+3, x+4, y+7, r.Colour)
		}
		y += mapLegendItemHeight
	}
	return y
}

// drawKeyText is the graph key's names (small_black_string, white while
// hovered; {SMALLFONT} is Font::regular) clipped to 94 px, and the status
// bar's string (black_stringid at the label's left − 1, top − 1).
func (w *MapWindow) drawKeyText(screen *ebiten.Image) {
	pal := globalG1Font.g1.GetPalette()
	var black color.Color = pal[paletteBlack0]
	white, _ := TextColours(2)
	x, y := w.X+w.W-104, w.Y+44
	for _, r := range w.keyRows {
		c := black
		if w.Hover&(1<<r.Bit) != 0 {
			c = white
		}
		DrawTextFont(screen, ClipString(r.Name, G1FontMediumNormal, 94), G1FontMediumNormal, x+6, y, c)
		y += mapLegendItemHeight
	}
	if w.Status != nil {
		if s := w.Status(w.Tab, w.Hover); s != "" {
			width := (w.W - 14) - 3 + 1 // statusBar: left 3, right width − 14
			// drawStringLeftClipped in TextRenderer's default font (medium bold)
			DrawTextFont(screen, ClipString(s, G1FontMediumBold, width), G1FontMediumBold, w.X+3-1, w.Y+(w.H-12)-1, black)
		}
	}
}

// drawTownNames is drawTownNames (0x0046C481): each town's name in the
// small font, purple outlined, centred on its map point and 3 px up, clipped
// to the scroll's crop (Org…, W × H on the screen).
func (w *MapWindow) drawTownNames(screen *ebiten.Image, t mapTarget) {
	if w.Towns == nil {
		return
	}
	clip, ok := screen.SubImage(image.Rect(w.X+t.OrgX, w.Y+t.OrgY, w.X+t.OrgX+t.W, w.Y+t.OrgY+t.H)).(*ebiten.Image)
	if !ok {
		return
	}
	body, outline := TextColours(5) // AdvancedColour(Colour::purple).outline()
	for _, tn := range w.Towns() {
		px, py := sim.LocationToMapWindowPos(tn.X, tn.Y, w.Rotation)
		px -= MeasureTextFont(tn.Name, G1FontSmall) / 2
		py -= 3
		DrawTextOutlinedFont(clip, tn.Name, G1FontSmall, w.X+t.OrgX+px-t.X, w.Y+t.OrgY+py-t.Y, body, outline)
	}
}
