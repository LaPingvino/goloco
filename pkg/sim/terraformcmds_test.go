package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// The tree commands on the title map: a ghost sapling placed on a free flat
// tile (free, marked ghost, growth 0), removed by the ghost removal; then a
// real one at a cost, a second on the same quarter refused (object in the
// way: named), removed for its clear cost.
func TestCreateRemoveTreeCmd(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	typ := -1
	for i, o := range om.TreeObjects {
		if o != nil {
			typ = i
			break
		}
	}
	if typ < 0 {
		t.Skip("no tree objects")
	}
	for y := 8; y < m.Height-8; y++ {
		for x := 8; x < m.Width-8; x++ {
			tl := m.TileAt(x, y)
			if len(tl.Order) != 1 || tl.Slope&0x1F != 0 || tl.WaterLevel != 0 {
				continue
			}
			a := TreePlacement{X: int32(x) * 32, Y: int32(y) * 32, Quadrant: 1, Rotation: 2, Type: uint8(typ)}
			if _, _, ok := m.CreateTreeCmd(c, CmdApply|CmdGhost|CmdNoPayment, a); !ok {
				continue // no-trees land
			}
			if len(tl.Trees) != 1 || tl.Trees[0].Growth != 0 || tl.Order[len(tl.Order)-1].Flags&elemFlagGhost == 0 {
				t.Fatalf("ghost: %+v, refs %+v", tl.Trees, tl.Order)
			}
			rm := TreeRemoval{X: a.X, Y: a.Y, Z: int32(tl.Trees[0].BaseZ) * 4, Type: a.Type, Quadrant: 1, Rotation: 2}
			if _, ok := m.RemoveTree(c, CmdApply|CmdGhost|CmdNoPayment, rm); !ok || len(tl.Trees) != 0 {
				t.Fatalf("ghost removal: ok %v, %d trees left", ok, len(tl.Trees))
			}
			dry, baseZ, ok := m.CreateTreeCmd(c, 0, a)
			if !ok || len(tl.Trees) != 0 || baseZ != tl.BaseZ {
				t.Fatalf("dry run: ok %v, %d trees, baseZ %d (surface %d)", ok, len(tl.Trees), baseZ, tl.BaseZ)
			}
			cost, _, ok := m.CreateTreeCmd(c, CmdApply, a)
			if !ok || cost != dry || cost <= 0 || len(tl.Trees) != 1 {
				t.Fatalf("apply: ok %v, cost %d (dry %d), %d trees", ok, cost, dry, len(tl.Trees))
			}
			for _, r := range tl.Order {
				if r.Kind == scenario.KindTree && r.Flags&elemFlagGhost != 0 {
					t.Error("a real tree marked ghost")
				}
			}
			c.ErrorText = 0
			if _, _, ok := m.CreateTreeCmd(c, 0, a); ok || c.ErrorText != i18n.StringIdInTheWay || c.ErrorArg == "" {
				t.Errorf("a second tree on the quarter: ok %v, error %d", ok, c.ErrorText)
			}
			if rc, ok := m.RemoveTree(c, CmdApply, rm); !ok || len(tl.Trees) != 0 {
				t.Errorf("removal: ok %v, cost %d, %d trees left", ok, rc, len(tl.Trees))
			}
			if _, ok := m.RemoveTree(c, CmdApply, rm); ok {
				t.Error("removing no tree succeeded")
			}
			t.Logf("tree %d at %d,%d: cost %d", typ, x, y, cost)
			return
		}
	}
	t.Skip("no free flat tile")
}

// RaiseLand / LowerLand with their centre: a flat free tile a step up and
// back, the position the centre at its land height; below the lowest height
// the refusal says so.
func TestRaiseLowerLandCmd(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	for y := 8; y < m.Height-8; y++ {
		for x := 8; x < m.Width-8; x++ {
			tl := m.TileAt(x, y)
			if len(tl.Order) != 1 || tl.Slope&0x1F != 0 || tl.WaterLevel != 0 || tl.BaseZ < 8 {
				continue
			}
			z := tl.BaseZ
			a := LandArgs{CX: int32(x)*32 + 16, CY: int32(y)*32 + 16, AX: x, AY: y, BX: x, BY: y, Corner: mapSelectionFull}
			cost, ok := m.RaiseLand(c, CmdApply, a)
			if !ok || cost <= 0 || tl.BaseZ != z+4 || c.Pos == nil || c.Pos[2] != int32(z+4)*4 {
				t.Fatalf("raise: ok %v, cost %d, baseZ %d → %d, pos %v", ok, cost, z, tl.BaseZ, c.Pos)
			}
			if _, ok := m.LowerLand(c, CmdApply, a); !ok || tl.BaseZ != z {
				t.Fatalf("lower: ok %v, baseZ %d", ok, tl.BaseZ)
			}
			a.Corner = 0 // one corner up: a slope
			if _, ok := m.RaiseLand(c, CmdApply, a); !ok || tl.Slope&0x1F == 0 || tl.BaseZ != z {
				t.Errorf("corner: ok %v, slope %#x, baseZ %d", ok, tl.Slope, tl.BaseZ)
			}
			return
		}
	}
	t.Skip("no free flat tile")
}
