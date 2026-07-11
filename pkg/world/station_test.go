package world

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// TestStationDeliveryLoop builds three connected straight track pieces, places a
// station at each end, seeds the near end with town buildings, spawns a vehicle
// and ticks until the board→travel→deliver loop fires OnCargoDelivered.
func TestStationDeliveryLoop(t *testing.T) {
	wld := newTestWorld(30, 30, 4)

	// Buildings clustered around the near-end station tile (15,15) so it has a
	// passenger supply. Eight buildings within the supply radius → 2 per 100t.
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			tl := wld.getTile(15+dx, 15+dy)
			tl.buildings = append(tl.buildings, scenario.BuildingElement{})
		}
	}

	wld.SetConstructionHead(15, 15) // head at (15,15), rot 0, baseZ 4
	for i := 0; i < 3; i++ {
		if !wld.PlaceTrackAtHead(0, 0) {
			t.Fatalf("PlaceTrackAtHead #%d failed", i)
		}
	}
	// Origins land on tiles 15, 14, 13 (line runs along -X). Stations at ends.
	if !wld.PlaceStationOnTile(15, 15, 0) {
		t.Fatal("PlaceStationOnTile(15,15) failed")
	}
	if !wld.PlaceStationOnTile(13, 15, 0) {
		t.Fatal("PlaceStationOnTile(13,15) failed")
	}
	if len(wld.simStations) != 2 {
		t.Fatalf("expected 2 registered stations, got %d", len(wld.simStations))
	}
	// The near-end station must have found the 8 nearby buildings.
	near := wld.stationAt(15, 15, 4)
	if near == nil || near.buildings != 8 {
		t.Fatalf("near station supply = %v (want buildings=8)", near)
	}
	// The station element and HasStation flag must be present for the renderer.
	if tl := wld.getTile(15, 15); len(tl.stations) != 1 || !tl.tracks[0].HasStation {
		t.Fatalf("station element/flag not set on tile (15,15): stations=%d hasFlag=%v",
			len(tl.stations), tl.tracks[0].HasStation)
	}

	var delivered uint32
	var deliverEvents int
	wld.OnCargoDelivered = func(slot uint8, amount uint32) {
		if slot != CargoSlotPass {
			t.Errorf("delivered to unexpected cargo slot %d", slot)
		}
		delivered += amount
		deliverEvents++
	}

	if !wld.SpawnTestVehicle(0) {
		t.Fatal("SpawnTestVehicle returned false")
	}
	sv := wld.simVehicles[0]

	for tick := 0; tick < 4000 && delivered == 0; tick++ {
		wld.TickVehicles()
		// Bookkeeping sanity every tick.
		if sv.boarded > vehiclePassengerCapacity {
			t.Fatalf("tick %d: boarded=%d exceeds capacity %d", tick, sv.boarded, vehiclePassengerCapacity)
		}
		for _, s := range wld.simStations {
			if s.waiting > stationWaitingCap {
				t.Fatalf("tick %d: station %d waiting=%d exceeds cap %d", tick, s.id, s.waiting, stationWaitingCap)
			}
		}
	}

	if delivered == 0 {
		t.Fatalf("no passengers delivered after 4000 ticks (near.waiting=%d boarded=%d)",
			near.waiting, sv.boarded)
	}
	if deliverEvents == 0 {
		t.Fatal("OnCargoDelivered never fired despite delivered>0")
	}
	t.Logf("delivered %d passengers across %d events; boarded=%d near.waiting=%d",
		delivered, deliverEvents, sv.boarded, near.waiting)
}

// TestPlaceStationRequiresTrack verifies a station cannot be placed on a tile
// with no track origin, and that a double-placement is rejected.
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
