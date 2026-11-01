package world

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

// Palette-map blending: the draw op OpenLoco uses for an image that is blended
// but has no primary colour — "only a transparency mask: just colour the
// pixels using the palette map" (Graphics/DrawSprite.cpp getDrawBlendOp,
// DrawBlendOp::transparent|dst). For every non-transparent pixel of the mask,
// the destination pixel's palette index i becomes map[i]. Water uses it with
// the water object's palette map (image +41, WaterObject::kColourPalette).
//
// goloco draws in RGBA, so the shader recovers i by matching the destination
// colour against the 256 palette colours, then writes palette[map[i]]. At zoom
// 0 every drawn pixel is an exact palette colour, so this is exact there.

// blendShaderSrc: imageSrc0 = mask (alpha only), imageSrc1 = destination (the
// terrain already drawn, same region), imageSrc2 = lookup table: entry i (0-255)
// at (i%64, i/64) holds palette[i], at (i%64, 4+i/64) holds palette[map[i]].
const blendShaderSrc = `//kage:unit pixels
package main

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	if imageSrc0At(src).a == 0 {
		return vec4(0)
	}
	// src is in image 0's coordinates; the other images have their own origins.
	c := imageSrc1At(src - imageSrc0Origin() + imageSrc1Origin())
	o := imageSrc2Origin()
	best := 0
	bestD := 10.0
	for i := 0; i < 256; i++ {
		p := imageSrc2At(o + vec2(float(i%64)+0.5, float(i/64)+0.5))
		d := abs(p.r-c.r) + abs(p.g-c.g) + abs(p.b-c.b)
		if d < bestD {
			bestD = d
			best = i
		}
	}
	m := imageSrc2At(o + vec2(float(best%64)+0.5, float(4+best/64)+0.5))
	return vec4(m.rgb, 1)
}
`

// exactWaterBlend switches water to the exact palette-map blend. OFF until the
// shader reads the terrain correctly: in a headless WebGL (swiftshader) run,
// imageSrc1At on the worldBuf sub-image returned black although the seabed is
// drawn there (debug variant returning c.rgb was black). Next step: check the
// sub-image origin mapping in pixel mode (Ebiten 2.9) with a tiny repro.
const exactWaterBlend = false

var (
	blendShader     *ebiten.Shader
	blendShaderErr  error
	blendShaderDone bool
)

func paletteBlendShader() *ebiten.Shader {
	if !blendShaderDone {
		blendShaderDone = true
		blendShader, blendShaderErr = ebiten.NewShader([]byte(blendShaderSrc))
		if blendShaderErr != nil {
			log.Printf("[Render] palette blend shader: %v", blendShaderErr)
		}
	}
	return blendShader
}

// blendTableKey identifies a lookup table: the map and the image size it must
// match (shader source images must all be the same size).
type blendTableKey struct {
	mapPtr *byte
	w, h   int
}

var blendTables = map[blendTableKey]*ebiten.Image{}

// paletteBlendTable returns a w×h image holding the palette (rows 0-3) and the
// palette after the map (rows 4-7), or nil if w×h can't hold them (w ≥ 64, h ≥ 8).
func paletteBlendTable(palette *[256]color.RGBA, palMap []byte, w, h int) *ebiten.Image {
	if len(palMap) < 256 || w < 64 || h < 8 {
		return nil
	}
	key := blendTableKey{&palMap[0], w, h}
	if img, ok := blendTables[key]; ok {
		return img
	}
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < 256; i++ {
		p := palette[i]
		p.A = 255
		rgba.SetRGBA(i%64, i/64, p)
		m := palette[palMap[i]]
		m.A = 255
		rgba.SetRGBA(i%64, 4+i/64, m)
	}
	img := ebiten.NewImageFromImage(rgba)
	blendTables[key] = img
	return img
}

// applyPaletteBlends runs every queued blend op: reads the destination region
// from src (the frame drawn so far) and writes the mapped pixels into dst.
func (w *World) applyPaletteBlends(src, dst *ebiten.Image) {
	sh := paletteBlendShader()
	if sh == nil {
		return
	}
	for _, op := range w.blendOps {
		if op.maskImg == nil || op.table == nil || op.tileRect.Empty() {
			continue
		}
		r := op.tileRect
		if !r.In(src.Bounds()) {
			r = r.Intersect(src.Bounds())
			if r.Dx() != op.tileRect.Dx() || r.Dy() != op.tileRect.Dy() {
				continue // partly off screen: all source regions must match in size
			}
		}
		opts := &ebiten.DrawRectShaderOptions{}
		opts.Images[0] = op.maskImg
		opts.Images[1] = src.SubImage(r).(*ebiten.Image)
		opts.Images[2] = op.table
		opts.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
		dst.DrawRectShader(r.Dx(), r.Dy(), sh, opts)
	}
}
