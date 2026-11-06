package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// VehicleRefit is GameCommands/Vehicles/VehicleRefit.cpp (0x0042F6DB): the
// first car's body takes cargoType — empty, its capacity rescaled by unit
// size (getNumUnitsForCargo), accepting every cargo of that category —, then
// updateTrainProperties. A train that can't be modified, or another
// company's, fails; one with no cars or not refittable succeeds doing
// nothing (upstream returns 0 with an empty error text). Cost 0.
func (m *Map) VehicleRefit(c *Cmd, flags uint8, headID uint16, cargoType uint8) (int32, bool) {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return 0, false
	}
	head := &m.Entities[tr.Head]
	if !m.canBeModified(tr) || !checkCompanyCompatibility(c, head.Owner) {
		return 0, false
	}
	om := c.Objects
	if len(tr.Cars) == 0 || om == nil {
		return 0, true
	}
	body := &m.Entities[tr.Cars[0][0].Body]
	obj := vehicleObject(om, body.ObjectID)
	if obj == nil || obj.Flags&vehicleFlagRefittable == 0 {
		return 0, true
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	cargo := om.CargoFor(obj) // VehicleObject::load's maxCargo / compatibleCargoCategories
	primary := bitScanForward32(uint32(cargo.CompatibleCargoCategories[0]))
	units := NumUnitsForCargo(om, uint32(cargo.MaxCargo[0]), primary, cargoType)
	body.Cargo.Type = cargoType
	body.Cargo.MaxQty = uint8(min(units, 0xFF))
	body.Cargo.Qty = 0
	var accepted uint32
	if newObj := cargoObject(om, cargoType); newObj != nil {
		for id, co := range om.CargoObjects {
			if co != nil && co.CargoCategory == newObj.CargoCategory {
				accepted |= 1 << id
			}
		}
	}
	body.Cargo.AcceptedTypes = accepted
	m.UpdateTrainProperties(tr, om)
	return 0, true
}

// vehicleFlagRefittable is VehicleObjectFlags::refittable.
const vehicleFlagRefittable = 1 << 14

// NumUnitsForCargo is Vehicles::getNumUnitsForCargo (0x0042F6B6): a capacity
// of the primary cargo in units of the new one (by unit size).
func NumUnitsForCargo(om *objects.ObjectManager, maxPrimary uint32, primaryID, newID uint8) uint32 {
	a, b := cargoObject(om, primaryID), cargoObject(om, newID)
	if a == nil || b == nil || b.UnitSize == 0 {
		return 0
	}
	return uint32(a.UnitSize) * maxPrimary / uint32(b.UnitSize)
}

func cargoObject(om *objects.ObjectManager, id uint8) *objects.CargoObject {
	if om == nil || int(id) >= len(om.CargoObjects) {
		return nil
	}
	return om.CargoObjects[id]
}

// bitScanForward32 is Numerics::bitScanForward (0xFF for none, as a cargo id).
func bitScanForward32(v uint32) uint8 {
	for i := 0; i < 32; i++ {
		if v&(1<<i) != 0 {
			return uint8(i)
		}
	}
	return 0xFF
}

// HasAnyCargo is VehicleHead::hasAnyCargo: any car's front bogie (secondary)
// or body (primary) holds cargo — when the window asks before a refit.
func (m *Map) HasAnyCargo(tr Train) bool {
	total := 0
	for _, car := range tr.Cars {
		if f := m.Entities[car[0].Front].Cargo; f.Type != 0xFF {
			total += int(f.Qty)
		}
		if b := m.Entities[car[0].Body].Cargo; b.Type != 0xFF {
			total += int(b.Qty)
		}
	}
	return uint16(total) > 0
}
