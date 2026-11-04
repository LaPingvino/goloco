package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Signal state and block occupancy — a port of Vehicles/Routing.cpp's
// findSignalOnTrack, getSignalState, setSignalState, findOccupationByBlock,
// setSignalsOccupiedState, sub_4A2AD7 and isBlockOccupied.
//
// PARTIAL: setSignalState's viewport invalidation and its signal animation
// (AnimationManager::createAnimation) are not ported; they only redraw.

// SignalStateFlags.
const (
	SignalOccupied       uint8 = 1 << 0
	SignalBlockedNoRoute uint8 = 1 << 1
	SignalOccupiedOneWay uint8 = 1 << 2
)

const signalRightSide = uint32(1) << 31

// findSignalOnTrack: the signal stored after the track element at loc with
// this piece, rotation, sequence index and track object.
func (m *Map) findSignalOnTrack(x, y, z int32, tad uint16, trackType, index uint8) (*scenario.SignalElement, *scenario.TrackElement) {
	t := m.TileAt(int(x>>5), int(y>>5))
	if t == nil {
		return nil, nil
	}
	for order, r := range t.Order {
		if r.Kind != scenario.KindTrack {
			continue
		}
		el := &t.Tracks[r.Index]
		if int32(el.BaseZ) != z/4 || !el.HasSignal || uint16(el.Rotation) != tad&3 || el.SequenceIndex != index ||
			el.TrackObjectID != trackType || uint16(el.TrackID) != tad>>3&0x3F {
			continue
		}
		n := t.nextElement(order)
		if n.Kind != scenario.KindSignal {
			return nil, el
		}
		return &t.Signals[n.Index], el
	}
	return nil, nil
}

// signalStart is where a piece's signals are looked up from: its start, or
// for a reversed piece its far end (and the side flips).
func signalStart(x, y, z int32, tad uint16, flags uint32) (int32, int32, int32, uint32) {
	if tad&4 != 0 {
		x, y, z = trackEndTile(x, y, z, tad)
		flags ^= signalRightSide
	}
	return x, y, z, flags
}

func pieceTileLoc(x, y, z int32, tad uint16, p PreviewTrack) (int32, int32, int32) {
	rx, ry := rotOffset(p.X, p.Y, uint8(tad&3))
	return x + int32(rx), y + int32(ry), z + int32(p.Z)
}

// SignalState is getSignalState (0x0048963F for flags 0xXXXX_XXXA).
func (m *Map) SignalState(x, y, z int32, tad uint16, trackType uint8, flags uint32) uint8 {
	x, y, z, flags = signalStart(x, y, z, tad, flags)
	pieces := TrackPieces[tad>>3&0x3F]
	sx, sy, sz := pieceTileLoc(x, y, z, tad, pieces[0])
	sig, _ := m.findSignalOnTrack(sx, sy, sz, tad, trackType, pieces[0].Index)
	if sig == nil {
		return 0
	}
	right := flags&signalRightSide != 0
	side, other := sig.Sides[0], sig.Sides[1]
	if right {
		side, other = other, side
	}
	var ret uint8
	if side.IsOccupied() {
		ret |= SignalOccupied
	}
	if !side.HasSignal() {
		ret |= SignalBlockedNoRoute
	}
	sideGhost, otherGhost := sig.LeftGhost, sig.RightGhost
	if right {
		sideGhost, otherGhost = otherGhost, sideGhost
	}
	if !other.HasSignal() || otherGhost {
		ret |= SignalOccupiedOneWay
	}
	if sideGhost && other.HasSignal() {
		ret |= SignalBlockedNoRoute
	}
	return ret
}

