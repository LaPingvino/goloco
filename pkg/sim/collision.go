package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Vehicle collisions (Vehicles/Vehicle.cpp checkForCollisions 0x004B1876,
// ignoreSelfCollision). The detection only: wiring it into the motion
// (updateTrackMotion's collision check, VehicleBogie::collision →
// destroyTrain, the crashed segment update) waits for a watched session —
// it changes how trains run, the title's included.

// mooreNeighbourhood is kMooreNeighbourhood (Vehicle.h), in its order.
var mooreNeighbourhood = [9][2]int32{{0, 0}, {0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}}

// ignoreSelfCollision: the candidate is one of the next 8 components after
// the source along nextCarId.
func (m *Map) ignoreSelfCollision(src, cand int) bool {
	id := m.Entities[src].NextCarID
	for range 8 {
		if id == 0xFFFF {
			return false
		}
		e := m.entityIndexOf(id)
		if e < 0 {
			return false
		}
		if e == cand {
			return true
		}
		id = m.Entities[e].NextCarID
	}
	return false
}

// CheckForCollisions is checkForCollisions: for a rail bogie at (x, y, z),
// the first rail vehicle part on the 3×3 tiles round it (in
// kMooreNeighbourhood's order) within 16 in z and Manhattan distance under
// 12 that has a body (body or bogie), the same owner, and is either another
// train's or — its own train's — not one of the 8 parts after it (nor it
// after the candidate). Its entity id, 0xFFFF none.
// PARTIAL: within a tile goloco walks its entity slice, upstream the tile's
// entity list (its own order).
func (m *Map) CheckForCollisions(bogie int, x, y, z int32) uint16 {
	b := &m.Entities[bogie]
	if b.Mode != 0 { // TransportMode::rail
		return 0xFFFF
	}
	tx, ty := x>>5, y>>5
	for _, n := range mooreNeighbourhood {
		for _, i := range m.vehiclesOnTile((tx+n[0])<<5, (ty+n[1])<<5) {
			if i == bogie {
				continue
			}
			v := &m.Entities[i]
			if v.Mode != 0 || abs32(z-int32(v.PosZ)) > 16 {
				continue
			}
			if abs32(int32(v.PosX)-x)+abs32(int32(v.PosY)-y) >= 12 {
				continue
			}
			if v.EntityType != scenario.VehicleTypeBodyContinued && v.EntityType != scenario.VehicleTypeBodyStart && v.EntityType != scenario.VehicleTypeBogie {
				continue
			}
			if v.Owner != b.Owner {
				continue
			}
			if v.HeadID != b.HeadID {
				return v.ID
			}
			if m.ignoreSelfCollision(bogie, i) || m.ignoreSelfCollision(i, bogie) {
				continue
			}
			return v.ID
		}
	}
	return 0xFFFF
}
