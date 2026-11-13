package sim

import "testing"

// The tile loop goes on where the save left it (GameState
// tileUpdateStartLocation; TileManager::update): title.dat's is tile (1, 5)
// of the 16×16 block. Which building meets which roll of its town's rng
// depends on it: from (0, 0) the loop's 19th pass pulled down a house of
// town 18 (Dallas, 13 people, at tile 322, 314) that the real game keeps,
// and the town's 1661 people were 1648 right after loading.
func TestTileLoopStartFromSave(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	if m.tileUpdateX != 1 || m.tileUpdateY != 5 {
		t.Fatalf("tile loop starts at (%d, %d), the save says (1, 5)", m.tileUpdateX, m.tileUpdateY)
	}
	s := &State{Map: *m}
	if got := s.Towns[18].Population; got != 1661 {
		t.Fatalf("town 18 loaded with %d people, want 1661", got)
	}
	for tick := 0; tick < 64; tick++ { // the game's tick: tiles, then the rest
		s.UpdateTiles(func(int, int) bool { return false }, om)
		s.UpdateVehicles(om)
		if got := s.Towns[18].Population; got != 1661 {
			t.Fatalf("town 18 has %d people after %d ticks, want 1661", got, tick+1)
		}
	}
	if b := s.buildingAt(322, 314, s.Tiles[314][322].Buildings[0].BaseZ); b == nil || !b.IsConstructed {
		t.Error("the house at (322, 314) is gone")
	}

	// TileManager::initialise (a generated landscape) starts the loop over.
	s.InitialiseTiles()
	if s.tileUpdateX != 0 || s.tileUpdateY != 0 {
		t.Errorf("after InitialiseTiles the loop starts at (%d, %d), want (0, 0)", s.tileUpdateX, s.tileUpdateY)
	}
}
