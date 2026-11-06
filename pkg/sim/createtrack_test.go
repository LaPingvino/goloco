package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// createTrack's start: an extra that isn't loaded is dropped, a height off
// the 16 grid is refused, and the cost is the track's price times the
// piece's cost factor / 256 (the curve's factor).
func TestCreateTrackCost(t *testing.T) {
	h := &objects.ObjectHeader{}
	copy(h.Name[:], "TRACKST ")
	track := &objects.TrackObject{Header: h, BuildCostFactor: 100, CostIndex: 0, ModNames: []string{"OVERHEAD"}}
	c := &Cmd{Company: 0, Objects: &objects.ObjectManager{TrackObjects: []*objects.TrackObject{track}}}
	m := &Map{}
	base := m.InflationAdjustedCost(100, 0, 10)
	mods, cost, ok := m.createTrackCost(c, trackPlacementArgs{Z: 32, Mods: 1})
	if !ok || mods != 0 {
		t.Errorf("mods %b ok %v", mods, ok)
	}
	if want := base * int32(TrackCostFactor[0]) / 256; cost != want {
		t.Errorf("straight cost %d, want %d", cost, want)
	}
	if _, _, ok := m.createTrackCost(c, trackPlacementArgs{Z: 40}); ok {
		t.Error("a height off the 16 grid must fail")
	}
	_, curve, _ := m.createTrackCost(c, trackPlacementArgs{Z: 32, TrackID: 6})
	if want := base * int32(TrackCostFactor[6]) / 256; curve != want {
		t.Errorf("curve cost %d, want %d", curve, want)
	}
	if _, _, ok := m.createTrackCost(c, trackPlacementArgs{TrackObjectID: 3}); ok {
		t.Error("an empty slot must fail")
	}
}
