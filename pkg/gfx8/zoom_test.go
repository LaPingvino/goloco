package gfx8

import (
	"bytes"
	"math/rand"
	"testing"
)

// upscale returns s with every pixel doubled in both directions (offsets
// doubled too); RLE sprites get one run per row covering the non-zero span.
func upscale(s *Sprite, k int) *Sprite {
	u := &Sprite{Width: s.Width * k, Height: s.Height * k, XOffset: s.XOffset * k, YOffset: s.YOffset * k,
		Pix: make([]uint8, s.Width*s.Height*k*k), HasTransparency: s.HasTransparency, RLE: s.RLE}
	for y := 0; y < u.Height; y++ {
		for x := 0; x < u.Width; x++ {
			u.Pix[y*u.Width+x] = s.Pix[(y/k)*s.Width+x/k]
		}
	}
	if s.RLE {
		u.Runs = make([][][2]int, u.Height)
		for y := range u.Runs {
			u.Runs[y] = [][2]int{{0, u.Width}}
		}
	}
	return u
}

func randSprite(r *rand.Rand, rle bool) *Sprite {
	w, h := 3+r.Intn(20), 3+r.Intn(20)
	s := &Sprite{Width: w, Height: h, XOffset: -r.Intn(10), YOffset: -r.Intn(10), Pix: make([]uint8, w*h), HasTransparency: true, RLE: rle}
	for i := range s.Pix {
		if r.Intn(3) > 0 {
			s.Pix[i] = uint8(1 + r.Intn(255))
		}
	}
	if rle {
		s.Runs = make([][][2]int, h)
		for y := range s.Runs {
			s.Runs[y] = [][2]int{{0, w}}
		}
	}
	return s
}

// TestZoomDownsample: a 2^z-upscaled sprite drawn at zoom z onto a target of
// the original size equals the original drawn at zoom 0 (positions on the grid).
func TestZoomDownsample(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for n := 0; n < 300; n++ {
		for _, rle := range []bool{false, true} {
			s := randSprite(r, rle)
			z := 1 + r.Intn(3)
			k := 1 << z
			x, y := r.Intn(40)-10, r.Intn(40)-10
			a, b := NewRenderTarget(32, 24), NewRenderTarget(32, 24)
			DrawSpriteNoise(a, s, x, y, OpTransparent, nil, nil)
			DrawSpriteZoom(b, upscale(s, k), x*k, y*k, z, OpTransparent, nil, nil)
			if !bytes.Equal(a.Bits, b.Bits) {
				t.Fatalf("case %d rle=%v zoom=%d at %d,%d: zoomed draw differs", n, rle, z, x, y)
			}
		}
	}
}

func TestNoZoomDraw(t *testing.T) {
	s := randSprite(rand.New(rand.NewSource(2)), false)
	s.NoZoomDraw = true
	rt := NewRenderTarget(16, 16)
	DrawSpriteZoom(rt, s, 8, 8, 1, OpTransparent, nil, nil)
	for _, p := range rt.Bits {
		if p != 0 {
			t.Fatal("noZoomDraw sprite drawn at zoom 1")
		}
	}
}
