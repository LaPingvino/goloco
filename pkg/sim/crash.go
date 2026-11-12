package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Train crashes: VehicleBase::destroyTrain (Vehicles/Vehicle.cpp 0x004AA464)
// and VehicleBogie::collision (Vehicles/VehicleBogie.cpp 0x004AA0DF), with
// applyDestructionToComponent. A bogie whose step hits another vehicle
// (CheckForCollisions, called from the motion) crashes its train and the one
// it hit.
//
// PARTIAL: explodeComponent's sound and its ExplosionCloud and
// VehicleCrashParticle entities are not made (goloco has no effect entities
// yet); the crashed parts don't yet fall and burn out (VehicleHead.cpp:4041,
// the bogies' and bodies' updateSegmentCrashed): a crashed train stands still.

const vehicleFlagUnk5 = 1 << 5 // VehicleFlags::unk_5: the bogie that collided

// DestroyTrain is destroyTrain, called by part `self` of train tr: the train
// is crashed (news for the first player company), nothing of it can be sold
// for money any more, the motor stops, and the bogies keep the speed it hit
// at; then the part that crashed it (or the bogie behind a body) is marked
// as exploding and taken off its tile.
func (m *Map) DestroyTrain(tr Train, self int) {
	head := &m.Entities[tr.Head]
	if head.Status != StatusCrashed && head.Status != StatusStuck {
		head.Status, head.CrashedTimeout = StatusCrashed, 0
		if head.Owner == m.PlayerCompanies[0] {
			m.PostMessage(nil, MsgVehicleCrashed, head.Owner, head.ID, 0xFFFF, 0xFFFF)
		}
	}
	for _, car := range tr.Cars {
		for _, cc := range car {
			for _, i := range []int{cc.Front, cc.Back, cc.Body} {
				m.Entities[i].RefundCost = 0
			}
		}
	}
	head.TotalRefundCost = 0
	v2 := &m.Entities[tr.Veh2]
	speed := uint32(v2.CurrentSpeed) // Speed32 getRaw
	v2.MotorState = MotorStopped
	for _, car := range tr.Cars {
		for _, cc := range car {
			m.Entities[cc.Front].Var5A, m.Entities[cc.Back].Var5A = speed, speed
		}
	}
	offTile := func(i int) {
		b := &m.Entities[i]
		b.Var5A |= 1 << 31
		b.TileX, b.TileY, b.TileBaseZ = 0, 0, 0
	}
	switch m.Entities[self].EntityType {
	case scenario.VehicleTypeVeh2:
		if len(tr.Cars) > 0 && len(tr.Cars[0]) > 0 {
			offTile(tr.Cars[0][0].Front)
		}
	case scenario.VehicleTypeBogie:
		offTile(self)
	default: // a body: the back bogie of its car component (else the train's last)
		explode := -1
	cars:
		for _, car := range tr.Cars {
			for _, cc := range car {
				explode = cc.Back
				if cc.Body == self {
					break cars
				}
			}
		}
		if explode >= 0 {
			offTile(explode)
		}
	}
}

// applyDestructionToComponent: the part explodes (PARTIAL: no effect
// entities) and its var_5A becomes an eighth of the speed, flagged exploding.
func (m *Map) applyDestructionToComponent(i int) {
	e := &m.Entities[i]
	e.Var5A &^= 1 << 31
	e.Var5A >>= 3
	e.Var5A |= 1 << 31
}

// BogieCollision is VehicleBogie::collision: the bogie's train is destroyed,
// the bogie and its car's body explode; the train it ran into is destroyed
// too (unless already crashed), and its part that was hit explodes along with
// that part's body. (Upstream looks for the hit part among this train's own
// cars, as written; a part of another train is never found there.)
func (m *Map) BogieCollision(tr Train, bogie int, hit uint16) {
	m.DestroyTrain(tr, bogie)
	m.applyDestructionToComponent(bogie)
	m.Entities[bogie].VehicleFlags |= vehicleFlagUnk5
	for _, car := range tr.Cars {
		for _, cc := range car {
			if cc.Front == bogie || cc.Back == bogie {
				m.applyDestructionToComponent(cc.Body)
				goto collided
			}
		}
	}
collided:
	hi := m.entityIndexOf(hit)
	if hi < 0 {
		return
	}
	other, err := m.TrainOf(m.Entities[hi].HeadID)
	if err != nil {
		return
	}
	if m.Entities[other.Head].Status != StatusCrashed {
		m.DestroyTrain(other, hi)
	}
	for _, car := range tr.Cars {
		for _, cc := range car {
			if cc.Front == hi {
				m.applyDestructionToComponent(cc.Front)
			}
			if cc.Back == hi {
				m.applyDestructionToComponent(cc.Back)
			}
			if cc.Front == hi || cc.Back == hi || cc.Body == hi {
				m.applyDestructionToComponent(cc.Body)
				return
			}
		}
	}
}
