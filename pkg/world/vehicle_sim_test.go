package world

import (
	"math"
	"testing"
)

// TestVehicleSimStraightLine places three connected straight track pieces,
// spawns a vehicle at the line's start and asserts it advances monotonically
// along the line, reaches the far end, then bounces back.
func TestVehicleSimStraightLine(t *testing.T) {
	wld := newTestWorld(30, 30, 4)
	wld.SetConstructionHead(15, 15) // head at (15,15), rot 0, baseZ 4

	// Three straight (trackID 0) pieces, rot 0. Each advances the head by the
	// piece delta kTrackCoordinates[0] = {0,0,-32,0,0}: origins land on tiles
	// (15,15), (14,15), (13,15); the head finishes at (12,15).
	for i := 0; i < 3; i++ {
		if !wld.PlaceTrackAtHead(0, 0) {
			t.Fatalf("PlaceTrackAtHead #%d failed, head=%+v", i, wld.ConstructionHeadState())
		}
	}

	if !wld.SpawnTestVehicle(0) {
		t.Fatal("SpawnTestVehicle returned false")
	}
	if len(wld.simVehicles) != 1 || len(wld.entities) != 1 {
		t.Fatalf("expected 1 sim vehicle + 1 entity, got %d/%d", len(wld.simVehicles), len(wld.entities))
	}
	sv := wld.simVehicles[0]
	ent := &wld.entities[0]

	// The line runs along -X (rot 0 straight: DX=-32). The vehicle should start
	// at the line's beginning (highest X, tile 15) with no predecessor.
	if sv.tileX != 15 || sv.tileY != 15 {
		t.Fatalf("spawn tile = (%d,%d), want (15,15)", sv.tileX, sv.tileY)
	}
	startX := ent.PosX
	if int(ent.PosY) != 15*32+tileCenterOffset {
		t.Fatalf("PosY = %d, want %d (constant along an X-line)", ent.PosY, 15*32+tileCenterOffset)
	}

	// Full travelled span: begin of tile 15 (x=15*32+16=496) to end of tile 13
	// (x=13*32+16-32=400). That is 3 pieces × 32 = 96 world units.
	const beginX = 15*32 + tileCenterOffset // 496
	const endX = beginX - 3*32              // 400
	if int(startX) != beginX {
		t.Fatalf("start PosX = %d, want %d", startX, beginX)
	}

	// Phase 1: advance forward, PosX must decrease monotonically until it
	// reaches the far end (endX), and PosY must stay constant.
	prevX := int(startX)
	reachedEnd := false
	for tick := 0; tick < 200; tick++ {
		wld.UpdateVehicles()
		x := int(ent.PosX)
		if int(ent.PosY) != 15*32+tileCenterOffset {
			t.Fatalf("tick %d: PosY drifted to %d", tick, ent.PosY)
		}
		if x > prevX {
			// Direction reversed — must only happen at/after the far end.
			if !reachedEnd {
				t.Fatalf("tick %d: PosX increased to %d before reaching end (min seen %d, want %d)",
					tick, x, prevX, endX)
			}
			break
		}
		prevX = x
		if x <= endX {
			reachedEnd = true
		}
	}
	if !reachedEnd {
		t.Fatalf("vehicle never reached the far end (min PosX %d, want %d)", prevX, endX)
	}
	if math.Abs(float64(prevX-endX)) > 2 {
		t.Errorf("far end PosX = %d, want ~%d", prevX, endX)
	}

	// Phase 2: keep ticking; the vehicle must bounce and climb back toward the
	// start, reaching (approximately) beginX again.
	maxX := prevX
	for tick := 0; tick < 300; tick++ {
		wld.UpdateVehicles()
		if int(ent.PosX) > maxX {
			maxX = int(ent.PosX)
		}
	}
	if maxX < beginX-2 {
		t.Errorf("after bounce, max PosX = %d, want it to return to ~%d", maxX, beginX)
	}
}
