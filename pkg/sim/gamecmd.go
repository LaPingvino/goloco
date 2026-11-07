package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// GameCommands::Flags.
const (
	CmdApply                   = 1 << 0
	CmdPreventBuildingClearing = 1 << 1
	CmdAllowNegativeCashFlow   = 1 << 2
	CmdNoErrorWindow           = 1 << 3
	CmdAiAllocated             = 1 << 4
	CmdNoPayment               = 1 << 5
	CmdGhost                   = 1 << 6
	CmdFlag7                   = 1 << 7
)

// CompanyNeutral is CompanyId::neutral.
const CompanyNeutral = 15

// Cmd is what a game command runs with besides its arguments: the updating
// company (GameCommands::getUpdatingCompanyId — neutral for towns, the
// generator and the editor) and gPrng2 (sounds draw from it).
type Cmd struct {
	Company uint8
	Rng2    *core.Prng
	Objects *objects.ObjectManager
	// Editor and Sandbox are SceneManager::isEditorMode / isSandboxMode.
	Editor, Sandbox bool
	// Players reports CompanyManager::isPlayerCompany.
	Players func(company uint8) bool
	// Pos is GameCommands::setPosition's — where a command's money effect
	// shows; nil: Location::null (spendMoneyEffect's viewport-centre case,
	// PARTIAL in the sim). Commands set it as upstream's do.
	Pos *[3]int32
}

func (c *Cmd) isPlayer(company uint8) bool { return c.Players != nil && c.Players(company) }

// TileClearance::clearTreeCollision: the tree's clear cost (inflation
// adjusted, divisor 12); applied (not ghost / AI allocated) the tree is
// removed (TileManager::removeTree).
func (m *Map) clearTreeCollision(c *Cmd, t *Tile, k int, px, py int32, flags uint8, cost *int32) ClearResult {
	tr := t.Trees[t.Order[k].Index]
	var obj *objects.TreeObject
	if int(tr.TreeObjectID) < len(c.Objects.TreeObjects) {
		obj = c.Objects.TreeObjects[tr.TreeObjectID]
	}
	if obj != nil {
		*cost += m.InflationAdjustedCost(obj.ClearCostFactor, obj.CostIndex, 12)
	}
	if flags&(CmdGhost|CmdAiAllocated) != 0 || flags&CmdApply == 0 {
		return ClearNoCollision
	}
	last := k == len(t.Order)-1
	m.removeTree(c, t, k, px, py, CmdApply, obj)
	if last {
		return ClearAllCollisionsRemoved
	}
	return ClearCollisionRemoved
}

// removeTree is TileManager::removeTree: the demolish sound (a gPrng2 draw —
// randNext(20003, 24098) — when a company, not the neutral one, removes a
// real tree), the town rating reduction (updateTownInfo: companies only), and
// the element removed.
func (m *Map) removeTree(c *Cmd, t *Tile, k int, px, py int32, flags uint8, obj *objects.TreeObject) {
	r := t.Order[k]
	if r.Flags&(elemFlagGhost|elemFlagAiAllocated) == 0 && c.Company != CompanyNeutral && c.Rng2 != nil {
		c.Rng2.RandNextRange(20003, 24098) // playDemolishTreeSound's frequency
	}
	if flags&CmdGhost == 0 {
		var rating int16
		if obj != nil {
			rating = obj.DemolishRatingReduction
		}
		m.updateTownInfo(px, py, 0, 0, rating, 0, c)
	}
	idx := int(r.Index)
	m.invalidateTile(px, py, int32(t.Trees[idx].BaseZ)*4, int32(t.Trees[idx].ClearZ)*4, ZoomEighth, 56) // ViewportManager::invalidate(pos, zMin, zMax, eighth, 56)
	t.Trees = append(t.Trees[:idx], t.Trees[idx+1:]...)
	t.removeRef(scenario.KindTree, idx)
}

// tileClearFunction is TileClearance's tileClearFunction (clearWithDefaultCollision
// / clearWithoutDefaultCollision): surfaces pass, trees are cleared, buildings
// removed, anything else a collision or not as asked.
func (m *Map) tileClearFunction(c *Cmd, t *Tile, k int, px, py int32, removed RemovedBuildings, flags uint8, cost *int32, defaultCollision bool) ClearResult {
	switch t.Order[k].Kind {
	case scenario.KindSurface:
		return ClearNoCollision
	case scenario.KindTree:
		return m.clearTreeCollision(c, t, k, px, py, flags, cost)
	case scenario.KindBuilding:
		return m.clearBuildingCollision(c, t, k, px, py, removed, flags, cost)
	}
	if defaultCollision {
		return ClearCollision
	}
	return ClearNoCollision
}

