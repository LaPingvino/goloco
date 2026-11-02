package main

import (
	"github.com/LaPingvino/goloco/pkg/ui"
	"os"
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
		w.Caption = cname + " - All Stations" // stringid_all_stations
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
