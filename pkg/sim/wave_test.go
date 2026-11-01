package sim

import "testing"

func waterMap(n int) *State {
	s := &State{Map: Map{Width: n, Height: n, Tiles: make([][]Tile, n)}}
	for y := range s.Tiles {
		s.Tiles[y] = make([]Tile, n)
		for x := range s.Tiles[y] {
			s.Tiles[y][x].WaterLevel = 4
		}
	}
	s.ResetWaves()
	return s
}

func TestWaveIndex(t *testing.T) {
	// OpenLoco WaveManager.h: (x & 7) | ((y & 7) << 3)
	for _, c := range []struct{ x, y, want int }{{0, 0, 0}, {7, 0, 7}, {0, 1, 8}, {9, 10, 1 | 2<<3}, {383, 383, 63}} {
		if got := WaveIndex(c.x, c.y); got != c.want {
			t.Errorf("WaveIndex(%d,%d) = %d, want %d", c.x, c.y, got, c.want)
		}
	}
}

func TestWaveNeedsWaterAllAround(t *testing.T) {
	s := waterMap(8)
	s.Tiles[3][4].WaterLevel = 0 // land east of (3,3)
	for i := 0; i < 2000; i++ {
		s.createWave(&s.Tiles[3][3], 3, 3, nil)
	}
	if !s.Waves[WaveIndex(3, 3)].Empty() || s.Tiles[3][3].WaveFlag {
		t.Fatal("a wave started next to land")
	}
}

func TestWaveLifetime(t *testing.T) {
	s := waterMap(8)
	for i := 0; i < 2000 && s.Waves[WaveIndex(3, 3)].Empty(); i++ {
		s.createWave(&s.Tiles[3][3], 3, 3, nil)
	}
	if s.Waves[WaveIndex(3, 3)].Empty() || !s.Tiles[3][3].WaveFlag {
		t.Fatal("no wave started on open water in 2000 tries")
	}
	// One frame per 4 ticks; freed after frame 16.
	for tick := uint32(1); tick <= 16*4; tick++ {
		s.UpdateWaves(tick)
		if tick < 16*4 && s.Waves[WaveIndex(3, 3)].Empty() {
			t.Fatalf("wave freed early at tick %d", tick)
		}
	}
	if !s.Waves[WaveIndex(3, 3)].Empty() || s.Tiles[3][3].WaveFlag {
		t.Fatal("wave still alive after 16 frames")
	}
}

func TestTileLoopStartMovesThroughBlock(t *testing.T) {
	s := waterMap(32)
	seen := map[[2]int]bool{}
	for i := 0; i < 256; i++ {
		seen[[2]int{s.tileUpdateX, s.tileUpdateY}] = true
		s.UpdateTiles(nil)
	}
	if len(seen) != 256 {
		t.Fatalf("start offset visited %d of 256 positions in a 16×16 block", len(seen))
	}
}
