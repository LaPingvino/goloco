package sim

import "testing"

// vehicleReverse on a stopped saved train: applied, the train is still
// sound and its head faces another way; another company's is refused.
func TestVehicleReverse(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	tr := firstRailTrain(t, m, trains, 1)
	head := m.Entities[tr.Head]
	if _, ok := m.VehicleReverse(&Cmd{Company: head.Owner ^ 1, Objects: om}, CmdApply, head.ID); ok {
		t.Fatal("another company's train reversed")
	}
	m.Entities[tr.Veh2].CurrentSpeed = 0
	c := &Cmd{Company: head.Owner, Objects: om}
	before := m.Entities[tr.Head].TrackAndDirection
	if _, ok := m.VehicleReverse(c, CmdApply, head.ID); !ok {
		t.Fatal("reverse refused")
	}
	after, err := m.TrainOf(head.ID)
	if err != nil {
		t.Fatalf("after: %v", err)
	}
	if got := m.Entities[after.Head].TrackAndDirection; got == before {
		t.Errorf("head still %#x", got)
	}
	if m.Entities[after.Head].HeadVar52 != head.HeadVar52 || c.Pos == nil {
		t.Errorf("var_52 %d (was %d), pos %v", m.Entities[after.Head].HeadVar52, head.HeadVar52, c.Pos)
	}
	t.Logf("train %d reversed: %#x → %#x", head.ID, before, m.Entities[after.Head].TrackAndDirection)
}
