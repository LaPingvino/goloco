package main

import (
	"os"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
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

// openVehicleListFor is VehicleList::open(company, type): the company's
// vehicle list (WindowType::vehicleList, numbered by the company) brought to
// the front on that tab, or opened.
func (g *Game) openVehicleListFor(company, tab int) { g.wins.Show(g.newVehicleList(company, tab)) }

// newVehicleList builds a company's vehicle list on a tab.
func (g *Game) newVehicleList(company, tab int) *ui.TownListWindow {
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
	w.WinType, w.Number = ui.WindowTypeVehicleList, company
	w.LighterFrame = true
	g.companySelect(w, company, func(id int) { g.turnList(w, ui.WindowTypeVehicleList, id, g.newVehicleList(id, tab)) })
	if cname != "" {
		w.Caption = cname + " - " + w.Caption // stringid_trains etc.
	}
	w.OnTab = func(t int) { g.openVehicleListFor(company, t) }
	g.listTabIcons(w, vehicleTabFrames, vehicleTabCounts, int(tc)) // recoloured to the company colour
	w.TabTooltips = labels(i18n.TooltipTrains, i18n.TooltipBuses, i18n.TooltipTrucks, i18n.TooltipTrams, i18n.TooltipAircraft, i18n.TooltipShips)
	w.TabDisabled = func(i int) bool { // disableUnavailableVehicleTypes: availableVehicles ^ 0x3F
		return company < len(g.w.Map.Companies) && g.w.Map.Companies[company].AvailableVehicles&(1<<i) == 0
	}
	w.Rebuild = func() *ui.TownListWindow { return g.newVehicleList(company, tab) }
	return w
}

// companySelect gives a list window the company select button (Vehicle and
// Station lists): the owner's face drawn on it (draw: competitor images by
// ownerEmotion, recoloured to the company's primary colour), and its dropdown
// of every company (populateCompanySelect: every non-empty slot), a choice
// turning the window to that company (onCompanyDropdown).
func (g *Game) companySelect(w *ui.TownListWindow, company int, open func(id int)) {
	w.Face, w.FaceColour = g.companyFace(company)
	for id := range g.w.CompanyNames {
		if g.w.CompanyNames[id] == 0 || id >= len(g.w.CompanyFaces) { // StringIds::empty: no company
			continue
		}
		face, col := g.companyFace(id)
		w.Companies = append(w.Companies, ui.CompanyChoice{ID: id, Name: g.w.CompanyName(id), Performance: int(g.w.CompanyFaces[id].Performance), Face: face, Colour: col})
	}
	w.OwnerID = company
	w.ScreenW, w.ScreenH = g.sw, g.sh
	w.OnCompany = open
	w.OpenDropdown = func(d *ui.DropdownMenu) { g.dropdown = d }
}

// turnList is a list's onCompanyDropdown: that company's list of the type,
// if open, comes to the front (bringToFront(type, company)); else this window
// turns to the company — rebuilt, where it stands.
func (g *Game) turnList(w *ui.TownListWindow, typ ui.WindowType, company int, rebuilt *ui.TownListWindow) {
	if open := g.wins.Find(typ, company); open != nil {
		g.wins.BringToFront(open)
		return
	}
	g.wins.Replace(w, rebuilt)
}

// openStationsWindow lists the controlling company's stations.
func (g *Game) openStationsWindow() {
	company, _ := g.w.ControllingCompany()
	g.openStationList(company, 0)
}

// stationTabMasks is StationList.cpp tabInformationByType's stationMask, by
// tab: all modes, rail, road, air, water (StationFlags transportMode*).
var stationTabMasks = [5]uint16{0xF, 1 << 0, 1 << 1, 1 << 2, 1 << 3}

// openStationList is StationList::open(company, tab): the company's stations of the tab's transport modes (populateStationList:
// owner, not flag_5 — not fully created —, flags & the tab's mask), captioned
// "{company} - All Stations" / "Rail Stations" …; a tab click reopens it on
// that tab.
func (g *Game) openStationList(company, tab int) { g.wins.Show(g.newStationList(company, tab)) }

// newStationList builds a company's station list (WindowType::stationList,
// numbered by the company) on a tab.
func (g *Game) newStationList(company, tab int) *ui.TownListWindow {
	cname := g.w.CompanyName(company)
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
	w.WinType, w.Number = ui.WindowTypeStationList, max(company, 0)
	w.LighterFrame = true
	if cname != "" {
		w.Caption = strings.Replace(i18n.Label(i18n.StringidAllStations+tab), "{STRINGID}", cname, 1) // "{STRINGID} - Rail Stations" …
	}
	w.CurrentTab = tab
	w.OnTab = func(t int) { g.openStationList(company, t) }
	g.companySelect(w, company, func(id int) { g.turnList(w, ui.WindowTypeStationList, id, g.newStationList(id, tab)) })
	w.OnRow = func(id int) { g.openStationWindow(id) }             // onScrollMouseDown: Station::open
	g.listTabIcons(w, stationTabFrames, stationTabCounts, int(tc)) // the company colour
	w.TabTooltips = labels(i18n.TooltipAllStations, i18n.TooltipRailStations, i18n.TooltipRoadStations, i18n.TooltipAirports, i18n.TooltipShipPorts)
	w.Rebuild = func() *ui.TownListWindow { return g.newStationList(company, tab) }
	return w
}

// openTownsWindow lists the scenario's towns with population and size.
func (g *Game) openTownsWindow() { g.wins.Show(g.newTownList()) }

// newTownList builds the town list (WindowType::townList; there is one).
func (g *Game) newTownList() *ui.TownListWindow {
	// The ported TownList window (pkg/ui/townlist.go).
	infos := g.w.TownList()
	rows := make([]ui.TownRow, len(infos))
	for i, t := range infos {
		rows[i] = ui.TownRow{ID: i, Name: t.Name, Type: t.Size, Size: t.SizeIndex, Population: int(t.Population), Stations: t.NumStations}
	}
	tc, wc := g.windowColours(false)
	w := ui.NewTownList(g.sw, rows, tc, wc)
	w.WinType = ui.WindowTypeTownList
	w.OnRow = func(id int) { // onScrollMouseDown: Town::open
		g.openTownWindow(id)
	}
	g.listTabIcons(w, townTabFrames, townTabCounts, -1)
	w.TabTooltips = labels(i18n.TooltipTownList, i18n.TooltipBuildTown, i18n.TooltipBuildBuildings, i18n.TooltipBuildMiscBuildings)
	// Outside the editor / sandbox the build-town, build-buildings and misc
	// buildings tabs are disabled (TownList.cpp open), so hidden.
	w.TabDisabled = func(i int) bool { return i > 0 }
	w.Rebuild = g.newTownList
	return w
}

// openTownWindow is Town::open: the Town window on its Town tab, its
// viewport on the town; the centre-view button moves the main view there.
// There is one per town (WindowType::town, numbered by the town): an open
// one comes to the front (bringToFront(type, number)).
func (g *Game) openTownWindow(id int) {
	if open := g.wins.Find(ui.WindowTypeTown, id); open != nil {
		g.wins.BringToFront(open)
		return
	}
	infos := g.w.TownList()
	if id < 0 || id >= len(infos) {
		return
	}
	t := infos[id]
	tc, wc := g.windowColours(false)
	w := ui.NewTownWindow(g.sw, g.sh, t.Name, t.Size, int(t.Population), tc, wc)
	w.WinType, w.Number = ui.WindowTypeTown, id
	w.View = func(vw, vh int) *gfx8.RenderTarget { return g.w.TownView(id, vw, vh) }
	w.Labels = func(dst *ebiten.Image) { g.w.TownViewLabels(id, dst) }
	w.OnCentreView = func() { g.w.CentreOnTown(id) }
	w.OnRename = func() { g.renameTown(id, w) }
	w.TabTooltips = labels(i18n.TooltipTown, i18n.TooltipPopulationGraph, i18n.TooltipTownRatingsEachCompany)
	w.RatingsTab = 2 // tab_company_ratings: each company the town rates
	w.Ratings = func() []ui.RatingLine {
		if id >= len(g.w.Map.Towns) {
			return nil
		}
		t := &g.w.Map.Towns[id]
		var lines []ui.RatingLine
		for c := range t.CompanyRatings {
			if t.CompaniesWithRating&(1<<c) == 0 {
				continue
			}
			r := (int(min(max(t.CompanyRatings[c], -1000), 1000)) + 1000) / 20
			rank := i18n.TownRatingAppalling
			switch {
			case r >= 70:
				rank = i18n.TownRatingExcellent
			case r >= 60:
				rank = i18n.TownRatingGood
			case r >= 50:
				rank = i18n.TownRatingAverage
			case r >= 25:
				rank = i18n.TownRatingPoor
			}
			lines = append(lines, ui.RatingLine{Name: g.w.CompanyName(c), Percent: r, Rank: i18n.Label(rank)})
		}
		return lines
	}
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
	g.wins.Show(w)
}

// openIndustryList is IndustryList::open (its list tab): every industry with
// its status and production; a row opens its Industry window.
func (g *Game) openIndustryList() { g.wins.Show(g.newIndustryList()) }

// newIndustryList builds the industry list's list tab (WindowType::
// industryList; there is one, its two tabs the same window).
func (g *Game) newIndustryList() *ui.TownListWindow {
	var rows []ui.IndustryListRow
	for _, r := range g.w.Map.IndustryList(g.r.ObjMgr) {
		rows = append(rows, ui.IndustryListRow{ID: r.ID, Name: r.Name, Status: r.Status, Produces: r.Produces,
			Transported: r.Transported, LastMonth: r.LastMonth, Unit: r.Unit})
	}
	tc, wc := g.windowColours(false)
	w := ui.NewIndustryList(g.sw, rows, tc, wc)
	w.WinType = ui.WindowTypeIndustryList
	w.OnRow = func(id int) { g.openIndustryWindow(id) } // onScrollMouseDown: Industry::open
	w.OnTab = func(t int) {
		if t == 1 {
			g.openNewIndustries()
		}
	}
	g.listTabIcons(w, industryTabFrames, industryTabCounts, -1)
	w.TabTooltips = labels(i18n.TooltipIndustriesList, i18n.TooltipFundNewIndustries)
	w.Rebuild = g.newIndustryList
	return w
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
			Image: obj.BuildingImageIDs, Colour: colour, Parts: parts, PartHeights: obj.PartHeights,
			Tooltip: sim.IndustryObjectTooltip(obj, om)})
	}
	// updateBuildableIndustries: rowHover is lastIndustryOption when it is
	// listed, else the first row (the tool then places that one).
	listed := false
	for _, c := range choices {
		listed = listed || c.ID == g.lastIndustryOption
	}
	if !listed && len(choices) > 0 {
		g.lastIndustryOption = choices[0].ID
	}
	tc, wc := g.windowColours(false)
	w := ui.NewIndustryGrid(g.sw, choices, g.lastIndustryOption, tc, wc, func(id int) {
		g.lastIndustryOption = id
		g.indTool.costOK = false // onScrollMouseDown: _industryPlacementCost = kFailure
	})
	w.PlacementCost = func() (int64, bool) { return int64(g.indTool.cost), g.indTool.ghost && g.indTool.costOK }
	w.WinType = ui.WindowTypeIndustryList
	w.OnClose = func() { g.industryToolAbort() } // onClose: toolCancel
	w.OnTab = func(t int) {
		if t == 0 {
			g.industryToolAbort()
			g.openIndustryList()
		}
	}
	g.listTabIcons(w, industryTabFrames, industryTabCounts, -1)
	w.TabTooltips = labels(i18n.TooltipIndustriesList, i18n.TooltipFundNewIndustries)
	g.wins.Show(w)
	g.industryToolStart(w)
}

