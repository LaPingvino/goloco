package scenario

import "math"

// initGeneratedLandscape creates a heightmap from the scenario's landscape-generation
// parameters (minLandHeight, topographyStyle, hillDensity) when the tile data is not
// present in the .SC5 file (landscapeGenerationDone flag is clear).
//
// This is a simplified approximation — OpenLoco's exact RNG and multi-pass generation
// is not replicated, but the result is a plausible landscape with matching characteristics.
//
// OpenLoco reference: src/OpenLoco/src/Scenario/LandscapeGenerator.cpp
func (sc *Scenario) initGeneratedLandscape() {
	w, h := sc.MapWidth, sc.MapHeight
	if w <= 0 || h <= 0 {
		w, h = DefaultMapSize, DefaultMapSize
		sc.MapWidth, sc.MapHeight = w, h
	}

	// Base height: OpenLoco stores minLandHeight as 0-15 (in SmallZ units).
	baseH := int(sc.Options.MinLandHeight)
	if baseH < 0 {
		baseH = 0
	}

	// Hill amplitude: scaled by hillDensity (0-100) and topography style.
	// Each SmallZ unit = 4 viewport pixels; typical terrain is 0-20 units.
	hillAmp := float64(sc.Options.HillDensity) * 0.18 // max ~18 units at density=100
	switch sc.Options.TopographyStyle {
	case 0: // flatLand
		hillAmp *= 0.1
	case 1: // smallHills
		hillAmp *= 0.5
	case 2: // mountains
		hillAmp *= 1.5
	case 3: // halfMountainsHills
		hillAmp *= 1.0
	case 4: // halfMountainsFlat
		hillAmp *= 0.7
	}

	// Use sum-of-sines noise so it's deterministic (no random seed needed).
	// Multiple frequencies give a natural-looking ridge pattern.
	heights := make([][]float64, h)
	for y := 0; y < h; y++ {
		heights[y] = make([]float64, w)
		for x := 0; x < w; x++ {
			fx := float64(x) / float64(w)
			fy := float64(y) / float64(h)
			// Sum of overlapping sinusoids at different scales/angles.
			v := math.Sin(fx*6.28*3+fy*6.28*2) * 0.4
			v += math.Sin(fx*6.28*7+fy*6.28*5) * 0.25
			v += math.Sin(fx*6.28*13+fy*6.28*11) * 0.15
			v += math.Cos(fx*6.28*4+fy*6.28*9) * 0.20
			heights[y][x] = v // range roughly -1..+1
		}
	}

	// Build tiles.
	sc.Tiles = make([][]Tile, h)
	for y := 0; y < h; y++ {
		sc.Tiles[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			// Convert noise to integer SmallZ height.
			elev := baseH + int(math.Round(heights[y][x]*hillAmp))
			if elev < 0 {
				elev = 0
			}
			if elev > 28 {
				elev = 28 // max terrain height in Locomotion
			}

			surface := SurfaceGrass
			// Water border (3 tiles wide).
			if x < 3 || x >= w-3 || y < 3 || y >= h-3 {
				surface = SurfaceWater
				elev = 0
			} else if elev == 0 {
				surface = SurfaceWater
			}

			sc.Tiles[y][x] = Tile{
				Surface: surface,
				Height:  uint8(elev),
			}
		}
	}
}
