package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A train built by hand: head 1 → veh1 2 → veh2 3 → car A (3 components:
// bogies 10/11 body 12, 13/14 body 15, 16/17 body 18) → car B (20/21 body 22)
// → tail 30.
func chainMap() *Map {
	var m Map
	add := func(id uint16, typ uint8, next uint16) {
		m.Entities = append(m.Entities, scenario.VehicleEntity{ID: id, EntityType: typ, NextCarID: next, HeadID: 1, ObjectSpriteType: uint8(id),
			Cargo: scenario.VehicleCargo{Type: 0xFF}}) // an empty slot, as saves hold it
	}
	add(1, scenario.VehicleTypeHead, 2)
	add(2, scenario.VehicleTypeVeh1, 3)
	add(3, scenario.VehicleTypeVeh2, 10)
	add(10, scenario.VehicleTypeBogie, 11)
	add(11, scenario.VehicleTypeBogie, 12)
	add(12, scenario.VehicleTypeBodyStart, 13)
	add(13, scenario.VehicleTypeBogie, 14)
	add(14, scenario.VehicleTypeBogie, 15)
	add(15, scenario.VehicleTypeBodyContinued, 16)
	add(16, scenario.VehicleTypeBogie, 17)
	add(17, scenario.VehicleTypeBogie, 18)
	add(18, scenario.VehicleTypeBodyContinued, 20)
	add(20, scenario.VehicleTypeBogie, 21)
	add(21, scenario.VehicleTypeBogie, 22)
	add(22, scenario.VehicleTypeBodyStart, 30)
	add(30, scenario.VehicleTypeTail, 0xFFFF)
	m.Entities[5].Cargo = scenario.VehicleCargo{Type: 3, MaxQty: 40, Qty: 7, TownFrom: 9} // car A's first body
	return &m
}

func order(m *Map) []uint16 {
	tr, _ := m.TrainOf(1)
	var out []uint16
	for _, car := range tr.Cars {
		for _, cc := range car {
			out = append(out, m.Entities[cc.Front].ID)
		}
		out = append(out, 0)
	}
	return out
}

func TestFlipCarMultiComponent(t *testing.T) {
	m := chainMap()
	saved := append([]scenario.VehicleEntity(nil), m.Entities...)
	tr, err := m.TrainOf(1)
	if err != nil {
		t.Fatal(err)
	}
	nf := m.FlipCar(tr, tr.Cars[0][0].Front)
	if got := order(m); len(got) != 6 || got[0] != 16 || got[1] != 13 || got[2] != 10 || got[4] != 20 {
		t.Fatalf("after a flip, components %v, want [16 13 10 0 20 0]", got)
	}
	tr1, err := m.TrainOf(1)
	if err != nil {
		t.Fatal(err)
	}
	if b := m.Entities[tr1.Cars[0][0].Body]; b.EntityType != scenario.VehicleTypeBodyStart || b.Cargo.Qty != 7 || b.Cargo.TownFrom != 9 || b.Flags38&2 == 0 {
		t.Errorf("new first body %+v", b)
	}
	if ob := m.Entities[5]; ob.Cargo.Type != 0xFF || ob.Cargo.TownFrom != 9 || ob.EntityType != scenario.VehicleTypeBodyContinued {
		t.Errorf("old first body cargo %+v (vanilla keeps townFrom)", ob.Cargo)
	}
	m.FlipCar(tr1, nf)
	// Vanilla does not reset townFrom: the body that was first after the
	// first flip (18) keeps the cargo's townFrom (9) once the second flip
	// moves the cargo back. Everything else is restored exactly.
	if m.Entities[11].Cargo.TownFrom != 9 {
		t.Errorf("entity 18 townFrom %d, want 9 (left by vanilla's partial reset)", m.Entities[11].Cargo.TownFrom)
	}
	m.Entities[11].Cargo.TownFrom = saved[11].Cargo.TownFrom
	for i := range m.Entities {
		if m.Entities[i] != saved[i] {
			t.Fatalf("entity %d differs after two flips: %+v vs %+v", m.Entities[i].ID, m.Entities[i], saved[i])
		}
	}
	// insertCarBefore: car B to the front (before car A's first bogie).
	tr2, _ := m.TrainOf(1)
	m.InsertCarBefore(tr2, tr2.Cars[1][0].Front, tr2.Head, tr2.Cars[0][0].Front)
	if got := order(m); len(got) != 6 || got[0] != 20 || got[2] != 10 {
		t.Errorf("after insertCarBefore, components %v, want [20 0 10 13 16 0]", got)
	}
}
