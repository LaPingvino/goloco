package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"log"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
)

// vehicle.go — the Vehicle window (Vehicle.cpp): its tabs' data and actions.
// The pickup (placing) tool is in vehicle_pickup.go, the Route tab's orders in
// vehicle_route.go.

// openVehicleDetails is Vehicle::Details::open (Main::open, then the Details
// tab): 265×177 grown to Details' minimum (192×182 for the controlling
// company's own vehicle, onResize), the window's colours the owner's, its
// caption the vehicle's name.
func (g *Game) openVehicleDetails(headID uint16) { g.openVehicleWindow(headID, 1) }

// openVehicleWindow is the Vehicle window's open on a tab (0 Main, 1 Details,
// 2 Cargo, 3 Finances, 4 Route): the vehicle's window — there is one per
// vehicle (WindowType::vehicle, numbered by its head) — comes to the front on
// that tab (WindowManager::bringToFront(type, number)), or a new one opens.
func (g *Game) openVehicleWindow(headID uint16, tab int) {
	if w := g.newVehicleWindow(headID, tab); w != nil {
		g.showAuto(w, ui.VehicleWindowSize[0], ui.VehicleWindowSize[1]) // Main::open's create: kWindowSize
	}
}

// refreshVehicleWindow rebuilds the vehicle's window, if it is open, from the
// game's state on a tab (< 0: the one it is on), where it stands.
func (g *Game) refreshVehicleWindow(headID uint16, tab int) {
	old := g.vehicleWin(headID)
	if old == nil {
		return
	}
	if tab < 0 {
		tab = old.Tab
	}
	if w := g.newVehicleWindow(headID, tab); w != nil {
		g.wins.Refresh(w)
	}
}

// routeSelOf is the vehicle window's selected order (orderTableIndex; -1:
// none, or no window).
func (g *Game) routeSelOf(headID uint16) int {
	if w := g.vehicleWin(headID); w != nil {
		return w.RouteSel
	}
	return -1
}

func (g *Game) setRouteSel(headID uint16, sel int) {
	if w := g.vehicleWin(headID); w != nil {
		w.RouteSel = sel
	}
}

// vehicleWinColour is the vehicle window's secondary colour, for a dropdown
// it opens (1 without the window).
func (g *Game) vehicleWinColour(headID uint16) uint8 {
	if w := g.vehicleWin(headID); w != nil {
		return w.Colour
	}
	return 1
}

