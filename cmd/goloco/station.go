package main

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"sort"
)

// station.go — the Station window (StationWindow.cpp) and its catchment.

// openStationWindow is Station::open: the Station window on its Station
// tab, its viewport on the station, the status bar its status; the
// centre-view button moves the main view there. The window takes the owning
// company's colour.
func (g *Game) openStationWindow(id int) {
	var info *sim.StationInfo
	for _, s := range g.w.StationList() {
		if s.ID == id {
			s := s
			info = &s
			break
		}
	}
	if info == nil {
		return
	}
	tc, wc := g.windowColours(false)
	if info.Owner >= 0 && info.Owner < len(g.w.CompanyColours) {
		tc = g.w.CompanyColours[info.Owner]
	}
	w := ui.NewStationWindow(g.sw, g.sh, info.Name, info.Status, tc, wc)
	w.Cargo, w.Accepted = g.stationCargoLines(id)
	if sk := g.skin(); sk != nil {
		w.SkinImage = int(sk.ImageOffset)
	}
	if info.Owner >= 0 && info.Owner < len(g.w.CompanyColours) {
		w.CompanyColour = g.w.CompanyColours[info.Owner]
	}
	if om := g.r.ObjMgr; om != nil && id < len(g.w.Map.StationNames) { // CargoRatings: every cargo with an origin
		for c, cs := range g.w.Map.StationNames[id].Cargo {
			if cs.Origin == 0xFFFF || c >= len(om.CargoObjects) || om.CargoObjects[c] == nil {
				continue
			}
			w.Ratings = append(w.Ratings, ui.StationRating{Name: om.CargoObjects[c].Name, Rating: int(cs.Rating)})
		}
	}
	// VehiclesStopping::populateVehicleList: every vehicle with an order at
	// this station, by type (which tabs show), each type's by name; rebuilt
	// about once a second while open (onUpdate).
	var byType [6][]ui.VehicleRow
	populated := -1
	populate := func() {
		byType = [6][]ui.VehicleRow{}
		w.VehicleTypes = [6]bool{}
		for _, v := range g.w.VehicleListRows() {
			if v.Type < 0 || v.Type >= 6 || !g.w.Map.VehicleStopsAt(v.HeadID, id) {
				continue
			}
			var cars []ui.TrainCarDraw
			for _, c := range g.w.TrainCars(v.HeadID) {
				cars = append(cars, ui.TrainCarDraw(c))
			}
			byType[v.Type] = append(byType[v.Type], ui.VehicleRow{HeadID: v.HeadID, Name: v.Name, Cars: cars, Status: v.Status, StatusRed: v.StatusRed})
			w.VehicleTypes[v.Type] = true
		}
		for t := range byType { // sortVehicleList: orderByName (strlogicalcmp)
			sort.SliceStable(byType[t], func(i, j int) bool { return core.StrLogicalCmp(byType[t][i].Name, byType[t][j].Name) < 0 })
		}
		populated = g.windowFrame
	}
	populate()
	w.Vehicles = func(t int) []ui.VehicleRow {
		if g.windowFrame-populated >= 60 {
			populate()
		}
		return byType[t]
	}
	w.OnVehicle = func(head uint16) { g.openVehicleDetails(head) } // onScrollMouseDown: Vehicle::Main::open
	w.View = func(vw, vh int) *gfx8.RenderTarget { return g.w.StationView(id, vw, vh) }
	w.Labels = func(dst *ebiten.Image) { g.w.StationViewLabels(id, dst) }
	w.OnCentreView = func() { g.w.CentreOnStation(id) }
	w.Catchment = g.catchmentStation == id
	w.OnCatchment = func() { // station_catchment: toggles this station's catchment
		if g.catchmentStation == id {
			g.showStationCatchment(-1)
		} else {
			g.showStationCatchment(id)
		}
		w.Catchment = g.catchmentStation == id
	}
	if own, ok := g.w.Map.StationOwner(id); ok && g.w.Map.IsPlayerCompany(int(own)) { // enableRenameByCaption
		w.OnRename = func() { g.renameStation(id, tc, wc) }
	}
	w.OnClose = func() { // Cargo::onClose: the catchment goes with the window
		if g.catchmentStation == id {
			g.showStationCatchment(-1)
		}
		g.stationWin = nil
	}
	g.stationWin = w
}

