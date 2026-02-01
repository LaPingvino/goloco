package graphics

import "log"

// Color palette and shade system for OpenLoco-style graphics

// Color represents a palette color index (0-255)
type Color uint8

// Standard OpenLoco colors
const (
	ColorBlack Color = iota
	ColorGrey
	ColorWhite
	ColorDarkPurple
	ColorLightPurple
	ColorBrightPurple
	ColorDarkBlue
	ColorLightBlue
	ColorIcyBlue
	ColorTeal
	ColorAquamarine
	ColorSaturatedGreen
	ColorDarkGreen
	ColorMossGreen
	ColorBrightGreen
	ColorOliveGreen
	ColorDarkOliveGreen
	ColorBrightYellow
	ColorYellow
	ColorMossYellow
	ColorDarkYellow
	ColorLightOrange
	ColorDarkOrange
	ColorLightBrown
	ColorSaturatedBrown
	ColorDarkBrown
	ColorSalmon
	ColorButterflyRed
	ColorDarkRed
	ColorBrightRed
	ColorDarkPink
	ColorBrightPink
	ColorTransparent = 0
)

// Color shade maps (simplified - in real OpenLoco these are much more complex)
var colorShadeMap = [32][8]uint8{
	// Black shades
	{0, 1, 2, 3, 4, 5, 6, 7},
	// Grey shades
	{8, 9, 10, 11, 12, 13, 14, 15},
	// White shades
	{16, 17, 18, 19, 20, 21, 22, 23},
	// Add more color shade progressions...
	// For now, simplified linear progression
}

// GetShade returns the palette index for a color at a specific shade level (0-7)
// This mimics OpenLoco's Colour::getShade function
// Ported from: Graphics/Colour.cpp
func GetShade(colour Color, shade uint8) uint8 {
	// Bounds check
	if colour > 31 {
		colour = ColorBlack
	}

	if shade < 8 {
		// Use primary shade map
		if int(colour) < len(colorShadeMap) {
			return colorShadeMap[colour][shade]
		}
		return uint8(colour)*8 + shade
	}

	// Extended shades (8-15) - use secondary map
	// For now, just offset from base
	return uint8(colour)*16 + shade
}

// GetTranslucent returns a translucent version of the color
func GetTranslucent(colour Color) Color {
	// Simplified - real OpenLoco has complex translucency tables
	return colour + 128
}

// InitColorMaps initializes the color shade mapping tables. Stub — the
// tables are populated with a synthetic linear gradient; they do not yet
// match the real Locomotion palette shade progressions.
//
// OpenLoco reference: src/OpenLoco/src/Graphics/Colour.cpp
//   initColourMap()
//
// In OpenLoco, initColourMap() fills _colourMapA and _colourMapB from
// hard-coded lookup tables that were reverse-engineered from the original
// Locomotion binary.  A third table, _translucentColourMap, stores the
// three ExtColour indices used for translucent rendering per colour.
// When the real tables are available (or extracted from G1.DAT palette
// data), replace the loop below with direct initialisation from those
// tables.
func InitColorMaps() {
	log.Println("[Graphics] InitColorMaps: stub — using synthetic linear shade progression")
	for c := 0; c < 32; c++ {
		for s := 0; s < 8; s++ {
			if c < len(colorShadeMap) {
				colorShadeMap[c][s] = uint8(c*8 + s)
			}
		}
	}
}
