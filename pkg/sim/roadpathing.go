package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// Road pathing — VehicleHead.cpp's roadAimlessWanderPathing (0x0047E481),
// roadLongestPathing (0x0047DF4A) with roadLongestPathingCalculate
// (0x0047E72F), and roadTargetedPathing with roadUpdateDistanceToTarget
// (0x0047E582). Like the track walks: ≤ 5 deep, ≤ 1280 weight.
//
// roadPathing (0x0047DFD0, the junction choice), isRoadRoutingResultBetter
// and calculateCompatibleRoadStations follow below.

const raDWithSignalMask = 0x7F | TaDHasSignal // basicRaDWithSignalMask

func (m *Map) roadWanderRecurse(x, y, z int32, tad uint16, p Pathing, om *objects.ObjectManager, st *wanderState) {
	if st.depth >= 5 {
		return
	}
	for {
		st.total += RoadUnkWeighting[((tad>>3)&0xF)%10]
		if st.total > pathingMaxWeight {
			return
		}
		e, ok := RoadConnectionEnd(x, y, z, tad&0x7F)
		if !ok {
			return
		}
		rc := m.RoadConnectionsOneWay(e.X, e.Y, e.Z, e.NextRotation, p.Company, p.TrackType, p.RequiredMods, p.QueryMods, om)
		if len(rc.Connections) == 0 {
			return
		}
		x, y, z = e.X, e.Y, e.Z
		tad = rc.Connections[0] & (TaDBasicMask | TaDHasSignal)
		if len(rc.Connections) == 1 {
			continue
		}
		for _, c := range rc.Connections {
			rs := *st
			rs.depth++
			m.roadWanderRecurse(x, y, z, c&(TaDBasicMask|TaDHasSignal), p, om, &rs)
			st.flags |= rs.flags
		}
		return
	}
}

// RoadAimlessWanderPathing is roadAimlessWanderPathing. Roads have no
// signals, so (as upstream) no flag is ever set: it returns 0.
func (m *Map) RoadAimlessWanderPathing(x, y, z int32, tad uint16, p Pathing, om *objects.ObjectManager) uint16 {
	var st wanderState
	m.roadWanderRecurse(x, y, z, tad, p, om, &st)
	return st.flags
}

func (m *Map) roadLongestRecurse(x, y, z int32, tad uint16, p Pathing, st *longestState) {
	if st.depth >= 5 {
		return
	}
	for {
		// Upstream keeps vanilla's mistake: the track weighting, by road id.
		st.total += TrackUnkWeighting[((tad>>3)&0xF)%44]
		if st.total > pathingMaxWeight {
			return
		}
		st.best = max(st.best, st.total)
		e, ok := RoadConnectionEnd(x, y, z, tad&0x7F)
		if !ok {
			return
		}
		rc := m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, p.Company, p.TrackType, p.RequiredMods, p.QueryMods)
		if len(rc.Connections) == 0 {
			return
		}
		x, y, z = e.X, e.Y, e.Z
		tad = rc.Connections[0] & (TaDBasicMask | TaDHasSignal)
		if len(rc.Connections) == 1 {
			continue
		}
		for _, c := range rc.Connections {
			rs := *st
			rs.depth++
			m.roadLongestRecurse(x, y, z, c&(TaDBasicMask|TaDHasSignal), p, &rs)
			st.best = rs.best
		}
		return
	}
}

// RoadLongestPathing is roadLongestPathing: of a junction's connections, the
// one leading furthest (the last of equals).
func (m *Map) RoadLongestPathing(x, y, z int32, conns []uint16, p Pathing) uint16 {
	var best uint16
	bestI := 0
	for i, c := range conns {
		var st longestState
		m.roadLongestRecurse(x, y, z, c&raDWithSignalMask, p, &st)
		if v := uint16(st.best); v >= best {
			best, bestI = v, i
		}
	}
	return conns[bestI]
}

// roadUpdateDistanceToTarget: Manhattan 3D distance.
func roadUpdateDistanceToTarget(x, y, z int32, t *RouteTarget, st *targetedState) {
	dist := absDiff(int32(int16(x)), int32(int16(t.X))) + absDiff(int32(int16(y)), int32(int16(t.Y))) + absDiff(int32(int16(z)), int32(int16(t.Z)))
	if dist <= uint32(st.result.BestDistToTarget) {
		if dist < uint32(st.result.BestDistToTarget) || st.total <= st.result.BestTrackWeighting {
			st.result.BestDistToTarget = uint16(dist)
			st.result.BestTrackWeighting = st.total
		}
	}
}

