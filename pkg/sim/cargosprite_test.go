package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// updateCargoSprite (VehicleBody.cpp 0x004AC039) by hand: a wagon of 40 with
// 4 cargo load frames, 2 animation frames and its coal sprites at offset 4:
// empty → frame 0; 20 of 40 → (128·4/256 = 2) + 4 = 6, × 2 = 12; full → (255·4/256
// = 3) + 4 = 7, × 2 = 14. No sprite type (0xFF) or no capacity: unchanged.
// goloco's bodies used to keep the frame they were saved with: a wagon loaded
// in play was drawn empty, and its saved cargoFrame differed from OpenLoco's.
func TestUpdateCargoSprite(t *testing.T) {
	const coal = 3
	veh := &objects.VehicleObject{}
	veh.BodySprites[1] = objects.VehicleBodySprite{NumCargoLoadFrames: 4, NumAnimationFrames: 2}
	veh.CargoSection[0] = objects.VehicleCargoSlot{MaxCargo: 40, Categories: []uint16{7}, SpriteOffsets: []uint8{4}}
	om := &objects.ObjectManager{Vehicles: []*objects.VehicleObject{veh}}
	om.CargoObjects = make([]*objects.CargoObject, coal+1)
	om.CargoObjects[coal] = &objects.CargoObject{CargoCategory: 7}
	m := &Map{}
	body := &scenario.VehicleEntity{ObjectSpriteType: 1, CargoFrame: 99}
	body.Cargo = scenario.VehicleCargo{Type: coal, MaxQty: 40}
	for _, c := range []struct {
		qty  uint8
		want uint8
	}{{0, 0}, {20, 12}, {40, 14}, {1, 0}} {
		body.Cargo.Qty = c.qty
		m.updateCargoSprite(body, om)
		if body.CargoFrame != c.want {
			t.Errorf("%d of 40: cargoFrame %d, want %d", c.qty, body.CargoFrame, c.want)
		}
	}
	body.Cargo.Qty, body.ObjectSpriteType = 40, 0xFF
	m.updateCargoSprite(body, om)
	if body.CargoFrame != 0 {
		t.Errorf("no sprite type: cargoFrame changed to %d", body.CargoFrame)
	}
}

// calculateRefundCost (VehicleHead.cpp 0x004B82AE) ends updateTrainProperties
// (VehicleHead.cpp 7002): a car added in play counts in the head's
// totalRefundCost at once — OpenLoco's save after adding a wagon to Boulder
// Breakers' Train 2 had it, goloco's only after the next month. And a new
// entity is named StringIds::empty_pop, which is 1 (StringIds.h), not 8.
func TestRefundCostAndNewEntityName(t *testing.T) {
	m := &Map{Entities: make([]scenario.VehicleEntity, 3)}
	m.Entities[1].RefundCost, m.Entities[2].RefundCost = 4000, 367
	tr := Train{Head: 0, Cars: []Car{{{Front: 1}}, {{Front: 2}}}}
	m.calculateRefundCost(tr)
	if got := m.Entities[0].TotalRefundCost; got != 4367 {
		t.Errorf("totalRefundCost %d, want 4367", got)
	}
	m2 := &Map{EntityLists: resetEntityLists()}
	e, ok := m2.newVehicleEntity(scenario.VehicleTypeBogie)
	if !ok || e.Name != 1 {
		t.Errorf("new entity name %d (ok %v), want empty_pop = 1", e.Name, ok)
	}
}
