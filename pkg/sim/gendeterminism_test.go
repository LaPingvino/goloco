package sim

import (
	"hash/fnv"
	"testing"
)

// The same seed generates the same landscape (a game's seed must reproduce
// it: OpenTTD keeps one generation seed for exactly this).
func TestGenerationIsDeterministic(t *testing.T) {
	_, om := loadDataMap(t, "title.dat")
	g1, _ := om.G1File.(G1Raw)
	sum := func() uint64 {
		m := NewGeneratedMap(1234, 5678, 1950, EditorDefaultTerrainOptions(om), g1, om)
		h := fnv.New64a()
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				tl := &m.Tiles[y][x]
				h.Write([]byte{tl.BaseZ, tl.Slope, tl.WaterLevel, byte(len(tl.Trees)), byte(len(tl.Order))})
			}
		}
		return h.Sum64()
	}
	a, b := sum(), sum()
	if a != b {
		t.Errorf("two generations from one seed differ: %x vs %x", a, b)
	}
}
