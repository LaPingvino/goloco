package sim

import (
	"bytes"
	"testing"
)

// cloneVehicle on a saved train: refused while one of its cars is locked for
// the company; unlocked, the query costs what the cars cost, and applied a
// new vehicle has the same cars, colours and orders.
func TestCloneVehicle(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	trains, _ := m.Trains()
	saved := firstRailTrain(t, m, trains, 2)
	head := m.Entities[saved.Head]
	c := &Cmd{Company: head.Owner, Objects: om}
	var objs []uint16
	for _, car := range saved.Cars {
		objs = append(objs, m.Entities[car[0].Front].ObjectID)
	}
	for len(m.UnlockedVehicles) <= int(c.Company) {
		m.UnlockedVehicles = append(m.UnlockedVehicles, [7]uint32{})
	}
	m.UnlockedVehicles[c.Company] = [7]uint32{}
	if _, ok := m.CloneVehicle(c, 0, head.ID, nil); ok {
		t.Fatal("cloned with its cars locked")
	}
	for _, id := range objs {
		m.UnlockedVehicles[c.Company][id/32] |= 1 << (id % 32)
	}
	var want int32
	for _, id := range objs {
		cost, _ := m.CreateVehicle(c, 0, id, false, nil)
		want += cost
	}
	if cost, ok := m.CloneVehicle(c, 0, head.ID, nil); !ok || cost != want {
		t.Fatalf("query: %d %v, want %d", cost, ok, want)
	}
	var nh uint16
	if _, ok := m.CloneVehicle(c, CmdApply, head.ID, &nh); !ok || nh == EntityIDNull {
		t.Fatalf("clone refused (%d)", nh)
	}
	src, _ := m.TrainOf(head.ID)
	dst, err := m.TrainOf(nh)
	if err != nil || len(dst.Cars) != len(src.Cars) {
		t.Fatalf("clone: %v, %d cars (source %d)", err, len(dst.Cars), len(src.Cars))
	}
	for i := range src.Cars {
		s, d := m.Entities[src.Cars[i][0].Body], m.Entities[dst.Cars[i][0].Body]
		if s.ObjectID != d.ObjectID || s.ColourPrimary != d.ColourPrimary || s.ColourSecondary != d.ColourSecondary {
			t.Errorf("car %d: object %d/%d colours %d,%d / %d,%d", i, s.ObjectID, d.ObjectID, s.ColourPrimary, s.ColourSecondary, d.ColourPrimary, d.ColourSecondary)
		}
	}
	so, do := m.OrderTable(m.Entities[src.Head].OrderTableOffset), m.OrderTable(m.Entities[dst.Head].OrderTableOffset)
	if len(so) != len(do) {
		t.Fatalf("%d orders cloned, source has %d", len(do), len(so))
	}
	for i := range so {
		n := orderSizes[so[i].Type()]
		if !bytes.Equal(m.Orders[so[i].Offset:so[i].Offset+n], m.Orders[do[i].Offset:do[i].Offset+n]) {
			t.Errorf("order %d differs", i)
		}
	}
	t.Logf("cloned train %d (%d cars, %d orders) as %d for %d", head.ID, len(src.Cars), len(so), nh, want)
}
