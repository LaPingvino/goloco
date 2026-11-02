package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Helpers of the rail reservation step (sub_4ACEF1) — Tile::trainStation,
// VehicleHead.cpp's bringTrackElementToFront and updateJunctionSignalLights,
// and Routing.cpp's sub_4A2A77 and setReverseSignalOccupiedInBlock.
//
// PARTIAL: sub_4ACEF1 itself is not ported yet.

// trainStation is Tile::trainStation: the station element after the track
// element with this track id, rotation and height that has a station.
func (t *Tile) trainStation(trackID, rotation, baseZ uint8) *scenario.StationElement {
	found := false
	for _, r := range t.Order {
		if found && r.Kind == scenario.KindStation {
			return &t.Stations[r.Index]
		}
		if r.Kind != scenario.KindTrack {
			continue
		}
		el := &t.Tracks[r.Index]
		found = false
		if el.BaseZ != baseZ || el.Rotation != rotation || el.TrackID != trackID || !el.HasStation {
			continue
		}
		found = true
	}
	return nil
}

// trackStartFromTad is getTrackStartFromTad.
func trackStartFromTad(x, y, z int32, tad uint16) (int32, int32, int32) {
	if tad&4 != 0 {
		return trackEndTile(x, y, z, tad)
	}
	return x, y, z
}

// BringTrackElementToFront is bringTrackElementToFront: on each tile of the
// chosen piece, its track element moves to the front of the run of track
// elements at that height (the order junction pieces are drawn and found
// in), unless it carries a signal or station.
func (m *Map) BringTrackElementToFront(x, y, z int32, trackType uint8, tad uint16) {
	sx, sy, sz := trackStartFromTad(x, y, z, tad)
	id, rot := uint8(tad>>3&0x3F), uint8(tad&3)
	if int(id) >= len(TrackPieces) {
		return
	}
	for _, p := range TrackPieces[id] {
		rx, ry := rotOffset(p.X, p.Y, rot)
		px, py, pz := sx+int32(rx), sy+int32(ry), sz+int32(p.Z)
		t := m.TileAt(int(px>>5), int(py>>5))
		if t == nil {
			continue
		}
		begin, last := -1, -1
		for oi, r := range t.Order {
			if r.Kind != scenario.KindTrack {
				begin = -1
				continue
			}
			el := &t.Tracks[r.Index]
			if int32(el.BaseZ)*4 != pz {
				begin = -1
				continue
			}
			if begin < 0 {
				begin = oi
			}
			if el.Rotation != rot || el.SequenceIndex != p.Index || el.TrackObjectID != trackType || el.TrackID != id {
				continue
			}
			if !el.HasSignal && !el.HasStation {
				last = oi
			}
			break
		}
		if last < 0 || begin < 0 || begin == last {
			continue
		}
		ref := t.Order[last]
		copy(t.Order[begin+1:last+1], t.Order[begin:last])
		t.Order[begin] = ref
	}
}

// UpdateJunctionSignalLights is updateJunctionSignalLights: the signal before
// a junction shows which way the route goes (left / right of the straightest
// connection), set through setSignalState's lights mode (16).
func (m *Map) UpdateJunctionSignalLights(x, y, z int32, targetRouting uint16, trackType uint8, newRouting uint16, conns []uint16) {
	if len(conns) == 0 {
		return
	}
	curv := func(r uint16) int8 { return CurvatureDegrees[((r&TaDBasicMask)>>2)%88] }
	minC, maxC, least := curv(conns[0]), curv(conns[0]), curv(conns[0])
	abs8 := func(v int8) int {
		if v < 0 {
			return -int(v)
		}
		return int(v)
	}
	for _, c := range conns[1:] {
		v := curv(c)
		minC, maxC = min(minC, v), max(maxC, v)
		if abs8(v) < abs8(least) {
			least = v
		}
	}
	nc := curv(newRouting)
	flags := uint32(0x10)
	if least != minC {
		flags |= 1 << 30
		if nc < least {
			flags |= 1 << 28
		}
	}
	if least != maxC {
		flags |= 1 << 29
		if nc > least {
			flags |= 1 << 27
		}
	}
	m.SetSignalState(x, y, z, targetRouting&TaDBasicMask, trackType, flags)
}

// Sub4A2A77 is sub_4A2A77: searching forwards only from (pos, tad) to the
// next signals — bit 0: one of them is clear or occupied one-way; bit 1: the
// line dead-ends.
func (m *Map) Sub4A2A77(x, y, z int32, tad uint16, company, trackType uint8) uint8 {
	var data uint16
	res := &RoutingResults{Reachable: NewInterestSet(SignalHashSetSize)}
	m.FindAllTracksFilterTransform(res, SearchExcludeReverseDirection, x, y, z, tad, company, trackType,
		func(in LocationOfInterest) bool {
			if in.TrackAndDirection&TaDHasSignal == 0 {
				return false
			}
			s := m.SignalState(in.X, in.Y, in.Z, in.TaD(), in.TrackType, 0)
			clear := s&SignalBlockedNoRoute == 0 && s&SignalOccupied == 0
			if s&SignalOccupiedOneWay != 0 || clear {
				data |= 1
			}
			return true
		}, nil)
	if res.HasDeadEnd {
		data |= 2
	}
	return uint8(data)
}

// SetReverseSignalOccupiedInBlock is setReverseSignalOccupiedInBlock: every
// signal bounding the block (searched with unk2) set occupied on its right
// side (1<<31 | 8).
func (m *Map) SetReverseSignalOccupiedInBlock(x, y, z int32, tad uint16, company, trackType uint8) {
	res := &RoutingResults{Reachable: NewInterestSet(SignalHashSetSize)}
	m.FindAllTracksFilterTransform(res, SearchUnk2, x, y, z, tad, company, trackType,
		func(in LocationOfInterest) bool {
			if in.TrackAndDirection&TaDHasSignal == 0 {
				return false
			}
			m.SetSignalState(in.X, in.Y, in.Z, in.TaD(), in.TrackType, signalRightSide|8)
			return true
		}, nil)
}
