package main

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"os"
	"strings"
)

// listwindows.go — the right-toolbar list windows: the ported VehicleList,
// StationList and TownList (pkg/ui/townlist.go's table window).

// openVehiclesWindow lists the scenario's vehicles.
func (g *Game) openVehiclesWindow() { g.openVehicleList(0) }

// openVehicleList opens the ported VehicleList window on a tab (vehicle type).
func (g *Game) openVehicleList(tab int) {
	company, _ := g.w.ControllingCompany()
	if os.Getenv("GOLOCO_DIAG_VEHLIST") != "" { // diagnostics: the first company owning this type
		for _, v := range g.w.VehicleListRows() {
			if v.Type == tab {
				company = v.Owner
				break
			}
		}
	}
	g.openVehicleListFor(company, tab)
}

// openVehicleListFor is VehicleList::open(company, type).
func (g *Game) openVehicleListFor(company, tab int) {
	cname := g.w.CompanyName(company)
	var rows []ui.VehicleRow
	for _, v := range g.w.VehicleListRows() {
		if v.Type == tab && v.Owner == company {
			var cars []ui.TrainCarDraw
			for _, c := range g.w.TrainCars(v.HeadID) {
				cars = append(cars, ui.TrainCarDraw(c))
			}
			rows = append(rows, ui.VehicleRow{Name: v.Name, Profit: v.Profit, AgeYears: v.AgeYears, Reliability: v.Reliability,
				RecentProfit: v.RecentProfit, DayCreated: v.DayCreated, Cars: cars, Status: v.Status, StatusRed: v.StatusRed, Cargos: v.Cargos})
		}
	}
	var tc, wc uint8 = 0, 1
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		tc, wc = g.r.ObjMgr.InterfaceSkin.WindowTitlebarColour, g.r.ObjMgr.InterfaceSkin.WindowPlayerColor
	}
	// The window's owner is the company, so WindowManager draws its primary
	// colour as the company's colour (CompanyManager::getCompanyColour).
	if company < len(g.w.CompanyColours) {
		tc = g.w.CompanyColours[company]
	}
	// The cargo filter's list: every loaded cargo object, by its slot.
	var cargos []ui.VehicleListCargo
	if g.objMgr != nil {
		for id, c := range g.objMgr.CargoObjects {
			if c != nil {
				cargos = append(cargos, ui.VehicleListCargo{ID: id, Name: c.Name, Sprite: int(c.ImageOffset)}) // unitInlineSprite
			}
		}
	}
	w := ui.NewVehicleList(g.sw, tab, rows, tc, wc, cargos)
	w.OpenDropdown = func(d *ui.DropdownMenu) { g.dropdown = d }
	w.Face, w.FaceColour = g.companyFace(company)
	// populateCompanySelect's companies: every non-empty slot.
	for id := range g.w.CompanyNames {
		if g.w.CompanyNames[id] == 0 || id >= len(g.w.CompanyFaces) { // StringIds::empty: no company
			continue
		}
		face, col := g.companyFace(id)
		w.Companies = append(w.Companies, ui.CompanyChoice{ID: id, Name: g.w.CompanyName(id), Performance: int(g.w.CompanyFaces[id].Performance), Face: face, Colour: col})
	}
	w.OwnerID = company
	w.ScreenW, w.ScreenH = g.sw, g.sh
	w.OnCompany = func(id int) { g.openVehicleListFor(id, tab) } // onCompanyDropdown: the window turns to that company
	w.OpenDropdown = func(d *ui.DropdownMenu) { g.dropdown = d }
	if cname != "" {
		w.Caption = cname + " - " + w.Caption // stringid_trains etc.
	}
	w.OnClose = func() { g.townList = nil }
	w.OnTab = func(t int) { g.openVehicleListFor(company, t) }
	g.townList = w
}

// openStationsWindow lists every placed/loaded station on the map.
func (g *Game) openStationsWindow() {
	// The ported StationList window (pkg/ui/townlist.go, shared table layout).
	// StationList::open(getControllingId()): the controlling company's stations.
	company, cname := g.w.ControllingCompany()
	var rows []ui.StationRow
	for _, s := range g.w.StationList() {
		if s.Owner >= 0 && s.Owner != company {
			continue
		}
		rows = append(rows, ui.StationRow{Name: s.Name, Status: s.Status, Waiting: s.Waiting, Accepts: s.Accepts,
			AcceptsKey: g.w.StationAcceptsKey(s.ID)})
	}
	var tc, wc uint8 = 0, 1
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		tc, wc = g.r.ObjMgr.InterfaceSkin.WindowTitlebarColour, g.r.ObjMgr.InterfaceSkin.WindowPlayerColor
	}
	// The window's owner is the company, so WindowManager draws its primary
	// colour as the company's colour (CompanyManager::getCompanyColour).
	if company < len(g.w.CompanyColours) {
		tc = g.w.CompanyColours[company]
	}
	w := ui.NewStationList(g.sw, rows, tc, wc)
	if cname != "" {
		w.Caption = strings.Replace(i18n.Label(i18n.StringidAllStations), "{STRINGID}", cname, 1) // "{STRINGID} - All Stations"
	}
	w.OnClose = func() { g.townList = nil }
	g.townList = w
}

// openTownsWindow lists the scenario's towns with population and size.
func (g *Game) openTownsWindow() {
	// The ported TownList window (pkg/ui/townlist.go).
	infos := g.w.TownList()
	rows := make([]ui.TownRow, len(infos))
	for i, t := range infos {
		rows[i] = ui.TownRow{Name: t.Name, Type: t.Size, Size: t.SizeIndex, Population: int(t.Population), Stations: t.NumStations}
	}
	var tc, wc uint8 = 0, 1
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		tc, wc = g.r.ObjMgr.InterfaceSkin.WindowTitlebarColour, g.r.ObjMgr.InterfaceSkin.WindowColour
	}
	w := ui.NewTownList(g.sw, rows, tc, wc)
	w.OnClose = func() { g.townList = nil }
	g.townList = w
}

