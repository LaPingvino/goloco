package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/scenario"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// A station's track, road and station elements hover and click as their
// station (_track / _road / getStationArguments): the track's station element
// right after it, the road's after its matching road; a ghost station, an
// AI-allocated one under a track, or a track without a station gives none.
func TestStationOfElement(t *testing.T) {
	w := &World{}
	w.Map = sim.Map{Width: 2, Height: 2, Tiles: [][]sim.Tile{make([]sim.Tile, 2), make([]sim.Tile, 2)}}
	tl := w.Map.TileAt(1, 1)
	tl.Tracks = []scenario.TrackElement{{HasStation: true, BaseZ: 8}, {BaseZ: 12}}
	tl.Roads = []scenario.RoadElement{{HasStationElement: true, BaseZ: 16, RoadID: 0, Rotation: 1}}
	tl.Stations = []scenario.StationElement{{StationID: 7}, {StationID: 9}}
	tl.Order = []scenario.ElemRef{
		{Kind: scenario.KindSurface}, {Kind: scenario.KindTrack, Index: 0}, {Kind: scenario.KindStation, Index: 0},
		{Kind: scenario.KindTrack, Index: 1}, {Kind: scenario.KindRoad, Index: 0}, {Kind: scenario.KindStation, Index: 1},
	}
	at := func(item any) (int, bool) {
		return w.stationOfElement(&paint.Struct{MapX: 32, MapY: 32, Item: item})
	}
	for _, c := range []struct {
		name string
		item any
		id   int
		ok   bool
	}{
		{"track with station", &tl.Tracks[0], 7, true},
		{"track without", &tl.Tracks[1], 0, false},
		{"road with station", &tl.Roads[0], 9, true},
		{"station element", &tl.Stations[1], 9, true},
	} {
		if id, ok := at(c.item); id != c.id || ok != c.ok {
			t.Errorf("%s: %d %v, want %d %v", c.name, id, ok, c.id, c.ok)
		}
	}
	tl.Stations[0].AiAllocated = true
	if _, ok := at(&tl.Tracks[0]); ok {
		t.Error("an AI-allocated station under a track resolved")
	}
	tl.Stations[1].Ghost = true
	if _, ok := at(&tl.Stations[1]); ok {
		t.Error("a ghost station resolved")
	}
}
