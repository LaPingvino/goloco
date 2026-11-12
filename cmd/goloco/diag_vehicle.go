package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// diag_vehicle.go — GOLOCO_DIAG_BUILDVEHICLE: the scenario dispatch and the
// vehicle scenarios (buying, placing, the Vehicle window's tabs, ghosts of
// the pickup tool). The map and station-tool scenarios are in diag_map.go.

// diagNeedsScene are the GOLOCO_DIAG_BUILDVEHICLE scenarios that wait for a
// painted scene (World.SceneReady) before they run.
var diagNeedsScene = map[string]bool{"routeadd": true, "routewaypoint": true, "express": true, "roadghost": true,
	"shipghost": true, "shipsail": true, "planefly": true, "planeghost": true, "newindustry": true,
	"newindustrybuild": true, "waves": true, "playable": true, "towns": true, "buildstation": true,
	"stationtool": true, "stationghost": true, "crash": true}

// diagScenario runs GOLOCO_DIAG_BUILDVEHICLE's scenario v (once): a number is
// the Build Vehicle window on that tab; the names set a situation up through
// the game's own commands and tools, for a shot or the log.
func (g *Game) diagScenario(v string) {
	switch v {
	case "buildstation":
		g.diagBuildStation()
	case "roadstationtool", "roadstationghost":
		g.diagRoadStationTool(v)
	case "stationtool", "stationghost":
		g.diagStationTool(v)
	case "towns":
		g.diagTownsView()
	case "playable":
		g.diagPlayable()
	case "drag":
		g.diagDragCar()
	case "sell":
		g.diagSell()
	case "rename":
		g.diagRename()
	case "refit":
		g.diagRefit()
	case "finances", "cargo":
		g.diagVehicleTab(v)
	case "moneyeffect":
		g.diagMoneyEffect()
	case "crash":
		g.diagCrash()
	case "airportghost":
		g.diagAirportGhost()
	case "dockghost", "docktool":
		g.diagDockTool(v)
	case "portmenu":
		g.diagPortMenu()
	case "menu":
		g.diagBuildVehiclesMenu()
	case "routeadd", "routewaypoint":
		g.diagRouteAdd(v)
	case "waves":
		g.diagWaves()
	case "newindustry", "newindustrybuild":
		g.diagNewIndustry(v)
	case "planeghost":
		g.diagPlaneGhost()
	case "planefly":
		g.diagPlaneFly()
	case "shipsail":
		g.diagShipSail()
	case "shipghost":
		g.diagShipGhost()
	case "roadghost":
		g.diagRoadGhost()
	case "express":
		g.diagExpress()
	case "route":
		g.diagRoute()
	case "main":
		g.diagVehicleMain()
	case "buy", "place":
		g.diagBuyPlace(v)
	default: // the Build Vehicle window on tab N
		tab, _ := strconv.Atoi(v)
		g.openBuildVehicle(tab)
	}
}

// diagPlayable (GOLOCO_DIAG_BUILDVEHICLE=playable) — DIAGNOSTIC ONLY: 5P end
// to end through the game's commands, as the player — a stopped rail train
// with two stops lifted, a copy bought (same cars), placed where it stood, its
// orders given, started; the view follows it, its window on Main
func (g *Game) diagPlayable() {
	m := &g.w.Map
	om := g.r.ObjMgr
	trains, _ := m.Trains()
	for _, tr := range trains {
		oh := m.Entities[tr.Head]
		if oh.Mode != 0 || oh.TileX == -1 || len(tr.Cars) == 0 {
			continue
		}
		var orders [][]byte
		stops := 0
		for _, o := range m.OrderRing(oh.OrderTableOffset, 0) {
			if o.Type() == sim.OrderStopAt {
				stops++
			}
			orders = append(orders, append([]byte(nil), m.Orders[o.Offset:o.Offset+sim.OrderSize(int(o.Type()))]...))
		}
		c := &sim.Cmd{Company: oh.Owner, Objects: om}
		if stops < 2 {
			continue
		}
		v2 := m.Entities[tr.Veh2] // where it stands, before it is lifted
		place := sim.VehiclePlacement{X: int32(v2.TileX), Y: int32(v2.TileY), Z: int32(v2.TileBaseZ) * 4, TrackAndDirection: v2.TrackAndDirection, TrackProgress: v2.SubPosition}
		if _, ok := m.VehiclePickup(c, sim.CmdApply, oh.ID); !ok {
			continue // moving: can't be modified
		}
		m.ControllingCompany, m.PlayerCompanies[0] = oh.Owner, oh.Owner
		m.Companies[oh.Owner].Cash += 5_000_000 // funded: the title's companies are nearly broke
		var cars []uint16
		for _, car := range tr.Cars {
			cars = append(cars, m.Entities[car[0].Body].ObjectID)
		}
		var h uint16
		_, okBuy := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehiclePurchases, func(f uint8) (int32, bool) { return m.CreateVehicle(c, f, cars[0], true, &h) })
		for _, car := range cars[1:] {
			m.AddCarToVehicle(c, sim.CmdApply, car, h, true)
		}
		place.Head = h
		_, okPlace := m.VehiclePlace(c, sim.CmdApply, place)
		for _, raw := range orders {
			m.VehicleOrderInsert(c, sim.CmdApply, h, m.Entities[func() int { t, _ := m.TrainOf(h); return t.Head }()].SizeOfOrderTable-1, raw)
		}
		_, okStart := m.VehicleChangeRunningMode(c, sim.CmdApply, h, sim.RunningModeStart)
		g.inTitleScreen = false
		g.diagShip, g.diagFollow = h, true
		g.w.CentreOnVehicle(h)
		log.Printf("[Diag] playable: replaced train %d by bought %d (%d cars, %d orders): bought %v placed %v started %v", oh.ID, h, len(cars), len(orders), okBuy, okPlace, okStart)
		g.openVehicleWindow(h, 0)
		if g.vehicleWin != nil {
			g.vehicleWin.X, g.vehicleWin.Y = 8, 40
		}
		break
	}
}

