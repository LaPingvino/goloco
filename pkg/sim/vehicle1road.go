package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Vehicle1's road motion helpers — a port of Vehicle1.cpp's
// getRoadOvertakeAvailability, veh1UpdateRoadMotionNewRoadPiece (0x0047D308),
// vehicle1UpdateRoadMotionByPieces (0x0047D2D6) and its NoMove twin
// (0x0047D142), applyOvertakeToVehicle1 (0x0047D46F) and
// applyChangeLaneToVehicle1 (0x0047D52B).
//
// lookaheadRoad (0x0047CEB7), updateRoadMotionNewRoadPiece (0x0047CABF) and
// Vehicle1::updateRoadMotion (0x0047CA71) drive them (getRoadOccupation:
// roadoccupation.go).

// Road trackAndDirection flags (AdditionalTaDFlags).
const (
	RaDIsOvertaking    uint16 = 1 << 7
	RaDIsChangingLane  uint16 = 1 << 8
	raDResetLaneChange        = ^(RaDIsChangingLane | RaDIsOvertaking)
)

// OvertakeResult.
const (
	NoOvertakeAvailable = 0
	MayBeOvertaken      = 1
	OvertakeAvailable   = 5
)

// Veh1 is the Vehicle1 state the road motion reads and writes, with what it
// needs of its train (the Vehicle2 behind it, the head).
type Veh1 struct {
	MotionComponent
	Var3C         int32
	Veh2MaxSpeed  int16
	HeadVar3C     *int32  // the head's var_3C (applyOvertake / applyChangeLane)
	HeadTaD       *uint16 // the head's trackAndDirection (applyChangeLane)
	HeadVar52     uint8   // the head's var_52 (1: no lookahead)
	HeadTrackType uint8   // the head's trackType (getRoadConnections)
	HeadVar53     uint8   // the head's required mods
}

// RoadOvertakeAvailability is getRoadOvertakeAvailability: a road tail on
// the piece (pos, tad) ahead — of a train moving (var_3C ≥ 0x220C0) and not
// broken down — that is faster than us (var_3C) or no slower (maxSpeed)
// means we may be overtaken; otherwise we may overtake it.
func (m *Map) RoadOvertakeAvailability(v *Veh1, x, y, z int32, tad uint16) int {
	trains, idx := m.trainsByHead()
	result := NoOvertakeAvailable
	for dy := int32(-1); dy <= 1; dy++ {
		for dx := int32(-1); dx <= 1; dx++ {
			for _, i := range m.vehiclesOnTile(x+dx*32, y+dy*32) {
				e := &m.Entities[i]
				if e.EntityType != scenario.VehicleTypeTail || e.Mode != 1 {
					continue
				}
				if int32(e.TileBaseZ)*4 != z || int32(e.TileX) != x || int32(e.TileY) != y || e.TrackAndDirection&TaDBasicMask != tad&TaDBasicMask {
					continue
				}
				tr, ok := trains[idx[e.HeadID]]
				if !ok {
					continue
				}
				o1, o2 := &m.Entities[tr.Veh1], &m.Entities[tr.Veh2]
				if o1.Var3C < 0x220C0 || o2.Var73&1 != 0 {
					continue
				}
				if v.Var3C < o1.Var3C || v.Veh2MaxSpeed <= o2.MaxSpeed {
					return MayBeOvertaken
				}
				result = OvertakeAvailable
			}
		}
	}
	return result
}

// trainsByHead maps head entity indices to their trains.
func (m *Map) trainsByHead() (map[int]Train, map[uint16]int) {
	idx := m.entityIndex()
	out := map[int]Train{}
	for i := range m.Entities {
		if m.Entities[i].EntityType == scenario.VehicleTypeHead {
			if t, err := m.trainOf(m.Entities[i].ID, idx); err == nil {
				out[i] = t
			}
		}
	}
	return out, idx
}

// veh1NewRoadPiece is veh1UpdateRoadMotionNewRoadPiece: onto the next
// routing's piece, marking its lanes occupied (no connection check).
func (m *Map) veh1NewRoadPiece(v *Veh1) bool {
	next := v.RoutingHandle.WithIndex(v.RoutingHandle.Index() + 1)
	routing := m.Routing(next)
	if routing == RoutingAllocatedButFree {
		return false
	}
	c := RoadCoordinates[v.TrackAndDirection&0x7F]
	nx, ny, nz := v.TileX+int32(c.DX), v.TileY+int32(c.DY), int32(v.TileBaseZ)*4+int32(c.DZ)
	rad := routing & 0x1FF
	m.UpdateRoadTileOccupancy(nx, ny, nz, rad, v.TrackType, true)
	v.RoutingHandle = next
	v.TrackAndDirection = rad
	v.TileX, v.TileY, v.TileBaseZ = nx, ny, uint8(nz/4)
	return true
}

