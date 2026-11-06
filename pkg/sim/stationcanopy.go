package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// sub_48D794 (World/Station.cpp): each train-station tile's sequence index —
// 0, or 1 for a middle tile (the platform piece that continues both ways)
// when the station isn't covered by something above it. Item ST: run after
// a train station is built or removed.

const bridgeObjFlagHasRoof = 1 << 0 // BridgeObjectFlags::hasRoof

// forEachStationElement is Station.cpp's: for each tile of the track piece
// under the station at (x, y, z), each station element at the piece's height
// on its track tile of that sequence index and rotation — with whether the
// surface came before it on the tile.
func (m *Map) forEachStationElement(x, y, z int32, rotation uint8, trackID uint8, fn func(t *Tile, k int, px, py, pz int32, passedSurface bool)) {
	if int(trackID) >= len(TrackPieces) {
		return
	}
	for _, piece := range TrackPieces[trackID] {
		ox, oy := rotOffset(piece.X, piece.Y, rotation)
		px, py, pz := x+int32(ox), y+int32(oy), z+int32(piece.Z)
		t := m.TileAt(int(px)>>5, int(py)>>5)
		if t == nil {
			continue
		}
		passed := false
		for k, r := range t.Order {
			if r.Kind == scenario.KindSurface {
				passed = true
				continue
			}
			if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) || int32(t.Stations[r.Index].BaseZ)*4 != pz {
				continue
			}
			if k == 0 || t.Order[k-1].Kind != scenario.KindTrack {
				continue
			}
			tr := &t.Tracks[t.Order[k-1].Index]
			if tr.SequenceIndex != piece.Index || tr.Rotation != rotation {
				continue
			}
			fn(t, k, px, py, pz, passed)
		}
	}
}

// isStationTrackConnected is Station.cpp's: a track piece of the owner with
// this station on it (not ghost or AI) that a train leaving with
// connectRotation enters at (x, y, z) — at its start, or at its end
// (flag_6, reversed).
func (m *Map) isStationTrackConnected(x, y, z int32, connectRotation, owner uint8, stationID uint16) bool {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return false
	}
	for k, r := range t.Order {
		if r.Kind != scenario.KindTrack || int(r.Index) >= len(t.Tracks) {
			continue
		}
		tr := &t.Tracks[r.Index]
		if !tr.HasStation || tr.Owner != owner || k+1 >= len(t.Order) {
			continue
		}
		nr := t.Order[k+1]
		if nr.Kind != scenario.KindStation || int(nr.Index) >= len(t.Stations) {
			continue
		}
		st := &t.Stations[nr.Index]
		if st.AiAllocated || st.Ghost || st.StationID != stationID {
			continue
		}
		tad := uint16(tr.TrackID)<<3 | uint16(tr.Rotation)
		if tr.SequenceIndex == 0 && connectRotation == TrackCoordinates[tad].RotBegin {
			if int32(tr.BaseZ)*4-int32(TrackPieces[tr.TrackID][0].Z) == z {
				return true
			}
		}
		if tr.IsLast {
			tad |= 1 << 2
			if connectRotation == TrackCoordinates[tad].RotBegin {
				p := TrackPieces[tr.TrackID][tr.SequenceIndex]
				if int32(tr.BaseZ)*4-(int32(p.Z)+int32(TrackCoordinates[tad].DZ)) == z {
					return true
				}
			}
		}
	}
	return false
}

// UpdateStationCanopies is sub_48D794: for each station tile on a piece's
// first tile, every station element of the piece gets sequence index 0 and
// counts as covered when it is under ground (before the surface), on a
// roofed bridge, or with something not at its height starting below its
// clear height above it; uncovered, with var_0B ≥ 2 and the station's
// track continuing both ways into this station, they get index 1.
func (m *Map) UpdateStationCanopies(id int, om *objects.ObjectManager) {
	st := &m.StationNames[id]
	for i := 0; i < int(st.TileSize) && i < len(st.Tiles); i++ {
		p := st.Tiles[i]
		rotation := uint8(p[2] & 3)
		x, y, z := int32(p[0]), int32(p[1]), int32(p[2]&^3)
		t := m.TileAt(int(x)>>5, int(y)>>5)
		if t == nil {
			continue
		}
		k := -1
		for j, r := range t.Order {
			if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) || int32(t.Stations[r.Index].BaseZ)*4 != z {
				continue
			}
			if t.Stations[r.Index].StationType != 0 {
				break
			}
			if j == 0 || t.Order[j-1].Kind != scenario.KindTrack || t.Tracks[t.Order[j-1].Index].SequenceIndex != 0 {
				continue
			}
			k = j
			break
		}
		if k < 0 {
			continue
		}
		el := &t.Stations[t.Order[k].Index]
		tr := &t.Tracks[t.Order[k-1].Index]
		var obj *objects.TrainStationObject
		if om != nil && int(el.ObjectID) < len(om.TrainStationObjects) {
			obj = om.TrainStationObjects[el.ObjectID]
		}
		covered := false
		m.forEachStationElement(x, y, z, rotation, tr.TrackID, func(tt *Tile, kk int, _, _, _ int32, passed bool) {
			se := &tt.Stations[tt.Order[kk].Index]
			se.SequenceIndex = 0
			if !passed {
				covered = true
				return
			}
			under := &tt.Tracks[tt.Order[kk-1].Index]
			if under.HasBridge && om != nil && int(under.BridgeID) < len(om.BridgeObjects) && om.BridgeObjects[under.BridgeID] != nil &&
				om.BridgeObjects[under.BridgeID].Flags&bridgeObjFlagHasRoof != 0 {
				covered = true
				return
			}
			last := len(tt.Order) - 1
			if kk == last {
				return
			}
			for j := kk + 1; ; j++ { // do … while (!el->isLast()): the tile's last element isn't checked
				bz := tt.baseZOf(tt.Order[j])
				if bz != se.BaseZ && int(bz)-4 < int(se.ClearZ) {
					covered = true
					return
				}
				if j+1 >= last {
					return
				}
			}
		})
		if covered || obj == nil || obj.Var0B == 0 || obj.Var0B == 1 {
			continue
		}
		tad := uint16(tr.TrackID)<<3 | uint16(tr.Rotation)
		c := TrackCoordinates[tad]
		if !m.isStationTrackConnected(x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ), c.RotEnd, tr.Owner, el.StationID) {
			continue
		}
		bx, by := x, y
		back := ReverseRotation[c.RotBegin]
		if back < 12 {
			bx += int32(RotationOffset[back][0])
			by += int32(RotationOffset[back][1])
		}
		if !m.isStationTrackConnected(bx, by, z, back, tr.Owner, el.StationID) {
			continue
		}
		m.forEachStationElement(x, y, z, rotation, tr.TrackID, func(tt *Tile, kk int, _, _, _ int32, _ bool) {
			tt.Stations[tt.Order[kk].Index].SequenceIndex = 1
		})
	}
}
