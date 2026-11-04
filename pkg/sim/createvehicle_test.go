package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// createBaseVehicle on a new game: head, veh1, veh2, tail in slots 0–3,
// linked in that order, the head in the vehicle-head list, an order table
// holding only End, ordinal 1; a second one: slots 4–7, ordinal 2, another
// routing handle. Both pass the train structure check (TrainOf).
func TestCreateBaseVehicle(t *testing.T) {
	m := &Map{EntityLists: resetEntityLists()}
	h1, ok := m.createBaseVehicle(0, 0, 0, 0)
	if !ok || h1 != 0 {
		t.Fatalf("first head %d ok %v", h1, ok)
	}
	want := []uint8{scenario.VehicleTypeHead, scenario.VehicleTypeVeh1, scenario.VehicleTypeVeh2, scenario.VehicleTypeTail}
	for i, e := range m.Entities {
		next := uint16(i + 1)
		if i == 3 {
			next = entityIDNull
		}
		if e.ID != uint16(i) || e.EntityType != want[i] || e.HeadID != 0 || e.NextCarID != next || e.TileX != -1 {
			t.Errorf("entity %d: %+v", i, e)
		}
	}
	head := m.Entities[0]
	if head.OrdinalNumber != 1 || head.Name != 4 || head.VehicleFlags&VehicleFlagCommandStop == 0 || head.StationID != 0xFFFF {
		t.Errorf("head %+v", head)
	}
	if m.OrderTableLength != 1 || head.OrderTableOffset != 0 || head.SizeOfOrderTable != 1 || m.Orders[0] != OrderEnd {
		t.Errorf("orders: length %d, head offset %d size %d", m.OrderTableLength, head.OrderTableOffset, head.SizeOfOrderTable)
	}
	if l := m.EntityLists; l.Counts[EntityListVehicleHead] != 1 || l.Counts[EntityListVehicle] != 3 || l.Heads[EntityListVehicleHead] != 0 {
		t.Errorf("lists: heads %v counts %v", l.Heads, l.Counts)
	}
	h2, _ := m.createBaseVehicle(0, 0, 0, 0)
	if h2 != 4 || m.Entities[4].OrdinalNumber != 2 || m.Entities[4].RoutingHandle == head.RoutingHandle || m.Entities[4].OrderTableOffset != 1 {
		t.Errorf("second: head %d %+v", h2, m.Entities[4])
	}
	for _, h := range []uint16{h1, h2} {
		if _, err := m.TrainOf(h); err != nil {
			t.Errorf("TrainOf(%d): %v", h, err)
		}
	}
}

// On the title map (local game data; skips without): the new head takes the
// save's free-list head, the order table grows at its end.
func TestCreateBaseVehicleOnSave(t *testing.T) {
	sc, err := scenario.LoadScenarioData("../../../locomotion/Data/title.dat")
	if err != nil || sc.EntityLists == nil {
		t.Skip("game data not present")
	}
	var m Map
	m.LoadScenario(sc)
	free, orders := m.EntityLists.Heads[EntityListNull], m.OrderTableLength
	h, ok := m.createBaseVehicle(0, 0, 0, 0)
	if !ok || h != free || m.OrderTableLength != orders+1 {
		t.Fatalf("head %d (free head was %d) ok %v; orders %d → %d", h, free, ok, orders, m.OrderTableLength)
	}
	if _, err := m.TrainOf(h); err != nil {
		t.Error(err)
	}
	t.Logf("new head %d, ordinal %d, orders %d → %d", h, m.Entities[len(m.Entities)-4].OrdinalNumber, orders, m.OrderTableLength)
}
