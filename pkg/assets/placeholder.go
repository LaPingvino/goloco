package assets

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// GeneratePlaceholderAtlas creates isometric diamond-shaped tiles for testing
func GeneratePlaceholderAtlas(outputPath string) error {
	// Create output directories
	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		return err
	}
	extractedPath := filepath.Join(outputPath, "extracted")
	err = os.MkdirAll(extractedPath, 0755)
	if err != nil {
		return err
	}

	// Create isometric diamond tiles (64 wide x 32 tall - standard 2:1 ratio)
	tileW := 64
	tileH := 32

	// Generate grass tile (green diamond)
	grassTile := createDiamondTile(tileW, tileH,
		color.RGBA{80, 160, 80, 255},   // main color
		color.RGBA{60, 140, 60, 255},   // darker edge
		color.RGBA{100, 180, 100, 255}) // lighter highlight
	err = savePNG(filepath.Join(extractedPath, "grass.png"), grassTile)
	if err != nil {
		return err
	}

	// Generate dirt tile (brown diamond)
	dirtTile := createDiamondTile(tileW, tileH,
		color.RGBA{139, 90, 43, 255},  // main color
		color.RGBA{100, 65, 30, 255},  // darker edge
		color.RGBA{170, 120, 70, 255}) // lighter highlight
	err = savePNG(filepath.Join(extractedPath, "dirt.png"), dirtTile)
	if err != nil {
		return err
	}

	// Generate water tile (blue diamond)
	waterTile := createDiamondTile(tileW, tileH,
		color.RGBA{64, 120, 192, 255},  // main color
		color.RGBA{40, 90, 160, 255},   // darker edge
		color.RGBA{100, 160, 220, 255}) // lighter highlight
	err = savePNG(filepath.Join(extractedPath, "water.png"), waterTile)
	if err != nil {
		return err
	}

	return nil
}

// createDiamondTile creates an isometric diamond-shaped tile
func createDiamondTile(w, h int, mainColor, darkColor, lightColor color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with transparent
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Diamond shape: for each row, calculate the horizontal extent
	// At y=0, we're at the top point (center)
	// At y=h/2, we're at the widest point (full width)
	// At y=h, we're at the bottom point (center)
	centerX := w / 2
	centerY := h / 2

	for y := 0; y < h; y++ {
		// Calculate horizontal extent at this row
		var halfWidth int
		if y < centerY {
			// Top half: expanding from center
			halfWidth = (y * centerX) / centerY
		} else {
			// Bottom half: contracting to center
			halfWidth = ((h - 1 - y) * centerX) / centerY
		}

		for x := centerX - halfWidth; x <= centerX+halfWidth; x++ {
			if x < 0 || x >= w {
				continue
			}

			// Determine color based on position (simple shading)
			var c color.RGBA
			if y < centerY {
				// Top half - use lighter color on left, main on right
				if x < centerX {
					c = lightColor
				} else {
					c = mainColor
				}
			} else {
				// Bottom half - use main on left, darker on right
				if x < centerX {
					c = mainColor
				} else {
					c = darkColor
				}
			}

			// Edge detection for outline
			isEdge := false
			if y < centerY {
				expectedHalfWidth := (y * centerX) / centerY
				if x == centerX-expectedHalfWidth || x == centerX+expectedHalfWidth {
					isEdge = true
				}
			} else {
				expectedHalfWidth := ((h - 1 - y) * centerX) / centerY
				if x == centerX-expectedHalfWidth || x == centerX+expectedHalfWidth {
					isEdge = true
				}
			}

			if isEdge {
				// Draw darker outline
				c = color.RGBA{
					uint8(max(0, int(c.R)-40)),
					uint8(max(0, int(c.G)-40)),
					uint8(max(0, int(c.B)-40)),
					255,
				}
			}

			img.SetRGBA(x, y, c)
		}
	}

	return img
}

func savePNG(path string, img *image.RGBA) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
