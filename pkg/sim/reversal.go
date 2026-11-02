package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// Reversing — VehicleHead.cpp's sub_4AD778, roadResetHead,
// isOnExpectedRoadOrTrack, sub_4ADB47 (the reversal) and tryReverse.
//
// PARTIAL (in sub_4ADB47): liftUpVehicle when the reversed train can't be
// placed (the commandStop flag is set as upstream; the lift-up is not
// ported), connectJacobsBogies (Jacobs-bogie vehicles only), and destroyTrain
// in tryReverse off the expected track.

const flags38Unk2 uint8 = 1 << 2

func (m *Map) copyMotion(dst, src int) {
	d, s := &m.Entities[dst], &m.Entities[src]
	d.TileX, d.TileY, d.TileBaseZ = s.TileX, s.TileY, s.TileBaseZ
	d.RoutingHandle, d.RemainingDistance = s.RoutingHandle, s.RemainingDistance
	d.TrackAndDirection, d.SubPosition = s.TrackAndDirection, s.SubPosition
	d.PosX, d.PosY, d.PosZ = s.PosX, s.PosY, s.PosZ
}

// roadResetHead: the head back onto Vehicle1, its routings beyond freed.
func (m *Map) roadResetHead(tr Train) {
	head, v1 := &m.Entities[tr.Head], &m.Entities[tr.Veh1]
	head.Flags38 |= flags38Unk2
	for _, h := range m.RoutingRing(RoutingHandle(v1.RoutingHandle), false) {
		if h != RoutingHandle(v1.RoutingHandle) {
			m.SetRouting(h, RoutingAllocatedButFree)
		}
	}
	m.copyMotion(tr.Head, tr.Veh1)
	head.Var3C = v1.Var3C
	head.Flags38 &^= flags38Unk2
}

// Sub4AD778 is sub_4AD778: everything reserved beyond Vehicle2 is released
// (road lanes; rail signals, blocks and crossings) and the head and Vehicle1
// are brought back onto Vehicle2.
func (m *Map) Sub4AD778(tr Train, om *objects.ObjectManager) {
	head, v1, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh1], &m.Entities[tr.Veh2]
	v2h := RoutingHandle(v2.RoutingHandle)
	x, y, z := int32(v2.TileX), int32(v2.TileY), int32(v2.TileBaseZ)*4
	if head.Mode == 1 {
		m.roadResetHead(tr)
		head.Flags38 |= flags38Unk2
		v1.Flags38 |= flags38Unk2
		for _, h := range m.RoutingRing(v2h, false) {
			tad := m.Routing(h) & TaDBasicMask
			if h != v2h {
				m.UpdateRoadTileOccupancy(x, y, z, tad, v2.TrackType, false)
			}
			c := RoadCoordinates[tad&0x7F]
			x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
			if h != v2h {
				m.SetRouting(h, RoutingAllocatedButFree)
			}
		}
	} else {
		head.Flags38 |= flags38Unk2
		v1.Flags38 |= flags38Unk2
		for _, h := range m.RoutingRing(v2h, false) {
			r := m.Routing(h)
			tad := r & TaDBasicMask
			m.UpdateBlockSignals(x, y, z, tad|r&TaDHasSignal, v2.Owner, v2.TrackType)
			if h != v2h {
				if r&TaDHasSignal != 0 {
					m.SetSignalState(x, y, z, tad, v2.TrackType, 0)
				}
				m.LeaveLevelCrossing(x, y, z, tad, LevelCrossingLeave)
			}
			c := TrackCoordinates[tad]
			x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
			if h != v2h {
				m.SetRouting(h, RoutingAllocatedButFree)
			}
		}
	}
	m.copyMotion(tr.Head, tr.Veh2)
	head.Var3C = 0
	head.Flags38 &^= flags38Unk2
	m.copyMotion(tr.Veh1, tr.Veh2)
	v1.Var3C = 0
	v1.Flags38 &^= flags38Unk2
}

// IsOnExpectedRoadOrTrack is isOnExpectedRoadOrTrack: Vehicle2's piece is
// still there (within 4 SmallZ, not ghost / AI).
func (m *Map) IsOnExpectedRoadOrTrack(tr Train) bool {
	v2 := &m.Entities[tr.Veh2]
	t := m.TileAt(int(v2.TileX)>>5, int(v2.TileY)>>5)
	if t == nil {
		return false
	}
	within := func(z uint8) bool { d := int(z) - int(v2.TileBaseZ); return d >= -4 && d <= 4 }
	if v2.Mode == 1 {
		for _, el := range t.Roads {
			if within(el.BaseZ) && !el.Ghost && !el.AiAllocated && uint16(el.RoadID) == v2.TrackAndDirection>>3&0xF {
				return true
			}
		}
		return false
	}
	for _, el := range t.Tracks {
		if within(el.BaseZ) && !el.Ghost && !el.AiAllocated && uint16(el.Rotation) == v2.TrackAndDirection&3 && uint16(el.TrackID) == v2.TrackAndDirection>>3&0x3F {
			return true
		}
	}
	return false
}

