package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A three-piece straight line of a road object with a mod 0 fitting
// straights (the title map's Tram Track and its catenary), laid on empty
// flat tiles — the title map has no tram — each next tile where the last
// piece's end leads (RoadConnectionEnd). Its first piece.
func findModdableRoad(t *testing.T, m *Map, c *Cmd) (int, int, scenario.RoadElement) {
	t.Helper()
	obj := -1
	for i := range 16 {
		if p, _, _, _, ok := m.roadModExtra(c, uint8(i), 0); ok && p&RoadCompatibleFlags[0] == RoadCompatibleFlags[0] {
			obj = i
			break
		}
	}
	if obj < 0 {
		t.Skip("no road object with a straight-fitting mod")
	}
	empty := func(x, y int, z uint8) bool {
		tl := m.TileAt(x, y)
		return tl != nil && len(tl.Order) == 1 && tl.Slope == 0 && tl.WaterLevel == 0 && tl.BaseZ == z
	}
	for y := 8; y < m.Height-8; y++ {
		for x := 8; x < m.Width-8; x++ {
			z := m.TileAt(x, y).BaseZ
			cx, cy, ok := x, y, true
			var tiles [3][2]int
			for i := range 3 {
				if !empty(cx, cy, z) {
					ok = false
					break
				}
				tiles[i] = [2]int{cx, cy}
				e, _ := RoadConnectionEnd(int32(cx)*32, int32(cy)*32, int32(z)*4, 0)
				cx, cy = int(e.X/32), int(e.Y/32)
			}
			if !ok {
				continue
			}
			el := scenario.RoadElement{RoadID: 0, Rotation: 0, BaseZ: z, ClearZ: z + 8, RoadObjectID: uint8(obj), Owner: CompanyNeutral, IsLast: true}
			for _, tp := range tiles {
				m.AddRoad(tp[0], tp[1], el)
			}
			return x, y, el
		}
	}
	t.Skip("no room for a three-piece line")
	return 0, 0, scenario.RoadElement{}
}

func TestCreateRoadMod(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	x, y, el := findModdableRoad(t, m, c)
	tl := m.TileAt(x, y)
	a := RoadModsPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation, Index: el.SequenceIndex, Type: 1, RoadObjectID: el.RoadObjectID, Section: ModSectionSingle}
	dry, ok := m.CreateRoadMod(c, 0, a)
	if !ok || dry <= 0 || tl.Roads[0].Mods&1 != 0 {
		t.Fatalf("dry run: ok %v cost %d mods %#x", ok, dry, tl.Roads[0].Mods)
	}
	m.ControllingCompany = el.Owner
	if _, ok := m.CreateRoadMod(c, CmdApply|CmdGhost, a); !ok || !tl.Roads[0].HasGhostMods || tl.Roads[0].Mods&1 != 0 {
		t.Fatal("ghost: not flagged only")
	}
	m.RemoveRoadMod(c, CmdApply|CmdGhost, a)
	if tl.Roads[0].HasGhostMods {
		t.Fatal("the ghost mod stayed")
	}
	if cost, ok := m.CreateRoadMod(c, CmdApply, a); !ok || cost != dry || tl.Roads[0].Mods&1 == 0 {
		t.Fatalf("apply: ok %v cost %d (dry %d)", ok, cost, dry)
	}
	if refund, ok := m.RemoveRoadMod(c, CmdApply, a); !ok || refund >= 0 || tl.Roads[0].Mods&1 != 0 {
		t.Errorf("removal: ok %v refund %d mods %#x", ok, refund, tl.Roads[0].Mods)
	}
	// All connected (the road walk): more than the one piece, each paid once.
	a.Section = ModSectionAllConnected
	count := func() int {
		n := 0
		for ty := 0; ty < m.Height; ty++ {
			for tx := 0; tx < m.Width; tx++ {
				for _, e := range m.TileAt(tx, ty).Roads {
					if e.Mods&1 != 0 && e.RoadObjectID == el.RoadObjectID {
						n++
					}
				}
			}
		}
		return n
	}
	before := count()
	cost, ok := m.CreateRoadMod(c, CmdApply, a)
	n := count() - before
	t.Logf("all connected: %d road tiles modded for %d (one straight %d)", n, cost, dry)
	if !ok || n != 3 || cost != 3*dry {
		t.Errorf("all connected: ok %v tiles %d cost %d", ok, n, cost)
	}
}
