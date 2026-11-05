package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
)

// Taking a vehicle off the track: VehicleHead::liftUpVehicle (0x004B08DD) with
// liftUpTail (VehicleTail.cpp 0x004B090F) and removeAllCargo (Vehicle.cpp
// 0x004AF16A), and the VehiclePickup game command (0x004B0826).
// PARTIAL: the pickup sound (playPickupSound, gPrng2) and the window
// invalidations; removeAllCargo's body cargo sprite update (drawn from the
// cargo state anyway, as load.go's note).

// KeepCargoModifyPickup is Config::keepCargoModifyPickup (a cheat, off).
var KeepCargoModifyPickup = false

// LiftUpVehicle is VehicleHead::liftUpVehicle.
func (m *Map) LiftUpVehicle(headID uint16, om *objects.ObjectManager) {
	tr, err := m.TrainOf(headID)
	if err != nil || m.Entities[tr.Head].TileX == -1 {
		return
	}
	m.Sub4AD778(tr, om)
	tr, _ = m.TrainOf(headID)
	m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 |= flags38Unk2 })
	m.liftUpTail(tr)
	m.eachComponent(tr, func(i int) { // 0x004B0A9A
		e := &m.Entities[i]
		e.Flags38 &^= flags38Unk2
		e.TileX = -1
		m.moveTo(e, -32768, 0, 0) // (0x8000, 0, 0)
	})
	if !KeepCargoModifyPickup {
		for _, car := range tr.Cars {
			for _, cc := range car {
				m.removeAllCargo(tr, cc, om)
			}
		}
	}
	m.Entities[tr.Veh1].Var3C = 0
	h := &m.Entities[tr.Head]
	h.Status = 0 // Status::unk_0
	h.StationID = 0xFFFF
}

// eachComponent is Vehicle::applyToComponents: head, veh1, veh2, every car's
// front bogie, back bogie and body, then the tail.
func (m *Map) eachComponent(tr Train, f func(i int)) {
	f(tr.Head)
	f(tr.Veh1)
	f(tr.Veh2)
	for _, car := range tr.Cars {
		for _, cc := range car {
			f(cc.Front)
			f(cc.Back)
			f(cc.Body)
		}
	}
	f(tr.Tail)
}

// liftUpTail is VehicleTail.cpp liftUpTail: what the tail holds released —
// road lanes, or the block behind (re-evaluated reversed from the piece's far
// end), then along its routings the blocks, signals and level crossings.
func (m *Map) liftUpTail(tr Train) {
	tail := &m.Entities[tr.Tail]
	th := RoutingHandle(tail.RoutingHandle)
	x, y, z := int32(tail.TileX), int32(tail.TileY), int32(tail.TileBaseZ)*4
	if tail.Mode == 1 {
		for _, h := range m.RoutingRing(th, false) {
			tad := m.Routing(h) & TaDBasicMask
			m.UpdateRoadTileOccupancy(x, y, z, tad, tail.TrackType, false)
			c := RoadCoordinates[tad&0x7F]
			x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
		}
		return
	}
	{
		tad := m.Routing(th) & TaDBasicMask
		ex, ey, ez := trackEndTile(x, y, z, tad)
		m.UpdateBlockSignals(ex, ey, ez, tad^(1<<2), tail.Owner, tail.TrackType) // setReversed(!isReversed): bit 2
	}
	for _, h := range m.RoutingRing(th, false) {
		r := m.Routing(h)
		tad := r & TaDBasicMask
		m.UpdateBlockSignals(x, y, z, tad, tail.Owner, tail.TrackType)
		if r&TaDHasSignal != 0 {
			m.SetSignalState(x, y, z, tad, tail.TrackType, 0)
		}
		m.LeaveLevelCrossing(x, y, z, tad, LevelCrossingLeave)
		c := TrackCoordinates[tad]
		x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
	}
}

// removeAllCargo is Vehicle.cpp removeAllCargo: the car component's cargo
// emptied, then the train's properties updated.
func (m *Map) removeAllCargo(tr Train, cc CarComponent, om *objects.ObjectManager) {
	m.Entities[cc.Front].Cargo.Qty = 0
	m.Entities[cc.Back].Cargo.Qty = 0
	m.Entities[cc.Body].Cargo.Qty = 0
	m.UpdateTrainProperties(tr, om)
}

// VehiclePickup is the vehiclePickup game command: 0 and true, or false when
// refused (no such vehicle, another company's, or it can't be modified).
// Applied, the vehicle is lifted off and told to stop (commandStop), its
// components' ghost flags cleared.
func (m *Map) VehiclePickup(c *Cmd, flags uint8, headID uint16) (int32, bool) {
	h := m.entity(headID)
	if h == nil || h.EntityType != 0 {
		return 0, false
	}
	tr, err := m.TrainOf(headID)
	if err != nil || h.Owner != c.Company || !m.canBeModified(tr) {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	m.LiftUpVehicle(headID, c.Objects)
	tr, _ = m.TrainOf(headID)
	m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 &^= flags38IsGhost })
	m.Entities[tr.Head].VehicleFlags |= VehicleFlagCommandStop
	return 0, true
}
