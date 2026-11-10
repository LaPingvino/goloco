package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Road mods (tram tracks…) on a piece, its block or all the road connected to
// it: Vehicles/Routing.cpp applyRoadModToRoad / removeRoadModToTrack /
// applyRoadModsToTrackNetwork / removeRoadModsToTrackNetwork, and the
// createRoadMod / removeRoadMod game commands (GameCommands/Road/).

// RoadModsPlacement is RoadModsPlacementArgs (and RoadModsRemovalArgs).
type RoadModsPlacement struct {
	X, Y, Z      int32
	Rotation     uint8
	RoadID       uint8
	Index        uint8
	Type         uint8 // the mods: bit i = the road object's mod i
	RoadObjectID uint8
	Section      ModSection
}

func roadModStart(in LocationOfInterest) (int32, int32, int32) {
	if rad := in.TaD(); rad&4 != 0 {
		return roadEndTile(in.X, in.Y, in.Z, rad)
	}
	return in.X, in.Y, in.Z
}

func addReverseRoadInterest(in LocationOfInterest, res *RoutingResults) {
	if res == nil {
		return
	}
	r := in
	r.X, r.Y, r.Z = roadEndTile(in.X, in.Y, in.Z, in.TaD())
	r.TrackAndDirection ^= 1 << 2
	res.Reachable.TryAdd(r)
}

func (m *Map) roadModElement(x, y, z int32, rot, seq, roadObj, roadID uint8) *scenario.RoadElement {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return nil
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
			continue
		}
		el := &t.Roads[r.Index]
		if int32(el.BaseZ)*smallZStep == z && el.Rotation == rot && el.SequenceIndex == seq && el.RoadObjectID == roadObj && el.RoadID == roadID {
			return el
		}
	}
	return nil
}

func (m *Map) roadModExtra(c *Cmd, roadObj uint8, i int) (pieces uint16, cost, sell int16, index uint8, ok bool) {
	road := c.roadObject(roadObj)
	if road == nil || i >= len(road.ModNames) || c.Objects == nil {
		return 0, 0, 0, 0, false
	}
	ex := c.Objects.Extra(road.ModNames[i])
	if ex == nil {
		return 0, 0, 0, 0, false
	}
	return ex.Pieces, ex.BuildCostFactor, ex.SellCostFactor, ex.CostIndex, true
}

// applyRoadModToRoad (0x0047A5E6): as applyTrackModToTrack with a road's two mods.
func (m *Map) applyRoadModToRoad(c *Cmd, in LocationOfInterest, flags uint8, res *RoutingResults, section ModSection, roadObj, mods uint8, total *int32, company uint8, allFailed *bool) bool {
	addReverseRoadInterest(in, res)
	rad := in.TaD()
	id, rot := uint8(rad>>3&0xF), uint8(rad&3)
	failed := false
	for i := range 2 {
		if mods&(1<<i) == 0 {
			continue
		}
		pieces, _, _, _, ok := m.roadModExtra(c, roadObj, i)
		pf := RoadCompatibleFlags[id]
		if !ok || pieces&pf != pf { // a missing mod object (upstream would read past the list): refused
			failed = true
			break
		}
	}
	if !failed && int(id) < len(RoadPieces) {
		sx, sy, sz := roadModStart(in)
		for _, pc := range RoadPieces[id] {
			ox, oy := rotOffset(pc.X, pc.Y, rot)
			el := m.roadModElement(sx+int32(ox), sy+int32(oy), sz+int32(pc.Z), rot, pc.Index, roadObj, id)
			if el == nil {
				break
			}
			if pc.Index == 0 {
				*allFailed = false
				for i := range 2 {
					if mods&(1<<i) != 0 && el.Mods&(1<<i) == 0 {
						if _, cf, _, ci, ok := m.roadModExtra(c, roadObj, i); ok {
							*total += m.InflationAdjustedCost(cf, ci, 10) * int32(RoadCostFactor[id]) / 256
						}
					}
				}
			}
			if flags&CmdApply != 0 {
				if flags&CmdGhost != 0 {
					if m.ControllingCompany == company {
						el.HasGhostMods = true
					}
				} else {
					el.Mods |= mods & 3
				}
			}
		}
	}
	return section == ModSectionBlock && in.TrackAndDirection&TaDHasSignal != 0
}

// removeRoadModToTrack (0x0047A8F0): the refund of the asked mods the piece
// has and, applying, a ghost's flag cleared or the mods taken off.
func (m *Map) removeRoadModToTrack(c *Cmd, in LocationOfInterest, flags uint8, res *RoutingResults, section ModSection, roadObj, mods uint8, total *int32, company uint8) bool {
	addReverseRoadInterest(in, res)
	rad := in.TaD()
	id, rot := uint8(rad>>3&0xF), uint8(rad&3)
	if int(id) < len(RoadPieces) {
		sx, sy, sz := roadModStart(in)
		for _, pc := range RoadPieces[id] {
			ox, oy := rotOffset(pc.X, pc.Y, rot)
			el := m.roadModElement(sx+int32(ox), sy+int32(oy), sz+int32(pc.Z), rot, pc.Index, roadObj, id)
			if el == nil {
				break
			}
			if pc.Index == 0 {
				for i := range 2 {
					if mods&(1<<i) != 0 && el.Mods&(1<<i) != 0 {
						if _, _, sf, ci, ok := m.roadModExtra(c, roadObj, i); ok {
							*total += m.InflationAdjustedCost(sf, ci, 10) * int32(RoadCostFactor[id]) / 256
						}
					}
				}
			}
			if flags&CmdApply != 0 {
				if flags&CmdGhost != 0 {
					if m.ControllingCompany == company {
						el.HasGhostMods = false
					}
				} else {
					el.Mods &^= mods & 3 // upstream loops i < 4: setMod(2 / 3) shifts past the byte
				}
			}
		}
	}
	return section == ModSectionBlock && in.TrackAndDirection&TaDHasSignal != 0
}

