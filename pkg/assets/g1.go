package assets

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
)

// G1 file format constants
const (
	G1ExpectedCountDisc  = 0x0F4A
	G1ExpectedCountSteam = 0x0F38
	DefaultPaletteIndex  = 304
)

// G1ElementFlags for sprite properties
type G1ElementFlags uint16

const (
	G1FlagNone            G1ElementFlags = 0
	G1FlagHasTransparency G1ElementFlags = 1 << 0
	G1FlagUnk1            G1ElementFlags = 1 << 1
	G1FlagIsRLECompressed G1ElementFlags = 1 << 2
	G1FlagIsR8G8B8Palette G1ElementFlags = 1 << 3
	G1FlagHasZoomSprites  G1ElementFlags = 1 << 4
	G1FlagNoZoomDraw      G1ElementFlags = 1 << 5
	G1FlagDuplicatePrev   G1ElementFlags = 1 << 6
)

// G1Header is the file header
type G1Header struct {
	NumEntries uint32
	TotalSize  uint32
}

// G1Element32 is the on-disk element header (16 bytes)
type G1Element32 struct {
	Offset     uint32
	Width      int16
	Height     int16
	XOffset    int16
	YOffset    int16
	Flags      G1ElementFlags
	ZoomOffset int16
}

// G1Element is a parsed sprite element with data pointer
type G1Element struct {
	Width      int16
	Height     int16
	XOffset    int16
	YOffset    int16
	Flags      G1ElementFlags
	ZoomOffset int16
	Data       []byte // raw pixel/RLE data
}

// G1File holds all parsed G1 data
type G1File struct {
	Header   G1Header
	Elements []G1Element
	Palette  [256]color.RGBA
}

// LoadG1 loads and parses a G1.DAT file
func LoadG1(path string) (*G1File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open g1: %w", err)
	}
	defer f.Close()

	// Read header
	var header G1Header
	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	fmt.Printf("G1: %d entries, %d bytes total\n", header.NumEntries, header.TotalSize)

	// Validate entry count
	if header.NumEntries != G1ExpectedCountDisc && header.NumEntries != G1ExpectedCountSteam {
		fmt.Printf("Warning: unexpected G1 entry count %d (expected %d or %d)\n",
			header.NumEntries, G1ExpectedCountDisc, G1ExpectedCountSteam)
	}

	// Read element headers
	elements32 := make([]G1Element32, header.NumEntries)
	for i := uint32(0); i < header.NumEntries; i++ {
		if err := binary.Read(f, binary.LittleEndian, &elements32[i]); err != nil {
			return nil, fmt.Errorf("read element %d: %w", i, err)
		}
	}

	// Read all element data
	elementData := make([]byte, header.TotalSize)
	if _, err := io.ReadFull(f, elementData); err != nil {
		return nil, fmt.Errorf("read element data: %w", err)
	}

	// Convert elements and extract data slices
	g1 := &G1File{
		Header:   header,
		Elements: make([]G1Element, header.NumEntries),
	}

	for i := uint32(0); i < header.NumEntries; i++ {
		e32 := &elements32[i]
		elem := &g1.Elements[i]

		elem.Width = e32.Width
		elem.Height = e32.Height
		elem.XOffset = e32.XOffset
		elem.YOffset = e32.YOffset
		elem.Flags = e32.Flags
		elem.ZoomOffset = e32.ZoomOffset

		// Calculate data bounds
		startOff := e32.Offset
		var endOff uint32
		if i+1 < header.NumEntries {
			endOff = elements32[i+1].Offset
		} else {
			endOff = header.TotalSize
		}

		if startOff < header.TotalSize && endOff <= header.TotalSize && startOff < endOff {
			elem.Data = elementData[startOff:endOff]
		}
	}

	// Load palette from element 304
	if err := g1.loadPalette(); err != nil {
		fmt.Printf("Warning: failed to load palette: %v\n", err)
		// Use a default grayscale palette as fallback
		for i := 0; i < 256; i++ {
			g1.Palette[i] = color.RGBA{uint8(i), uint8(i), uint8(i), 255}
		}
	}

	return g1, nil
}

