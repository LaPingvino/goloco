package world

import "testing"

func TestPlaceStationRequiresTrack(t *testing.T) {
	wld := newTestWorld(10, 10, 4)
	if wld.PlaceStationOnTile(5, 5, 0) {
		t.Error("station placed on a tile with no track")
	}
	wld.SetConstructionHead(5, 5)
	if !wld.PlaceTrackAtHead(0, 0) {
		t.Fatal("PlaceTrackAtHead failed")
	}
	if !wld.PlaceStationOnTile(5, 5, 0) {
		t.Fatal("first PlaceStationOnTile failed")
	}
	if wld.PlaceStationOnTile(5, 5, 0) {
		t.Error("second PlaceStationOnTile should be rejected (already a station)")
	}
}