// companyFace is the owner's image of a company: its competitor object's
// images[ownerEmotion], recoloured to mainColours.primary (0 if unknown).
func (g *Game) companyFace(company int) (int, uint8) {
	if g.objMgr == nil || company >= len(g.w.CompanyFaces) {
		return 0, 0
	}
	f := g.w.CompanyFaces[company]
	if int(f.Competitor) >= len(g.objMgr.CompetitorObjects) || g.objMgr.CompetitorObjects[f.Competitor] == nil || f.Emotion >= 9 {
		return 0, 0
	}
	var colour uint8
	if company < len(g.w.CompanyColours) {
		colour = g.w.CompanyColours[company]
	}
	return int(g.objMgr.CompetitorObjects[f.Competitor].Images[f.Emotion]), colour
}

// openBuildVehicle is BuildVehicle::open(tab, isTabId = true) for the
// controlling company (5P.f): the opening track tab, the list, at the screen's
// middle (PARTIAL: createWindow's free-spot placement).
func (g *Game) openBuildVehicle(tab int) {
	if g.w == nil || g.r == nil || g.r.ObjMgr == nil {
		return
	}
	m, om := &g.w.Map, g.r.ObjMgr
	company, _ := g.w.ControllingCompany()
	o := sim.DefaultBuildVehicleOptions(uint8(company))
	tabs := m.BuildVehicleTrackTabs(om, uint8(tab), o)
	trackTab := m.BuildVehicleTrackTab(uint8(tab), tabs)
	var rows []ui.BVRow
	if len(tabs) > 0 {
		for _, id := range m.BuildableVehicles(om, uint8(tab), tabs[trackTab], o) {
			v := om.Vehicles[id]
			rows = append(rows, ui.BVRow{ID: id, Vehicle: v, Name: v.DisplayName})
		}
	}
	var tc, wc uint8 = 0, 1
	if om.InterfaceSkin != nil {
		tc, wc = om.InterfaceSkin.WindowTitlebarColour, om.InterfaceSkin.WindowPlayerColor
	}
	var p, s uint8
	if company >= 0 && company < len(m.CompanyColourSchemes) {
		p, s = m.CompanyColourSchemes[company].Main[0], m.CompanyColourSchemes[company].Main[1]
	}
	sw, sh := g.sw, g.sh
	g.buildVehicle = &ui.BuildVehicleWindow{X: (sw - 500) / 2, Y: (sh - 400) / 2, W: 500, H: 400, TitleColour: tc, Colour: wc,
		Tab: tab, TrackTab: trackTab, Disabled: m.BuildVehicleDisabledTabs(uint8(company)), NumTrackTabs: len(tabs),
		Rows: rows, Hover: -1, Primary: p, Secondary: s,
		OnClose: func() { g.buildVehicle = nil }, OnTab: g.openBuildVehicle,
		OnBuy: func(id uint16) { g.buyVehicle(tab, id) }}
}

// buyVehicle is onScrollMouseDown's purchase: the createVehicle command (no
// target vehicle) for the controlling company through doCommand (cash checked,
// paid under VehiclePurchases), then the window refreshed (sub_4B92A5: here
// opened afresh on its tab). PARTIAL: Vehicle::Details::open on the new
// vehicle (the Vehicle window, 5P.g — until then the bought vehicle waits
// unplaced, as upstream's does until placed, and shows in the vehicle list),
// the error window ("Can't build …" with the reason), Config::buildLockedVehicles
// (false).
func (g *Game) buyVehicle(tab int, id uint16) {
	if g.w == nil || g.r == nil || g.r.ObjMgr == nil {
		return
	}
	m := &g.w.Map
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	var head uint16
	cost, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehiclePurchases, func(flags uint8) (int32, bool) {
		return m.CreateVehicle(c, flags, id, false, &head)
	})
	if !ok {
		log.Printf("[Game] can't build vehicle %d", id)
		return
	}
	log.Printf("[Game] bought vehicle %d (head %d) for %d", id, head, cost)
	g.openVehicleDetails(head)
	g.openBuildVehicle(tab)
}

// openVehicleDetails is Vehicle::Details::open (Main::open, then the Details
// tab): 265×177 grown to Details' minimum (192×182 for the controlling
// company's own vehicle, onResize), the window's colours the owner's, its
// caption the vehicle's name. PARTIAL: createWindow's placement (centred
// here), bringing an already open window to the front.
func (g *Game) openVehicleDetails(headID uint16) { g.openVehicleWindow(headID, 1) }