// diagDragCar (GOLOCO_DIAG_BUILDVEHICLE=drag) — DIAGNOSTIC ONLY: a stopped
// saved train's last car dragged over its second row (GOLOCO_DIAG_DRAG_DROP:
// dropped there)
func (g *Game) diagDragCar() {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.VehicleType != 0 || len(tr.Cars) < 3 {
			continue
		}
		m.ControllingCompany = h.Owner
		m.Entities[tr.Veh2].CurrentSpeed = 0
		m.Entities[tr.Head].VehicleFlags |= sim.VehicleFlagCommandStop
		g.inTitleScreen = false
		g.openVehicleWindow(h.ID, 1)
		w := g.vehicleWin
		order := func() []uint16 {
			t, _ := m.TrainOf(h.ID)
			var ids []uint16
			for _, car := range t.Cars {
				ids = append(ids, m.Entities[car[0].Front].ID)
			}
			return ids
		}
		before := order()
		listY := w.Y + 44 + 1 // the car list's inside (Details carList top 44)
		w.HandleClick(w.X+60, listY+(len(tr.Cars)-1)*22+11)
		w.DragTo(w.X+80, listY+22+11)
		w.MouseOver(w.X+80, listY+22+11) // scrollDrag: the drop line
		log.Printf("[Diag] drag: vehicle %d, dragging %v car %d", h.ID, w.Dragging, w.DragCar)
		if os.Getenv("GOLOCO_DIAG_DRAG_DROP") != "" {
			w.DragEnd(w.X+80, listY+22+11)
			if g.prompt != nil && g.prompt.OnOK != nil { // the cargo question: OK
				g.prompt.OnOK()
				g.prompt = nil
			}
			log.Printf("[Diag] drag: %v → %v", before, order())
		}
		break
	}
}

// diagSell (GOLOCO_DIAG_BUILDVEHICLE=sell) — DIAGNOSTIC ONLY: as a saved
// train's company, a vehicle bought, its Details tab; GOLOCO_DIAG_SELL_OK: the
// bin pressed
func (g *Game) diagSell() {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.VehicleType != 0 || len(tr.Cars) == 0 {
			continue
		}
		m.ControllingCompany = h.Owner
		c := &sim.Cmd{Company: h.Owner, Objects: g.r.ObjMgr}
		var nh uint16
		if _, ok := m.CreateVehicle(c, sim.CmdApply, m.Entities[tr.Cars[0][0].Body].ObjectID, true, &nh); !ok {
			continue
		}
		g.inTitleScreen = false
		g.openVehicleWindow(nh, 1)
		cash := m.Companies[h.Owner].Cash
		if os.Getenv("GOLOCO_DIAG_SELL_OK") != "" {
			g.vehicleWin.HandleClick(g.vehicleWin.X+g.vehicleWin.W-14, g.vehicleWin.Y+108) // the bin (right bar, 96–119)
			log.Printf("[Diag] sell: vehicle %d gone %v, window closed %v, cash %d → %d", nh, m.Entities != nil && func() bool { _, err := m.TrainOf(nh); return err != nil }(), g.vehicleWin == nil, cash, m.Companies[h.Owner].Cash)
		} else {
			log.Printf("[Diag] sell: vehicle %d, Details open", nh)
		}
		break
	}
}

