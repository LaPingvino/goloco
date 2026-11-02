package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Helpers of the rail reservation step (sub_4ACEF1) — Tile::trainStation,
// VehicleHead.cpp's bringTrackElementToFront and updateJunctionSignalLights,
// and Routing.cpp's sub_4A2A77 and setReverseSignalOccupiedInBlock.

// trainStation is Tile::trainStation: the station element after the track
// element with this track id, rotation and height that has a station.
func (t *Tile) trainStation(trackID, rotation, baseZ uint8) *scenario.StationElement {
	found := false
	for _, r := range t.Order {
		if found && r.Kind == scenario.KindStation {
			return &t.Stations[r.Index]
		}
		if r.Kind != scenario.KindTrack {
			continue
		}
		el := &t.Tracks[r.Index]
		found = false
		if el.BaseZ != baseZ || el.Rotation != rotation || el.TrackID != trackID || !el.HasStation {
			continue
		}
		found = true
	}
	return nil
}

// trackStartFromTad is getTrackStartFromTad.
func trackStartFromTad(x, y, z int32, tad uint16) (int32, int32, int32) {
	if tad&4 != 0 {
		return trackEndTile(x, y, z, tad)
	}
	return x, y, z
}

// BringTrackElementToFront is bringTrackElementToFront: on each tile of the
// chosen piece, its track element moves to the front of the run of track
// elements at that height (the order junction pieces are drawn and found
// in), unless it carries a signal or station.
func (m *Map) BringTrackElementToFront(x, y, z int32, trackType uint8, tad uint16) {
	sx, sy, sz := trackStartFromTad(x, y, z, tad)
	id, rot := uint8(tad>>3&0x3F), uint8(tad&3)
	if int(id) >= len(TrackPieces) {
		return
	}
	for _, p := range TrackPieces[id] {
		rx, ry := rotOffset(p.X, p.Y, rot)
		px, py, pz := sx+int32(rx), sy+int32(ry), sz+int32(p.Z)
		t := m.TileAt(int(px>>5), int(py>>5))
		if t == nil {
			continue
		}
		begin, last := -1, -1
		for oi, r := range t.Order {
			if r.Kind != scenario.KindTrack {
				begin = -1
				continue
			}
			el := &t.Tracks[r.Index]
			if int32(el.BaseZ)*4 != pz {
				begin = -1
				continue
			}
			if begin < 0 {
				begin = oi
			}
			if el.Rotation != rot || el.SequenceIndex != p.Index || el.TrackObjectID != trackType || el.TrackID != id {
				continue
			}
			if !el.HasSignal && !el.HasStation {
				last = oi
			}
			break
		}
		if last < 0 || begin < 0 || begin == last {
			continue
		}
		ref := t.Order[last]
		copy(t.Order[begin+1:last+1], t.Order[begin:last])
		t.Order[begin] = ref
	}
}

// UpdateJunctionSignalLights is updateJunctionSignalLights: the signal before
// a junction shows which way the route goes (left / right of the straightest
// connection), set through setSignalState's lights mode (16).
func (m *Map) UpdateJunctionSignalLights(x, y, z int32, targetRouting uint16, trackType uint8, newRouting uint16, conns []uint16) {
	if len(conns) == 0 {
		return
	}
	curv := func(r uint16) int8 { return CurvatureDegrees[((r&TaDBasicMask)>>2)%88] }
	minC, maxC, least := curv(conns[0]), curv(conns[0]), curv(conns[0])
	abs8 := func(v int8) int {
		if v < 0 {
			return -int(v)
		}
		return int(v)
	}
	for _, c := range conns[1:] {
		v := curv(c)
		minC, maxC = min(minC, v), max(maxC, v)
		if abs8(v) < abs8(least) {
			least = v
		}
	}
	nc := curv(newRouting)
	flags := uint32(0x10)
	if least != minC {
		flags |= 1 << 30
		if nc < least {
			flags |= 1 << 28
		}
	}
	if least != maxC {
		flags |= 1 << 29
		if nc > least {
			flags |= 1 << 27
		}
	}
	m.SetSignalState(x, y, z, targetRouting&TaDBasicMask, trackType, flags)
}

