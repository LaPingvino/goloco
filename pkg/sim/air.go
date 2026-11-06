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
