package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// VehicleHead::autoLayoutTrain (VehicleHead.cpp 544) and connectJacobsBogies
// (Vehicle.cpp 896), item 5P.a: after a car is added the train orders itself
// — the first powered car to the front, alternating-direction cars turned
// in turn, top-and-tail cars to the ends, centre cars to the middle, two
// flag_04 cars into the middle of a set of 4+ — then the cars are flipped /
// moved into that order, alternating car sprites set, Jacobs bogies joined.
// A head with shuntCheat keeps the order the player gave.

const (
	vehObjFlagAlternatingDirection = 1 << 0
	vehObjFlagTopAndTailPosition   = 1 << 1
	vehObjFlag04                   = 1 << 4
	vehObjFlagCenterPosition       = 1 << 5
	vehObjFlagAlternatingCarSprite = 1 << 7
)

type carMeta struct {
	frontID                 uint16
	flags                   uint16
	power                   uint16
	isReversed, wasReversed bool
}

func (m *Map) autoLayoutTrain(headID uint16, om *objects.ObjectManager) {
	tr, err := m.TrainOf(headID)
	if err != nil || len(tr.Cars) == 0 {
		return
	}
	shunt := m.Entities[tr.Head].VehicleFlags&vehicleFlagShuntCheat != 0
	var cd []carMeta
	for _, car := range tr.Cars {
		f, body := &m.Entities[car[0].Front], &m.Entities[car[0].Body]
		var flags, power uint16
		if int(f.ObjectID) < len(om.Vehicles) && om.Vehicles[f.ObjectID] != nil {
			flags, power = om.Vehicles[f.ObjectID].Flags, om.Vehicles[f.ObjectID].Power
		}
		rev := body.Flags38&f38IsReversed != 0
		cd = append(cd, carMeta{f.ID, flags, power, rev, rev})
	}
	toFront := func(i int) { c := cd[i]; copy(cd[1:i+1], cd[:i]); cd[0] = c }
	toBack := func(i int) { c := cd[i]; copy(cd[i:], cd[i+1:]); cd[len(cd)-1] = c }
	if !shunt { // the first powered car to the front
		for i := range cd {
			if cd[i].power != 0 {
				toFront(i)
				break
			}
		}
	}
	cur := false // alternating direction
	for i := range cd {
		if cd[i].flags&vehObjFlagAlternatingDirection != 0 {
			cd[i].isReversed = cur
			cur = !cur
		}
	}
	if !shunt { // top-and-tail cars: the first to the front, the last to the back
		isFirst, lastTT := true, -1
		for i := 0; i < len(cd); i++ {
			if cd[i].power != 0 || cd[i].flags&vehObjFlagTopAndTailPosition != 0 {
				if isFirst {
					if cd[i].flags&vehObjFlagTopAndTailPosition != 0 {
						cd[i].isReversed = false
						toFront(i)
					}
					isFirst = false
					continue
				}
				if cd[i].flags&vehObjFlagTopAndTailPosition != 0 {
					lastTT = i
				}
			}
			isFirst = false
		}
		if lastTT != -1 {
			cd[lastTT].isReversed = true
			toBack(lastTT)
		}
	}
	if !shunt { // centre cars in the middle (stable_partition, then rotate), biased to the back when odd
		var non, mid []carMeta
		for _, c := range cd {
			if c.flags&vehObjFlagCenterPosition == 0 {
				non = append(non, c)
			} else {
				mid = append(mid, c)
			}
		}
		k := len(non)/2 + len(non)%2
		cd = append(append(append([]carMeta{}, non[:k]...), mid...), non[k:]...)
	}
	if !shunt { // 4+ flag_04 cars: two of them into the middle
		n4 := 0
		for _, c := range cd {
			if c.flags&vehObjFlag04 != 0 {
				n4++
			}
		}
		if n4 >= 4 {
			middle := len(cd)/2 + 1
			if middle < len(cd) {
				var moved []carMeta
				for i := 1; i < len(cd)-1; i++ {
					if cd[i].flags&vehObjFlag04 != 0 {
						moved = append(moved, cd[i])
						cd = append(cd[:i], cd[i+1:]...)
						i--
						if len(moved) == 2 {
							break
						}
					}
				}
				for len(moved) < 2 { // upstream's toBeMoved array holds a zeroed entry when fewer were found
					moved = append(moved, carMeta{frontID: entityIDNull})
				}
				ins := func(at int, c carMeta) { cd = append(cd[:at], append([]carMeta{c}, cd[at:]...)...) }
				ins(middle-1, moved[0])
				ins(middle, moved[1])
			}
		}
	}
	if !shunt { // apply: from the back, each car flipped if it should be and moved in front of dest
		idx := m.entityIndex()
		dest := tr.Tail
		for i := len(cd) - 1; i >= 0; i-- {
			fi, ok := idx[cd[i].frontID]
			if !ok {
				continue
			}
			cur, _ := m.TrainOf(headID)
			if cd[i].isReversed != cd[i].wasReversed {
				fi = m.FlipCar(cur, fi)
				cur, _ = m.TrainOf(headID)
			}
			m.InsertCarBefore(cur, fi, cur.Head, dest)
			dest = fi
		}
	}
	tr, err = m.TrainOf(headID)
	if err != nil {
		return
	}
	front := true // alternating car sprites
	for _, car := range tr.Cars {
		f := &m.Entities[car[0].Front]
		obj := om.Vehicles[f.ObjectID]
		if obj == nil || obj.Flags&vehObjFlagAlternatingCarSprite == 0 {
			continue
		}
		body := &m.Entities[car[0].Body]
		body.BodyIndex = 1
		if front {
			body.BodyIndex = 0
		}
		body.ObjectSpriteType = obj.CarComponents[body.BodyIndex].BodySpriteIdx &^ spriteIndexIsReversed
		front = !front
	}
	m.connectJacobsBogies(headID, om)
}

