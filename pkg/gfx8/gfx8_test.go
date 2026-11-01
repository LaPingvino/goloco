package gfx8

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
)

func identityPlus(n uint8) []uint8 {
	m := make([]uint8, 256)
	for i := range m {
		m[i] = uint8(i) + n
	}
	return m
}

func TestDecodeRLE(t *testing.T) {
	// 4×2 sprite: row 0 = run of 2 at x=1 (last); row 1 = run of 1 at x=0, run of 1 at x=3 (last)
	row0 := []byte{0x80 | 2, 1, 7, 8}
	row1 := []byte{1, 0, 9, 0x80 | 1, 3, 10}
	data := []byte{4, 0, byte(4 + len(row0)), 0}
	data = append(append(data, row0...), row1...)
	s, err := Decode(&assets.G1Element{Width: 4, Height: 2, Flags: assets.G1FlagIsRLECompressed, Data: data})
	if err != nil {
		t.Fatal(err)
	}
	want := []uint8{0, 7, 8, 0, 9, 0, 0, 10}
	for i := range want {
		if s.Pix[i] != want[i] {
			t.Fatalf("pix = %v, want %v", s.Pix, want)
		}
	}
}

func TestBlendOps(t *testing.T) {
	s := &Sprite{Width: 3, Height: 1, Pix: []uint8{0, 5, 6}, HasTransparency: true}
	fresh := func() *RenderTarget {
		rt := NewRenderTarget(3, 1)
		copy(rt.Bits, []uint8{20, 21, 22})
		return rt
	}
	// transparent: index 0 skipped, others copied
	rt := fresh()
	DrawSprite(rt, s, 0, 0, OpTransparent, nil)
	if got := rt.Bits; got[0] != 20 || got[1] != 5 || got[2] != 6 {
		t.Errorf("transparent: %v", got)
	}
	// src remap: copied through the map
	rt = fresh()
	DrawSprite(rt, s, 0, 0, OpTransparent|OpSrc, identityPlus(100))
	if got := rt.Bits; got[0] != 20 || got[1] != 105 || got[2] != 106 {
		t.Errorf("src: %v", got)
	}
	// dst (mask): destination mapped where the source is non-zero — water, glass
	rt = fresh()
	DrawSprite(rt, s, 0, 0, OpTransparent|OpDst, identityPlus(100))
	if got := rt.Bits; got[0] != 20 || got[1] != 121 || got[2] != 122 {
		t.Errorf("dst: %v", got)
	}
	// dst mapping to 0 leaves the pixel alone
	m := identityPlus(0)
	m[21] = 0
	rt = fresh()
	DrawSprite(rt, s, 0, 0, OpTransparent|OpDst, m)
	if got := rt.Bits; got[1] != 21 {
		t.Errorf("dst→0 must not write: %v", got)
	}
	// src+dst: row src-1, column dst
	big := make([]uint8, 256*7)
	big[(5-1)*256+21] = 77
	big[(6-1)*256+22] = 88
	rt = fresh()
	DrawSprite(rt, s, 0, 0, OpTransparent|OpSrc|OpDst, big)
	if got := rt.Bits; got[1] != 77 || got[2] != 88 {
		t.Errorf("src+dst: %v", got)
	}
}

func TestOpFor(t *testing.T) {
	s := &Sprite{HasTransparency: true}
	if OpFor(s, false, true) != OpTransparent|OpDst {
		t.Error("blended without primary must be a dst mask (water)")
	}
	if OpFor(&Sprite{}, false, false) != OpNone {
		t.Error("opaque sprite without transparency must copy raw")
	}
}

func TestDrawSpriteClipsAndOffsets(t *testing.T) {
	s := &Sprite{Width: 2, Height: 2, XOffset: -1, YOffset: -1, Pix: []uint8{1, 2, 3, 4}, HasTransparency: true}
	rt := NewRenderTarget(2, 2)
	DrawSprite(rt, s, 0, 0, OpTransparent, nil) // only the bottom-right pixel lands at (0,0)
	if rt.Bits[0] != 4 || rt.Bits[1] != 0 || rt.Bits[2] != 0 {
		t.Errorf("offset/clip: %v", rt.Bits)
	}
}

func TestDrawSpriteMasked(t *testing.T) {
	img := &Sprite{Width: 3, Height: 1, Pix: []uint8{9, 9, 9}}
	mask := &Sprite{Width: 3, Height: 1, Pix: []uint8{0xFF, 0, 0x0F}}
	rt := NewRenderTarget(3, 1)
	copy(rt.Bits, []uint8{1, 2, 3})
	DrawSpriteMasked(rt, img, mask, 0, 0)
	// 9&0xFF=9 written; 9&0=0 skipped; 9&0x0F=9 written
	if rt.Bits[0] != 9 || rt.Bits[1] != 2 || rt.Bits[2] != 9 {
		t.Fatalf("masked: %v", rt.Bits)
	}
}

func TestNoiseMaskUsesSpriteWidthStride(t *testing.T) {
	s := &Sprite{Width: 2, Height: 2, Pix: []uint8{5, 5, 5, 5}, HasTransparency: true}
	// noise image 4 wide; with the sprite's width (2) as stride, row 1 reads noise[2..3]
	noise := &Sprite{Width: 4, Height: 2, Pix: []uint8{0xFF, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}}
	rt := NewRenderTarget(2, 2)
	DrawSpriteNoise(rt, s, 0, 0, OpTransparent, nil, noise)
	want := []uint8{5, 0, 0, 5}
	for i := range want {
		if rt.Bits[i] != want[i] {
			t.Fatalf("noise: %v, want %v", rt.Bits, want)
		}
	}
	// RLE sprites ignore the noise image (vanilla behaviour)
	s.RLE = true
	rt = NewRenderTarget(2, 2)
	DrawSpriteNoise(rt, s, 0, 0, OpTransparent, nil, noise)
	if rt.Bits[1] != 5 {
		t.Fatalf("RLE sprite must ignore noise: %v", rt.Bits)
	}
}