// SetSignalState is setSignalState (0x0048963F): flags & 0xFFFF selects
// the change — 16: set lights from bits 27–30; 8: occupied; >8: free;
// otherwise the state (unk4) — on every signal of the piece, on the side
// bit 31 names (flipped for a reversed piece).
func (m *Map) SetSignalState(x, y, z int32, tad uint16, trackType uint8, flags uint32) {
	unk1 := flags & 0xFFFF
	x, y, z, flags = signalStart(x, y, z, tad, flags)
	for _, p := range TrackPieces[tad>>3&0x3F] {
		sx, sy, sz := pieceTileLoc(x, y, z, tad, p)
		sig, foundTrack := m.findSignalOnTrack(sx, sy, sz, tad, trackType, p.Index)
		if sig == nil {
			return // upstream asserts: a signalled piece has a signal on every tile
		}
		side := &sig.Sides[0]
		if flags&signalRightSide != 0 {
			side = &sig.Sides[1]
		}
		switch {
		case unk1 == 16:
			var lights uint8
			if flags&(1<<30) != 0 {
				lights |= 1 << 2
				if flags&(1<<28) != 0 {
					lights |= 1 << 0
				}
			}
			if flags&(1<<29) != 0 {
				lights |= 1 << 3
				if flags&(1<<27) != 0 {
					lights |= 1 << 1
				}
			}
			if side.AllLights() != lights {
				side.SetAllLights(lights)
				should := foundTrack.IsLast // isFlag6
				if flags&(1<<31) != 0 {
					should = foundTrack.SequenceIndex == 0
				}
				if should {
					m.invalidateTile(sx, sy, sz, sz+24, ZoomHalf, 0)
				}
			}
		case unk1 == 8:
			side.SetIsOccupied(true)
		case unk1 > 8:
			side.SetIsOccupied(false)
		default:
			side.SetUnk4(uint8(unk1 & 3))
			if unk1&3 == 0 {
				side.SetAllLights(side.AllLights() & 0xC)
			}
			animate := foundTrack.SequenceIndex == 0
			if flags&(1<<31) != 0 {
				animate = foundTrack.IsLast // isFlag6
			}
			if animate {
				// PARTIAL: AnimationManager::createAnimation(0, signalLoc, z / 4) —
				// the animation manager is not ported (the signal's frames don't step).
				m.invalidateTile(sx, sy, sz, sz+24, ZoomHalf, 0)
			}
		}
	}
}

// vehiclesOnTile is EntityManager::EntityTileList for vehicles: indices of
// the vehicle components whose position lies on the tile.
func (m *Map) vehiclesOnTile(x, y int32) []int {
	var out []int
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType > scenario.VehicleTypeTail {
			continue
		}
		if int32(e.PosX)>>5 == x>>5 && int32(e.PosY)>>5 == y>>5 {
			out = append(out, i)
		}
	}
	return out
}

// findOccupationByBlock (0x004A2AF0): sets *occupied when a vehicle stands
// on the interest's piece (either direction); stops at signals.
func (m *Map) findOccupationByBlock(in LocationOfInterest, occupied *uint16) bool {
	tad := in.TaD()
	nx, ny, nz := trackEndTile(in.X, in.Y, in.Z, tad)
	back := tad ^ 4
	sx, sy := in.X, in.Y
	if tad&4 != 0 {
		sx, sy = nx, ny
	}
	for _, p := range TrackPieces[tad>>3&0x3F] {
		rx, ry := rotOffset(p.X, p.Y, uint8(tad&3))
		for _, i := range m.vehiclesOnTile(sx+int32(rx), sy+int32(ry)) {
			e := &m.Entities[i]
			if e.Flags38&(1<<0|1<<2) != 0 {
				continue
			}
			lx, ly, lz := int32(e.TileX), int32(e.TileY), int32(e.TileBaseZ)*4
			if lx == in.X && ly == in.Y && lz == in.Z && e.TrackAndDirection == tad {
				*occupied = 1
				break
			}
			if lx == nx && ly == ny && lz == nz && e.TrackAndDirection == back {
				*occupied = 1
				break
			}
		}
	}
	return in.TrackAndDirection&TaDHasSignal != 0
}

// setSignalsOccupiedState (0x004A2CE7): every signal bounding the block
// gets occupied (8) or free (9), on its right side.
func (m *Map) setSignalsOccupiedState(res *RoutingResults, occupied uint16) {
	for _, in := range res.Reachable.All() {
		if in.TrackAndDirection&TaDHasSignal == 0 {
			continue
		}
		f := signalRightSide | 9
		if occupied&1 != 0 {
			f = signalRightSide | 8
		}
		m.SetSignalState(in.X, in.Y, in.Z, in.TaD(), in.TrackType, f)
	}
}

// UpdateBlockSignals is sub_4A2AD7: the block from (loc, tad) is searched
// for vehicles and its bounding signals set occupied or free.
func (m *Map) UpdateBlockSignals(x, y, z int32, tad uint16, company, trackType uint8) {
	var occupied uint16
	res := &RoutingResults{Reachable: NewInterestSet(SignalHashSetSize)}
	m.FindAllTracksFilterTransform(res, SearchUnk2, x, y, z, tad, company, trackType,
		func(in LocationOfInterest) bool { return m.findOccupationByBlock(in, &occupied) },
		func(r *RoutingResults) { m.setSignalsOccupiedState(r, occupied) })
}

// IsBlockOccupied is isBlockOccupied (0x004A2A39).
func (m *Map) IsBlockOccupied(x, y, z int32, tad uint16, company, trackType uint8) bool {
	var occupied uint16
	res := &RoutingResults{Reachable: NewInterestSet(SignalHashSetSize)}
	m.FindAllTracksFilterTransform(res, SearchUnk2, x, y, z, tad, company, trackType,
		func(in LocationOfInterest) bool { return m.findOccupationByBlock(in, &occupied) }, nil)
	return occupied&1 != 0
}