// diagRename (GOLOCO_DIAG_BUILDVEHICLE=rename) — DIAGNOSTIC ONLY: as a saved
// train's company, its window's caption clicked, " Express" typed
func (g *Game) diagRename() {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		if h := m.Entities[tr.Head]; h.VehicleType == 0 {
			m.ControllingCompany = h.Owner
			g.inTitleScreen = false
			g.openVehicleWindow(h.ID, 3)
			g.vehicleWin.HandleClick(g.vehicleWin.X+40, g.vehicleWin.Y+6)
			if g.textInput != nil {
				g.textInput.Keys([]rune(" Express"), nil, false, false)
			}
			log.Printf("[Diag] rename: vehicle %d, text input open %v", h.ID, g.textInput != nil)
			if os.Getenv("GOLOCO_DIAG_RENAME_OK") != "" && g.textInput != nil { // Enter: OK
				g.textInput.Keys(nil, nil, true, false)
				g.textInput = nil
				nh := m.Entities[tr.Head]
				log.Printf("[Diag] rename: now %q, window %q", m.VehicleName(nh.Name, int(nh.OrdinalNumber)), g.vehicleWin.Title)
			}
			break
		}
	}
}

// diagRefit (GOLOCO_DIAG_BUILDVEHICLE=refit) — DIAGNOSTIC ONLY: as a saved
// train's company, a refittable vehicle bought, its Cargo tab, the refit
// dropdown
func (g *Game) diagRefit() {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		if h.VehicleType != 0 {
			continue
		}
		m.ControllingCompany, m.PlayerCompanies[0] = h.Owner, h.Owner // the diag plays as it (canRefit: a player's)
		c := &sim.Cmd{Company: h.Owner, Objects: g.r.ObjMgr}
		for id, ob := range g.r.ObjMgr.Vehicles {
			var nh uint16
			if ob == nil || ob.Flags&(1<<14) == 0 {
				continue
			}
			if _, ok := m.CreateVehicle(c, sim.CmdApply, uint16(id), true, &nh); !ok {
				continue
			}
			g.inTitleScreen = false
			g.openVehicleWindow(nh, 2)
			x, y := g.vehicleWin.X+g.vehicleWin.W-14, g.vehicleWin.Y+56 // the refit button (right bar, 44–67)
			g.vehicleWin.HandleClick(x, y)
			log.Printf("[Diag] refit: vehicle %d (%s), dropdown %v", nh, ob.DisplayName, g.dropdown != nil)
			break
		}
		break
	}
}

// diagVehicleTab (GOLOCO_DIAG_BUILDVEHICLE=finances / cargo) — DIAGNOSTIC
// ONLY: the first saved train with a last income (cargo: holding cargo), that
// tab
func (g *Game) diagVehicleTab(v string) {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		loaded := false
		for _, car := range tr.Cars {
			loaded = loaded || m.Entities[car[0].Body].Cargo.Qty != 0
		}
		if h.VehicleType == 0 && m.Entities[tr.Veh1].LastIncome.Day != -1 && (v == "finances" || loaded && len(tr.Cars) > 2) {
			g.inTitleScreen = false
			g.openVehicleWindow(h.ID, map[string]int{"finances": 3, "cargo": 2}[v])
			g.vehicleWin.MouseOver(g.vehicleWin.X+60, g.vehicleWin.Y+44+1+22+11) // the second row hovered
			log.Printf("[Diag] %s: vehicle %d", v, h.ID)
			break
		}
	}
}

// diagBuildVehiclesMenu (GOLOCO_DIAG_BUILDVEHICLE=menu) — DIAGNOSTIC ONLY: the
// toolbar's Build Vehicles menu instead
func (g *Game) diagBuildVehiclesMenu() {
	for i := range g.toolbar.Buttons {
		if g.toolbar.Buttons[i].Tooltip == "Build Vehicles" {
			g.handleToolbarButton(i)
		}
	}
}

