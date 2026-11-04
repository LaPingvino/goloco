// Package gfx8 is a port of OpenLoco's software drawing: an 8-bit palette-
// indexed render target and sprite blits with every blend mode
// (Graphics/DrawSprite.cpp getDrawBlendOp, DrawSpriteHelper.hpp blitPixel).
// It is the exact CPU reference the rest of goloco's rendering moves onto (LOOP
// item 3i); faster paths must match it. No rendering-library dependencies.
package gfx8

import (
	"encoding/binary"
	"errors"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// Transparent is palette index 0 (OpenLoco PaletteIndex::transparent).
const Transparent = 0

// BlendOp selects how sprite pixels combine with the target (DrawBlendOp).
type BlendOp uint8

const (
	OpNone        BlendOp = 0
	OpTransparent BlendOp = 1 << 0 // skip index-0 source pixels
	OpSrc         BlendOp = 1 << 1 // remap the source pixel through the palette map
	OpDst         BlendOp = 1 << 2 // remap the destination pixel through the palette map
	OpNoiseMask   BlendOp = 1 << 3 // AND the source with a noise mask (0 or 0xFF)
)

// RenderTarget is an 8-bit indexed frame (OpenLoco Gfx::RenderTarget, zoom 0).
type RenderTarget struct {
	Width, Height int
	Bits          []uint8 // Width×Height palette indices, row-major
	// OriginX, OriginY are rt.x / rt.y: the viewport position (zoom-0
	// pixels) of the target's top-left, subtracted after the zoom alignment
	// (DrawSpriteZoom / DrawImageZoom; zero for plain UI drawing).
	OriginX, OriginY int
}

// NewRenderTarget returns a target filled with index 0.
func NewRenderTarget(w, h int) *RenderTarget {
	return &RenderTarget{Width: w, Height: h, Bits: make([]uint8, w*h)}
}

// Sprite is a decoded g1 element: palette indices plus its draw offsets.
type Sprite struct {
	Width, Height    int
	XOffset, YOffset int
	Pix              []uint8 // Width×Height indices; 0 = transparent
	HasTransparency  bool
	RLE              bool // stored RLE-compressed (no noise mask: vanilla didn't support it)

	// Zoom (G1 element flags and zoomOffset): at zoom > 0 a sprite with
	// HasZoomSprites is drawn as image - ZoomOffset at half size; NoZoomDraw
	// sprites are not drawn at all.
	HasZoomSprites, NoZoomDraw bool
	ZoomOffset                 int
	// Index is the g1 image index, set by a sprite cache that knows it
	// (DrawSpriteAtZoom recurses through Index − ZoomOffset); 0 = unknown.
	Index int

	// spans: per row, the runs of non-transparent pixels (start, end) — what
	// upstream's RLE drawing walks; made by Decode, so a transparent copy
	// moves whole runs. nil (sprites made elsewhere) = test every pixel.
	spans [][][2]int32
	// Runs are the RLE rows as stored: per row, [x0, n] pairs (drawRLESprite
	// walks these; the pixels themselves are in Pix).
	Runs [][][2]int
}

// Decode turns a g1 element into indices (raw or RLE, as stored).
func Decode(e *assets.G1Element) (*Sprite, error) {
	w, h := int(e.Width), int(e.Height)
	if w <= 0 || h <= 0 {
		return nil, errors.New("gfx8: empty sprite")
	}
	s := &Sprite{Width: w, Height: h, XOffset: int(e.XOffset), YOffset: int(e.YOffset), Pix: make([]uint8, w*h),
		HasTransparency: e.Flags&assets.G1FlagHasTransparency != 0,
		HasZoomSprites:  e.Flags&assets.G1FlagHasZoomSprites != 0, NoZoomDraw: e.Flags&assets.G1FlagNoZoomDraw != 0,
		ZoomOffset: int(e.ZoomOffset)}
	if e.Flags&assets.G1FlagIsRLECompressed == 0 {
		if len(e.Data) < w*h {
			return nil, errors.New("gfx8: raw data too short")
		}
		copy(s.Pix, e.Data[:w*h])
		s.spans = rowSpans(s)
		return s, nil
	}
	s.HasTransparency = true
	s.RLE = true
	s.Runs = make([][][2]int, h)
	d := e.Data
	if len(d) < h*2 {
		return nil, errors.New("gfx8: RLE data too short")
	}
	for y := 0; y < h; y++ {
		p := int(binary.LittleEndian.Uint16(d[y*2:]))
		for p+2 <= len(d) {
			n, x0 := int(d[p]&0x7F), int(d[p+1])
			last := d[p]&0x80 != 0
			p += 2
			if p+n > len(d) {
				break
			}
			s.Runs[y] = append(s.Runs[y], [2]int{x0, n})
			for i := 0; i < n; i++ {
				if x := x0 + i; x < w {
					s.Pix[y*w+x] = d[p+i]
				}
			}
			p += n
			if last {
				break
			}
		}
	}
	s.spans = rowSpans(s)
	return s, nil
}

// blend is OpenLoco's blend(paletteMap, src, dst): row src-1, column dst.
func blend(palMap []uint8, src, dst uint8) uint8 {
	return palMap[(int(src)-1)*256+int(dst)]
}

// blitPixel is DrawSpriteHelper.hpp blitPixel: returns whether dst changed.
func blitPixel(op BlendOp, src uint8, dst *uint8, palMap []uint8, noise uint8) bool {
	if op&OpNoiseMask != 0 {
		src &= noise
	}
	if op&OpTransparent != 0 && src == Transparent {
		return false
	}
	switch {
	case op&OpSrc != 0 && op&OpDst != 0:
		// Upstream asserts the src row exists in the map (DrawSpriteHelper.hpp
		// blend); release builds read past it. Shadow maps are one row, so their
		// images are masks of index 1. Out of range: leave dst (Go would panic).
		if (int(src)-1)*256+int(*dst) >= len(palMap) {
			return false
		}
		px := blend(palMap, src, *dst)
		if op&OpTransparent != 0 && px == Transparent {
			return false
		}
		*dst = px
	case op&OpSrc != 0:
		px := palMap[src]
		if op&OpTransparent != 0 && px == Transparent {
			return false
		}
		*dst = px
	case op&OpDst != 0:
		px := palMap[*dst]
		if op&OpTransparent != 0 && px == Transparent {
			return false
		}
		*dst = px
	default:
		*dst = src
	}
	return true
}

// OpFor picks the blend op as getDrawBlendOp does: hasPrimary = the image is
// remapped (palMap is a colour remap), blended = translucent/blend image.
func OpFor(s *Sprite, hasPrimary, blended bool) BlendOp {
	switch {
	case hasPrimary && blended:
		return OpTransparent | OpSrc | OpDst
	case hasPrimary:
		return OpTransparent | OpSrc
	case blended:
		return OpTransparent | OpDst
	case !s.HasTransparency:
		return OpNone
	}
	return OpTransparent
}

// DrawSprite draws s with its offsets at (x, y) on rt (zoom 0), clipped.
func DrawSprite(rt *RenderTarget, s *Sprite, x, y int, op BlendOp, palMap []uint8) {
	DrawSpriteNoise(rt, s, x, y, op, palMap, nil)
}

// DrawSpriteNoise is DrawSprite with an optional noise image (ImageId
// withNoiseMask). As upstream (DrawSprite.cpp getDrawBlendOp) the noise op only
// applies to non-RLE sprites, and (DrawSpriteBMP.hpp) the noise byte for sprite
// pixel (x, y) is read at noise[sprite.Width*y + x] — the sprite's width is the
// stride into the noise image.
func DrawSpriteNoise(rt *RenderTarget, s *Sprite, x, y int, op BlendOp, palMap []uint8, noise *Sprite) {
	if noise != nil && !s.RLE {
		op |= OpNoiseMask
	} else {
		op &^= OpNoiseMask
	}
	x0, y0 := x+s.XOffset, y+s.YOffset
	// Clip the sprite rectangle to the target once (same pixels as testing every
	// pixel against the bounds).
	sxMin, sxMax := max(0, -x0), min(s.Width, rt.Width-x0)
	syMin, syMax := max(0, -y0), min(s.Height, rt.Height-y0)
	if sxMin >= sxMax || syMin >= syMax {
		return
	}
	// Fast path for the commonest op, a transparent copy: blitPixel with
	// OpTransparent alone writes src wherever src != 0.
	if op == OpTransparent && len(s.spans) == s.Height {
		for sy := syMin; sy < syMax; sy++ {
			base, src := (y0+sy)*rt.Width+x0, s.Pix[sy*s.Width:]
			for _, r := range s.spans[sy] {
				a, b := max(int(r[0]), sxMin), min(int(r[1]), sxMax)
				if a < b {
					copy(rt.Bits[base+a:base+b], src[a:b])
				}
			}
		}
		return
	}
	if op == OpTransparent {
		for sy := syMin; sy < syMax; sy++ {
			src := s.Pix[sy*s.Width+sxMin : sy*s.Width+sxMax]
			dst := rt.Bits[(y0+sy)*rt.Width+x0+sxMin:]
			for i, p := range src {
				if p != Transparent {
					dst[i] = p
				}
			}
		}
		return
	}
	// The other common ops without noise, each blitPixel's case as its own
	// loop (no call per pixel); anything else goes through blitPixel below.
	if fast := drawRowsFast(rt, s, x0, y0, sxMin, sxMax, syMin, syMax, op, palMap); fast {
		return
	}
	for sy := syMin; sy < syMax; sy++ {
		row := (y0 + sy) * rt.Width
		for sx := sxMin; sx < sxMax; sx++ {
			nm := uint8(0xFF)
			if op&OpNoiseMask != 0 {
				if i := s.Width*sy + sx; i < len(noise.Pix) {
					nm = noise.Pix[i]
				} else {
					nm = 0
				}
			}
			blitPixel(op, s.Pix[sy*s.Width+sx], &rt.Bits[row+x0+sx], palMap, nm)
		}
	}
}

// DrawSpriteMasked is drawImageMasked at zoom 0 (SoftwareDrawingContext.cpp
// drawImageMaskedZoom/drawMaskedZoom): the image is placed by its own offsets,
// the mask is read at the same image-local positions with the image's row
// width, and dst = mask & image wherever that is non-zero. Upstream supports
// only raw (non-RLE) images here; the mask is assumed to share the layout.
func DrawSpriteMasked(rt *RenderTarget, img, mask *Sprite, x, y int) {
	x0, y0 := x+img.XOffset, y+img.YOffset
	// Clipped once per sprite and row (the same pixels as testing each one).
	sxMin, sxMax := max(0, -x0), min(img.Width, rt.Width-x0)
	syMin, syMax := max(0, -y0), min(img.Height, rt.Height-y0)
	if sxMin >= sxMax || syMin >= syMax {
		return
	}
	for sy := syMin; sy < syMax; sy++ {
		lo, hi := sy*img.Width+sxMin, min(sy*img.Width+sxMax, len(mask.Pix))
		if lo >= hi {
			continue
		}
		m, src := mask.Pix[lo:hi], img.Pix[lo:hi]
		dst := rt.Bits[(y0+sy)*rt.Width+x0+sxMin:]
		dst = dst[:len(m)]
		for i := range m {
			if v := m[i] & src[i]; v != 0 {
				dst[i] = v
			}
		}
	}
}

// drawRowsFast draws the clipped rows for the ops that need no noise byte and
// always have OpTransparent (the remap, the translucent blend, both) or are a
// plain copy, exactly as blitPixel would; false = not one of those.
func drawRowsFast(rt *RenderTarget, s *Sprite, x0, y0, sxMin, sxMax, syMin, syMax int, op BlendOp, palMap []uint8) bool {
	switch op {
	case OpNone:
		for sy := syMin; sy < syMax; sy++ {
			copy(rt.Bits[(y0+sy)*rt.Width+x0+sxMin:], s.Pix[sy*s.Width+sxMin:sy*s.Width+sxMax])
		}
	case OpTransparent | OpSrc:
		pm := palMap[:256]
		for sy := syMin; sy < syMax; sy++ {
			src := s.Pix[sy*s.Width+sxMin : sy*s.Width+sxMax]
			dst := rt.Bits[(y0+sy)*rt.Width+x0+sxMin:]
			dst = dst[:len(src)]
			for i, p := range src {
				if p != Transparent {
					if px := pm[p]; px != Transparent {
						dst[i] = px
					}
				}
			}
		}
	case OpTransparent | OpDst:
		pm := palMap[:256]
		for sy := syMin; sy < syMax; sy++ {
			src := s.Pix[sy*s.Width+sxMin : sy*s.Width+sxMax]
			dst := rt.Bits[(y0+sy)*rt.Width+x0+sxMin:]
			dst = dst[:len(src)]
			for i, p := range src {
				if p != Transparent {
					if px := pm[dst[i]]; px != Transparent {
						dst[i] = px
					}
				}
			}
		}
	case OpTransparent | OpSrc | OpDst:
		n := len(palMap)
		for sy := syMin; sy < syMax; sy++ {
			src := s.Pix[sy*s.Width+sxMin : sy*s.Width+sxMax]
			dst := rt.Bits[(y0+sy)*rt.Width+x0+sxMin:]
			dst = dst[:len(src)]
			for i, p := range src {
				if p != Transparent {
					if k := (int(p)-1)*256 + int(dst[i]); k < n {
						if px := palMap[k]; px != Transparent {
							dst[i] = px
						}
					}
				}
			}
		}
	default:
		return false
	}
	return true
}

// rowSpans finds each row's runs of non-transparent pixels.
func rowSpans(s *Sprite) [][][2]int32 {
	out := make([][][2]int32, s.Height)
	for y := range s.Height {
		row := s.Pix[y*s.Width : (y+1)*s.Width]
		var spans [][2]int32
		for x := 0; x < len(row); {
			for x < len(row) && row[x] == Transparent {
				x++
			}
			a := x
			for x < len(row) && row[x] != Transparent {
				x++
			}
			if a < x {
				spans = append(spans, [2]int32{int32(a), int32(x)})
			}
		}
		out[y] = spans
	}
	return out
}
