package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// Road pathing — VehicleHead.cpp's roadAimlessWanderPathing (0x0047E481),
// roadLongestPathing (0x0047DF4A) with roadLongestPathingCalculate
// (0x0047E72F), and roadTargetedPathing with roadUpdateDistanceToTarget
// (0x0047E582). Like the track walks: ≤ 5 deep, ≤ 1280 weight.
//
// PARTIAL: roadPathing (0x0047DFD0, the junction choice) and
// isRoadRoutingResultBetter are not ported yet.

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
