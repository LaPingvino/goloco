package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/olmath"
)

// Aircraft fly (AW.b): VehicleHead::updateAir (VehicleHead.cpp 0x004A9051)
// with airplaneApproachTarget (0x004A94A9), sub_4A95CB (0x004A95CB),
// airplaneLoadingUpdate (0x004A95F5) and movePlaneTo (0x00426CA4) — the
// edges and statuses are in air.go. As for boats, an aircraft's tick is
// the head's: var_5C, the breakdown, updateAir. PARTIAL: the body's
// sub_4AAB0B (animation; not ported for land either), the touchdown sound.

const (
	motorAirplaneAtTaxiSpeed uint8 = 5 // MotorState::airplaneAtTaxiSpeed
	pitchUp12deg             uint8 = 2 // Pitch::up12deg

	vehicleFlagAircraftIsTaildragger = 1 << 8  // VehicleObjectFlags
	vehicleFlagAircraftFlaresLanding = 1 << 11 // (Concorde)
)

// updateAircraft is VehicleHead::update for an aircraft.
func (m *Map) updateAircraft(tr Train, om *objects.ObjectManager) {
	head := &m.Entities[tr.Head]
	if head.Var5C != 0 {
		head.Var5C--
	}
	m.updateBreakdown(tr, om)
	m.updateAir(tr, om)
}

// updateAir is VehicleHead::updateAir: stopped, it calls at its station;
// (un)loading it does that; else its status and target speed from its edge,
// the speed towards it (+2 mph a tick; braking by 2, 5, 11 or 25 by speed),
// the yaw and pitch a step towards the target (nose up when slow for a
// taildragger, climbing under 350, flaring under 180 for a plane that
// flares), then: out of reach of its target, on towards it; at a terminal,
// arrival; else the next edge taken (or, none free, a stop when slow).
func (m *Map) updateAir(tr Train, om *objects.ObjectManager) {
	head, v1, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh1], &m.Entities[tr.Veh2]
	switch head.Status {
	case headStatusStopped:
		if head.VehicleFlags&VehicleFlagCommandStop == 0 {
			m.callAtAirport(tr)
		}
		return
	case headStatusUnloading:
		m.UpdateUnloadCargo(tr, om)
		return
	case headStatusLoading:
		m.airplaneLoadingUpdate(tr, om)
		return
	}
	head.Status = headStatusTravelling
	st, target := m.airplaneGetNewStatus(tr, om)
	head.Status = st
	v1.TargetSpeed = target
	m.AdvanceToNextRoutableOrder(head.OrderTableOffset, &head.CurrentOrder, head.SizeOfOrderTable)
	t1 := int32(v1.TargetSpeed) * mph
	switch s := v2.CurrentSpeed; {
	case s == t1:
		v2.MotorState = motorAirplaneAtTaxiSpeed
		if s != 20*mph {
			v2.MotorState = MotorCoasting
		}
	case s > t1:
		v2.MotorState = MotorCoasting
		dec := 2 * mph
		if s >= 130*mph {
			dec = 5 * mph
			if s >= 400*mph {
				dec = 11 * mph
				if s >= 600*mph {
					dec = 25 * mph
				}
			}
		}
		if t1 == 20*mph {
			v2.MotorState = MotorBraking
		}
		v2.CurrentSpeed = max(t1, max(0, s-dec))
	default:
		v2.MotorState = MotorAccelerating
		v2.CurrentSpeed = min(s+2*mph, t1)
	}
	p := m.airplaneApproachParams(tr, om)
	targetYaw := p.targetYaw
	if p.isHeliTakeOffEnd { // a helicopter rises or sinks on the spot
		targetYaw = head.SpriteYaw
		v2.MotorState = MotorAccelerating
		if p.targetZ < int32(head.PosZ) {
			v2.MotorState = MotorCoasting
		}
	}
	if targetYaw != head.SpriteYaw {
		if (targetYaw-head.SpriteYaw)&0x3F > 0x20 {
			head.SpriteYaw--
		} else {
			head.SpriteYaw++
		}
		head.SpriteYaw &= 0x3F
	}
	targetPitch := uint8(0)
	obj := m.aircraftObject(tr, om)
	if obj != nil && v2.CurrentSpeed < 50*mph && obj.Flags&vehicleFlagAircraftIsTaildragger != 0 {
		targetPitch = pitchUp12deg
	}
	if p.targetZ > int32(head.PosZ) && v2.CurrentSpeed <= 350*mph {
		targetPitch = pitchUp12deg
	}
	if obj != nil && p.targetZ < int32(head.PosZ) && v2.CurrentSpeed <= 180*mph && obj.Flags&vehicleFlagAircraftFlaresLanding != 0 {
		targetPitch = pitchUp12deg
	}
	switch {
	case targetPitch < head.SpritePitch:
		head.SpritePitch--
	case targetPitch > head.SpritePitch:
		head.SpritePitch++
	}
	if p.isHeliTakeOffEnd {
		v2.CurrentSpeed = 8 * mph
		if p.targetZ != int32(head.PosZ) {
			m.airplaneApproachTarget(tr, p)
			return
		}
	} else {
		tolerance := int32(480)
		if head.AirportMovementEdge != airportEdgeNull {
			tolerance = 5
			if v2.CurrentSpeed >= 70*mph {
				tolerance = 30
			}
		}
		if p.manhattan > tolerance {
			m.airplaneApproachTarget(tr, p)
			return
		}
	}
	if head.StationID != stationIDNull && head.AirportMovementEdge != airportEdgeNull {
		if flags, _, _, _, ok := m.AirportGetMovementEdgeTarget(om, head.StationID, head.AirportMovementEdge); ok {
			// touchdown: produceTouchdownAirportSound (PARTIAL)
			if flags&objects.AirportNodeTaxiing != 0 {
				m.updateLastJourneyAverageSpeed(tr)
			}
			if flags&objects.AirportNodeTerminal != 0 {
				m.airplaneAtTerminal(tr)
				return
			}
		}
	}
	edge := uint8(airportEdgeNull)
	if head.StationID != stationIDNull {
		edge = head.AirportMovementEdge
	}
	if next := m.airportGetNextMovementEdge(tr, edge, om); next != airportEdgeNoValid {
		m.airplaneTakeEdge(tr, next, om)
		m.airplaneApproachTarget(tr, p)
		return
	}
	if v2.CurrentSpeed > 30*mph {
		m.airplaneApproachTarget(tr, p)
		return
	}
	v2.CurrentSpeed, v2.MotorState = 0, MotorStopped
}

