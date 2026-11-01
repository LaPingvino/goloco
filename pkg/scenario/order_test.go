package scenario

import (
	"os"
	"testing"
)

// TestTileOrder: every element of every tile is in Order exactly once, in
// stored order, and the surface is present.
func TestTileOrder(t *testing.T) {
	const f = "../../../locomotion/Data/title.dat"
	if _, err := os.Stat(f); err != nil {
		t.Skip("game data not present")
	}
	sc, err := LoadScenarioData(f)
	if err != nil {
		t.Fatal(err)
	}
	underground := 0
	for y := range sc.MapHeight {
		for x := range sc.MapWidth {
			tl := sc.GetTile(x, y)
			n := 1 + len(tl.Trees) + len(tl.Buildings) + len(tl.Walls) + len(tl.Tracks) + len(tl.Roads) + len(tl.Stations) + len(tl.Signals) + len(tl.Industries)
			if len(tl.Order) != n {
				t.Fatalf("tile %d,%d: Order has %d, tile has %d elements", x, y, len(tl.Order), n)
			}
			seen := map[ElemRef]bool{}
			surfaceAt := -1
			for i, r := range tl.Order {
				if seen[r] {
					t.Fatalf("tile %d,%d: %v twice", x, y, r)
				}
				seen[r] = true
				if r.Kind == KindSurface {
					surfaceAt = i
				}
			}
			if surfaceAt < 0 {
				t.Fatalf("tile %d,%d: no surface in Order", x, y)
			}
			if surfaceAt > 0 {
				underground++
			}
		}
	}
	t.Logf("tiles with elements stored before the surface (underground): %d", underground)
}
