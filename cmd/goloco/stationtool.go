package main

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// stationtool.go — the construction window's Station tab on the map: the
// train station, road stop, dock and airport tools and their ghosts
// (Construction/StationTab.cpp).

// stationArgsAt is getTrainStationPlacementArgsFromCursor with the selected
// station type; ok false off track.
func (g *Game) stationArgsAt(sx, sy int) (sim.TrainStationPlacement, bool) {
	if g.r == nil || g.r.ObjMgr == nil {
		return sim.TrainStationPlacement{}, false
	}
	list := g.r.ObjMgr.AvailableCompatibleStations(g.buildObjID, false, int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.TrainStationPlacement{}, false
	}
	return g.w.TrainStationPlacementAt(sx, sy, uint8(pickSlot(list, g.consStation)))
}

func (g *Game) stationCmd() (*sim.Cmd, sim.G1Raw) {
	company, _ := g.w.ControllingCompany()
	g1, _ := g.r.ObjMgr.G1File.(sim.G1Raw)
	return &sim.Cmd{Company: uint8(company), Objects: g.r.ObjMgr}, g1
}

// stationToolUpdate is StationTab's onToolUpdateTrainStation (0x0049E421 side):
// off track, the ghost goes (onToolUpdateFail); on the same piece it stays;
// else the old ghost goes and a new one is placed (apply | noErrorWindow |
// noPayment | ghost), its cost kept; the catchment shown on the map and the
// ghost's accepted / produced cargo (calcAcceptedCargoTrainStationGhost).
func (g *Game) stationToolUpdate(sx, sy int) {
	if g.buildMode == buildModeAirport {
		g.airportToolUpdate(sx, sy)
		return
	}
	if g.buildMode == buildModeDock {
		g.dockToolUpdate(sx, sy)
		return
	}
	if g.consIsRoad() {
		g.roadStationToolUpdate(sx, sy)
		return
	}
	a, ok := g.stationArgsAt(sx, sy)
	if !ok {
		g.removeStationGhost()
		return
	}
	if gh := g.stationGhost; gh != nil && gh.X == a.X && gh.Y == a.Y && gh.Z == a.Z && gh.Rotation == a.Rotation &&
		gh.TrackID == a.TrackID && gh.Index == a.Index && gh.TrackObjectID == a.TrackObjectID {
		return
	}
	g.removeStationGhost()
	c, g1 := g.stationCmd()
	cost, ok := g.w.Map.CreateTrainStation(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, a, g1)
	g.stationGhostCost = cost
	if ok {
		g.stationGhost = &a
		m := &g.w.Map // constructingStationId and the ghost's catchment cargo
		g.stationGhostStation = m.LastAdjoiningStation
		x, y := m.LastAdjoiningStationX, m.LastAdjoiningStationY
		if x == -1 {
			x, y = a.X, a.Y
		}
		g.stationGhostAccepted, g.stationGhostProduced = m.CalcAcceptedCargoTrainStationGhost(g.stationGhostStation, x, y, 0xFFFFFFFF, g.r.ObjMgr)
		g.w.SetCatchmentDisplay(m.GhostCatchmentDisplay(g.stationGhostStation, x, y, g.r.ObjMgr)) // catchmentArea + setCatchmentDisplay / sub_491BF5 (flag_0)
	}
}

// removeStationGhost is removeConstructionGhosts' station part: the ghost
// removed as it was placed. Only with a station ghost showing
// (GhostVisibilityFlags::station) is the catchment display cleared — so the
// Station window's own catchment (showStationCatchment) stays.
func (g *Game) removeStationGhost() {
	if g.stationGhost != nil || g.roadStationGhost != nil || g.dockGhost != nil || g.airportGhost != nil {
		g.w.SetCatchmentDisplay(nil) // sub_491BC6 + resetMapSelectionFlag(catchmentArea)
		g.catchmentStation = -1      // the catchment shown is the ghost's from here
	}
	if rg := g.roadStationGhost; rg != nil { // stationGhostType & (1 << 7): RoadStationRemovalArgs
		g.roadStationGhost = nil
		c, g1 := g.stationCmd()
		g.w.Map.RemoveRoadStation(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost,
			sim.RoadStationRemoval{X: rg.X, Y: rg.Y, Z: rg.Z, Rotation: rg.Rotation, RoadID: rg.RoadID, Index: rg.Index, RoadObjectID: rg.RoadObjectID}, g1)
	}
	gh := g.stationGhost
	if gh == nil {
		return
	}
	g.stationGhost = nil
	c, g1 := g.stationCmd()
	g.w.Map.RemoveTrainStation(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost,
		sim.TrainStationRemoval{X: gh.X, Y: gh.Y, Z: gh.Z, Rotation: gh.Rotation, TrackID: gh.TrackID, Index: gh.Index, TrackObjectID: gh.TrackObjectID}, g1)
}

