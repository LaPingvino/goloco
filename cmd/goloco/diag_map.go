package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/LaPingvino/goloco/pkg/ui"
)

// diag_map.go — GOLOCO_DIAG_BUILDVEHICLE's map scenarios: the Station tab's
// tools (train station, road stop, dock, airport), the port menu, views on
// towns and water, a new industry, money popups.

// diagBuildStation (GOLOCO_DIAG_BUILDVEHICLE=buildstation) — DIAGNOSTIC ONLY:
// CreateTrainStation on the first free straight track piece that takes one,
// the view on it
func (g *Game) diagBuildStation() {
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
}

// diagRoadStationTool (GOLOCO_DIAG_BUILDVEHICLE=roadstationtool /
// roadstationghost) — DIAGNOSTIC ONLY: the Station tab's road tool over the
// first free straight road near a town, through picking (ghost: hover, leave,
// back, held for the shot; tool: clicked)
func (g *Game) diagRoadStationTool(v string) {
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
}

// diagStationTool (GOLOCO_DIAG_BUILDVEHICLE=stationtool / stationghost) —
// DIAGNOSTIC ONLY: the Station tab's tool clicked on the screen over the first
// free straight track piece of a company — through picking, as the player
// would (stationghost: near a town first, for the catchment lists)
func (g *Game) diagStationTool(v string) {
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
}

// diagTownsView (GOLOCO_DIAG_BUILDVEHICLE=towns) — DIAGNOSTIC ONLY: the view
// on the biggest town (GOLOCO_DIAG_ZOOM: at that zoom), for the town labels
func (g *Game) diagTownsView() {
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
}

// diagMoneyEffect (GOLOCO_DIAG_BUILDVEHICLE=moneyeffect) — DIAGNOSTIC ONLY:
// money popups at the screen's centre, one made whenever none is alive (they
// live 160 ticks)
func (g *Game) diagMoneyEffect() {
	g.inTitleScreen = false
	g.w.Map.TitleMode = false
	g.diagMoney = true
}

// diagAirportGhost (GOLOCO_DIAG_BUILDVEHICLE=airportghost) — DIAGNOSTIC ONLY:
// the airport window's tool over the first spot an airport takes (dry run), as
// the player; the ghost held, rotated once
func (g *Game) diagAirportGhost() {
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
}

// diagDockTool (GOLOCO_DIAG_BUILDVEHICLE=dockghost / docktool) — DIAGNOSTIC
// ONLY: the dock window's tool over the first shore spot a dock takes (dry
// run), as a station owner — hover (ghost held) or click
func (g *Game) diagDockTool(v string) {
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
}

// diagPortMenu (GOLOCO_DIAG_BUILDVEHICLE=portmenu) — DIAGNOSTIC ONLY: the
// toolbar's Port/Airport menu
func (g *Game) diagPortMenu() {
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
}

// diagWaves (GOLOCO_DIAG_BUILDVEHICLE=waves) — DIAGNOSTIC ONLY: the view on
// the map's most open water (the tile with the most water within 6 tiles), for
// waves
func (g *Game) diagWaves() {
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
}

// diagNewIndustry (GOLOCO_DIAG_BUILDVEHICLE=newindustry / newindustrybuild) —
// DIAGNOSTIC ONLY: a farm-type industry by the createIndustry command at a
// searched place (built at once, or — newindustrybuild — under construction),
// the view on it
func (g *Game) diagNewIndustry(v string) {
	g.inTitleScreen = false
	m := &g.w.Map
	om := g.r.ObjMgr
	if id, x, y, ok := m.DiagNewFarmIndustry(om, v == "newindustry"); ok {
		g.w.CenterOn(fmt.Sprintf("tile:%d,%d", x>>5, y>>5))
		log.Printf("[Diag] newindustry: industry %d built at tile (%d, %d)", id, x>>5, y>>5)
	} else {
		log.Printf("[Diag] newindustry: none could be built")
	}
}
