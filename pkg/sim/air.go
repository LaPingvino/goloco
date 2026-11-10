package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Aircraft (AW.b): VehicleHead.cpp's airport movement — the airport
// object's movement graph (objects/airport.go) placed at the station's
// airport, the approach parameters and the status an aircraft takes.
// PARTIAL: VehicleHead::updateAir itself (airportGetNextMovementEdge,
// sub_4A9348, airplaneApproachTarget, airplaneLoadingUpdate) — AW.b.

// airportStationElement is the station element of the airport at the
// station's airportStartPos tile and height (any station element there, as
// upstream takes the first at that baseZ).
func (m *Map) airportStationElement(stationID uint16) (*scenario.StationElement, [3]int16, bool) {
	if int(stationID) >= len(m.StationNames) {
		return nil, [3]int16{}, false
	}
	pos := m.StationNames[stationID].AirportStartPos
	t := m.TileAt(int(pos[0])>>5, int(pos[1])>>5)
	if t == nil {
		return nil, pos, false
	}
	for i := range t.Stations {
		if int16(t.Stations[i].BaseZ) == pos[2]/4 {
			return &t.Stations[i], pos, true
		}
	}
	return nil, pos, false
}

// AirportGetMovementEdgeTarget is airportGetMovementEdgeTarget (0x00426E26):
// the flags and world position of the node edge curEdge leads to, at the
// target station's airport — the node rotated about the tile's centre by
// the airport's rotation, from its start position; on the ground (taxiing)
// at the node's height above the airport, else at the airport's height
// + 255 in flight, or 960 (above everything) for the nodes neither taxiing
// nor in flight. ok false: no airport there (upstream asserts).
func (m *Map) AirportGetMovementEdgeTarget(om *objects.ObjectManager, stationID uint16, curEdge uint8) (flags uint16, x, y, z int32, ok bool) {
	el, pos, found := m.airportStationElement(stationID)
	if !found || om == nil || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil {
		return 0, 0, 0, 0, false
	}
	a := om.AirportObjects[el.ObjectID]
	if int(curEdge) >= len(a.MovementEdges) || int(a.MovementEdges[curEdge].NextNode) >= len(a.MovementNodes) {
		return 0, 0, 0, 0, false
	}
	node := a.MovementNodes[a.MovementEdges[curEdge].NextNode]
	rx, ry := rotOffset(node.X-16, node.Y-16, el.Rotation)
	x, y = int32(rx)+16+int32(pos[0]), int32(ry)+16+int32(pos[1])
	z = int32(node.Z) + int32(pos[2])
	if node.Flags&objects.AirportNodeTaxiing == 0 {
		z = int32(pos[2]) + 255
		if node.Flags&objects.AirportNodeInFlight == 0 {
			z = 960
		}
	}
	return node.Flags, x, y, z, true
}

const (
	airportEdgeNull    = 0xFF   // kAirportMovementNodeNull
	stationFlagAirport = 1 << 6 // StationFlags::flag_6: the station has an airport
)

// airplaneApproach is AirplaneApproachTargetParams.
type airplaneApproach struct {
	targetYaw        uint8
	targetZ          int32
	manhattan        int32
	isHeliTakeOffEnd bool
}

// airplaneApproachParams is VehicleHead::sub_427122: where the aircraft
// heads — its movement edge's target at its airport; or, flying to a
// station (its current order's, when it has none of its own), that
// station's airport start (its position when it has no airport) at 960;
// with no station at all, the middle of the map at 960 — and the yaw to
// it, its height and the Manhattan distance.
func (m *Map) airplaneApproachParams(tr Train, om *objects.ObjectManager) airplaneApproach {
	head := &m.Entities[tr.Head]
	var res airplaneApproach
	target := uint16(stationIDNull)
	var tx, ty, tz int32
	have := false
	if head.StationID == stationIDNull {
		if orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder); len(orders) > 0 {
			o := orders[0]
			if o.Type() != OrderRouteWaypoint && orderFlags[o.Type()&7]&OrderHasStation != 0 {
				target = uint16(o.Station())
			}
		}
	} else if head.AirportMovementEdge == airportEdgeNull || int(head.StationID) >= len(m.StationNames) || m.StationNames[head.StationID].Flags&stationFlagAirport == 0 {
		target = head.StationID
	} else if flags, x, y, z, ok := m.AirportGetMovementEdgeTarget(om, head.StationID, head.AirportMovementEdge); ok {
		res.isHeliTakeOffEnd = flags&objects.AirportNodeHeliTakeoffEnd != 0
		tx, ty, tz, have = x, y, z, true
	}
	if !have {
		tx, ty, tz = 6143, 6143, 960
		if target != stationIDNull && int(target) < len(m.StationNames) {
			st := &m.StationNames[target]
			tx, ty = int32(st.X), int32(st.Y)
			if st.Flags&stationFlagAirport != 0 {
				tx, ty = int32(st.AirportStartPos[0]), int32(st.AirportStartPos[1])
			}
		}
	}
	dx, dy := tx-int32(head.PosX), ty-int32(head.PosY)
	res.targetYaw = Yaw1Plane(int16(dx), int16(dy))
	res.targetZ = tz
	res.manhattan = abs32(dx) + abs32(dy)
	return res
}