// callAtAirport is setStationVisitedTypes, checkIfAtOrderStation and
// beginUnloading (updateAir's stopped case, sub_4A95CB).
func (m *Map) callAtAirport(tr Train) {
	head := &m.Entities[tr.Head]
	if int(head.StationID) < len(m.StationNames) {
		m.StationNames[head.StationID].VisitedTypes |= 1 << (head.VehicleType & 7)
	}
	if orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder); len(orders) > 1 { // checkIfAtOrderStation
		if t := orders[0].Type(); (t == OrderStopAt || t == OrderRouteThrough) && uint16(orders[0].Station()) == head.StationID {
			head.CurrentOrder = uint16(orders[1].Offset - int(head.OrderTableOffset))
		}
	}
	m.beginUnloading(tr)
}

// airplaneAtTerminal is sub_4A95CB: told to stop, it stops; else it calls.
func (m *Map) airplaneAtTerminal(tr Train) {
	head := &m.Entities[tr.Head]
	if head.VehicleFlags&VehicleFlagCommandStop != 0 {
		head.Status = headStatusStopped
		m.Entities[tr.Veh2].CurrentSpeed = 0
		return
	}
	m.callAtAirport(tr)
}

// airplaneLoadingUpdate is VehicleHead::airplaneLoadingUpdate: stopped
// while loading; loaded, its next order and status, and — when an edge is
// free — on to it (approaching at its own height, the params upstream
// leaves uninitialised fixed so); else it waits, loading.
func (m *Map) airplaneLoadingUpdate(tr Train, om *objects.ObjectManager) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	v2.CurrentSpeed, v2.MotorState = 0, MotorStopped
	if m.UpdateLoadCargo(tr, om) {
		return
	}
	m.AdvanceToNextRoutableOrder(head.OrderTableOffset, &head.CurrentOrder, head.SizeOfOrderTable)
	head.Status = headStatusTravelling
	head.Status, _ = m.airplaneGetNewStatus(tr, om)
	edge := uint8(airportEdgeNull)
	if head.StationID != stationIDNull {
		edge = head.AirportMovementEdge
	}
	if next := m.airportGetNextMovementEdge(tr, edge, om); next != airportEdgeNoValid {
		m.airplaneTakeEdge(tr, next, om)
		m.airplaneApproachTarget(tr, airplaneApproach{targetZ: int32(m.Entities[tr.Head].PosZ)})
		return
	}
	head.Status = headStatusLoading
}