// diagRouteAdd (GOLOCO_DIAG_BUILDVEHICLE=routeadd / routewaypoint) —
// DIAGNOSTIC ONLY: the Route tool: click the first station (or plain track: a
// waypoint) found on screen
func (g *Game) diagRouteAdd(v string) {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		if h := m.Entities[tr.Head]; h.TileX != -1 && h.SizeOfOrderTable > 1 {
			m.ControllingCompany = h.Owner // the diag plays as the train's company (the title map's are AI)
			g.openVehicleWindow(h.ID, 4)
			log.Printf("[Diag] route tool on: %v (company %d)", g.routeTool, h.Owner)
			g.w.CentreOnVehicle(h.ID) // its stations in view
			done := false
			for sy := 40; sy < g.sh-40 && !done; sy += 8 {
				for sx := 0; sx < g.sw && !done; sx += 8 {
					hit, ok := g.w.RouteAt(sx, sy, h.Mode)
					if ok && (hit.Waypoint != nil) == (v == "routewaypoint") {
						log.Printf("[Diag] %+v under %d,%d — clicking", hit, sx, sy)
						g.routeToolDown(sx, sy)
						done = true
					}
				}
			}
			for _, l := range g.vehicleWin.Route {
				log.Printf("[Diag] route after the click: %q", l.Text)
			}
			break
		}
	}
}

// diagPlaneGhost (GOLOCO_DIAG_BUILDVEHICLE=planeghost) — DIAGNOSTIC ONLY: buy
// an aircraft for an airport's owner, its pickup tool on, hover over the
// airport until a ghost stands at a terminal
func (g *Game) diagPlaneGhost() {
	m := &g.w.Map
	om := g.r.ObjMgr
	var sid uint16 = 0xFFFF
	for i, st := range m.StationNames {
		if st.Name != 0 && st.Flags&(1<<6) != 0 && st.Owner < 16 {
			sid = uint16(i)
			break
		}
	}
	var planeObj uint16 = 0xFFFF
	for i, v := range om.Vehicles {
		if v != nil && v.Mode == objects.TransportModeAir {
			planeObj = uint16(i)
			break
		}
	}
	if sid == 0xFFFF || planeObj == 0xFFFF {
		log.Printf("[Diag] planeghost: no airport or aircraft object")
	} else {
		st := m.StationNames[sid]
		m.ControllingCompany = st.Owner
		g.inTitleScreen = false
		g.w.CenterOn(fmt.Sprintf("tile:%d,%d", int(st.AirportStartPos[0])/32+2, int(st.AirportStartPos[1])/32+2))
		var head uint16
		if _, ok := m.CreateVehicle(&sim.Cmd{Company: st.Owner, Objects: om}, sim.CmdApply, planeObj, true, &head); !ok {
			log.Printf("[Diag] planeghost: buying aircraft %d refused", planeObj)
		} else {
			g.openVehicleWindow(head, 1)
			if x, y, ok := g.vehicleWin.PickupButton(); ok {
				g.vehicleWin.HandleClick(x, y)
			}
			if g.vehicleWin != nil {
				g.vehicleWin.X, g.vehicleWin.Y = 4, 40
			}
			for r := 0; r < 200 && !m.VehicleIsGhost(head); r += 4 {
				for _, o := range [][2]int{{r, r / 2}, {-r, r / 2}, {r, -r / 2}, {-r, -r / 2}, {0, r}, {0, -r}} {
					hx, hy := g.sw/2+o[0], g.sh/2+o[1]
					if g.overWindow(hx, hy) {
						continue
					}
					g.pickupToolUpdate(hx, hy)
					if m.VehicleIsGhost(head) {
						g.diagMouse = &[2]int{hx, hy}
						break
					}
				}
			}
			trh, _ := m.TrainOf(head)
			h := m.Entities[trh.Head]
			log.Printf("[Diag] planeghost: station %d, aircraft %d (object %d); pickup tool %v; a ghost now %v at %d,%d,%d edge %d; mouse %v",
				sid, head, planeObj, g.pickupOn, m.VehicleIsGhost(head), h.PosX, h.PosY, h.PosZ, h.AirportMovementEdge, g.diagMouse)
		}
	}
}

// diagPlaneFly (GOLOCO_DIAG_BUILDVEHICLE=planefly) — DIAGNOSTIC ONLY: follow
// the title map's aircraft at an airport (its status logged)
func (g *Game) diagPlaneFly() {
	g.inTitleScreen = false
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		if h := m.Entities[tr.Head]; h.Mode == 2 && h.TileX != -1 && h.StationID != 0xFFFF {
			g.diagShip, g.diagFollow = h.ID, true
			g.w.CentreOnVehicle(h.ID)
			log.Printf("[Diag] planefly: aircraft %d at station %d", h.ID, h.StationID)
			break
		}
	}
}