// stationToolUp is StationTab's onToolUpTrainStation (0x004A5390): the
// placement from the track under the cursor (getTrainStationPlacementArgsFromCursor),
// refused on another track type (wrong_type_of_track_road), then
// CreateTrainStation with apply, after removeConstructionGhosts; a failure
// opens the error window titled cant_build_pop3_string with the station's
// name and the command's reason (when the command sets one). PARTIAL: the
// construct sound.
func (g *Game) stationToolUp(sx, sy int) {
	if g.buildMode == buildModeAirport {
		g.airportToolUp(sx, sy)
		return
	}
	if g.buildMode == buildModeDock {
		g.dockToolUp(sx, sy)
		return
	}
	if g.consIsRoad() {
		g.roadStationToolUp(sx, sy)
		return
	}
	g.removeStationGhost()
	a, ok := g.stationArgsAt(sx, sy)
	if !ok {
		return
	}
	title := ""
	om := g.r.ObjMgr
	if list := om.AvailableCompatibleStations(g.buildObjID, false, int(g.w.CurrentYear)); len(list) > 0 {
		if id := pickSlot(list, g.consStation); id >= 0 && id < len(om.TrainStationObjects) && om.TrainStationObjects[id] != nil {
			title = ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), om.TrainStationObjects[id].Name)
		}
	}
	if int(a.TrackObjectID) != g.buildObjID {
		g.openError("", i18n.Label(i18n.WrongTypeOfTrackRoad))
		return
	}
	c, g1 := g.stationCmd()
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return g.w.Map.CreateTrainStation(c, f, a, g1) }); !ok { // GameCommands::doCommand: checked, then charged
		g.openError(title, cmdErrorText(c))
	}
}

// roadStationArgsAt is getRoadStationPlacementArgsFromCursor with the
// selected stop type.
func (g *Game) roadStationArgsAt(sx, sy int) (sim.RoadStationPlacement, int, bool) {
	if g.r == nil || g.r.ObjMgr == nil {
		return sim.RoadStationPlacement{}, -1, false
	}
	list := g.r.ObjMgr.AvailableCompatibleStations(g.buildObjID, true, int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.RoadStationPlacement{}, -1, false
	}
	typ := pickSlot(list, g.consStation)
	a, ok := g.w.RoadStationPlacementAt(sx, sy, uint8(typ))
	return a, typ, ok
}

// roadStationToolUpdate is StationTab's onToolUpdateRoadStation: as the train
// one, with the road stop ghost; the catchment's cargo filtered by the stop
// (a passenger stop: its cargo only; a freight bay: all but that).
func (g *Game) roadStationToolUpdate(sx, sy int) {
	a, typ, ok := g.roadStationArgsAt(sx, sy)
	if !ok {
		g.removeStationGhost()
		return
	}
	if gh := g.roadStationGhost; gh != nil && gh.X == a.X && gh.Y == a.Y && gh.Z == a.Z && gh.Rotation == a.Rotation &&
		gh.RoadID == a.RoadID && gh.Index == a.Index && gh.RoadObjectID == a.RoadObjectID {
		return
	}
	g.removeStationGhost()
	c, g1 := g.stationCmd()
	cost, ok := g.w.Map.CreateRoadStation(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, a, g1)
	g.stationGhostCost = cost
	if !ok {
		return
	}
	g.roadStationGhost = &a
	m := &g.w.Map
	g.stationGhostStation = m.LastAdjoiningStation
	x, y := m.LastAdjoiningStationX, m.LastAdjoiningStationY
	if x == -1 {
		x, y = a.X, a.Y
	}
	g.w.SetCatchmentDisplay(m.GhostCatchmentDisplay(g.stationGhostStation, x, y, g.r.ObjMgr))
	g.stationGhostAccepted, g.stationGhostProduced = m.CalcAcceptedCargoTrainStationGhost(g.stationGhostStation, x, y, sim.RoadStationCargoFilter(g.r.ObjMgr, typ), g.r.ObjMgr)
}

