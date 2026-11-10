package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Track mods (overhead wires, third rail…) on a piece, its block or all the
// track connected to it: Vehicles/Routing.cpp applyTrackModToTrack /
// removeTrackModToTrack / applyTrackModsToTrackNetwork /
// removeTrackModsToTrackNetwork, and the createTrackMod / removeTrackMod
// game commands (GameCommands/Track/CreateTrackMod.cpp, RemoveTrackMod.cpp).

// ModSection is Track::ModSection.
type ModSection uint8

const (
	ModSectionSingle       ModSection = 0
	ModSectionBlock        ModSection = 1
	ModSectionAllConnected ModSection = 2
)

// TrackModsPlacement is TrackModsPlacementArgs (and TrackModsRemovalArgs):
// a tile of the piece, its rotation, piece, sequence index, the mods (bit i =
// the track object's mod i), the track object and the section.
type TrackModsPlacement struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	Index         uint8
	Type          uint8
	TrackObjectID uint8
	Section       ModSection
}

// trackModStart: the piece's first tile for the interest (a reversed one
// starts at its far end).
func trackModStart(in LocationOfInterest) (int32, int32, int32) {
	tad := in.TaD()
	if tad&4 != 0 {
		return trackEndTile(in.X, in.Y, in.Z, tad)
	}
	return in.X, in.Y, in.Z
}

// addReverseInterest: outside single mode the piece's reverse is marked
// visited — mods have no direction, applying to it again would pay twice.
func addReverseInterest(in LocationOfInterest, res *RoutingResults) {
	if res == nil {
		return
	}
	r := in
	r.X, r.Y, r.Z = trackEndTile(in.X, in.Y, in.Z, in.TaD())
	r.TrackAndDirection ^= 1 << 2
	res.Reachable.TryAdd(r)
}

// trackModElement is the walk's element lookup on one tile of the piece
// (base height, rotation, sequence index, track object, piece).
func (m *Map) trackModElement(x, y, z int32, rot, seq, trackObj, trackID uint8) *scenario.TrackElement {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return nil
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindTrack || int(r.Index) >= len(t.Tracks) {
			continue
		}
		el := &t.Tracks[r.Index]
		if int32(el.BaseZ)*smallZStep == z && el.Rotation == rot && el.SequenceIndex == seq && el.TrackObjectID == trackObj && el.TrackID == trackID {
			return el
		}
	}
	return nil
}

// trackModCost: mod i's inflation-adjusted factor times the piece's cost factor / 256.
func (m *Map) trackModCost(c *Cmd, trackObj uint8, i int, trackID uint8, sell bool) int32 {
	track := c.trackObject(trackObj)
	if track == nil || i >= len(track.ModNames) || c.Objects == nil {
		return 0
	}
	ex := c.Objects.Extra(track.ModNames[i])
	if ex == nil {
		return 0
	}
	f := ex.BuildCostFactor
	if sell {
		f = ex.SellCostFactor
	}
	return m.InflationAdjustedCost(f, ex.CostIndex, 10) * int32(TrackCostFactor[trackID]) / 256
}

// applyTrackModToTrack (Routing.cpp): refused (a placement
// failure) when a mod can't go on this kind of piece; otherwise the cost of
// the mods the piece lacks (read on its first tile, which also counts as a
// successful placement) and, applying, a ghost's flag (the controlling
// company's only) or the mods set. In block mode the walk stops at a signal.
func (m *Map) applyTrackModToTrack(c *Cmd, in LocationOfInterest, flags uint8, res *RoutingResults, section ModSection, trackObj, mods uint8, total *int32, company uint8, allFailed *bool) bool {
	addReverseInterest(in, res)
	tad := in.TaD()
	id, rot := uint8(tad>>3), uint8(tad&3)
	failed := false
	if track := c.trackObject(trackObj); track != nil {
		for i := range 4 {
			if mods&(1<<i) == 0 {
				continue
			}
			var pieces uint16
			found := false
			if i < len(track.ModNames) && c.Objects != nil {
				if e := c.Objects.Extra(track.ModNames[i]); e != nil {
					pieces, found = e.Pieces, true
				}
			}
			pf := TrackCompatibleFlags[id]
			if !found || pieces&pf != pf { // a missing mod object (upstream would read past the list): refused
				failed = true
				break
			}
		}
	}
	if !failed {
		sx, sy, sz := trackModStart(in)
		for _, pc := range TrackPieces[id] {
			ox, oy := rotOffset(pc.X, pc.Y, rot)
			x, y, z := sx+int32(ox), sy+int32(oy), sz+int32(pc.Z)
			el := m.trackModElement(x, y, z, rot, pc.Index, trackObj, id)
			if el == nil {
				break
			}
			if pc.Index == 0 {
				*allFailed = false
				for i := range 4 {
					if mods&(1<<i) != 0 && el.Mods&(1<<i) == 0 {
						*total += m.trackModCost(c, trackObj, i, id, false)
					}
				}
			}
			if flags&CmdApply != 0 {
				if flags&CmdGhost != 0 {
					if m.ControllingCompany == company {
						el.HasGhostMods = true
					}
				} else {
					el.Mods |= mods & 0xF
				}
			}
		}
	}
	return section == ModSectionBlock && in.TrackAndDirection&TaDHasSignal != 0
}