// newVehicleWindow builds the Vehicle window of a vehicle on a tab from the
// game's state (nil: no such vehicle). An open window's order selection is
// kept; a new one has none (orderTableIndex -1).
func (g *Game) newVehicleWindow(headID uint16, tab int) *ui.VehicleWindow {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return nil
	}
	routeSel := g.routeSelOf(headID)
	h := m.Entities[tr.Head]
	company, _ := g.w.ControllingCompany()
	own := int(h.Owner) == company
	title := ""
	for _, r := range m.VehicleListRows() {
		if r.HeadID == headID {
			title = r.Name
		}
	}
	var cars []ui.TrainCarDraw
	var names []string
	for _, c := range g.w.TrainCars(headID) {
		cars = append(cars, ui.TrainCarDraw(c))
		n := ""
		if c.Object != nil {
			n = c.Object.DisplayName
		}
		names = append(names, n)
	}
	tc, wc := g.windowColours(true)
	if int(h.Owner) < len(g.w.CompanyColours) {
		tc = g.w.CompanyColours[h.Owner]
	}
	// Each tab's onResize: setSize(kMinWindowSize, kMaxWindowSize) — the
	// window (265 × 177 as created, or as the player left it) brought within
	// the tab's limits when it opens or changes tab.
	var lim [4]int
	switch tab {
	case 0: // Main
		lim = [4]int{192, 177, 600, 440}
	case 1: // Details: 182 high with the paint tools (the controlling company's own)
		lim = [4]int{ui.VehicleDetailsMinSize[0], ui.VehicleDetailsMinSize[1], 400, 440}
		if own {
			lim[0], lim[1] = ui.VehicleDetailsMinSizePaint[0], ui.VehicleDetailsMinSizePaint[1]
		}
	case 2: // Cargo
		lim = [4]int{ui.VehicleCargoMinSize[0], ui.VehicleCargoMinSize[1], 400, 440}
	case 3: // Finances: one size
		lim = [4]int{ui.VehicleFinancesSize[0], ui.VehicleFinancesSize[1], ui.VehicleFinancesSize[0], ui.VehicleFinancesSize[1]}
	default: // Route
		lim = [4]int{265, ui.VehicleRouteMinHeight, 600, 440}
	}
	vw := &ui.VehicleWindow{WinBase: ui.WinBase{WinType: ui.WindowTypeVehicle, Number: int(headID), W: ui.VehicleWindowSize[0], H: ui.VehicleWindowSize[1],
		MinW: lim[0], MinH: lim[1], MaxW: lim[2], MaxH: lim[3], WinFlags: ui.WindowFlagResizable}, TitleColour: tc, Colour: wc,
		Title: title, VehicleType: int(h.VehicleType), Cars: cars, Names: names,
		State:        ui.VehicleDetailsState{Own: own, Placed: h.TileX != -1, NoCars: len(tr.Cars) == 0},
		OnPickup:     func() { g.onPickup(headID) },
		PickupActive: g.pickupOn && g.pickupHead == headID, PickupImage: g.pickupImage(&h), BuildNewImage: g.buildNewImage(&h),
		OnBuildNew: func(x, y, w, hh int) { g.buildNewOpen(headID, x, y, w, hh) },
		Route:      g.routeLines(&h), RouteSel: routeSel, Hover: -1, Finances: g.vehicleFinances(tr), Cargo: g.vehicleCargo(tr, cars),
		OnRouteSelect: func(row int) { g.setRouteSel(headID, row); g.refreshVehicleWindow(headID, 4) },
		OnRoute:       func(a string) { g.routeAction(headID, a); g.refreshVehicleWindow(headID, 4) },
		Express:       m.Entities[tr.Veh1].Var48&sim.Flags48ExpressMode != 0,
		Tab:           tab, Main: ui.VehicleMainState{Own: own, Manual: h.VehicleFlags&sim.VehicleFlagManualControl != 0,
			NoCars: len(tr.Cars) == 0, Exhausted: m.Entities[tr.Veh1].Var3C >= 0x3689,
			AirWaterMoving: (h.Mode == 2 || h.Mode == 3) && h.Status != sim.StatusStopped && h.Status != sim.StatusLoading && h.TileX != -1,
			Unplaced:       h.TileX == -1, ToolActive: g.pickupOn && g.pickupHead == headID, WaitingAtSignal: h.Status == sim.StatusWaitingAtSignal},
		StopStartImage: stopStartImage(h.VehicleFlags),
		OnTab: func(t int) { // switchTab
			g.refreshVehicleWindow(headID, t)
			if w := g.vehicleWin(headID); w != nil {
				g.wins.MoveInsideScreenEdges(w)
			}
		},
		View:              func(w, h int) *gfx8.RenderTarget { return g.w.VehicleView(headID, w, h) },
		Labels:            func(dst *ebiten.Image) { g.w.VehicleViewLabels(headID, dst) },
		OnCentreView:      func() { g.centreViewportOpen(headID) },
		OnStopStart:       func(x, y, w, hh int) { g.stopStartOpen(headID, x, y, w, hh) },
		OnRefit:           func(x, y, w, hh int) { g.refitOpen(headID, x, y, w, hh) },
		OnOrderDropdown:   func(kind string, x, y, w, hh int) { g.orderDropdown(headID, kind, x, y, w, hh) },
		OnPassSignal:      func() { g.passSignal(headID) },
		OnChangeDirection: func() { g.changeDirection(headID) },
		OnSpeed:           func(speed int8) { g.speedControl(headID, speed) },
		ManualPower:       h.ManualPower}
	if own {
		vw.OnRename = func() { g.renameVehicle(headID, tc, wc) }
	}
	// Common's onClose: the tool goes with its window (ToolManager::
	// isToolActive(type, number) → toolCancel) — another vehicle's stays.
	vw.OnClose = func() {
		if g.pickupOn && g.pickupHead == headID {
			g.pickupToolAbort()
			g.pickupOn = false
		}
		if g.routeHead == headID {
			g.routeTool = false
		}
	}
	vw.OnSell = func() { g.sellVehicle(headID) }
	vw.OnDropSell = func(row int) { g.dropCar(headID, row, -2) }
	vw.OnDropMove = func(row, dest int) { g.dropCar(headID, row, dest) }
	return vw
}