// roadStationToolUp is StationTab's onToolUpRoadStation (0x004A548F): after
// removeConstructionGhosts, CreateRoadStation with apply on the road under
// the cursor; a failure opens the error window titled cant_build_pop3_string
// with the stop's name. PARTIAL: the construct sound, the command's own
// error text.
func (g *Game) roadStationToolUp(sx, sy int) {
	g.removeStationGhost()
	a, typ, ok := g.roadStationArgsAt(sx, sy)
	if !ok {
		return
	}
	om := g.r.ObjMgr
	title := ""
	if typ >= 0 && typ < len(om.RoadStationObjects) && om.RoadStationObjects[typ] != nil {
		title = ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), om.RoadStationObjects[typ].Name)
	}
	c, g1 := g.stationCmd()
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return g.w.Map.CreateRoadStation(c, f, a, g1) }); !ok { // GameCommands::doCommand: checked, then charged
		g.openError(title, "")
	}
}

// ghostTabInfo is the Station tab's ghost part: cState.stationCost, the
// station it joins (or New Station) and the catchment's cargo lists.
func (g *Game) ghostTabInfo(st *ui.StationTabState) {
	om := g.r.ObjMgr
	st.Cost = int64(g.stationGhostCost) // cState.stationCost while a ghost stands
	st.Ghost = true
	st.GhostName = i18n.Label(i18n.NewStation)
	if id := g.stationGhostStation; id >= 0 && id < len(g.w.Map.StationNames) {
		st.GhostName = g.w.Map.StationName(id)
	}
	lines := func(bits uint32) []ui.CargoLine {
		var out []ui.CargoLine
		for i := 0; i < 32; i++ {
			if bits&(1<<i) == 0 || i >= len(om.CargoObjects) || om.CargoObjects[i] == nil {
				continue
			}
			c := om.CargoObjects[i]
			out = append(out, ui.CargoLine{Name: c.Name, Image: int(c.ImageOffset)})
		}
		return out
	}
	st.Accepted, st.Produced = lines(g.stationGhostAccepted), lines(g.stationGhostProduced)
}

// dockArgsAt is getDockPlacementArgsFromCursor: the surface or water under
// the cursor (getSurfaceOrWaterLocFromUi), the dock placed by
// getDockPlacementArgs with the selected dock.
func (g *Game) dockArgsAt(sx, sy int) (sim.PortPlacement, int, bool) {
	if g.r == nil || g.r.ObjMgr == nil {
		return sim.PortPlacement{}, -1, false
	}
	list := g.r.ObjMgr.AvailableDocks(int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.PortPlacement{}, -1, false
	}
	px, py, ok := g.w.SurfaceOrWaterAt(sx, sy)
	if !ok {
		return sim.PortPlacement{}, -1, false
	}
	typ := pickSlot(list, g.consStation)
	a, ok := g.w.Map.DockPlacementAt(px, py, uint8(typ), g.r.ObjMgr)
	return a, typ, ok
}

// dockToolUpdate is StationTab's onToolUpdateDock: the ghost dock follows
// the cursor (createPort with apply | noErrorWindow | noPayment | ghost),
// its cost, the catchment shown (sub_491D20) and its cargo
// (calcAcceptedCargoDockGhost). PARTIAL: the construction arrow and the
// selected 2×2 tiles (map selection).
func (g *Game) dockToolUpdate(sx, sy int) {
	a, _, ok := g.dockArgsAt(sx, sy)
	if !ok {
		g.removeStationGhost()
		return
	}
	if gh := g.dockGhost; gh != nil && gh.X == a.X && gh.Y == a.Y && gh.Z == a.Z && gh.Rotation == a.Rotation && gh.Type == a.Type {
		return
	}
	g.removeStationGhost()
	c, g1 := g.stationCmd()
	cost, ok := g.w.Map.CreatePort(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, a, g1)
	g.stationGhostCost = cost
	if !ok {
		return
	}
	g.dockGhost = &a
	m := &g.w.Map
	g.stationGhostStation = m.LastAdjoiningStation
	x, y := m.LastAdjoiningStationX, m.LastAdjoiningStationY
	if x == -1 {
		x, y = a.X, a.Y
	}
	g.w.SetCatchmentDisplay(m.GhostDockCatchmentDisplay(g.stationGhostStation, x, y, g.r.ObjMgr))
	g.stationGhostAccepted, g.stationGhostProduced = m.CalcAcceptedCargoDockGhost(g.stationGhostStation, x, y, 0xFFFFFFFF, g.r.ObjMgr)
}