// openVehicleWindow opens (or refreshes) the Vehicle window on a tab: 0 Main,
// 1 Details.
func (g *Game) openVehicleWindow(headID uint16, tab int) {
	if g.w == nil {
		return
	}
	if g.vehicleWin == nil || g.routeHead != headID { // a new window: no order selected (orderTableIndex -1)
		g.routeSel, g.routeHead = -1, headID
	}
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	company, _ := g.w.ControllingCompany()
	own := int(h.Owner) == company
	// Route's onUpdate: its tool is set while the tab shows the company's own
	// vehicle (and no other tool — the pickup — holds the cursor).
	g.routeTool = tab == 4 && own && !g.pickupOn
	title := ""
	for _, r := range m.VehicleListRows() {
		if r.HeadID == headID {
			title = r.Name
		}
	}
	var cars []ui.TrainCarDraw
	var names []string
	for _, c := range g.w.TrainCars(headID) {
		cars = append(cars, ui.TrainCarDraw(c))
		n := ""
		if c.Object != nil {
			n = c.Object.DisplayName
		}
		names = append(names, n)
	}
	var tc, wc uint8 = 0, 1
	if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
		tc, wc = g.r.ObjMgr.InterfaceSkin.WindowTitlebarColour, g.r.ObjMgr.InterfaceSkin.WindowPlayerColor
	}
	if int(h.Owner) < len(g.w.CompanyColours) {
		tc = g.w.CompanyColours[h.Owner]
	}
	ww, wh := ui.VehicleWindowSize[0], ui.VehicleWindowSize[1]
	min := ui.VehicleDetailsMinSize
	if own {
		min = ui.VehicleDetailsMinSizePaint
	}
	ww, wh = max(ww, min[0]), max(wh, min[1])
	if tab == 4 {
		wh = max(wh, ui.VehicleRouteMinHeight) // Route::kMinWindowSize
	}
	if tab == 3 {
		ww, wh = ui.VehicleFinancesSize[0], ui.VehicleFinancesSize[1] // Finances: min = max size
	}
	if tab == 2 {
		ww, wh = max(ui.VehicleWindowSize[0], ui.VehicleCargoMinSize[0]), max(ui.VehicleWindowSize[1], ui.VehicleCargoMinSize[1])
	}
	g.vehicleWin = &ui.VehicleWindow{X: (g.sw - ww) / 2, Y: (g.sh - wh) / 2, W: ww, H: wh, TitleColour: tc, Colour: wc,
		Title: title, VehicleType: int(h.VehicleType), Cars: cars, Names: names,
		State:   ui.VehicleDetailsState{Own: own, Placed: h.TileX != -1, NoCars: len(tr.Cars) == 0},
		OnClose: func() { g.vehicleWin, g.pickupOn, g.routeTool = nil, false, false }, OnPickup: func() { g.onPickup(headID) },
		PickupActive: g.pickupOn && g.pickupHead == headID, PickupImage: g.pickupImage(&h),
		Route: g.routeLines(&h), RouteSel: g.routeSel, Hover: -1, Finances: g.vehicleFinances(tr), Cargo: g.vehicleCargo(tr, cars),
		OnRouteSelect: func(row int) { g.routeSel = row; g.openVehicleWindow(headID, 4) },
		OnRoute:       func(a string) { g.routeAction(headID, a); g.openVehicleWindow(headID, 4) },
		Express:       m.Entities[tr.Veh1].Var48&sim.Flags48ExpressMode != 0,
		Tab:           tab, Main: ui.VehicleMainState{Own: own, Manual: h.VehicleFlags&sim.VehicleFlagManualControl != 0},
		StopStartImage: stopStartImage(h.VehicleFlags),
		OnTab:          func(t int) { g.openVehicleWindow(headID, t) },
		View:           func(w, h int) *gfx8.RenderTarget { return g.w.VehicleView(headID, w, h) },
		OnCentreView:   func() { g.centreViewportOpen(headID) },
		OnStopStart:    func(x, y, w, hh int) { g.stopStartOpen(headID, x, y, w, hh) },
		OnRefit:        func(x, y, w, hh int) { g.refitOpen(headID, x, y, w, hh) }}
	if own {
		g.vehicleWin.OnRename = func() { g.renameVehicle(headID, tc, wc) }
	}
	g.vehicleWin.OnSell = func() { g.sellVehicle(headID) }
	g.vehicleWin.OnDropSell = func(row int) { g.dropCar(headID, row, -2) }
	g.vehicleWin.OnDropMove = func(row, dest int) { g.dropCar(headID, row, dest) }
}

// stopStartImage is Main::prepareDraw's: yellow in manual control, red when
// told to stop, else green.
func stopStartImage(flags uint16) int {
	switch {
	case flags&sim.VehicleFlagManualControl != 0:
		return ui.ImgYellowFlag
	case flags&sim.VehicleFlagCommandStop != 0:
		return ui.ImgRedFlag
	}
	return ui.ImgGreenFlag
}

// stopStartOpen is Main's stopStartOpen: a text dropdown at the button with
// Stop and Start (Manual only with the driver cheat, not ported), the current
// mode marked, the other highlighted; onStopStartDropdown runs
// VehicleChangeRunningMode and the window shows the new flag.
func (g *Game) stopStartOpen(headID uint16, x, y, w, h int) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	head := m.Entities[tr.Head]
	selected := 0 // Stop
	if head.VehicleFlags&sim.VehicleFlagCommandStop == 0 {
		selected = 1 // Start
	}
	items := []ui.DropdownItem{{Text: i18n.Label(i18n.Stop)}, {Text: i18n.Label(i18n.Start)}}
	items[selected].Selected = true
	d := ui.NewDropdownMenu(x, y+h, items) // Dropdown::showText below the button
	if g.vehicleWin != nil {
		d.Colour = g.vehicleWin.Colour
	}
	d.Highlighted = 1 - selected // stop → start highlighted, start → stop
	d.OnSelect = func(item int) {
		mode := sim.RunningModeStop
		if item == 1 {
			mode = sim.RunningModeStart
		}
		company, _ := g.w.ControllingCompany()
		c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
		if _, ok := m.VehicleChangeRunningMode(c, sim.CmdApply, headID, mode); !ok {
			log.Printf("[Game] can't %s vehicle %d", []string{"stop", "start"}[item], headID)
		}
		if g.vehicleWin != nil {
			g.openVehicleWindow(headID, g.vehicleWin.Tab)
		}
	}
	g.dropdown = d
}

// pickupImage is getPickupImageIdandTooltip's image for rail: the track
// object's image + kUiPickupFromTrack (16) when placed, kUiPlaceOnTrack (17)
// when not. PARTIAL: road, air and water (0: no image).
func (g *Game) pickupImage(h *scenario.VehicleEntity) int {
	if h.Mode != 0 || g.r == nil || g.r.ObjMgr == nil {
		return 0
	}
	t := g.r.ObjMgr.GetTrackObjectByIndex(int(h.TrackType))
	if t == nil {
		return 0
	}
	if h.TileX != -1 {
		return int(t.Image) + 16
	}
	return int(t.Image) + 17
}

// onPickup is Common::onPickup for rail: a vehicle not on the map takes the
// placing tool (toolSet; the tool is goloco's pickupHead); a placed one is
// lifted (VehiclePickup) and the window reopened (callOnMouseUp tabDetails).
// PARTIAL: the main view's unfocus, the error title, road / air / water.
func (g *Game) onPickup(headID uint16) {
	if g.w == nil || g.r == nil {
		return
	}
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	if h.TileX == -1 {
		g.pickupOn, g.pickupHead, g.ghostProgress = true, headID, -1 // toolSet; _ghostTrackProgress = -1
		g.ghostAirStation = 0xFFFF                                   // _ghostAirportStationId = null
		g.openVehicleDetails(headID)
		return
	}
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if _, ok := m.VehiclePickup(c, sim.CmdApply, headID); !ok {
		log.Printf("[Game] can't remove vehicle %d", headID)
		return
	}
	g.openVehicleDetails(headID)
}

