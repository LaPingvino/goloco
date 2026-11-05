package sim

import (
	"github.com/LaPingvino/goloco/pkg/coords"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Putting a vehicle on the track: VehicleManager::placeDownVehicle
// (0x004B05E4) with moveComponentToSubPosition.

// PlaceDownResult is VehicleManager::PlaceDownResult.
type PlaceDownResult uint8

const (
	PlaceDownOkay PlaceDownResult = iota // Okay
	PlaceDownUnk0                        // not_enough_space_or_vehicle_in_the_way
	PlaceDownUnk1                        // vehicle_approaching_or_in_the_way
)

const flags48Flag2 uint8 = 1 << 2 // Flags48::flag2

// PlaceDownVehicle is placeDownVehicle: the vehicle (off the map) put with
// its front at (x, y, baseZ) on the piece trackAndDirection at
// initialSubPosition — placed reversed from the piece's far end, pushed
// forward onto it (positionVehicleOnTrack), then turned round (sub_4ADB47) —
// stopped. Refused when the block (or road lane) is taken; Unk0 when the
// train does not fit (then lifted up again).
func (m *Map) PlaceDownVehicle(headID uint16, x, y int32, baseZ uint8, trackAndDirection, initialSubPosition uint16, om *objects.ObjectManager) PlaceDownResult {
	tr, err := m.TrainOf(headID)
	if err != nil {
		return PlaceDownUnk0
	}
	head := &m.Entities[tr.Head]
	if head.TileX != -1 {
		return PlaceDownOkay
	}
	z := int32(baseZ) * 4
	rightHand := m.TrafficHandedness != 0
	var sub uint16
	rx, ry, rz := x, y, z
	rtad := trackAndDirection
	if head.Mode != 1 {
		if m.Sub4A2A58(x, y, z, trackAndDirection, head.Owner, head.TrackType)&1 != 0 {
			return PlaceDownUnk1
		}
		if m.IsBlockOccupied(x, y, z, trackAndDirection, head.Owner, head.TrackType) {
			return PlaceDownUnk1
		}
		sub = uint16(len(TrackSubPosition(trackAndDirection))) - 1 - initialSubPosition
		rx, ry, rz = trackEndTile(x, y, z, trackAndDirection)
		rtad ^= 1 << 2 // setReversed(!isReversed)
	} else {
		sub = uint16(len(RoadSubPosition(trackAndDirection, rightHand))) - 1 - initialSubPosition
		c := RoadCoordinates[trackAndDirection&0x7F] // getUnkRoad(basicRad)
		rx, ry, rz = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
		if c.RotEnd < 12 {
			rx -= int32(RotationOffset[c.RotEnd][0])
			ry -= int32(RotationOffset[c.RotEnd][1])
		}
		rtad ^= 1 << 2
		rtad ^= 1 << 7
		if rtad&(1<<8) != 0 { // isChangingLane
			rtad ^= 1 << 7
			if rtad&(1<<7) == 0 { // !isOvertaking
				rtad ^= 1 << 8
			}
		}
	}
	h := RoutingHandle(head.RoutingHandle)
	m.ResetRoutings(h)
	m.SetRouting(h, rtad)
	if head.Mode == 1 {
		if m.RoadOccupation(rx, ry, rz, rtad, om)&(RoadLaneOccupied|RoadLevelCrossingClosed) != 0 {
			return PlaceDownUnk1
		}
	} else {
		if m.Sub4A2A58(rx, ry, rz, rtad, head.Owner, head.TrackType)&1 != 0 {
			return PlaceDownUnk1
		}
		if m.IsBlockOccupied(rx, ry, rz, rtad, head.Owner, head.TrackType) {
			return PlaceDownUnk1
		}
	}
	// 0x004B077E
	for _, i := range trainComponents(tr) {
		e := &m.Entities[i]
		e.RoutingHandle = head.RoutingHandle
		e.TileX, e.TileY, e.TileBaseZ = coords.WorldUnits(rx), coords.WorldUnits(ry), uint8(rz/4)
		e.SubPosition = sub
		e.TrackAndDirection = rtad
		e.RemainingDistance = 0
		m.moveComponentToSubPosition(e, rightHand)
	}
	head = &m.Entities[tr.Head]
	head.Var3C = 0
	m.Entities[tr.Veh1].Var3C = 0
	m.ApplyVehicleObjectLength(tr, om)
	old := head.HeadVar52
	head.HeadVar52 = 1
	failure := m.PositionVehicleOnTrack(tr, true, om)
	head = &m.Entities[tr.Head]
	head.HeadVar52 = old
	if failure {
		m.LiftUpVehicle(headID, om)
		return PlaceDownUnk0
	}
	head.HeadVar52 = 1
	m.Sub4ADB47(tr, true, om)
	tr, _ = m.TrainOf(headID)
	head = &m.Entities[tr.Head]
	head.HeadVar52 = old
	head.Status = StatusStopped
	m.Entities[tr.Veh1].Var48 |= flags48Flag2
	return PlaceDownOkay
}

// moveComponentToSubPosition is VehicleManager.cpp's: the component moved to
// its sub-position on its piece, facing along it.
func (m *Map) moveComponentToSubPosition(e *scenario.VehicleEntity, rightHand bool) {
	var path []MoveInfo
	if e.Mode == 1 {
		path = RoadSubPosition(e.TrackAndDirection, rightHand)
	} else {
		path = TrackSubPosition(e.TrackAndDirection)
	}
	if int(e.SubPosition) >= len(path) {
		return
	}
	mi := path[e.SubPosition]
	e.SpriteYaw, e.SpritePitch = mi.Yaw, mi.Pitch
	m.moveTo(e, int32(e.TileX)+int32(mi.X), int32(e.TileY)+int32(mi.Y), int32(e.TileBaseZ)*4+int32(mi.Z))
}