// stopStartImage is Main::prepareDraw's: yellow in manual control, red when
// told to stop, else green.
func stopStartImage(flags uint16) int {
	switch {
	case flags&sim.VehicleFlagManualControl != 0:
		return ui.ImgYellowFlag
	case flags&sim.VehicleFlagCommandStop != 0:
		return ui.ImgRedFlag
	}
	return ui.ImgGreenFlag
}

// stopStartOpen is Main's stopStartOpen: a text dropdown at the button with
// Stop and Start, and Manual for a rail vehicle with the driver cheat on
// (SceneManager::isDriverCheatEnabled — g.driverCheat; the Cheats window that
// turns it on is not ported), the current mode marked (manual control above
// the others), Start highlighted after Stop and Stop otherwise;
// onStopStartDropdown runs VehicleChangeRunningMode and the window shows the
// new flag.
func (g *Game) stopStartOpen(headID uint16, x, y, w, h int) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	head := m.Entities[tr.Head]
	selected := 0 // Stop
	if head.VehicleFlags&sim.VehicleFlagCommandStop == 0 {
		selected = 1 // Start
	}
	items := []ui.DropdownItem{{Text: i18n.Label(i18n.Stop)}, {Text: i18n.Label(i18n.Start)}}
	if head.Mode == 0 && g.driverCheat { // TransportMode::rail
		items = append(items, ui.DropdownItem{Text: i18n.Label(i18n.Manual)})
	}
	if head.VehicleFlags&sim.VehicleFlagManualControl != 0 {
		selected = 2
	}
	if selected < len(items) {
		items[selected].Selected = true
	}
	d := ui.NewDropdownMenu(x, y+h, items) // Dropdown::showText below the button
	d.Colour = g.vehicleWinColour(headID)
	d.Highlighted = 0 // stop → start highlighted, start or manual → stop
	if selected == 0 {
		d.Highlighted = 1
	}
	d.OnSelect = func(item int) {
		mode := sim.RunningModeStop
		switch item {
		case 1:
			mode = sim.RunningModeStart
		case 2:
			mode = sim.RunningModeDriveManually
		}
		c := g.playerCmd()
		if _, ok := m.VehicleChangeRunningMode(c, sim.CmdApply, headID, mode); !ok {
			log.Printf("[Game] can't %s vehicle %d", []string{"stop", "start", "drive"}[item], headID)
		}
		g.refreshVehicleWindow(headID, -1)
	}
	g.dropdown = d
}

// pickupImage is getPickupImageIdandTooltip's image: by transport mode the
// track's, road's (0xFF: the default track type's) or water object's
// pick-up / place image, or the airport ones — pick-up when placed, place
// when not. (It returned none but for rail, so a road vehicle, aircraft or
// ship had an empty, invisible place button.)
func (g *Game) pickupImage(h *scenario.VehicleEntity) int {
	om := g.r.ObjMgr
	if om == nil {
		return 0
	}
	placed := h.TileX != -1
	pick := func(base uint32, pickup, place int) int { // isPlaced ? pickup : place
		if placed {
			return int(base) + pickup
		}
		return int(base) + place
	}
	switch h.Mode {
	case 0: // rail: TrackObj kUiPickupFromTrack / kUiPlaceOnTrack
		if t := om.GetTrackObjectByIndex(int(h.TrackType)); t != nil {
			return pick(t.Image, 16, 17)
		}
	case 1: // road: RoadObj kUiPickupFromRoad / kUiPlaceOnRoad, 0xFF the default track type
		id := int(h.TrackType)
		if h.TrackType == 0xFF {
			id = int(g.w.Map.LastTrackTypeOption) // GameState defaultTrackTypeObjectId (0x1AD)
		}
		if r := om.GetRoadObjectByIndex(id); r != nil {
			return pick(r.Image, 32, 33)
		}
	case 2: // air: ImageIds pickupFromAirport / placeOnAirport
		return pick(0, 2372, 2373)
	case 3: // water: Water kUiPickupFromWater / kUiPlaceOnWater
		if om.WaterObj != nil {
			return pick(om.WaterObj.ImageOffset, 58, 59)
		}
	}
	return 0
}

