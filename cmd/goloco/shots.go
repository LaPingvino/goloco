package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// shots.go — captures of the finished frame: --shots, a script's "shot".

// maybeSaveShot implements --shots mode: save the full frame every shotsEvery,
// quit when the requested count has been captured. Script `shot` commands are
// honoured here too.
func (g *Game) maybeSaveShot(screen *ebiten.Image) {
	if os.Getenv("GOLOCO_DIAG_TOWNS") != "" && g.townList == nil && !g.diagTownsDone && g.w != nil {
		g.diagTownsDone = true
		if d := os.Getenv("GOLOCO_DIAG_TOWNS"); d == "stations" {
			g.openStationsWindow()
		} else if d == "vehicles" {
			g.openVehiclesWindow()
		} else {
			g.openTownsWindow()
		}
		g.townList.Hover = 2
		if d := os.Getenv("GOLOCO_DIAG_TOWNS"); d == "scrolled" { // a click on the track below the thumb
			l := g.townList.List
			g.townList.HandleClick(g.townList.X+l[0]+l[2]-6, g.townList.Y+l[1]+l[3]-20)
		}
	}
	if d := os.Getenv("GOLOCO_DIAG_CONSTRUCTION"); (d == "window" || d == "road" || d == "extend" || d == "signal" || d == "overhead" || d == "station" || d == "sigdrop" || d == "bridgedrop") && !g.diagConsDone {
		g.diagConsDone = true
		if d == "signal" {
			g.consTab = 2
		}
		if d == "station" {
			g.consTab = 1
		}
		if d == "bridgedrop" { // the bridge list open
			defer func() {
				x, y := g.consWindowOrigin()
				f, _ := g.constructionFrame(x, y)
				g.openBridgeDropdown(x, y, f.SkinColour, f.CompanyColour)
			}()
		}
		if d == "sigdrop" { // the signal tab with its dropdown open
			g.consTab = 2
			defer func() {
				x, y := g.consWindowOrigin()
				g.openConsDropdown(x, y, g.r.ObjMgr.InterfaceSkin.WindowConstructionColour, false)
			}()
		}
		if d == "overhead" { // the tram road (its overhead wires), one ticked
			g.buildMode = buildModeRoad
			for range 8 {
				if st := g.overheadTabState(); st.Names[0] != "" {
					break
				}
				g.consCycleType(1)
			}
			g.consTab, g.consMods = 3, 1
		}
		if d == "extend" {
			g.buildMode = buildModeTrack
			mw, mh := g.w.GetMapSize()
			g.w.SetConstructionHead(mw/2, mh/2)
		}
		if d == "road" {
			g.buildMode = buildModeRoad
			if names := g.consObjectNames(); len(names) > 0 && names[0] == "" {
				g.consCycleType(1)
			}
		}
		g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
	}
	if v := os.Getenv("GOLOCO_DIAG_BUILDVEHICLE"); v != "" && !g.buildVehDiagDone && g.w != nil && ((v != "routeadd" && v != "routewaypoint" && v != "express" && v != "roadghost" && v != "shipghost" && v != "shipsail" && v != "planefly" && v != "planeghost" && v != "newindustry" && v != "newindustrybuild" && v != "waves" && v != "playable" && v != "towns" && v != "buildstation" && v != "stationtool" && v != "stationghost") || g.w.SceneReady()) { // the Build Vehicle window on tab N
		g.buildVehDiagDone = true
		if v == "buildstation" { // DIAGNOSTIC ONLY: CreateTrainStation on the first free straight track piece that takes one, the view on it
			m := &g.w.Map
			om := g.r.ObjMgr
			g1, _ := om.G1File.(sim.G1Raw)
			typ := uint8(0)
			for i, s := range om.TrainStationObjects {
				if s != nil {
					typ = uint8(i)
					break
				}
			}
			done := false
			for y := 4; y < m.Height-4 && !done; y++ {
				for x := 4; x < m.Width-4 && !done; x++ {
					tl := m.TileAt(x, y)
					for _, r := range tl.Order {
						if r.Kind != scenario.KindTrack {
							continue
						}
						tr := tl.Tracks[r.Index]
						if tr.TrackID != 0 || tr.HasStation || tr.HasSignal || tr.HasLevelCrossing || tr.Ghost || tr.AiAllocated || tr.HasBridge {
							continue
						}
						c := &sim.Cmd{Company: tr.Owner, Objects: om}
						a := sim.TrainStationPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(tr.BaseZ) * 4, Rotation: tr.Rotation, TrackObjectID: tr.TrackObjectID, Type: typ}
						if _, ok := m.CreateTrainStation(c, 0, a, g1); !ok {
							continue
						}
						cost, ok := m.CreateTrainStation(c, sim.CmdApply, a, g1)
						g.inTitleScreen = false
						g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
						log.Printf("[Diag] buildstation: at %d,%d company %d: ok %v cost %d", x, y, tr.Owner, ok, cost)
						done = true
						break
					}
				}
			}
		} else if v == "roadstationtool" || v == "roadstationghost" { // DIAGNOSTIC ONLY: the Station tab's road tool over the first free straight road near a town, through picking (ghost: hover, leave, back, held for the shot; tool: clicked)
			m := &g.w.Map
			done := false
			for y := 4; y < m.Height-4 && !done; y++ {
				for x := 4; x < m.Width-4 && !done; x++ {
					near := false
					for _, t := range m.Towns {
						if dx, dy := int(t.X)/32-x, int(t.Y)/32-y; dx*dx+dy*dy <= 64 {
							near = true
						}
					}
					if !near {
						continue
					}
					tl := m.TileAt(x, y)
					for _, r := range tl.Order {
						if r.Kind != scenario.KindRoad {
							continue
						}
						rd := tl.Roads[r.Index]
						if rd.RoadID != 0 || rd.HasStationElement || rd.HasSignalElement || rd.HasLevelCrossing || rd.Ghost || rd.HasBridge {
							continue
						}
						g.inTitleScreen = false
						g.buildMode, g.consTab, g.buildObjID = buildModeRoad, 1, int(rd.RoadObjectID)
						g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
						sx, sy := g.w.TileToScreenPx(x, y)
						ghost := func() bool {
							for _, st := range m.TileAt(x, y).Stations {
								if st.Ghost && st.StationType == 1 {
									return true
								}
							}
							return false
						}
						if v == "roadstationghost" {
							g.stationToolUpdate(sx, sy+16)
							on := ghost()
							g.stationToolUpdate(5, 5)
							off := ghost()
							g.stationToolUpdate(sx, sy+16)
							log.Printf("[Diag] roadstationghost: tile %d,%d: ghost on hover %v, after leaving %v, back %v, cost %d, accepts %#x produces %#x", x, y, on, off, ghost(), g.stationGhostCost, g.stationGhostAccepted, g.stationGhostProduced)
							g.diagHoldGhost = true
							g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
						} else {
							g.stationToolUp(sx, sy+16)
							has := false
							for _, st := range m.TileAt(x, y).Stations {
								has = has || (!st.Ghost && st.StationType == 1)
							}
							log.Printf("[Diag] roadstationtool: clicked tile %d,%d: stop now %v, error window %v", x, y, has, g.errorWin != nil)
						}
						done = true
						break
					}
				}
			}
		} else if v == "stationtool" || v == "stationghost" { // DIAGNOSTIC ONLY: the Station tab's tool clicked on the screen over the first free straight track piece of a company — through picking, as the player would (stationghost: near a town first, for the catchment lists)
			m := &g.w.Map
			done := false
			nearTown := func(x, y int) bool {
				for _, t := range m.Towns {
					if dx, dy := int(t.X)/32-x, int(t.Y)/32-y; dx*dx+dy*dy <= 64 {
						return true
					}
				}
				return false
			}
			for pass := 0; pass < 2 && !done; pass++ {
				for y := 4; y < m.Height-4 && !done; y++ {
					for x := 4; x < m.Width-4 && !done; x++ {
						if pass == 0 && (v != "stationghost" || !nearTown(x, y)) {
							continue
						}
						tl := m.TileAt(x, y)
						for _, r := range tl.Order {
							if r.Kind != scenario.KindTrack {
								continue
							}
							tr := tl.Tracks[r.Index]
							if tr.TrackID != 0 || tr.HasStation || tr.HasSignal || tr.HasLevelCrossing || tr.Ghost || tr.AiAllocated || tr.HasBridge {
								continue
							}
							g.inTitleScreen = false
							m.ControllingCompany = tr.Owner
							g.buildMode, g.consTab, g.buildObjID = buildModeTrack, 1, int(tr.TrackObjectID)
							g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
							sx, sy := g.w.TileToScreenPx(x, y)
							if v == "stationghost" { // hover: a ghost appears; off the track it goes; back on, it stays for the shot
								g.stationToolUpdate(sx, sy+16)
								ghost := func() bool {
									for _, st := range m.TileAt(x, y).Stations {
										if st.Ghost {
											return true
										}
									}
									return false
								}
								on := ghost()
								g.stationToolUpdate(5, 5)
								off := ghost()
								g.stationToolUpdate(sx, sy+16)
								log.Printf("[Diag] stationghost: tile %d,%d: ghost on hover %v, after leaving %v, back %v, cost %d", x, y, on, off, ghost(), g.stationGhostCost)
								g.diagHoldGhost = true // the ghost and the tab stay for the shot, its cost in the window
								g.windowMgr.OpenWindow(g.newPortedConstructionWindow())
							} else {
								g.stationToolUp(sx, sy+16)
								log.Printf("[Diag] stationtool: clicked %d,%d (tile %d,%d company %d): station now %v", sx, sy+16, x, y, tr.Owner, m.TileAt(x, y).Tracks[r.Index].HasStation)
								g.openError(ui.FormatErrorArgs(i18n.Label(i18n.CantBuildPop3String), "City Station"), i18n.Label(i18n.WrongTypeOfTrackRoad)) // the window, held for the shot
								g.errorWin.Hold = true
								log.Printf("[Diag] stationtool: error window %q", g.errorWin.Lines)
							}
							done = true
							break
						}
					}
				}
			}
		} else if v == "towns" { // DIAGNOSTIC ONLY: the view on the biggest town (GOLOCO_DIAG_ZOOM: at that zoom), for the town labels
			g.inTitleScreen = false
			m := &g.w.Map
			best := 0
			for i, t := range m.Towns {
				if t.Population > m.Towns[best].Population {
					best = i
				}
			}
			if z, err := strconv.Atoi(os.Getenv("GOLOCO_DIAG_ZOOM")); err == nil {
				g.w.SetZoom(z)
			}
			t := m.Towns[best]
			g.w.CenterOn(fmt.Sprintf("tile:%d,%d", int(t.X)/32, int(t.Y)/32))
			log.Printf("[Diag] towns: %s, frame %+v", m.TownName(t.Name, best), t.LabelFrame)
		} else if v == "playable" { // DIAGNOSTIC ONLY: 5P end to end through the game's commands, as the player — a stopped rail train with two stops lifted, a copy bought (same cars), placed where it stood, its orders given, started; the view follows it, its window on Main
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
		} else if v == "drag" { // DIAGNOSTIC ONLY: a stopped saved train's last car dragged over its second row (GOLOCO_DIAG_DRAG_DROP: dropped there)
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
		} else if v == "sell" { // DIAGNOSTIC ONLY: as a saved train's company, a vehicle bought, its Details tab; GOLOCO_DIAG_SELL_OK: the bin pressed
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
		} else if v == "rename" { // DIAGNOSTIC ONLY: as a saved train's company, its window's caption clicked, " Express" typed
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
		} else if v == "refit" { // DIAGNOSTIC ONLY: as a saved train's company, a refittable vehicle bought, its Cargo tab, the refit dropdown
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
		} else if v == "finances" || v == "cargo" { // DIAGNOSTIC ONLY: the first saved train with a last income (cargo: holding cargo), that tab
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
		} else if v == "moneyeffect" { // DIAGNOSTIC ONLY: money popups at the screen's centre, one made whenever none is alive (they live 160 ticks)
			g.inTitleScreen = false
			g.w.Map.TitleMode = false
			g.diagMoney = true
		} else if v == "airportghost" { // DIAGNOSTIC ONLY: the airport window's tool over the first spot an airport takes (dry run), as the player; the ghost held, rotated once
			m := &g.w.Map
			m.ControllingCompany = m.PlayerCompanies[0]
			g.inTitleScreen = false
			g.openPortConstruction(0)
			g.airportRotation = 1
			c, g1 := g.stationCmd()
			done := false
			for y := 10; y < m.Height-10 && !done; y += 3 {
				for x := 10; x < m.Width-10 && !done; x += 3 {
					a, ok := m.AirportPlacementAt(int32(x)*32, int32(y)*32, uint8(g.consStation), g.airportRotation, g.r.ObjMgr)
					if !ok || m.TileAt(x, y).WaterLevel != 0 {
						continue
					}
					if _, ok := m.CreateAirport(c, 0, a, g1); !ok {
						continue
					}
					g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
					sx, sy := g.w.TileToScreenPx(x, y)
					g.stationToolUpdate(sx, sy+16)
					log.Printf("[Diag] airportghost: tile %d,%d rotation %d: ghost %v cost %d accepts %#x produces %#x", x, y, a.Rotation, g.airportGhost != nil, g.stationGhostCost, g.stationGhostAccepted, g.stationGhostProduced)
					g.diagHoldGhost = true
					done = true
				}
			}
		} else if v == "dockghost" || v == "docktool" { // DIAGNOSTIC ONLY: the dock window's tool over the first shore spot a dock takes (dry run), as a station owner — hover (ghost held) or click
			m := &g.w.Map
			for i := range m.StationNames {
				if m.StationNames[i].Name != 0xFFFF {
					m.ControllingCompany = m.StationNames[i].Owner
					break
				}
			}
			g.inTitleScreen = false
			g.openPortConstruction(1)
			c, g1 := g.stationCmd()
			done := false
			for y := 3; y < m.Height-4 && !done; y++ {
				for x := 3; x < m.Width-4 && !done; x++ {
					a, ok := m.DockPlacementAt(int32(x)*32, int32(y)*32, uint8(g.consStation), g.r.ObjMgr)
					if !ok || m.TileAt(x, y).WaterLevel != 0 {
						continue
					}
					if _, ok := m.CreatePort(c, 0, a, g1); !ok {
						continue
					}
					g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x, y))
					sx, sy := g.w.TileToScreenPx(x, y)
					if v == "dockghost" {
						g.stationToolUpdate(sx, sy+16)
						log.Printf("[Diag] dockghost: tile %d,%d rotation %d: ghost %v cost %d accepts %#x produces %#x", x, y, a.Rotation, g.dockGhost != nil, g.stationGhostCost, g.stationGhostAccepted, g.stationGhostProduced)
						g.diagHoldGhost = true
					} else {
						g.stationToolUp(sx, sy+16)
						docks := 0
						for _, st := range m.TileAt(x, y).Stations {
							if st.StationType == 3 && !st.Ghost {
								docks++
							}
						}
						log.Printf("[Diag] docktool: clicked tile %d,%d: dock elements %d, error window %v", x, y, docks, g.errorWin != nil)
					}
					done = true
				}
			}
		} else if v == "portmenu" { // DIAGNOSTIC ONLY: the toolbar's Port/Airport menu
			g.inTitleScreen = false
			for i := range g.toolbar.Buttons {
				if g.toolbar.Buttons[i].Tooltip == "Port/Airport" {
					g.handleToolbarButton(i)
				}
			}
			log.Printf("[Diag] portmenu: lastAirport %d lastShipPort %d, menu open %v", g.w.Map.LastAirport, g.w.Map.LastShipPort, g.dropdown != nil)
			if os.Getenv("GOLOCO_DIAG_PORT_DOCK") != "" { // choose Ship Port: the dock window (as a station owner: the title's company 0 is black)
				g.dropdown = nil
				for i := range g.w.Map.StationNames {
					if g.w.Map.StationNames[i].Name != 0xFFFF {
						g.w.Map.ControllingCompany = g.w.Map.StationNames[i].Owner
						break
					}
				}
				g.openPortConstruction(1)
				f, _ := g.constructionFrame(0, 0)
				c, _ := g.w.ControllingCompany()
				log.Printf("[Diag] portmenu: dock window, mode %d tab %d station %d, title %q raw %q, company %d colour %d", g.buildMode, g.consTab, g.consStation, f.Title, i18n.Str(i18n.TitleShipPort), c, f.CompanyColour)
			}
		} else if v == "menu" { // the toolbar's Build Vehicles menu instead
			for i := range g.toolbar.Buttons {
				if g.toolbar.Buttons[i].Tooltip == "Build Vehicles" {
					g.handleToolbarButton(i)
				}
			}
		} else if v == "routeadd" || v == "routewaypoint" { // the Route tool: click the first station (or plain track: a waypoint) found on screen
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
		} else if v == "waves" { // DIAGNOSTIC ONLY: the view on the map's most open water (the tile with the most water within 6 tiles), for waves
			g.inTitleScreen = false
			m := &g.w.Map
			best, bx, by := -1, 0, 0
			for y := 6; y < m.Height-6; y += 2 {
				for x := 6; x < m.Width-6; x += 2 {
					n := 0
					for dy := -6; dy <= 6; dy += 2 {
						for dx := -6; dx <= 6; dx += 2 {
							if m.TileAt(x+dx, y+dy).WaterLevel != 0 {
								n++
							}
						}
					}
					if n > best {
						best, bx, by = n, x, y
					}
				}
			}
			g.w.CenterOn(fmt.Sprintf("tile:%d,%d", bx, by))
			log.Printf("[Diag] waves: view on tile (%d, %d), %d of 49 sampled tiles water", bx, by, best)
		} else if v == "newindustry" || v == "newindustrybuild" { // DIAGNOSTIC ONLY: a farm-type industry by the createIndustry command at a searched place (built at once, or — newindustrybuild — under construction), the view on it
			g.inTitleScreen = false
			m := &g.w.Map
			om := g.r.ObjMgr
			if id, x, y, ok := m.DiagNewFarmIndustry(om, v == "newindustry"); ok {
				g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x>>5, y>>5))
				log.Printf("[Diag] newindustry: industry %d built at tile (%d, %d)", id, x>>5, y>>5)
			} else {
				log.Printf("[Diag] newindustry: none could be built")
			}
		} else if v == "planeghost" { // DIAGNOSTIC ONLY: buy an aircraft for an airport's owner, its pickup tool on, hover over the airport until a ghost stands at a terminal
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
		} else if v == "planefly" { // DIAGNOSTIC ONLY: follow the title map's aircraft at an airport (its status logged)
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
		} else if v == "shipsail" { // DIAGNOSTIC ONLY: two docks (at the water's height, in their stations' tiles) on one lake, a ship placed at the first with "stop at" the second, started — watch it sail
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
		} else if v == "shipghost" { // DIAGNOSTIC ONLY: the title map has no docks or boats — add a dock on water near the view's centre, buy a ship, hover over it
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
		} else if v == "roadghost" { // DIAGNOSTIC ONLY: lift a road vehicle, its pickup tool on, hover until a placement — the ghost left standing for the shot
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
		} else if v == "express" { // the Route tab of the first placed train, as its owner: a click on Express (then left open)
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
		} else if v == "route" { // the Route tab of the first train with orders
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
		} else if v == "main" { // buy a train, then the Vehicle window's Main tab (clicked)
			g.openBuildVehicle(0)
			if w := g.buildVehicle; w != nil {
				w.HandleClick(w.X+3+1+5, w.Y+102+1+5)
			}
			g.buildVehicle = nil
			if vw := g.vehicleWin; vw != nil {
				vw.HandleClick(vw.X+3+15, vw.Y+15+13) // the Main tab
				log.Printf("[Diag] vehicle window tab %d, flag image %d", g.vehicleWin.Tab, g.vehicleWin.StopStartImage)
			}
		} else if v == "buy" || v == "place" { // trains, then a click on the first row
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
		} else {
			tab, _ := strconv.Atoi(v)
			g.openBuildVehicle(tab)
		}
	}
	if os.Getenv("GOLOCO_DIAG_VEHLIST") != "" && !g.vehDiagDone && g.w != nil { // the vehicle list (trains of their first owner)
		g.vehDiagDone = true
		g.openVehicleList(0)
		if os.Getenv("GOLOCO_DIAG_VEHLIST") == "cargo" && g.townList != nil { // transporting cargo, the cargo dropdown open
			w := g.townList
			w.HandleClick(w.X+270, w.Y+205)
			if g.dropdown != nil && g.dropdown.OnSelect != nil {
				g.dropdown.OnSelect(1)
				g.dropdown = nil
			}
			w.HandleClick(w.X+400, w.Y+205)
		}
		if os.Getenv("GOLOCO_DIAG_VEHLIST") == "company" && g.townList != nil { // the company select dropdown open
			w := g.townList
			w.HandleClick(w.X+w.W-20, w.Y+20)
		}
	}
	if os.Getenv("GOLOCO_DIAG_OPTIONS") != "" && g.options == nil && !g.diagOptionsDone {
		g.diagOptionsDone = true
		g.openOptionsWindow()
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "controls" && g.options != nil {
			g.options.Tab = ui.OptTabControls
			g.options.Controls = [3]bool{true, false, true}
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regional" && g.options != nil {
			g.options.Tab = ui.OptTabRegional
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "framesolid" && g.options != nil {
			ui.FrameStyle = 1 // the Rendering tab with solid frames
			g.options.Tab = ui.OptTabRendering
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "prompt" { // the replace-file confirmation
			g.options = nil
			g.prompt = ui.NewPromptOkCancel("Save Game", "Replace existing file: “Bottleneck Blues”?", "Replace", g.sw, g.sh, nil)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "renderscale" && g.options != nil {
			g.options.Tab = ui.OptTabRendering // and a click on Vehicles min. display scale
			g.options.HandleClick(g.options.X+240, g.options.Y+66)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "framelimit" && g.options != nil {
			g.options.Tab = ui.OptTabDisplay // and a click on Frame rate limit
			g.options.HandleClick(g.options.X+240, g.options.Y+114)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "displayres" && g.options != nil {
			g.options.Tab = ui.OptTabDisplay // in fullscreen mode, a click on Display Resolution
			g.options.ScreenMode = 1
			g.options.HandleClick(g.options.X+240, g.options.Y+82)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "displaymode" && g.options != nil {
			g.options.Tab = ui.OptTabDisplay // and a click on Screen mode
			g.options.HandleClick(g.options.X+240, g.options.Y+66)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionalcurrency" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Preferred currency
			g.options.HandleClick(g.options.X+190, g.options.Y+122)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionalcurrent" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Current currency (as in a game)
			g.options.TitleMode = false
			g.options.HandleClick(g.options.X+190, g.options.Y+107)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionalalways" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // in a game, a click on 'always' (the current currency greys)
			g.options.TitleMode = false
			g.options.HandleClick(g.options.X+100, g.options.Y+153)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionallang" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Language
			g.options.HandleClick(g.options.X+190, g.options.Y+52)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionalheights" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Heights
			g.options.HandleClick(g.options.X+190, g.options.Y+87)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "regionaldrop" && g.options != nil {
			g.options.Tab = ui.OptTabRegional // and a click on Distance and Speed
			g.options.HandleClick(g.options.X+190, g.options.Y+72)
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "display" && g.options != nil {
			g.options.Tab = ui.OptTabDisplay
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "company" && g.options != nil {
			g.options.Tab = ui.OptTabCompany
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "rendering" && g.options != nil {
			g.options.Tab = ui.OptTabRendering
		}
		if os.Getenv("GOLOCO_DIAG_OPTIONS") == "misc" && g.options != nil {
			g.options.Tab = ui.OptTabMisc
		}
	}
	if os.Getenv("GOLOCO_DIAG_BROWSE") != "" && g.browse == nil && !g.diagBrowseDone {
		g.diagBrowseDone = true
		if os.Getenv("GOLOCO_DIAG_BROWSE") == "preview" { // a synthetic saved-game preview on the first file
			g.openLoadGameWindow()
			if g.browse != nil {
				img := make([]byte, 250*200)
				for i := range img {
					img[i] = byte(10 + (i%250)/25*10 + (i/250)/50) // bands of palette colours
				}
				g.browse.Preview = &scenario.SaveDetails{Company: "Joop's Railways", Owner: "Joop", Date: 55000, ChallengeProgress: 42, Image: img}
				g.browse.ChangeDir(filepath.Join(g.browse.Dir, "Scenarios"))
				g.browse.SelectFirstFile()
			}
		} else if os.Getenv("GOLOCO_DIAG_BROWSE") == "save" {
			g.openSaveWindow()
		} else {
			g.openFileWindow("Load Game", false)
		}
	}
	if os.Getenv("GOLOCO_DIAG_NEWGAME") != "" && g.scenarioSelect == nil && !g.diagNewGameDone {
		// Diagnostic: open the ScenarioSelect window with its first scenario selected.
		g.diagNewGameDone = true
		g.openNewGameWindow()
		g.scenarioSelect.SelectFirst()
		if os.Getenv("GOLOCO_DIAG_NEWGAME") == "many" {
			g.scenarioSelect.DiagRepeat(3)
			g.scenarioSelect.Tab = 3
			g.scenarioSelect.HandleWheel(-4)
		}
	}
	if os.Getenv("GOLOCO_DIAG_TIMEPANEL") != "" {
		st := ui.TimePanelState{Date: uilayout.DateDMY(1, "January", 1950), Speed: 1}
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			st.SkinColour = g.r.ObjMgr.InterfaceSkin.TimeToolbarColour
			st.SkinImage = int(g.r.ObjMgr.InterfaceSkin.ImageOffset)
		}
		ui.DrawTimePanel(screen, st)
		ps := ui.PlayerPanelState{Cash: 2500000}
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			ps.SkinColour = g.r.ObjMgr.InterfaceSkin.PlayerInfoToolbarColour
		}
		g.playerPanelCompany(&ps)
		ui.DrawPlayerPanel(screen, ps)
	}
	if os.Getenv("GOLOCO_DIAG_WINDOW") != "" {
		// Diagnostic: an upstream-style window frame + caption (FrameWidget,
		// CaptionWidget ports) for headless shots; the colour is the skin's window colour.
		var c uint8 = 1
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			c = g.r.ObjMgr.InterfaceSkin.WindowColour
		}
		ui.DrawWindowFrame(screen, 200, 60, 360, 220, c, false)
		ui.DrawWindowCaption(screen, 201, 61, 358, "Select Scenario")
	}
	if os.Getenv("GOLOCO_DIAG_CONSTRUCTION") == "frame" && g.r != nil && g.r.ObjMgr != nil {
		// Diagnostic: the ported construction frame for the first track object.
		f := &ui.ConstructionFrame{X: 10, Y: 40, Title: "Railway", FrameNo: g.titleFrame}
		if sk := g.r.ObjMgr.InterfaceSkin; sk != nil {
			f.SkinColour = sk.WindowConstructionColour
		}
		if len(g.w.CompanyColours) > 0 {
			f.CompanyColour = g.w.CompanyColours[0]
		}
		if len(g.r.ObjMgr.TrackObjects) > 0 && g.r.ObjMgr.TrackObjects[0] != nil {
			f.TrackImage = int(g.r.ObjMgr.TrackObjects[0].Image)
			f.Title = g.r.ObjMgr.TrackObjects[0].Name
		}
		f.Draw(screen)
		if len(g.r.ObjMgr.TrackObjects) > 0 && g.r.ObjMgr.TrackObjects[0] != nil {
			ui.DrawConstructionTab(screen, f, ui.ConstructionTabState{Traits: g.r.ObjMgr.TrackObjects[0].TrackPieces,
				Piece: "straight", Gradient: "level"})
		}
		f.DrawCaption(screen)
	}
	if os.Getenv("GOLOCO_DIAG_PROGRESS") != "" {
		// Diagnostic: the ProgressBar window at 3/4 in both styles, for headless shots.
		ui.DrawProgressBar(screen, "Loading...", 190, 0)
	}
	if os.Getenv("GOLOCO_DIAG_TOOLTIP") != "" {
		// Diagnostic: a tooltip over the frame (ToolTip.cpp port), for headless shots.
		var c uint8
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			c = g.r.ObjMgr.InterfaceSkin.TooltipColour
		}
		ui.DrawTooltip(screen, "Change viewport settings", 380, 120, c)
		ui.DrawTooltip(screen, "Build tracks, stations, signals and train depots, and place them on the map", 560, 240, c)
	}
	if os.Getenv("GOLOCO_DIAG_DROPDOWN") != "" {
		// Diagnostic: a toolbar-style dropdown on every frame (drawn last, just
		// before the capture), for headless screenshots of the Dropdown.cpp port.
		d := ui.NewDropdownMenu(40, 28, []ui.DropdownItem{{Text: "Load Game"}, {Text: "Save Game"}, {Separator: true},
			{Text: "Screenshot"}, {Separator: true}, {Text: "About", Disabled: true}, {Text: "Options"}, {Separator: true},
			{Text: "Quit to Menu"}, {Text: "Quit to Desktop"}})
		d.Highlighted = 1
		if g.r != nil && g.r.ObjMgr != nil && g.r.ObjMgr.InterfaceSkin != nil {
			d.SetToolbarColour(g.r.ObjMgr.InterfaceSkin.TopToolbarPrimaryColour)
		}
		d.Draw(screen)
	}
	if name := g.scriptShotName(); name != "" {
		name = debugPath(name)
		if f, err := os.Create(name); err == nil {
			png.Encode(f, screen.SubImage(screen.Bounds()))
			f.Close()
			log.Printf("[Script] saved %s", name)
		}
	}
	if g.shotsEvery <= 0 || g.shotsLeft <= 0 || time.Since(g.shotsLast) < g.shotsEvery {
		return
	}
	g.shotsLast = time.Now()
	g.shotsN++
	name := debugPath(fmt.Sprintf("shot_%02d.png", g.shotsN))
	if paint.Dump != nil { // GOLOCO_SHOT_PAINTDUMP: rtX order type image x y w h minX minY minZ maxX maxY maxZ quadrant
		os.WriteFile(strings.TrimSuffix(name, ".png")+".paint.txt", []byte(strings.Join(*paint.Dump, "\n")+"\n"), 0o644)
		paint.Dump = nil
	}
	if f, err := os.Create(name); err == nil {
		if err := png.Encode(f, screen.SubImage(screen.Bounds())); err != nil {
			log.Printf("[Shots] encode %s: %v", name, err)
		}
		f.Close()
		log.Printf("[Shots] saved %s (%d remaining)", name, g.shotsLeft)
	} else {
		log.Printf("[Shots] create %s: %v", name, err)
	}
	g.shotsLeft--
	if g.shotsLeft == 0 {
		g.quitAfterSave = true
	}
}

// debugPath prefixes capture filenames with the gitignored debug/ directory
// (created on demand) unless the caller already gave a path.
func debugPath(name string) string {
	if strings.ContainsRune(name, '/') {
		return name
	}
	_ = os.MkdirAll("debug", 0o755)
	return "debug/" + name
}