// Sub4A2A77 is sub_4A2A77: searching forwards only from (pos, tad) to the
// next signals — bit 0: one of them is clear or occupied one-way; bit 1: the
// line dead-ends.
func (m *Map) Sub4A2A77(x, y, z int32, tad uint16, company, trackType uint8) uint8 {
	var data uint16
	res := &RoutingResults{Reachable: NewInterestSet(SignalHashSetSize)}
	m.FindAllTracksFilterTransform(res, SearchExcludeReverseDirection, x, y, z, tad, company, trackType,
		func(in LocationOfInterest) bool {
			if in.TrackAndDirection&TaDHasSignal == 0 {
				return false
			}
			s := m.SignalState(in.X, in.Y, in.Z, in.TaD(), in.TrackType, 0)
			clear := s&SignalBlockedNoRoute == 0 && s&SignalOccupied == 0
			if s&SignalOccupiedOneWay != 0 || clear {
				data |= 1
			}
			return true
		}, nil)
	if res.HasDeadEnd {
		data |= 2
	}
	return uint8(data)
}

// SetReverseSignalOccupiedInBlock is setReverseSignalOccupiedInBlock: every
// signal bounding the block (searched with unk2) set occupied on its right
// side (1<<31 | 8).
func (m *Map) SetReverseSignalOccupiedInBlock(x, y, z int32, tad uint16, company, trackType uint8) {
	res := &RoutingResults{Reachable: NewInterestSet(SignalHashSetSize)}
	m.FindAllTracksFilterTransform(res, SearchUnk2, x, y, z, tad, company, trackType,
		func(in LocationOfInterest) bool {
			if in.TrackAndDirection&TaDHasSignal == 0 {
				return false
			}
			m.SetSignalState(in.X, in.Y, in.Z, in.TaD(), in.TrackType, signalRightSide|8)
			return true
		}, nil)
}

// Sub4ACEE7Result is Sub4ACEE7Result: Al 0 moved on, 1 reservation full,
// 2 no route / blocked, 3 waiting at a signal (Flags: its reverse state),
// 4 arrived at a station to stop at (StationID).
type Sub4ACEE7Result struct {
	Al        uint8
	Flags     uint8
	StationID int // -1: null
}

const (
	flags48PassSignal  uint8 = 1 << 0
	flags48ExpressMode uint8 = 1 << 1
)