// pickupToolDown is pickupToolDownLand for rail: the placement under the
// cursor, VehiclePlace, then pickupToolPlacementCommandCallback — shift starts
// the vehicle, the tool is cancelled, the window shown again. false: no track
// there (the tool stays); rail or road by the vehicle's mode (World.PlacementAt).
// PARTIAL: air and water, the error window.
func (g *Game) pickupToolDown(sx, sy int) bool {
	head := g.pickupHead
	switch g.pickupMode() {
	case 2:
		return g.pickupToolDownAir(sx, sy)
	case 3:
		return g.pickupToolDownWater(sx, sy)
	}
	p, ok := g.w.PlacementAt(head, sx, sy)
	if !ok {
		return false
	}
	m := &g.w.Map
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if g.ghostProgress != -1 && g.ghostPlace.TrackAndDirection == p.TrackAndDirection && g.ghostPlace.X == p.X &&
		g.ghostPlace.Y == p.Y && g.ghostPlace.Z == p.Z && int(g.ghostPlace.TrackProgress) == int(p.TrackProgress) && m.VehicleIsGhost(head) {
		p.ConvertGhost = true // the ghost where it stands becomes the vehicle
	} else {
		g.removeLandGhost()
	}
	g.ghostProgress = -1
	if _, ok := m.VehiclePlace(c, sim.CmdApply, p); !ok {
		log.Printf("[Game] can't place vehicle %d here", head)
		return false
	}
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		m.VehicleChangeRunningMode(c, sim.CmdApply, head, sim.RunningModeStart)
	}
	g.pickupOn = false
	g.openVehicleWindow(head, 0) // callOnMouseUp(tabMain)
	return true
}

// buildVehiclesMenu is ToolbarTop.cpp buildVehiclesMenuMouseDown (0x0043AD1F):
// one item per vehicle type the controlling company may build
// (availableVehicles), its build image recoloured to the company colour with
// build_trains … (menu_sprite_stringid), 25 px items (showBelow, flag 1<<6);
// the highlighted item is lastBuildVehiclesOption taken as an item index, as
// upstream does (an enum value, not the item's place — kept). Choosing one is
// buildVehiclesMenuDropdown: lastBuildVehiclesOption = the type, then
// BuildVehicle::openByType. PARTIAL: the button face doesn't yet show the last
// type's image (kBuildVehicleImages).
func (g *Game) buildVehiclesMenu(btn *ui.ToolbarButton) {
	if g.w == nil || g.r == nil || g.r.ObjMgr == nil || g.r.ObjMgr.InterfaceSkin == nil {
		return
	}
	m := &g.w.Map
	company, _ := g.w.ControllingCompany()
	var avail uint16
	if company >= 0 && company < len(m.Companies) {
		avail = m.Companies[company].AvailableVehicles
	}
	var colour uint8
	if company >= 0 && company < len(g.w.CompanyColours) {
		colour = g.w.CompanyColours[company]
	}
	skin := int(g.r.ObjMgr.InterfaceSkin.ImageOffset)
	var items []ui.DropdownItem
	for vt := 0; vt < 6; vt++ {
		if avail&(1<<vt) == 0 {
			continue
		}
		t := vt
		items = append(items, ui.DropdownItem{MenuSprite: true, Sprite: skin + objects.ISBuildVehicleFrame0[vt], SpriteColour: colour,
			Text: i18n.Label(i18n.BuildTrains + t), Action: func() {
				m.LastBuildVehiclesOption = uint8(t)
				g.openBuildVehicle(t)
			}})
	}
	if len(items) == 0 {
		return
	}
	d := ui.NewDropdownShowText(btn.X, btn.Y, btn.Width, btn.Height, ui.DropdownColour, items, 25, g.sw, g.sh)
	if sk := g.r.ObjMgr.InterfaceSkin; btn.ColorGroup >= 0 && btn.ColorGroup < 4 {
		cols := [4]uint8{sk.TopToolbarPrimaryColour, sk.TopToolbarSecondaryColour, sk.TopToolbarTertiaryColour, sk.TopToolbarQuaternaryColour}
		d.SetToolbarColour(cols[btn.ColorGroup])
	}
	if h := int(m.LastBuildVehiclesOption); h < len(items) {
		d.Highlighted = h
	}
	g.dropdown = d
}

// pickupToolUpdate is pickupToolUpdateLand for rail and road: the placement under the
// cursor; unchanged since the last ghost, nothing; else the old ghost lifted
// and the train placed there as a ghost (apply | ghost | noErrorWindow).
func (g *Game) pickupToolUpdate(sx, sy int) {
	switch g.pickupMode() {
	case 2:
		g.pickupToolUpdateAir(sx, sy)
		return
	case 3:
		g.pickupToolUpdateWater(sx, sy)
		return
	}
	p, ok := g.w.PlacementAt(g.pickupHead, sx, sy)
	if !ok {
		g.removeLandGhost()
		return
	}
	if g.ghostProgress != -1 && g.ghostPlace.TrackAndDirection == p.TrackAndDirection && g.ghostPlace.X == p.X &&
		g.ghostPlace.Y == p.Y && g.ghostPlace.Z == p.Z && g.ghostProgress == int(p.TrackProgress) {
		return
	}
	g.removeLandGhost()
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if _, ok := g.w.Map.VehiclePlace(c, sim.CmdApply|sim.CmdGhost|sim.CmdNoErrorWindow, p); ok {
		g.ghostPlace, g.ghostProgress = p, int(p.TrackProgress)
	}
}

// removeLandGhost is Vehicle.cpp's: a ghost-placed vehicle lifted again
// (VehiclePickup, apply | noErrorWindow | ghost); _ghostTrackProgress = -1.
func (g *Game) removeLandGhost() {
	if g.pickupMode() == 3 {
		g.removeBoatGhost()
		return
	}
	if g.w != nil && g.r != nil && g.w.Map.VehicleIsGhost(g.pickupHead) {
		company, _ := g.w.ControllingCompany()
		c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
		g.w.Map.VehiclePickup(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdGhost, g.pickupHead)
	}
	g.ghostProgress = -1
}

// overWindow reports whether a screen point is on one of goloco's windows or
// bars, where the tool does not act (the viewport is under none of them).
func (g *Game) overWindow(x, y int) bool {
	return g.vehicleWin != nil && g.vehicleWin.Contains(x, y) ||
		g.buildVehicle != nil && g.buildVehicle.Contains(x, y) ||
		g.townList != nil && g.townList.Contains(x, y) ||
		g.options != nil && g.options.Contains(x, y) ||
		g.toolbar.HandleClick(x, y) >= 0 ||
		uilayout.InTimePanel(x, y, g.sw, g.sh) || uilayout.InPlayerPanel(x, y, g.sw, g.sh)
}

