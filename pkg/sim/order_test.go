package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// checkOrder: Order covers every element once, in non-decreasing baseZ.
func checkOrder(t *testing.T, tl *Tile) {
	t.Helper()
	n := 1 + len(tl.Tracks) + len(tl.Roads) + len(tl.Stations)
	if len(tl.Order) != n {
		t.Fatalf("Order has %d refs, tile has %d elements: %v", len(tl.Order), n, tl.Order)
	}
	seen := map[scenario.ElemRef]bool{}
	prev := -1
	for _, r := range tl.Order {
		if seen[r] {
			t.Fatalf("ref %v twice", r)
		}
		seen[r] = true
		if z := int(tl.baseZOf(r)); z < prev {
			t.Fatalf("Order not by height: %v", tl.Order)
		} else {
			prev = z
		}
	}
}

func TestOrderFollowsMutations(t *testing.T) {
	s := &State{Map: Map{Width: 1, Height: 1, Tiles: [][]Tile{{{BaseZ: 8, Order: []scenario.ElemRef{{Kind: scenario.KindSurface}}}}}}}
	tl := s.TileAt(0, 0)
	// A tunnel track below the surface, two at surface height, a road above.
	s.AddTrack(0, 0, scenario.TrackElement{TrackID: 0, BaseZ: 8})
	s.AddTrack(0, 0, scenario.TrackElement{TrackID: 0, BaseZ: 4})
	s.AddRoad(0, 0, scenario.RoadElement{RoadID: 0, BaseZ: 12})
	s.AddTrack(0, 0, scenario.TrackElement{TrackID: 1, BaseZ: 8})
	checkOrder(t, tl)
	if tl.Order[0] != (scenario.ElemRef{Kind: scenario.KindTrack, Index: 1}) {
		t.Fatalf("the tunnel track should come first: %v", tl.Order)
	}
	// Equal height goes after the existing ones (insertElement's >=).
	if tl.Order[1].Kind != scenario.KindSurface || tl.Order[2].Index != 0 || tl.Order[3].Index != 2 {
		t.Fatalf("equal-height order wrong: %v", tl.Order)
	}
	tl.Tracks = append(tl.Tracks[:1], tl.Tracks[2:]...) // the tunnel track (index 1) goes
	tl.removeRef(scenario.KindTrack, 1)
	checkOrder(t, tl)
	tl.Roads = tl.Roads[:0]
	tl.removeRef(scenario.KindRoad, 0)
	checkOrder(t, tl)
	if len(tl.Tracks) != 2 || len(tl.Roads) != 0 {
		t.Fatalf("removals: %d tracks %d roads", len(tl.Tracks), len(tl.Roads))
	}
}
