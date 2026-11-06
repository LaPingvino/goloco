package sim

import (
	"fmt"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Train is Vehicles::Vehicle (Vehicle.cpp): a head's component chain —
// head → Vehicle1 → Vehicle2 → cars → tail, following nextCarId. Fields are
// indices into Map.Entities.
type Train struct {
	Head, Veh1, Veh2, Tail int
	Cars                   []Car
}

// Car is Vehicles::Car: the CarComponents from a body_start up to (not
// including) the next body_start or the tail.
type Car []CarComponent

// CarComponent is Vehicles::CarComponent: front bogie, back bogie, body.
type CarComponent struct{ Front, Back, Body int }

// entityIndex maps entity ids to their index in Entities.
func (m *Map) entityIndex() map[uint16]int {
	idx := make(map[uint16]int, len(m.Entities))
	for i := range m.Entities {
		idx[m.Entities[i].ID] = i
	}
	return idx
}

// TrainOf is the Vehicle(head) constructor with refreshCars; an error is
// upstream's "Bad vehicle structure".
func (m *Map) TrainOf(headID uint16) (Train, error) {
	return m.trainOf(headID, m.entityIndex())
}

func (m *Map) trainOf(headID uint16, idx map[uint16]int) (Train, error) {
	bad := fmt.Errorf("bad vehicle structure (head %d)", headID)
	at := func(id uint16, typ uint8) (int, bool) {
		i, ok := idx[id]
		if !ok || id == 0xFFFF || m.Entities[i].EntityType != typ {
			return 0, false
		}
		return i, true
	}
	next := func(i int) uint16 { return m.Entities[i].NextCarID }
	var t Train
	var ok bool
	if t.Head, ok = at(headID, scenario.VehicleTypeHead); !ok {
		return t, bad
	}
	if t.Veh1, ok = at(next(t.Head), scenario.VehicleTypeVeh1); !ok {
		return t, bad
	}
	if t.Veh2, ok = at(next(t.Veh1), scenario.VehicleTypeVeh2); !ok {
		return t, bad
	}
	cur := next(t.Veh2)
	for {
		i, found := idx[cur]
		if !found {
			return t, bad
		}
		if m.Entities[i].EntityType == scenario.VehicleTypeTail {
			t.Tail = i
			return t, nil
		}
		// CarComponent(component): bogie, bogie, body.
		var cc CarComponent
		if cc.Front, ok = at(cur, scenario.VehicleTypeBogie); !ok {
			return t, bad
		}
		if cc.Back, ok = at(next(cc.Front), scenario.VehicleTypeBogie); !ok {
			return t, bad
		}
		bi, found := idx[next(cc.Back)]
		if !found {
			return t, bad
		}
		bt := m.Entities[bi].EntityType
		if bt != scenario.VehicleTypeBodyStart && bt != scenario.VehicleTypeBodyContinued {
			return t, bad
		}
		cc.Body = bi
		if bt == scenario.VehicleTypeBodyStart || len(t.Cars) == 0 {
			t.Cars = append(t.Cars, Car{})
		}
		t.Cars[len(t.Cars)-1] = append(t.Cars[len(t.Cars)-1], cc)
		cur = next(bi)
	}
}

// Trains builds every head's Train, in entity order.
func (m *Map) Trains() ([]Train, error) {
	idx := m.entityIndex()
	var out []Train
	// VehicleList (VehicleManager.cpp): the vehicle-head list from its head
	// along next — newest first, the order upstream updates them in. Without
	// the lists (or a list that misses a head), slot order.
	if l := m.EntityLists; l != nil {
		n := 0
		for i := range m.Entities {
			if m.Entities[i].EntityType == scenario.VehicleTypeHead {
				n++
			}
		}
		seen := 0
		for id := l.Heads[EntityListVehicleHead]; id != entityIDNull && int(id) < len(l.Links) && seen <= n; id = l.Links[id].Next {
			i, ok := idx[id]
			if !ok || m.Entities[i].EntityType != scenario.VehicleTypeHead {
				break
			}
			t, err := m.trainOf(id, idx)
			if err != nil {
				return out, err
			}
			out = append(out, t)
			seen++
		}
		if seen == n {
			return out, nil
		}
		out = out[:0]
	}
	for i := range m.Entities {
		if m.Entities[i].EntityType != scenario.VehicleTypeHead {
			continue
		}
		t, err := m.trainOf(m.Entities[i].ID, idx)
		if err != nil {
			return out, err
		}
		out = append(out, t)
	}
	return out, nil
}

// UpdateDistances is VehicleUpdateDistances as VehicleHead::update sets them:
// both are Vehicle2's currentSpeed raw >> 7.
func UpdateDistances(veh2CurrentSpeed int32) (unkDistance1, unkDistance2 int32) {
	return veh2CurrentSpeed >> 7, veh2CurrentSpeed >> 7
}