// diagShipSail (GOLOCO_DIAG_BUILDVEHICLE=shipsail) — DIAGNOSTIC ONLY: two
// docks (at the water's height, in their stations' tiles) on one lake, a ship
// placed at the first with "stop at" the second, started — watch it sail
func (g *Game) diagShipSail() {
	m := &g.w.Map
	om := g.r.ObjMgr
	var shipObj uint16 = 0xFFFF
	for i, v := range om.Vehicles {
		if v != nil && v.Mode == objects.TransportModeWater {
			shipObj = uint16(i)
			break
		}
	}
	var company uint8 = 0xFF
	var stations []uint16
	for i, st := range m.StationNames {
		if st.Name != 0 && st.Owner < 16 && (company == 0xFF || st.Owner == company) && int(st.TileSize) < len(st.Tiles) {
			company = st.Owner
			stations = append(stations, uint16(i))
			if len(stations) == 2 {
				break
			}
		}
	}
	if shipObj == 0xFFFF || len(stations) < 2 || len(om.DockObjects) == 0 || om.DockObjects[0] == nil {
		log.Printf("[Diag] shipsail: no ship object, two stations of a company, or dock object")
	} else {
		d := om.DockObjects[0]
		ox, oy := sim.RotateOffset(d.BoatPosition[0], d.BoatPosition[1], 0)
		var docks [][3]int32
		var level uint8
		for y := 4; y < m.Height-4 && len(docks) < 2; y++ {
			for x := 4; x < m.Width-4 && len(docks) < 2; x++ {
				t := m.TileAt(x, y)
				bx, by := int32(x)*32+int32(ox)+32, int32(y)*32+int32(oy)+32
				bt := m.TileAt(int(bx>>5), int(by>>5))
				if t == nil || bt == nil || t.WaterLevel != 0 || bt.WaterLevel == 0 || len(t.Stations) > 0 || (level != 0 && bt.WaterLevel != level) {
					continue
				}
				if len(docks) == 1 && (abs(int(int32(x)*32-docks[0][0]))+abs(int(int32(y)*32-docks[0][1])) < 6*32 || abs(int(int32(x)*32-docks[0][0]))+abs(int(int32(y)*32-docks[0][1])) > 14*32) {
					continue
				}
				level = bt.WaterLevel
				docks = append(docks, [3]int32{int32(x) * 32, int32(y) * 32, int32(bt.WaterLevel) * 16})
			}
		}
		if len(docks) < 2 {
			log.Printf("[Diag] shipsail: no two shore tiles on one water")
		} else {
			for i, dk := range docks {
				tl := m.TileAt(int(dk[0]>>5), int(dk[1]>>5))
				tl.Stations = append(tl.Stations, scenario.StationElement{StationType: 3, StationID: stations[i], Owner: company, BaseZ: uint8(dk[2] / 4)})
				st := &m.StationNames[stations[i]]
				st.Tiles[st.TileSize] = [3]int16{int16(dk[0]), int16(dk[1]), int16(dk[2])}
				st.TileSize++
			}
			m.ControllingCompany = company
			g.inTitleScreen = false
			own := &sim.Cmd{Company: company, Objects: om}
			var head uint16
			_, okBuy := m.CreateVehicle(own, sim.CmdApply, shipObj, true, &head)
			_, okPlace := m.VehiclePlaceWater(own, sim.CmdApply, sim.VehicleWaterPlacement{X: docks[0][0], Y: docks[0][1], Z: docks[0][2], Head: head})
			_, okOrder := m.VehicleOrderInsert(own, sim.CmdApply, head, 0, sim.OrderStopAtRaw(int(stations[1])))
			_, okStart := m.VehicleChangeRunningMode(own, sim.CmdApply, head, sim.RunningModeStart)
			mx, my := (docks[0][0]+docks[1][0])/2, (docks[0][1]+docks[1][1])/2
			g.w.CenterOn(fmt.Sprintf("tile:%d,%d", mx>>5, my>>5))
			g.diagShip = head
			log.Printf("[Diag] shipsail: docks at tiles %d,%d and %d,%d (water level %d, stations %d → %d, company %d); ship %d bought %v, placed %v, ordered %v, started %v",
				docks[0][0]>>5, docks[0][1]>>5, docks[1][0]>>5, docks[1][1]>>5, level, stations[0], stations[1], company, head, okBuy, okPlace, okOrder, okStart)
		}
	}
}