// centreViewportOpen is Main's onCentreViewportControl: a text dropdown at the
// button — move the main view to show the vehicle, or focus it on the vehicle
// (following it) — and onCentreViewportDropdown.
func (g *Game) centreViewportOpen(headID uint16) {
	vw := g.vehicleWin(headID)
	if vw == nil {
		return
	}
	x, y, _ := vw.CentreViewButton()
	d := ui.NewDropdownMenu(x, y, []ui.DropdownItem{{Text: i18n.Label(i18n.DropdownViewportMove)}, {Text: i18n.Label(i18n.DropdownViewportFocus)}})
	d.Colour = vw.Colour
	d.OnSelect = func(item int) {
		if item <= 0 {
			g.w.CentreOnVehicle(headID) // viewportCentreMain, without locking
			return
		}
		g.w.FocusOnVehicle(headID)
	}
	g.dropdown = d
}

// vehicleFinances is what Finances::draw formats: Vehicle1's lastIncome (its
// day, then each cargo's unit name with the quantity, the distance in tiles,
// the age in days and the profit), the head's last average speed, the
// running cost, veh2's totalRecentProfit / 4, the head's totalRefundCost.
func (g *Game) vehicleFinances(tr sim.Train) ui.VehicleFinances {
	m := &g.w.Map
	head, v1 := m.Entities[tr.Head], m.Entities[tr.Veh1]
	var f ui.VehicleFinances
	li := v1.LastIncome
	if li.Day != -1 {
		y, mo, d := uilayout.CalcDate(uint32(li.Day))
		f.IncomeHeader = ui.FillArgs(i18n.Str(i18n.LastIncomeOnDate), uilayout.DateDMY(d, i18n.Month(mo+1), y))
		for i := 0; i < 4; i++ {
			c := li.CargoTypes[i]
			if c == 0xFF {
				continue
			}
			unit := ""
			if om := g.r.ObjMgr; om != nil && int(c) < len(om.CargoObjects) && om.CargoObjects[c] != nil {
				unit = ui.ObjectString(om.CargoObjects[c].UnitNamePlural)
				if li.CargoQtys[i] == 1 {
					unit = ui.ObjectString(om.CargoObjects[c].UnitNameSingular)
				}
			}
			f.Income = append(f.Income, ui.FillArgs(i18n.Str(i18n.TransportedBlocksInDays), unit, ui.Comma16(int(li.CargoQtys[i])),
				ui.Comma16(int(li.CargoDistances[i])), ui.Comma16(int(li.CargoAges[i])), uilayout.Currency48(int64(li.CargoProfits[i]))))
		}
	} else {
		f.IncomeHeader = i18n.Str(i18n.LastIncomeNa)
	}
	if head.LastAverageSpeed != 0 {
		f.AverageSpeed = ui.FillArgs(i18n.Str(i18n.LastJourneyAverageSpeed), uilayout.Velocity(int(head.LastAverageSpeed)))
	}
	f.RunningCost = ui.FillArgs(i18n.Str(i18n.VehicleMonthlyRunningCost), uilayout.Currency48(int64(m.RunningCost(tr, g.r.ObjMgr))))
	var profit int64
	for _, r := range m.VehicleListRows() {
		if int(r.HeadID) == int(head.ID) {
			profit = r.Profit // totalRecentProfit / 4
		}
	}
	f.Profit = ui.FillArgs(i18n.Str(i18n.VehicleMonthlyProfit), uilayout.Currency48(profit))
	f.SaleValue = ui.FillArgs(i18n.Str(i18n.SaleValueOfVehicle), uilayout.Currency48(int64(head.TotalRefundCost)))
	return f
}