// airplaneGetNewStatus is VehicleHead::airplaneGetNewStatus (0x004273DF):
// the status and target speed (mph) from the movement edge the aircraft is
// on — its var_03 (1 landing at the slow speed, 3 landing to a stop, 4
// taxiing in at 20, else approaching) unless it is 2 on an edge whose
// var_00 isn't 1; else its var_00 (2 taking off at full speed, 3 taking
// off to a stop, else taxiing out at 20). Off any airport: travelling at
// full speed (the slow one when broken down).
func (m *Map) airplaneGetNewStatus(tr Train, om *objects.ObjectManager) (uint8, int16) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	full := v2.MaxSpeed
	if v2.Var73&flags73BrokenDown != 0 {
		full = v2.RackRailMaxSpeed
	}
	if head.StationID == stationIDNull || head.AirportMovementEdge == airportEdgeNull {
		return headStatusTravelling, full
	}
	el, _, ok := m.airportStationElement(head.StationID)
	if !ok || om == nil || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil ||
		int(head.AirportMovementEdge) >= len(om.AirportObjects[el.ObjectID].MovementEdges) {
		return headStatusTravelling, v2.MaxSpeed // upstream asserts
	}
	e := om.AirportObjects[el.ObjectID].MovementEdges[head.AirportMovementEdge]
	al, cl := e.Var03, e.Var00
	if al != 0 && (cl == 1 || al != 2) {
		switch al {
		case 1:
			return headStatusLanding, v2.RackRailMaxSpeed
		case 3:
			return headStatusLanding, 0
		case 4:
			return headStatusTaxiing1, 20
		default:
			return headStatusApproaching, v2.RackRailMaxSpeed
		}
	}
	switch cl {
	case 2:
		return headStatusTakingOff, full
	case 3:
		return headStatusTakingOff, 0
	}
	return headStatusTaxiing2, 20
}

const (
	airportEdgeNoValid = 0xFE // kAirportMovementNoValidEdge

	vehicleFlagAircraftIsHelicopter = 1 << 13 // VehicleObjectFlags::aircraftIsHelicopter

	airportFlagAcceptsLightPlanes = 1 << 2 // AirportObjectFlags
	airportFlagAcceptsHeavyPlanes = 1 << 3
	airportFlagAcceptsHelicopter  = 1 << 4
)

// aircraftObject is the aircraft's vehicle object (its first car's).
func (m *Map) aircraftObject(tr Train, om *objects.ObjectManager) *objects.VehicleObject {
	if om == nil || len(tr.Cars) == 0 || len(tr.Cars[0]) == 0 {
		return nil
	}
	id := int(m.Entities[tr.Cars[0][0].Body].ObjectID)
	if id >= len(om.Vehicles) {
		return nil
	}
	return om.Vehicles[id]
}

// compatibleAirportType is VehicleBogie::getCompatibleAirportType: the
// airport flags that take it — helicopters a helicopter pad, planes under
// 50 t light or heavy runways, the rest heavy ones.
func compatibleAirportType(v *objects.VehicleObject) uint16 {
	switch {
	case v.Flags&vehicleFlagAircraftIsHelicopter != 0:
		return airportFlagAcceptsHelicopter
	case v.Weight < 50:
		return airportFlagAcceptsHeavyPlanes | airportFlagAcceptsLightPlanes
	}
	return airportFlagAcceptsHeavyPlanes
}

// edgeFree is the occupancy test both searches share: none of the edges
// that must be clear is occupied, and — when any are named — not all of
// those of which one must be clear.
func edgeFree(e objects.AirportMovementEdge, occupied uint32) bool {
	if occupied&e.MustBeClear != 0 {
		return false
	}
	return e.AtLeastOneClear == 0 || occupied&e.AtLeastOneClear != e.AtLeastOneClear
}

