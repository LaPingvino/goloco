package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
)

// Aircraft are placed at airports and lifted off them (AW.b part 4): the
// game commands vehiclePlaceAir (GameCommands/Vehicles/VehiclePlaceAir.cpp,
// 0x004267BE) and vehiclePickupAir (VehiclePickupAir.cpp), with
// getAirportMovementNodeLoc (World/Station.cpp). PARTIAL: the sounds
// (playPlacedownSound, playPickupSound), the viewport invalidation.

// VehicleAirPlacement is VehicleAirPlacementArgs: the airport's station, the
// movement node to put the aircraft on, the aircraft, and whether a ghost
// becomes the aircraft.
type VehicleAirPlacement struct {
	StationID    uint16
	AirportNode  uint8
	Head         uint16
	ConvertGhost bool
}

// AirportMovementNodeLoc is getAirportMovementNodeLoc: the world position of
// one of the station's airport's movement nodes — rotated about the tile
// centre from the airport's start, on the ground when taxiing, else at the
// airport's height + 255 in flight or 960.
func (m *Map) AirportMovementNodeLoc(om *objects.ObjectManager, stationID uint16, node uint8) (x, y, z int32, ok bool) {
	el, pos, found := m.airportStationElement(stationID)
	if !found || om == nil || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil {
		return 0, 0, 0, false
	}
	a := om.AirportObjects[el.ObjectID]
	if int(node) >= len(a.MovementNodes) {
		return 0, 0, 0, false
	}
	n := a.MovementNodes[node]
	rx, ry := rotOffset(n.X-16, n.Y-16, el.Rotation)
	x, y, z = int32(rx)+16+int32(pos[0]), int32(ry)+16+int32(pos[1]), int32(n.Z)+int32(pos[2])
	if n.Flags&objects.AirportNodeTaxiing == 0 {
		z = int32(pos[2]) + 255
		if n.Flags&objects.AirportNodeInFlight == 0 {
			z = 30 * 32
		}
	}
	return x, y, z, true
}

// VehiclePlaceAir is the vehiclePlaceAir game command: the aircraft put on
// the node (the first edge into it is the one it takes and marks occupied),
// facing away from where that edge comes from (a taildragger nose up),
// stopped and told to stop. Refused: the airport's tile missing, a ghost or
// AI-allocated, no edge into the node, that edge's must-be-clear edges or
// the airport's own (var_B6) occupied, another company's airport, or an
// airport that doesn't take the aircraft.
func (m *Map) VehiclePlaceAir(c *Cmd, flags uint8, a VehicleAirPlacement) (int32, bool) {
	if int(a.StationID) >= len(m.StationNames) {
		return 0, false
	}
	st := &m.StationNames[a.StationID]
	tr, err := m.TrainOf(a.Head)
	if err != nil {
		return 0, false
	}
	head := &m.Entities[tr.Head]
	if !checkCompanyCompatibility(c, head.Owner) {
		return 0, false
	}
	if !a.ConvertGhost && (head.TileX != -1 || len(tr.Cars) == 0) {
		return 0, false
	}
	if a.ConvertGhost {
		m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 &^= flags38IsGhost })
		return 0, true
	}
	pos := st.AirportStartPos
	el := m.dockAt(int32(pos[0]), int32(pos[1]), int32(pos[2])) // the station element at that height
	if el == nil || el.Ghost || el.AiAllocated || c.Objects == nil || int(el.ObjectID) >= len(c.Objects.AirportObjects) || c.Objects.AirportObjects[el.ObjectID] == nil {
		return 0, false
	}
	ap := c.Objects.AirportObjects[el.ObjectID]
	prev := -1
	for i, e := range ap.MovementEdges {
		if e.NextNode == a.AirportNode {
			prev = i
			break
		}
	}
	if prev < 0 {
		return 0, false
	}
	if st.AirportOccupiedEdges&ap.MovementEdges[prev].MustBeClear != 0 || st.AirportOccupiedEdges&ap.PlacementClearEdges != 0 {
		return 0, false // vehicle_approaching_or_in_the_way
	}
	if !checkCompanyCompatibility(c, st.Owner) {
		return 0, false
	}
	v := m.aircraftObject(tr, c.Objects)
	if v == nil || ap.Flags&compatibleAirportType(v) == 0 {
		return 0, false // airport_type_not_suitable_for_aircraft
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	px, py, pz, ok1 := m.AirportMovementNodeLoc(c.Objects, a.StationID, a.AirportNode)
	qx, qy, _, ok2 := m.AirportMovementNodeLoc(c.Objects, a.StationID, ap.MovementEdges[prev].CurNode)
	if !ok1 || !ok2 {
		return 0, false
	}
	yaw := Yaw1(int16(px-qx), int16(py-qy)) ^ (1 << 5)
	pitch := uint8(0)
	if v.Flags&vehicleFlagAircraftIsTaildragger != 0 {
		pitch = pitchUp12deg
	}
	m.movePlaneTo(tr, px, py, pz, yaw, pitch)
	head = &m.Entities[tr.Head]
	head.Status = StatusStopped
	head.VehicleFlags |= VehicleFlagCommandStop
	head.StationID = a.StationID
	head.AirportMovementEdge = uint8(prev)
	st.AirportOccupiedEdges |= 1 << prev
	m.Entities[tr.Veh1].Var48 |= flags48Flag2
	v2 := &m.Entities[tr.Veh2]
	v2.CurrentSpeed, v2.MotorState = 0, MotorStopped
	if flags&CmdGhost != 0 {
		m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 |= flags38IsGhost })
	}
	return 0, true
}

// VehiclePickupAir is the vehiclePickupAir game command: the aircraft
// lifted off (its edge freed), every component off the map, no station,
// told to stop, ghost flags cleared, its cargo gone (unless
// KeepCargoModifyPickup).
func (m *Map) VehiclePickupAir(c *Cmd, flags uint8, headID uint16) (int32, bool) {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return 0, false
	}
	head := &m.Entities[tr.Head]
	if !checkCompanyCompatibility(c, head.Owner) || !m.canBeModified(tr) {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	if head.TileX != -1 {
		m.eachComponent(tr, func(i int) {
			e := &m.Entities[i]
			e.TileX = -1
			m.moveTo(e, -32768, 0, 0) // (0x8000, 0, 0)
		})
		head = &m.Entities[tr.Head]
		if head.StationID != stationIDNull && int(head.StationID) < len(m.StationNames) && head.AirportMovementEdge < 32 {
			m.StationNames[head.StationID].AirportOccupiedEdges &^= 1 << head.AirportMovementEdge
		}
	}
	head.Status = vehicleStatusUnk0
	head.StationID = stationIDNull
	head.VehicleFlags |= VehicleFlagCommandStop
	m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 &^= flags38IsGhost })
	if !KeepCargoModifyPickup {
		for _, car := range tr.Cars {
			for _, cc := range car {
				m.removeAllCargo(tr, cc, c.Objects)
			}
		}
	}
	return 0, true
}
