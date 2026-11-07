package sim

import "github.com/LaPingvino/goloco/pkg/objects"

// createTrack (GameCommands/Track/CreateTrack.cpp 321–585), PL.1 (e): the
// game command a player builds track with (the construction window's "Build
// this…"). Ported in parts, as createRoad was; this is its start: the
// arguments, the track extras that fit the piece, the height rule and the
// cost of the track and its extras.
// createTrackCheck (part 3) is the per-piece checks and the bridge / tunnel
// costs. PARTIAL: the apply step (inserting the elements), the company
// observation, emotion and sound. Until then the window still builds with
// PlaceTrackAtHead (no cost — LOOP PL.1 e).

// trackPlacementArgs is TrackPlacementArgs.
type trackPlacementArgs struct {
	X, Y, Z       int32
	Rotation      uint8 // 0–3
	TrackID       uint8 // 0–43
	Mods          uint8 // the track extras asked for (bits 0–3)
	Bridge        uint8 // a bridge object slot, 0xFF none
	TrackObjectID uint8
	UnkFlags      uint8
}

// trackObject is the track object in a slot, or nil.
func (c *Cmd) trackObject(id uint8) *objects.TrackObject {
	if c.Objects != nil && int(id) < len(c.Objects.TrackObjects) {
		return c.Objects.TrackObjects[id]
	}
	return nil
}

// createTrackCost is createTrack's start (0x0049BB98 up to the pieces): the
// extras asked for that fit this piece (its compatibleFlags all in the
// extra's trackPieces), the height (a multiple of 16: no Loco piece has
// startsAtHalfHeight, the RCT2 leftover), and the cost — the track's and each
// valid extra's inflation-adjusted build cost (divisor 10) times the piece's
// cost factor / 256.
func (m *Map) createTrackCost(c *Cmd, a trackPlacementArgs) (validMods uint8, cost int32, ok bool) {
	track := c.trackObject(a.TrackObjectID)
	if track == nil || int(a.TrackID) >= len(TrackCompatibleFlags) {
		return 0, 0, false
	}
	compat := TrackCompatibleFlags[a.TrackID]
	validMods = a.Mods
	extra := func(i int) *objects.ExtraObject {
		if i >= len(track.ModNames) || c.Objects == nil {
			return nil
		}
		return c.Objects.Extra(track.ModNames[i])
	}
	for i := range 4 {
		if a.Mods&(1<<i) == 0 {
			continue
		}
		if ex := extra(i); ex == nil || ex.Pieces&compat != compat {
			validMods &^= 1 << i
		}
	}
	if a.Z&0xF != 0 {
		return validMods, 0, false
	}
	factor := int32(TrackCostFactor[a.TrackID])
	cost = m.InflationAdjustedCost(track.BuildCostFactor, track.CostIndex, 10) * factor / 256
	for i := range 4 {
		if validMods&(1<<i) != 0 {
			ex := extra(i)
			cost += m.InflationAdjustedCost(ex.BuildCostFactor, ex.CostIndex, 10) * factor / 256
		}
	}
	return validMods, cost, true
}

// trackPiecePlan is one piece tile createTrackCheck cleared: what the apply
// step (0x0049C015 on) places there.
type trackPiecePlan struct {
	X, Y, Z          int32
	Index            uint8
	BaseZ, ClearZ    uint8
	Quarter          uint8 // the rotated subTileClearance
	HasBridge        bool  // flags_1136073 bit 1 for this tile
	HasLevelCrossing bool
}

// createTrackResult is what createTrackCheck found.
type createTrackResult struct {
	Cost      int32
	ValidMods uint8
	Pieces    []trackPiecePlan
	State     trackCmdState
}