// veh1MoveByPieces is vehicle1UpdateRoadMotionByPieces: advance until n new
// pieces are entered (or the routings run out — then, as upstream, without
// moving), returning the distance covered.
func (m *Map) veh1MoveByPieces(v *Veh1, n uint8) int32 {
	ix, iy, iz := v.X, v.Y, v.Z
	var moved int32
	for i := uint8(0); i < n; {
		sub := v.SubPosition + 1
		if int(sub) >= len(RoadSubPosition(v.TrackAndDirection, v.RightHand)) {
			if !m.veh1NewRoadPiece(v) {
				return moved // upstream returns without moveTo ("seems wrong but its what vanilla does")
			}
			sub = 0
			i++
		}
		v.SubPosition = sub
		mi := RoadSubPosition(v.TrackAndDirection, v.RightHand)[sub]
		nx, ny, nz := int32(mi.X)+v.TileX, int32(mi.Y)+v.TileY, int32(mi.Z)+int32(v.TileBaseZ)*4
		moved += MovementNibbleToDistance[MovementNibble(ix, iy, iz, nx, ny, nz)]
		ix, iy, iz = nx, ny, nz
		v.SpriteYaw, v.SpritePitch = mi.Yaw, mi.Pitch
	}
	v.X, v.Y, v.Z = ix, iy, iz
	return moved
}

// veh1MoveByPiecesNoMove is vehicle1UpdateRoadMotionByPiecesNoMove: the
// distance the same advance would cover, changing nothing.
func (m *Map) veh1MoveByPiecesNoMove(v *Veh1, n uint8) int32 {
	ix, iy, iz := v.X, v.Y, v.Z
	x, y, z := v.TileX, v.TileY, int32(v.TileBaseZ)*4
	rad, h, sub := v.TrackAndDirection, v.RoutingHandle, v.SubPosition
	var moved int32
	for i := uint8(0); i < n; {
		ns := sub + 1
		if int(ns) >= len(RoadSubPosition(rad, v.RightHand)) {
			h = h.WithIndex(h.Index() + 1)
			routing := m.Routing(h)
			if routing == RoutingAllocatedButFree {
				return moved
			}
			c := RoadCoordinates[rad&0x7F]
			x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
			rad = routing & 0x1FF
			ns = 0
			i++
		}
		sub = ns
		mi := RoadSubPosition(rad, v.RightHand)[ns]
		nx, ny, nz := int32(mi.X)+x, int32(mi.Y)+y, int32(mi.Z)+z
		moved += MovementNibbleToDistance[MovementNibble(ix, iy, iz, nx, ny, nz)]
		ix, iy, iz = nx, ny, nz
	}
	return moved
}

// markLaneChange rewrites the next n routings' lane flags: first changing
// lane, middle overtaking, the last (overtake only) both.
func (m *Map) markLaneChange(v *Veh1, n uint8, overtake bool) {
	ring := m.RoutingRing(v.RoutingHandle, false)
	for i := 0; i < int(n) && i+1 < len(ring); i++ {
		h := ring[i+1]
		r := m.Routing(h) & raDResetLaneChange
		switch {
		case i == 0:
			r |= RaDIsChangingLane
		case overtake && i == int(n)-1:
			r |= RaDIsChangingLane | RaDIsOvertaking
		default:
			r |= RaDIsOvertaking
		}
		m.SetRouting(h, r)
	}
}

// ApplyOvertake is applyOvertakeToVehicle1.
func (m *Map) ApplyOvertake(v *Veh1, n uint8) {
	without := m.veh1MoveByPiecesNoMove(v, n)
	m.markLaneChange(v, n, true)
	d := m.veh1MoveByPieces(v, n)
	v.Var3C += d
	if v.HeadVar3C != nil {
		*v.HeadVar3C += d - without
	}
}

// ApplyChangeLane is applyChangeLaneToVehicle1.
func (m *Map) ApplyChangeLane(v *Veh1, n uint8) {
	without := m.veh1MoveByPiecesNoMove(v, n)
	m.markLaneChange(v, n, false)
	d := m.veh1MoveByPieces(v, n)
	v.Var3C += d
	if v.HeadVar3C != nil {
		*v.HeadVar3C += d - without
	}
	if v.HeadTaD != nil {
		*v.HeadTaD &= raDResetLaneChange
		if n <= 1 {
			*v.HeadTaD |= RaDIsChangingLane
		} else {
			*v.HeadTaD |= RaDIsOvertaking
		}
	}
}