// centreViewportOpen is Main's onCentreViewportControl: a text dropdown at the
// button — move the main view to show the vehicle, or focus it on the vehicle
// (following it) — and onCentreViewportDropdown.
func (g *Game) centreViewportOpen(headID uint16) {
	if g.vehicleWin == nil {
		return
	}
	x, y, _ := g.vehicleWin.CentreViewButton()
	d := ui.NewDropdownMenu(x, y, []ui.DropdownItem{{Text: i18n.Label(i18n.DropdownViewportMove)}, {Text: i18n.Label(i18n.DropdownViewportFocus)}})
	d.Colour = g.vehicleWin.Colour
	d.OnSelect = func(item int) {
		if item <= 0 {
			g.w.CentreOnVehicle(headID) // viewportCentreMain, without locking
			return
		}
		g.w.FocusOnVehicle(headID)
	}
	g.dropdown = d
}

// routeLines is Route::drawScroll's rows: "no route defined" for an empty
// table, each order as orderString formats it (the station's name; the
// cargo's for unload / wait), the current one marked, then the end of the
// list. PARTIAL: the cargo sprite after unload / wait.
func (g *Game) routeLines(h *scenario.VehicleEntity) []ui.RouteLine {
	m := &g.w.Map
	var lines []ui.RouteLine
	if h.SizeOfOrderTable == 1 {
		lines = append(lines, ui.RouteLine{Text: i18n.Label(i18n.NoRouteDefined)})
	}
	format := [6]int{i18n.OrdersEnd, i18n.OrdersStopAt, i18n.OrdersRouteThrough, i18n.OrdersRouteThoughtWaypoint, i18n.OrdersUnloadAll, i18n.OrdersWaitForFullLoadOf}
	for _, o := range m.OrderTable(h.OrderTableOffset) {
		t := o.Type()
		arg := ""
		switch t {
		case sim.OrderStopAt, sim.OrderRouteThrough:
			arg = m.StationName(o.Station())
		case sim.OrderUnloadAll, sim.OrderWaitFor:
			if c := int(o.Cargo()); g.r != nil && g.r.ObjMgr != nil && c < len(g.r.ObjMgr.CargoObjects) && g.r.ObjMgr.CargoObjects[c] != nil {
				arg = g.r.ObjMgr.CargoObjects[c].Name
			}
		}
		text := strings.Replace(i18n.Label(format[t]), "{STRINGID}", arg, 1)
		text = strings.TrimSpace(strings.Replace(text, "{SPRITE}", "", 1))
		lines = append(lines, ui.RouteLine{Text: text, Current: int(h.OrderTableOffset)+int(h.CurrentOrder) == o.Offset})
	}
	return append(lines, ui.RouteLine{Text: i18n.Label(i18n.OrdersEnd)})
}

// routeAction is Route::onMouseUp's order buttons: delete (the selected order,
// or the last), up / down (the selected one, the selection following it),
// skip (the current order), local / express (the vehicle's running mode,
// toggled only when the other one is on), reverse (the whole table, the
// current order following its order). PARTIAL: generateNumDisplayFrames after
// a change (the map's order numbers) — not ported for any of them.
func (g *Game) routeAction(headID uint16, action string) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	orders := m.OrderTable(h.OrderTableOffset)
	rel := func(o sim.Order) uint16 { return uint16(o.Offset - int(h.OrderTableOffset)) }
	sel := g.routeSel
	switch action {
	case "local", "express": // widx::localMode / expressMode: the player's own only
		if company < 0 || h.Owner != uint8(company) {
			return
		}
		if express := m.Entities[tr.Veh1].Var48&sim.Flags48ExpressMode != 0; express != (action == "express") {
			m.VehicleChangeRunningMode(c, sim.CmdApply, headID, sim.RunningModeToggleLocalExpress)
		}
	case "delete": // onOrderDelete
		if h.SizeOfOrderTable <= 1 || len(orders) == 0 {
			return
		}
		target := orders[len(orders)-1]
		if sel >= 0 && sel < len(orders) {
			target = orders[sel]
		}
		m.VehicleOrderDelete(c, sim.CmdApply, headID, rel(target))
		if g.routeSel >= len(m.OrderTable(m.Entities[tr.Head].OrderTableOffset)) {
			g.routeSel = -1 // no order selected any more
		}
	case "up", "down": // onOrderMove
		if h.SizeOfOrderTable <= 1 || sel < 0 || sel >= len(orders) {
			return
		}
		if action == "up" {
			if _, ok := m.VehicleOrderUp(c, sim.CmdApply, headID, rel(orders[sel])); ok && g.routeSel > 0 {
				g.routeSel--
			}
		} else if _, ok := m.VehicleOrderDown(c, sim.CmdApply, headID, rel(orders[sel])); ok && sel+1 < len(orders) {
			g.routeSel++
		}
	case "skip":
		m.VehicleOrderSkip(c, sim.CmdApply, headID)
	case "reverse": // orderReverseCommand: the selection stays (orderTableIndex untouched)
		m.VehicleOrderReverse(c, sim.CmdApply, headID)
	}
	g.openVehicleWindow(headID, 4)
}

// routeToolDown is Route::onToolDown for stops: the station under the cursor
// (World.RouteStopAt) added through addNewOrder — inserted before the selected
// order, else before the table's end; the selection moves down with the order
// it was on; plain track or road gives a route-through waypoint instead.
// PARTIAL: the waypoint sound; water waypoints.
func (g *Game) routeToolDown(sx, sy int) {
	m := &g.w.Map
	tr, err := m.TrainOf(g.routeHead)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	hit, ok := g.w.RouteAt(sx, sy, h.Mode)
	if !ok {
		return
	}
	order := sim.OrderStopAtRaw(hit.Station)
	if hit.Waypoint != nil {
		order = hit.Waypoint // route through (0x004B5160 / 0x004B5223)
	}
	offset := uint16(h.SizeOfOrderTable - 1) // addNewOrder: before the End order
	if orders := m.OrderTable(h.OrderTableOffset); g.routeSel >= 0 && g.routeSel < len(orders) {
		offset = uint16(orders[g.routeSel].Offset - int(h.OrderTableOffset))
	}
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	before := h.SizeOfOrderTable
	if _, ok := m.VehicleOrderInsert(c, sim.CmdApply, g.routeHead, offset, order); !ok {
		return
	}
	if m.Entities[tr.Head].SizeOfOrderTable != before && g.routeSel != -1 {
		g.routeSel++
	}
	g.openVehicleWindow(g.routeHead, 4)
}

