package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// VehicleHead::updateTrainProperties (0x004B7CC3) with setTrainModFlags:
// the train's required track mods (head var_53), rack-rail mods (Vehicle1
// var_49), fasterAroundCurves, each car's weight with its cargo, the train's
// power, weight, maximum and rack-rail speeds, and accepted cargo types.
//
// PARTIAL: the creation day (Vehicle1 dayCreated), the driving-sound object
// swap, calculateRefundCost and recalculateTrainMinReliability are not
// ported (none of them moves a vehicle).

func (m *Map) setTrainModFlags(tr Train, modNames []string, om *objects.ObjectManager) {
	canRoll := true
	var roadMods, rackMods uint8
	for _, car := range tr.Cars {
		veh := vehicleObject(om, m.Entities[car[0].Front].ObjectID)
		if veh == nil {
			continue
		}
		if veh.BodySprites[0].NumRollFrames == 1 {
			canRoll = false
		}
		if veh.Flags&vehicleFlagRackRail != 0 {
			for i, n := range modNames {
				if sameObject(n, veh.RackRailName) {
					rackMods |= 1 << i
					break
				}
			}
		}
		for _, extra := range veh.ExtraNames {
			for j, n := range modNames {
				if sameObject(n, extra) {
					roadMods |= 1 << j
					break
				}
			}
		}
	}
	head := &m.Entities[tr.Head]
	head.HeadVar53 = roadMods
	m.Entities[tr.Veh1].Var49 = rackMods
	head.Flags38 &^= flags38FasterCurves
	if canRoll {
		head.Flags38 |= flags38FasterCurves
	}
}

// UpdateTrainProperties is updateTrainProperties (see the file comment).
func (m *Map) UpdateTrainProperties(tr Train, om *objects.ObjectManager) {
	head := &m.Entities[tr.Head]
	switch head.Mode {
	case 1:
		id := head.TrackType // 0xFF: GameState lastTrackTypeOption (not kept: no mods then)
		if rd := om.GetRoadObjectByIndex(int(id)); rd != nil && id != 0xFF {
			m.setTrainModFlags(tr, rd.ModNames, om)
		} else {
			m.setTrainModFlags(tr, nil, om)
		}
	case 0:
		if tr0 := om.GetTrackObjectByIndex(int(head.TrackType)); tr0 != nil {
			m.setTrainModFlags(tr, tr0.ModNames, om)
		}
	}
	var power, weight, accepted uint32
	maxSpeed, rackSpeed := speed16Max, speed16Max
	cargoWeight := func(c *VehicleCargoRef) uint16 {
		if c.Type == 0xFF || int(c.Type) >= len(om.CargoObjects) || om.CargoObjects[c.Type] == nil {
			return 0
		}
		return uint16((uint32(om.CargoObjects[c.Type].UnitWeight) * uint32(c.Qty)) / 256)
	}
	for _, car := range tr.Cars {
		front, back, body := &m.Entities[car[0].Front], &m.Entities[car[0].Back], &m.Entities[car[0].Body]
		veh := vehicleObject(om, front.ObjectID)
		if veh == nil {
			continue
		}
		power += uint32(veh.Power)
		w := veh.Weight
		for _, c := range []*VehicleCargoRef{{front.Cargo.Type, front.Cargo.Qty}, {back.Cargo.Type, back.Cargo.Qty}, {body.Cargo.Type, body.Cargo.Qty}} {
			w += cargoWeight(c)
		}
		accepted |= front.Cargo.AcceptedTypes | back.Cargo.AcceptedTypes | body.Cargo.AcceptedTypes
		front.TotalWeight = w // totalCarWeight
		weight += uint32(w)
		maxSpeed = min(maxSpeed, int16(veh.Speed))
		rackSpeed = min(rackSpeed, int16(veh.Speed))
		if veh.Flags&vehicleFlagRackRail != 0 {
			rackSpeed = min(rackSpeed, int16(veh.RackSpeed))
		}
		if veh.Mode == 2 { // air
			rackSpeed = int16(veh.RackSpeed)
		}
	}
	head.TrainAcceptedCargoTypes = accepted
	v2 := &m.Entities[tr.Veh2]
	v2.TotalPower = uint16(min(power, 0xFFFF))
	v2.TotalWeight = uint16(min(weight, 0xFFFF))
	v2.MaxSpeed = maxSpeed
	v2.RackRailMaxSpeed = rackSpeed
	v2.Var4F = -1
}

// VehicleCargoRef is the part of a VehicleCargo the weight sums read.
type VehicleCargoRef struct{ Type, Qty uint8 }