// createTrackCheck is createTrack's checks (CreateTrack.cpp 386–585 without
// the insertion): the cost and, piece by piece, the height (not below 16,
// clearZ not above 236), water and the bridge a raised piece needs (one asked
// for, high enough, suitable for the piece), the clearance with the track
// clear function (junctions, level crossings and their cost), and the
// position rules (not partly above / below ground, not under water or too
// close to its surface); then the bridge's and a tunnel's cost. It changes
// nothing unless flags apply (the clear functions remove trees / buildings
// and mark level crossings then).
func (m *Map) createTrackCheck(c *Cmd, a trackPlacementArgs, flags uint8) (createTrackResult, bool) {
	var r createTrackResult
	st := &r.State
	st.Bridge = 0xFF
	validMods, cost, ok := m.createTrackCost(c, a)
	if !ok || int(a.TrackID) >= len(TrackPieces) {
		return r, false
	}
	r.ValidMods, r.Cost = validMods, cost
	pieces := TrackPieces[a.TrackID]
	removed := RemovedBuildings{}
	for _, piece := range pieces {
		ox, oy := rotOffset(piece.X, piece.Y, a.Rotation)
		x, y, z := a.X+int32(ox), a.Y+int32(oy), a.Z+int32(piece.Z)
		quarter := rotateQuarterTile(piece.SubTileClearance, a.Rotation)
		unk := rotl4bit(piece.Flags&0xF, a.Rotation)
		if z < 16 { // error_too_low
			return r, false
		}
		baseZ := uint8(z / 4)
		clearZ := int(baseZ) + (int(piece.ClearZ)+32)/4
		st.Flags73 &^= 1 << 1
		if m.validCoords(x, y) {
			t := m.TileAt(int(x/tileSize), int(y/tileSize))
			if t.WaterLevel != 0 {
				st.Flags73 |= 1 << 7
			}
			if isBridgeRequired(baseZ, t, piece.Flags&(1<<4) != 0, unk) {
				st.Flags73 |= 1<<1 | 1<<0
				heightDiff := (int(baseZ) - int(t.BaseZ)) / microToSmallZStep
				if a.Bridge == 0xFF { // bridge_needed
					return r, false
				}
				b := c.bridge(a.Bridge)
				if b == nil || heightDiff > int(b.MaxHeight) { // too_far_above_ground_for_bridge_type
					return r, false
				}
				st.MaxBridgeDiff = max(st.MaxBridgeDiff, uint8(heightDiff))
				if b.DisabledTrackCfg&TrackMiscFlags[a.TrackID] != 0 { // bridge_type_unsuitable_for_this_configuration
					return r, false
				}
				clearZ += int(b.ClearHeight) / 4
			}
		}
		if clearZ > 236 { // error_too_high
			return r, false
		}
		plan := trackPiecePlan{X: x, Y: y, Z: z, Index: piece.Index, BaseZ: baseZ, ClearZ: uint8(clearZ), Quarter: quarter}
		args := trackClearArgs{X: x, Y: y, Z: z, Rotation: a.Rotation, TrackID: a.TrackID, UnkFlags: a.UnkFlags, BridgeID: a.Bridge,
			TrackObjectID: a.TrackObjectID, Index: piece.Index, IsLastIndex: int(piece.Index) == len(pieces)-1, Flags: flags}
		clear := func(t *Tile, k int) ClearResult {
			return m.trackClearFunction(c, st, t, k, &r.Cost, &plan.HasLevelCrossing, removed, args)
		}
		ok, pos := m.ApplyClearAtStandardHeight(x, y, baseZ, uint8(clearZ), quarter&0xF, quarter>>4, clear)
		if !ok {
			return r, false
		}
		if plan.HasLevelCrossing && c.Objects != nil && int(m.DefaultLevelCrossingType) < len(c.Objects.LevelCrossingObjects) {
			if lc := c.Objects.LevelCrossingObjects[m.DefaultLevelCrossingType]; lc != nil {
				r.Cost += m.InflationAdjustedCost(lc.CostFactor, lc.CostIndex, 10)
			}
		}
		ground := pos & (PosAboveGround | PosUnderground)
		if st.PosFlags != 0 && st.PosFlags&ground == 0 { // cant_build_partly_above_partly_below_ground
			return r, false
		}
		st.PosFlags = ground
		if pos&(PosPartiallyUnderwater|PosUnderwater) != 0 { // under water / too close to its surface
			return r, false
		}
		plan.HasBridge = st.Flags73&(1<<1) != 0
		r.Pieces = append(r.Pieces, plan)
	}
	factor := int32(TrackCostFactor[a.TrackID])
	if st.Flags73&(1<<0) != 0 {
		if b := c.bridge(a.Bridge); b != nil {
			height := int32(st.MaxBridgeDiff) * int32(b.HeightCostFactor)
			cost := m.InflationAdjustedCost(int16(int32(b.BaseCostFactor)+height), b.CostIndex, 10) * factor / 256
			if st.Flags73&(1<<7) != 0 {
				cost *= 2
			}
			r.Cost += cost
		}
	}
	if st.PosFlags&PosUnderground != 0 {
		if track := c.trackObject(a.TrackObjectID); track != nil {
			r.Cost += m.InflationAdjustedCost(track.TunnelCostFactor, 2, 8) * factor / 256
		}
	}
	return r, true
}
