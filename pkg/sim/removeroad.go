package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// GameCommands/Road/RemoveRoad.cpp: taking a road piece up — by a player (the
// construction window's remove) or as a ghost (the road ghost moved on).

// RoadRemoval is RoadRemovalArgs: one of the piece's tiles (world units, its
// base height), its rotation, the piece, which of its tiles that is, and the
// road object.
type RoadRemoval struct {
	X, Y, Z      int32
	Rotation     uint8
	RoadID       uint8
	Index        uint8
	RoadObjectID uint8
}

// roadElementAt is getRoadElement: the road element on the tile at x, y with
// this base height, the removal's rotation, road object and piece and this
// sequence index, ghost and AI-allocated as the flags say; a ghost must be the
// updating company's (nil: none — ownership of others is checked by the caller).
func (m *Map) roadElementAt(c *Cmd, a RoadRemoval, flags uint8, x, y, z int32, seq uint8) (*Tile, int) {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return nil, -1
	}
	ghost, ai := flags&CmdGhost != 0, flags&CmdAiAllocated != 0
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad || int(r.Index) >= len(t.Roads) {
			continue
		}
		el := &t.Roads[r.Index]
		if int32(el.BaseZ)*smallZStep != z || el.Rotation != a.Rotation || el.SequenceIndex != seq ||
			el.RoadObjectID != a.RoadObjectID || el.RoadID != a.RoadID || el.Ghost != ghost || el.AiAllocated != ai {
			continue
		}
		if ghost && el.Owner != c.Company {
			return nil, -1 // upstream returns nullptr here, it doesn't look further
		}
		return t, int(r.Index)
	}
	return nil, -1
}

// roadRemoveCost is 0x00477A10: the piece's refund read from its first tile —
// none for a neutral (town) road; else the road's inflation-adjusted
// sellCostFactor (divisor 10) times the piece's cost factor / 256, and each
// extra on it likewise unless the road is any-road-type compatible (0 when
// that tile has gone).
func (m *Map) roadRemoveCost(c *Cmd, a RoadRemoval, p0 PreviewTrack, sx, sy, sz int32, flags uint8) int32 {
	dx, dy := rotOffset(p0.X, p0.Y, a.Rotation)
	t, i := m.roadElementAt(c, a, flags, sx+int32(dx), sy+int32(dy), sz+int32(p0.Z), p0.Index)
	if t == nil {
		return 0
	}
	el := &t.Roads[i]
	road := c.roadObject(el.RoadObjectID)
	if road == nil {
		return 0
	}
	factor := int32(RoadCostFactor[a.RoadID])
	var total int32
	if el.Owner != CompanyNeutral {
		total += m.InflationAdjustedCost(road.SellCostFactor, road.CostIndex, 10) * factor / 256
	}
	if road.Flags&roadFlagAnyRoadTypeCompatible == 0 {
		for k := range 2 {
			if el.Mods&(1<<k) == 0 || k >= len(road.ModNames) || c.Objects == nil {
				continue
			}
			if ex := c.Objects.Extra(road.ModNames[k]); ex != nil {
				total += m.InflationAdjustedCost(ex.SellCostFactor, ex.CostIndex, 10) * factor / 256
			}
		}
	}
	return total
}