// vehicleCargo is what Cargo::draw formats: per car (its first component's
// front bogie and body) the body's primary and the bogie's secondary cargo,
// each with a quantity "{unit} from {station}", or "Empty" when it carries
// something but holds none; the totals and capacities by cargo type
// (generateCargoTotalString / CapacityString: cargoLUTToString, "Empty" when
// none); refit for the player's own refittable vehicle (canRefit).
func (g *Game) vehicleCargo(tr sim.Train, cars []ui.TrainCarDraw) ui.VehicleCargo {
	m := &g.w.Map
	om := g.r.ObjMgr
	unit := func(c uint8, n uint32) string {
		if om == nil || int(c) >= len(om.CargoObjects) || om.CargoObjects[c] == nil {
			return ""
		}
		if n == 1 {
			return ui.ObjectString(om.CargoObjects[c].UnitNameSingular)
		}
		return ui.ObjectString(om.CargoObjects[c].UnitNamePlural)
	}
	var out ui.VehicleCargo
	var totals, caps [32]uint32
	for i, car := range tr.Cars {
		front, body := m.Entities[car[0].Front], m.Entities[car[0].Body]
		row := ui.VehicleCargoRow{}
		if i < len(cars) {
			row.Car = cars[i]
		}
		for _, cg := range []scenario.VehicleCargo{front.Cargo, body.Cargo} {
			if cg.Type != 0xFF && int(cg.Type) < len(totals) {
				totals[cg.Type] += uint32(cg.Qty)
				caps[cg.Type] += uint32(cg.MaxQty)
			}
		}
		if body.Cargo.Type != 0xFF {
			if front.Cargo.Qty != 0 || body.Cargo.Qty != 0 {
				for _, cg := range []scenario.VehicleCargo{body.Cargo, front.Cargo} { // drawCargoText: primary, then secondary
					if cg.Qty == 0 {
						continue
					}
					row.Lines = append(row.Lines, ui.FillArgs(i18n.Str(i18n.CargoFrom), unit(cg.Type, uint32(cg.Qty)), ui.Comma16(int(cg.Qty)), m.StationName(int(cg.TownFrom))))
				}
			} else {
				row.Lines = []string{i18n.Str(i18n.CargoEmpty)}
			}
		}
		out.Rows = append(out.Rows, row)
	}
	lut := func(t *[32]uint32) string { // cargoLUTToString
		s := ""
		for c, n := range t {
			if n == 0 {
				continue
			}
			if s != "" {
				s += ", "
			}
			s += ui.FillArgs(unit(uint8(c), n), ui.Comma16(int(n)))
		}
		if s == "" {
			s = i18n.Str(i18n.CargoEmpty2)
		}
		return s
	}
	out.Total = ui.FillArgs(i18n.Str(i18n.TotalStringid), lut(&totals))
	out.Capacity = ui.FillArgs(i18n.Str(i18n.VehicleCapacityStringid), lut(&caps))
	if h := m.Entities[tr.Head]; m.IsPlayerCompany(int(h.Owner)) && len(tr.Cars) > 0 && om != nil {
		if id := m.Entities[tr.Cars[0][0].Front].ObjectID; int(id) < len(om.Vehicles) && om.Vehicles[id] != nil {
			out.CanRefit = om.Vehicles[id].Flags&(1<<14) != 0 // VehicleObjectFlags::refittable
		}
	}
	return out
}