func (m *Map) roadTargetedRecurse(x, y, z int32, tad uint16, p Pathing, allowedStationTypes uint32, t *RouteTarget, om *objects.ObjectManager, st *targetedState) {
	curStation, curStationObj := -1, uint8(0)
	if st.depth >= 5 {
		return
	}
	for {
		reached := false
		if t.StationID >= 0 {
			if curStation == t.StationID && allowedStationTypes&(1<<curStationObj) != 0 {
				fwd := m.RoadOccupation(x, y, z, tad, om)
				if fwd&RoadHasStation != 0 {
					if fwd&RoadLaneOccupied == 0 {
						reached = true
					} else if m.RoadOccupation(x, y, z, tad^4, om)&RoadLaneOccupied == 0 {
						reached = true
					}
				}
			}
		} else if x == t.X && y == t.Y && z == t.Z && tad&0x7F == t.TaD {
			reached = true
		} else if x == t.RX && y == t.RY && z == t.RZ && tad&0x7F == t.ReverseTaD {
			reached = true
		}
		if reached {
			if processReachedTargetRouteEnd(st) {
				return
			}
		} else {
			roadUpdateDistanceToTarget(x, y, z, t, st)
		}
		st.total += RoadUnkWeighting[((tad>>3)&0xF)%10]
		if st.total > pathingMaxWeight {
			return
		}
		e, ok := RoadConnectionEnd(x, y, z, tad&0x7F)
		if !ok {
			return
		}
		rc := m.RoadConnectionsOneWay(e.X, e.Y, e.Z, e.NextRotation, p.Company, p.TrackType, p.RequiredMods, p.QueryMods, om)
		if len(rc.Connections) == 0 {
			return
		}
		x, y, z = e.X, e.Y, e.Z
		tad = rc.Connections[0] & 0x807F
		curStation, curStationObj = rc.StationID, rc.StationObjectID
		if len(rc.Connections) == 1 {
			continue
		}
		for _, c := range rc.Connections {
			rs := *st
			rs.depth++
			m.roadTargetedRecurse(x, y, z, c&0x807F, p, allowedStationTypes, t, om, &rs)
			st.result.SignalState = min(st.result.SignalState, rs.result.SignalState)
			st.result.BestDistToTarget = rs.result.BestDistToTarget
			st.result.BestTrackWeighting = rs.result.BestTrackWeighting
		}
		return
	}
}

// RoadTargetedPathing is roadTargetedPathing: a station target counts as
// reached at a stop of an allowed station type with a free lane (either way).
func (m *Map) RoadTargetedPathing(x, y, z int32, tad uint16, p Pathing, allowedStationTypes uint32, t *RouteTarget, om *objects.ObjectManager) RoutingResult {
	st := targetedState{result: RoutingResult{BestDistToTarget: 0xFFFF, BestTrackWeighting: 0xFFFFFFFF, SignalState: RouteNull}}
	m.roadTargetedRecurse(x, y, z, tad, p, allowedStationTypes, t, om, &st)
	return st.result
}

// RoadStationFlags.
const (
	roadStationPassenger uint8 = 1 << 1
	roadStationFreight   uint8 = 1 << 2
)

// cargoSlotByName is the cargo slot a dependency header names (the
// renderer's cargoSlot: case-insensitive, 0 if absent).
func cargoSlotByName(om *objects.ObjectManager, name string) int {
	for i, c := range om.CargoObjects {
		if c != nil && c.Header != nil && sameObject(c.Header.GetName(), name) {
			return i
		}
	}
	return 0
}

// CalculateCompatibleRoadStations is calculateCompatibleRoadStations: the
// road station types (of kMaxRoadStationObjects = 16) this vehicle may use —
// passenger stops for cargo it carries, freight stops for cargo it doesn't
// (upstream's "Eh?" kept), and every other kind.
func CalculateCompatibleRoadStations(trainAcceptedCargoTypes uint32, om *objects.ObjectManager) uint32 {
	var out uint32
	for i := 0; i < 16; i++ {
		rs := om.GetRoadStationObjectByIndex(i)
		if rs == nil {
			continue
		}
		cargo := uint32(1) << (cargoSlotByName(om, rs.CargoName) & 31)
		switch {
		case rs.Flags&roadStationPassenger != 0:
			if trainAcceptedCargoTypes&cargo != 0 {
				out |= 1 << i
			}
		case rs.Flags&roadStationFreight != 0:
			if trainAcceptedCargoTypes&cargo == 0 {
				out |= 1 << i
			}
		default:
			out |= 1 << i
		}
	}
	return out
}