// openIndustryWindow is IndustryWindow open (industry id): the window on its
// Industry tab, its viewport on the industry, its status getStatusString; the
// centre-view button moves the main view there.
func (g *Game) openIndustryWindow(id int) {
	if open := g.wins.Find(ui.WindowTypeIndustry, id); open != nil { // one per industry: bringToFront(type, number)
		g.wins.BringToFront(open)
		return
	}
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
	w.WinType, w.Number = ui.WindowTypeIndustry, id
	w.View = func(vw, vh int) *gfx8.RenderTarget { return g.w.IndustryView(idx, vw, vh) }
	w.Labels = func(dst *ebiten.Image) { g.w.IndustryViewLabels(idx, dst) }
	w.OnCentreView = func() { g.w.CentreOnIndustry(idx) }
	w.OnRename = func() { g.renameIndustry(idx, w) }
	w.TabTooltips = labels(i18n.TooltipIndustry, i18n.TooltipProductionGraph, i18n.TooltipProductionGraph, i18n.TooltipStatistics)
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
	g.wins.Show(w)
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

// renameTown is renameTownPrompt and textInput: title_town_name,
// prompt_type_new_town_name with the town's name, its current name; then
// renameTown for the controlling company (Miscellaneous); refused, the error
// window (error_cant_rename_town). The window shows the new name.
func (g *Game) renameTown(id int, win *ui.ViewWindow) {
	infos := g.w.TownList()
	if id < 0 || id >= len(infos) {
		return
	}
	tc, wc := g.windowColours(false)
	g.textInput = ui.OpenTextInput(i18n.Str(i18n.TitleTownName), ui.FillArgs(i18n.Str(i18n.PromptTypeNewTownName), infos[id].Name),
		infos[id].Name, 31, g.sw, g.sh, tc, wc, true, func(text string) {
			m := &g.w.Map
			c := g.playerCmd()
			g1, _ := g.r.ObjMgr.G1File.(sim.G1Raw)
			if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureMiscellaneous, func(f uint8) (int32, bool) {
				return m.RenameTown(c, f, id, text, g1)
			}); !ok {
				g.openError(i18n.Label(i18n.ErrorCantRenameTown), cmdErrorText(c))
				return
			}
			if infos := g.w.TownList(); id < len(infos) {
				win.Title = infos[id].Name
			}
			g.refreshList() // the open list repopulated
		})
}

