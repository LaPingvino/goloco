package sim

// Track pathing at junctions without a target — a port of VehicleHead.cpp's
// trackLongestPathing (0x004AC34D) with trackLongestPathingCalculate
// (0x004ACBFF) and trackAimlessWanderPathing (0x004AC884), each a recursion
// (at most 5 deep, at most 1280 weight) over getTrackConnections that stops
// at occupied or no-route signals.
//
// PARTIAL: trackPathing (0x004AC3D3: the per-junction choice over these for
// the head's orders) and the road equivalents are not ported yet.

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

// Targeted track pathing — trackTargetedPathing (0x004AC94F) with
// trackUpdateDistanceToTarget (0x004AC98B), processReachedTargetRouteEnd
// (0x004AC9FD) and isTrackRoutingResultBetter (0x004AC6DA).

// RouteSignalState.
const (
	RouteNoSignals           uint32 = 1
	RouteSignalClear         uint32 = 2
	RouteSignalBlockedOneWay uint32 = 3
	RouteSignalBlockedTwoWay uint32 = 4
	RouteSignalNoRoute       uint32 = 6
	RouteNull                uint32 = 0xFFFFFFFF
)

// RoutingResult is RoutingResult.
type RoutingResult struct {
	BestDistToTarget   uint16
	BestTrackWeighting uint32
	SignalState        uint32
}

// RouteTarget is Sub4AC94FTarget: a station (StationID ≥ 0), or a piece
// (pos, tad) reachable either way (reversePos, reverseTad).
type RouteTarget struct {
	StationID  int // -1: none
	X, Y, Z    int32
	TaD        uint16
	RX, RY, RZ int32
	ReverseTaD uint16
}

type targetedState struct {
	depth  uint16
	total  uint32
	result RoutingResult
}

func absDiff(a, b int32) uint32 {
	if a > b {
		return uint32(a - b)
	}
	return uint32(b - a)
}

// trackUpdateDistanceToTarget: dist = min(dx,dy)/4 + max(dx,dy) + dz.
func trackUpdateDistanceToTarget(x, y, z int32, t *RouteTarget, st *targetedState) {
	dx, dy, dz := absDiff(int32(int16(x)), int32(int16(t.X))), absDiff(int32(int16(y)), int32(int16(t.Y))), absDiff(int32(int16(z)), int32(int16(t.Z)))
	dist := min(dx, dy)/4 + max(dx, dy) + dz
	if dist <= uint32(st.result.BestDistToTarget) {
		if dist < uint32(st.result.BestDistToTarget) || st.total <= st.result.BestTrackWeighting {
			st.result.BestDistToTarget = uint16(dist)
			st.result.BestTrackWeighting = st.total
		}
	}
}

// processReachedTargetRouteEnd: true when this is the best route so far.
func processReachedTargetRouteEnd(st *targetedState) bool {
	if st.result.BestDistToTarget != 0 || st.total <= st.result.BestTrackWeighting {
		st.result.BestDistToTarget = 0
		st.result.BestTrackWeighting = st.total
		if st.result.SignalState == RouteNull {
			st.result.SignalState = RouteNoSignals
		}
		return true
	}
	return false
}

func (m *Map) trackTargetedRecurse(x, y, z int32, tad uint16, p Pathing, t *RouteTarget, st *targetedState) {
	curStation := -1
	if st.depth >= 5 {
		return
	}
	for {
		reached := false
		if t.StationID >= 0 {
			reached = curStation == t.StationID
		} else if x == t.X && y == t.Y && z == t.Z && tad&TaDBasicMask == t.TaD {
			reached = true
		} else if x == t.RX && y == t.RY && z == t.RZ && tad&TaDBasicMask == t.ReverseTaD {
			reached = true
		}
		if reached {
			if processReachedTargetRouteEnd(st) {
				return
			}
		} else {
			trackUpdateDistanceToTarget(x, y, z, t, st)
		}
		if tad&TaDHasSignal != 0 {
			s := m.SignalState(x, y, z, tad&^TaDHasSignal, p.TrackType, 0)
			switch {
			case s&SignalBlockedNoRoute != 0:
				if st.result.SignalState == RouteNull {
					st.result.SignalState = RouteSignalNoRoute
				}
				return
			case s&SignalOccupied != 0:
				if st.result.SignalState == RouteNull {
					st.result.SignalState = RouteSignalBlockedTwoWay
					if s&SignalOccupiedOneWay != 0 {
						st.result.SignalState = RouteSignalBlockedOneWay
					}
				}
			default:
				if st.result.SignalState == RouteNull {
					st.result.SignalState = RouteSignalClear
				}
			}
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
		curStation = tc.StationID
		if len(tc.Connections) == 1 {
			continue
		}
		sig := st.result.SignalState
		for _, c := range tc.Connections {
			rs := *st
			rs.depth++
			m.trackTargetedRecurse(x, y, z, c&(TaDBasicMask|TaDHasSignal), p, t, &rs)
			sig = min(sig, rs.result.SignalState)
			st.result.BestDistToTarget = rs.result.BestDistToTarget
			st.result.BestTrackWeighting = rs.result.BestTrackWeighting
		}
		st.result.SignalState = sig
		return
	}
}

// TrackTargetedPathing is trackTargetedPathing (0x004AC94F).
func (m *Map) TrackTargetedPathing(x, y, z int32, tad uint16, p Pathing, t *RouteTarget) RoutingResult {
	st := targetedState{result: RoutingResult{BestDistToTarget: 0xFFFF, BestTrackWeighting: 0xFFFFFFFF, SignalState: RouteNull}}
	m.trackTargetedRecurse(x, y, z, tad, p, t, &st)
	return st.result
}

// IsTrackRoutingResultBetter is isTrackRoutingResultBetter (0x004AC6DA).
func IsTrackRoutingResultBetter(base, n RoutingResult) bool {
	if base.SignalState == RouteNull {
		return true
	}
	newReached, baseReached := n.BestDistToTarget == 0, base.BestDistToTarget == 0
	if newReached && n.SignalState == RouteSignalBlockedOneWay {
		if base.SignalState <= RouteSignalClear && base.BestTrackWeighting > 288 && n.BestTrackWeighting*5/4 <= base.BestTrackWeighting {
			return true
		}
	}
	if baseReached && base.SignalState == RouteSignalBlockedOneWay {
		if n.SignalState <= RouteSignalClear && n.BestTrackWeighting > 288 && base.BestTrackWeighting*5/4 <= n.BestTrackWeighting {
			return false
		}
	}
	if !newReached && n.SignalState == RouteSignalBlockedOneWay {
		if !baseReached && base.SignalState == RouteSignalClear {
			if int(n.BestDistToTarget)*5/4 <= int(base.BestDistToTarget) || int(n.BestDistToTarget)+320 <= int(base.BestDistToTarget) {
				return true
			}
		}
	}
	if !baseReached && base.SignalState == RouteSignalBlockedOneWay {
		if !newReached && n.SignalState == RouteSignalClear {
			if int(base.BestDistToTarget)*5/4 <= int(n.BestDistToTarget) || int(base.BestDistToTarget)+320 <= int(n.BestDistToTarget) {
				return false
			}
		}
	}
	if newReached && !baseReached {
		return true
	}
	if !newReached && baseReached {
		return false
	}
	if n.SignalState < base.SignalState {
		return true
	}
	if n.SignalState > base.SignalState {
		return false
	}
	if n.BestDistToTarget < base.BestDistToTarget {
		return true
	}
	return n.BestDistToTarget == base.BestDistToTarget && n.BestTrackWeighting <= base.BestTrackWeighting
}
