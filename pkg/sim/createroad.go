package sim

// createRoad (GameCommands/Road/CreateRoad.cpp 594–1105), item 5.9f3 — the
// game command towns grow their roads with. Ported in parts; this is its
// start: the arguments, the road extras that fit the piece, and the cost.

// roadPlacementArgs is RoadPlacementArgs.
type roadPlacementArgs struct {
	X, Y, Z      int32
	Rotation     uint8 // 0–3
	RoadID       uint8 // 0–15
	Mods         uint8 // the road extras asked for (bits 0, 1)
	Bridge       uint8 // a bridge object slot, 0xFF none
	RoadObjectID uint8
	UnkFlags     uint8
}

// createRoadCost is createRoad's start: the road extras asked for that fit
// this piece (the piece's compatibleFlags all in the extra's roadPieces),
// whether the height is valid (a multiple of 16), and the cost — the road's
// and each valid extra's inflation-adjusted build cost (divisor 10) times the
// piece's cost factor / 256.
func (m *Map) createRoadCost(c *Cmd, a roadPlacementArgs) (validMods uint8, cost int32, ok bool) {
	road := c.roadObject(a.RoadObjectID)
	if road == nil || int(a.RoadID) >= len(RoadCompatibleFlags) {
		return 0, 0, false
	}
	compat := RoadCompatibleFlags[a.RoadID]
	validMods = a.Mods
	extra := func(i int) (pieces uint16, buildCost int16, costIndex uint8, found bool) {
		if i >= len(road.ModNames) || c.Objects == nil {
			return 0, 0, 0, false
		}
		if ex := c.Objects.Extra(road.ModNames[i]); ex != nil {
			return ex.Pieces, ex.BuildCostFactor, ex.CostIndex, true
		}
		return 0, 0, 0, false
	}
	for i := range 2 {
		if a.Mods&(1<<i) == 0 {
			continue
		}
		if pieces, _, _, found := extra(i); !found || pieces&compat != compat {
			validMods &^= 1 << i
		}
	}
	if a.Z&0xF != 0 {
		return validMods, 0, false
	}
	factor := int32(RoadCostFactor[a.RoadID])
	cost = m.InflationAdjustedCost(road.BuildCostFactor, road.CostIndex, 10) * factor / 256
	for i := range 2 {
		if validMods&(1<<i) != 0 {
			_, bc, ci, _ := extra(i)
			cost += m.InflationAdjustedCost(bc, ci, 10) * factor / 256
		}
	}
	return validMods, cost, true
}

// roadPiecePlan is one piece tile createRoadCheck cleared: what the apply
// step (0x00476408 on) places there.
type roadPiecePlan struct {
	X, Y, Z            int32
	Index              uint8
	BaseZ, ClearZ      uint8
	Quarter            uint8 // the rotated subTileClearance
	HasBridge          bool  // flags_1136073 bit 1 for this tile
	HasLevelCrossing   bool
	LevelCrossingObjID uint8
}

// createRoadResult is what createRoadCheck found: the cost, each piece's
// plan, the road station met (if any), and the junction bits (roadIdUnk).
type createRoadResult struct {
	Cost       int32
	ValidMods  uint8
	Pieces     []roadPiecePlan
	HasStation bool
	StationID  uint16
	RoadIDUnk  [16]uint8
	State      roadCmdState
}

// createRoadCheck is createRoad's checks (CreateRoad.cpp 594–803): the cost
// and, piece by piece, the height (not below 16, clearZ not above 236), water
// and the bridge a raised piece needs (one asked for, high enough, suitable
// for the piece), the clearance with the road clear function (level
// crossings, road stations, junctions; their costs), and the position rules
// (not partly above / below ground, not under water or too close to its
// surface). It changes nothing unless a.Flags applies (the clear functions
// remove trees / buildings then). The pieces' insertion is the apply step.
// PARTIAL: company observation / emotion events; the error texts.
func (m *Map) createRoadCheck(c *Cmd, a roadPlacementArgs, flags uint8) (createRoadResult, bool) {
	var r createRoadResult
	st := &r.State
	st.Bridge, st.AltRoadObject = 0xFF, 0xFF
	validMods, cost, ok := m.createRoadCost(c, a)
	if !ok {
		return r, false
	}
	r.ValidMods, r.Cost = validMods, cost
	r.StationID = 0xFFFF
	r.RoadIDUnk[a.RoadID] |= 1 << a.Rotation
	pieces := RoadPieces[a.RoadID]
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
				if b == nil || heightDiff > int(b.MaxHeight) { // too far above ground for this bridge
					return r, false
				}
				st.MaxBridgeDiff = max(st.MaxBridgeDiff, uint8(heightDiff))
				if b.DisabledTrackCfg&RoadMiscFlags[a.RoadID] != 0 { // unsuitable for this configuration
					return r, false
				}
				clearZ += int(b.ClearHeight) / 4
			}
		}
		if clearZ > 236 { // error_too_high
			return r, false
		}
		plan := roadPiecePlan{X: x, Y: y, Z: z, Index: piece.Index, BaseZ: baseZ, ClearZ: uint8(clearZ), Quarter: quarter, LevelCrossingObjID: 0xFF}
		args := roadClearArgs{X: x, Y: y, Z: z, Rotation: a.Rotation, RoadID: a.RoadID, BridgeID: a.Bridge, RoadObjectID: a.RoadObjectID,
			Index: piece.Index, IsLastIndex: int(piece.Index) == len(pieces)-1, Flags: flags, UnkFlags: a.UnkFlags}
		clear := func(t *Tile, k int) ClearResult {
			res := m.roadClearFunction(c, st, t, k, &r.Cost, removed, args)
			plan.HasLevelCrossing = plan.HasLevelCrossing || res.HasLevelCrossing
			r.HasStation = r.HasStation || res.HasStation
			if res.LevelCrossingObjID != 0xFF {
				plan.LevelCrossingObjID = res.LevelCrossingObjID
			}
			if res.StationID != 0xFFFF {
				r.StationID = res.StationID
			}
			for i := range r.RoadIDUnk {
				r.RoadIDUnk[i] |= res.RoadIDUnk[i]
			}
			return res.Res
		}
		ok, pos := m.ApplyClearAtStandardHeight(x, y, baseZ, uint8(clearZ), quarter&0xF, quarter>>4, clear)
		if !ok {
			return r, false
		}
		if plan.LevelCrossingObjID == 0xFF {
			plan.LevelCrossingObjID = m.DefaultLevelCrossingType
		}
		if plan.HasLevelCrossing && c.Objects != nil && int(plan.LevelCrossingObjID) < len(c.Objects.LevelCrossingObjects) {
			if lc := c.Objects.LevelCrossingObjects[plan.LevelCrossingObjID]; lc != nil {
				r.Cost += m.InflationAdjustedCost(lc.CostFactor, lc.CostIndex, 10)
			}
		}
		ground := pos & (PosAboveGround | PosUnderground)
		if st.PosFlags != 0 && st.PosFlags&ground == 0 { // partly above, partly below ground
			return r, false
		}
		st.PosFlags = ground
		if pos&(PosPartiallyUnderwater|PosUnderwater) != 0 { // under water / too close to its surface
			return r, false
		}
		plan.HasBridge = st.Flags73&(1<<1) != 0
		r.Pieces = append(r.Pieces, plan)
	}
	return r, true
}

// validCoords is World::validCoords: inside the map.
func (m *Map) validCoords(x, y int32) bool {
	return x >= 0 && y >= 0 && x < int32(m.Width*tileSize) && y < int32(m.Height*tileSize)
}
