package main

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/sim"
	"github.com/hajimehoshi/ebiten/v2"
)

// vehicle_pickup.go — the Vehicle window's pickup tool: lifting a vehicle and
// placing it on track or road, water, or at an airport, with its ghost
// (Vehicle.cpp pickupTool*).

// onPickup is Common::onPickup for rail: a vehicle not on the map takes the
// placing tool (toolSet; the tool is goloco's pickupHead); a placed one is
// lifted (VehiclePickup) and the window reopened (callOnMouseUp tabDetails).
// PARTIAL: the main view's unfocus, the error title, road / air / water.
func (g *Game) onPickup(headID uint16) {
	m := &g.w.Map
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	h := m.Entities[tr.Head]
	if h.TileX == -1 {
		g.closeTerraform()                                           // toolSet: the tool is this window's now
		g.pickupOn, g.pickupHead, g.ghostProgress = true, headID, -1 // toolSet; _ghostTrackProgress = -1
		g.ghostAirStation = 0xFFFF                                   // _ghostAirportStationId = null
		g.openVehicleDetails(headID)
		return
	}
	c := g.playerCmd()
	if _, ok := m.VehiclePickup(c, sim.CmdApply, headID); !ok {
		log.Printf("[Game] can't remove vehicle %d", headID)
		return
	}
	g.openVehicleDetails(headID)
}

// pickupMode is the picked-up vehicle's TransportMode (0xFF: none).
func (g *Game) pickupMode() uint8 {
	tr, err := g.w.Map.TrainOf(g.pickupHead)
	if err != nil {
		return 0xFF
	}
	return g.w.Map.Entities[tr.Head].Mode
}

// pickupToolDown is pickupToolDownLand for rail: the placement under the
// cursor, VehiclePlace, then pickupToolPlacementCommandCallback — shift starts
// the vehicle, the tool is cancelled, the window shown again. false: no track
// there (the tool stays); rail or road by the vehicle's mode (World.PlacementAt).
// PARTIAL: air and water, the error window.
func (g *Game) pickupToolDown(sx, sy int) bool {
	head := g.pickupHead
	switch g.pickupMode() {
	case 2:
		return g.pickupToolDownAir(sx, sy)
	case 3:
		return g.pickupToolDownWater(sx, sy)
	}
	p, ok := g.w.PlacementAt(head, sx, sy)
	if !ok {
		return false
	}
	m := &g.w.Map
	c := g.playerCmd()
	if g.ghostProgress != -1 && g.ghostPlace.TrackAndDirection == p.TrackAndDirection && g.ghostPlace.X == p.X &&
		g.ghostPlace.Y == p.Y && g.ghostPlace.Z == p.Z && int(g.ghostPlace.TrackProgress) == int(p.TrackProgress) && m.VehicleIsGhost(head) {
		p.ConvertGhost = true // the ghost where it stands becomes the vehicle
	} else {
		g.removeLandGhost()
	}
	g.ghostProgress = -1
	if _, ok := m.VehiclePlace(c, sim.CmdApply, p); !ok {
		log.Printf("[Game] can't place vehicle %d here", head)
		return false
	}
	g.pickupPlaced(c, head)
	return true
}

// pickupPlaced is pickupToolPlacementCommandCallback: shift starts the placed
// vehicle, the tool is cancelled, the window shown again on Main.
func (g *Game) pickupPlaced(c *sim.Cmd, head uint16) {
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.w.Map.VehicleChangeRunningMode(c, sim.CmdApply, head, sim.RunningModeStart)
	}
	g.pickupOn = false
	g.openVehicleWindow(head, 0) // callOnMouseUp(tabMain)
}

// pickupToolUpdate is pickupToolUpdateLand for rail and road: the placement under the
// cursor; unchanged since the last ghost, nothing; else the old ghost lifted
// and the train placed there as a ghost (apply | ghost | noErrorWindow).
func (g *Game) pickupToolUpdate(sx, sy int) {
	switch g.pickupMode() {
	case 2:
		g.pickupToolUpdateAir(sx, sy)
		return
	case 3:
		g.pickupToolUpdateWater(sx, sy)
		return
	}
	p, ok := g.w.PlacementAt(g.pickupHead, sx, sy)
	if !ok {
		g.removeLandGhost()
		return
	}
	if g.ghostProgress != -1 && g.ghostPlace.TrackAndDirection == p.TrackAndDirection && g.ghostPlace.X == p.X &&
		g.ghostPlace.Y == p.Y && g.ghostPlace.Z == p.Z && g.ghostProgress == int(p.TrackProgress) {
		return
	}
	g.removeLandGhost()
	c := g.playerCmd()
	if _, ok := g.w.Map.VehiclePlace(c, sim.CmdApply|sim.CmdGhost|sim.CmdNoErrorWindow, p); ok {
		g.ghostPlace, g.ghostProgress = p, int(p.TrackProgress)
	}
}

// removeLandGhost is Vehicle.cpp's: a ghost-placed vehicle lifted again
// (VehiclePickup, apply | noErrorWindow | ghost); _ghostTrackProgress = -1.
func (g *Game) removeLandGhost() {
	if g.w.Map.VehicleIsGhost(g.pickupHead) {
		c := g.playerCmd()
		g.w.Map.VehiclePickup(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdGhost, g.pickupHead)
	}
	g.ghostProgress = -1
}

