package ui

import (
	"image/color"
	"math/rand"
	"testing"
)

// The flat palette-index table answers as the map did: a palette colour its
// lowest index (0 skipped), any other colour the nearest by squared RGB
// distance (lowest index on a tie) — checked by brute force.
func TestPaletteIndexTable(t *testing.T) {
	palInv = nil
	pal := make([]color.RGBA, 256)
	rng := rand.New(rand.NewSource(1))
	for i := range pal {
		pal[i] = color.RGBA{uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256)), 255}
	}
	pal[200] = pal[17] // a repeated colour: index 17 wins
	inv := paletteIndex(pal)
	brute := func(r, g, b uint8) uint8 {
		best, bd := uint8(1), 1<<30
		for i := 1; i < len(pal); i++ {
			c := pal[i]
			d := sq(int(c.R)-int(r)) + sq(int(c.G)-int(g)) + sq(int(c.B)-int(b))
			if d < bd {
				best, bd = uint8(i), d
			}
		}
		return best
	}
	if got := inv.lookup(pal, pal[200].R, pal[200].G, pal[200].B); got != 17 && pal[17] == pal[200] {
		t.Errorf("repeated colour: %d", got)
	}
	for range 20000 {
		r, g, b := uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256))
		if got, want := inv.lookup(pal, r, g, b), brute(r, g, b); got != want {
			t.Fatalf("(%d,%d,%d): %d, want %d", r, g, b, got, want)
		}
	}
	palInv = nil
}
