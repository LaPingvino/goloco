package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

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

// createRoadApply is createRoad's apply step (CreateRoad.cpp 776–865), on
// the plan createRoadCheck made with the same flags: per piece the walls on
// the tile below the road removed, the surface's industry and growth / snow
// cleared at the road's height, and the road element inserted — its heights,
// rotation, objects, owner, extras, bridge, level crossing, station flag,
// last-piece, ghost and AI flags.
// PARTIAL: the player company's emotion event; Scenario options'
// madeAnyChanges (not kept); viewport invalidation (the renderer redraws).
func (m *Map) createRoadApply(c *Cmd, a roadPlacementArgs, flags uint8, r *createRoadResult) bool {
	if flags&CmdApply == 0 {
		return true
	}
	ghostOrAI := flags&(CmdGhost|CmdAiAllocated) != 0
	for _, p := range r.Pieces {
		t := m.TileAt(int(p.X/tileSize), int(p.Y/tileSize))
		if t == nil {
			return false
		}
		if !ghostOrAI {
			t.removeAllWallsBelow(p.BaseZ)
			if z := int(t.BaseZ) * smallZStep; z <= int(p.Z) && int(p.Z)-z < 12*smallZStep {
				m.removeSurfaceIndustry(p.X, p.Y, t, c)
				t.setTerrainStyleAsCleared()
			}
		}
		el := scenario.RoadElement{
			BaseZ: p.BaseZ, ClearZ: p.ClearZ, Rotation: a.Rotation, RoadObjectID: a.RoadObjectID,
			SequenceIndex: p.Index, RoadID: a.RoadID, Owner: c.Company,
			Mods:              r.ValidMods & 3,
			BridgeID:          a.Bridge & 7, // setBridgeObjectId keeps 3 bits
			HasBridge:         p.HasBridge,
			HasStationElement: r.HasStation,
			IsLast:            int(p.Index) == len(RoadPieces[a.RoadID])-1,
			Ghost:             flags&CmdGhost != 0,
			AiAllocated:       flags&CmdAiAllocated != 0,
		}
		if m.RoadAnyTypeCompatible&(1<<a.RoadObjectID) != 0 && c.Company != CompanyNeutral {
			el.Mods |= 1 // setUnk7_40: mods bit 0
		}
		if p.HasLevelCrossing && flags&CmdAiAllocated == 0 {
			el.HasLevelCrossing, el.LevelCrossingObjectID = true, p.LevelCrossingObjID
		}
		t.insertRoadElement(el, p.Quarter&0xF)
		if flags&CmdAiAllocated == 0 { // shouldInvalidateTile (Config showAiPlanningAsGhosts: default off)
			m.mapInvalidateTileFull(p.X, p.Y)
		}
	}
	return true
}

// insertRoadElement is TileManager::insertElementRoad: the road goes after
// every element at or below its height, but before a road station at its own
// height; its ref carries the flags byte (quarters, ghost, AI, flag_6) that
// the clearance checks read.
func (t *Tile) insertRoadElement(el scenario.RoadElement, quarters uint8) {
	t.insertRoadElementAt(el, quarters, true)
}

// insertRoadElementAt: beforeRoadStation false is the plain insertElement
// (after everything at or below its height).
func (t *Tile) insertRoadElementAt(el scenario.RoadElement, quarters uint8, beforeRoadStation bool) {
	t.Roads = append(t.Roads, el)
	at := len(t.Order)
	for i, r := range t.Order {
		bz := t.baseZOf(r)
		if el.BaseZ < bz || (beforeRoadStation && bz == el.BaseZ && r.Kind == scenario.KindStation && t.Stations[r.Index].StationType == 1) {
			at = i
			break
		}
	}
	f := quarters & 0xF
	if el.Ghost {
		f |= elemFlagGhost
	}
	if el.AiAllocated {
		f |= elemFlagAiAllocated
	}
	if el.IsLast {
		f |= 1 << 6
	}
	t.Order = append(t.Order, scenario.ElemRef{})
	copy(t.Order[at+1:], t.Order[at:])
	t.Order[at] = scenario.ElemRef{Kind: scenario.KindRoad, Index: uint16(len(t.Roads) - 1), Flags: f}
}

// removeAllWallsBelow is TileManager::removeAllWallsOnTileBelow: every wall
// except those wholly above baseZ + 12 (and reaching above baseZ) is removed.
func (t *Tile) removeAllWallsBelow(baseZ uint8) {
	for i := len(t.Walls) - 1; i >= 0; i-- {
		w := t.Walls[i]
		if baseZ < w.ClearZ && int(baseZ)+12 <= int(w.BaseZ) {
			continue
		}
		t.Walls = append(t.Walls[:i], t.Walls[i+1:]...)
		t.removeRef(scenario.KindWall, i)
	}
}

// removeSurfaceIndustry is SurfaceElement::removeIndustry: an industrial
// surface becomes plain again (growth and variation 0) and the closest town
// loses 30 rating points' worth (updateTownInfo(pos, 0, 0, -30, 0)).
func (m *Map) removeSurfaceIndustry(x, y int32, t *Tile, c *Cmd) {
	if !t.IsIndustrial {
		return
	}
	t.IsIndustrial, t.GrowthStage, t.Variation = false, 0, 0
	m.updateTownInfo(x, y, 0, 0, -30, 0, c)
}

// setTerrainStyleAsCleared is TileManager's: a non-industrial surface's growth
// stage and snow coverage (slope bits 5–7) back to 0.
func (t *Tile) setTerrainStyleAsCleared() {
	if t.IsIndustrial {
		return
	}
	if t.GrowthStage > 0 {
		t.GrowthStage = 0
		t.Slope &^= 0xE0
	}
	t.Slope &^= 0xE0
}