// diagShipGhost (GOLOCO_DIAG_BUILDVEHICLE=shipghost) — DIAGNOSTIC ONLY: the
// title map has no docks or boats — add a dock on water near the view's
// centre, buy a ship, hover over it
func (g *Game) diagShipGhost() {
	m := &g.w.Map
	om := g.r.ObjMgr
	var shipObj uint16 = 0xFFFF
	for i, v := range om.Vehicles {
		if v != nil && v.Mode == objects.TransportModeWater {
			shipObj = uint16(i)
			break
		}
	}
	var company uint8 = 0xFF
	var stationID uint16
	for i, st := range m.StationNames {
		if st.Name != 0 && st.Owner < 16 {
			company, stationID = st.Owner, uint16(i)
			break
		}
	}
	if shipObj == 0xFFFF || company == 0xFF || len(om.DockObjects) == 0 || om.DockObjects[0] == nil {
		log.Printf("[Diag] shipghost: no ship object, company station or dock object")
	} else {
		m.ControllingCompany = company
		g.inTitleScreen = false
		d := om.DockObjects[0]
		ox, oy := sim.RotateOffset(d.BoatPosition[0], d.BoatPosition[1], 0)
		var tx, ty int = -1, -1 // a tile whose dock boat position lies on water
		for y := 4; y < m.Height-4 && tx < 0; y++ {
			for x := 4; x < m.Width-4; x++ {
				t := m.TileAt(x, y)
				if _, water := m.GetHeight(int32(x)*32+int32(ox)+32, int32(y)*32+int32(oy)+32); water != 0 && t != nil && len(t.Stations) == 0 {
					if _, here := m.GetHeight(int32(x)*32, int32(y)*32); here == 0 { // the dock itself on the shore
						tx, ty = x, y
						break
					}
				}
			}
		}
		if tx < 0 {
			log.Printf("[Diag] shipghost: no shore tile with water at the boat position")
		} else {
			t := m.TileAt(tx, ty)
			t.Stations = append(t.Stations, scenario.StationElement{StationType: 3, StationID: stationID, Owner: company, BaseZ: t.BaseZ})
			g.w.CenterOn(fmt.Sprintf("tile:%d,%d", tx, ty))
			var head uint16
			if _, ok := m.CreateVehicle(&sim.Cmd{Company: company, Objects: om}, sim.CmdApply, shipObj, true, &head); !ok {
				log.Printf("[Diag] shipghost: buying ship %d refused", shipObj)
			} else {
				g.openVehicleWindow(head, 1)
				if x, y, ok := g.vehicleWin.PickupButton(); ok {
					g.vehicleWin.HandleClick(x, y)
				}
				if g.vehicleWin != nil { // out of the way of the dock at the centre (the click re-opened it centred)
					g.vehicleWin.X, g.vehicleWin.Y = 4, 40
				}
				for r := 0; r < 120 && !m.VehicleIsGhost(head); r += 4 {
					for _, o := range [][2]int{{r, r / 2}, {-r, r / 2}, {r, -r / 2}, {-r, -r / 2}, {0, r}, {0, -r}} {
						hx, hy := g.sw/2+o[0], g.sh/2+o[1]
						if g.overWindow(hx, hy) {
							continue
						}
						g.pickupToolUpdate(hx, hy)
						if m.VehicleIsGhost(head) {
							g.diagMouse = &[2]int{hx, hy}
							break
						}
					}
				}
				tr, _ := m.TrainOf(head)
				v2 := m.Entities[tr.Veh2]
				log.Printf("[Diag] shipghost: dock added at tile %d,%d (station %d, company %d), ship %d bought; pickup tool %v; a ghost now %v at %d,%d,%d; mouse %v",
					tx, ty, stationID, company, head, g.pickupOn, m.VehicleIsGhost(head), v2.PosX, v2.PosY, v2.PosZ, g.diagMouse)
			}
		}
	}
}

// diagRoadGhost (GOLOCO_DIAG_BUILDVEHICLE=roadghost) — DIAGNOSTIC ONLY: lift a
// road vehicle, its pickup tool on, hover until a placement — the ghost left
// standing for the shot
func (g *Game) diagRoadGhost() {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		h := m.Entities[tr.Head]
		v2 := m.Entities[tr.Veh2]
		// a vehicle of the ordinary streets (road object 1, or any road type): trams need their own tracks
		if h.TileX == -1 || h.Mode != 1 || int(h.Owner) >= len(m.Companies) || (h.TrackType != 1 && h.TrackType != 0xFF) {
			continue
		}
		if _, ok := m.VehiclePickup(&sim.Cmd{Company: h.Owner, Objects: g.r.ObjMgr}, sim.CmdApply, h.ID); !ok {
			continue
		}
		m.ControllingCompany = h.Owner
		g.inTitleScreen = false // windows draw over the title world
		g.w.CenterOn(fmt.Sprintf("tile:%d,%d", v2.TileX/32, v2.TileY/32))
		g.openVehicleWindow(h.ID, 1)
		g.vehicleWin.X, g.vehicleWin.Y = 8, 40 // out of the way of the road at the centre
		if x, y, ok := g.vehicleWin.PickupButton(); ok {
			g.vehicleWin.HandleClick(x, y)
		}
		log.Printf("[Diag] road vehicle %d of company %d lifted at tile %d,%d; pickup tool on: %v", h.ID, h.Owner, v2.TileX/32, v2.TileY/32, g.pickupOn)
		for d := 0; d < 200 && g.pickupOn && !m.VehicleIsGhost(h.ID); d += 4 { // outward from the centre
			for _, o := range [][2]int{{d, 0}, {-d, 0}, {0, d}, {0, -d}, {d, d}, {-d, -d}} {
				sx, sy := g.sw/2+o[0], g.sh/2+o[1]
				if g.overWindow(sx, sy) {
					continue
				}
				if p, ok := g.w.PlacementAt(h.ID, sx, sy); ok {
					g.pickupToolUpdate(sx, sy)
					g.diagMouse = &[2]int{sx, sy} // the mouse stays there: the frame's own pickupToolUpdate keeps the ghost
					log.Printf("[Diag] hover at %d,%d: placement %+v; a ghost now: %v", sx, sy, p, m.VehicleIsGhost(h.ID))
					break
				}
			}
		}
		break
	}
}

