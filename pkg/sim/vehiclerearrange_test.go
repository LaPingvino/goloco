package sim

import "testing"

func carFronts(m *Map, h uint16) []uint16 {
	tr, _ := m.TrainOf(h)
	var ids []uint16
	for _, car := range tr.Cars {
		ids = append(ids, m.Entities[car[0].Front].ID)
	}
	return ids
}

func TestVehicleRearrange(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	saved := firstRailTrain(t, m, trains, 3)
	company := m.Entities[saved.Head].Owner
	c := &Cmd{Company: company, Objects: om}
	m.Entities[saved.Veh2].CurrentSpeed = 0
	sh := m.Entities[saved.Head].ID
	before := carFronts(m, sh)
	last, second := before[len(before)-1], before[1]
	run := func(src, dst uint16) bool {
		_, ok := m.DoCommandFor(c, CmdApply, ExpenditureVehiclePurchases, func(f uint8) (int32, bool) { return m.VehicleRearrange(c, f, src, dst) })
		return ok
	}
	if !run(last, second) {
		t.Fatal("rearranging within the train refused")
	}
	after := carFronts(m, sh)
	if len(after) != len(before) || after[1] != last {
		t.Errorf("within: %v → %v (want %d second)", before, after, last)
	}
	// Between two bought vehicles: the second car of A to the end of B (B's tail).
	objID := m.Entities[saved.Cars[len(saved.Cars)-1][0].Body].ObjectID
	var a, b uint16
	m.CreateVehicle(c, CmdApply, objID, true, &a)
	m.AddCarToVehicle(c, CmdApply, objID, a, true)
	m.CreateVehicle(c, CmdApply, objID, true, &b)
	trA, _ := m.TrainOf(a)
	trB, _ := m.TrainOf(b)
	moved := m.Entities[trA.Cars[1][0].Front].ID
	m.Entities[trA.Cars[1][0].Body].Cargo.Qty = 5
	if !run(moved, m.Entities[trB.Tail].ID) {
		t.Fatal("moving a car to another vehicle refused")
	}
	fa, fb := carFronts(m, a), carFronts(m, b)
	if len(fa) != 1 || len(fb) != 2 || fb[1] != moved {
		t.Errorf("between: A %v, B %v (moved %d)", fa, fb, moved)
	}
	trB, _ = m.TrainOf(b)
	if q := m.Entities[trB.Cars[1][0].Body].Cargo.Qty; q != 0 {
		t.Errorf("the moved car kept %d cargo", q)
	}
	if _, err := m.Trains(); err != nil {
		t.Error(err)
	}
	// Another company can't.
	if _, ok := m.VehicleRearrange(&Cmd{Company: company ^ 1, Objects: om}, 0, fb[1], fb[0]); ok {
		t.Error("another company rearranged it")
	}
	t.Logf("within %v → %v; A %v, B %v", before, after, fa, fb)
}
