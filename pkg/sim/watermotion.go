package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
)

// Boats move (AW.a): VehicleHead::updateWater (VehicleHead.cpp 0x00427122
// region, 2330) and updateWaterMotion (2833) — Vehicle1 / Vehicle2 / bodies /
// bogies / the tail do nothing for water vehicles (their update returns at
// once), so a boat's tick is the head's: var_5C, the breakdown, updateWater.
// PARTIAL: the body's sub_4AAB0B (animation; not ported for land either),
// produceLeavingDockSound, the driving sounds.

const (
	waterStopping    = 1 << 0 // WaterMotionFlags::isStopping
	waterLeavingDock = 1 << 1 // isLeavingDock
	waterReachedDest = 1 << 2 // hasReachedADestination
	waterReachedDock = 1 << 3 // hasReachedDock
)

// veh1Frac is Vehicle1's var_4E / var_50: the sub-pixel parts of its
// position, in the bytes the head reads as trainAcceptedCargoTypes.
func veh1Frac(m *Map, tr Train) (uint16, uint16) {
	v := m.Entities[tr.Veh1].TrainAcceptedCargoTypes
	return uint16(v), uint16(v >> 16)
}

func setVeh1Frac(m *Map, tr Train, x, y uint16) {
	m.Entities[tr.Veh1].TrainAcceptedCargoTypes = uint32(x) | uint32(y)<<16
}

// updateBoat is VehicleHead::update for a water vehicle.
func (m *Map) updateBoat(tr Train, om *objects.ObjectManager) {
	head := &m.Entities[tr.Head]
	if head.Var5C != 0 {
		head.Var5C--
	}
	m.updateBreakdown(tr, om)
	m.updateWater(tr, om)
}

// updateWater is VehicleHead::updateWater.
func (m *Map) updateWater(tr Train, om *objects.ObjectManager) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	if head.Status == StatusStopped {
		if head.VehicleFlags&VehicleFlagCommandStop != 0 {
			return
		}
		if head.StationID != stationIDNull {
			v2.CurrentSpeed = 0
			m.arriveAtDock(tr)
			return
		}
	}
	if head.VehicleFlags&VehicleFlagCommandStop != 0 {
		if m.updateWaterMotion(tr, waterStopping, om)&waterReachedDest == 0 {
			return
		}
		head.Status = StatusStopped
		v2.CurrentSpeed, v2.MotorState = 0, MotorStopped
		return
	}
	switch head.Status {
	case headStatusUnloading:
		m.UpdateUnloadCargo(tr, om)
	case headStatusLoading:
		if m.UpdateLoadCargo(tr, om) {
			return
		}
		head.JourneyStartTicks = m.ScenarioTicks // beginNewJourney
		head.JourneyStartX, head.JourneyStartY = int16(v2.PosX), int16(v2.PosY)
		head.BreakdownFlags |= BreakdownJourneyStarted
		m.AdvanceToNextRoutableOrder(head.OrderTableOffset, &head.CurrentOrder, head.SizeOfOrderTable)
		head.Status = m.sub427BF2(tr)
		m.updateWaterMotion(tr, waterLeavingDock, om)
	default:
		head.Status = m.sub427BF2(tr)
		m.AdvanceToNextRoutableOrder(head.OrderTableOffset, &head.CurrentOrder, head.SizeOfOrderTable)
		if m.updateWaterMotion(tr, 0, om)&waterReachedDock == 0 {
			return
		}
		head = &m.Entities[tr.Head]
		if head.VehicleFlags&VehicleFlagCommandStop != 0 {
			head.Status = StatusStopped
			v2.CurrentSpeed, v2.MotorState = 0, MotorStopped
			return
		}
		v2.CurrentSpeed = 0
		m.arriveAtDock(tr)
	}
}

// arriveAtDock is updateWater's arrival: setStationVisitedTypes,
// checkIfAtOrderStation, updateLastJourneyAverageSpeed, beginUnloading.
func (m *Map) arriveAtDock(tr Train) {
	head := &m.Entities[tr.Head]
	if int(head.StationID) < len(m.StationNames) {
		m.StationNames[head.StationID].VisitedTypes |= 1 << (head.VehicleType & 7)
	}
	if orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder); len(orders) > 1 { // checkIfAtOrderStation
		if t := orders[0].Type(); (t == OrderStopAt || t == OrderRouteThrough) && uint16(orders[0].Station()) == head.StationID {
			head.CurrentOrder = uint16(orders[1].Offset - int(head.OrderTableOffset))
		}
	}
	m.updateLastJourneyAverageSpeed(tr)
	m.beginUnloading(tr)
}

// sub427BF2: travelling, or approaching when it has a station.
func (m *Map) sub427BF2(tr Train) uint8 {
	if m.Entities[tr.Head].StationID == stationIDNull {
		return StatusTravelling
	}
	return StatusApproaching
}