// loadPalette extracts the palette from the default palette element
func (g1 *G1File) loadPalette() error {
	if DefaultPaletteIndex >= len(g1.Elements) {
		return errors.New("palette index out of range")
	}

	elem := &g1.Elements[DefaultPaletteIndex]
	if elem.Flags&G1FlagIsR8G8B8Palette == 0 {
		return errors.New("palette element doesn't have R8G8B8 flag")
	}

	// Palette data is sequential R, G, B bytes
	// xOffset indicates starting palette index, width indicates count
	startIdx := int(elem.XOffset)
	count := int(elem.Width)

	if len(elem.Data) < count*3 {
		return fmt.Errorf("palette data too small: %d < %d", len(elem.Data), count*3)
	}

	for i := 0; i < count; i++ {
		idx := startIdx + i
		if idx >= 0 && idx < 256 {
			// Data is R, G, B order
			g1.Palette[idx] = color.RGBA{
				R: elem.Data[i*3+0],
				G: elem.Data[i*3+1],
				B: elem.Data[i*3+2],
				A: 255,
			}
		}
	}

	// Make index 0 transparent (typically the transparent color)
	g1.Palette[0] = color.RGBA{0, 0, 0, 0}

	return nil
}

// DecodeSprite decodes a sprite element to an RGBA image
func (g1 *G1File) DecodeSprite(index int) (*image.RGBA, error) {
	if index < 0 || index >= len(g1.Elements) {
		return nil, fmt.Errorf("sprite index %d out of range", index)
	}

	elem := &g1.Elements[index]
	if elem.Width <= 0 || elem.Height <= 0 {
		return nil, fmt.Errorf("invalid sprite dimensions: %dx%d", elem.Width, elem.Height)
	}

	w := int(elem.Width)
	h := int(elem.Height)

	// Create output image
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Check if RLE compressed
	if elem.Flags&G1FlagIsRLECompressed != 0 {
		if err := g1.decodeRLE(elem, img); err != nil {
			return nil, fmt.Errorf("decode RLE: %w", err)
		}
	} else {
		// Raw palette indices
		if err := g1.decodeRaw(elem, img); err != nil {
			return nil, fmt.Errorf("decode raw: %w", err)
		}
	}

	return img, nil
}

// decodeRaw decodes uncompressed palette index data
func (g1 *G1File) decodeRaw(elem *G1Element, img *image.RGBA) error {
	w := int(elem.Width)
	h := int(elem.Height)

	if len(elem.Data) < w*h {
		return fmt.Errorf("raw data too small: %d < %d", len(elem.Data), w*h)
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := elem.Data[y*w+x]
			img.SetRGBA(x, y, g1.Palette[idx])
		}
	}

	return nil
}

// decodeRLE decodes RLE compressed sprite data
func (g1 *G1File) decodeRLE(elem *G1Element, img *image.RGBA) error {
	w := int(elem.Width)
	h := int(elem.Height)
	data := elem.Data

	// RLE format: first h*2 bytes are uint16 LE line offsets
	if len(data) < h*2 {
		return errors.New("RLE data too small for line offsets")
	}

	for y := 0; y < h; y++ {
		lineOffset := int(binary.LittleEndian.Uint16(data[y*2 : y*2+2]))
		if lineOffset >= len(data) {
			continue // Skip invalid lines
		}

		p := lineOffset
		for {
			if p+2 > len(data) {
				break
			}

			dataSize := int(data[p])
			firstX := int(data[p+1])
			p += 2

			isLast := (dataSize & 0x80) != 0
			dataSize &= 0x7F

			if p+dataSize > len(data) {
				break
			}

			// Copy pixel data
			for i := 0; i < dataSize; i++ {
				x := firstX + i
				if x >= 0 && x < w {
					idx := data[p+i]
					img.SetRGBA(x, y, g1.Palette[idx])
				}
			}
			p += dataSize

			if isLast {
				break
			}
		}
	}

	return nil
}

// GetSpriteCount returns the number of sprites
func (g1 *G1File) GetSpriteCount() int {
	return len(g1.Elements)
}