// Vehicle object flags the car reversal reads.
const (
	vehicleFlagTopAndTail = 1 << 1
	vehicleFlagCenter     = 1 << 5
	vehicleFlag08         = 1 << 8
	vehicleFlagShuntCheat = 1 << 7 // VehicleFlags (on the head)
)

// Sub4ADB47 is sub_4ADB47: the train turns round where it stands.
func (m *Map) Sub4ADB47(tr Train, unk bool, om *objects.ObjectManager) {
	m.Sub4AD778(tr, om)
	head := &m.Entities[tr.Head]
	if head.Status == StatusApproaching {
		head.Status = StatusTravelling
	}
	v2, tail := &m.Entities[tr.Veh2], &m.Entities[tr.Tail]
	x, y, z := int32(tail.TileX), int32(tail.TileY), int32(tail.TileBaseZ)*4
	road := head.Mode == 1
	for _, h := range m.RoutingRing(RoutingHandle(tail.RoutingHandle), false) {
		r := m.Routing(h)
		tad := r & TaDBasicMask
		if road {
			m.UpdateRoadTileOccupancy(x, y, z, tad, head.TrackType, false)
			c := RoadCoordinates[tad&0x7F]
			x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
		} else {
			if r&TaDHasSignal != 0 {
				m.SetSignalState(x, y, z, tad, tail.TrackType, 0)
			}
			c := TrackCoordinates[tad]
			x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
		}
	}
	// The routings under the train, from Vehicle2 back to the tail, reversed
	// into the start of the table.
	var copied []uint16
	it, end := RoutingHandle(v2.RoutingHandle), RoutingHandle(tail.RoutingHandle)
	for it.Index() != end.Index() {
		copied = append(copied, m.Routing(it))
		it = it.WithIndex(it.Index() - 1)
	}
	copied = append(copied, m.Routing(it))
	it = RoutingHandle(v2.RoutingHandle).WithIndex(0)
	for _, r := range copied {
		r ^= 1 << 2
		if road {
			r ^= RaDIsOvertaking
			if r&RaDIsChangingLane != 0 {
				r ^= RaDIsOvertaking
				if r&RaDIsOvertaking == 0 {
					r ^= RaDIsChangingLane
				}
			}
		}
		r &^= 1 << 14
		m.SetRouting(it, r)
		it = it.WithIndex(it.Index() + 1)
	}
	for it.Index() != 0 {
		m.FreeRouting(it)
		it = it.WithIndex(it.Index() + 1)
	}
	if !unk {
		lastBody := -1
		for _, car := range tr.Cars {
			for _, cc := range car {
				lastBody = cc.Body
			}
		}
		reverseCars := false
		if lastBody >= 0 {
			if obj := vehicleObject(om, m.Entities[lastBody].ObjectID); obj != nil {
				switch {
				case head.VehicleFlags&vehicleFlagShuntCheat != 0 && head.VehicleFlags&VehicleFlagManualControl != 0:
					reverseCars = true
				case obj.Flags&vehicleFlag08 != 0:
				case obj.Flags&vehicleFlagTopAndTail != 0:
					reverseCars = true
				case obj.Power == 0:
				default:
					reverseCars = obj.Flags&vehicleFlagCenter == 0
				}
			}
		}
		if reverseCars {
			dest := tr.Tail
			idx := m.entityIndex()
			for steps := 0; steps < 1024; steps++ {
				ci, ok := idx[m.Entities[tr.Veh2].NextCarID]
				if !ok || ci == dest || m.Entities[ci].EntityType != 3 {
					break
				}
				cur, err := m.TrainOf(head.ID)
				if err != nil {
					break
				}
				nf := m.FlipCar(cur, ci)
				cur, _ = m.TrainOf(head.ID)
				m.InsertCarBefore(cur, nf, tr.Head, dest)
				dest = nf
			}
			if t2, err := m.TrainOf(head.ID); err == nil {
				tr = t2
			}
		}
	}
	m.ApplyVehicleObjectLength(tr, om)
	v2 = &m.Entities[tr.Veh2]
	newTad := v2.TrackAndDirection
	nx, ny, nz := int32(v2.TileX), int32(v2.TileY), int32(v2.TileBaseZ)*4
	var newSub uint16
	if road {
		newSub = uint16(len(RoadSubPosition(newTad, m.TrafficHandedness != 0))) - 1 - v2.SubPosition
		c := RoadCoordinates[newTad&0x7F]
		nx, ny, nz = nx+int32(c.DX), ny+int32(c.DY), nz+int32(c.DZ)
		if c.RotEnd < 12 {
			nx -= int32(RotationOffset[c.RotEnd][0])
			ny -= int32(RotationOffset[c.RotEnd][1])
		}
		newTad ^= 4
		newTad ^= RaDIsOvertaking
		if newTad&RaDIsChangingLane != 0 {
			newTad ^= RaDIsOvertaking
			if newTad&RaDIsOvertaking == 0 {
				newTad ^= RaDIsChangingLane
			}
		}
	} else {
		newSub = uint16(len(TrackSubPosition(newTad&TaDBasicMask))) - 1 - v2.SubPosition
		nx, ny, nz = trackEndTile(nx, ny, nz, newTad&TaDBasicMask)
		newTad ^= 4
	}
	comps := trainComponents(tr)
	for _, i := range comps {
		e := &m.Entities[i]
		e.RoutingHandle = uint16(RoutingHandle(e.RoutingHandle).WithIndex(0))
		e.TileX, e.TileY, e.TileBaseZ = coords.WorldUnits(nx), coords.WorldUnits(ny), uint8(nz/4)
		e.SubPosition, e.TrackAndDirection = newSub, newTad
	}
	if road {
		m.UpdateRoadTileOccupancy(nx, ny, nz, newTad&TaDBasicMask, head.TrackType, true)
	}
	var path []MoveInfo
	if road {
		path = RoadSubPosition(head.TrackAndDirection, m.TrafficHandedness != 0)
	} else {
		path = TrackSubPosition(head.TrackAndDirection & TaDBasicMask)
	}
	if int(head.SubPosition) < len(path) {
		mi := path[head.SubPosition]
		for _, i := range comps {
			e := &m.Entities[i]
			m.setPos(e, nx+int32(mi.X), ny+int32(mi.Y), nz+int32(mi.Z))
			e.SpritePitch, e.SpriteYaw = mi.Pitch, mi.Yaw
		}
	}
	head.Var3C = -v2.RemainingDistance
	m.Entities[tr.Veh1].Var3C = -v2.RemainingDistance
	if m.PositionVehicleOnTrack(tr, false, om) {
		head.VehicleFlags |= VehicleFlagCommandStop // liftUpVehicle: PARTIAL
		return
	}
	// connectJacobsBogies: PARTIAL
	if !road {
		m.LeaveLevelCrossing(int32(head.TileX), int32(head.TileY), int32(head.TileBaseZ)*4, head.TrackAndDirection&TaDBasicMask, LevelCrossingEnter)
	}
	head.Var3C = 0
	m.Entities[tr.Veh1].Var3C = 0
	m.Entities[tr.Veh1].TargetSpeed = 0
	m.Entities[tr.Veh2].CurrentSpeed = 0
	m.UpdateTrainProperties(tr, om)
}