// refitOpen is Cargo::onRefitButton: a dropdown below the button of every
// cargo with the refit flag — its plural unit name with the capacity the car
// would have (getNumUnitsForCargo), the current one selected, the first
// highlighted; onDropdown asks first when the train holds cargo
// (confirmComponentChange, the keep-cargo cheat not consulted), then
// VehicleRefit.
func (g *Game) refitOpen(headID uint16, x, y, w, h int) {
	m := &g.w.Map
	om := g.r.ObjMgr
	tr, err := m.TrainOf(headID)
	if err != nil || len(tr.Cars) == 0 || om == nil {
		return
	}
	obj := om.Vehicles[m.Entities[tr.Cars[0][0].Front].ObjectID]
	if obj == nil {
		return
	}
	cargo := om.CargoFor(obj)
	primary := uint8(0xFF)
	for i := 0; i < 32; i++ {
		if uint32(cargo.CompatibleCargoCategories[0])&(1<<i) != 0 {
			primary = uint8(i)
			break
		}
	}
	current := m.Entities[tr.Cars[0][0].Body].Cargo.Type
	var items []ui.DropdownItem
	var ids []uint8
	for id, co := range om.CargoObjects {
		if co == nil || co.Flags&(1<<1) == 0 { // CargoObjectFlags::refit
			continue
		}
		n := sim.NumUnitsForCargo(om, uint32(cargo.MaxCargo[0]), primary, uint8(id))
		items = append(items, ui.DropdownItem{Text: ui.FillArgs(ui.ObjectString(co.UnitNamePlural), ui.Comma16(int(n))), Selected: uint8(id) == current, MoveX10: true})
		ids = append(ids, uint8(id))
	}
	if len(items) == 0 {
		return
	}
	d := ui.NewDropdownMenu(x, y+h, items)
	d.Colour = g.vehicleWinColour(headID)
	d.Highlighted = 0
	d.OnSelect = func(item int) {
		apply := func() {
			c := g.playerCmd()
			if _, ok := m.VehicleRefit(c, sim.CmdApply, headID, ids[item]); !ok {
				log.Printf("[Game] %s", i18n.Str(i18n.CantRefitVehicle))
			}
			g.refreshVehicleWindow(headID, 2)
		}
		if t, err := m.TrainOf(headID); err == nil && m.HasAnyCargo(t) {
			g.prompt = ui.NewPromptOkCancel(i18n.Str(i18n.ConfirmVehicleComponentRefitCargoWarningTitle),
				i18n.Str(i18n.ConfirmVehicleComponentRefitCargoWarningMessage), i18n.Str(i18n.ConfirmVehicleComponentRefitCargoWarningConfirm), g.sw, g.sh, apply)
			return
		}
		apply()
	}
	g.dropdown = d
}

// renameVehicle is Common::renameVehicle: the text input titled "Name
// {type}", "Type new name for this {type}:", on the current name (inputSize
// kUserStringSize − 1); textInput: VehicleRename with the text (an empty
// one changes nothing — upstream returns before the command).
func (g *Game) renameVehicle(headID uint16, titleColour, colour uint8) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	typ := i18n.Str([]int{i18n.Train, i18n.Bus, i18n.Truck, i18n.Tram, i18n.Aircraft, i18n.Ship}[min(int(h.VehicleType), 5)])
	g.textInput = ui.OpenTextInput(ui.FillArgs(i18n.Str(i18n.TitleNameVehicle), typ), ui.FillArgs(i18n.Str(i18n.PromptEnterNewVehicleName), typ),
		m.VehicleName(h.Name, int(h.OrdinalNumber)), 31, g.sw, g.sh, titleColour, colour, true, func(text string) {
			if text == "" {
				return
			}
			if _, ok := m.VehicleRename(g.playerCmd(), sim.CmdApply, headID, text); !ok {
				log.Printf("[Game] %s", i18n.Str(i18n.CantRenameThisVehicle))
			}
			g.refreshVehicleWindow(headID, -1)
		})
}

// sellVehicle is Details' remove button: VehicleSell for the head, as a
// company command (VehicleDisposals); sold, its window closes (deleteTrain's
// WindowManager::close) and the pickup tool with it.
func (g *Game) sellVehicle(headID uint16) {
	m := &g.w.Map
	c := g.playerCmd()
	cost, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehicleDisposals, func(flags uint8) (int32, bool) {
		return m.VehicleSell(c, flags, headID)
	})
	if !ok {
		log.Printf("[Game] can't sell vehicle %d", headID)
		return
	}
	log.Printf("[Game] sold vehicle %d for %d", headID, -cost)
	g.wins.CloseType(ui.WindowTypeVehicle, int(headID))
}