// pickupMode is the picked-up vehicle's TransportMode (0xFF: none).
func (g *Game) pickupMode() uint8 {
	if g.w == nil {
		return 0xFF
	}
	tr, err := g.w.Map.TrainOf(g.pickupHead)
	if err != nil {
		return 0xFF
	}
	return g.w.Map.Entities[tr.Head].Mode
}

// pickupToolUpdateWater is Vehicle.cpp's (0x004B2B9E): the dock nearest the
// cursor (World.WaterPlacementAt); the same as the last ghost, nothing; else
// the old ghost lifted and the boat placed there as a ghost
// (_ghostTrackProgress 0 marks a boat ghost).
func (g *Game) pickupToolUpdateWater(sx, sy int) {
	p, ok := g.w.WaterPlacementAt(g.pickupHead, sx, sy)
	if !ok {
		g.removeBoatGhost()
		return
	}
	if g.ghostProgress == 0 && g.ghostWater.X == p.X && g.ghostWater.Y == p.Y && g.ghostWater.Z == p.Z {
		return
	}
	g.ghostWater = p
	g.removeBoatGhost()
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if _, ok := g.w.Map.VehiclePlaceWater(c, sim.CmdApply|sim.CmdGhost|sim.CmdNoErrorWindow, p); ok {
		g.ghostProgress = 0
	}
}

// removeBoatGhost is Vehicle.cpp's: a ghost boat lifted
// (VehiclePickupWater, apply | noErrorWindow | ghost); _ghostTrackProgress = -1.
func (g *Game) removeBoatGhost() {
	if g.w != nil && g.r != nil && g.w.Map.VehicleIsGhost(g.pickupHead) {
		company, _ := g.w.ControllingCompany()
		c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
		g.w.Map.VehiclePickupWater(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdGhost, g.pickupHead)
	}
	g.ghostProgress = -1
}

// pickupToolDownWater is Vehicle.cpp's: the dock nearest the cursor; the
// ghost standing there becomes the boat (convertGhost), else the ghost goes
// and the boat is placed; then as pickupToolPlacementCommandCallback.
// PARTIAL: pickupToolError's window.
func (g *Game) pickupToolDownWater(sx, sy int) bool {
	head := g.pickupHead
	p, ok := g.w.WaterPlacementAt(head, sx, sy)
	if !ok {
		return false
	}
	m := &g.w.Map
	if g.ghostProgress == 0 && g.ghostWater.X == p.X && g.ghostWater.Y == p.Y && g.ghostWater.Z == p.Z && m.VehicleIsGhost(head) {
		p.ConvertGhost = true
	} else {
		g.removeBoatGhost()
	}
	g.ghostProgress = -1
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if _, ok := m.VehiclePlaceWater(c, sim.CmdApply, p); !ok {
		log.Printf("[Game] can't place vehicle %d here", head)
		return false
	}
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		m.VehicleChangeRunningMode(c, sim.CmdApply, head, sim.RunningModeStart)
	}
	g.pickupOn = false
	g.openVehicleWindow(head, 0)
	return true
}

// pickupToolUpdateAir is Vehicle.cpp's: the airport terminal nearest the
// cursor (World.AirPlacementAt); the same as the ghost's, nothing; else the
// old ghost lifted and the aircraft placed there as a ghost
// (_ghostAirportStationId / _ghostAirportNode).
func (g *Game) pickupToolUpdateAir(sx, sy int) {
	p, ok := g.w.AirPlacementAt(g.pickupHead, sx, sy)
	if !ok {
		g.removeAirplaneGhost()
		return
	}
	if g.ghostAirStation != 0xFFFF && g.ghostAirStation == p.StationID && g.ghostAirNode == p.AirportNode {
		return
	}
	g.removeAirplaneGhost()
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if _, ok := g.w.Map.VehiclePlaceAir(c, sim.CmdApply|sim.CmdGhost|sim.CmdNoErrorWindow, p); ok {
		g.ghostAirStation, g.ghostAirNode = p.StationID, p.AirportNode
	}
}

// removeAirplaneGhost is Vehicle.cpp's: a ghost aircraft lifted
// (VehiclePickupAir, apply | noErrorWindow | ghost); no ghost airport.
func (g *Game) removeAirplaneGhost() {
	if g.w != nil && g.r != nil && g.w.Map.VehicleIsGhost(g.pickupHead) {
		company, _ := g.w.ControllingCompany()
		c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
		g.w.Map.VehiclePickupAir(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdGhost, g.pickupHead)
	}
	g.ghostAirStation = 0xFFFF
}

// pickupToolDownAir is Vehicle.cpp's (0x004B2E18): the terminal nearest the
// cursor; the ghost standing there becomes the aircraft (convertGhost), else
// the ghost goes and it is placed; then as
// pickupToolPlacementCommandCallback. PARTIAL: pickupToolError's window.
func (g *Game) pickupToolDownAir(sx, sy int) bool {
	head := g.pickupHead
	p, ok := g.w.AirPlacementAt(head, sx, sy)
	if !ok {
		return false
	}
	m := &g.w.Map
	if g.ghostAirStation == p.StationID && g.ghostAirNode == p.AirportNode && m.VehicleIsGhost(head) {
		p.ConvertGhost = true
	}
	if !p.ConvertGhost {
		g.removeAirplaneGhost()
	}
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	if _, ok := m.VehiclePlaceAir(c, sim.CmdApply, p); !ok {
		log.Printf("[Game] can't place vehicle %d here", head)
		return false
	}
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		m.VehicleChangeRunningMode(c, sim.CmdApply, head, sim.RunningModeStart)
	}
	g.pickupOn = false
	g.openVehicleWindow(head, 0)
	return true
}

