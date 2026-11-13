package main

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// vehicle_route.go — the Vehicle window's Route tab: the order list, its
// buttons, and the tool adding stops from the map (Vehicle.cpp Route).

// routeLines is Route::drawScroll's rows: "no route defined" for an empty
// table, each order as orderString formats it (the station's name; the
// cargo's for unload / wait), the current one marked, then the end of the
// list. PARTIAL: the cargo sprite after unload / wait.
func (g *Game) routeLines(h *scenario.VehicleEntity) []ui.RouteLine {
	m := &g.w.Map
	var lines []ui.RouteLine
	if h.SizeOfOrderTable == 1 {
		lines = append(lines, ui.RouteLine{Text: i18n.Label(i18n.NoRouteDefined)})
	}
	format := [6]int{i18n.OrdersEnd, i18n.OrdersStopAt, i18n.OrdersRouteThrough, i18n.OrdersRouteThoughtWaypoint, i18n.OrdersUnloadAll, i18n.OrdersWaitForFullLoadOf}
	for _, o := range m.OrderTable(h.OrderTableOffset) {
		t := o.Type()
		arg := ""
		switch t {
		case sim.OrderStopAt, sim.OrderRouteThrough:
			arg = m.StationName(o.Station())
		case sim.OrderUnloadAll, sim.OrderWaitFor:
			if c := int(o.Cargo()); g.r.ObjMgr != nil && c < len(g.r.ObjMgr.CargoObjects) && g.r.ObjMgr.CargoObjects[c] != nil {
				arg = g.r.ObjMgr.CargoObjects[c].Name
			}
		}
		text := strings.Replace(i18n.Label(format[t]), "{STRINGID}", arg, 1)
		text = strings.TrimSpace(strings.Replace(text, "{SPRITE}", "", 1))
		lines = append(lines, ui.RouteLine{Text: text, Current: int(h.OrderTableOffset)+int(h.CurrentOrder) == o.Offset})
	}
	return append(lines, ui.RouteLine{Text: i18n.Label(i18n.OrdersEnd)})
}

// routeAction is Route::onMouseUp's order buttons: delete (the selected order,
// or the last), up / down (the selected one, the selection following it),
// skip (the current order), local / express (the vehicle's running mode,
// toggled only when the other one is on), reverse (the whole table, the
// current order following its order). PARTIAL: generateNumDisplayFrames after
// a change (the map's order numbers) — not ported for any of them.
func (g *Game) routeAction(headID uint16, action string) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	company, _ := g.w.ControllingCompany()
	c := g.playerCmd()
	orders := m.OrderTable(h.OrderTableOffset)
	rel := func(o sim.Order) uint16 { return uint16(o.Offset - int(h.OrderTableOffset)) }
	sel := g.routeSelOf(headID)
	switch action {
	case "local", "express": // widx::localMode / expressMode: the player's own only
		if company < 0 || h.Owner != uint8(company) {
			return
		}
		if express := m.Entities[tr.Veh1].Var48&sim.Flags48ExpressMode != 0; express != (action == "express") {
			m.VehicleChangeRunningMode(c, sim.CmdApply, headID, sim.RunningModeToggleLocalExpress)
		}
	case "delete": // onOrderDelete
		if h.SizeOfOrderTable <= 1 || len(orders) == 0 {
			return
		}
		target := orders[len(orders)-1]
		if sel >= 0 && sel < len(orders) {
			target = orders[sel]
		}
		m.VehicleOrderDelete(c, sim.CmdApply, headID, rel(target))
		if sel >= len(m.OrderTable(m.Entities[tr.Head].OrderTableOffset)) {
			g.setRouteSel(headID, -1) // no order selected any more
		}
	case "up", "down": // onOrderMove
		if h.SizeOfOrderTable <= 1 || sel < 0 || sel >= len(orders) {
			return
		}
		if action == "up" {
			if _, ok := m.VehicleOrderUp(c, sim.CmdApply, headID, rel(orders[sel])); ok && sel > 0 {
				g.setRouteSel(headID, sel-1)
			}
		} else if _, ok := m.VehicleOrderDown(c, sim.CmdApply, headID, rel(orders[sel])); ok && sel+1 < len(orders) {
			g.setRouteSel(headID, sel+1)
		}
	case "skip":
		m.VehicleOrderSkip(c, sim.CmdApply, headID)
	case "reverse": // orderReverseCommand: the selection stays (orderTableIndex untouched)
		m.VehicleOrderReverse(c, sim.CmdApply, headID)
	}
	g.refreshVehicleWindow(headID, 4)
}

