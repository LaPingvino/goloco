package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// VehicleRearrange is GameCommands/Vehicles/VehicleRearrange.cpp (0x004AF1DF):
// the car whose front bogie is source moves in front of dest (a car's front
// bogie, or a vehicle's tail: to the end) — within one vehicle or into
// another. The test pass checks both companies, canBeModified for both, the
// same track type and that dest's vehicle takes the car; applying, a car
// moved to another vehicle loses its cargo (the keep-cargo cheat, not kept by
// goloco, is off as by default), both vehicles are lifted, the car inserted,
// both laid out and updated, and put back where they stood (the source only
// while it still has cars). Cost 0.
func (m *Map) VehicleRearrange(c *Cmd, flags uint8, source, dest uint16) (int32, bool) {
	src, dst := m.entity(source), m.entity(dest)
	if src == nil || dst == nil {
		return 0, false
	}
	srcTr, err1 := m.TrainOf(src.HeadID)
	dstTr, err2 := m.TrainOf(dst.HeadID)
	if err1 != nil || err2 != nil {
		return 0, false
	}
	om := c.Objects
	if flags&CmdApply == 0 {
		if !checkCompanyCompatibility(c, src.Owner) || !checkCompanyCompatibility(c, dst.Owner) {
			return 0, false
		}
		if !m.canBeModified(srcTr) || !m.canBeModified(dstTr) {
			return 0, false
		}
		if src.TrackType != dst.TrackType { // incompatible_vehicle
			return 0, false
		}
		if src.EntityType != scenario.VehicleTypeBogie {
			return 0, false
		}
		if !m.isVehicleTypeCompatible(dstTr, src.ObjectID, om) {
			return 0, false
		}
		return 0, true
	}
	if src.EntityType != scenario.VehicleTypeBogie {
		return 0, false
	}
	srcHeadID, dstHeadID := src.HeadID, dst.HeadID
	if srcHeadID != dstHeadID {
		for _, car := range srcTr.Cars {
			for _, cc := range car {
				if m.Entities[cc.Front].ID == source {
					m.removeAllCargo(srcTr, cc, om)
				}
			}
		}
	}
	type backup struct {
		placed      bool
		x, y        int32
		z           uint8
		tad, subPos uint16
	}
	pickup := func(headID uint16) backup {
		h := m.entity(headID)
		if h == nil || h.TileX == -1 {
			return backup{}
		}
		b := backup{true, int32(h.TileX), int32(h.TileY), h.TileBaseZ, h.TrackAndDirection, h.SubPosition}
		m.pickUpForRemoval(c, headID, h.Mode, om)
		return b
	}
	sb := pickup(srcHeadID)
	db := pickup(dstHeadID)
	srcTr, _ = m.TrainOf(srcHeadID)
	dstTr, _ = m.TrainOf(dstHeadID)
	si, di := m.entityIndexOf(source), m.entityIndexOf(dest)
	if si < 0 || di < 0 {
		return 0, false
	}
	m.InsertCarBefore(srcTr, si, dstTr.Head, di)
	m.autoLayoutTrain(dstHeadID, om)
	if tr, err := m.TrainOf(dstHeadID); err == nil {
		m.UpdateTrainProperties(tr, om)
	}
	if srcHeadID != dstHeadID {
		m.autoLayoutTrain(srcHeadID, om)
		if tr, err := m.TrainOf(srcHeadID); err == nil {
			m.UpdateTrainProperties(tr, om)
		}
	}
	if tr, err := m.TrainOf(srcHeadID); sb.placed && err == nil && len(tr.Cars) > 0 {
		m.PlaceDownVehicle(srcHeadID, sb.x, sb.y, sb.z, sb.tad, sb.subPos, om)
	}
	if db.placed {
		m.PlaceDownVehicle(dstHeadID, db.x, db.y, db.z, db.tad, db.subPos, om)
	}
	return 0, true
}

// entityIndexOf is an entity id's index in Entities (-1: none).
func (m *Map) entityIndexOf(id uint16) int {
	for i := range m.Entities {
		if m.Entities[i].ID == id {
			return i
		}
	}
	return -1
}
