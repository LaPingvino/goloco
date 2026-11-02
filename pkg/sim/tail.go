package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// The tail's work behind the train — a port of VehicleTail.cpp
// (leaveLevelCrossing 0x004794BC, VehicleTail::update 0x004AA24A) and
// VehicleBase::updateRoadTileOccupancy.
//
// PARTIAL: leaveLevelCrossing's level crossing animation
// (AnimationManager::createAnimation) is not ported; it only redraws.

// LevelCrossingEnter / LevelCrossingLeave are leaveLevelCrossing's unk.
const (
	LevelCrossingEnter = 8
	LevelCrossingLeave = 9
)

// LeaveLevelCrossing is leaveLevelCrossing: the straight road (road id 0)
// crossing the piece's first tile is marked in use (unk 8) or free.
func (m *Map) LeaveLevelCrossing(x, y, z int32, tad uint16, unk int) {
	if tad&4 != 0 {
		x, y, z = trackEndTile(x, y, z, tad)
	}
	p := TrackPieces[tad>>3&0x3F][0]
	x, y, z = pieceTileLoc(x, y, z, tad, p)
	t := m.TileAt(int(x>>5), int(y>>5))
	if t == nil {
		return
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad {
			continue
		}
		el := &t.Roads[r.Index]
		if int32(el.BaseZ) != z/4 || el.RoadID != 0 {
			continue
		}
		el.Unk7_10 = unk == LevelCrossingEnter
	}
}

// UpdateRoadTileOccupancy is VehicleBase::updateRoadTileOccupancy: the lanes
// rad uses (getRoadOccupationMask(rad >> 2) >> 4) are set or cleared on the
// matching road elements within 4 SmallZ; returns the road type (an
// any-road-type vehicle, trackType 0xFF, takes the road's and marks it).
func (m *Map) UpdateRoadTileOccupancy(x, y, z int32, rad uint16, trackType uint8, setOccupied bool) uint8 {
	roadType := trackType
	t := m.TileAt(int(x>>5), int(y>>5))
	if t == nil {
		return roadType
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad {
			continue
		}
		el := &t.Roads[r.Index]
		d := int32(el.BaseZ) - z/4
		if d < -4 || d > 4 || uint16(el.Rotation) != rad&3 || uint16(el.RoadID) != rad>>3&0xF || el.AiAllocated {
			continue
		}
		lanes := RoadOccupationMasks[(rad>>2)&0x7F] >> 4
		if setOccupied {
			el.LaneOccupation = (el.LaneOccupation | lanes) & 3
		} else {
			el.LaneOccupation &^= lanes
		}
		if trackType == 0xFF {
			if m.RoadAnyTypeCompatible&(1<<el.RoadObjectID) != 0 {
				el.Mods |= 1 // setUnk7_40: the same bit as mods bit 0
				roadType = el.RoadObjectID
			}
		} else {
			roadType = trackType
		}
	}
	return roadType
}

// TailUpdate is VehicleTail::update for land vehicles: the tail moves by
// distance; once it has left a piece its routing is freed and the piece
// released — road lanes cleared, or for rail the signal behind it freed,
// the block it left re-evaluated (sub_4A2AD7) and a level crossing left.
// It returns false where upstream destroys the train (noRouteFound).
func (m *Map) TailUpdate(c *MotionComponent, distance int32, requiredMods uint8) bool {
	oldHandle := c.RoutingHandle
	ox, oy, oz := c.TileX, c.TileY, int32(c.TileBaseZ)*4
	next := m.TrackNewPiece
	if c.Road {
		next = m.RoadNewPiece
	}
	r := c.UpdateMotion(distance, func(c *MotionComponent, f *uint32) bool {
		return next(c, f, requiredMods, false, 0)
	})
	if r.Flags&MotionNoRouteFound != 0 {
		return false
	}
	if oldHandle == c.RoutingHandle {
		return true
	}
	ref := m.Routing(oldHandle)
	tad := ref & 0x1FF // TrackAndDirection((ref & 0x1F8) >> 3, ref & 7)
	m.FreeRouting(oldHandle)
	if c.Road {
		m.UpdateRoadTileOccupancy(ox, oy, oz, tad, c.TrackType, false)
		return true
	}
	if ref&TaDHasSignal != 0 {
		m.SetSignalState(ox, oy, oz, tad, c.TrackType, 0)
	}
	nx, ny, nz := trackEndTile(ox, oy, oz, ref&TaDBasicMask)
	m.UpdateBlockSignals(nx, ny, nz, tad^4, c.Owner, c.TrackType)
	m.LeaveLevelCrossing(ox, oy, oz, tad, LevelCrossingLeave)
	return true
}