// dockToolUp is StationTab's onToolUpDock (0x004A55AB): after
// removeConstructionGhosts, createPort with apply; a failure opens the error
// window titled cant_build_pop3_string with the dock's name.
func (g *Game) dockToolUp(sx, sy int) {
	g.removeStationGhost()
	a, typ, ok := g.dockArgsAt(sx, sy)
	if !ok {
		return
	}
	om := g.r.ObjMgr
	title := ""
	if typ >= 0 && typ < len(om.DockObjects) && om.DockObjects[typ] != nil {
		title = ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), om.DockObjects[typ].Name)
	}
	c, g1 := g.stationCmd()
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return g.w.Map.CreatePort(c, f, a, g1) }); !ok { // GameCommands::doCommand: checked, then charged
		g.openError(title, "")
	}
}

// airportArgsAt is getAirportPlacementArgsFromCursor: the surface or water
// under the cursor, the selected airport at the window's rotation.
func (g *Game) airportArgsAt(sx, sy int) (sim.AirportPlacement, int, bool) {
	if g.r == nil || g.r.ObjMgr == nil {
		return sim.AirportPlacement{}, -1, false
	}
	list := g.r.ObjMgr.AvailableAirports(int(g.w.CurrentYear))
	if len(list) == 0 {
		return sim.AirportPlacement{}, -1, false
	}
	px, py, ok := g.w.SurfaceOrWaterAt(sx, sy)
	if !ok {
		return sim.AirportPlacement{}, -1, false
	}
	typ := pickSlot(list, g.consStation)
	a, ok := g.w.Map.AirportPlacementAt(px, py, uint8(typ), g.airportRotation, g.r.ObjMgr)
	return a, typ, ok
}

// airportToolUpdate is StationTab's onToolUpdateAirport: the ghost airport
// follows the cursor, its cost, the catchment (sub_491C6F) and its cargo
// (calcAcceptedCargoAirportGhost). PARTIAL: the construction arrow and the
// selected tiles of its extents.
func (g *Game) airportToolUpdate(sx, sy int) {
	a, typ, ok := g.airportArgsAt(sx, sy)
	if !ok {
		g.removeStationGhost()
		return
	}
	if gh := g.airportGhost; gh != nil && *gh == a {
		return
	}
	g.removeStationGhost()
	c, g1 := g.stationCmd()
	cost, ok := g.w.Map.CreateAirport(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdNoPayment|sim.CmdGhost, a, g1)
	g.stationGhostCost = cost
	if !ok {
		return
	}
	g.airportGhost = &a
	m := &g.w.Map
	g.stationGhostStation = m.LastAdjoiningStation
	x, y := m.LastAdjoiningStationX, m.LastAdjoiningStationY
	if x == -1 {
		x, y = a.X, a.Y
	}
	ap := g.r.ObjMgr.AirportObjects[typ]
	g.w.SetCatchmentDisplay(m.GhostAirportCatchmentDisplay(g.stationGhostStation, ap, x, y, a.Rotation, g.r.ObjMgr))
	g.stationGhostAccepted, g.stationGhostProduced = m.CalcAcceptedCargoAirportGhost(g.stationGhostStation, ap, x, y, a.Rotation, 0xFFFFFFFF, g.r.ObjMgr)
}

// airportToolUp is StationTab's onToolUpAirport (0x004A5550): after
// removeConstructionGhosts, createAirport with apply; a failure opens the
// error window titled cant_build_pop3_string with the airport's name.
func (g *Game) airportToolUp(sx, sy int) {
	g.removeStationGhost()
	a, typ, ok := g.airportArgsAt(sx, sy)
	if !ok {
		return
	}
	om := g.r.ObjMgr
	title := ""
	if typ >= 0 && typ < len(om.AirportObjects) && om.AirportObjects[typ] != nil {
		title = ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), om.AirportObjects[typ].Name)
	}
	c, g1 := g.stationCmd()
	if _, ok := g.w.Map.DoCommandFor(c, sim.CmdApply, sim.ExpenditureConstruction, func(f uint8) (int32, bool) { return g.w.Map.CreateAirport(c, f, a, g1) }); !ok { // GameCommands::doCommand: checked, then charged
		g.openError(title, "")
	}
}