// updateWaterMotion is VehicleHead::updateWaterMotion: the water under the
// boat marked as a ship route (no longer industrial), the speed towards the
// target (5 mph near or at a station or stopping, else its maximum unless
// broken down), and — when it has reached the head's target (or is leaving
// a dock) — the next one: route-through waypoints passed, the dock left
// (flag_6 cleared), waterPathfind, a new dock taken (flag_6 set); then the
// yaw turned one step towards the target and Vehicle1 / Vehicle2 moved on.
func (m *Map) updateWaterMotion(tr Train, flags uint8, om *objects.ObjectManager) uint8 {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	if t := m.TileAt(int(v2.PosX)>>5, int(v2.PosY)>>5); t != nil && t.WaterLevel != 0 {
		if t.IsIndustrial {
			t.IsIndustrial = false
			t.GrowthStage = 0
		}
		t.Variation = 0 // setIndustry(IndustryId(0)): byte 7
		t.Type6Flag = true
	}
	target := int32(5 * mph)
	if head.StationID == stationIDNull && flags&waterStopping == 0 && v2.Var73&flags73BrokenDown == 0 {
		target = int32(v2.MaxSpeed) * mph
	}
	switch {
	case target == v2.CurrentSpeed:
		v2.MotorState = MotorCoasting
	case target < v2.CurrentSpeed:
		dec := int32(mph)
		if v2.CurrentSpeed >= 50*mph {
			dec = 3 * mph
		}
		v2.MotorState = MotorBraking
		v2.CurrentSpeed = max(target, max(v2.CurrentSpeed-dec, 0))
	default:
		v2.MotorState = MotorAccelerating
		v2.CurrentSpeed = min(target, v2.CurrentSpeed+21845) // + 0.333333 mph
	}
	dist := abs32(int32(head.PosX)-int32(v2.PosX)) + abs32(int32(head.PosY)-int32(v2.PosY))
	tolerance := int32(3)
	if v2.CurrentSpeed >= 20*mph {
		tolerance = 16
		if v2.CurrentSpeed > 70*mph {
			tolerance = 24
		}
	}
	if flags&waterLeavingDock != 0 || dist <= tolerance {
		flags |= waterReachedDest
		if head.StationID != stationIDNull && flags&waterLeavingDock == 0 {
			flags |= waterReachedDock
		}
		if flags&waterStopping != 0 {
			return flags
		}
		if orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder); len(orders) > 1 && orders[0].Type() == OrderRouteWaypoint {
			if x, y, _ := orders[0].Waypoint(); x == int32(head.PosX)&^0x1F && y == int32(head.PosY)&^0x1F {
				head.CurrentOrder = uint16(orders[1].Offset - int(head.OrderTableOffset))
			}
		}
		if flags&waterLeavingDock == 0 && head.StationID != stationIDNull {
			return flags
		}
		if head.StationID != stationIDNull {
			m.setDockFlag6(int32(head.TileX), int32(head.TileY), head.TileBaseZ, false)
			head.StationID = stationIDNull
		}
		r := m.WaterPathfind(tr, dockBoatPosition(om))
		m.moveTo(head, r.HeadX, r.HeadY, 32)
		if r.StationID != stationIDNull {
			head.StationID = r.StationID
			head.TileX, head.TileY, head.TileBaseZ = coords.WorldUnits(r.StationX), coords.WorldUnits(r.StationY), uint8(r.StationZ/4)
			m.setDockFlag6(r.StationX, r.StationY, uint8(r.StationZ/4), true)
		}
	}
	targetYaw := Yaw4(int16(int32(head.PosX)-int32(v2.PosX)), int16(int32(head.PosY)-int32(v2.PosY)))
	if targetYaw != v2.SpriteYaw {
		if (targetYaw-v2.SpriteYaw)&0x3F > 0x20 {
			v2.SpriteYaw--
		} else {
			v2.SpriteYaw++
		}
		v2.SpriteYaw &= 0x3F
	}
	// calculateNextPosition: Vehicle2's position on 16.16 with Vehicle1's
	// sub-pixel parts, moved by the speed along the yaw.
	dx, dy := olmath.ComputeXYVector(v2.CurrentSpeed>>5, v2.SpriteYaw)
	fx, fy := veh1Frac(m, tr)
	bx := int32(fx) + int32(v2.PosX)<<16 + dx
	by := int32(fy) + int32(v2.PosY)<<16 + dy
	setVeh1Frac(m, tr, uint16(bx), uint16(by))
	m.moveBoatTo(tr, bx>>16, by>>16, int32(v2.PosZ), v2.SpriteYaw, 0)
	return flags
}

// setDockFlag6 sets or clears flag_6 of the dock station element at the tile
// and height (not ghost or AI).
func (m *Map) setDockFlag6(x, y int32, baseZ uint8, on bool) {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return
	}
	for i := range t.Stations {
		el := &t.Stations[i]
		if el.Ghost || el.AiAllocated || el.BaseZ != baseZ {
			continue
		}
		el.Flag6 = on
		return
	}
}

// dockBoatPosition reads the dock objects' boat positions for WaterPathfind.
func dockBoatPosition(om *objects.ObjectManager) func(uint8) (int16, int16, bool) {
	return func(obj uint8) (int16, int16, bool) {
		if om == nil || int(obj) >= len(om.DockObjects) || om.DockObjects[obj] == nil {
			return 0, 0, false
		}
		d := om.DockObjects[obj]
		return d.BoatPosition[0], d.BoatPosition[1], true
	}
}
