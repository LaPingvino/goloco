package sim

import "testing"

func TestVehicleSell(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	saved := firstRailTrain(t, m, trains, 2)
	company := m.Entities[saved.Head].Owner
	c := &Cmd{Company: company, Objects: om}
	objID := m.Entities[saved.Cars[len(saved.Cars)-1][0].Body].ObjectID
	var h uint16
	if _, ok := m.CreateVehicle(c, CmdApply, objID, true, &h); !ok {
		t.Fatal("buying refused")
	}
	if _, ok := m.AddCarToVehicle(c, CmdApply, objID, h, true); !ok {
		t.Fatal("adding a car refused")
	}
	n0, counts0 := len(m.Entities), m.Companies[company].TransportTypeCount
	tr, _ := m.TrainOf(h)
	if len(tr.Cars) != 2 {
		t.Fatalf("%d cars", len(tr.Cars))
	}
	second := m.Entities[tr.Cars[1][0].Front]
	if _, ok := m.VehicleSell(&Cmd{Company: company ^ 1, Objects: om}, 0, h); ok {
		t.Error("another company may sell it")
	}
	// One car: its front bogie's refund, the vehicle left with one car.
	cost, ok := m.VehicleSell(c, CmdApply, second.ID)
	tr, err := m.TrainOf(h)
	if !ok || cost != -int32(second.RefundCost) || err != nil || len(tr.Cars) != 1 || len(m.Entities) != n0-3*1 {
		t.Fatalf("selling a car: cost %d (refund %d) ok %v; %d cars, err %v; entities %d → %d", cost, second.RefundCost, ok, len(tr.Cars), err, n0, len(m.Entities))
	}
	// The whole vehicle: every car's refund, all its entities gone.
	want := -int32(m.Entities[tr.Cars[0][0].Front].RefundCost)
	cost, ok = m.VehicleSell(c, CmdApply, h)
	if !ok || cost != want || m.entity(h) != nil || len(m.Entities) != n0-3*2-4 {
		t.Fatalf("selling the vehicle: cost %d want %d ok %v; left %v; entities %d → %d", cost, want, ok, m.entity(h) != nil, n0, len(m.Entities))
	}
	if got := m.Companies[company].TransportTypeCount; got[0] != counts0[0]-1 {
		t.Errorf("trains counted %d, were %d", got[0], counts0[0])
	}
	// A placed saved train, sold whole: gone, the rest still sound.
	sid := m.Entities[saved.Head].ID
	if _, ok := m.VehicleSell(c, CmdApply, sid); !ok {
		t.Fatal("selling a saved train refused")
	}
	after, err := m.Trains()
	if err != nil || len(after) != len(trains)-1 || m.entity(sid) != nil {
		t.Errorf("after selling a saved train: %d trains (were %d), err %v", len(after), len(trains), err)
	}
	t.Logf("sold a car (%d), a bought vehicle and saved train %d", second.RefundCost, sid)
}

// A player's crashed vehicle, told to start, is released, stopped and sold
// (VehicleChangeRunningMode's 0x004B69C7 branch).
func TestCrashedVehicleSoldOnStart(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	saved := firstRailTrain(t, m, trains, 1)
	company := m.Entities[saved.Head].Owner
	m.PlayerCompanies[0] = company
	c := &Cmd{Company: company, Objects: om}
	sid := m.Entities[saved.Head].ID
	m.Entities[saved.Head].Status = StatusCrashed
	if _, ok := m.VehicleChangeRunningMode(c, CmdApply, sid, RunningModeStart); !ok {
		t.Fatal("refused")
	}
	if m.entity(sid) != nil {
		t.Error("the crashed vehicle is still there")
	}
	if _, err := m.Trains(); err != nil {
		t.Error(err)
	}
}