// showStationCatchment is StationWindow's showStationCatchment: the selected
// station (_lastSelectedStation, −1 none) changes; the old one's display is
// cleared (sub_491BC6, catchmentArea reset) and the new one's set
// (setCatchmentDisplay flag_0, catchmentArea). PARTIAL: Construction's
// sub_4A6FAC (the construction window's own ghost display) first.
func (g *Game) showStationCatchment(id int) {
	if id == g.catchmentStation {
		return
	}
	g.catchmentStation = id
	if id < 0 {
		g.w.SetCatchmentDisplay(nil)
		return
	}
	g.w.SetCatchmentDisplay(g.w.Map.StationCatchmentDisplay(id, g.r.ObjMgr))
}

// stationCargoLines is what the Station window's Cargo tab shows: for each
// cargo with a quantity, its sprite, the units drawn ((min(q, 400) + 9) /
// 10), "N <unit name>" (unitNameSingular for 1) and, from another station,
// "(en-route from <that station>)" split over the two lines as
// station_cargo_en_route_start / _end; and the accepted cargo's sprites.
func (g *Game) stationCargoLines(id int) (lines []ui.StationCargoLine, accepted []int) {
	m := &g.w.Map
	om := g.r.ObjMgr
	if om == nil || id < 0 || id >= len(m.StationNames) {
		return nil, nil
	}
	names := map[int]string{}
	for _, s := range g.w.StationList() {
		names[s.ID] = s.Name
	}
	st := &m.StationNames[id]
	for c := range st.Cargo {
		cs := st.Cargo[c]
		if c >= len(om.CargoObjects) || om.CargoObjects[c] == nil {
			continue
		}
		co := om.CargoObjects[c]
		if cs.Flags&1 != 0 { // isAccepted
			accepted = append(accepted, int(co.ImageOffset))
		}
		if cs.Quantity == 0 {
			continue
		}
		unit := co.UnitNamePlural
		if cs.Quantity == 1 {
			unit = co.UnitNameSingular
		}
		name := ui.FillArgs(ui.ObjectString(unit), ui.Comma32(int64(cs.Quantity)))
		l := ui.StationCargoLine{Sprite: int(co.ImageOffset), Units: (int(min(cs.Quantity, 400)) + 9) / 10}
		if int(cs.Origin) != id {
			l.Text = ui.StripCodes(ui.FillArgs(i18n.Str(i18n.StationCargoEnRouteStart), name))
			l.Origin = ui.StripCodes(ui.FillArgs(i18n.Str(i18n.StationCargoEnRouteEnd), names[int(cs.Origin)]))
		} else {
			l.Text = ui.StripCodes(ui.FillArgs(i18n.Str(i18n.StationCargo), name))
		}
		lines = append(lines, l)
	}
	return lines, accepted
}

// renameStation is renameStationPrompt and textInput: the text input
// (title_station_name, prompt_type_new_station_name with the station's
// town, the current name),
// then RenameStation for the controlling company (Miscellaneous); refused,
// the error window (error_cant_rename_station). The window shows the new
// name.
func (g *Game) renameStation(id int, titleColour, colour uint8) {
	m := &g.w.Map
	town := "" // args2: station->town
	if infos := g.w.TownList(); id < len(m.StationNames) && int(m.StationNames[id].Town) < len(infos) {
		town = infos[m.StationNames[id].Town].Name
	}
	g.textInput = ui.OpenTextInput(i18n.Str(i18n.TitleStationName), ui.FillArgs(i18n.Str(i18n.PromptTypeNewStationName), town),
		m.StationName(id), 31, g.sw, g.sh, titleColour, colour, true, func(text string) {
			c := g.playerCmd()
			if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureMiscellaneous, func(f uint8) (int32, bool) {
				return m.RenameStation(c, f, id, text, g.r.ObjMgr)
			}); !ok {
				g.openError(i18n.Label(i18n.ErrorCantRenameStation), cmdErrorText(c))
				return
			}
			if g.stationWin != nil {
				g.stationWin.Title = m.StationName(id)
			}
			g.w.StationsChanged()
		})
}