// isBridgeRequired is CreateRoad.cpp's isBridgeRequired: above the surface
// always; level with it unless the piece is flat-capable (no unk4 preview
// flag) and its corner flags are none or the surface's slope corners.
func isBridgeRequired(baseZ uint8, t *Tile, pieceUnk4 bool, unk uint8) bool {
	if baseZ > t.BaseZ {
		return true
	}
	if baseZ == t.BaseZ {
		if !pieceUnk4 && (unk == 0 || unk == t.Slope&0x0F) {
			return false
		}
		return true
	}
	return false
}

// RemovedBuildings is TileClearance::RemovedBuildings: the start (x, y, z) of
// each building a command has already removed.
type RemovedBuildings map[[3]int32]bool

// buildingTileOffsets is getBuildingTileOffsets (Industry.cpp): one tile, or
// the four of a large one (World::kOffsets order).
func buildingTileOffsets(large bool) [][2]int32 {
	if large {
		return [][2]int32{{0, 0}, {0, 32}, {32, 32}, {32, 0}}
	}
	return [][2]int32{{0, 0}}
}

// clearBuildingCollision is TileClearance::clearBuildingCollision: a
// headquarters is a collision; any other building is removed once (by its
// start tile) through the removeBuilding command, its cost added.
func (m *Map) clearBuildingCollision(c *Cmd, t *Tile, k int, px, py int32, removed RemovedBuildings, flags uint8, cost *int32) ClearResult {
	b := t.Buildings[t.Order[k].Index]
	obj := c.Objects.GetBuildingObjectByIndex(int(b.ObjectID))
	if obj != nil && obj.Flags&objects.BuildingFlagIsHeadquarters != 0 {
		return ClearCollision
	}
	off := buildingTileOffsets(true)[b.SequenceIndex&3]
	start := [3]int32{px - off[0], py - off[1], int32(b.BaseZ) * 4}
	if removed[start] {
		return ClearNoCollision
	}
	removed[start] = true
	rf := flags
	if flags&CmdApply != 0 || len(removed) != 1 {
		rf |= CmdFlag7
	}
	if flags&(CmdGhost|CmdAiAllocated) != 0 {
		rf &^= CmdGhost | CmdAiAllocated | CmdApply
	}
	last := k == len(t.Order)-1 // wasRemoveOnLastElement (the element checked is the one removed)
	bc, ok := m.RemoveBuilding(c, start[0], start[1], start[2], rf)
	if !ok {
		return ClearCollisionErrorSet
	}
	*cost += bc
	if flags&CmdApply == 0 || flags&(CmdGhost|CmdAiAllocated) != 0 {
		return ClearNoCollision
	}
	if last {
		return ClearAllCollisionsRemoved
	}
	return ClearCollisionRemoved
}

// RemoveBuilding is the removeBuilding game command (RemoveBuilding.cpp): the
// building at (x, y, z) on the tile — refused, outside the editor and unless
// ghost / flag 7, when indestructible (a player in sandbox excepted) or, for
// a town building, when the company's rating in the closest town is below 0
// — its tiles' building elements at that height removed when applied, and
// its clear cost (inflation adjusted, divisor 8). ok false on failure.
//
// PARTIAL: createDestructExplosion (a company's real building) draws gPrng2
// for the sound here but does not create the ExplosionSmoke entity yet.
func (m *Map) RemoveBuilding(c *Cmd, x, y, z int32, flags uint8) (int32, bool) {
	t := m.TileAt(int(x/tileSize), int(y/tileSize))
	if t == nil {
		return 0, false
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindBuilding || int32(t.Buildings[r.Index].BaseZ) != z/4 {
			continue
		}
		obj := c.Objects.GetBuildingObjectByIndex(int(t.Buildings[r.Index].ObjectID))
		if obj == nil {
			return 0, false
		}
		if !c.Editor && flags&(CmdGhost|CmdFlag7) == 0 {
			if !(c.Sandbox && c.isPlayer(c.Company)) && obj.Flags&objects.BuildingFlagIndestructible != 0 {
				return 0, false
			}
			if obj.Flags&objects.BuildingFlagMiscBuilding == 0 && c.Company != CompanyNeutral {
				if i, _, ok := m.ClosestTownAndDensity(x, y); ok && m.Towns[i].CompanyRatings[c.Company] < 0 {
					return 0, false
				}
			}
		}
		for _, o := range buildingTileOffsets(obj.Flags&objects.BuildingFlagLargeTile != 0) {
			sx, sy := x+o[0], y+o[1]
			st := m.TileAt(int(sx/tileSize), int(sy/tileSize))
			if st == nil {
				continue
			}
			// upstream's loop moves on after a removal (the next element slid
			// into the removed one's place is not looked at), as here.
			for k := 0; k < len(st.Order); k++ {
				sr := st.Order[k]
				if sr.Kind != scenario.KindBuilding || int32(st.Buildings[sr.Index].BaseZ) != z/4 {
					continue
				}
				if flags&CmdApply != 0 {
					m.removeBuildingElement(c, st, k, sx, sy)
				}
			}
		}
		return m.InflationAdjustedCost(obj.ClearCostFactor, obj.ClearCostIndex, 8), true
	}
	return 0, false
}

