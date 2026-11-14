package sim

import (
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The Terraform window's game commands as a company runs them
// (GameCommands/Terraform/RaiseLand.cpp, LowerLand.cpp, CreateTree.cpp,
// RemoveTree.cpp): the tile-range commands of terraform.go with their
// centre, the tree commands with flags, cost and ghosts.

// LandArgs is RaiseLandArgs / LowerLandArgs: the centre (world units, where
// the cost shows), the tiles a..b and the corner (MapSelectionType: 0–3 a
// corner, 4 the whole tile).
type LandArgs struct {
	CX, CY         int32
	AX, AY, BX, BY int
	Corner         uint8
}

// RaiseLand is the raiseLand game command (0x00463702): raiseLand over the
// range, the position its centre at the land height there. The caller pays
// under Construction. PARTIAL: the construct sound, madeAnyChanges.
func (m *Map) RaiseLand(c *Cmd, flags uint8, a LandArgs) (int32, bool) {
	cost, ok := m.raiseLand(c, a.AX, a.AY, a.BX, a.BY, a.Corner, flags)
	if !ok {
		return 0, false
	}
	land, _ := m.GetHeight(a.CX, a.CY)
	c.Pos = &[3]int32{a.CX, a.CY, int32(land)}
	return cost, true
}

// LowerLand is the lowerLand game command (0x00463751), as RaiseLand.
func (m *Map) LowerLand(c *Cmd, flags uint8, a LandArgs) (int32, bool) {
	cost, ok := m.lowerLand(c, a.AX, a.AY, a.BX, a.BY, a.Corner, flags)
	if !ok {
		return 0, false
	}
	land, _ := m.GetHeight(a.CX, a.CY)
	c.Pos = &[3]int32{a.CX, a.CY, int32(land)}
	return cost, true
}

// TreePlacement is TreePlacementArgs: the tile (world units), the quadrant,
// rotation, tree object and colour; buildImmediately (fully grown, as the
// generator and the editor plant) and requiresFullClearance (nothing above).
type TreePlacement struct {
	X, Y                                    int32
	Quadrant, Rotation, Type, Colour        uint8
	BuildImmediately, RequiresFullClearance bool
}

// CreateTreeCmd is the createTree game command (0x004BB138): refused off the
// map, under water (cant_build_this_underwater), on no-trees land
// (land_type_not_suitable) or without clearance over its quarter; applied,
// the tree element — a sapling (growth 0, the object's initial height)
// unless built immediately, a ghost when asked, else its rating to the
// closest town. The build cost (inflation adjusted, divisor 12) and the
// placed tree's baseZ (Terraform::setLastPlacedTree). The position the
// tile's centre at its land height; the caller pays under Construction.
// PARTIAL: madeAnyChanges; checkFreeElementsAndReorganise's own error.
func (m *Map) CreateTreeCmd(c *Cmd, flags uint8, a TreePlacement) (cost int32, baseZ uint8, ok bool) {
	om := c.Objects
	if a.X < 0 || a.Y < 0 || a.X >= int32(m.Width*tileSize) || a.Y >= int32(m.Height*tileSize) {
		return 0, 0, false
	}
	land, _ := m.GetHeight(a.X, a.Y)
	c.Pos = &[3]int32{a.X + tileSize/2, a.Y + tileSize/2, int32(land)}
	if om == nil || int(a.Type) >= len(om.TreeObjects) || om.TreeObjects[a.Type] == nil {
		return 0, 0, false
	}
	to := om.TreeObjects[a.Type]
	q := a.Quadrant & 3
	qLand, _ := m.GetHeight(a.X+quadrantCentreOffsets[q][0]-1, a.Y+quadrantCentreOffsets[q][1]-1)
	t := m.TileAt(int(a.X/tileSize), int(a.Y/tileSize))
	if t.WaterLevel&0x1F != 0 && int16(t.WaterLevel&0x1F)*microZStep-1 > qLand {
		c.ErrorText = i18n.CantBuildThisUnderwater
		return 0, 0, false
	}
	if int(t.TerrainIndex) < len(om.LandObjects) && om.LandObjects[t.TerrainIndex] != nil && om.LandObjects[t.TerrainIndex].Flags&objects.LandFlagNoTrees != 0 {
		c.ErrorText = i18n.LandTypeNotSuitable
		return 0, 0, false
	}
	baseZ = uint8(qLand / smallZStep)
	clearZ := baseZ + to.InitialHeight/smallZStep
	if a.RequiresFullClearance {
		clearZ = 0xFF
	}
	baseQuarter := uint8(1) << (q ^ 2)
	m.LastCollision = nil
	if !m.CanConstructAt(a.X, a.Y, baseZ, clearZ, baseQuarter, 0x0F) {
		c.setCollisionError(m)
		return 0, 0, false
	}
	cost = m.InflationAdjustedCost(to.BuildCostFactor, to.CostIndex, 12)
	if flags&CmdApply == 0 {
		return cost, baseZ, true
	}
	el := scenario.TreeElement{TreeObjectID: a.Type, Rotation: a.Rotation, Quadrant: q, Colour: a.Colour,
		Season: to.CurrentSeason, Unk7l: 7, BaseZ: baseZ, ClearZ: to.InitialHeight/smallZStep + baseZ}
	if a.BuildImmediately {
		el.Growth = to.Growth - 1
		el.ClearZ = to.Height/smallZStep + baseZ
		if int(baseZ)-4 > int(m.CurrentSnowLine) && uint16(to.Flags)&treeHasSnowVariation != 0 {
			el.HasSnow = true
		}
	}
	if kMaxElements-m.EntriesEnd <= kMaxElementsOnOneTile { // checkFreeElementsAndReorganise
		m.EntriesEnd = m.countElements() // its full reorganise (the periodic defrags first: PARTIAL)
	}
	m.EntriesEnd += len(t.Order) + 1 // insertElement copies the tile to the end of the entries
	t.Trees = append(t.Trees, el)
	t.insertRef(scenario.KindTree, len(t.Trees)-1, baseZ)
	for i := range t.Order { // the new ref carries the occupied quarter
		if t.Order[i].Kind == scenario.KindTree && int(t.Order[i].Index) == len(t.Trees)-1 {
			t.Order[i].Flags = baseQuarter
			if flags&CmdGhost != 0 {
				t.Order[i].Flags |= elemFlagGhost
			}
		}
	}
	if flags&CmdGhost == 0 {
		m.updateTownInfo(a.X, a.Y, 0, 0, to.Rating, 0, c)
	}
	m.invalidateTile(a.X, a.Y, int32(el.BaseZ)*4, int32(el.ClearZ)*4, ZoomEighth, 56)
	return cost, baseZ, true
}

// setCollisionError is TileClearance's setCollisionErrorMessage after a
// refused clearance: what the clear function said, else what was in the way
// (collisionMessage), else object_in_the_way.
func (c *Cmd) setCollisionError(m *Map) {
	if c.ErrorText == 0 {
		c.ErrorText, c.ErrorArg = m.collisionMessage(c)
	}
	if c.ErrorText == 0 {
		c.ErrorText = i18n.ObjectInTheWay
	}
	m.LastCollision = nil
}

// TreeRemoval is TreeRemovalArgs: the tile and height (world units), the
// tree object, its quadrant and rotation.
type TreeRemoval struct {
	X, Y, Z                  int32
	Type, Quadrant, Rotation uint8
}

// RemoveTree is the removeTree game command (0x004BB392): the first tree on
// the tile at that height, quadrant, rotation and of that type goes when
// applying (TileManager::removeTree: the town rating unless a ghost
// removal); its clear cost (inflation adjusted, divisor 12). No such tree:
// failure. The position the tile's centre at its land height.
func (m *Map) RemoveTree(c *Cmd, flags uint8, a TreeRemoval) (int32, bool) {
	t := m.TileAt(int(a.X/tileSize), int(a.Y/tileSize))
	if t == nil || a.X < 0 || a.Y < 0 {
		return 0, false
	}
	land, _ := m.GetHeight(a.X, a.Y)
	c.Pos = &[3]int32{a.X + tileSize/2, a.Y + tileSize/2, int32(land)}
	for k, r := range t.Order {
		if r.Kind != scenario.KindTree {
			continue
		}
		tr := t.Trees[r.Index]
		if int32(tr.BaseZ)*4 != a.Z || tr.Quadrant != a.Quadrant || tr.Rotation != a.Rotation || tr.TreeObjectID != a.Type {
			continue
		}
		var obj *objects.TreeObject
		if c.Objects != nil && int(tr.TreeObjectID) < len(c.Objects.TreeObjects) {
			obj = c.Objects.TreeObjects[tr.TreeObjectID]
		}
		var cost int32
		if obj != nil {
			cost = m.InflationAdjustedCost(obj.ClearCostFactor, obj.CostIndex, 12)
		}
		if flags&CmdApply != 0 {
			m.removeTree(c, t, k, a.X, a.Y, flags, obj)
		}
		return cost, true
	}
	return 0, false
}
