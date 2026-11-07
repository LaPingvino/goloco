package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// GameCommands/Track/RemoveTrack.cpp: taking a track piece up — by a player
// (the construction window's remove) or as a ghost (removeConstructionGhosts,
// the ghost under the cursor moved on).

// TrackRemoval is TrackRemovalArgs: the tile of one of the piece's tiles
// (world units, its base height), its rotation, the piece, which of its
// tiles that is, and the track object.
type TrackRemoval struct {
	X, Y, Z       int32
	Rotation      uint8
	TrackID       uint8
	Index         uint8
	TrackObjectID uint8
}

// trackElementAt is getElTrackAt: the track element on the tile at x, y with
// the removal's rotation, track object and piece, this base height and
// sequence index, ghost and AI-allocated as the flags say, and — for a
// ghost — the updating company's (−1 none).
func (m *Map) trackElementAt(c *Cmd, a TrackRemoval, flags uint8, x, y, z int32, seq uint8) (*Tile, int) {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return nil, -1
	}
	ghost, ai := flags&CmdGhost != 0, flags&CmdAiAllocated != 0
	for _, r := range t.Order {
		if r.Kind != scenario.KindTrack || int(r.Index) >= len(t.Tracks) {
			continue
		}
		el := &t.Tracks[r.Index]
		if el.Rotation != a.Rotation || int32(el.BaseZ)*smallZStep != z || el.SequenceIndex != seq ||
			el.TrackObjectID != a.TrackObjectID || el.TrackID != a.TrackID || el.Ghost != ghost || el.AiAllocated != ai {
			continue
		}
		if ghost && el.Owner != c.Company { // non-ghost ownership is checked by the caller (neutral allowed)
			continue
		}
		return t, int(r.Index)
	}
	return nil, -1
}

// trackRemoveCost is 0x0049CC23: the piece's sell price — the track's
// inflation-adjusted sellCostFactor (divisor 10) times the piece's cost
// factor / 256, and each of its extras' likewise — read from its first tile
// (0 when that tile has gone).
func (m *Map) trackRemoveCost(c *Cmd, a TrackRemoval, p0 PreviewTrack, sx, sy, sz int32, flags uint8) int32 {
	dx, dy := rotOffset(p0.X, p0.Y, a.Rotation)
	t, i := m.trackElementAt(c, a, flags, sx+int32(dx), sy+int32(dy), sz+int32(p0.Z), p0.Index)
	if t == nil {
		return 0
	}
	el := &t.Tracks[i]
	track := c.trackObject(el.TrackObjectID)
	if track == nil {
		return 0
	}
	factor := int32(TrackCostFactor[a.TrackID])
	total := m.InflationAdjustedCost(track.SellCostFactor, track.CostIndex, 10) * factor / 256
	for k := range 4 {
		if el.Mods&(1<<k) == 0 || k >= len(track.ModNames) || c.Objects == nil {
			continue
		}
		if ex := c.Objects.Extra(track.ModNames[k]); ex != nil {
			total += m.InflationAdjustedCost(ex.SellCostFactor, ex.CostIndex, 10) * factor / 256
		}
	}
	return total
}

// RemoveTrack is the removeTrack game command (0x0049C7F2), the position
// its tile's centre: refused when the element isn't there or (not a ghost)
// belongs to another company; a station on it goes first (TrainStationRemoval
// as a nested command); then every tile of the piece is found (refused when
// one is missing) and, applying, removed; the cost is the piece's sell price
// plus, when it stood on a bridge, the bridge's. The caller pays under
// Construction. PARTIAL: a signal on the piece refuses it (SignalRemoval is
// not ported), setLevelCrossingFlags after each tile, the demolish sound.
func (m *Map) RemoveTrack(c *Cmd, flags uint8, a TrackRemoval, g1 G1Raw) (int32, bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z}
	if int(a.TrackID) >= len(TrackPieces) {
		return 0, false
	}
	t, i := m.trackElementAt(c, a, flags, a.X, a.Y, a.Z, a.Index)
	if t == nil {
		return 0, false
	}
	el := t.Tracks[i]
	if flags&CmdGhost == 0 && !c.companyCompatible(el.Owner) {
		return 0, false
	}
	var total int32
	if el.HasSignal {
		return 0, false // PARTIAL: SignalRemoval
	}
	if el.HasStation {
		cost, ok := nested(flags, func(f uint8) (int32, bool) {
			return m.RemoveTrainStation(c, f, TrainStationRemoval{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, TrackID: a.TrackID, Index: a.Index, TrackObjectID: a.TrackObjectID}, g1)
		})
		if !ok {
			return 0, false
		}
		total += cost
	}
	pieces := TrackPieces[a.TrackID]
	if int(a.Index) >= len(pieces) {
		return 0, false
	}
	p := pieces[a.Index]
	dx, dy := rotOffset(p.X, p.Y, a.Rotation)
	sx, sy, sz := a.X-int32(dx), a.Y-int32(dy), a.Z-int32(p.Z)
	pieceCost := m.trackRemoveCost(c, a, pieces[0], sx, sy, sz, flags)
	hadBridge, bridge := false, uint8(0)
	for _, pc := range pieces {
		ox, oy := rotOffset(pc.X, pc.Y, a.Rotation)
		x, y, z := sx+int32(ox), sy+int32(oy), sz+int32(pc.Z)
		if flags&CmdAiAllocated == 0 { // shouldInvalidateTile (showAiPlanningAsGhosts off)
			m.mapInvalidateTileFull(x, y)
		}
		pt, pi := m.trackElementAt(c, a, flags, x, y, z, pc.Index)
		if pt == nil {
			return 0, false
		}
		if pt.Tracks[pi].HasBridge {
			hadBridge, bridge = true, pt.Tracks[pi].BridgeID
		}
		if flags&CmdApply == 0 {
			continue
		}
		pt.Tracks = append(pt.Tracks[:pi], pt.Tracks[pi+1:]...)
		pt.removeRef(scenario.KindTrack, pi)
	}
	total += pieceCost
	if hadBridge && c.Objects != nil && int(bridge) < len(c.Objects.BridgeObjects) {
		if b := c.Objects.BridgeObjects[bridge]; b != nil {
			total += m.InflationAdjustedCost(b.SellCostFactor, b.CostIndex, 10) * int32(TrackCostFactor[a.TrackID]) / 256
		}
	}
	return total, true
}
