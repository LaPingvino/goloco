package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// The train's car chain — Vehicle.cpp flipCar (0x004AF16A-side) and
// insertCarBefore (0x004AF4D6), over entity ids and nextCarId links.
// (Window invalidations are upstream's redraw only.)

// previousComponent is previousVehicleComponent: the component of the train
// (from its head) whose nextCarId is id; -1 if none.
func (m *Map) previousComponent(headIdx int, id uint16) int {
	idx := m.entityIndex()
	cur := headIdx
	for steps := 0; steps < 4096; steps++ {
		next := m.Entities[cur].NextCarID
		if next == id {
			return cur
		}
		i, ok := idx[next]
		if !ok {
			return -1
		}
		cur = i
	}
	return -1
}

// FlipCar is flipCar: the car starting at frontBogie turns round — bogie
// sprites swapped, bodies' isReversed toggled, its components linked in
// reverse order (the new first body becomes body_start), and the first
// component's cargo / breakdown / reliability state moved to the new first.
// Returns the car's new front bogie (entity index).
func (m *Map) FlipCar(tr Train, frontBogie int) int {
	var car Car
	for _, c := range tr.Cars {
		if c[0].Front == frontBogie {
			car = c
			break
		}
	}
	if car == nil {
		return frontBogie
	}
	for _, cc := range car {
		f, b, body := &m.Entities[cc.Front], &m.Entities[cc.Back], &m.Entities[cc.Body]
		f.ObjectSpriteType, b.ObjectSpriteType = b.ObjectSpriteType, f.ObjectSpriteType
		body.Flags38 ^= 2 // isReversed
	}
	if len(car) == 1 {
		return frontBogie
	}
	prev := m.previousComponent(tr.Head, m.Entities[frontBogie].ID)
	oldFirst, newFirst := car[0], car[len(car)-1]
	m.Entities[newFirst.Body].EntityType = scenario.VehicleTypeBodyStart
	if prev >= 0 {
		m.Entities[prev].NextCarID = m.Entities[newFirst.Front].ID
	}
	m.Entities[oldFirst.Body].NextCarID = m.Entities[newFirst.Body].NextCarID
	for i := len(car) - 2; i >= 0; i-- {
		m.Entities[car[i].Body].EntityType = scenario.VehicleTypeBodyContinued
		m.Entities[car[i+1].Body].NextCarID = m.Entities[car[i].Front].ID
	}
	nb, ob := &m.Entities[newFirst.Body], &m.Entities[oldFirst.Body]
	nf, of := &m.Entities[newFirst.Front], &m.Entities[oldFirst.Front]
	nb.Cargo, nb.BreakdownFlags, nb.BreakdownTimeout = ob.Cargo, ob.BreakdownFlags, ob.BreakdownTimeout
	nf.Cargo, nf.BreakdownFlags, nf.BreakdownTimeout = of.Cargo, of.BreakdownFlags, of.BreakdownTimeout
	nf.TotalWeight, nf.Reliability, nf.TimeoutToBreakdown = of.TotalWeight, of.Reliability, of.TimeoutToBreakdown
	// vanilla does not reset every value (townFrom stays)
	for _, c := range []*scenario.VehicleCargo{&ob.Cargo, &of.Cargo} {
		c.AcceptedTypes, c.Type, c.MaxQty, c.Qty, c.NumDays = 0, 0xFF, 0, 0, 0
	}
	ob.BreakdownFlags, ob.BreakdownTimeout = 0, 0
	of.BreakdownFlags, of.BreakdownTimeout = 0, 0
	return newFirst.Front
}

// InsertCarBefore is insertCarBefore: the car starting at source moves (to
// dest's train) in front of dest.
func (m *Map) InsertCarBefore(srcTrain Train, source int, destHead int, dest int) {
	if source == dest {
		return
	}
	srcPrev := m.previousComponent(srcTrain.Head, m.Entities[source].ID)
	destHeadID := m.Entities[destHead].ID
	for _, car := range srcTrain.Cars {
		if car[0].Front != source {
			continue
		}
		last := car[0].Body
		for _, cc := range car {
			m.Entities[cc.Front].HeadID = destHeadID
			m.Entities[cc.Back].HeadID = destHeadID
			m.Entities[cc.Body].HeadID = destHeadID
			last = cc.Body
		}
		if srcPrev >= 0 {
			m.Entities[srcPrev].NextCarID = m.Entities[last].NextCarID
		}
		m.Entities[last].NextCarID = m.Entities[dest].ID
		break
	}
	if dp := m.previousComponent(destHead, m.Entities[dest].ID); dp >= 0 {
		m.Entities[dp].NextCarID = m.Entities[source].ID
	}
}