// diagExpress (GOLOCO_DIAG_BUILDVEHICLE=express) — DIAGNOSTIC ONLY: the Route
// tab of the first placed train, as its owner: a click on Express (then left
// open)
func (g *Game) diagExpress() {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		if h := m.Entities[tr.Head]; h.TileX != -1 {
			m.ControllingCompany = h.Owner // DIAGNOSTIC ONLY: play as its owner
			g.inTitleScreen = false        // DIAGNOSTIC ONLY: draw windows over the title world (it stops cycling maps)
			g.openVehicleWindow(h.ID, 4)
			log.Printf("[Diag] train %d express before: %v", h.ID, g.vehicleWin.Express)
			lw := ui.VehicleRouteLayout(g.vehicleWin.W, g.vehicleWin.H, g.vehicleWin.State.HiddenTabs, false)
			e := lw[ui.VrExpressMode]
			g.vehicleWin.HandleClick(g.vehicleWin.X+(e.L+e.R)/2, g.vehicleWin.Y+(e.T+e.B)/2)
			log.Printf("[Diag] train %d express after the click: %v (veh1 var_48 %#x)", h.ID, g.vehicleWin.Express, m.Entities[tr.Veh1].Var48)
			break
		}
	}
}

// diagRoute (GOLOCO_DIAG_BUILDVEHICLE=route) — DIAGNOSTIC ONLY: the Route tab
// of the first train with orders
func (g *Game) diagRoute() {
	m := &g.w.Map
	trains, _ := m.Trains()
	for _, tr := range trains {
		if h := m.Entities[tr.Head]; h.TileX != -1 && h.SizeOfOrderTable > 1 {
			g.openVehicleWindow(h.ID, 4)
			for _, l := range g.vehicleWin.Route {
				log.Printf("[Diag] route of %d: %q current %v", h.ID, l.Text, l.Current)
			}
			if g.vehicleWin.OnRouteSelect != nil { // select the first order, move it down
				g.vehicleWin.OnRouteSelect(0)
				g.vehicleWin.OnRoute("down")
				for _, l := range g.vehicleWin.Route {
					log.Printf("[Diag] after down: %q", l.Text)
				}
				log.Printf("[Diag] selection now row %d", g.routeSel)
			}
			break
		}
	}
}

// diagVehicleMain (GOLOCO_DIAG_BUILDVEHICLE=main) — DIAGNOSTIC ONLY: buy a
// train, then the Vehicle window's Main tab (clicked)
func (g *Game) diagVehicleMain() {
	g.openBuildVehicle(0)
	if w := g.buildVehicle; w != nil {
		w.HandleClick(w.X+3+1+5, w.Y+102+1+5)
	}
	g.buildVehicle = nil
	if vw := g.vehicleWin; vw != nil {
		vw.HandleClick(vw.X+3+15, vw.Y+15+13) // the Main tab
		log.Printf("[Diag] vehicle window tab %d, flag image %d", g.vehicleWin.Tab, g.vehicleWin.StopStartImage)
	}
}

