package graphics

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// RenderTarget wraps an Ebiten image with OpenLoco-style interface
type RenderTarget struct {
	Image  *ebiten.Image
	Width  int16
	Height int16
	Pitch  int32 // bytes per row
}

// NewRenderTarget creates a RenderTarget from an Ebiten image
func NewRenderTarget(img *ebiten.Image) *RenderTarget {
	bounds := img.Bounds()
	return &RenderTarget{
		Image:  img,
		Width:  int16(bounds.Dx()),
		Height: int16(bounds.Dy()),
		Pitch:  int32(bounds.Dx() * 4), // 4 bytes per pixel (RGBA)
	}
}

// NewRenderTargetWithSize creates a new RenderTarget with specified dimensions
func NewRenderTargetWithSize(width, height int16) *RenderTarget {
	img := ebiten.NewImage(int(width), int(height))
	return NewRenderTarget(img)
}
