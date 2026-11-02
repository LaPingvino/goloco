package main

import (
	"github.com/LaPingvino/goloco/pkg/ui"
)

// listwindows.go — the right-toolbar list windows: the ported VehicleList,
// StationList and TownList (pkg/ui/townlist.go's table window).

// openVehiclesWindow lists save-loaded vehicle entities and runtime SimVehicles.
func (g *Game) openVehiclesWindow() { g.openVehicleList(0) }

// openVehicleList opens the ported VehicleList window on a tab (vehicle type).
func (g *Game) openVehicleList(tab int) {
	// VehicleList::open(getControllingId(), type): that company's vehicles.
	company, cname := g.w.ControllingCompany()
	var rows []ui.VehicleRow
	for _, v := range g.w.VehicleListRows() {
		if v.Type == tab && v.Owner == company {
			rows = append(rows, ui.VehicleRow{Name: v.Name, Profit: v.Profit, AgeYears: v.AgeYears, Reliability: v.Reliability})
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
	w := ui.NewVehicleList(g.sw, tab, rows, tc, wc)
	if cname != "" {
		w.Caption = cname + " - " + w.Caption // stringid_trains etc.
	}
	w.OnClose = func() { g.townList = nil }
	w.OnTab = g.openVehicleList
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
		rows = append(rows, ui.StationRow{Name: s.Name, Status: s.Status, Waiting: s.Waiting, Accepts: s.Accepts})
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
