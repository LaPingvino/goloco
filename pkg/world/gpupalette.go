package world

import (
	"image"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// The world's 8-bit frame turned into colours on the GPU (LOOP item FPS (3)):
// instead of converting every pixel through the palette on the CPU and
// uploading 4 bytes a pixel, the palette indices are uploaded as they are —
// four to a texel, in the left quarter of one image — with the 256-colour
// palette in row 0 just right of them, and a shader looks each pixel's colour
// up. A quarter of the upload, no conversion loop. The picture is the same
// (index 0 drawn black, as ToRGBA does). GOLOCO_GPUPAL=0 goes back to the CPU.
var gpuPaletteOn = os.Getenv("GOLOCO_GPUPAL") != "0"

const gpuPaletteShaderSrc = `//kage:unit pixels

package main

// QW is the width of the packed-index area in texels (ceil(width / 4)); the
// palette's 256 entries sit at x = QW … QW + 255 of row 0.
var QW float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	p := dstPos.xy - imageDstOrigin()
	x := floor(p.x)
	t := imageSrc0At(imageSrc0Origin() + vec2(floor(x/4)+0.5, p.y))
	m := mod(x, 4)
	v := t.r
	if m >= 3 {
		v = t.a
	} else if m >= 2 {
		v = t.b
	} else if m >= 1 {
		v = t.g
	}
	i := floor(v*255 + 0.5)
	c := imageSrc0At(imageSrc0Origin() + vec2(QW+i+0.5, 0.5))
	return vec4(c.rgb, 1)
}
`

var gpuPaletteShader *ebiten.Shader

// gpuPal is the packed frame and its palette for one screen size.
type gpuPal struct {
	img    *ebiten.Image
	w, h   int
	qw     int
	palSet bool
	buf    []byte
}

// gpuPaletteFits reports whether a w-wide screen has room for the palette row
// beside the packed indices (QW + 256 ≤ w: from 342 pixels wide).
func gpuPaletteFits(w int) bool { return (w+3)/4+256 <= w }

func (w *World) gpuPalFor(sw, sh int) *gpuPal {
	if gpuPaletteShader == nil {
		s, err := ebiten.NewShader([]byte(gpuPaletteShaderSrc))
		if err != nil {
			gpuPaletteOn = false // no shader: the CPU path
			return nil
		}
		gpuPaletteShader = s
		log.Printf("[Scene] palette on the GPU (GOLOCO_GPUPAL=0: on the CPU)")
	}
	g := w.gpal
	if g == nil || g.w != sw || g.h != sh {
		if g != nil {
			g.img.Deallocate()
		}
		g = &gpuPal{img: ebiten.NewImageWithOptions(image.Rect(0, 0, sw, sh), &ebiten.NewImageOptions{Unmanaged: true}), w: sw, h: sh, qw: (sw + 3) / 4}
		w.gpal = g
	}
	return g
}

// setPalette writes the palette row (index 0 black, as ToRGBA).
func (g *gpuPal) setPalette(pal *[256]color.RGBA) {
	b := make([]byte, 256*4)
	for i, c := range pal {
		b[i*4], b[i*4+1], b[i*4+2], b[i*4+3] = c.R, c.G, c.B, 255
	}
	b[0], b[1], b[2] = 0, 0, 0
	g.img.SubImage(rect(g.qw, 0, g.qw+256, 1)).(*ebiten.Image).WritePixels(b)
	g.palSet = true
}

// upload writes the indices of screen columns [x0, x1) (widened to whole
// texels), straight from the 8-bit frame.
func (g *gpuPal) upload(rt *gfx8.RenderTarget, x0, x1 int) {
	a, b := x0/4, (x1+3)/4
	n := (b - a) * 4 * g.h
	if cap(g.buf) < n {
		g.buf = make([]byte, n)
	}
	buf := g.buf[:n]
	row := (b - a) * 4
	for y := 0; y < g.h; y++ {
		src := rt.Bits[y*rt.Width:]
		dst := buf[y*row : (y+1)*row]
		lo, hi := a*4, min(b*4, rt.Width)
		copy(dst, src[lo:hi])
		for i := hi - lo; i < row; i++ {
			dst[i] = 0
		}
	}
	g.img.SubImage(rect(a, 0, b, g.h)).(*ebiten.Image).WritePixels(buf)
}

// draw paints the frame onto the screen through the palette shader.
func (g *gpuPal) draw(screen *ebiten.Image) {
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = g.img
	op.Uniforms = map[string]any{"QW": float32(g.qw)}
	screen.DrawRectShader(g.w, g.h, gpuPaletteShader, op)
}

func rect(x0, y0, x1, y1 int) image.Rectangle { return image.Rect(x0, y0, x1, y1) }
