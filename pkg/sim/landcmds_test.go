package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Clearing a 3×3 patch with trees on the title map takes the trees off (at a
// cost) and leaves track / road; the dry run leaves the trees and costs the same.
func TestClearLand(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	trees := func(x0, y0 int) (n, other int) {
		for y := y0; y < y0+3; y++ {
			for x := x0; x < x0+3; x++ {
				tl := m.TileAt(x, y)
				n += len(tl.Trees)
				other += len(tl.Tracks) + len(tl.Roads) + len(tl.Buildings)
			}
		}
		return
	}
	for y := 8; y < m.Height-8; y += 3 {
		for x := 8; x < m.Width-8; x += 3 {
			n, other := trees(x, y)
			if n < 3 || other != 0 {
				continue
			}
			cx, cy := int32(x+1)*32+16, int32(y+1)*32+16
			dry, ok := m.ClearLand(c, 0, cx, cy, x, y, x+2, y+2)
			if n2, _ := trees(x, y); !ok || n2 != n {
				t.Fatalf("dry run: ok %v, trees %d → %d", ok, n, n2)
			}
			cost, ok := m.ClearLand(c, CmdApply, cx, cy, x, y, x+2, y+2)
			left, _ := trees(x, y)
			t.Logf("%d trees at %d,%d cleared for %d (dry %d)", n, x, y, cost, dry)
			if !ok || left != 0 || cost != dry || cost == 0 {
				t.Errorf("apply: ok %v, %d trees left, cost %d (dry %d)", ok, left, cost, dry)
			}
			return
		}
	}
	t.Skip("no tree-only 3×3 patch")
}

// RemoveWall: only the wall at that height and rotation; a ghost wall only by
// a ghost removal.
func TestRemoveWall(t *testing.T) {
	m, _ := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral}
	x, y := 20, 20
	tl := m.TileAt(x, y)
	z := tl.BaseZ
	add := func(rot uint8, ghost bool) {
		tl.Walls = append(tl.Walls, scenario.WallElement{Rotation: rot, BaseZ: z, ClearZ: z + 4})
		f := uint8(0)
		if ghost {
			f = elemFlagGhost
		}
		tl.Order = append(tl.Order, scenario.ElemRef{Kind: scenario.KindWall, Index: uint16(len(tl.Walls) - 1), Flags: f})
	}
	n := len(tl.Walls)
	add(1, false)
	add(2, true)
	wz := int32(z) * 4
	m.RemoveWall(c, 0, int32(x)*32, int32(y)*32, wz, 1)
	if len(tl.Walls) != n+2 {
		t.Fatal("a dry run removed a wall")
	}
	m.RemoveWall(c, CmdApply, int32(x)*32, int32(y)*32, wz, 2)
	if len(tl.Walls) != n+2 {
		t.Fatal("a real removal took the ghost wall")
	}
	m.RemoveWall(c, CmdApply, int32(x)*32, int32(y)*32, wz, 1)
	m.RemoveWall(c, CmdApply|CmdGhost, int32(x)*32, int32(y)*32, wz, 2)
	if len(tl.Walls) != n {
		t.Fatalf("walls %d, want %d", len(tl.Walls), n)
	}
	for _, r := range tl.Order {
		if r.Kind == scenario.KindWall && int(r.Index) >= len(tl.Walls) {
			t.Fatal("a dangling wall ref")
		}
	}
}