// Sub4ACEF1 is sub_4ACEF1 (rail): with free room in its reservation the head
// moves on (by unk1 − var_3C, at most max(var_113612C·4, 0xCC48)); at the
// end of what it has it handles the station it is on (stop / route through /
// non-express stop), chooses the next piece (trackPathing, or the longest
// line when var_52 is 1), sets junction lights, enters level crossings,
// obeys signals (no route, occupied waits, one-way checks), reserves that
// piece and advances a waypoint order it reaches. Upstream's module static
// _vehicleMangled_113623B is Map.vehicleMangled113623B.
func (m *Map) Sub4ACEF1(tr Train, unk1 uint32, var113612C uint32, isPlaceDown bool) Sub4ACEE7Result {
	head, v1 := &m.Entities[tr.Head], &m.Entities[tr.Veh1]
	null := Sub4ACEE7Result{StationID: -1}
	// begin()+2 and begin()+3 (raw index steps, as upstream's iterator ++
	// does without checking end()) must be free.
	hh := RoutingHandle(head.RoutingHandle)
	if m.Routing(hh.WithIndex(hh.Index()+2)) != RoutingAllocatedButFree || m.Routing(hh.WithIndex(hh.Index()+3)) != RoutingAllocatedButFree {
		return Sub4ACEE7Result{Al: 1, StationID: -1}
	}
	c := m.component(tr.Head)
	var res MotionResult
	if head.HeadVar52 == 1 {
		res = c.UpdateMotion(0, func(c *MotionComponent, f *uint32) bool { return m.TrackNewPiece(c, f, head.HeadVar53, false, 0) })
		c.RemainingDistance += res.RemainingDistance
	} else {
		d := min(int32(unk1)-head.Var3C, int32(max(var113612C*4, 0xCC48)))
		res = c.UpdateMotion(d, func(c *MotionComponent, f *uint32) bool { return m.TrackNewPiece(c, f, head.HeadVar53, false, 0) })
		head.Var3C += d - res.RemainingDistance
	}
	m.store(tr.Head, &c)
	if !res.Blocked() {
		return null
	}
	x, y, z := int32(head.TileX), int32(head.TileY), int32(head.TileBaseZ)*4
	tad := head.TrackAndDirection & TaDBasicMask
	tileStation := -1
	if t := m.TileAt(int(x>>5), int(y>>5)); t != nil {
		if st := t.trainStation(uint8(tad>>3&0x3F), uint8(tad&3), head.TileBaseZ); st != nil && !st.Ghost && !st.AiAllocated {
			tileStation = int(st.StationID)
		}
	}
	requiredMods, queryMods := head.HeadVar53, v1.Var49
	e, _ := TrackConnectionEnd(x, y, z, tad)
	tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, head.Owner, head.TrackType, requiredMods, queryMods)
	if head.HeadVar52 != 1 && tileStation >= 0 && tileStation != tc.StationID {
		orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder)
		processed := false
		if len(orders) > 0 {
			o := orders[0]
			switch o.Type() {
			case OrderStopAt:
				if tileStation == o.Station() {
					processed = true
					if head.VehicleFlags&VehicleFlagManualControl == 0 {
						return Sub4ACEE7Result{Al: 4, StationID: tileStation}
					}
				}
			case OrderRouteThrough:
				if tileStation == o.Station() {
					next := orders[1%len(orders)]
					head.CurrentOrder = uint16(next.Offset - int(head.OrderTableOffset))
					processed = true
				}
			}
		}
		if !processed && int(head.StationID) != tileStation && v1.Var48&flags48ExpressMode == 0 {
			return Sub4ACEE7Result{Al: 4, StationID: tileStation}
		}
	}
	if len(tc.Connections) == 0 {
		return Sub4ACEE7Result{Al: 2, StationID: -1}
	}
	conn := tc.Connections[0]
	p := Pathing{Company: head.Owner, TrackType: head.TrackType, RequiredMods: requiredMods, QueryMods: queryMods}
	if len(tc.Connections) > 1 {
		if head.HeadVar52 == 1 {
			conn = m.TrackLongestPathing(e.X, e.Y, e.Z, tc.Connections, p)
		} else {
			var st PathingState
			conn = m.TrackPathing(head.OrderTableOffset, head.CurrentOrder, e.X, e.Y, e.Z, tc.Connections, p, false, &st)
		}
		conn |= 1 << 14
		m.BringTrackElementToFront(e.X, e.Y, e.Z, head.TrackType, conn&TaDBasicMask)
		// Back through the routings (up to 6) to the signal before the junction.
		rx, ry, rz := x, y, z
		h := RoutingHandle(head.RoutingHandle)
		for i := 0; i < 6; i, h = i+1, h.WithIndex(h.Index()-1) {
			r := m.Routing(h)
			if r == RoutingAllocatedButFree {
				break
			}
			if h != RoutingHandle(head.RoutingHandle) {
				cc := TrackCoordinates[r&TaDBasicMask]
				rx, ry, rz = rx-int32(cc.DX), ry-int32(cc.DY), rz-int32(cc.DZ)
			}
			if r&TaDHasSignal != 0 {
				m.UpdateJunctionSignalLights(rx, ry, rz, r, head.TrackType, conn, tc.Connections)
				break
			} else if r&(1<<14) != 0 {
				break
			}
		}
	} else {
		if tc.HasLevelCrossing {
			m.LeaveLevelCrossing(e.X, e.Y, e.Z, conn&TaDBasicMask, LevelCrossingEnter)
		}
		if conn&TaDHasSignal != 0 {
			ctad := conn & TaDBasicMask
			key := SignalOccupied | SignalOccupiedOneWay | SignalBlockedNoRoute
			if isPlaceDown {
				key = SignalOccupied
			}
			s := m.SignalState(e.X, e.Y, e.Z, ctad, head.TrackType, 0) & key
			switch {
			case s&SignalBlockedNoRoute != 0:
				return Sub4ACEE7Result{Al: 2, StationID: -1}
			case s&SignalOccupied != 0:
				rs := m.SignalState(e.X, e.Y, e.Z, ctad, head.TrackType, signalRightSide)
				m.vehicleMangled113623B = rs
				if head.Var5C == 0 {
					v1.Var48 &^= flags48PassSignal
					return Sub4ACEE7Result{Al: 3, Flags: rs, StationID: -1}
				}
			default:
				if v1.Var48&flags48PassSignal != 0 {
					v1.Var48 &^= flags48PassSignal
					if m.IsBlockOccupied(e.X, e.Y, e.Z, ctad, head.Owner, head.TrackType) {
						m.SetSignalState(e.X, e.Y, e.Z, ctad, head.TrackType, 8)
						return Sub4ACEE7Result{Al: 3, Flags: m.vehicleMangled113623B, StationID: -1}
					}
				}
				if s&SignalOccupiedOneWay != 0 {
					if v1.HeadVar52 != 0 { // Vehicle1 var_52
						m.vehicleMangled113623B |= 1 << 7
						v1.HeadVar52--
						return Sub4ACEE7Result{Al: 3, Flags: m.vehicleMangled113623B, StationID: -1}
					} else if m.Sub4A2A77(e.X, e.Y, e.Z, ctad, head.Owner, head.TrackType)&3 == 0 {
						m.vehicleMangled113623B |= 1 << 7
						v1.HeadVar52 = 55
						return Sub4ACEE7Result{Al: 3, Flags: m.vehicleMangled113623B, StationID: -1}
					}
				}
			}
			m.SetReverseSignalOccupiedInBlock(e.X, e.Y, e.Z, ctad, head.Owner, head.TrackType)
			if head.Var5C == 0 {
				m.SetSignalState(e.X, e.Y, e.Z, ctad, head.TrackType, 1)
			}
			edi := uint32(2)
			rx, ry, rz := x, y, z
			for h := RoutingHandle(head.RoutingHandle); m.Routing(h) != RoutingAllocatedButFree; h = h.WithIndex(h.Index() - 1) {
				r := m.Routing(h)
				if h != RoutingHandle(head.RoutingHandle) {
					cc := TrackCoordinates[r&TaDBasicMask]
					rx, ry, rz = rx-int32(cc.DX), ry-int32(cc.DY), rz-int32(cc.DZ)
				}
				if r&TaDHasSignal != 0 {
					m.SetSignalState(rx, ry, rz, r&TaDBasicMask, head.TrackType, edi)
					edi = min(edi+1, 3)
				}
				if h.WithIndex(h.Index()-1) == RoutingHandle(head.RoutingHandle) {
					break // a full lap of the ring
				}
			}
		}
	}
	m.SetRouting(RoutingHandle(head.RoutingHandle).WithIndex(RoutingHandle(head.RoutingHandle).Index()+1), conn)
	if head.HeadVar52 == 1 {
		return null
	}
	orders := m.OrderRing(head.OrderTableOffset, head.CurrentOrder)
	if len(orders) == 0 || orders[0].Type() != OrderRouteWaypoint {
		return null
	}
	wo := orders[0]
	wx, wy, wz := wo.Waypoint()
	wtad := uint16(wo.TrackID())<<3 | uint16(wo.Direction())
	if e.X != wx || e.Y != wy || e.Z != wz || conn&TaDBasicMask != wtad {
		cx, cy, cz := trackEndTile(e.X, e.Y, e.Z, conn&TaDBasicMask)
		cc := TrackCoordinates[conn&TaDBasicMask]
		rt := (conn&TaDBasicMask^4)&^3 | uint16(ReverseRotation[cc.RotEnd]&3)
		if cx != wx || cy != wy || cz != wz || rt != wtad {
			return null
		}
	}
	next := orders[1%len(orders)]
	head.CurrentOrder = uint16(next.Offset - int(head.OrderTableOffset))
	return null
}
