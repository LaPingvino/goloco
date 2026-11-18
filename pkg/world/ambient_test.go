package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// updateAmbientNoise's scan: the 12 × 12 tiles from 5 before the centre's
// tile (clamped to the map); a tile with water counts as water; a surface
// that is its tile's last element counts as wilderness when snowy or at
// baseZ 64 and up; every tree after the surface counts.
func TestAmbientCounts(t *testing.T) {
	w := newTestWorld(20, 20, 8)
	for y := range 20 {
		for x := range 20 {
			w.Map.Tiles[y][x].Order = []scenario.ElemRef{{Kind: scenario.KindSurface}}
		}
	}
	set := func(x, y int, f func(tl *tile)) { f(&w.Map.Tiles[y][x]) }
	set(5, 5, func(tl *tile) { tl.WaterLevel = 2 })               // in: water
	set(16, 16, func(tl *tile) { tl.WaterLevel = 2 })             // in (the far corner)
	set(17, 16, func(tl *tile) { tl.WaterLevel = 2 })             // out
	set(4, 10, func(tl *tile) { tl.WaterLevel = 2 })              // out
	set(6, 6, func(tl *tile) { tl.BaseZ = 64 })                   // wilderness
	set(7, 6, func(tl *tile) { tl.Slope = 2 << 5 })               // snowy: wilderness
	set(8, 6, func(tl *tile) { tl.WaterLevel, tl.BaseZ = 1, 70 }) // water first
	set(9, 6, func(tl *tile) {                                    // high, but trees on it: not last; two trees
		tl.BaseZ = 70
		tl.Trees = []scenario.TreeElement{{}, {}}
		tl.Order = append(tl.Order, scenario.ElemRef{Kind: scenario.KindTree}, scenario.ElemRef{Kind: scenario.KindTree, Index: 1})
	})
	water, wild, trees := w.ambientCounts(nil, 10*32+5, 10*32+20)
	if water != 3 || wild != 2 || trees != 2 {
		t.Errorf("counts %d %d %d, want 3 water, 2 wilderness, 2 trees", water, wild, trees)
	}
	// near the map's corner the range is clamped: tiles 0…6
	water, _, _ = w.ambientCounts(nil, 1*32, 1*32)
	if water != 1 {
		t.Errorf("clamped: water %d, want 1 (5, 5)", water)
	}
}