// IsRoadRoutingResultBetter is isRoadRoutingResultBetter.
func IsRoadRoutingResultBetter(base, n RoutingResult) bool {
	newReached, baseReached := n.BestDistToTarget == 0, base.BestDistToTarget == 0
	if newReached && n.SignalState != RouteNull {
		if base.SignalState == RouteNull || !baseReached {
			return true
		}
		return n.BestTrackWeighting < base.BestTrackWeighting
	}
	if baseReached && base.SignalState != RouteNull {
		return false
	}
	if n.SignalState != RouteNull && base.SignalState == RouteNull {
		return true
	}
	if n.BestDistToTarget < base.BestDistToTarget {
		return true
	}
	return n.BestDistToTarget == base.BestDistToTarget && n.BestTrackWeighting < base.BestTrackWeighting
}

// RoadPathing is roadPathing (0x0047DFD0): the junction choice for a road
// vehicle, as trackPathing but with the road searches, the road target
// (getUnkRoad for a waypoint's reverse end), allowed station types, and the
// first result taken on a first run whatever it is.
func (m *Map) RoadPathing(orderTableOffset uint32, currentOrder uint16, x, y, z int32, conns []uint16, p Pathing, allowedStationTypes uint32, isSecondRun bool, st *PathingState, om *objects.ObjectManager) uint16 {
	st.HadNewResult = false
	if !isSecondRun {
		st.Result.SignalState = RouteNull
	}
	randVal := m.Rng1.RandNext()
	var target *RouteTarget
	if ring := m.OrderRing(orderTableOffset, currentOrder); len(ring) > 0 {
		o := ring[0]
		switch o.Type() {
		case OrderStopAt, OrderRouteThrough:
			target = &RouteTarget{StationID: o.Station()}
			if o.Station() < len(m.StationNames) {
				sn := m.StationNames[o.Station()]
				target.X, target.Y, target.Z = int32(sn.X), int32(sn.Y), int32(sn.Z)
			}
		case OrderRouteWaypoint:
			wx, wy, wz := o.Waypoint()
			tad := uint16(o.TrackID())<<3 | uint16(o.Direction())
			c := RoadCoordinates[tad%80]
			rx, ry, rz := wx+int32(c.DX), wy+int32(c.DY), wz+int32(c.DZ)
			if c.RotEnd < 12 {
				rx -= int32(RotationOffset[c.RotEnd][0])
				ry -= int32(RotationOffset[c.RotEnd][1])
			}
			target = &RouteTarget{StationID: -1, X: wx, Y: wy, Z: wz, TaD: tad, RX: rx, RY: ry, RZ: rz, ReverseTaD: tad ^ 4}
		}
	}
	if target != nil {
		best := 0
		for i, c := range conns {
			conn := c & raDWithSignalMask
			ct := conn & 0x7F
			if (x == target.X && y == target.Y && z == target.Z && ct == target.TaD) ||
				(x == target.RX && y == target.RY && z == target.RZ && ct == target.ReverseTaD) {
				st.Result = RoutingResult{SignalState: RouteNoSignals}
				st.HadNewResult = true
				return c
			}
			r := m.RoadTargetedPathing(x, y, z, conn, p, allowedStationTypes, target, om)
			if (!st.HadNewResult && !isSecondRun) || IsRoadRoutingResultBetter(st.Result, r) {
				st.Result = r
				st.HadNewResult = true
				best = i
			}
		}
		return conns[best]
	}
	var bestValue uint16
	best := 0
	for i, c := range conns {
		f := m.RoadAimlessWanderPathing(x, y, z, c&raDWithSignalMask, p, om)
		if v := wanderScore[f&7] + uint16(randVal&7); v >= bestValue {
			bestValue, best = v, i
		}
		randVal = randVal>>3 | randVal<<29
	}
	return conns[best]
}
