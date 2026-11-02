package gfx8

import (
	"bytes"
	"math/rand"
	"testing"
)

func randSpriteWH(r *rand.Rand, w, h int, rle bool) *Sprite {
	s := &Sprite{Width: w, Height: h, XOffset: r.Intn(9) - 4, YOffset: r.Intn(9) - 4, Pix: make([]uint8, w*h), RLE: rle, HasTransparency: true}
	for i := range s.Pix {
		if r.Intn(4) > 0 {
			s.Pix[i] = uint8(1 + r.Intn(250))
		}
	}
	if rle {
		s.Runs = runsFromPix(s)
	}
	return s
}

// runsFromPix builds the RLE rows ([x0, n] pairs of opaque pixels).
func runsFromPix(s *Sprite) [][][2]int {
	runs := make([][][2]int, s.Height)
	for y := range s.Height {
		for x := 0; x < s.Width; {
			if s.Pix[y*s.Width+x] == 0 {
				x++
				continue
			}
			x0 := x
			for x < s.Width && s.Pix[y*s.Width+x] != 0 {
				x++
			}
			runs[y] = append(runs[y], [2]int{x0, x - x0})
		}
	}
	return runs
}

// Moving the target's origin and the position together by a multiple of
// 2^zoom draws the same pixels (the alignment is on the position before the
// origin is subtracted).
func TestZoomOriginShift(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	for c := 0; c < 200; c++ {
		zoom := r.Intn(4)
		s := randSpriteWH(r, 1+r.Intn(20), 1+r.Intn(20), c%2 == 1)
		x, y := r.Intn(80)-20, r.Intn(80)-20
		a := NewRenderTarget(16, 16)
		DrawSpriteZoom(a, s, x, y, zoom, OpTransparent, nil, nil)
		k := (r.Intn(9) - 4) << zoom
		m := (r.Intn(9) - 4) << zoom
		b := NewRenderTarget(16, 16)
		b.OriginX, b.OriginY = k, m
		DrawSpriteZoom(b, s, x+k, y+m, zoom, OpTransparent, nil, nil)
		if !bytes.Equal(a.Bits, b.Bits) {
			t.Fatalf("case %d (zoom %d, rle %v, shift %d,%d): differ", c, zoom, s.RLE, k, m)
		}
	}
}

// drawImagePaletteSet's recursion: an image with zoom sprites at zoom 1 is
// its image − zoomOffset drawn at zoom 0 with the position and origin halved.
func TestDrawImageZoomRecursion(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	for c := 0; c < 100; c++ {
		small := randSpriteWH(r, 1+r.Intn(10), 1+r.Intn(10), c%2 == 1)
		big := randSpriteWH(r, 2*small.Width, 2*small.Height, false)
		big.HasZoomSprites, big.ZoomOffset = true, 7
		table := map[int]*Sprite{10: big, 3: small}
		look := func(id int) *Sprite { return table[id] }
		x, y, ox, oy := r.Intn(60)-10, r.Intn(60)-10, r.Intn(20)-10, r.Intn(20)-10
		a := NewRenderTarget(24, 24)
		a.OriginX, a.OriginY = ox, oy
		DrawImageZoom(a, look, 10, x, y, 1, OpTransparent, nil, nil)
		b := NewRenderTarget(24, 24)
		b.OriginX, b.OriginY = ox>>1, oy>>1
		DrawSpriteZoom(b, small, x>>1, y>>1, 0, OpTransparent, nil, nil)
		if !bytes.Equal(a.Bits, b.Bits) {
			t.Fatalf("case %d: the recursion differs from the half-size image", c)
		}
	}
}