// diagBuyPlace (GOLOCO_DIAG_BUILDVEHICLE=buy / place) — DIAGNOSTIC ONLY:
// trains, then a click on the first row
func (g *Game) diagBuyPlace(v string) {
	if v == "place" { // DIAGNOSTIC ONLY: lift a rail train that may be lifted, play as its owner, look at its freed track
		m := &g.w.Map
		trains, _ := m.Trains()
		for _, tr := range trains {
			h := m.Entities[tr.Head]
			v2 := m.Entities[tr.Veh2]
			if h.TileX == -1 || h.Mode != 0 || int(h.Owner) >= len(m.Companies) {
				continue
			}
			if _, ok := m.VehiclePickup(&sim.Cmd{Company: h.Owner, Objects: g.r.ObjMgr}, sim.CmdApply, h.ID); !ok {
				continue
			}
			log.Printf("[Diag] lifted train %d of company %d at tile %d,%d; controlling company %d → %d", h.ID, h.Owner, v2.TileX/32, v2.TileY/32, m.ControllingCompany, h.Owner)
			m.ControllingCompany = h.Owner
			m.Companies[h.Owner].Cash += 10_000_000 // diagnostic: the title map's companies are nearly broke
			g.w.CenterOn(fmt.Sprintf("tile:%d,%d", v2.TileX/32, v2.TileY/32))
			break
		}
	}
	g.openBuildVehicle(0)
	if w := g.buildVehicle; w != nil {
		w.HandleClick(w.X+3+1+5, w.Y+102+1+5)
	}
	if v == "place" && g.vehicleWin != nil { // then the pickup button, and the first track under a 16-px grid
		g.buildVehicle = nil
		if x, y, ok := g.vehicleWin.PickupButton(); ok {
			g.vehicleWin.HandleClick(x, y)
		}
		log.Printf("[Diag] pickup tool on: %v", g.pickupOn)
		placed := false
		for sy := 40; sy < g.sh-40 && !placed && g.pickupOn; sy += 16 {
			for sx := 16; sx < g.sw-16 && !placed; sx += 16 {
				if g.vehicleWin.Contains(sx, sy) {
					continue
				}
				if p, ok := g.w.RailPlacementAt(g.pickupHead, sx, sy); ok && !g.diagLogged {
					g.diagLogged = true
					m := &g.w.Map
					tr, _ := m.TrainOf(g.pickupHead)
					h := m.Entities[tr.Head]
					log.Printf("[Diag] first placement %+v; the train: trackType %d, needs mods %#x", p, h.TrackType, h.HeadVar53)
					if t := m.TileAt(int(p.X)>>5, int(p.Y)>>5); t != nil {
						for _, el := range t.Tracks {
							log.Printf("[Diag]   tile track: id %d rot %d z %d obj %d owner %d mods %#x ai %v ghost %v", el.TrackID, el.Rotation, int(el.BaseZ)*4, el.TrackObjectID, el.Owner, el.Mods, el.AiAllocated, el.Ghost)
						}
					}
				}
				g.pickupToolUpdate(sx, sy) // the hover first: a ghost where the click will place
				ghost := g.w.Map.VehicleIsGhost(g.pickupHead)
				if placed = g.pickupToolDown(sx, sy); placed {
					log.Printf("[Diag] placed at screen %d,%d (a ghost there before the click: %v; still a ghost: %v); tool on: %v, window placed: %v",
						sx, sy, ghost, g.w.Map.VehicleIsGhost(g.pickupHead), g.pickupOn, g.vehicleWin.State.Placed)
					if x, y, ok := g.vehicleWin.StopStartButton(); ok { // then Main's stop / start → Start
						flag := g.vehicleWin.StopStartImage
						g.vehicleWin.HandleClick(x, y)
						if g.dropdown != nil && g.dropdown.OnSelect != nil {
							g.dropdown.OnSelect(1)
							g.dropdown = nil
						}
						tr, _ := g.w.Map.TrainOf(g.pickupHead)
						log.Printf("[Diag] Main tab %v; stop / start: flag %d → %d, commandStop now %v", g.vehicleWin.Tab == 0, flag, g.vehicleWin.StopStartImage,
							g.w.Map.Entities[tr.Head].VehicleFlags&sim.VehicleFlagCommandStop != 0)
						if x, y, ok := g.vehicleWin.CentreViewButton(); ok { // then centre view → focus
							bx, by := g.tileUnderCursor(g.sw/2, g.sh/2)
							g.vehicleWin.HandleClick(x+12, y-12)
							if g.dropdown != nil && g.dropdown.OnSelect != nil {
								g.dropdown.OnSelect(1)
								g.dropdown = nil
							}
							ax, ay := g.tileUnderCursor(g.sw/2, g.sh/2)
							v2 := g.w.Map.Entities[tr.Veh2]
							log.Printf("[Diag] focus on the train: screen centre tile %d,%d → %d,%d (the train's veh2 at tile %d,%d)", bx, by, ax, ay, v2.TileX/32, v2.TileY/32)
						}
					}
				}
			}
		}
	}
}