// renameIndustry is renameIndustryPrompt and textInput: outside the editor
// only an industry the player's company founded (flag_04, a player owner);
// title_industry_name, prompt_enter_new_industry_name with the industry's
// town (as upstream), its current name; then renameIndustry
// (Miscellaneous); refused, error_cant_rename_industry.
func (g *Game) renameIndustry(idx int, win *ui.ViewWindow) {
	m := &g.w.Map
	if idx < 0 || idx >= len(m.Industries) || !m.IndustryFoundedByPlayer(idx) {
		return
	}
	name, _ := m.IndustryName(idx, g.r.ObjMgr)
	town := ""
	if infos := g.w.TownList(); int(m.Industries[idx].Town) < len(infos) {
		town = infos[m.Industries[idx].Town].Name
	}
	tc, wc := g.windowColours(false)
	g.textInput = ui.OpenTextInput(i18n.Str(i18n.TitleIndustryName), ui.FillArgs(i18n.Str(i18n.PromptEnterNewIndustryName), town),
		name, 31, g.sw, g.sh, tc, wc, true, func(text string) {
			c := g.playerCmd()
			if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureMiscellaneous, func(f uint8) (int32, bool) {
				return m.RenameIndustry(c, f, idx, text, g.r.ObjMgr)
			}); !ok {
				g.openError(i18n.Label(i18n.ErrorCantRenameIndustry), cmdErrorText(c))
				return
			}
			win.Title, _ = m.IndustryName(idx, g.r.ObjMgr)
		})
}