// ApplyRoadModsToTrackNetwork is applyRoadModsToTrackNetwork.
func (m *Map) ApplyRoadModsToTrackNetwork(c *Cmd, x, y, z int32, rad uint16, company, roadType, flags uint8, section ModSection, mods uint8) ApplyTrackModsResult {
	res := ApplyTrackModsResult{AllPlacementsFailed: true}
	if section == ModSectionSingle {
		m.applyRoadModToRoad(c, LocationOfInterest{x, y, z, rad, company, roadType}, flags, nil, section, roadType, mods, &res.Cost, company, &res.AllPlacementsFailed)
		return res
	}
	rr := &RoutingResults{Reachable: NewInterestSet(TrackModHashSetSize)}
	m.FindAllRoadsFilterTransform(rr, 0, x, y, z, rad, company, roadType, func(in LocationOfInterest) bool {
		return m.applyRoadModToRoad(c, in, flags, rr, section, roadType, mods, &res.Cost, company, &res.AllPlacementsFailed)
	}, nil)
	res.NetworkTooComplex = rr.Reachable.count >= rr.Reachable.max
	return res
}

// RemoveRoadModsToTrackNetwork is removeRoadModsToTrackNetwork: the refund.
func (m *Map) RemoveRoadModsToTrackNetwork(c *Cmd, x, y, z int32, rad uint16, company, roadType, flags uint8, section ModSection, mods uint8) int32 {
	var cost int32
	if section == ModSectionSingle {
		m.removeRoadModToTrack(c, LocationOfInterest{x, y, z, rad, company, roadType}, flags, nil, section, roadType, mods, &cost, company)
		return cost
	}
	rr := &RoutingResults{Reachable: NewInterestSet(TrackModHashSetSize)}
	m.FindAllRoadsFilterTransform(rr, 0, x, y, z, rad, company, roadType, func(in LocationOfInterest) bool {
		return m.removeRoadModToTrack(c, in, flags, rr, section, roadType, mods, &cost, company)
	}, nil)
	return cost
}

// roadModsTarget: the element, company compatibility, the piece's first tile
// (createRoadMod rotates the piece offset by args.rotation; the lookup made
// it equal to the element's) and its rad.
func (m *Map) roadModsTarget(c *Cmd, a RoadModsPlacement) (x, y, z int32, rad uint16, owner uint8, ok bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	el := m.roadModElement(a.X, a.Y, a.Z, a.Rotation, a.Index, a.RoadObjectID, a.RoadID)
	if el == nil || !c.companyCompatible(el.Owner) || int(el.RoadID) >= len(RoadPieces) || int(el.SequenceIndex) >= len(RoadPieces[el.RoadID]) {
		return 0, 0, 0, 0, 0, false
	}
	p := RoadPieces[el.RoadID][el.SequenceIndex]
	dx, dy := rotOffset(p.X, p.Y, el.Rotation)
	return a.X - int32(dx), a.Y - int32(dy), a.Z - int32(p.Z), uint16(el.RoadID)<<3 | uint16(el.Rotation&3), el.Owner, true
}

// CreateRoadMod is the createRoadMod game command (0x0047A21E); refused
// (track_road_unsuitable) when no piece took the mods. The caller pays under
// Construction. PARTIAL: the too-much-track error window.
func (m *Map) CreateRoadMod(c *Cmd, flags uint8, a RoadModsPlacement) (int32, bool) {
	x, y, z, rad, owner, ok := m.roadModsTarget(c, a)
	if !ok {
		return 0, false
	}
	r := m.ApplyRoadModsToTrackNetwork(c, x, y, z, rad, owner, a.RoadObjectID, flags, a.Section, a.Type)
	if r.AllPlacementsFailed {
		c.ErrorText = i18n.TrackRoadUnsuitable
		return 0, false
	}
	return r.Cost, true
}

// RemoveRoadMod is the removeRoadMod game command (0x0047A42F): the refund.
func (m *Map) RemoveRoadMod(c *Cmd, flags uint8, a RoadModsPlacement) (int32, bool) {
	x, y, z, rad, owner, ok := m.roadModsTarget(c, a)
	if !ok {
		return 0, false
	}
	return m.RemoveRoadModsToTrackNetwork(c, x, y, z, rad, owner, a.RoadObjectID, flags, a.Section, a.Type), true
}
