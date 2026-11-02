package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Track network search — a port of Vehicles/Routing.cpp's
// LocationOfInterestHashSet, findAllUsableTrackPieces,
// findAllUsableTrackInNetwork and findAllTracksFilterTransform (the walk
// that signals, block occupancy and track mods are built on).

// TrackNetworkSearchFlags.
const (
	SearchExcludeReverseDirection uint16 = 1 << 1
	SearchUnk2                    uint16 = 1 << 2
)

// Hash set sizes (kSignalHashSetSize, kTrackModHashSetSize).
const (
	SignalHashSetSize   = 0x400
	TrackModHashSetSize = 0x1000
)

// LocationOfInterest: a piece (trackAndDirection with the hasSignal bit) at
// a position, for a company and track object.
type LocationOfInterest struct {
	X, Y, Z           int32
	TrackAndDirection uint16
	Company           uint8
	TrackType         uint8
}

// TaD is tad(): the trackAndDirection without its flags (& 0x1FF).
func (l LocationOfInterest) TaD() uint16 { return l.TrackAndDirection & 0x1FF }

const nullCompany = 0xFF // CompanyId::null

var emptyInterest = LocationOfInterest{X: -1, Y: -1, Z: 0, Company: nullCompany}

func (l LocationOfInterest) empty() bool { return l.X == -1 && l.Y == -1 && l.Z == 0 }

// InterestSet is LocationOfInterestHashSet: open addressing, linear probing,
// at most size-100 entries (kMinFreeSlots); iteration is by slot index.
type InterestSet struct {
	locs  []LocationOfInterest
	count int
	mask  uint16
	max   int
}

func NewInterestSet(size int) *InterestSet {
	s := &InterestSet{locs: make([]LocationOfInterest, size), mask: uint16(size - 1), max: size - 100}
	for i := range s.locs {
		s.locs[i] = emptyInterest
	}
	return s
}

func (s *InterestSet) hash(l LocationOfInterest) uint16 {
	// coordinate_t is int16: the C++ arithmetic runs on promoted ints.
	x, y, z := int32(int16(l.X)), int32(int16(l.Y)), int32(int16(l.Z))
	return uint16((((x^z)/32)^y)^int32(l.TrackAndDirection)) & s.mask
}

// TryAdd is tryAdd: false if present or full.
func (s *InterestSet) TryAdd(l LocationOfInterest) bool {
	i := s.hash(l)
	for ; !s.locs[i].empty(); i = (i + 1) & s.mask {
		if s.locs[i] == l {
			return false
		}
	}
	if s.count >= s.max {
		return false
	}
	s.locs[i] = l
	s.count++
	return true
}

// All is the set in iteration (slot) order.
func (s *InterestSet) All() []LocationOfInterest {
	var out []LocationOfInterest
	for _, l := range s.locs {
		if !l.empty() {
			out = append(out, l)
		}
	}
	return out
}

// RoutingResults is RoutingResults.
type RoutingResults struct {
	Reachable  *InterestSet
	HasDeadEnd bool
}

// TrackEndTile is trackEndTile, exported for tests and callers.
func TrackEndTile(x, y, z int32, tad uint16) (int32, int32, int32) { return trackEndTile(x, y, z, tad) }

// trackEndTile is the tile a (reversed) piece's far end sits on: start +
// getUnkTrack.pos, minus kRotationOffset[rotationEnd] for non-diagonal ends.
func trackEndTile(x, y, z int32, tad uint16) (int32, int32, int32) {
	c := TrackCoordinates[tad&0x1FF]
	x, y, z = x+int32(c.DX), y+int32(c.DY), z+int32(c.DZ)
	if c.RotEnd < 12 {
		x -= int32(RotationOffset[c.RotEnd][0])
		y -= int32(RotationOffset[c.RotEnd][1])
	}
	return x, y, z
}

