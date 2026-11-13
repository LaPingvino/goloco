package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/sim"
)

func TestPlaceStationRequiresTrack(t *testing.T) {
	wld := newTestWorld(10, 10, 4)
	if wld.PlaceStationOnTile(5, 5, 0) {
		t.Error("station placed on a tile with no track")
	}
	if _, ok := wld.LayTrack(5*32, 5*32, 4*4, 0, 0, 0); !ok {
		t.Fatal("LayTrack failed")
	}
	if !wld.PlaceStationOnTile(5, 5, 0) {
		t.Fatal("first PlaceStationOnTile failed")
	}
	if wld.PlaceStationOnTile(5, 5, 0) {
		t.Error("second PlaceStationOnTile should be rejected (already a station)")
	}
}

// newTestWorld builds a synthetic flat World (no renderer).
func newTestWorld(w, h int, baseZ uint8) *World {
	tiles := make([][]tile, h)
	for y := 0; y < h; y++ {
		tiles[y] = make([]tile, w)
		for x := 0; x < w; x++ {
			tiles[y][x] = tile{BaseZ: baseZ}
		}
	}
	return &World{State: sim.State{Map: sim.Map{Width: w, Height: h, Tiles: tiles}}}
}