// routeToolDown is Route::onToolDown for stops: the station under the cursor
// (World.RouteStopAt) added through addNewOrder — inserted before the selected
// order, else before the table's end; the selection moves down with the order
// it was on; plain track or road gives a route-through waypoint instead.
// PARTIAL: the waypoint sound; water waypoints.
func (g *Game) routeToolDown(sx, sy int) {
	m := &g.w.Map
	tr, err := m.TrainOf(g.routeHead)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	hit, ok := g.w.RouteAt(sx, sy, h.Mode)
	if !ok {
		return
	}
	order := sim.OrderStopAtRaw(hit.Station)
	if hit.Waypoint != nil {
		order = hit.Waypoint // route through (0x004B5160 / 0x004B5223)
	}
	g.addNewOrder(g.routeHead, order)
}

// addNewOrder is Route's addNewOrder: the order inserted before the selected
// one (or before End), the selection following it when the table grew;
// refused, the error window (orders_cant_insert and the reason).
// PARTIAL: generateNumDisplayFrames.
func (g *Game) addNewOrder(headID uint16, order []byte) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	offset := uint16(h.SizeOfOrderTable - 1)
	sel := g.routeSelOf(headID)
	if orders := m.OrderTable(h.OrderTableOffset); sel >= 0 && sel < len(orders) {
		offset = uint16(orders[sel].Offset - int(h.OrderTableOffset))
	}
	c := g.playerCmd()
	before := h.SizeOfOrderTable
	if _, ok := m.DoCommandFor(c, sim.CmdApply, 0, func(f uint8) (int32, bool) { return m.VehicleOrderInsert(c, f, headID, offset, order) }); !ok {
		g.openError(i18n.Label(i18n.OrdersCantInsert), cmdErrorText(c))
		return
	}
	if tr, err := m.TrainOf(headID); err == nil && m.Entities[tr.Head].SizeOfOrderTable != before && sel != -1 {
		g.setRouteSel(headID, sel+1)
	}
	g.refreshVehicleWindow(headID, 4)
}

// orderDropdown is Route's createOrderDropdown: one item per cargo the
// vehicle's cars carry (trainAcceptedCargoTypes), "Unload all {cargo}
// {sprite}" (orders_unload_all2) or "Wait for full load of …"
// (orders_wait_for_full_load_of2), shown as text at the button in the
// window's secondary colour, the first highlighted; the one picked is added
// as an OrderUnloadAll / OrderWaitFor of that cargo (onDropdown).
func (g *Game) orderDropdown(headID uint16, kind string, x, y, w, h int) {
	m := &g.w.Map
	om := g.r.ObjMgr
	tr, err := m.TrainOf(headID)
	if err != nil || om == nil {
		return
	}
	accepted := m.Entities[tr.Head].TrainAcceptedCargoTypes
	format, typ := i18n.Str(i18n.OrdersUnloadAll2), sim.OrderUnloadAll
	if kind == "wait" {
		format, typ = i18n.Str(i18n.OrdersWaitForFullLoadOf2), sim.OrderWaitFor
	}
	if i := strings.Index(format, "{SPRITE}"); i >= 0 {
		format = format[:i] // the inline cargo sprite: TrailSprite
	}
	var items []ui.DropdownItem
	var ids []uint8
	for id, co := range om.CargoObjects {
		if co == nil || id >= 32 || accepted&(1<<id) == 0 {
			continue
		}
		items = append(items, ui.DropdownItem{Text: ui.FillArgs(format, co.Name), TrailSprite: int(co.ImageOffset)})
		ids = append(ids, uint8(id))
	}
	if len(items) == 0 {
		return
	}
	colour := g.vehicleWinColour(headID)
	d := ui.NewDropdownShowText(x, y, w, h, colour, items, 10, g.sw, g.sh)
	d.Highlighted = 0
	d.OnSelect = func(item int) {
		if item >= 0 && item < len(ids) {
			g.addNewOrder(headID, []byte{typ | ids[item]<<3})
		}
	}
	g.dropdown = d
}

// updateRouteTool is Route::onUpdate's tool: the front-most window
// (WindowManager::isInFront) showing the Route tab of one of the controlling
// company's vehicles takes the tool (toolSet) unless it has it; it loses it
// when its window closes or leaves the tab (switchTab's toolCancel).
// PARTIAL: not while the pickup tool is on (upstream's toolSet cancels it).
func (g *Game) updateRouteTool() {
	if g.routeTool {
		if w := g.vehicleWin(g.routeHead); w == nil || w.Tab != 4 {
			g.routeTool = false
		}
	}
	if g.pickupOn {
		g.routeTool = false
		return
	}
	for _, win := range g.wins.Windows {
		w, ok := win.(*ui.VehicleWindow)
		if !ok || w.Tab != 4 || !w.Main.Own || !g.wins.IsInFront(w) {
			continue
		}
		if head := uint16(w.Number); !g.routeTool || g.routeHead != head {
			g.routeTool, g.routeHead = true, head
		}
	}
}
