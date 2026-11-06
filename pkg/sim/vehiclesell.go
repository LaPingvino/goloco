package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Selling (GameCommands/Vehicles/VehicleSell.cpp 0x004AED34) and what it
// deletes with (VehicleManager::deleteCar / deleteTrain).

// VehicleSell sells a whole vehicle (id = its head: every car's front bogie's
// refundCost) or one car (id = that car's front bogie: its refundCost); the
// cost is minus the refund. A test (no apply) checks the company and
// canBeModified; apply deletes it — a car taken out of a placed vehicle, the
// vehicle laid out again and put back where its head was (unless no car is
// left). PARTIAL: setPosition (the money effect's place), the dangling tail's
// timeout (trainDanglingTimeout not kept).
func (m *Map) VehicleSell(c *Cmd, flags uint8, id uint16) (int32, bool) {
	base := m.entity(id)
	if base == nil {
		return 0, false
	}
	tr, err := m.TrainOf(base.HeadID)
	if err != nil {
		return 0, false
	}
	head := &m.Entities[tr.Head]
	whole := head.ID == id
	var refund int64
	carIdx := -1
	if whole {
		for _, car := range tr.Cars {
			refund += int64(m.Entities[car[0].Front].RefundCost)
		}
	} else {
		if base.EntityType != scenario.VehicleTypeBogie {
			return 0, false
		}
		for i, car := range tr.Cars {
			if m.Entities[car[0].Front].ID == id {
				carIdx = i
			}
		}
		refund = int64(base.RefundCost)
	}
	if flags&CmdApply == 0 {
		if !checkCompanyCompatibility(c, base.Owner) || !m.canBeModified(tr) {
			return 0, false
		}
		return int32(-refund), true
	}
	om := c.Objects
	headID := head.ID
	if whole {
		m.deleteTrain(c, headID)
		return int32(-refund), true
	}
	if carIdx < 0 {
		return 0, false
	}
	placed := head.TileX != -1
	px, py, pz, ptad, psub := int32(head.TileX), int32(head.TileY), head.TileBaseZ, head.TrackAndDirection, head.SubPosition
	if placed {
		m.pickUpForRemoval(c, headID, head.Mode, om)
	}
	tr, _ = m.TrainOf(headID)
	m.deleteCar(tr, carIdx)
	m.autoLayoutTrain(headID, om)
	tr, _ = m.TrainOf(headID)
	m.UpdateTrainProperties(tr, om)
	m.ApplyBreakdownToTrain(tr, om)
	if placed && len(tr.Cars) > 0 {
		m.PlaceDownVehicle(headID, px, py, pz, ptad, psub, om)
	}
	return int32(-refund), true
}

// pickUpForRemoval lifts a vehicle as deleteTrain and VehicleSell do: rail
// and road liftUpVehicle, air and water their pickup commands.
func (m *Map) pickUpForRemoval(c *Cmd, headID uint16, mode uint8, om *objects.ObjectManager) {
	switch objects.TransportMode(mode) {
	case objects.TransportModeAir:
		m.VehiclePickupAir(c, CmdApply, headID)
	case objects.TransportModeWater:
		m.VehiclePickupWater(c, CmdApply, headID)
	default:
		m.LiftUpVehicle(headID, om)
	}
}

// deleteCar is VehicleManager::deleteCar: the car's components unlinked (the
// component before its front bogie now points past its last body) and freed.
func (m *Map) deleteCar(tr Train, carIdx int) {
	car := tr.Cars[carIdx]
	frontID := m.Entities[car[0].Front].ID
	var ids []uint16
	for _, cc := range car {
		ids = append(ids, m.Entities[cc.Front].ID, m.Entities[cc.Back].ID, m.Entities[cc.Body].ID)
	}
	next := m.Entities[car[len(car)-1].Body].NextCarID
	for i := range m.Entities {
		if e := &m.Entities[i]; e.NextCarID == frontID && e.HeadID == m.Entities[tr.Head].ID {
			e.NextCarID = next
			break
		}
	}
	for _, id := range ids {
		m.removeEntity(id)
	}
}

// deleteTrain is VehicleManager::deleteTrain: lifted off (pickUpForRemoval),
// every car deleted, then the routing handle, the orders, the head, veh1,
// veh2 and tail freed (deleteBaseVehicle) and the company's transport counts
// recounted. PARTIAL: the windows (goloco's vehicle window is closed by its
// caller), the vehicle's noise and messages (not kept), the company's owner
// status (not kept).
func (m *Map) deleteTrain(c *Cmd, headID uint16) {
	h := m.entity(headID)
	if h == nil {
		return
	}
	owner := h.Owner
	m.pickUpForRemoval(c, headID, h.Mode, c.Objects)
	for {
		tr, err := m.TrainOf(headID)
		if err != nil || len(tr.Cars) == 0 {
			break
		}
		m.deleteCar(tr, 0)
	}
	m.deleteBaseVehicle(headID)
	m.RecalculateTransportCounts(owner)
}
