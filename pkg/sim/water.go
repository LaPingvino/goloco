package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Boats are placed at docks and lifted from them: the game commands
// vehiclePlaceWater (GameCommands/Vehicles/VehiclePlaceWater.cpp, 0x004267BE)
// and vehiclePickupWater (VehiclePickupWater.cpp, 0x004279CC). PARTIAL: the
// sounds (playWaterPlacedownSound, playPickupSound), the viewport
// invalidation.

// VehicleWaterPlacement is VehicleWaterPlacementArgs: the dock's station
// element position (z its base height), the vehicle, and whether a ghost
// becomes the vehicle.
type VehicleWaterPlacement struct {
	X, Y, Z      int32
	Head         uint16
	ConvertGhost bool
}

// dockAt is the station element at (x, y) whose base height is z.
func (m *Map) dockAt(x, y, z int32) *scenario.StationElement {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return nil
	}
	for i := range t.Stations {
		if int32(t.Stations[i].BaseZ)*4 == z {
			return &t.Stations[i]
		}
	}
	return nil
}

// VehiclePlaceWater is the vehiclePlaceWater game command: the boat put at
// the dock's boat position (the dock object's, rotated, from the tile's
// corner + 32, 32), facing out of the dock, stopped, the dock marked
// occupied (flag_6).
func (m *Map) VehiclePlaceWater(c *Cmd, flags uint8, a VehicleWaterPlacement) (int32, bool) {
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
	el := m.dockAt(a.X, a.Y, a.Z)
	if el == nil || el.Ghost || el.AiAllocated {
		return 0, false
	}
	if int(el.StationID) >= len(m.StationNames) || !checkCompanyCompatibility(c, m.StationNames[el.StationID].Owner) {
		return 0, false
	}
	if el.Flag6 {
		return 0, false // vehicle_approaching_or_in_the_way
	}
	var dock *objects.DockObject
	if c.Objects != nil && int(el.ObjectID) < len(c.Objects.DockObjects) {
		dock = c.Objects.DockObjects[el.ObjectID]
	}
	if dock == nil {
		return 0, false
	}
	ox, oy := rotOffset(dock.BoatPosition[0], dock.BoatPosition[1], el.Rotation) // Math::Vector::rotate
	bx, by, bz := a.X+int32(ox)+32, a.Y+int32(oy)+32, a.Z
	if _, water := m.GetHeight(bx, by); water == 0 {
		return 0, false // noWater
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	yaw := ((el.Rotation + 1) & 3) * 16
	m.moveBoatTo(tr, bx, by, bz, yaw, 0) // Pitch::flat
	head = &m.Entities[tr.Head]
	m.moveTo(head, bx, by, bz+32)
	head.Status = StatusStopped
	head.VehicleFlags |= VehicleFlagCommandStop
	head.StationID = el.StationID
	head.TileX, head.TileY, head.TileBaseZ = coords.WorldUnits(a.X), coords.WorldUnits(a.Y), uint8(a.Z/4)
	el.Flag6 = true
	m.Entities[tr.Veh1].Var48 |= flags48Flag2
	if flags&CmdGhost != 0 {
		m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 |= flags38IsGhost })
	}
	return 0, true
}

// moveBoatTo is VehicleHead::moveBoatTo (0x00426E26 region): veh1, veh2 and
// the first car's body moved to the point (tileX 0), the body turned to yaw
// and pitch.
func (m *Map) moveBoatTo(tr Train, x, y, z int32, yaw, pitch uint8) {
	for _, i := range []int{tr.Veh1, tr.Veh2} {
		m.moveTo(&m.Entities[i], x, y, z)
		m.Entities[i].TileX = 0
	}
	if len(tr.Cars) == 0 || len(tr.Cars[0]) == 0 {
		return
	}
	body := &m.Entities[tr.Cars[0][0].Body]
	m.moveTo(body, x, y, z)
	body.SpriteYaw, body.SpritePitch = yaw, pitch
	body.TileX = 0
}

// VehiclePickupWater is the vehiclePickupWater game command: the boat
// lifted off its dock (the dock's flag_6 cleared), every component off the
// map, ghost flags cleared, stopped, its cargo gone (unless
// KeepCargoModifyPickup).
func (m *Map) VehiclePickupWater(c *Cmd, flags uint8, headID uint16) (int32, bool) {
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
	if head.StationID != stationIDNull {
		if t := m.TileAt(int(head.TileX)>>5, int(head.TileY)>>5); t != nil { // getTrackLoc
			for i := range t.Stations {
				el := &t.Stations[i]
				if el.BaseZ != head.TileBaseZ || el.Ghost || el.AiAllocated {
					continue
				}
				el.Flag6 = false
				head.StationID = stationIDNull
			}
		}
	}
	m.eachComponent(tr, func(i int) {
		e := &m.Entities[i]
		e.TileX = -1
		m.moveTo(e, -32768, 0, 0) // (0x8000, 0, 0)
	})
	head = &m.Entities[tr.Head]
	head.Status = vehicleStatusUnk0
	m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 &^= flags38IsGhost })
	head.VehicleFlags |= VehicleFlagCommandStop
	if !KeepCargoModifyPickup {
		for _, car := range tr.Cars {
			for _, cc := range car {
				m.removeAllCargo(tr, cc, c.Objects)
			}
		}
	}
	return 0, true
}

// RotateOffset is Math::Vector::rotate of an object's offset by an element's
// rotation (for the UI's placement searches).
func RotateOffset(x, y int16, rotation uint8) (int16, int16) { return rotOffset(x, y, rotation) }
