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
	w.OnRow = func(id int) { g.openStationWindow(id) } // onScrollMouseDown: Station::open
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
	w := ui.NewIndustryGrid(g.sw, choices, g.lastIndustryOption, tc, wc, func(id int) { g.lastIndustryOption = id })
	w.OnClose = func() { g.townList = nil }
	w.OnTab = func(t int) {
		if t == 0 {
			g.openIndustryList()
		}
	}
	g.townList = w
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
		w.Production = func(n int) (string, []uint8, int, int) {
			label := ""
			if c := int(produced[0]); c < len(g.r.ObjMgr.CargoObjects) && g.r.ObjMgr.CargoObjects[c] != nil {
				label = g.r.ObjMgr.CargoObjects[c].UnitsAndCargoName // producedCargoType[0] on both tabs
			}
			ind := &g.w.Map.Industries[idx]
			size := min(int(ind.HistorySize[n]), len(ind.History[n]))
			return label, ind.History[n][:size], int(g.w.Map.CurrentMonth), int(g.w.Map.CurrentYear)
		}
	}
	g.viewWin = w
}