// removeTrackModToTrack (0x004A6136): the refund of the asked mods the piece
// has (first tile) and, applying, a ghost's flag cleared (the controlling
// company's) or the mods taken off.
func (m *Map) removeTrackModToTrack(c *Cmd, in LocationOfInterest, flags uint8, res *RoutingResults, section ModSection, trackObj, mods uint8, total *int32, company uint8) bool {
	addReverseInterest(in, res)
	tad := in.TaD()
	id, rot := uint8(tad>>3), uint8(tad&3)
	sx, sy, sz := trackModStart(in)
	for _, pc := range TrackPieces[id] {
		ox, oy := rotOffset(pc.X, pc.Y, rot)
		x, y, z := sx+int32(ox), sy+int32(oy), sz+int32(pc.Z)
		el := m.trackModElement(x, y, z, rot, pc.Index, trackObj, id)
		if el == nil {
			break
		}
		if pc.Index == 0 {
			for i := range 4 {
				if mods&(1<<i) != 0 && el.Mods&(1<<i) != 0 {
					*total += m.trackModCost(c, trackObj, i, id, true)
				}
			}
		}
		if flags&CmdApply != 0 {
			if flags&CmdGhost != 0 {
				if m.ControllingCompany == company {
					el.HasGhostMods = false
				}
			} else {
				el.Mods &^= mods & 0xF
			}
		}
	}
	return section == ModSectionBlock && in.TrackAndDirection&TaDHasSignal != 0
}

// ApplyTrackModsResult is Vehicles::ApplyTrackModsResult.
type ApplyTrackModsResult struct {
	Cost                int32
	AllPlacementsFailed bool
	NetworkTooComplex   bool
}

// ApplyTrackModsToTrackNetwork: single — the one piece; otherwise the walk
// over the network (no search flags) with the apply as its filter.
func (m *Map) ApplyTrackModsToTrackNetwork(c *Cmd, x, y, z int32, tad uint16, company, trackType, flags uint8, section ModSection, mods uint8) ApplyTrackModsResult {
	res := ApplyTrackModsResult{AllPlacementsFailed: true}
	if section == ModSectionSingle {
		m.applyTrackModToTrack(c, LocationOfInterest{x, y, z, tad, company, trackType}, flags, nil, section, trackType, mods, &res.Cost, company, &res.AllPlacementsFailed)
		return res
	}
	rr := &RoutingResults{Reachable: NewInterestSet(TrackModHashSetSize)}
	m.FindAllTracksFilterTransform(rr, 0, x, y, z, tad, company, trackType, func(in LocationOfInterest) bool {
		return m.applyTrackModToTrack(c, in, flags, rr, section, trackType, mods, &res.Cost, company, &res.AllPlacementsFailed)
	}, nil)
	res.NetworkTooComplex = rr.Reachable.count >= rr.Reachable.max
	return res
}

// RemoveTrackModsToTrackNetwork is removeTrackModsToTrackNetwork: the refund.
func (m *Map) RemoveTrackModsToTrackNetwork(c *Cmd, x, y, z int32, tad uint16, company, trackType, flags uint8, section ModSection, mods uint8) int32 {
	var cost int32
	if section == ModSectionSingle {
		m.removeTrackModToTrack(c, LocationOfInterest{x, y, z, tad, company, trackType}, flags, nil, section, trackType, mods, &cost, company)
		return cost
	}
	rr := &RoutingResults{Reachable: NewInterestSet(TrackModHashSetSize)}
	m.FindAllTracksFilterTransform(rr, 0, x, y, z, tad, company, trackType, func(in LocationOfInterest) bool {
		return m.removeTrackModToTrack(c, in, flags, rr, section, trackType, mods, &cost, company)
	}, nil)
	return cost
}

// trackModsTarget is the two commands' shared start: the element (any ghost
// state), company compatibility, the piece's first tile and its tad.
func (m *Map) trackModsTarget(c *Cmd, a TrackModsPlacement) (x, y, z int32, tad uint16, owner uint8, ok bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	el := m.trackModElement(a.X, a.Y, a.Z, a.Rotation, a.Index, a.TrackObjectID, a.TrackID)
	if el == nil || int(el.TrackID) >= len(TrackPieces) || int(el.SequenceIndex) >= len(TrackPieces[el.TrackID]) || !c.companyCompatible(el.Owner) {
		return 0, 0, 0, 0, 0, false
	}
	p := TrackPieces[el.TrackID][el.SequenceIndex]
	dx, dy := rotOffset(p.X, p.Y, el.Rotation)
	return a.X - int32(dx), a.Y - int32(dy), a.Z - int32(p.Z), uint16(el.TrackID)<<3 | uint16(el.Rotation&3), el.Owner, true
}

// CreateTrackMod is the createTrackMod game command (0x004A6479): the mods
// over the section; refused (track_road_unsuitable) when no piece took them.
// The caller pays under Construction.
// PARTIAL: the "too much track, some not upgraded" error window when the
// network was too complex (NetworkTooComplex is returned for it).
func (m *Map) CreateTrackMod(c *Cmd, flags uint8, a TrackModsPlacement) (int32, bool) {
	x, y, z, tad, owner, ok := m.trackModsTarget(c, a)
	if !ok {
		return 0, false
	}
	r := m.ApplyTrackModsToTrackNetwork(c, x, y, z, tad, owner, a.TrackObjectID, flags, a.Section, a.Type)
	if r.AllPlacementsFailed {
		c.ErrorText = i18n.TrackRoadUnsuitable
		return 0, false
	}
	return r.Cost, true
}

// RemoveTrackMod is the removeTrackMod game command: the refund over the section.
func (m *Map) RemoveTrackMod(c *Cmd, flags uint8, a TrackModsPlacement) (int32, bool) {
	x, y, z, tad, owner, ok := m.trackModsTarget(c, a)
	if !ok {
		return 0, false
	}
	return m.RemoveTrackModsToTrackNetwork(c, x, y, z, tad, owner, a.TrackObjectID, flags, a.Section, a.Type), true
}
