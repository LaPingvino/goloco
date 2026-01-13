// Package graphics - stub implementations to allow build
package graphics

// G1Element represents a game sprite/image
type G1Element struct {
	Offset     []byte
	Width      int16
	Height     int16
	XOffset    int16
	YOffset    int16
	Flags      G1ElementFlags
	ZoomOffset int16
}

// G1ElementFlags represent image properties
type G1ElementFlags uint16

const (
	G1FlagsNone         G1ElementFlags = 0
	G1HasTransparency   G1ElementFlags = 1 << 0
	G1IsRLECompressed   G1ElementFlags = 1 << 2
	G1HasZoomSprites    G1ElementFlags = 1 << 4
	G1DuplicatePrevious G1ElementFlags = 1 << 6
)

// G1Header represents the header of a G1 image table
type G1Header struct {
	NumEntries uint32
	TotalSize  uint32
}

// Colour represents an RGB colour
type Colour struct {
	R, G, B, A uint8
}

// PaletteEntry is a single palette color
type PaletteEntry struct {
	B, G, R, A uint8
}

// Global palette
var GlobalPalette [256]PaletteEntry

// LoadG1 loads the G1 image table (stub)
func LoadG1() error {
	return nil
}

// GetG1Element returns a specific G1 element by ID (stub)
func GetG1Element(id uint32) *G1Element {
	return &G1Element{}
}
