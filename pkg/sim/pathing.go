package sim

// Track pathing at junctions without a target — a port of VehicleHead.cpp's
// trackLongestPathing (0x004AC34D) with trackLongestPathingCalculate
// (0x004ACBFF) and trackAimlessWanderPathing (0x004AC884), each a recursion
// (at most 5 deep, at most 1280 weight) over getTrackConnections that stops
// at occupied or no-route signals.
//
// PARTIAL: targeted pathing (trackTargetedPathing / sub_4AC94F and friends)
// and the road equivalents are not ported yet.

const pathingMaxWeight = 1280

// Pathing is the vehicle identity the walks query connections for.
type Pathing struct {
	Company, TrackType, RequiredMods, QueryMods uint8
}

type longestState struct {
	depth       uint16
	total, best uint32
}

func (m *Map) trackLongestRecurse(x, y, z int32, tad uint16, p Pathing, st *longestState) {
	if st.depth >= 5 {
		return
	}
	for {
		if tad&TaDHasSignal != 0 {
			basic := tad &^ TaDHasSignal
			if m.SignalState(x, y, z, basic, p.TrackType, 0)&SignalOccupied != 0 {
				return
			}
			if m.SignalState(x, y, z, basic, p.TrackType, signalRightSide)&SignalBlockedNoRoute != 0 {
				return
			}
		}
		st.total += TrackUnkWeighting[(tad>>3)&0x3F%44]
		if st.total > pathingMaxWeight {
			return
		}
		st.best = max(st.best, st.total)
		e, ok := TrackConnectionEnd(x, y, z, tad&TaDBasicMask)
		if !ok {
			return
		}
		tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, p.Company, p.TrackType, p.RequiredMods, p.QueryMods)
		if len(tc.Connections) == 0 {
			return
		}
		x, y, z = e.X, e.Y, e.Z
		tad = tc.Connections[0] & (TaDBasicMask | TaDHasSignal)
		if len(tc.Connections) == 1 {
			continue
		}
		for _, c := range tc.Connections {
			rs := *st
			rs.depth++
			m.trackLongestRecurse(x, y, z, c&(TaDBasicMask|TaDHasSignal), p, &rs)
			st.best = rs.best
		}
		return
	}
}

// TrackLongestPathingCalculate is trackLongestPathingCalculate: the best
// weight reachable from (pos, tad).
func (m *Map) TrackLongestPathingCalculate(x, y, z int32, tad uint16, p Pathing) uint16 {
	var st longestState
	m.trackLongestRecurse(x, y, z, tad, p, &st)
	return uint16(st.best)
}

// TrackLongestPathing is trackLongestPathing: of a junction's connections,
// the one leading furthest (the last of equals).
func (m *Map) TrackLongestPathing(x, y, z int32, conns []uint16, p Pathing) uint16 {
	var best uint16
	bestI := 0
	for i, c := range conns {
		if v := m.TrackLongestPathingCalculate(x, y, z, c&(TaDBasicMask|TaDHasSignal), p); v >= best {
			best, bestI = v, i
		}
	}
	return conns[bestI]
}

// Aimless wander flags (Sub4AC884State unkFlags).
const (
	WanderBlockedNoRoute uint16 = 1 << 0
	WanderPassedSignal   uint16 = 1 << 1
	WanderOccupied       uint16 = 1 << 2
)

type wanderState struct {
	depth uint16
	flags uint16
	total uint32
}

func (m *Map) trackWanderRecurse(x, y, z int32, tad uint16, p Pathing, st *wanderState) {
	if st.depth >= 5 {
		return
	}
	for {
		if tad&TaDHasSignal != 0 {
			s := m.SignalState(x, y, z, tad&TaDBasicMask, p.TrackType, 0)
			if s&SignalBlockedNoRoute != 0 {
				st.flags |= WanderBlockedNoRoute
				return
			} else if s&SignalOccupied != 0 {
				st.flags |= WanderOccupied
				return
			}
			st.flags |= WanderPassedSignal
		}
		st.total += TrackUnkWeighting[(tad>>3)&0x3F%44]
		if st.total > pathingMaxWeight {
			return
		}
		e, ok := TrackConnectionEnd(x, y, z, tad&TaDBasicMask)
		if !ok {
			return
		}
		tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, p.Company, p.TrackType, p.RequiredMods, p.QueryMods)
		if len(tc.Connections) == 0 {
			return
		}
		x, y, z = e.X, e.Y, e.Z
		tad = tc.Connections[0] & (TaDBasicMask | TaDHasSignal)
		if len(tc.Connections) == 1 {
			continue
		}
		for _, c := range tc.Connections {
			rs := *st
			rs.depth++
			m.trackWanderRecurse(x, y, z, c&(TaDBasicMask|TaDHasSignal), p, &rs)
			st.flags |= rs.flags
		}
		return
	}
}

// TrackAimlessWanderPathing is trackAimlessWanderPathing: what lies ahead of
// (pos, tad) within reach — a no-route signal, a passable one, an occupied one.
func (m *Map) TrackAimlessWanderPathing(x, y, z int32, tad uint16, p Pathing) uint16 {
	var st wanderState
	m.trackWanderRecurse(x, y, z, tad, p, &st)
	return st.flags
}
