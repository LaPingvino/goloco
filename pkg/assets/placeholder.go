package assets

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// GeneratePlaceholderAtlas creates a simple test sprite atlas with basic tiles
// This is used for development/testing before real asset extraction is working
func GeneratePlaceholderAtlas(outputPath string) error {
	// Create output directory
	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		return err
	}

	// Create a simple 1024x1024 atlas with a grid of basic sprites
	atlas := image.NewRGBA(image.Rect(0, 0, 1024, 1024))

	// Fill with background (light gray)
	bgColor := color.RGBA{200, 200, 200, 255}
	for y := 0; y < 1024; y++ {
		for x := 0; x < 1024; x++ {
			atlas.SetRGBA(x, y, bgColor)
		}
	}

	// Draw a grid of 64x64 tiles (16x16 grid = 256 tiles)
	tileSize := 64
	colors := []color.RGBA{
		{100, 200, 100, 255}, // green
		{100, 150, 200, 255}, // blue
		{200, 100, 100, 255}, // red
		{200, 200, 100, 255}, // yellow
		{150, 100, 150, 255}, // purple
		{100, 200, 200, 255}, // cyan
	}

	for ty := 0; ty < 16; ty++ {
		for tx := 0; tx < 16; tx++ {
			colorIdx := (tx + ty*16) % len(colors)
			col := colors[colorIdx]

			// Draw tile
			x0 := tx * tileSize
			y0 := ty * tileSize
			for y := y0; y < y0+tileSize && y < 1024; y++ {
				for x := x0; x < x0+tileSize && x < 1024; x++ {
					atlas.SetRGBA(x, y, col)
				}
			}

			// Draw border
			borderColor := color.RGBA{50, 50, 50, 255}
			for i := 0; i < tileSize; i++ {
				if x0+i < 1024 {
					atlas.SetRGBA(x0+i, y0, borderColor)
					atlas.SetRGBA(x0+i, y0+tileSize-1, borderColor)
				}
				if y0+i < 1024 {
					atlas.SetRGBA(x0, y0+i, borderColor)
					atlas.SetRGBA(x0+tileSize-1, y0+i, borderColor)
				}
			}
		}
	}

	// Save atlas as PNG
	outFile := filepath.Join(outputPath, "placeholder_atlas.png")
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	err = png.Encode(f, atlas)
	if err != nil {
		return err
	}

	return nil
}
