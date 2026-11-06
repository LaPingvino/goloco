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

// Sold as a company command: the refund comes back as cash, booked under
// VehicleDisposals (expenditures −= payment: the refund adds); a moving one
// is refused.
func TestVehicleSellPays(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	saved := firstRailTrain(t, m, trains, 1)
	company := m.Entities[saved.Head].Owner
	c := &Cmd{Company: company, Objects: om}
	var refund int64
	for _, car := range saved.Cars {
		refund += int64(m.Entities[car[0].Front].RefundCost)
	}
	m.Entities[saved.Veh2].CurrentSpeed = 0 // canBeModified: a moving train can't be sold
	co := &m.Companies[company]
	cash, booked := co.Cash, co.Expenditures[0][ExpenditureVehicleDisposals]
	for _, tr := range trains { // a running placed train (well past the slow-speed limit) is refused
		v2, v1 := m.Entities[tr.Veh2], m.Entities[tr.Veh1]
		if tr.Head != saved.Head && m.Entities[tr.Head].TileX != -1 && v2.CurrentSpeed != 0 && v1.Var3C > 0x3689 {
			if _, ok := m.VehicleSell(&Cmd{Company: m.Entities[tr.Head].Owner, Objects: om}, 0, m.Entities[tr.Head].ID); ok {
				t.Error("a moving train may be sold")
			}
			break
		}
	}
	cost, ok := m.DoCommandFor(c, CmdApply, ExpenditureVehicleDisposals, func(f uint8) (int32, bool) {
		return m.VehicleSell(c, f, m.Entities[saved.Head].ID)
	})
	if !ok || int64(cost) != -refund {
		t.Fatalf("cost %d ok %v, refund %d", cost, ok, refund)
	}
	if got := int64(co.Cash) - int64(cash); got != refund {
		t.Errorf("cash rose by %d, refund %d", got, refund)
	}
	if got := int64(co.Expenditures[0][ExpenditureVehicleDisposals]) - int64(booked); got != refund { // expenditures -= payment (−refund)
		t.Errorf("booked %d under disposals, want %d", got, refund)
	}
}