// TryReverse is tryReverse: on its expected piece the train reverses (with
// var_52 set for the duration); otherwise upstream destroys it (PARTIAL:
// reported as false).
func (m *Map) TryReverse(tr Train, om *objects.ObjectManager) bool {
	if !m.IsOnExpectedRoadOrTrack(tr) {
		return false
	}
	head := &m.Entities[tr.Head]
	tmp := head.HeadVar52
	head.HeadVar52 = 1
	m.Sub4ADB47(tr, false, om)
	head.HeadVar52 = tmp
	return true
}

// Signal timeouts (categoriseTimeElapsed).
const (
	TimeoutOK         = 0
	TimeoutFirst      = 1
	TimeoutTurnaround = 2
	busSignalTimeout  = 960  // kBusSignalTimeout
	tramSignalTimeout = 2880 // kTramSignalTimeout
)

// CategoriseTimeElapsed is categoriseTimeElapsed (road): how long Vehicle2
// has stood at Vehicle1's spot — a first timeout (160, or 128 / 64 while
// overtaking) and a turn-round one (960 bus / 2880 tram; 544 / 128 while
// overtaking, which also sets var_5C 40).
func (m *Map) CategoriseTimeElapsed(tr Train, om *objects.ObjectManager) int {
	head, v1, v2 := &m.Entities[tr.Head], &m.Entities[tr.Veh1], &m.Entities[tr.Veh2]
	if v2.RoutingHandle != v1.RoutingHandle || v2.SubPosition != v1.SubPosition {
		v1.TimeAtSignal = 0
		return TimeoutOK
	}
	first, turn := uint16(160), uint16(busSignalTimeout)
	isRoad := head.TrackType == 0xFF
	if !isRoad {
		rd := om.GetRoadObjectByIndex(int(head.TrackType))
		isRoad = rd != nil && rd.Flags&roadFlagIsRoad != 0
	}
	overtaking := v1.TrackAndDirection&RaDIsOvertaking != 0
	if isRoad {
		if overtaking {
			first, turn = 128, 544
		}
	} else {
		turn = tramSignalTimeout
		if overtaking {
			first, turn = 64, 128
		}
	}
	v1.TimeAtSignal++
	if v1.TimeAtSignal == first {
		return TimeoutFirst
	}
	if v1.TimeAtSignal == turn {
		head.Var5C = 40
		return TimeoutTurnaround
	}
	return TimeoutOK
}

// Sub4A8DB7 is sub_4A8DB7: give up the reservation ahead (sub_4AD778) and
// stop approaching.
func (m *Map) Sub4A8DB7(tr Train, om *objects.ObjectManager) {
	m.Sub4AD778(tr, om)
	if h := &m.Entities[tr.Head]; h.Status == StatusApproaching {
		h.Status = StatusTravelling
	}
}