// removeBuildingElement is TileManager::removeBuildingElement: a company's
// real building explodes (see RemoveBuilding's PARTIAL); the first tile of a
// real town building takes its capacity (and, built, its population), the
// rating reduction, one building and its amenity from the closest town; the
// element goes.
func (m *Map) removeBuildingElement(c *Cmd, t *Tile, k int, px, py int32) {
	r := t.Order[k]
	b := t.Buildings[r.Index]
	if r.Flags&(elemFlagGhost|elemFlagAiAllocated) == 0 && c.Company != CompanyNeutral && c.Rng2 != nil {
		c.Rng2.RandNextRange(20003, 24098) // createDestructExplosion's sound frequency
	}
	if b.SequenceIndex == 0 && r.Flags&elemFlagGhost == 0 {
		if obj := c.Objects.GetBuildingObjectByIndex(int(b.ObjectID)); obj != nil && obj.Flags&objects.BuildingFlagMiscBuilding == 0 {
			capacity := uint32(-int32(obj.ProducedQuantity[0]))
			pop := capacity
			if !b.IsConstructed {
				pop = 0
			}
			if i := m.updateTownInfo(px, py, pop, capacity, obj.DemolishRatingReduction, -1, c); i >= 0 && obj.TownAmenityCategory != 0xFF {
				m.Towns[i].AmenityCounts[obj.TownAmenityCategory&7]--
			}
		}
	}
	idx := int(r.Index)
	m.invalidateTile(px, py, int32(t.Buildings[idx].BaseZ)*4, int32(t.Buildings[idx].ClearZ)*4, ZoomEighth, defaultRadius)
	t.Buildings = append(t.Buildings[:idx], t.Buildings[idx+1:]...)
	t.removeRef(scenario.KindBuilding, idx)
}

// doCommand is GameCommands::doCommand's loc_4313C6 for a command run from
// the sim (towns, as the neutral company): the command first without apply —
// a failure ends it there —, then, when applying, again with apply. The
// query pass is where a command refuses (an indestructible building, a town's
// permission: the clear functions only skip those checks while applying), so
// applying without it would build where upstream refuses.
// PARTIAL: ensureCompanyFunding and the payment (the neutral company pays
// nothing), the editor's zero cost, the error texts and sounds.
func doCommand(flags uint8, run func(flags uint8) (int32, bool)) (int32, bool) {
	cost, ok := run(flags &^ CmdApply)
	if !ok || flags&CmdApply == 0 {
		return cost, ok
	}
	cost2, ok := run(flags)
	if !ok {
		return 0, false
	}
	return min(cost, cost2), true
}

// ExpenditureConstruction is ExpenditureType::Construction (createTrack's
// setExpenditureType).
const ExpenditureConstruction = 12

// ExpenditureVehiclePurchases is ExpenditureType::VehiclePurchases (the
// createVehicle command's setExpenditureType).
const ExpenditureVehiclePurchases = 13

// ExpenditureVehicleDisposals is ExpenditureType::VehicleDisposals (the
// sellVehicle command's).
const ExpenditureVehicleDisposals = 14

// DoCommandFor is GameCommands::doCommand's loc_4313C6 for a command a
// company runs at the top level (nest level 1), with the payment: the query
// pass (free in the editor), ensureCompanyFunding unless AllowNegativeCashFlow
// or NoPayment (a refusal ends it), then the applying pass, the lower of the
// two costs, applyPaymentToCompany under the expenditure type and, for the
// controlling company, the flying cost text (spendMoneyEffect at the
// command's position + 24 up; Cmd.Pos). PARTIAL: the error window texts /
// sounds.
func (m *Map) DoCommandFor(c *Cmd, flags uint8, expenditure uint8, run func(flags uint8) (int32, bool)) (int32, bool) {
	cost, ok := run(flags &^ CmdApply)
	if !ok {
		return 0, false
	}
	if c.Editor {
		cost = 0
	}
	if flags&(CmdAllowNegativeCashFlow|CmdNoPayment) == 0 && cost != 0 && !c.Editor {
		if !m.ensureCompanyFunding(int(c.Company), cost) {
			return 0, false
		}
	}
	if flags&CmdApply == 0 {
		return cost, true
	}
	cost2, ok := run(flags)
	if !ok {
		return 0, false
	}
	if c.Editor {
		cost = 0
	}
	cost = min(cost, cost2)
	if flags&CmdNoPayment == 0 {
		m.ApplyPaymentToCompany(c.Company, cost, expenditure)
		if cost != 0 && c.Company == m.ControllingCompany && c.Pos != nil { // the flying cost text
			m.SpendMoneyEffect(c.Pos[0], c.Pos[1], c.Pos[2]+24, c.Company, cost)
		}
	}
	return cost, true
}