// Lookahead types (LookaheadType).
const (
	LookaheadNone = iota
	LookaheadOvertake
	LookaheadNoRoute
	LookaheadLevelCrossing
	LookaheadChangeLane
)

const roadFlagIsRoad uint16 = 1 << 6 // RoadObjectFlags::isRoad (else tram track)

// LookaheadRoad is lookaheadRoad: whether Vehicle1, entering its next piece,
// should cross a level crossing in one go, overtake or change lane, or may
// not go on — and over how many pieces.
func (m *Map) LookaheadRoad(v *Veh1, om *objects.ObjectManager) (int, uint8) {
	n := uint8(1)
	c := RoadCoordinates[v.TrackAndDirection&0x7F]
	x, y, z := v.TileX+int32(c.DX), v.TileY+int32(c.DY), int32(v.TileBaseZ)*4+int32(c.DZ)
	ring := m.RoutingRing(v.RoutingHandle, false)
	if len(ring) < 2 {
		return LookaheadNone, 0
	}
	k := 1
	tad := m.Routing(ring[k]) & TaDBasicMask
	occ := m.RoadOccupation(x, y, z, tad, om)
	step := func() bool { // pos += getUnkRoad(tad); routingIter++; false at end()
		c := RoadCoordinates[tad&0x7F]
		x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
		k++
		return k < len(ring)
	}
	if occ&RoadHasLevelCrossing != 0 {
		if occ&(RoadLaneOccupied|RoadLevelCrossingClosed) != 0 {
			return LookaheadNoRoute, 0
		}
		for n++; n < 255; n++ {
			if !step() {
				return LookaheadNoRoute, 0
			}
			tad = m.Routing(ring[k]) & TaDBasicMask
			f := m.RoadOccupation(x, y, z, tad, om)
			if f&(RoadLaneOccupied|RoadLevelCrossingClosed) != 0 {
				return LookaheadNoRoute, 0
			}
			if f&RoadHasLevelCrossing == 0 {
				return LookaheadLevelCrossing, n
			}
		}
		return LookaheadNoRoute, 0
	}
	overtake := OvertakeAvailable
	if occ&RoadLaneOccupied != 0 {
		overtake = m.RoadOvertakeAvailability(v, x, y, z, tad)
		if overtake == MayBeOvertaken {
			return LookaheadNone, 0
		}
	} else {
		nc := RoadCoordinates[tad&0x7F]
		nx, ny, nz := x+int32(nc.DX), y+int32(nc.DY), z+int32(nc.DZ)
		if k+1 >= len(ring) {
			return LookaheadNone, 0
		}
		ntad := m.Routing(ring[k+1]) & TaDBasicMask
		if m.RoadOccupation(nx, ny, nz, ntad, om)&RoadLaneOccupied == 0 {
			return LookaheadNone, 0
		}
		overtake = m.RoadOvertakeAvailability(v, nx, ny, nz, ntad)
		if overtake == MayBeOvertaken {
			return LookaheadNone, 0
		}
	}
	if v.TrackType != 0xFF {
		if rd := om.GetRoadObjectByIndex(int(v.TrackType)); rd == nil || rd.Flags&roadFlagIsRoad == 0 {
			return LookaheadNone, 0
		}
	}
	if tad&(RaDIsChangingLane|RaDIsOvertaking) != 0 {
		return LookaheadNone, 0
	}
	btf := m.RoadOccupation(x, y, z, tad^RaDIsOvertaking, om)
	if btf&(RoadHasLevelCrossing|RoadLaneOccupied) != 0 {
		return LookaheadNone, 0
	}
	need := 0
	if overtake == OvertakeAvailable {
		need = 5
	}
	i := 0
	for n++; n < 11; n++ {
		if !step() {
			if btf&RoadHasStation == 0 {
				return LookaheadNone, 0
			}
			return LookaheadChangeLane, n - 1
		}
		tad = m.Routing(ring[k]) & TaDBasicMask
		btf = m.RoadOccupation(x, y, z, tad^RaDIsOvertaking, om)
		if btf&(RoadHasLevelCrossing|RoadLaneOccupied) != 0 {
			return LookaheadNone, 0
		}
		if m.RoadOccupation(x, y, z, tad, om)&RoadLaneOccupied != 0 {
			i = 0
			continue
		}
		i++
		if i <= need {
			continue
		}
		return LookaheadOvertake, n
	}
	return LookaheadNone, 0
}

// RoadMotionNewPieceResult.
const (
	roadContinue = iota
	roadNoFurther
	roadPerformedLookahead
)

