package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Track and road connection ends — Map/Track/Track.cpp getTrackConnectionEnd /
// getRoadConnectionEnd over TrackData::getUnkTrack / getUnkRoad
// (TrackCoordinates / RoadCoordinates, track_data_gen.go / road_data_gen.go).

// ConnectionEnd is Track::ConnectionEnd: the position (world units) of the
// tile the piece's end leads into, and the rotation it arrives with.
type ConnectionEnd struct {
	X, Y, Z      int32
	NextRotation uint8
}

func connectionEnd(c TrackCoord, x, y, z int32) ConnectionEnd {
	return ConnectionEnd{x + int32(c.DX), y + int32(c.DY), z + int32(c.DZ), c.RotEnd}
}

// TrackConnectionEnd is getTrackConnectionEnd(pos, trackAndDirection).
func TrackConnectionEnd(x, y, z int32, trackAndDirection uint16) (ConnectionEnd, bool) {
	if int(trackAndDirection) >= len(TrackCoordinates) {
		return ConnectionEnd{}, false
	}
	return connectionEnd(TrackCoordinates[trackAndDirection], x, y, z), true
}

// RoadConnectionEnd is getRoadConnectionEnd(pos, trackAndDirection).
func RoadConnectionEnd(x, y, z int32, trackAndDirection uint16) (ConnectionEnd, bool) {
	if int(trackAndDirection) >= len(RoadCoordinates) {
		return ConnectionEnd{}, false
	}
	return connectionEnd(RoadCoordinates[trackAndDirection], x, y, z), true
}

// AdditionalTaDFlags (Map/Track/Track.h).
const (
	TaDBasicMask  uint16 = 0b0000_0001_1111_1111
	TaDBridgeMask uint16 = 0b0000_1110_0000_0000
	TaDHasBridge  uint16 = 1 << 12
	TaDHasMods    uint16 = 1 << 13
	TaDHasSignal  uint16 = 1 << 15
)

// TrackConnections is Track::TrackConnections; StationID is -1 for
// StationId::null.
type TrackConnections struct {
	Connections      []uint16
	HasLevelCrossing bool
	StationID        int
}

// nextElement is TileElementEntry::next(): the element stored after the
// order-th one, or Kind 0xFF at the tile's end.
func (t *Tile) nextElement(order int) scenario.ElemRef {
	if order+1 < len(t.Order) {
		return t.Order[order+1]
	}
	return scenario.ElemRef{Kind: 0xFF}
}

// connectionFlags adds getElTrackConnection's bridge, mods and signal flags.
func (t *Tile) connectionFlags(order int, el *scenario.TrackElement, tad uint16, queryMods uint8) uint16 {
	if el.HasBridge {
		tad |= uint16(el.BridgeID)<<9 | TaDHasBridge
	}
	if queryMods&el.Mods != 0 {
		tad |= TaDHasMods
	}
	if el.HasSignal {
		if n := t.nextElement(order); n.Kind == scenario.KindSignal {
			if sg := &t.Signals[n.Index]; !sg.AiAllocated && !sg.Ghost {
				tad |= TaDHasSignal
			}
		}
	}
	return tad
}

const nullTaD uint16 = 0xFFFF

// elTrackConnection is getElTrackConnection: el connects if it begins here
// (sequence 0) with nextRotation at baseZ, or — reversed (bit 2) — if it
// ends here (flag_6) with nextRotation at baseZ.
func (t *Tile) elTrackConnection(order int, el *scenario.TrackElement, nextRotation, baseZ, queryMods uint8) uint16 {
	if int(el.TrackID) >= len(TrackPieceTiles) {
		return nullTaD
	}
	piece := TrackPieceTiles[el.TrackID]
	if el.SequenceIndex == 0 {
		tad := uint16(el.TrackID)<<3 | uint16(el.Rotation)
		if nextRotation == TrackCoordinates[tad].RotBegin && int(baseZ) == int(el.BaseZ)-int(piece[0][2])/4 {
			return t.connectionFlags(order, el, tad, queryMods)
		}
	}
	if !el.IsLast {
		return nullTaD
	}
	tad := uint16(el.TrackID)<<3 | 1<<2 | uint16(el.Rotation)
	if nextRotation != TrackCoordinates[tad].RotBegin || int(el.SequenceIndex) >= len(piece) {
		return nullTaD
	}
	if int(baseZ) != int(el.BaseZ)-(int(piece[el.SequenceIndex][2])+int(TrackCoordinates[tad].DZ))/4 {
		return nullTaD
	}
	return t.connectionFlags(order, el, tad, queryMods)
}

// TrackConnections is getTrackConnections (0x004A2638): the track pieces on
// the tile at (x, y, z) (world units) that a piece arriving with
// nextRotation continues onto, for this company, track object and mods.
func (m *Map) TrackConnections(x, y, z int32, nextRotation, company, trackObjectID, requiredMods, queryMods uint8) TrackConnections {
	res := TrackConnections{StationID: -1}
	baseZ := uint8(z / 4)
	t := m.TileAt(int(x/32), int(y/32))
	if t == nil || x < 0 || y < 0 {
		return res
	}
	for order, r := range t.Order {
		if r.Kind != scenario.KindTrack {
			continue
		}
		el := &t.Tracks[r.Index]
		if el.Owner != company || el.TrackObjectID != trackObjectID || el.Mods&requiredMods != requiredMods || el.Ghost || el.AiAllocated {
			continue
		}
		c := t.elTrackConnection(order, el, nextRotation, baseZ, queryMods)
		if c == nullTaD {
			continue
		}
		res.Connections = append(res.Connections, c)
		if el.HasStation {
			if n := t.nextElement(order); n.Kind == scenario.KindStation {
				if st := &t.Stations[n.Index]; !st.AiAllocated && !st.Ghost {
					res.StationID = int(st.StationID)
				}
			}
		}
		if el.HasLevelCrossing {
			res.HasLevelCrossing = true
		}
	}
	return res
}
