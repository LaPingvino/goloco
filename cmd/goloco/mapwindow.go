package main

import (
	"image"

	"github.com/LaPingvino/goloco/pkg/gfx8"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// openMap is MapWindow::open (0x0046B490; PORT-PLANS A, steps 1a, 1c, 1d):
// the one map window, brought to the front when open; the skin's
// windowTitlebarColour and windowMapColour; the buffer cleared (clearMap),
// filled as it updates; centred on the main view (centerOnViewPoint); the
// overall tab, town names shown, no row hovered; assignIndustryColours and
// assignRouteColours; mapFrameNumber 0.
func (g *Game) openMap() {
	if w := g.wins.Find(ui.WindowTypeMap, ui.AnyNumber); w != nil {
		g.wins.BringToFront(w)
		return
	}
	var tc, mc uint8 = 0, 1
	if sk := g.skin(); sk != nil {
		tc, mc = sk.WindowTitlebarColour, sk.WindowMapColour
	}
	w := ui.NewMapWindow(sim.NewMapBuffer(), tc, mc) // clearMap
	w.Rotation = g.w.Map.CurrentRotation
	w.MainView = g.mapMainView
	w.OnCentre = g.w.CentreOnMapView // moveMainViewToMapView
	if sk := g.skin(); sk != nil {
		w.SkinImage = int(sk.ImageOffset)
	}
	m := &g.w.Map
	if pc := int(m.ControllingCompany); pc < len(m.CompanyColourSchemes) { // getPlayerCompany()->mainColours.primary
		w.VehicleColour = m.CompanyColourSchemes[pc].Main[0]
	}
	w.Key, w.Status, w.Vehicles, w.Towns = g.mapKey, g.mapStatus, g.mapVehicles, g.mapTowns
	g.showAuto(w, 0, 0)
	w.CentreOnViewPoint()
	if om := g.r.ObjMgr; om != nil {
		g.mapColours.AssignIndustryColours(om, g.mapPixel)
		m.AssignRouteColours(&g.mapColours, om, g.mapPixel)
	}
}

// mapPixel is getG1Element(image)->offset[i]: a G1 element's byte (0 out of
// range).
func (g *Game) mapPixel(img uint32, i int) uint8 {
	g1 := g.r.G1
	if g1 == nil || int(img) >= len(g1.Elements) || i < 0 || i >= len(g1.Elements[img].Data) {
		return 0
	}
	return g1.Elements[img].Data[i]
}

// mapShade is Colours::getShade(colour, shade).
func (g *Game) mapShade(colour uint8, shade int) uint8 {
	if g.r.G1 == nil {
		return 0
	}
	cm, err := g.r.G1.GetPaletteMap(int(colour))
	if err != nil {
		return 0
	}
	return gfx8.Shade(cm, shade)
}

// mapKey is the tab's graph key rows (drawGraphKey*); the companies' is
// drawGraphKeyCompanies (0x0046D6E1): each company (CompanyManager::
// companies(): the non-empty ones) its primary colour's shade 6 and name.
func (g *Game) mapKey(tab int) []ui.MapKeyRow {
	om := g.r.ObjMgr
	switch tab {
	case sim.MapTabOverall:
		return ui.MapOverallKey()
	case sim.MapTabVehicles:
		return ui.MapVehiclesKey()
	case sim.MapTabIndustries:
		if om != nil {
			return ui.MapIndustriesKey(om, &g.mapColours)
		}
	case sim.MapTabRoutes:
		if om != nil {
			return ui.MapRoutesKey(om, &g.mapColours)
		}
	case sim.MapTabOwnership:
		m := &g.w.Map
		var rows []ui.MapKeyRow
		for id := range min(len(m.CompanyNames), 15) {
			if m.CompanyEmpty(id) || id >= len(m.CompanyColourSchemes) {
				continue
			}
			rows = append(rows, ui.MapKeyRow{Bit: id, Colour: g.mapShade(m.CompanyColourSchemes[id].Main[0], 6), Name: m.CompanyName(id)})
		}
		return rows
	}
	return nil
}

// mapStatus is draw's status bar string: formatVehicleString (0x0046D81F) on
// the vehicles tab, formatIndustryString (0x0046D87C) on the industries
// tab, empty otherwise.
func (g *Game) mapStatus(tab int, hover uint32) string {
	switch tab {
	case sim.MapTabVehicles:
		return sim.MapVehicleStatus(g.mapVehicleCounts, hover)
	case sim.MapTabIndustries:
		if g.r.ObjMgr != nil {
			return g.w.Map.MapIndustryStatus(hover, g.r.ObjMgr)
		}
	}
	return ""
}

// mapVehicles is drawScroll's countVehiclesOnMap (the vehicles tab: the
// status bar's counts) and drawVehiclesOnMap, at the view's rotation.
func (g *Game) mapVehicles(tab int, flashing uint32, frame uint16) []sim.MapMark {
	marks, counts, counted := g.w.Map.MapVehicles(tab, &g.mapColours, flashing, frame, g.w.Rotation(), g.mapShade)
	if counted {
		g.mapVehicleCounts = counts
	}
	return marks
}

// mapTowns is TownManager::towns() for drawTownNames: each town's position
// and name.
func (g *Game) mapTowns() []ui.MapTownName {
	m := &g.w.Map
	out := make([]ui.MapTownName, 0, len(m.Towns))
	for i, t := range m.Towns {
		out = append(out, ui.MapTownName{X: int32(t.X), Y: int32(t.Y), Name: m.TownName(t.Name, i)})
	}
	return out
}

// mapMainView is the main window's viewports[0] as the map window reads it:
// viewX / viewY and viewWidth / viewHeight (viewport coordinates), and
// getCurrentRotation.
func (g *Game) mapMainView() ui.MapMainView {
	v := g.w.MainViewport().View
	return ui.MapMainView{X: v.X, Y: v.Y, W: v.W << v.Zoom, H: v.H << v.Zoom, Rotation: g.w.Rotation(), OK: true}
}

// rotateView is a rotation of the main view by its callers — Shortcuts.cpp
// rotateView (l. 101), MouseInput.cpp's wheel with the right button held
// (l. 1916), ToolbarTop.cpp rotateMenuDropdown (l. 645): viewportRotateRight
// / Left with updateLabels (World.RotateView), then
// MapWindow::centerOnViewPoint.
func (g *Game) rotateView(right bool) {
	g.w.RotateView(right)
	if w, ok := g.wins.Find(ui.WindowTypeMap, ui.AnyNumber).(*ui.MapWindow); ok {
		w.CentreOnViewPoint()
	}
}

// mapChatMenu is TimePanel.cpp mapMouseDown (0x0043A67F) outside a network
// game: a one-item menu below the button, the map (its icon by the view's
// rotation); choosing it is mapDropdown's MapWindow::open.
func (g *Game) mapChatMenu(x, y, w, h int) {
	skin := 0
	if sk := g.skin(); sk != nil {
		skin = int(sk.ImageOffset)
	}
	items := []ui.DropdownItem{{MenuSprite: true, Sprite: skin + objects.ISToolbarMenuMapNorth + int(g.w.Map.CurrentRotation&3),
		Text: i18n.Label(i18n.MenuMap), Action: g.openMap}}
	d := ui.NewDropdownShowText(x, y, w, h, ui.DropdownColour, items, 25, g.sw, g.sh)
	d.OpenedByPress(image.Rect(x, y, x+w, y+h))
	d.Highlighted = 0
	g.dropdown = d
}

// updateMap is MapWindow's onUpdate (0x0046BA5B, each frame): mapFrameNumber
// counts, a new view rotation clears the buffer (var_846), then 80 rows are
// filled for the tab (setMapPixels, _flashingItems = var_854), then the
// window's part (frameNo, setHoverItemTab: ui MapWindow.Update).
func (g *Game) updateMap(w *ui.MapWindow, mx, my int, onTop, rightHeld bool) {
	w.FrameNo++
	m := &g.w.Map
	if m.CurrentRotation != w.Rotation {
		w.Rotation = m.CurrentRotation
		w.Buf.Clear()
	}
	w.Flashing = w.Hover
	f := &sim.MapFill{OM: g.r.ObjMgr, Pixel: g.mapPixel, Shade: g.mapShade, Flashing: w.Flashing, Colours: &g.mapColours}
	for range 80 {
		m.FillMapRow(w.Buf, int(w.Rotation), w.Tab, f)
	}
	w.Update(mx, my, onTop, rightHeld)
}
