package main

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// buildvehicle.go — the Build Vehicle window (BuildVehicle.cpp): buying,
// cloning, and the Vehicle window's build-new button.

// openBuildVehicle is BuildVehicle::open(tab, isTabId = true) for the
// controlling company (5P.f): the opening track tab, the list, at the screen's
// middle (PARTIAL: createWindow's free-spot placement).
func (g *Game) openBuildVehicle(tab int) { g.openBuildVehicleFor(tab, sim.EntityIDNull) }

// openBuildVehicleFor is openBuildVehicle with _buildTargetVehicle = target
// (BuildVehicle::openByVehicleId: the list limited to what fits that
// vehicle, and a row adds a car to it); EntityIDNull builds new vehicles.
// PARTIAL: the caption still says "Build …" for a target.
func (g *Game) openBuildVehicleFor(tab int, target uint16) {
	if g.r.ObjMgr == nil {
		return
	}
	m, om := &g.w.Map, g.r.ObjMgr
	company, _ := g.w.ControllingCompany()
	o := sim.DefaultBuildVehicleOptions(uint8(company))
	o.TargetHead = target
	tabs := m.BuildVehicleTrackTabs(om, uint8(tab), o)
	trackTab := m.BuildVehicleTrackTab(uint8(tab), tabs)
	var rows []ui.BVRow
	if len(tabs) > 0 {
		for _, id := range m.BuildableVehicles(om, uint8(tab), tabs[trackTab], o) {
			v := om.Vehicles[id]
			rows = append(rows, ui.BVRow{ID: id, Vehicle: v, Name: v.DisplayName})
		}
	}
	tc, wc := g.windowColours(true)
	var p, s uint8
	if company >= 0 && company < len(m.CompanyColourSchemes) {
		p, s = m.CompanyColourSchemes[company].Main[0], m.CompanyColourSchemes[company].Main[1]
	}
	sw, sh := g.sw, g.sh
	g.buildVehicle = &ui.BuildVehicleWindow{X: (sw - 500) / 2, Y: (sh - 400) / 2, W: 500, H: 400, TitleColour: tc, Colour: wc,
		Tab: tab, TrackTab: trackTab, Disabled: m.BuildVehicleDisabledTabs(uint8(company)), NumTrackTabs: len(tabs),
		Rows: rows, Hover: -1, Primary: p, Secondary: s,
		OnClose: func() { g.buildVehicle = nil }, OnTab: func(t int) { g.openBuildVehicleFor(t, target) },
		OnBuy: func(id uint16) { g.buyVehicle(tab, id, target) }}
}

// buyVehicle is onScrollMouseDown's purchase: the createVehicle command (no
// target vehicle) for the controlling company through doCommand (cash checked,
// paid under VehiclePurchases), then the window refreshed (sub_4B92A5: here
// opened afresh on its tab). PARTIAL: Vehicle::Details::open on the new
// vehicle (the Vehicle window, 5P.g — until then the bought vehicle waits
// unplaced, as upstream's does until placed, and shows in the vehicle list),
// the error window ("Can't build …" with the reason), Config::buildLockedVehicles
// (false).
func (g *Game) buyVehicle(tab int, id, target uint16) {
	if g.r.ObjMgr == nil {
		return
	}
	m := &g.w.Map
	c := g.playerCmd()
	var head uint16
	cost, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehiclePurchases, func(flags uint8) (int32, bool) {
		if target != sim.EntityIDNull { // gcArgs.vehicleId = _buildTargetVehicle: a car added to it
			head = target
			return m.AddCarToVehicle(c, flags, id, target, false)
		}
		return m.CreateVehicle(c, flags, id, false, &head)
	})
	if !ok {
		log.Printf("[Game] can't build vehicle %d", id)
		return
	}
	log.Printf("[Game] bought vehicle %d (head %d) for %d", id, head, cost)
	if target == sim.EntityIDNull { // only a new vehicle opens its window
		g.openVehicleDetails(head)
	} else if g.vehicleWin != nil {
		g.openVehicleWindow(head, 1) // the target's window shows its new car
	}
	g.openBuildVehicleFor(tab, target) // sub_4B92A5: the list again
}

// cloneVehicle is Vehicle::Details' cloneVehicle: the cloneVehicle command
// for the controlling company (paid under VehiclePurchases); done, the new
// vehicle's Details window opens; refused, the error window (cant_clone_vehicle
// and the reason).
func (g *Game) cloneVehicle(headID uint16) {
	if g.r.ObjMgr == nil {
		return
	}
	m := &g.w.Map
	c := g.playerCmd()
	var head uint16
	if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehiclePurchases, func(flags uint8) (int32, bool) {
		return m.CloneVehicle(c, flags, headID, &head)
	}); !ok {
		g.openError(i18n.Label(i18n.CantCloneVehicle), cmdErrorText(c))
		return
	}
	if head != sim.EntityIDNull {
		g.openVehicleDetails(head)
	}
}

// buildNewOpen is Vehicle Details' buildNew dropdown (onMouseDown): Modify
// vehicle (BuildVehicle::openByVehicleId) and Clone vehicle (cloneVehicle).
func (g *Game) buildNewOpen(headID uint16, x, y, w, h int) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	vt := int(m.Entities[tr.Head].VehicleType)
	items := []ui.DropdownItem{{Text: i18n.Label(i18n.DropdownModifyVehicle)}, {Text: i18n.Label(i18n.DropdownCloneVehicle)}}
	items[0].Selected = true
	d := ui.NewDropdownMenu(x, y+h, items)
	if g.vehicleWin != nil {
		d.Colour = g.vehicleWin.Colour
	}
	d.Highlighted = 0
	d.OnSelect = func(item int) {
		switch item {
		case 0:
			g.openBuildVehicleFor(vt, headID)
		case 1:
			g.cloneVehicle(headID)
		}
	}
	g.dropdown = d
}

// buildNewImage is Details::prepareDraw's buildNew image: the interface
// skin's build_additional_<vehicle type> (411 + type).
func (g *Game) buildNewImage(h *scenario.VehicleEntity) int {
	if g.skin() == nil || h.VehicleType > 5 {
		return 0
	}
	return int(g.skin().ImageOffset) + 411 + int(h.VehicleType)
}