// roadObjFlagUnk02 is RoadObjectFlags::unk_02: the road fills in junctions
// (createRoad's 0x47656B).
const roadObjFlagUnk02 = 1 << 2

// createRoad is the game command (CreateRoad.cpp 594–1105): the checks
// (createRoadCheck), the pieces placed (createRoadApply), the junction pieces
// filled in (createRoadJunctions), then a bridge's and a tunnel's cost. The
// cost, and false where upstream returns kFailure.
// PARTIAL: a road station met is not recalculated (recalculateStationModes /
// Center — not yet ported — then UpdateStationLabel, which is); the
// placement sound.
func (m *Map) createRoad(c *Cmd, a roadPlacementArgs, flags uint8) (int32, bool) {
	r, ok := m.createRoadCheck(c, a, flags)
	c.PosFlags = r.State.PosFlags
	if !ok || !m.createRoadApply(c, a, flags, &r) {
		return 0, false
	}
	road := c.roadObject(a.RoadObjectID)
	st := &r.State
	if flags&CmdApply != 0 && flags&(CmdGhost|CmdAiAllocated) == 0 && r.RoadIDUnk[0]&(1<<7) != 0 && road.Flags&roadObjFlagUnk02 != 0 {
		m.createRoadJunctions(c, a, flags, &r)
	}
	if st.Flags73&(1<<0) != 0 && st.Flags73&(1<<6) == 0 {
		if b := c.bridge(a.Bridge); b != nil {
			height := int32(st.MaxBridgeDiff) * int32(b.HeightCostFactor)
			cost := m.InflationAdjustedCost(int16(int32(b.BaseCostFactor)+height), b.CostIndex, 10) * int32(RoadCostFactor[a.RoadID]) / 256
			if st.Flags73&(1<<7) != 0 {
				cost *= 2
			}
			r.Cost += cost
		}
	}
	if st.PosFlags&PosUnderground != 0 && st.Flags73&(1<<6) == 0 {
		// upstream prices the tunnel with getTrackMiscData(roadId): kept
		r.Cost += m.InflationAdjustedCost(road.TunnelCostFactor, 2, 8) * int32(TrackCostFactor[a.RoadID]) / 256
	}
	return r.Cost, true
}

// createRoadJunctions is createRoad's 0x47656B: with the straight piece's
// junction bit set (roadIdUnk[0] bit 7) and a road that fills junctions,
// for each rotation the missing curve / straight pieces (road ids 1, 2, 0)
// that the met junction needs are placed at the command's position, as
// upstream decides from the gathered bits (roadIdUnk).
func (m *Map) createRoadJunctions(c *Cmd, a roadPlacementArgs, flags uint8, r *createRoadResult) {
	u := &r.RoadIDUnk
	t := m.TileAt(int(a.X/tileSize), int(a.Y/tileSize))
	if t == nil {
		return
	}
	place := func(rotation, roadID uint8) {
		if z := int(t.BaseZ) * smallZStep; z <= int(a.Z) && int(a.Z)-z < 12*smallZStep {
			m.removeSurfaceIndustry(a.X, a.Y, t, c)
			t.setTerrainStyleAsCleared()
		}
		bz := uint8(a.Z / smallZStep)
		el := scenario.RoadElement{BaseZ: bz, ClearZ: bz + 8, Rotation: rotation & 3, RoadObjectID: a.RoadObjectID, RoadID: roadID,
			Owner: c.Company, Mods: r.ValidMods & 3, BridgeID: a.Bridge & 7, HasBridge: r.State.Flags73&(1<<1) != 0, IsLast: true,
			Ghost: flags&CmdGhost != 0, AiAllocated: flags&CmdAiAllocated != 0}
		t.insertRoadElementAt(el, 0xF, false)
	}
	for i := uint8(0); i < 4; i++ {
		f0, f1, f2, f3 := uint8(1)<<i, uint8(1)<<((i+1)&3), uint8(1)<<((i+2)&3), uint8(1)<<((i+3)&3)
		has := func(id int, f uint8) bool { return u[id]&f != 0 }
		left := has(0, f0) && (has(0, f3) || has(0, f1) || has(1, f1) || has(2, f2)) ||
			has(0, f2) && has(0, f3) ||
			has(0, f3) && (has(2, f0) || has(1, f3)) ||
			has(2, f0) && (has(1, f1) || has(2, f2)) ||
			has(1, f3) && (has(1, f1) || has(2, f2)) ||
			has(0, f1) && has(2, f0)
		if left && !has(2, f1) {
			p := !has(1, f0)
			u[1] |= f0
			if p {
				place(i, 1)
			}
		}
		right := has(0, f2) && (has(2, f2) || has(1, f1)) ||
			has(1, f0) && (has(1, f2) || has(2, f3)) ||
			has(2, f1) && (has(1, f2) || has(2, f3)) ||
			has(0, f1) && (has(0, f2) || has(1, f3))
		if right && !has(1, f0) {
			p := !has(2, f1)
			u[2] |= f1
			if p {
				place(i+1, 2)
			}
		}
		straight1 := has(1, f0) && (has(1, f1) || has(2, f2)) ||
			has(2, f0) && (has(1, f1) || has(2, f2)) ||
			has(1, f3) && (has(1, f1) || has(2, f2)) ||
			has(1, f0) && (has(1, f2) || has(2, f3)) ||
			has(2, f1) && (has(1, f2) || has(2, f3)) ||
			has(2, f1) && has(1, f1)
		if straight1 && !has(0, f2) {
			p := !has(0, f0)
			u[0] |= f0
			if p {
				place(i, 0)
			}
		}
		if has(2, f1) && has(2, f2) && !has(0, f2) {
			p := !has(0, f0)
			u[0] |= f0
			if p {
				place(i, 0)
			}
		}
	}
}