// dropCar is Details::scrollDragEnd for this vehicle's car at row: dest −2
// the remove button (VehicleSell of that car, VehicleDisposals), else the row
// it goes before (−1: the tail) — VehicleRearrange; each asked first when the
// train holds cargo (confirmComponentChange; the keep-cargo cheat is not
// kept, so it never skips the question).
func (g *Game) dropCar(headID uint16, row, dest int) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil || row < 0 || row >= len(tr.Cars) {
		return
	}
	src := m.Entities[tr.Cars[row][0].Front].ID
	c := g.playerCmd()
	var run func()
	var title, msg, ok int
	if dest == -2 {
		title, msg, ok = i18n.ConfirmVehicleComponentSellCargoWarningTitle, i18n.ConfirmVehicleComponentSellCargoWarningMessage, i18n.ConfirmVehicleComponentSellCargoWarningConfirm
		run = func() {
			if _, done := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureVehicleDisposals, func(f uint8) (int32, bool) { return m.VehicleSell(c, f, src) }); !done {
				log.Printf("[Game] can't sell car %d", src)
			}
			g.openVehicleWindow(headID, 1)
		}
	} else {
		dst := m.Entities[tr.Tail].ID
		if dest >= 0 && dest < len(tr.Cars) {
			dst = m.Entities[tr.Cars[dest][0].Front].ID
		}
		title, msg, ok = i18n.ConfirmVehicleComponentMoveCargoWarningTitle, i18n.ConfirmVehicleComponentMoveCargoWarningMessage, i18n.ConfirmVehicleComponentMoveCargoWarningConfirm
		run = func() {
			if _, done := m.DoCommandFor(c, sim.CmdApply, 1, func(f uint8) (int32, bool) { return m.VehicleRearrange(c, f, src, dst) }); !done { // TrainRunningCosts
				log.Printf("[Game] can't move car %d", src)
			}
			g.openVehicleWindow(headID, 1)
		}
	}
	if m.HasAnyCargo(tr) {
		g.prompt = ui.NewPromptOkCancel(i18n.Str(title), i18n.Str(msg), i18n.Str(ok), g.sw, g.sh, run)
		return
	}
	run()
}

// passSignal is Main's passSignal button: VehiclePassSignal for the
// controlling company (cost 0, TrainRunningCosts); refused, the error window
// (cant_pass_signal_at_danger).
func (g *Game) passSignal(headID uint16) {
	m := &g.w.Map
	c := g.playerCmd()
	if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureTrainRunningCosts, func(f uint8) (int32, bool) {
		return m.VehiclePassSignal(c, f, headID)
	}); !ok {
		g.openError(i18n.Label(i18n.CantPassSignalAtDanger), cmdErrorText(c))
	}
}

// changeDirection is Main's onChangeDirection: with this vehicle's pickup
// tool active, the placement direction flips (pickupDirection ^= 1);
// otherwise VehicleReverse for the controlling company, refused → the error
// window (cant_reverse_train).
func (g *Game) changeDirection(headID uint16) {
	m := &g.w.Map
	if g.pickupOn && g.pickupHead == headID {
		m.PickupDirection ^= 1
		return
	}
	c := g.playerCmd()
	if _, ok := m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureTrainRunningCosts, func(f uint8) (int32, bool) {
		return m.VehicleReverse(c, f, headID)
	}); !ok {
		g.openError(i18n.Label(i18n.CantReverseTrain), cmdErrorText(c))
	}
}

// speedControl is Main's onSpeedControl: VehicleSpeedControl for the
// controlling company (cost 0, TrainRunningCosts), the slider's thumb
// following the head's new manualPower.
func (g *Game) speedControl(headID uint16, speed int8) {
	m := &g.w.Map
	c := g.playerCmd()
	m.DoCommandFor(c, sim.CmdApply, sim.ExpenditureTrainRunningCosts, func(f uint8) (int32, bool) {
		return m.VehicleSpeedControl(c, f, headID, speed)
	})
	if vw := g.vehicleWin(headID); vw != nil {
		vw.ManualPower = speed
	}
}