// vehicleFinances is what Finances::draw formats: Vehicle1's lastIncome (its
// day, then each cargo's unit name with the quantity, the distance in tiles,
// the age in days and the profit), the head's last average speed, the
// running cost, veh2's totalRecentProfit / 4, the head's totalRefundCost.
func (g *Game) vehicleFinances(tr sim.Train) ui.VehicleFinances {
	m := &g.w.Map
	head, v1 := m.Entities[tr.Head], m.Entities[tr.Veh1]
	var f ui.VehicleFinances
	li := v1.LastIncome
	if li.Day != -1 {
		y, mo, d := uilayout.CalcDate(uint32(li.Day))
		f.IncomeHeader = ui.FillArgs(i18n.Str(i18n.LastIncomeOnDate), uilayout.DateDMY(d, i18n.Month(mo+1), y))
		for i := 0; i < 4; i++ {
			c := li.CargoTypes[i]
			if c == 0xFF {
				continue
			}
			unit := ""
			if om := g.r.ObjMgr; om != nil && int(c) < len(om.CargoObjects) && om.CargoObjects[c] != nil {
				unit = ui.ObjectString(om.CargoObjects[c].UnitNamePlural)
				if li.CargoQtys[i] == 1 {
					unit = ui.ObjectString(om.CargoObjects[c].UnitNameSingular)
				}
			}
			f.Income = append(f.Income, ui.FillArgs(i18n.Str(i18n.TransportedBlocksInDays), unit, ui.Comma16(int(li.CargoQtys[i])),
				ui.Comma16(int(li.CargoDistances[i])), ui.Comma16(int(li.CargoAges[i])), uilayout.Currency48(int64(li.CargoProfits[i]))))
		}
	} else {
		f.IncomeHeader = i18n.Str(i18n.LastIncomeNa)
	}
	if head.LastAverageSpeed != 0 {
		f.AverageSpeed = ui.FillArgs(i18n.Str(i18n.LastJourneyAverageSpeed), uilayout.Velocity(int(head.LastAverageSpeed)))
	}
	f.RunningCost = ui.FillArgs(i18n.Str(i18n.VehicleMonthlyRunningCost), uilayout.Currency48(int64(m.RunningCost(tr, g.r.ObjMgr))))
	var profit int64
	for _, r := range m.VehicleListRows() {
		if int(r.HeadID) == int(head.ID) {
			profit = r.Profit // totalRecentProfit / 4
		}
	}
	f.Profit = ui.FillArgs(i18n.Str(i18n.VehicleMonthlyProfit), uilayout.Currency48(profit))
	f.SaleValue = ui.FillArgs(i18n.Str(i18n.SaleValueOfVehicle), uilayout.Currency48(int64(head.TotalRefundCost)))
	return f
}

// vehicleCargo is what Cargo::draw formats: per car (its first component's
// front bogie and body) the body's primary and the bogie's secondary cargo,
// each with a quantity "{unit} from {station}", or "Empty" when it carries
// something but holds none; the totals and capacities by cargo type
// (generateCargoTotalString / CapacityString: cargoLUTToString, "Empty" when
// none); refit for the player's own refittable vehicle (canRefit).
func (g *Game) vehicleCargo(tr sim.Train, cars []ui.TrainCarDraw) ui.VehicleCargo {
	m := &g.w.Map
	om := g.r.ObjMgr
	unit := func(c uint8, n uint32) string {
		if om == nil || int(c) >= len(om.CargoObjects) || om.CargoObjects[c] == nil {
			return ""
		}
		if n == 1 {
			return ui.ObjectString(om.CargoObjects[c].UnitNameSingular)
		}
		return ui.ObjectString(om.CargoObjects[c].UnitNamePlural)
	}
	var out ui.VehicleCargo
	var totals, caps [32]uint32
	for i, car := range tr.Cars {
		front, body := m.Entities[car[0].Front], m.Entities[car[0].Body]
		row := ui.VehicleCargoRow{}
		if i < len(cars) {
			row.Car = cars[i]
		}
		for _, cg := range []scenario.VehicleCargo{front.Cargo, body.Cargo} {
			if cg.Type != 0xFF && int(cg.Type) < len(totals) {
				totals[cg.Type] += uint32(cg.Qty)
				caps[cg.Type] += uint32(cg.MaxQty)
			}
		}
		if body.Cargo.Type != 0xFF {
			if front.Cargo.Qty != 0 || body.Cargo.Qty != 0 {
				for _, cg := range []scenario.VehicleCargo{body.Cargo, front.Cargo} { // drawCargoText: primary, then secondary
					if cg.Qty == 0 {
						continue
					}
					row.Lines = append(row.Lines, ui.FillArgs(i18n.Str(i18n.CargoFrom), unit(cg.Type, uint32(cg.Qty)), ui.Comma16(int(cg.Qty)), m.StationName(int(cg.TownFrom))))
				}
			} else {
				row.Lines = []string{i18n.Str(i18n.CargoEmpty)}
			}
		}
		out.Rows = append(out.Rows, row)
	}
	lut := func(t *[32]uint32) string { // cargoLUTToString
		s := ""
		for c, n := range t {
			if n == 0 {
				continue
			}
			if s != "" {
				s += ", "
			}
			s += ui.FillArgs(unit(uint8(c), n), ui.Comma16(int(n)))
		}
		if s == "" {
			s = i18n.Str(i18n.CargoEmpty2)
		}
		return s
	}
	out.Total = ui.FillArgs(i18n.Str(i18n.TotalStringid), lut(&totals))
	out.Capacity = ui.FillArgs(i18n.Str(i18n.VehicleCapacityStringid), lut(&caps))
	if h := m.Entities[tr.Head]; m.IsPlayerCompany(int(h.Owner)) && len(tr.Cars) > 0 && om != nil {
		if id := m.Entities[tr.Cars[0][0].Front].ObjectID; int(id) < len(om.Vehicles) && om.Vehicles[id] != nil {
			out.CanRefit = om.Vehicles[id].Flags&(1<<14) != 0 // VehicleObjectFlags::refittable
		}
	}
	return out
}

