package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// Whether to reverse — VehicleHead.cpp's pathingShouldReverse (0x004ACCDC)
// with trackPathingShouldReverse (0x004ACCE6) / roadPathingShouldReverse
// (0x004ACDE0), sub_4AC1C2 and Routing.cpp's sub_4A2A58: decisions only; the
// reversal itself (sub_4ADB47) is item 5.6d.

// pathingFrom runs the pathing step a train would take from the end of
// (x, y, z, tad), as the head (its orders) — false when nothing connects.
func (m *Map) trackPathingFrom(head int, x, y, z int32, tad uint16, company, trackType uint8, p Pathing, second bool, st *PathingState) bool {
	e, ok := TrackConnectionEnd(x, y, z, tad)
	if !ok {
		return false
	}
	tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, company, trackType, p.RequiredMods, p.QueryMods)
	if len(tc.Connections) == 0 {
		return false
	}
	h := &m.Entities[head]
	m.TrackPathing(h.OrderTableOffset, h.CurrentOrder, e.X, e.Y, e.Z, tc.Connections, p, second, st)
	return true
}

// PathingShouldReverse is pathingShouldReverse: the pathing from the train's
// far end (the tail, reversed) gives a better route than from the head.
func (m *Map) PathingShouldReverse(tr Train, om *objects.ObjectManager) bool {
	head, v1, tail := &m.Entities[tr.Head], &m.Entities[tr.Veh1], &m.Entities[tr.Tail]
	p := Pathing{Company: head.Owner, TrackType: head.TrackType, RequiredMods: head.HeadVar53, QueryMods: v1.Var49}
	var st PathingState
	if head.Mode == 0 {
		if !m.trackPathingFrom(tr.Head, int32(head.TileX), int32(head.TileY), int32(head.TileBaseZ)*4, head.TrackAndDirection&TaDBasicMask, head.Owner, head.TrackType, p, false, &st) {
			return false
		}
		tad := tail.TrackAndDirection & TaDBasicMask
		x, y, z := trackEndTile(int32(tail.TileX), int32(tail.TileY), int32(tail.TileBaseZ)*4, tad)
		if !m.trackPathingFrom(tr.Head, x, y, z, tad^4, tail.Owner, tail.TrackType, p, true, &st) {
			return false
		}
		return st.HadNewResult
	}
	if head.TrackType != 0xFF {
		if rd := om.GetRoadObjectByIndex(int(head.TrackType)); rd == nil || rd.Flags&roadFlagIsRoad == 0 {
			return false
		}
	}
	allowed := CalculateCompatibleRoadStations(head.TrainAcceptedCargoTypes, om)
	roadFrom := func(x, y, z int32, rad uint16, company, roadType uint8, second bool) bool {
		e, ok := RoadConnectionEnd(x, y, z, rad)
		if !ok {
			return false
		}
		rc := m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, company, roadType, p.RequiredMods, p.QueryMods)
		if len(rc.Connections) == 0 {
			return false
		}
		m.RoadPathing(head.OrderTableOffset, head.CurrentOrder, e.X, e.Y, e.Z, rc.Connections, p, allowed, second, &st, om)
		return true
	}
	if !roadFrom(int32(head.TileX), int32(head.TileY), int32(head.TileBaseZ)*4, head.TrackAndDirection&0x7F, head.Owner, head.TrackType, false) {
		return false
	}
	rad := tail.TrackAndDirection & 0x7F
	c := RoadCoordinates[rad]
	// (upstream's road version adds getUnkRoad's offset without the
	// rotation-offset correction the track version applies)
	x, y, z := int32(tail.TileX)+int32(c.DX), int32(tail.TileY)+int32(c.DY), int32(tail.TileBaseZ)*4+int32(c.DZ)
	if !roadFrom(x, y, z, rad^4, tail.Owner, tail.TrackType, true) {
		return false
	}
	return st.HadNewResult
}

// Sub4A2A58 is sub_4A2A58: the signals bounding the block ahead (searched
// with unk2) — bit 0: one is occupied; bit 1: one is not.
func (m *Map) Sub4A2A58(x, y, z int32, tad uint16, company, trackType uint8) uint8 {
	var unk uint16
	res := &RoutingResults{Reachable: NewInterestSet(SignalHashSetSize)}
	m.FindAllTracksFilterTransform(res, SearchUnk2, x, y, z, tad, company, trackType,
		func(in LocationOfInterest) bool {
			if in.TrackAndDirection&TaDHasSignal == 0 {
				return false
			}
			if m.SignalState(in.X, in.Y, in.Z, in.TaD(), in.TrackType, signalRightSide)&SignalOccupied != 0 {
				unk |= 1
			} else {
				unk |= 2
			}
			return true
		}, nil)
	return uint8(unk)
}

// Sub4AC1C2 is sub_4AC1C2: a single connection ahead whose block has a
// free signal (sub_4A2A58 bit 1) — the head may pass the signal it waits at.
func (m *Map) Sub4AC1C2(tr Train) bool {
	head := &m.Entities[tr.Head]
	e, ok := TrackConnectionEnd(int32(head.TileX), int32(head.TileY), int32(head.TileBaseZ)*4, head.TrackAndDirection&TaDBasicMask)
	if !ok {
		return false
	}
	tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, head.Owner, head.TrackType, 0, 0)
	if len(tc.Connections) != 1 {
		return false
	}
	return m.Sub4A2A58(e.X, e.Y, e.Z, tc.Connections[0]&TaDBasicMask, head.Owner, head.TrackType)&2 != 0
}
