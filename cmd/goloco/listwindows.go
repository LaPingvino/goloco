package main

import (
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
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	company, _ := g.w.ControllingCompany()
	own := int(h.Owner) == company
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
	g.vehicleWin = &ui.VehicleWindow{X: (g.sw - ww) / 2, Y: (g.sh - wh) / 2, W: ww, H: wh, TitleColour: tc, Colour: wc,
		Title: title, VehicleType: int(h.VehicleType), Cars: cars, Names: names,
		State:   ui.VehicleDetailsState{Own: own, Placed: h.TileX != -1, NoCars: len(tr.Cars) == 0},
		OnClose: func() { g.vehicleWin, g.pickupOn = nil, false }, OnPickup: func() { g.onPickup(headID) },
		PickupActive: g.pickupOn && g.pickupHead == headID, PickupImage: g.pickupImage(&h),
		Tab: tab, Main: ui.VehicleMainState{Own: own, Manual: h.VehicleFlags&sim.VehicleFlagManualControl != 0},
		StopStartImage: stopStartImage(h.VehicleFlags),
		OnTab:          func(t int) { g.openVehicleWindow(headID, t) },
		OnStopStart:    func(x, y, w, hh int) { g.stopStartOpen(headID, x, y, w, hh) }}
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
// there (the tool stays). PARTIAL: the ghost (pickupToolUpdate), the error
// window.
func (g *Game) pickupToolDown(sx, sy int) bool {
	head := g.pickupHead
	p, ok := g.w.RailPlacementAt(head, sx, sy)
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

// pickupToolUpdate is pickupToolUpdateLand for rail: the placement under the
// cursor; unchanged since the last ghost, nothing; else the old ghost lifted
// and the train placed there as a ghost (apply | ghost | noErrorWindow).
func (g *Game) pickupToolUpdate(sx, sy int) {
	p, ok := g.w.RailPlacementAt(g.pickupHead, sx, sy)
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