// RemoveRoad is the removeRoad game command (0x004775A5), the position its
// tile's centre: refused when the element isn't there or (not a ghost) the
// owner isn't compatible; then each tile of the piece that is found is
// (applying) removed — upstream skips a missing tile rather than refusing; the
// cost is the piece's refund plus, when it was the only road on a bridge, the
// bridge's (OpenLoco's addition). A road stop on the piece goes first (a
// nested RoadStationRemoval), unless another road at that height also uses
// it. The caller pays under Construction.
// After each tile setLevelCrossingFlags. PARTIAL: madeAnyChanges; the sound.
func (m *Map) RemoveRoad(c *Cmd, flags uint8, a RoadRemoval, g1 G1Raw) (int32, bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	if int(a.RoadID) >= len(RoadPieces) {
		return 0, false
	}
	t, i := m.roadElementAt(c, a, flags, a.X, a.Y, a.Z, a.Index)
	if t == nil {
		return 0, false
	}
	el := t.Roads[i]
	if !c.companyCompatible(el.Owner) {
		return 0, false
	}
	var total int32
	if st := m.roadStationAt(a.X, a.Y, a.Z); el.HasStationElement && !m.otherRoadStationUser(t, i, el.BaseZ) && st != nil && !st.Ghost {
		cost, ok := nested(flags, func(f uint8) (int32, bool) {
			return m.RemoveRoadStation(c, f, RoadStationRemoval{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, RoadID: a.RoadID, Index: a.Index, RoadObjectID: a.RoadObjectID}, g1)
		})
		if !ok {
			return 0, false
		}
		total += cost
	}
	pieces := RoadPieces[a.RoadID]
	if int(a.Index) >= len(pieces) {
		return 0, false
	}
	p := pieces[a.Index]
	dx, dy := rotOffset(p.X, p.Y, a.Rotation)
	sx, sy, sz := a.X-int32(dx), a.Y-int32(dy), a.Z-int32(p.Z)
	pieceCost := m.roadRemoveCost(c, a, pieces[0], sx, sy, sz, flags)
	removeBridge, bridge := false, uint8(0)
	for _, pc := range pieces {
		ox, oy := rotOffset(pc.X, pc.Y, a.Rotation)
		x, y, z := sx+int32(ox), sy+int32(oy), sz+int32(pc.Z)
		if flags&CmdAiAllocated == 0 { // shouldInvalidateTile
			m.mapInvalidateTileFull(x, y)
		}
		pt, pi := m.roadElementAt(c, a, flags, x, y, z, pc.Index)
		if pt == nil {
			continue
		}
		if pt.Roads[pi].HasBridge {
			// OverlapRoads(args.pos): the bridge goes only with the only road there.
			removeBridge = m.roadsAtHeight(a.X, a.Y, a.Z) == 1
			bridge = pt.Roads[pi].BridgeID
		}
		if flags&CmdApply == 0 {
			continue
		}
		pt.Roads = append(pt.Roads[:pi], pt.Roads[pi+1:]...)
		pt.removeRef(scenario.KindRoad, pi)
		m.setLevelCrossingFlags(x, y, z)
	}
	total += pieceCost
	if removeBridge && el.Owner != CompanyNeutral && c.Objects != nil && int(bridge) < len(c.Objects.BridgeObjects) {
		if b := c.Objects.BridgeObjects[bridge]; b != nil {
			total += m.InflationAdjustedCost(b.SellCostFactor, b.CostIndex, 10) * int32(RoadCostFactor[a.RoadID]) / 256
		}
	}
	return total, true
}

// otherRoadStationUser: another road at this height on the tile (OverlapRoads)
// that also has a station element — then the stop stays.
func (m *Map) otherRoadStationUser(t *Tile, self int, baseZ uint8) bool {
	for _, r := range t.Order {
		if r.Kind == scenario.KindRoad && int(r.Index) != self && int(r.Index) < len(t.Roads) {
			if o := t.Roads[r.Index]; o.BaseZ == baseZ && o.HasStationElement {
				return true
			}
		}
	}
	return false
}

// roadsAtHeight is OverlapRoads' count: the road elements at this height on
// the tile, in one run of the tile's element order.
func (m *Map) roadsAtHeight(x, y, z int32) int {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return 0
	}
	n, started := 0, false
	for _, r := range t.Order {
		isRoad := r.Kind == scenario.KindRoad && int(r.Index) < len(t.Roads) && int32(t.Roads[r.Index].BaseZ)*smallZStep == z
		if !isRoad {
			if started {
				break
			}
			continue
		}
		started = true
		n++
	}
	return n
}