// veh1RoadNewPiece is Vehicle1's updateRoadMotionNewRoadPiece.
func (m *Map) veh1RoadNewPiece(v *Veh1, flags *uint32, om *objects.ObjectManager) int {
	next := v.RoutingHandle.WithIndex(v.RoutingHandle.Index() + 1)
	routing := m.Routing(next)
	if routing == RoutingAllocatedButFree {
		return roadNoFurther
	}
	if v.HeadVar52 != 1 {
		switch typ, n := m.LookaheadRoad(v, om); typ {
		case LookaheadOvertake:
			m.ApplyOvertake(v, n)
			return roadPerformedLookahead
		case LookaheadNoRoute:
			return roadNoFurther
		case LookaheadLevelCrossing:
			v.Var3C += m.veh1MoveByPieces(v, n)
			return roadPerformedLookahead
		case LookaheadChangeLane:
			m.ApplyChangeLane(v, n)
			return roadPerformedLookahead
		}
	}
	x, y, z := v.TileX, v.TileY, int32(v.TileBaseZ)*4
	e, ok := RoadConnectionEnd(x, y, z, v.TrackAndDirection&0x7F)
	if !ok {
		return roadNoFurther
	}
	tc := m.RoadConnections(e.X, e.Y, e.Z, e.NextRotation, v.Owner, v.HeadTrackType, v.HeadVar53, 0)
	rad := routing & 0x1FF
	found := false
	for _, conn := range tc.Connections {
		if conn&0x7F == routing&0x7F {
			found = true
			break
		}
	}
	if !found {
		*flags |= MotionNoRouteFound
		return roadNoFurther
	}
	bad := RoadLaneOccupied
	if v.HeadVar52 != 1 {
		bad |= RoadLevelCrossingClosed
	}
	if m.RoadOccupation(e.X, e.Y, e.Z, rad, om)&bad != 0 {
		return roadNoFurther
	}
	v.TrackType = m.UpdateRoadTileOccupancy(e.X, e.Y, e.Z, rad, v.TrackType, true)
	v.RoutingHandle = next
	v.TrackAndDirection = rad
	v.TileX, v.TileY, v.TileBaseZ = e.X, e.Y, uint8(e.Z/4)
	return roadContinue
}

// Veh1UpdateRoadMotion is Vehicle1::updateRoadMotion: as updateRoadMotion,
// but a new piece goes through the lookahead (an overtake, lane change or
// level crossing run ends the step with remainingDistance 0, unmoved further).
func (m *Map) Veh1UpdateRoadMotion(v *Veh1, distance int32, om *objects.ObjectManager) MotionResult {
	var res MotionResult
	v.RemainingDistance += distance
	moved := false
	ix, iy, iz := v.X, v.Y, v.Z
	for v.RemainingDistance >= 0x368A {
		moved = true
		sub := v.SubPosition + 1
		if int(sub) >= len(RoadSubPosition(v.TrackAndDirection, v.RightHand)) {
			switch m.veh1RoadNewPiece(v, &res.Flags, om) {
			case roadNoFurther:
				res.RemainingDistance = v.RemainingDistance - 0x3689
				v.RemainingDistance = 0x3689
				res.Flags |= MotionUnkM00
				goto done
			case roadPerformedLookahead:
				res.RemainingDistance = 0
				return res
			}
			sub = 0
		}
		v.SubPosition = sub
		mi := RoadSubPosition(v.TrackAndDirection, v.RightHand)[sub]
		nx, ny, nz := int32(mi.X)+v.TileX, int32(mi.Y)+v.TileY, int32(mi.Z)+int32(v.TileBaseZ)*4
		v.RemainingDistance -= MovementNibbleToDistance[MovementNibble(ix, iy, iz, nx, ny, nz)]
		ix, iy, iz = nx, ny, nz
		v.SpriteYaw, v.SpritePitch = mi.Yaw, mi.Pitch
	}
done:
	if moved {
		v.X, v.Y, v.Z = ix, iy, iz
	}
	return res
}

// Vehicle1RoadMove is the move that ends Vehicle1::updateRoad: by
// min(currentSpeed/2 − var_3C, max(unkDistance1·4, 0xCC48)) through
// Veh1UpdateRoadMotion; var_3C grows by what was travelled. It returns the
// motion flags (noRouteFound: upstream then calls the head's sub_4AD93A).
func (m *Map) Vehicle1RoadMove(v *Veh1, veh2CurrentSpeed, unkDistance1 int32, om *objects.ObjectManager) uint32 {
	d := veh2CurrentSpeed/2 - v.Var3C
	d = min(d, max(unkDistance1*4, 0xCC48))
	res := m.Veh1UpdateRoadMotion(v, d, om)
	v.Var3C += d - res.RemainingDistance
	return res.Flags
}
