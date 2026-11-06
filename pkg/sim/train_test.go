package sim

import "testing"

// Trains follows upstream's VehicleList — the vehicle-head list from its head
// along next — on a save that keeps the lists: every head, each once, in the
// list's order.
func TestTrainsFollowVehicleHeadList(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	if m.EntityLists == nil {
		t.Skip("no entity lists in the save")
	}
	trains, err := m.Trains()
	if err != nil {
		t.Fatal(err)
	}
	l := m.EntityLists
	var want []uint16
	for id := l.Heads[EntityListVehicleHead]; id != entityIDNull; id = l.Links[id].Next {
		want = append(want, id)
	}
	if len(trains) != len(want) {
		t.Fatalf("%d trains, the list holds %d", len(trains), len(want))
	}
	slot := 0
	for i := range trains {
		if m.Entities[trains[i].Head].ID != want[i] {
			t.Fatalf("train %d is head %d, the list's %d", i, m.Entities[trains[i].Head].ID, want[i])
		}
		if i > 0 && trains[i].Head < trains[i-1].Head {
			slot++
		}
	}
	t.Logf("%d trains in list order; %d steps against slot order", len(trains), slot)
}
