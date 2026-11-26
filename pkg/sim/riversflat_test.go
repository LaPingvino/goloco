package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/core"
)

// generateRivers returns for flat land from the original generator
// (MapGenerator.cpp, upstream d9cb11a #4020): no river, no rng draw. Any
// other style, or another generator, still carves them.
func TestGenerateRiversSkipsFlatLand(t *testing.T) {
	for _, c := range []struct {
		style, generator uint8
		carves           bool
	}{{0, 0, false}, {1, 0, true}, {0, 1, true}} {
		m := &Map{Width: 64, Height: 64}
		m.Rng1 = *core.NewPrngSeed(12345, 67890)
		hm := NewHeightMap(64, 64)
		for i := range hm.Height {
			for j := range hm.Width {
				*hm.At(j, i) = 20
			}
		}
		o := TerrainOptions{TopographyStyle: c.style, Generator: c.generator, SeaLevel: 4, MinLandHeight: 2,
			NumRiverbeds: 2, MinRiverWidth: 3, MaxRiverWidth: 6, RiverbankWidth: 1, RiverMeanderRate: 2}
		m.GenerateRivers(o, hm)
		carved := false
		for _, v := range hm.data {
			carved = carved || v != 20
		}
		drew := m.Rng1.Srand0() != 12345 || m.Rng1.Srand1() != 67890
		if carved != c.carves || drew != c.carves {
			t.Errorf("style %d generator %d: carved %v drew %v, want %v", c.style, c.generator, carved, drew, c.carves)
		}
	}
}