// airplaneApproachTarget is VehicleHead::airplaneApproachTarget: the head
// moved by its speed along its yaw (a helicopter at its take-off end along
// the target yaw) — calculateNextPosition, Vehicle1's sub-pixel parts —
// and its height towards the target: the last stretch (within 28) by 1–3 a
// tick with the speed, else sloped so as to arrive there.
func (m *Map) airplaneApproachTarget(tr Train, p airplaneApproach) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	yaw := head.SpriteYaw
	if p.isHeliTakeOffEnd {
		yaw = p.targetYaw
	}
	dx, dy := olmath.ComputeXYVector(v2.CurrentSpeed>>5, yaw)
	fx, fy := veh1Frac(m, tr)
	bx := int32(fx) + int32(head.PosX)<<16 + dx
	by := int32(fy) + int32(head.PosY)<<16 + dy
	setVeh1Frac(m, tr, uint16(bx), uint16(by))
	z := p.targetZ
	pz := int32(head.PosZ)
	if p.targetZ != pz {
		if p.manhattan <= 28 {
			shift := int32(1)
			if v2.CurrentSpeed >= 50*mph {
				shift++
				if v2.CurrentSpeed >= 100*mph {
					shift++
				}
			}
			if p.targetZ < pz {
				z = max(p.targetZ, pz-shift)
			} else {
				z = min(p.targetZ, pz+shift)
			}
		} else {
			p1 := ((p.targetZ - pz) * (v2.CurrentSpeed >> 16)) >> 5
			p2 := p.manhattan - 18
			if p1%p2 < 0 {
				z = pz + p1/p2 - 1
			} else {
				z = pz + p1/p2 + 1
			}
		}
	}
	m.movePlaneTo(tr, bx>>16, by>>16, int32(int16(z)), head.SpriteYaw, head.SpritePitch)
}

// movePlaneTo is VehicleHead::movePlaneTo: the head, Vehicle1, Vehicle2 and
// the body to the point (tileX 0), the first bogie — its shadow — on the
// ground beneath, flat, all facing yaw (the body pitched).
func (m *Map) movePlaneTo(tr Train, x, y, z int32, yaw, pitch uint8) {
	for _, i := range []int{tr.Head, tr.Veh1, tr.Veh2} {
		m.moveTo(&m.Entities[i], x, y, z)
		m.Entities[i].TileX = 0
	}
	if len(tr.Cars) == 0 || len(tr.Cars[0]) == 0 {
		return
	}
	c := tr.Cars[0][0]
	land, _ := m.GetHeight(x, y)
	shadow := &m.Entities[c.Front]
	m.moveTo(shadow, x, y, int32(land))
	shadow.SpriteYaw, shadow.SpritePitch, shadow.TileX = yaw, 0, 0
	body := &m.Entities[c.Body]
	m.moveTo(body, x, y, z)
	body.SpriteYaw, body.SpritePitch, body.TileX = yaw, pitch, 0
}