// airportGetNextMovementEdge is VehicleHead::airportGetNextMovementEdge
// (0x00427214): arriving (no edge yet), the first free edge from an entry
// node (flag2); on an edge, airportEdgeNull when a take-off has reached the
// runway's end (it flies), else the first free edge on from the node it
// leads to — helicopters never onto a runway's take-off, planes never onto
// a helipad's; airportEdgeNoValid when none is free.
func (m *Map) airportGetNextMovementEdge(tr Train, curEdge uint8, om *objects.ObjectManager) uint8 {
	head := &m.Entities[tr.Head]
	el, _, ok := m.airportStationElement(head.StationID)
	if !ok || om == nil || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil {
		return airportEdgeNull // upstream asserts
	}
	a := om.AirportObjects[el.ObjectID]
	occupied := m.StationNames[head.StationID].AirportOccupiedEdges
	if curEdge == airportEdgeNull {
		for i, e := range a.MovementEdges {
			if a.MovementNodes[e.CurNode].Flags&objects.AirportNodeFlag2 != 0 && edgeFree(e, occupied) {
				return uint8(i)
			}
		}
		return airportEdgeNoValid
	}
	target := a.MovementEdges[curEdge].NextNode
	if head.Status == headStatusTakingOff && a.MovementNodes[target].Flags&objects.AirportNodeTakeoffEnd != 0 {
		return airportEdgeNull
	}
	avoid := uint16(objects.AirportNodeHeliTakeoffBegin) // a plane: no helipad take-offs
	if v := m.aircraftObject(tr, om); v != nil && v.Flags&vehicleFlagAircraftIsHelicopter != 0 {
		avoid = objects.AirportNodeTakeoffBegin // a helicopter: no runway take-offs
	}
	for i, e := range a.MovementEdges {
		if e.CurNode != target || a.MovementNodes[e.NextNode].Flags&avoid != 0 {
			continue
		}
		if edgeFree(e, occupied) {
			return uint8(i)
		}
	}
	return airportEdgeNoValid
}

// beginNewJourney is VehicleHead::beginNewJourney: the journey's start time
// and place (Vehicle2's) for updateLastJourneyAverageSpeed.
func (m *Map) beginNewJourney(tr Train) {
	head, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh2]
	head.JourneyStartTicks = m.ScenarioTicks
	head.JourneyStartX, head.JourneyStartY = int16(v2.PosX), int16(v2.PosY)
	head.BreakdownFlags |= BreakdownJourneyStarted
}

// airplaneTakeEdge is sub_4A9348's state change (its every path then calls
// airplaneApproachTarget, which the caller does): the edge it was on freed;
// a new edge taken and marked occupied — or, leaving the airport
// (airportEdgeNull), a journey begun and, when its next stop is an airport
// that takes it (any one for an AI's aircraft), that station its target —
// else, for the controlling company, unableToLandAtAirport is posted.
func (m *Map) airplaneTakeEdge(tr Train, newEdge uint8, om *objects.ObjectManager) {
	head := &m.Entities[tr.Head]
	if head.StationID != stationIDNull && head.AirportMovementEdge != airportEdgeNull && int(head.StationID) < len(m.StationNames) {
		m.StationNames[head.StationID].AirportOccupiedEdges &^= 1 << head.AirportMovementEdge
	}
	if newEdge != airportEdgeNull {
		head.AirportMovementEdge = newEdge
		if head.StationID != stationIDNull && int(head.StationID) < len(m.StationNames) {
			m.StationNames[head.StationID].AirportOccupiedEdges |= 1 << newEdge
		}
		return
	}
	m.beginNewJourney(tr)
	head.AirportMovementEdge = airportEdgeNull
	if head.SizeOfOrderTable == 1 {
		return
	}
	orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder)
	if len(orders) == 0 || orders[0].Type() != OrderStopAt {
		return
	}
	sid := uint16(orders[0].Station())
	if int(sid) >= len(m.StationNames) || m.StationNames[sid].Flags&stationFlagAirport == 0 {
		return
	}
	if !m.IsPlayerCompany(int(head.Owner)) {
		head.StationID = sid
		return
	}
	el, _, ok := m.airportStationElement(sid)
	if !ok || om == nil || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil {
		return // upstream asserts
	}
	if v := m.aircraftObject(tr, om); v != nil && om.AirportObjects[el.ObjectID].Flags&compatibleAirportType(v) != 0 {
		head.StationID = sid
	} else if head.Owner == m.ControllingCompany {
		m.PostMessage(om, MsgUnableToLandAtAirport, head.Owner, head.ID, sid, 0xFFFF)
	}
}
