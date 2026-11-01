package gfx8

import (
	"os"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// TestDecodeMatchesAssetsDecoder checks, on the real g1.DAT, that every sprite
// decoded to indices and looked up in the palette equals the existing RGBA
// decoder's output — the equivalence step before rendering moves onto gfx8.
func TestDecodeMatchesAssetsDecoder(t *testing.T) {
	const path = "../../../locomotion/Data/g1.DAT"
	if _, err := os.Stat(path); err != nil {
		t.Skip("game data not present")
	}
	g1, err := assets.LoadG1(path)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for i := range g1.Elements {
		e := &g1.Elements[i]
		if e.Width <= 0 || e.Height <= 0 || e.Flags&assets.G1FlagIsR8G8B8Palette != 0 || len(e.Data) == 0 {
			continue
		}
		want, err := g1.DecodeSprite(i)
		if err != nil {
			continue
		}
		s, err := Decode(e)
		if err != nil {
			t.Fatalf("sprite %d: %v", i, err)
		}
		for y := 0; y < s.Height; y++ {
			for x := 0; x < s.Width; x++ {
				idx := s.Pix[y*s.Width+x]
				o := want.PixOffset(x, y)
				got := g1.Palette[idx]
				if idx == Transparent {
					if want.Pix[o+3] != 0 && want.Pix[o+3] != got.A {
						t.Fatalf("sprite %d (%d,%d): index 0 but RGBA decoder has alpha %d", i, x, y, want.Pix[o+3])
					}
					continue
				}
				if want.Pix[o] != got.R || want.Pix[o+1] != got.G || want.Pix[o+2] != got.B {
					t.Fatalf("sprite %d (%d,%d): index %d → %v, RGBA decoder %v", i, x, y, idx, got, want.Pix[o:o+4])
				}
			}
		}
		checked++
	}
	t.Logf("%d sprites identical", checked)
}
