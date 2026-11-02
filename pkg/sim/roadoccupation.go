package sim

import (
	"math/bits"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// getRoadOccupation (VehicleHead.cpp 0x0047D5D6): what occupies the lane a
// road vehicle on (pos, tad) would use.

// RoadOccupationFlags.
const (
	RoadLaneOccupied        uint8 = 1 << 0
	RoadLevelCrossingClosed uint8 = 1 << 1
	RoadHasLevelCrossing    uint8 = 1 << 2
	RoadHasStation          uint8 = 1 << 3
	RoadIsOneWay            uint8 = 1 << 4
)

const roadFlagIsOneWay uint16 = 1 << 0 // RoadObjectFlags::isOneWay

// k4F7338 lanes (VehicleHead.cpp k4F7338Lane0 / Lane1), by road id;
// right-hand traffic swaps them (kRightHand4F7338).
var roadLaneMasks = [2][10]uint8{
	{65, 64, 69, 64, 64, 65, 65, 65, 65, 80},
	{20, 21, 16, 16, 16, 20, 20, 20, 20, 80},
}

func (m *Map) roadUnkThing() [2][10]uint8 {
	if m.TrafficHandedness != 0 {
		return [2][10]uint8{roadLaneMasks[1], roadLaneMasks[0]}
	}
	return roadLaneMasks
}

func roadIsOneWay(om *objects.ObjectManager, roadObjectID uint8) bool {
	rd := om.GetRoadObjectByIndex(int(roadObjectID))
	return rd != nil && rd.Flags&roadFlagIsOneWay != 0
}

// RoadOccupation is getRoadOccupation.
func (m *Map) RoadOccupation(x, y, z int32, tad uint16, om *objects.ObjectManager) uint8 {
	id := (tad >> 3) & 0xF
	if id >= 10 {
		return 0
	}
	overtaking := tad&RaDIsOvertaking != 0
	reversed := tad&4 != 0
	if RoadReverseLane[id] != 1 {
		backToFront := overtaking
		sx, sy, sz := x, y, z
		if reversed {
			c := RoadCoordinates[tad&0x7F]
			sx, sy, sz = sx+int32(c.DX), sy+int32(c.DY), sz+int32(c.DZ)
			if c.RotEnd < 12 {
				sx -= int32(RotationOffset[c.RotEnd][0])
				sy -= int32(RotationOffset[c.RotEnd][1])
			}
			backToFront = !backToFront
		}
		rot := uint8(tad & 3)
		var compatible *scenario.RoadElement
		for _, p := range RoadPieces[id] {
			rx, ry := rotOffset(p.X, p.Y, rot)
			cx, cy, cz := sx+int32(rx), sy+int32(ry), sz+int32(p.Z)
			t := m.TileAt(int(cx>>5), int(cy>>5))
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindRoad {
					continue
				}
				el := &t.Roads[r.Index]
				if d := int32(el.BaseZ) - cz/4; d < -4 || d > 4 {
					continue
				}
				isCompatible := el.Rotation == rot || el.SequenceIndex == p.Index
				overlap := !isCompatible || uint16(el.RoadID) != id
				if isCompatible {
					compatible = el
				}
				lane := uint8(0b01)
				if backToFront != overlap {
					lane = 0b10
				}
				if el.LaneOccupation&lane == 0 {
					continue
				}
				if roadIsOneWay(om, el.RoadObjectID) {
					return RoadIsOneWay | RoadLaneOccupied
				}
				return RoadLaneOccupied
			}
		}
		if compatible != nil && roadIsOneWay(om, compatible.RoadObjectID) {
			return RoadIsOneWay
		}
		return 0
	}
	t := m.TileAt(int(x>>5), int(y>>5))
	if t == nil {
		return 0
	}
	things := m.roadUnkThing()
	side := 0
	if overtaking != reversed {
		side = 1
	}
	dh := bits.RotateLeft8(things[side][id], int(tad&3)*2)
	var res uint8
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad {
			continue
		}
		el := &t.Roads[r.Index]
		if d := int32(el.BaseZ) - z/4; d < -4 || d > 4 {
			continue
		}
		if el.HasStationElement {
			res |= RoadHasStation
		}
		for lane := range 2 {
			if el.RoadID >= 10 || bits.RotateLeft8(things[lane][el.RoadID], int(el.Rotation)*2)&dh == 0 {
				continue
			}
			if el.HasLevelCrossing {
				res |= RoadHasLevelCrossing
				if el.Unk7_10 {
					res |= RoadLevelCrossingClosed
				}
			}
			if el.LaneOccupation&(1<<lane) != 0 {
				res |= RoadLaneOccupied
			}
			if roadIsOneWay(om, el.RoadObjectID) {
				res |= RoadIsOneWay
			}
		}
	}
	return res
}
