package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Vehicle placement — Vehicle.cpp applyVehicleObjectLength (+ToBogies) and
// VehicleHead.cpp positionVehicleOnTrack (0x004AE15C-side).
//
// PARTIAL (5.6b): liftUpVehicle and connectJacobsBogies are not ported.

const objDistToHighPrecisionDistance = 2179 // kObjDistToHighPrecisionDistance

// applyVehicleObjectLengthToBogies spaces each car's bogies along the train
// by its object's bogie positions and body lengths (reversed cars from their
// last component), starting at startDistance; returns where the train ends.
func (m *Map) applyVehicleObjectLengthToBogies(tr Train, start int32, om *objects.ObjectManager) int32 {
	d := start
	for _, car := range tr.Cars {
		veh := vehicleObject(om, m.Entities[car[0].Front].ObjectID)
		if veh == nil {
			continue
		}
		reversed := m.Entities[car[0].Body].Flags38&2 != 0 // isReversed
		for k, cc := range car {
			idx := k
			if reversed {
				idx = int(veh.NumCarComponents) - 1 - k
			}
			if idx < 0 || idx >= len(veh.CarComponents) {
				continue
			}
			oc := veh.CarComponents[idx]
			fpos, bpos := int32(oc.FrontBogiePosition), int32(oc.BackBogiePosition)
			if reversed {
				fpos, bpos = bpos, fpos
			}
			m.Entities[cc.Front].RemainingDistance = d - fpos*objDistToHighPrecisionDistance
			if oc.BodySpriteIdx != 0xFF {
				d -= int32(veh.BodySprites[oc.BodySpriteIdx&0x7F].HalfLength) * objDistToHighPrecisionDistance * 2
			}
			m.Entities[cc.Back].RemainingDistance = d + bpos*objDistToHighPrecisionDistance
		}
	}
	return d
}

// ApplyVehicleObjectLength is applyVehicleObjectLength: the train's length
// measured from 0, then laid out again so that the tail ends at 0.
func (m *Map) ApplyVehicleObjectLength(tr Train, om *objects.ObjectManager) {
	neg := -m.applyVehicleObjectLengthToBogies(tr, 0, om)
	for _, i := range []int{tr.Head, tr.Veh1, tr.Veh2} {
		m.Entities[i].RemainingDistance = neg
	}
	m.applyVehicleObjectLengthToBogies(tr, neg, om)
	m.Entities[tr.Tail].RemainingDistance = 0
}

// trainComponents is Vehicle::applyToComponents' order: head, Vehicle1,
// Vehicle2, each car's front bogie, back bogie, body, then the tail.
func trainComponents(tr Train) []int {
	out := []int{tr.Head, tr.Veh1, tr.Veh2}
	for _, car := range tr.Cars {
		for _, cc := range car {
			out = append(out, cc.Front, cc.Back, cc.Body)
		}
	}
	return append(out, tr.Tail)
}

// PositionVehicleOnTrack is positionVehicleOnTrack: reserve ahead (up to 32
// sub_4ACEE7 steps), then settle every component on its path with a zero
// move (bodies between their bogies), and re-evaluate the signal blocks at
// both ends of the head's piece. true: a component could not be placed
// (blocked with ≥ 2454 left over, or held at Vehicle1's routing).
func (m *Map) PositionVehicleOnTrack(tr Train, isPlaceDown bool, om *objects.ObjectManager) bool {
	head := &m.Entities[tr.Head]
	for range 32 {
		var r Sub4ACEE7Result
		if head.Mode == 1 {
			r = m.Sub47DA8D(tr, 0, 0, om)
		} else {
			r = m.Sub4ACEF1(tr, 0, 0, isPlaceDown)
		}
		if r.Al != 0 {
			break
		}
	}
	bad := false
	var front, back int = -1, -1
	for _, i := range trainComponents(tr) {
		e := &m.Entities[i]
		if e.Flags38&1 != 0 { // a body: sub_4AC255
			if front >= 0 && back >= 0 {
				f, b := &m.Entities[front], &m.Entities[back]
				p := PlaceBody(int32(f.PosX), int32(f.PosY), int32(f.PosZ), int32(b.PosX), int32(b.PosY), int32(b.PosZ), e.ObjectSpriteType, vehicleObject(om, e.ObjectID))
				m.setPos(e, p.X, p.Y, p.Z)
				if p.Turned {
					e.SpritePitch, e.SpriteYaw = p.Pitch, p.Yaw
				}
			}
			continue
		}
		var res MotionResult
		if e.EntityType == scenario.VehicleTypeVeh1 && e.Mode == 1 {
			v := &Veh1{MotionComponent: m.component(i), Var3C: e.Var3C, Veh2MaxSpeed: m.Entities[tr.Veh2].MaxSpeed,
				HeadVar3C: &head.Var3C, HeadTaD: &head.TrackAndDirection, HeadVar52: head.HeadVar52, HeadTrackType: head.TrackType, HeadVar53: head.HeadVar53}
			res = m.Veh1UpdateRoadMotion(v, 0, om)
			m.store(i, &v.MotionComponent)
			e.Var3C = v.Var3C
		} else {
			c := m.component(i)
			next := m.TrackNewPiece
			if c.Road {
				next = m.RoadNewPiece
			}
			res = c.UpdateMotion(0, func(c *MotionComponent, f *uint32) bool { return next(c, f, head.HeadVar53, false, 0) })
			m.store(i, &c)
		}
		if e.EntityType != scenario.VehicleTypeHead && res.Flags&(MotionUnkM00|MotionUnkM03) != 0 {
			if res.Flags&MotionUnkM03 != 0 || res.RemainingDistance >= 2454 {
				bad = true
			}
		}
		if e.EntityType == scenario.VehicleTypeBogie {
			front, back = back, i
		}
	}
	if head.Mode != 1 {
		x, y, z := int32(head.TileX), int32(head.TileY), int32(head.TileBaseZ)*4
		tad := m.Routing(RoutingHandle(head.RoutingHandle)) & TaDBasicMask
		rx, ry, rz := trackEndTile(x, y, z, tad)
		m.UpdateBlockSignals(rx, ry, rz, tad^4, head.Owner, head.TrackType)
		m.UpdateBlockSignals(x, y, z, tad, head.Owner, head.TrackType)
	}
	return bad
}

func (m *Map) setPos(e *scenario.VehicleEntity, x, y, z int32) {
	e.PosX, e.PosY, e.PosZ = coords.WorldUnits(x), coords.WorldUnits(y), int16(z)
}
