package main

import (
	"github.com/LaPingvino/goloco/pkg/audio"
	"github.com/LaPingvino/goloco/pkg/objects"
	"log"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// buildvehicle.go — the Build Vehicle window (BuildVehicle.cpp): buying,
// cloning, and the Vehicle window's build-new button.

// openBuildVehicle is BuildVehicle::open(tab, isTabId = true) for the
// controlling company (5P.f): the opening track tab, the list; the window
// stack places it (createWindow).
func (g *Game) openBuildVehicle(tab int) { g.openBuildVehicleFor(tab, sim.EntityIDNull) }

// openBuildVehicleFor is openBuildVehicle with _buildTargetVehicle = target
// (BuildVehicle::openByVehicleId: the list limited to what fits that
// vehicle, and a row adds a car to it); EntityIDNull builds new vehicles.
// PARTIAL: the caption still says "Build …" for a target.
func (g *Game) openBuildVehicleFor(tab int, target uint16) {
	if t := g.buildTargetVehicle(); t != sim.EntityIDNull { // open's sub_4B92A5: the top editing window wins
		target = t
	}
	if bw := g.newBuildVehicleWindow(tab, target); bw != nil {
		g.showAuto(bw, 0, 0)
	}
}

// buildTargetVehicle is sub_4B92A5's _buildTargetVehicle: the front-most
// Vehicle window on its Details tab whose vehicle the controlling company
// owns (getTopEditingVehicleWindow), EntityIDNull when there is none. So a
// bought locomotive's Details window, opened by the purchase, makes the next
// purchase a car added to it.
func (g *Game) buildTargetVehicle() uint16 {
	company, _ := g.w.ControllingCompany()
	for i := len(g.wins.Windows) - 1; i >= 0; i-- {
		vw, ok := g.wins.Windows[i].(*ui.VehicleWindow)
		if !ok || vw.Tab != 1 {
			continue
		}
		tr, err := g.w.Map.TrainOf(uint16(vw.Number))
		if err != nil || int(g.w.Map.Entities[tr.Head].Owner) != company {
			continue
		}
		return uint16(vw.Number)
	}
	return sim.EntityIDNull
}

// refreshBuildVehicle is sub_4B92A5 on the open Build Vehicle window: its
// target and list again, on its tab.
func (g *Game) refreshBuildVehicle() {
	company, _ := g.w.ControllingCompany()
	old, ok := g.wins.Find(ui.WindowTypeBuildVehicle, company).(*ui.BuildVehicleWindow)
	if !ok {
		return
	}
	if bw := g.newBuildVehicleWindow(old.Tab, g.buildTargetVehicle()); bw != nil {
		g.wins.Refresh(bw)
	}
}

// updateBuildVehicle is onUpdate's refresh: the window's target no longer
// the top editing Vehicle window (closed, or another tab) gets sub_4B92A5.
func (g *Game) updateBuildVehicle() {
	company, _ := g.w.ControllingCompany()
	if bw, ok := g.wins.Find(ui.WindowTypeBuildVehicle, company).(*ui.BuildVehicleWindow); ok && bw.Target != g.buildTargetVehicle() {
		g.refreshBuildVehicle()
	}
}

// newBuildVehicleWindow builds the window (WindowType::buildVehicle,
// numbered by the controlling company) on a tab for a target.
func (g *Game) newBuildVehicleWindow(tab int, target uint16) *ui.BuildVehicleWindow {
	if g.r.ObjMgr == nil {
		return nil
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
	bw := &ui.BuildVehicleWindow{WinBase: ui.WinBase{WinType: ui.WindowTypeBuildVehicle, Number: company, W: 500, H: 400}, TitleColour: tc, Colour: wc,
		Tab: tab, TrackTab: trackTab, Target: target, Disabled: m.BuildVehicleDisabledTabs(uint8(company)), NumTrackTabs: len(tabs),
		Rows: rows, Hover: -1, Primary: p, Secondary: s,
		OnTab: func(t int) { g.wins.Quietly(func() { g.openBuildVehicleFor(t, target) }) }, // switchTab
		TabIcon: func(i int) (int, int) { // drawTransportTypeTabs: frame (frameNo / 2) & 0xF on the active tab
			sk := g.skin()
			if sk == nil || i >= len(objects.ISBuildVehicleFrame0) {
				return -1, -1
			}
			img := int(sk.ImageOffset) + objects.ISBuildVehicleFrame0[i]
			if i == tab {
				img += (g.windowFrame / 2) & 0xF
			}
			return img, int(p)
		},
		TrackTabIcon: func(i int) (int, int) { // drawTrackTypeTabs
			if i >= len(tabs) {
				return -1, -1
			}
			switch t := tabs[i]; {
			case t == 0xFF: // air / water: the toolbar's airport or ship-port image, not recoloured
				if sk := g.skin(); sk != nil {
					if tab == 4 {
						return int(sk.ImageOffset) + objects.ISToolbarMenuAirport, -1
					}
					return int(sk.ImageOffset) + objects.ISToolbarMenuShipPort, -1
				}
			case t&(1<<7) != 0: // a road: its image, animated (frameNo / 4) & 0x1F when active
				if r := om.GetRoadObjectByIndex(int(t &^ (1 << 7))); r != nil {
					img := int(r.Image)
					if i == trackTab {
						img += (g.windowFrame / 4) & 0x1F
					}
					return img, int(p)
				}
			default: // a track: kUiPreviewImage0, animated (frameNo / 4) & 0xF when active
				if tr := om.GetTrackObjectByIndex(int(t)); tr != nil {
					img := int(tr.Image)
					if i == trackTab {
						img += (g.windowFrame / 4) & 0xF
					}
					return img, int(p)
				}
			}
			return -1, -1
		},
		OnBuy: func(id uint16) { g.buyVehicle(tab, id, g.buildTargetVehicle()) }}
	bw.OnResize() // resizable, its bounds (callOnResize)
	return bw
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
	company, _ := g.w.ControllingCompany()
	if bw, ok := g.wins.Find(ui.WindowTypeBuildVehicle, company).(*ui.BuildVehicleWindow); ok {
		g.playSound(audio.SoundClickDown, audio.ChannelUI, bw.W/2+bw.X) // onScrollMouseDown: the window's middle
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
	} else {
		g.refreshVehicleWindow(head, 1) // the target's window shows its new car
	}
	g.refreshBuildVehicle() // sub_4B92A5: the new vehicle's Details window is the target now

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
	d.Colour = g.vehicleWinColour(headID)
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
