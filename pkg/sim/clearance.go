package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Tile clearance — Map/TileClearance.cpp canConstructAtWithClear (the
// standard and any-height collision types, a clear function, the position
// flags, the error texts it sets: setCollisionErrorMessage, off_edge_of_map,
// cannot_build_partly_above_below_water, another_company_is_about_to_build_here)
// and canConstructAt (no clear function: any collision fails). The command
// whose error text it sets is c; nil for the callers that are no command (a
// tree's growth, a town's building search): no text is kept.

const (
	elemFlagGhost       = 1 << 4
	elemFlagAiAllocated = 1 << 5
)

// clearZOf is an element's clearZ (the surface's for KindSurface).
func (t *Tile) clearZOf(r scenario.ElemRef) uint8 {
	switch r.Kind {
	case scenario.KindTrack:
		return t.Tracks[r.Index].ClearZ
	case scenario.KindStation:
		return t.Stations[r.Index].ClearZ
	case scenario.KindSignal:
		return t.Signals[r.Index].ClearZ
	case scenario.KindBuilding:
		return t.Buildings[r.Index].ClearZ
	case scenario.KindTree:
		return t.Trees[r.Index].ClearZ
	case scenario.KindWall:
		return t.Walls[r.Index].ClearZ
	case scenario.KindRoad:
		return t.Roads[r.Index].ClearZ
	case scenario.KindIndustry:
		return t.Industries[r.Index].ClearZ
	}
	return t.SurfaceClearZ()
}

// ClearResult is TileClearance's ClearFuncResult.
type ClearResult uint8

const (
	ClearAllCollisionsRemoved ClearResult = iota
	ClearCollision
	ClearCollisionErrorSet
	ClearNoCollision
	ClearCollisionRemoved
)

func isCollisionResult(r ClearResult) bool {
	return r == ClearCollision || r == ClearCollisionErrorSet
}

// ClearFunc is the clear function a game command passes: called on a
// colliding element (t.Order[k]); it may remove it (ClearCollisionRemoved —
// the tile is then scanned again from the start).
type ClearFunc func(t *Tile, k int) ClearResult

// ElementPositionFlags (TileManager.h).
const (
	PosAboveGround         = 1 << 0
	PosUnderground         = 1 << 1
	PosPartiallyUnderwater = 1 << 2
	PosUnderwater          = 1 << 3
)

// companyFlagUnk1 is CompanyFlags::unk1 (companyFlagUnk2, challenge.go: set
// while an AI converts its aiAllocated items into real ones).
const companyFlagUnk1 = 1 << 1

// companyNull is CompanyId::null.
const companyNull = 0xFF

// tileOwner is TileManager::getTileOwner: a track's, road's or airport's /
// dock's owner; a train or road station's is the track or road element
// before it, a signal's the track before it; else none (companyNull).
func (t *Tile) tileOwner(k int) uint8 {
	r := t.Order[k]
	prev := func(kind uint8) (scenario.ElemRef, bool) {
		if k == 0 || t.Order[k-1].Kind != kind {
			return scenario.ElemRef{}, false
		}
		return t.Order[k-1], true
	}
	switch r.Kind {
	case scenario.KindTrack:
		return t.Tracks[r.Index].Owner
	case scenario.KindRoad:
		return t.Roads[r.Index].Owner
	case scenario.KindStation:
		switch st := &t.Stations[r.Index]; st.StationType {
		case 0: // trainStation
			if p, ok := prev(scenario.KindTrack); ok {
				return t.Tracks[p.Index].Owner
			}
		case 1: // roadStation
			if p, ok := prev(scenario.KindRoad); ok {
				return t.Roads[p.Index].Owner
			}
		default:
			return st.Owner
		}
	case scenario.KindSignal:
		if p, ok := prev(scenario.KindTrack); ok {
			return t.Tracks[p.Index].Owner
		}
	}
	return companyNull
}

// aiCompanyAboutToBuildCheck is TileClearance's (0x00462BB3) on the element
// k: the owner of what an AI has allocated there — on the surface, of the
// first allocated element above it with an owner. A neutral or non-player
// company may not build there; a player company may build over another
// player's or its own (a collision, as any) and over an AI's allocation
// unless that AI is already converting it (CompanyFlags::unk2) — the AI is
// told (unk1, upstream's modification from vanilla).
func (m *Map) aiCompanyAboutToBuildCheck(c *Cmd, t *Tile, k int) ClearResult {
	about := func() ClearResult {
		c.setErrorText(i18n.AnotherCompanyIsAboutToBuildHere)
		return ClearCollisionErrorSet
	}
	ownerResult := func(owner uint8) ClearResult {
		updating := CompanyNeutral
		if c != nil {
			updating = int(c.Company)
		}
		if updating == CompanyNeutral || !c.isPlayer(uint8(updating)) {
			return about()
		}
		if owner == uint8(updating) || c.isPlayer(owner) {
			return ClearCollision
		}
		if int(owner) < len(m.CompanyFaces) {
			f := &m.CompanyFaces[owner]
			if f.ChallengeFlags&companyFlagUnk2 != 0 {
				return about()
			}
			f.ChallengeFlags |= companyFlagUnk1
		}
		return ClearNoCollision
	}
	if t.Order[k].Kind != scenario.KindSurface {
		owner := t.tileOwner(k)
		if owner == companyNull {
			return ClearNoCollision
		}
		return ownerResult(owner)
	}
	for j := k + 1; j < len(t.Order); j++ {
		if t.Order[j].Flags&elemFlagAiAllocated == 0 {
			continue
		}
		if owner := t.tileOwner(j); owner != companyNull {
			return ownerResult(owner)
		}
	}
	return about()
}

// setErrorText is GameCommands::setErrorText for the command (none: c nil).
func (c *Cmd) setErrorText(text int) {
	if c != nil {
		c.ErrorText, c.ErrorArg = text, ""
	}
}

func callClear(t *Tile, k int, clear ClearFunc) ClearResult {
	if clear == nil {
		return ClearCollision
	}
	return clear(t, k)
}

// canConstructAtWithClear is TileClearance's canConstructAtWithClear: ok,
// and the position flags it leaves (getPositionFlags). Refused, the
// command's error text says why: a collision names what is in the way
// (setCollisionErrorMessage); a collisionErrorSet keeps the text the clear
// function or the checks set.
func (m *Map) canConstructAtWithClear(c *Cmd, px, py int32, baseZ, clearZ, baseQuarter, zQuarter uint8, anyHeight bool, clear ClearFunc) (bool, uint8) {
	var pos uint8 = PosAboveGround
	drawable := func(v int32) bool { return v >= tileSize && v < int32(m.Width*tileSize)-tileSize-1 }
	if !drawable(px) || !drawable(py) {
		c.setErrorText(i18n.OffEdgeOfMap)
		return false, pos
	}
	t := m.TileAt(int(px/tileSize), int(py/tileSize))
	for {
		removed := false
		for k := 0; k < len(t.Order); k++ {
			var res ClearResult
			if t.Order[k].Kind == scenario.KindSurface {
				res = m.checkSurface(c, t, k, baseZ, clearZ, baseQuarter, zQuarter, clear, &pos)
			} else {
				res = m.checkNonSurface(c, t, k, baseZ, clearZ, baseQuarter, anyHeight, clear)
			}
			switch res {
			case ClearAllCollisionsRemoved:
				return true, pos
			case ClearCollision:
				m.setCollisionErrorMessage(c, t, k)
				return false, pos
			case ClearCollisionErrorSet:
				return false, pos
			case ClearCollisionRemoved:
				removed = true
			}
			if removed {
				break
			}
		}
		if !removed {
			return true, pos
		}
	}
}

// ApplyClearAtStandardHeight / ApplyClearAtAllHeights are TileClearance's.
func (m *Map) ApplyClearAtStandardHeight(c *Cmd, px, py int32, baseZ, clearZ, baseQuarter, zQuarter uint8, clear ClearFunc) (bool, uint8) {
	return m.canConstructAtWithClear(c, px, py, baseZ, clearZ, baseQuarter, zQuarter, false, clear)
}

func (m *Map) ApplyClearAtAllHeights(c *Cmd, px, py int32, baseZ, clearZ, baseQuarter, zQuarter uint8, clear ClearFunc) (bool, uint8) {
	return m.canConstructAtWithClear(c, px, py, baseZ, clearZ, baseQuarter, zQuarter, true, clear)
}

// CanConstructAt is TileClearance::canConstructAt: whether something from
// baseZ to clearZ over the quarters baseQuarter (z quarters zQuarter) fits on
// the tile at world (px, py).
func (m *Map) CanConstructAt(c *Cmd, px, py int32, baseZ, clearZ uint8, baseQuarter, zQuarter uint8) bool {
	ok, _ := m.canConstructAtWithClear(c, px, py, baseZ, clearZ, baseQuarter, zQuarter, false, nil)
	return ok
}

// checkNonSurface is canConstructAtCheckNonSurfaceElement.
func (m *Map) checkNonSurface(c *Cmd, t *Tile, k int, baseZ, clearZ, baseQuarter uint8, anyHeight bool, clear ClearFunc) ClearResult {
	r := t.Order[k]
	if anyHeight {
		if r.Kind == scenario.KindTree || r.Kind == scenario.KindBuilding || r.Kind == scenario.KindIndustry {
			return callClear(t, k, clear)
		}
		if r.Kind == scenario.KindStation {
			if st := t.Stations[r.Index].StationType; st == 2 || st == 3 {
				return callClear(t, k, clear)
			}
		}
	}
	if baseZ >= t.clearZOf(r) || clearZ <= t.baseZOf(r) || r.Flags&elemFlagGhost != 0 || r.Flags&0x0F&baseQuarter == 0 {
		return ClearNoCollision
	}
	if r.Flags&elemFlagAiAllocated == 0 {
		return callClear(t, k, clear)
	}
	if clear != nil {
		if res := clear(t, k); !isCollisionResult(res) {
			return res
		}
	}
	return m.aiCompanyAboutToBuildCheck(c, t, k)
}

// checkSurface is canConstructAtCheckSurfaceElement.
func (m *Map) checkSurface(c *Cmd, t *Tile, k int, baseZ, clearZ, baseQuarter, zQuarter uint8, clear ClearFunc, pos *uint8) ClearResult {
	r := t.Order[k]
	if r.Flags&elemFlagAiAllocated != 0 {
		if res := m.aiCompanyAboutToBuildCheck(c, t, k); res != ClearNoCollision {
			return res
		}
	}
	if waterZ := int(t.WaterLevel&0x1F) * microToSmallZStep; waterZ != 0 {
		if clearZ > t.SurfaceClearZ() && int(baseZ) < waterZ+4 {
			*pos |= PosPartiallyUnderwater
			if int(baseZ) < waterZ {
				*pos |= PosUnderwater
				if int(clearZ) > waterZ {
					if clear == nil {
						c.setErrorText(i18n.CannotBuildPartlyAboveBelowWater)
						return ClearCollisionErrorSet
					}
					if res := clear(t, k); isCollisionResult(res) {
						if res == ClearCollision {
							c.setErrorText(i18n.CannotBuildPartlyAboveBelowWater)
						}
						return ClearCollisionErrorSet
					}
				}
			}
		}
	}
	if zQuarter == 0x0F {
		return ClearNoCollision
	}
	if clearZ <= t.BaseZ {
		*pos |= PosUnderground
		*pos &^= PosAboveGround
		return ClearNoCollision
	}
	if surfaceCornersFit(t, baseZ, baseQuarter, zQuarter) {
		return ClearNoCollision
	}
	return callClear(t, k, clear)
}

// surfaceCornersFit is the slope-corner part of canConstructAtCheckSurfaceElement.
func surfaceCornersFit(t *Tile, baseZ, baseQuarter, zQuarter uint8) bool {
	slope := t.Slope & 0x1F
	n, e, s, w := t.BaseZ, t.BaseZ, t.BaseZ, t.BaseZ
	if slope&slopeCornerNorth != 0 {
		n += smallZStep
		if slope == slopeDownSouth|slopeDoubleHeight {
			n += smallZStep
		}
	}
	if slope&slopeCornerEast != 0 {
		e += smallZStep
		if slope == slopeDownWest|slopeDoubleHeight {
			e += smallZStep
		}
	}
	if slope&slopeCornerSouth != 0 {
		s += smallZStep
		if slope == slopeDownNorth|slopeDoubleHeight {
			s += smallZStep
		}
	}
	if slope&slopeCornerWest != 0 {
		w += smallZStep
		if slope == slopeDownEast|slopeDoubleHeight {
			w += smallZStep
		}
	}
	dbl := int(baseZ) + 8
	ok := func(bit uint8, cz uint8) bool {
		return baseQuarter&bit == 0 || ((zQuarter&bit != 0 || baseZ >= cz) && dbl >= int(cz))
	}
	return ok(1, n) && ok(2, e) && ok(4, s) && ok(8, w)
}

// MaxHeight is PaintSession::setMaxHeight's _maxHeight for this tile: the
// highest clearZ of its elements (water, and an industrial surface + 24,
// count too), × kSmallZStep, + 32 — how far the tile can reach up the
// screen (the paint's vertical cull).
func (t *Tile) MaxHeight() int {
	var maxClearZ uint8
	for _, r := range t.Order {
		maxClearZ = max(maxClearZ, t.clearZOf(r))
		if r.Kind != scenario.KindSurface {
			continue
		}
		if t.WaterLevel != 0 {
			maxClearZ = max(maxClearZ, uint8(int(t.WaterLevel)*microToSmallZStep))
		}
		if t.IsIndustrial {
			maxClearZ = max(maxClearZ, t.SurfaceClearZ()+24)
		}
	}
	return int(maxClearZ)*smallZStep + 32
}

// setCollisionErrorMessage is TileClearance's (0x00462C8E): the command's
// error text for what is in the way (the element k of the tile), with the
// name it carries.
func (m *Map) setCollisionErrorMessage(c *Cmd, t *Tile, k int) {
	if c == nil {
		return
	}
	r := t.Order[k]
	in := func(name string) { c.ErrorText, c.ErrorArg = i18n.StringIdInTheWay, name }
	om := c.Objects
	switch r.Kind {
	case scenario.KindSurface:
		c.setErrorText(i18n.RaiseOrLowerLandFirst)
	case scenario.KindTrack:
		if o := c.trackObject(t.Tracks[r.Index].TrackObjectID); o != nil {
			in(o.Name)
		}
	case scenario.KindStation:
		names := [4]int{i18n.CaptStation, i18n.CaptStation, i18n.CaptAirport, i18n.CaptShipPort}
		if st := t.Stations[r.Index].StationType; int(st) < len(names) {
			in(i18n.Label(names[st]))
		}
	case scenario.KindSignal:
		in(i18n.Label(i18n.CaptSignal))
	case scenario.KindBuilding:
		if om != nil && int(t.Buildings[r.Index].ObjectID) < len(om.BuildingObjects) && om.BuildingObjects[t.Buildings[r.Index].ObjectID] != nil {
			in(om.BuildingObjects[t.Buildings[r.Index].ObjectID].Name)
		}
	case scenario.KindTree:
		if om != nil && int(t.Trees[r.Index].TreeObjectID) < len(om.TreeObjects) && om.TreeObjects[t.Trees[r.Index].TreeObjectID] != nil {
			in(om.TreeObjects[t.Trees[r.Index].TreeObjectID].Name)
		}
	case scenario.KindWall:
		c.setErrorText(i18n.ObjectInTheWay)
	case scenario.KindRoad:
		if o := c.roadObject(t.Roads[r.Index].RoadObjectID); o != nil {
			in(o.Name)
		}
	case scenario.KindIndustry: // the industry's name, formatted with its town
		for i := range m.Industries {
			if m.Industries[i].ID != t.Industries[r.Index].IndustryID {
				continue
			}
			if name, ok := m.IndustryName(i, om); ok {
				in(name)
			}
			break
		}
	}
}
