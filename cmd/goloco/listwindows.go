package main

import (
	"os"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// listwindows.go — the right-toolbar list windows: the ported VehicleList,
// StationList and TownList (pkg/ui/townlist.go's table window), and the Town
// window a town's row opens.

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
	tc, wc := g.windowColours(true)
	// The window's owner is the company, so WindowManager draws its primary
	// colour as the company's colour (CompanyManager::getCompanyColour).
	if company < len(g.w.CompanyColours) {
		tc = g.w.CompanyColours[company]
	}
	// The cargo filter's list: every loaded cargo object, by its slot.
	var cargos []ui.VehicleListCargo
	if g.r.ObjMgr != nil {
		for id, c := range g.r.ObjMgr.CargoObjects {
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
	g.listTabIcons(w, vehicleTabFrames, vehicleTabCounts, int(tc)) // recoloured to the company colour
	w.TabDisabled = func(i int) bool { // disableUnavailableVehicleTypes: availableVehicles ^ 0x3F
		return company < len(g.w.Map.Companies) && g.w.Map.Companies[company].AvailableVehicles&(1<<i) == 0
	}
	g.townList = w
}

// openStationsWindow lists every placed/loaded station on the map.
func (g *Game) openStationsWindow() { g.openStationList(0) }

// stationTabMasks is StationList.cpp tabInformationByType's stationMask, by
// tab: all modes, rail, road, air, water (StationFlags transportMode*).
var stationTabMasks = [5]uint16{0xF, 1 << 0, 1 << 1, 1 << 2, 1 << 3}

// openStationList is StationList::open(the controlling company, tab): the
// company's stations of the tab's transport modes (populateStationList:
// owner, not flag_5 — not fully created —, flags & the tab's mask), captioned
// "{company} - All Stations" / "Rail Stations" …; a tab click reopens it on
// that tab.
func (g *Game) openStationList(tab int) {
	company, cname := g.w.ControllingCompany()
	var rows []ui.StationRow
	for _, s := range g.w.StationList() {
		if s.Owner >= 0 && s.Owner != company {
			continue
		}
		if s.Flags&(1<<5) != 0 || s.Flags&stationTabMasks[tab] == 0 {
			continue
		}
		rows = append(rows, ui.StationRow{ID: s.ID, Name: s.Name, Status: s.Status, Waiting: s.Waiting, Accepts: s.Accepts,
			AcceptsKey: g.w.StationAcceptsKey(s.ID)})
	}
	tc, wc := g.windowColours(true)
	// The window's owner is the company, so WindowManager draws its primary
	// colour as the company's colour (CompanyManager::getCompanyColour).
	if company < len(g.w.CompanyColours) {
		tc = g.w.CompanyColours[company]
	}
	w := ui.NewStationList(g.sw, rows, tc, wc)
	if cname != "" {
		w.Caption = strings.Replace(i18n.Label(i18n.StringidAllStations+tab), "{STRINGID}", cname, 1) // "{STRINGID} - Rail Stations" …
	}
	w.CurrentTab = tab
	w.OnTab = func(t int) { g.openStationList(t) }
	w.OnClose = func() { g.townList = nil }
	w.OnRow = func(id int) { g.openStationWindow(id) }             // onScrollMouseDown: Station::open
	g.listTabIcons(w, stationTabFrames, stationTabCounts, int(tc)) // the company colour
	g.townList = w
}

// openTownsWindow lists the scenario's towns with population and size.
func (g *Game) openTownsWindow() {
	// The ported TownList window (pkg/ui/townlist.go).
	infos := g.w.TownList()
	rows := make([]ui.TownRow, len(infos))
	for i, t := range infos {
		rows[i] = ui.TownRow{ID: i, Name: t.Name, Type: t.Size, Size: t.SizeIndex, Population: int(t.Population), Stations: t.NumStations}
	}
	tc, wc := g.windowColours(false)
	w := ui.NewTownList(g.sw, rows, tc, wc)
	w.OnClose = func() { g.townList = nil }
	w.OnRow = func(id int) { // onScrollMouseDown: Town::open
		g.openTownWindow(id)
	}
	g.listTabIcons(w, townTabFrames, townTabCounts, -1)
	g.townList = w
}

// openTownWindow is Town::open: the Town window on its Town tab, its
// viewport on the town; the centre-view button moves the main view there.
func (g *Game) openTownWindow(id int) {
	infos := g.w.TownList()
	if id < 0 || id >= len(infos) {
		return
	}
	t := infos[id]
	tc, wc := g.windowColours(false)
	w := ui.NewTownWindow(g.sw, g.sh, t.Name, t.Size, int(t.Population), tc, wc)
	w.View = func(vw, vh int) *gfx8.RenderTarget { return g.w.TownView(id, vw, vh) }
	w.OnCentreView = func() { g.w.CentreOnTown(id) }
	w.Graph = func(i int) (ui.GraphTab, bool) { // the Population tab: the town's history from historyMinPopulation
		if i != 1 || id >= len(g.w.Map.Towns) {
			return ui.GraphTab{}, false
		}
		t := &g.w.Map.Towns[id]
		size := min(int(t.HistorySize), len(t.History))
		return ui.GraphTab{Caption: i18n.TitleTownPopulation, History: t.History[:size], FirstTick: int(t.HistoryMinPopulation),
			Month: int(g.w.Map.CurrentMonth), Year: int(g.w.Map.CurrentYear), Layout: ui.TownGraphLayout}, true
	}
	// drawTabs: the town (toolbar_menu_towns), population (tab_population, 8
	// frames, recoloured to the window's secondary colour) and company ratings
	// (tab_ratings, 16 frames); the current one animated by frameNo / 4.
	w.TabIcon = func(i int) (int, int, int) {
		sk := g.skin()
		if sk == nil {
			return -1, -1, -1
		}
		anim := func(n int) int {
			if i == w.Tab {
				return (g.windowFrame / 4) % n
			}
			return 0
		}
		base := int(sk.ImageOffset)
		switch i {
		case 0:
			return base + 189, -1, -1
		case 1:
			return base + 57 + anim(8), int(wc), -1
		}
		return base + 145 + anim(16), -1, -1
	}
	w.OnClose = func() { g.viewWin = nil }
	g.viewWin = w
}

// openIndustryList is IndustryList::open (its list tab): every industry with
// its status and production; a row opens its Industry window.
func (g *Game) openIndustryList() {
	var rows []ui.IndustryListRow
	for _, r := range g.w.Map.IndustryList(g.r.ObjMgr) {
		rows = append(rows, ui.IndustryListRow{ID: r.ID, Name: r.Name, Status: r.Status, Produces: r.Produces,
			Transported: r.Transported, LastMonth: r.LastMonth, Unit: r.Unit})
	}
	tc, wc := g.windowColours(false)
	w := ui.NewIndustryList(g.sw, rows, tc, wc)
	w.OnClose = func() { g.townList = nil }
	w.OnRow = func(id int) { g.openIndustryWindow(id) } // onScrollMouseDown: Industry::open
	w.OnTab = func(t int) {
		if t == 1 {
			g.openNewIndustries()
		}
	}
	g.listTabIcons(w, industryTabFrames, industryTabCounts, -1)
	g.townList = w
}

// openNewIndustries is the industry list's second tab (NewIndustries):
// updateBuildableIndustries — outside the editor, the industries the player
// may found (canBeFoundedByPlayer) and whose years include this one — each
// with its inflation-adjusted cost (costFactor, costIndex, divisor 3).
func (g *Game) openNewIndustries() {
	om, m := g.r.ObjMgr, &g.w.Map
	var choices []ui.NewIndustryChoice
	for i, obj := range om.IndustryObjects {
		if obj == nil || obj.Flags&objects.IndustryFlagCanBeFoundedByPlayer == 0 {
			continue
		}
		if y := int(m.CurrentYear); y < int(obj.DesignedYear) || y > int(obj.ObsoleteYear) {
			continue
		}
		var parts []uint8
		if len(obj.Variations) > 0 {
			parts = obj.Variations[0]
		}
		colour := uint8(0) // bitScanReverse(availableColours), else black
		for c := 31; c >= 0; c-- {
			if obj.AvailableColours&(1<<c) != 0 {
				colour = uint8(c)
				break
			}
		}
		choices = append(choices, ui.NewIndustryChoice{ID: i, Name: obj.Strings[0], Cost: int64(m.InflationAdjustedCost(obj.CostFactor, obj.CostIndex, 3)),
			Image: obj.BuildingImageIDs, Colour: colour, Parts: parts, PartHeights: obj.PartHeights})
	}
	tc, wc := g.windowColours(false)
	w := ui.NewIndustryGrid(g.sw, choices, g.lastIndustryOption, tc, wc, func(id int) {
		g.lastIndustryOption = id
		g.indTool.costOK = false // onScrollMouseDown: _industryPlacementCost = kFailure
	})
	w.PlacementCost = func() (int64, bool) { return int64(g.indTool.cost), g.indTool.ghost && g.indTool.costOK }
	w.OnClose = func() { g.industryToolAbort(); g.townList = nil } // onClose: toolCancel
	w.OnTab = func(t int) {
		if t == 0 {
			g.industryToolAbort()
			g.openIndustryList()
		}
	}
	g.listTabIcons(w, industryTabFrames, industryTabCounts, -1)
	g.townList = w
	g.industryToolStart(w)
}

// openIndustryWindow is IndustryWindow open (industry id): the window on its
// Industry tab, its viewport on the industry, its status getStatusString; the
// centre-view button moves the main view there.
func (g *Game) openIndustryWindow(id int) {
	m := &g.w.Map
	idx := -1
	for i := range m.Industries {
		if int(m.Industries[i].ID) == id && m.Industries[i].Name != 0xFFFF {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}
	name, _ := m.IndustryName(idx, g.r.ObjMgr)
	tc, wc := g.windowColours(false)
	w := ui.NewIndustryWindow(g.sw, g.sh, name, m.IndustryStatusString(idx, g.r.ObjMgr), tc, wc)
	w.View = func(vw, vh int) *gfx8.RenderTarget { return g.w.IndustryView(idx, vw, vh) }
	w.OnCentreView = func() { g.w.CentreOnIndustry(idx) }
	w.OnClose = func() { g.viewWin = nil }
	w.CargoTab = 3 // tab_transported
	w.Cargo = func() (bool, bool, []ui.CargoMonthLine) {
		rec, prod, months := g.w.Map.IndustryCargoMonths(idx, g.r.ObjMgr)
		lines := make([]ui.CargoMonthLine, len(months))
		for k, c := range months {
			lines[k] = ui.CargoMonthLine{Produced: c.Produced, Unit: c.Unit, Qty: c.Qty, Percent: c.Percent}
		}
		return rec, prod, lines
	}
	if obj := g.r.ObjMgr.GetIndustryObjectByIndex(int(m.Industries[idx].ObjectID)); obj != nil { // setDisabledWidgets
		produced, _ := g.r.ObjMgr.IndustryCargoIDs(obj)
		w.TabDisabled[1], w.TabDisabled[2] = produced[0] == 0xFF, produced[1] == 0xFF
		// drawTabs: the industry (toolbar_menu_industries), the production
		// tabs (tab_production, 8 frames, in the window's secondary colour,
		// with the cargo's inline sprite) and transported (tab_transported, 7
		// frames); the current one animated by frameNo / 4.
		w.TabIcon = func(i int) (int, int, int) {
			sk := g.skin()
			if sk == nil {
				return -1, -1, -1
			}
			anim := func(n int) int {
				if i == w.Tab {
					return (g.windowFrame / 4) % n
				}
				return 0
			}
			base := int(sk.ImageOffset)
			switch i {
			case 0:
				return base + 191, -1, -1
			case 1, 2:
				cargo := -1
				if c := int(produced[i-1]); c < len(g.r.ObjMgr.CargoObjects) && g.r.ObjMgr.CargoObjects[c] != nil {
					cargo = int(g.r.ObjMgr.CargoObjects[c].ImageOffset) // unitInlineSprite
				}
				return base + 89 + anim(8), int(wc), cargo
			}
			return base + 161 + anim(7), -1, -1
		}
		w.Graph = func(i int) (ui.GraphTab, bool) { // the Production tabs (1, 2)
			if i != 1 && i != 2 {
				return ui.GraphTab{}, false
			}
			label := ""
			if c := int(produced[0]); c < len(g.r.ObjMgr.CargoObjects) && g.r.ObjMgr.CargoObjects[c] != nil {
				label = g.r.ObjMgr.CargoObjects[c].UnitsAndCargoName // producedCargoType[0] on both tabs
			}
			ind := &g.w.Map.Industries[idx]
			n := i - 1
			size := min(int(ind.HistorySize[n]), len(ind.History[n]))
			return ui.GraphTab{Caption: i18n.TitleIndustryMonthlyProduction, Label: label, History: ind.History[n][:size],
				Month: int(g.w.Map.CurrentMonth), Year: int(g.w.Map.CurrentYear), Layout: ui.IndustryGraphLayout}, true
		}
	}
	g.viewWin = w
}

// listTabIcons is a list window's drawTabs: tab i's interface skin image
// (skin->img + frames[i]), the current tab animated through its counts[i]
// frames by the window's frameNo / 2 (1: still), recoloured to colour (< 0:
// none — the company colour for the station and vehicle lists).
func (g *Game) listTabIcons(w *ui.TownListWindow, frames, counts []int, colour int) {
	w.TabIcon = func(i int) (int, int) {
		sk := g.skin()
		if sk == nil || i >= len(frames) {
			return -1, -1
		}
		img := int(sk.ImageOffset) + frames[i]
		if i == w.CurrentTab && counts[i] > 1 {
			img += (g.windowFrame / 2) % counts[i]
		}
		return img, colour
	}
}

// The lists' tab images (InterfaceSkinObject.h ImageIds).
var (
	townTabFrames     = []int{189, 363, 379, 395} // toolbar_menu_towns, build_town / build_buildings / build_misc_buildings_frame_0
	townTabCounts     = []int{1, 16, 16, 16}
	industryTabFrames = []int{191, 347} // toolbar_menu_industries, build_industry_frame_0
	industryTabCounts = []int{1, 16}
	stationTabFrames  = []int{objects.ISAllStations, objects.ISAllStations + 1, objects.ISAllStations + 2, objects.ISAllStations + 3, objects.ISAllStations + 4}
	stationTabCounts  = []int{1, 1, 1, 1, 1}
	// the vehicle list's: trains, buses, trucks, trams, aircraft, ships (vehicle_*_frame_0, 8 frames each)
	vehicleTabFrames = []int{objects.ISVehicleTrainFrame0, objects.ISVehicleBusesFrame0, objects.ISVehicleTrucksFrame0,
		objects.ISVehicleTramsFrame0, objects.ISVehicleAircraftFrame0, objects.ISVehicleShipsFrame0}
	vehicleTabCounts = []int{8, 8, 8, 8, 8, 8}
)