// connectJacobsBogies is Vehicle.cpp 896 (with its "changes from vanilla").
func (m *Map) connectJacobsBogies(headID uint16, om *objects.ObjectManager) {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return
	}
	idx := m.entityIndex()
	obj := func(e *scenario.VehicleEntity) *objects.VehicleObject {
		if int(e.ObjectID) < len(om.Vehicles) {
			return om.Vehicles[e.ObjectID]
		}
		return nil
	}
	component := func(front int) CarComponent { // CarComponent(frontBogie): front, the next (back), its next (body)
		b := idx[m.Entities[front].NextCarID]
		return CarComponent{Front: front, Back: b, Body: idx[m.Entities[b].NextCarID]}
	}
	found := 0
	var prev, prev2 CarComponent
	for _, car := range tr.Cars {
		f, b, body := &m.Entities[car[0].Front], &m.Entities[car[0].Back], &m.Entities[car[0].Body]
		if body.Flags38&f38JacobsBogieAvailable != 0 {
			nf, ok := idx[body.NextCarID]
			if !ok {
				return
			}
			next := component(nf)
			if found >= 1 && m.Entities[prev.Body].Flags38&f38JacobsBogieAvailable != 0 {
				if found < 2 {
					return // upstream throws: no second previous component
				}
				if o := obj(f); o != nil {
					f.ObjectSpriteType = o.CarComponents[f.BodyIndex].FrontBogieSpriteIdx
				}
				m.Entities[next.Front].ObjectSpriteType = spriteIndexNull
				m.Entities[prev2.Back].ObjectSpriteType = spriteIndexNull
			} else {
				f.ObjectSpriteType, b.ObjectSpriteType, body.ObjectSpriteType = spriteIndexNull, spriteIndexNull, spriteIndexNull
				nb := &m.Entities[next.Body]
				if o := obj(nb); o != nil {
					m.Entities[next.Front].ObjectSpriteType = o.CarComponents[nb.BodyIndex].FrontBogieSpriteIdx
					if nb.Flags38&f38IsReversed != 0 {
						m.Entities[next.Front].ObjectSpriteType = o.CarComponents[nb.BodyIndex].BackBogieSpriteIdx
					}
				}
			}
		}
		for _, cc := range car {
			cb := &m.Entities[cc.Body]
			if cb.Flags38&f38JacobsBogieAvailable != 0 && cb.EntityType == scenario.VehicleTypeBodyContinued {
				if found == 0 {
					return // upstream throws
				}
				m.Entities[cc.Front].ObjectSpriteType, m.Entities[cc.Back].ObjectSpriteType, cb.ObjectSpriteType = spriteIndexNull, spriteIndexNull, spriteIndexNull
				pb := &m.Entities[prev.Back]
				if o := obj(pb); o != nil {
					pb.ObjectSpriteType = o.CarComponents[pb.BodyIndex].BackBogieSpriteIdx
					if m.Entities[prev.Body].Flags38&f38IsReversed != 0 {
						pb.ObjectSpriteType = o.CarComponents[pb.BodyIndex].FrontBogieSpriteIdx
					}
				}
			}
			prev2, prev = prev, cc
			found++
		}
	}
}