// refreshList is the lists' onUpdate repopulating: every open list (that can
// rebuild) built anew from the game's current data, keeping where it is, its
// scroll, its sort and its hover row (by its id); not while a mouse button is held (a
// scrollbar drag or a click in progress).
func (g *Game) refreshList() {
	if g.inMousePressed(ebiten.MouseButtonLeft) {
		return
	}
	for _, win := range append([]ui.Win(nil), g.wins.Windows...) {
		if old, ok := win.(*ui.TownListWindow); ok && old.Rebuild != nil {
			g.refreshListWindow(old)
		}
	}
}

func (g *Game) refreshListWindow(old *ui.TownListWindow) {
	w := old.Rebuild()
	if w == nil {
		return
	}
	g.wins.Replace(old, w)
	if w.Sort != nil && w.SortMode != old.SortMode {
		w.SortMode = old.SortMode
		w.Sort(old.SortMode)
	}
	w.Scroll, w.Hover = old.Scroll, old.Hover
	if old.Hover >= 0 && old.Hover < len(old.RowIDs) { // rowHover is the row's id: it follows the row
		w.Hover = -1
		for i, id := range w.RowIDs {
			if id == old.RowIDs[old.Hover] {
				w.Hover = i
				break
			}
		}
	}
}

// labels is the active language's strings of ids as stored, font and colour
// codes kept (widget tooltips: DrawTooltip reads them).
func labels(ids ...int) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = i18n.Str(id)
	}
	return out
}
