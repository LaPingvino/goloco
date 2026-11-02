package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Vehicle1's road motion helpers — a port of Vehicle1.cpp's
// getRoadOvertakeAvailability, veh1UpdateRoadMotionNewRoadPiece (0x0047D308),
// vehicle1UpdateRoadMotionByPieces (0x0047D2D6) and its NoMove twin
// (0x0047D142), applyOvertakeToVehicle1 (0x0047D46F) and
// applyChangeLaneToVehicle1 (0x0047D52B).
//
// PARTIAL: lookaheadRoad, getRoadOccupation and Vehicle1::updateRoadMotion
// that drive these are not ported yet.

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
	Var3C        int32
	Veh2MaxSpeed int16
	HeadVar3C    *int32  // the head's var_3C (applyOvertake / applyChangeLane)
	HeadTaD      *uint16 // the head's trackAndDirection (applyChangeLane)
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