// pickupToolAbort is Vehicle.cpp's (0x004B3035): the tool's ghost, if any,
// lifted by its own transport mode's command — land (VehiclePickup), air
// (VehiclePickupAir) or water (the boat's).
func (g *Game) pickupToolAbort() {
	switch g.pickupMode() {
	case 2:
		g.removeAirplaneGhost()
	case 3:
		g.removeBoatGhost()
	default:
		g.removeLandGhost()
	}
}

// pickupToolDownWater is Vehicle.cpp's: the dock nearest the cursor; the
// ghost standing there becomes the boat (convertGhost), else the ghost goes
// and the boat is placed; then as pickupToolPlacementCommandCallback.
// PARTIAL: pickupToolError's window.
func (g *Game) pickupToolDownWater(sx, sy int) bool {
	head := g.pickupHead
	p, ok := g.w.WaterPlacementAt(head, sx, sy)
	if !ok {
		return false
	}
	m := &g.w.Map
	if g.ghostProgress == 0 && g.ghostWater.X == p.X && g.ghostWater.Y == p.Y && g.ghostWater.Z == p.Z && m.VehicleIsGhost(head) {
		p.ConvertGhost = true
	} else {
		g.removeBoatGhost()
	}
	g.ghostProgress = -1
	c := g.playerCmd()
	if _, ok := m.VehiclePlaceWater(c, sim.CmdApply, p); !ok {
		log.Printf("[Game] can't place vehicle %d here", head)
		return false
	}
	g.pickupPlaced(c, head)
	return true
}

// pickupToolUpdateWater is Vehicle.cpp's (0x004B2B9E): the dock nearest the
// cursor (World.WaterPlacementAt); the same as the last ghost, nothing; else
// the old ghost lifted and the boat placed there as a ghost
// (_ghostTrackProgress 0 marks a boat ghost).
func (g *Game) pickupToolUpdateWater(sx, sy int) {
	p, ok := g.w.WaterPlacementAt(g.pickupHead, sx, sy)
	if !ok {
		g.removeBoatGhost()
		return
	}
	if g.ghostProgress == 0 && g.ghostWater.X == p.X && g.ghostWater.Y == p.Y && g.ghostWater.Z == p.Z {
		return
	}
	g.ghostWater = p
	g.removeBoatGhost()
	c := g.playerCmd()
	if _, ok := g.w.Map.VehiclePlaceWater(c, sim.CmdApply|sim.CmdGhost|sim.CmdNoErrorWindow, p); ok {
		g.ghostProgress = 0
	}
}

// removeBoatGhost is Vehicle.cpp's: a ghost boat lifted
// (VehiclePickupWater, apply | noErrorWindow | ghost); _ghostTrackProgress = -1.
func (g *Game) removeBoatGhost() {
	if g.w.Map.VehicleIsGhost(g.pickupHead) {
		c := g.playerCmd()
		g.w.Map.VehiclePickupWater(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdGhost, g.pickupHead)
	}
	g.ghostProgress = -1
}

// pickupToolDownAir is Vehicle.cpp's (0x004B2E18): the terminal nearest the
// cursor; the ghost standing there becomes the aircraft (convertGhost), else
// the ghost goes and it is placed; then as
// pickupToolPlacementCommandCallback. PARTIAL: pickupToolError's window.
func (g *Game) pickupToolDownAir(sx, sy int) bool {
	head := g.pickupHead
	p, ok := g.w.AirPlacementAt(head, sx, sy)
	if !ok {
		return false
	}
	m := &g.w.Map
	if g.ghostAirStation == p.StationID && g.ghostAirNode == p.AirportNode && m.VehicleIsGhost(head) {
		p.ConvertGhost = true
	}
	if !p.ConvertGhost {
		g.removeAirplaneGhost()
	}
	c := g.playerCmd()
	if _, ok := m.VehiclePlaceAir(c, sim.CmdApply, p); !ok {
		log.Printf("[Game] can't place vehicle %d here", head)
		return false
	}
	g.pickupPlaced(c, head)
	return true
}

// pickupToolUpdateAir is Vehicle.cpp's: the airport terminal nearest the
// cursor (World.AirPlacementAt); the same as the ghost's, nothing; else the
// old ghost lifted and the aircraft placed there as a ghost
// (_ghostAirportStationId / _ghostAirportNode).
func (g *Game) pickupToolUpdateAir(sx, sy int) {
	p, ok := g.w.AirPlacementAt(g.pickupHead, sx, sy)
	if !ok {
		g.removeAirplaneGhost()
		return
	}
	if g.ghostAirStation != 0xFFFF && g.ghostAirStation == p.StationID && g.ghostAirNode == p.AirportNode {
		return
	}
	g.removeAirplaneGhost()
	c := g.playerCmd()
	if _, ok := g.w.Map.VehiclePlaceAir(c, sim.CmdApply|sim.CmdGhost|sim.CmdNoErrorWindow, p); ok {
		g.ghostAirStation, g.ghostAirNode = p.StationID, p.AirportNode
	}
}

// removeAirplaneGhost is Vehicle.cpp's: a ghost aircraft lifted
// (VehiclePickupAir, apply | noErrorWindow | ghost); no ghost airport.
func (g *Game) removeAirplaneGhost() {
	if g.w.Map.VehicleIsGhost(g.pickupHead) {
		c := g.playerCmd()
		g.w.Map.VehiclePickupAir(c, sim.CmdApply|sim.CmdNoErrorWindow|sim.CmdGhost, g.pickupHead)
	}
	g.ghostAirStation = 0xFFFF
}