// findAllUsableTrackPieces (0x004A313B): with SearchUnk2, every track
// element sharing a tile edge (connectFlags) with any tile of the interest's
// piece is added, forwards and reversed.
func (m *Map) findAllUsableTrackPieces(queue *[]LocationOfInterest, flags uint16, in LocationOfInterest, filter func(LocationOfInterest) bool, res *RoutingResults) {
	if flags&SearchUnk2 == 0 {
		return
	}
	tad := in.TaD()
	id, dir := tad>>3, uint8(tad&3)
	nx, ny, nz := in.X, in.Y, in.Z
	if tad&4 != 0 {
		nx, ny, nz = trackEndTile(nx, ny, nz, tad)
	}
	if int(id) >= len(TrackPieces) {
		return
	}
	for _, piece := range TrackPieces[id] {
		connect := piece.ConnectFlags[dir]
		rx, ry := rotOffset(piece.X, piece.Y, dir)
		px, py, pz := nx+int32(rx), ny+int32(ry), nz+int32(piece.Z)
		t := m.TileAt(int(px>>5), int(py>>5))
		if t == nil {
			continue
		}
		for _, r := range t.Order {
			if r.Kind != scenario.KindTrack {
				continue
			}
			el := &t.Tracks[r.Index]
			if int32(el.BaseZ) != pz/4 || el.AiAllocated || el.Ghost {
				continue
			}
			if int(el.TrackID) >= len(TrackPieces) || int(el.SequenceIndex) >= len(TrackPieces[el.TrackID]) {
				continue
			}
			target := TrackPieces[el.TrackID][el.SequenceIndex]
			if target.ConnectFlags[el.Rotation]&connect == 0 {
				continue
			}
			if el.Rotation == dir && el.SequenceIndex == piece.Index && el.TrackObjectID == in.TrackType && uint16(el.TrackID) == id {
				continue
			}
			tx, ty := rotOffset(target.X, target.Y, el.Rotation)
			sx, sy, sz := px-int32(tx), py-int32(ty), int32(el.BaseZ)*4-int32(target.Z)
			tad2 := uint16(el.TrackID)<<3 | uint16(el.Rotation)
			fwd := LocationOfInterest{sx, sy, sz, tad2, el.Owner, el.TrackObjectID}
			if res.Reachable.TryAdd(fwd) && !filter(fwd) {
				m.findAllUsableTrackPieces(queue, flags, fwd, filter, res)
				*queue = append(*queue, fwd)
			}
			ex, ey, ez := trackEndTile(sx, sy, sz, tad2)
			rev := LocationOfInterest{ex, ey, ez, tad2 ^ 4, el.Owner, el.TrackObjectID}
			if res.Reachable.TryAdd(rev) && !filter(rev) {
				m.findAllUsableTrackPieces(queue, flags, rev, filter, res)
				*queue = append(*queue, rev)
			}
		}
	}
}

// findAllUsableTrackInNetwork (0x004A2FE6): the pieces connected at the
// interest's end, and (unless excluded) behind its start.
func (m *Map) findAllUsableTrackInNetwork(queue *[]LocationOfInterest, flags uint16, in LocationOfInterest, filter func(LocationOfInterest) bool, res *RoutingResults) {
	visit := func(x, y, z int32, conns []uint16) {
		for _, c := range conns {
			l := LocationOfInterest{x, y, z, c & (TaDBasicMask | TaDHasSignal), in.Company, in.TrackType}
			if res.Reachable.TryAdd(l) && !filter(l) {
				m.findAllUsableTrackPieces(queue, flags, l, filter, res)
				*queue = append(*queue, l)
			}
		}
	}
	e, _ := TrackConnectionEnd(in.X, in.Y, in.Z, in.TaD())
	tc := m.TrackConnections(e.X, e.Y, e.Z, e.NextRotation, in.Company, in.TrackType, 0, 0)
	if len(tc.Connections) > 0 {
		visit(e.X, e.Y, e.Z, tc.Connections)
	} else {
		res.HasDeadEnd = true
	}
	if flags&SearchExcludeReverseDirection == 0 {
		x, y, z := trackEndTile(in.X, in.Y, in.Z, in.TaD())
		rot := ReverseRotation[TrackCoordinates[in.TaD()].RotEnd]
		tc2 := m.TrackConnections(x, y, z, rot, in.Company, in.TrackType, 0, 0)
		visit(x, y, z, tc2.Connections)
	}
}

// FindAllTracksFilterTransform (0x004A2E46): a stack-driven walk from
// (loc, trackAndDirection); filter returns true where the walk stops
// (e.g. a signal ending a block); transform runs on the results.
func (m *Map) FindAllTracksFilterTransform(res *RoutingResults, flags uint16, x, y, z int32, trackAndDirection uint16, company, trackType uint8, filter func(LocationOfInterest) bool, transform func(*RoutingResults)) {
	stack := []LocationOfInterest{{x, y, z, trackAndDirection, company, trackType}}
	for len(stack) > 0 {
		in := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		m.findAllUsableTrackInNetwork(&stack, flags, in, filter, res)
	}
	if transform != nil {
		transform(res)
	}
}