// refitOpen is Cargo::onRefitButton: a dropdown below the button of every
// cargo with the refit flag — its plural unit name with the capacity the car
// would have (getNumUnitsForCargo), the current one selected, the first
// highlighted; onDropdown asks first when the train holds cargo
// (confirmComponentChange, the keep-cargo cheat not consulted), then
// VehicleRefit.
func (g *Game) refitOpen(headID uint16, x, y, w, h int) {
	m := &g.w.Map
	om := g.r.ObjMgr
	tr, err := m.TrainOf(headID)
	if err != nil || len(tr.Cars) == 0 || om == nil {
		return
	}
	obj := om.Vehicles[m.Entities[tr.Cars[0][0].Front].ObjectID]
	if obj == nil {
		return
	}
	cargo := om.CargoFor(obj)
	primary := uint8(0xFF)
	for i := 0; i < 32; i++ {
		if uint32(cargo.CompatibleCargoCategories[0])&(1<<i) != 0 {
			primary = uint8(i)
			break
		}
	}
	current := m.Entities[tr.Cars[0][0].Body].Cargo.Type
	var items []ui.DropdownItem
	var ids []uint8
	for id, co := range om.CargoObjects {
		if co == nil || co.Flags&(1<<1) == 0 { // CargoObjectFlags::refit
			continue
		}
		n := sim.NumUnitsForCargo(om, uint32(cargo.MaxCargo[0]), primary, uint8(id))
		items = append(items, ui.DropdownItem{Text: ui.FillArgs(ui.ObjectString(co.UnitNamePlural), ui.Comma16(int(n))), Selected: uint8(id) == current, MoveX10: true})
		ids = append(ids, uint8(id))
	}
	if len(items) == 0 {
		return
	}
	d := ui.NewDropdownMenu(x, y+h, items)
	if g.vehicleWin != nil {
		d.Colour = g.vehicleWin.Colour
	}
	d.Highlighted = 0
	d.OnSelect = func(item int) {
		apply := func() {
			company, _ := g.w.ControllingCompany()
			c := &sim.Cmd{Company: uint8(company), Objects: om}
			if _, ok := m.VehicleRefit(c, sim.CmdApply, headID, ids[item]); !ok {
				log.Printf("[Game] %s", i18n.Str(i18n.CantRefitVehicle))
			}
			if g.vehicleWin != nil {
				g.openVehicleWindow(headID, 2)
			}
		}
		if t, err := m.TrainOf(headID); err == nil && m.HasAnyCargo(t) {
			g.prompt = ui.NewPromptOkCancel(i18n.Str(i18n.ConfirmVehicleComponentRefitCargoWarningTitle),
				i18n.Str(i18n.ConfirmVehicleComponentRefitCargoWarningMessage), i18n.Str(i18n.ConfirmVehicleComponentRefitCargoWarningConfirm), g.sw, g.sh, apply)
			return
		}
		apply()
	}
	g.dropdown = d
}

// renameVehicle is Common::renameVehicle: the text input titled "Name
// {type}", "Type new name for this {type}:", on the current name (inputSize
// kUserStringSize − 1); textInput: VehicleRename with the text (an empty
// one changes nothing — upstream returns before the command).
func (g *Game) renameVehicle(headID uint16, titleColour, colour uint8) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	typ := i18n.Str([]int{i18n.Train, i18n.Bus, i18n.Truck, i18n.Tram, i18n.Aircraft, i18n.Ship}[min(int(h.VehicleType), 5)])
	tab := 0
	if g.vehicleWin != nil {
		tab = g.vehicleWin.Tab
	}
	g.textInput = ui.OpenTextInput(ui.FillArgs(i18n.Str(i18n.TitleNameVehicle), typ), ui.FillArgs(i18n.Str(i18n.PromptEnterNewVehicleName), typ),
		m.VehicleName(h.Name, int(h.OrdinalNumber)), 31, g.sw, g.sh, titleColour, colour, true, func(text string) {
			if text == "" {
				return
			}
			company, _ := g.w.ControllingCompany()
			if _, ok := m.VehicleRename(&sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}, sim.CmdApply, headID, text); !ok {
				log.Printf("[Game] %s", i18n.Str(i18n.CantRenameThisVehicle))
			}
			if g.vehicleWin != nil {
				g.openVehicleWindow(headID, tab)
			}
		})
}

// sellVehicle is Details' remove button: VehicleSell for the head, as a
// company command (VehicleDisposals); sold, its window closes (deleteTrain's
// WindowManager::close) and the pickup tool with it.
func (g *Game) sellVehicle(headID uint16) {
	m := &g.w.Map
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	cost, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehicleDisposals, func(flags uint8) (int32, bool) {
		return m.VehicleSell(c, flags, headID)
	})
	if !ok {
		log.Printf("[Game] can't sell vehicle %d", headID)
		return
	}
	log.Printf("[Game] sold vehicle %d for %d", headID, -cost)
	g.vehicleWin, g.pickupOn, g.routeTool = nil, false, false
}

// dropCar is Details::scrollDragEnd for this vehicle's car at row: dest −2
// the remove button (VehicleSell of that car, VehicleDisposals), else the row
// it goes before (−1: the tail) — VehicleRearrange; each asked first when the
// train holds cargo (confirmComponentChange; the keep-cargo cheat is not
// kept, so it never skips the question).
func (g *Game) dropCar(headID uint16, row, dest int) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil || row < 0 || row >= len(tr.Cars) {
		return
	}
	src := m.Entities[tr.Cars[row][0].Front].ID
	company, _ := g.w.ControllingCompany()
	c := &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}
	var run func()
	var title, msg, ok int
	if dest == -2 {
		title, msg, ok = i18n.ConfirmVehicleComponentSellCargoWarningTitle, i18n.ConfirmVehicleComponentSellCargoWarningMessage, i18n.ConfirmVehicleComponentSellCargoWarningConfirm
		run = func() {
			if _, done := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehicleDisposals, func(f uint8) (int32, bool) { return m.VehicleSell(c, f, src) }); !done {
				log.Printf("[Game] can't sell car %d", src)
			}
			g.openVehicleWindow(headID, 1)
		}
	} else {
		dst := m.Entities[tr.Tail].ID
		if dest >= 0 && dest < len(tr.Cars) {
			dst = m.Entities[tr.Cars[dest][0].Front].ID
		}
		title, msg, ok = i18n.ConfirmVehicleComponentMoveCargoWarningTitle, i18n.ConfirmVehicleComponentMoveCargoWarningMessage, i18n.ConfirmVehicleComponentMoveCargoWarningConfirm
		run = func() {
			if _, done := m.DoCommandFor(c, sim.CmdApply, 1, func(f uint8) (int32, bool) { return m.VehicleRearrange(c, f, src, dst) }); !done { // TrainRunningCosts
				log.Printf("[Game] can't move car %d", src)
			}
			g.openVehicleWindow(headID, 1)
		}
	}
	if m.HasAnyCargo(tr) {
		g.prompt = ui.NewPromptOkCancel(i18n.Str(title), i18n.Str(msg), i18n.Str(ok), g.sw, g.sh, run)
		return
	}
	run()
}
