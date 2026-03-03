package objects

import "fmt"

// WaterObject represents the water surface sprites.
// OpenLoco reference: src/OpenLoco/src/Objects/WaterObject.h
// sizeof(WaterObject) == 0xE
//
// Water sprite index layout (relative to ImageOffset):
//
//	0-29:  misc sprites
//	30-34: water surface shapes [flat, slope variants] — non-blended (attached)
//	35-39: water surface shapes — blended with water colour
//	60+:   wave animation frames
//
// shape = kSlopeToWaterShape[slope & 0x0F] where table is:
//
//	{0,0,0,0, 0,0,0,2, 0,0,0,3, 0,1,4,0}
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintSurface.cpp paintSurfaceWater()
type WaterObject struct {
	Header      *ObjectHeader
	Name        string
	ImageOffset uint32 // base sprite index in G1 pool
}

// kSlopeToWaterShape maps tile slope (4-bit lower nibble) to water sprite shape index.
// OpenLoco reference: src/OpenLoco/src/Paint/PaintSurface.cpp kSlopeToWaterShape
var KSlopeToWaterShape = [16]uint8{
	0, 0, 0, 0,
	0, 0, 0, 2,
	0, 0, 0, 3,
	0, 1, 4, 0,
}

// FlatWaterSpriteOffset is the sprite offset for a flat water tile (shape=0, non-blended).
const FlatWaterSpriteOffset = 30

// BlendedWaterSpriteOffset is the sprite offset for blended water (shape=0, blended).
const BlendedWaterSpriteOffset = 35

// ParseWaterObject parses a water object from decompressed data.
func ParseWaterObject(header *ObjectHeader, data []byte) (*WaterObject, error) {
	return ParseWaterObjectWithG1(header, data, nil)
}

// ParseWaterObjectWithG1 parses a water object and optionally loads sprites into the G1 pool.
// OpenLoco reference: src/OpenLoco/src/Objects/WaterObject.cpp::load
func ParseWaterObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*WaterObject, error) {
	if len(data) < 0x0E {
		return nil, fmt.Errorf("water object data too small: %d bytes", len(data))
	}

	w := &WaterObject{Header: header}

	// Fixed struct: StringId(2) + costIndex(1) + var03(1) + costFactor(2) + image(4) + mapPixelImage(4) = 14 bytes
	// String table begins after the fixed struct at offset 0x0E.
	// OpenLoco: string table format [langID][null-terminated string]...[0xFF]
	offset := 0x0E

	for offset < len(data) && data[offset] != 0xFF {
		offset++ // skip language ID
		strStart := offset
		for offset < len(data) && data[offset] != 0 {
			offset++
		}
		if w.Name == "" {
			w.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF
	}

	// Load sprites into G1 pool
	if g1 != nil {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading water image table to G1: %w", err)
		}
		w.ImageOffset = imgRes.ImageOffset
	}

	return w, nil
}

// GetWaterSpriteIndex returns the G1 sprite index for a water surface tile.
// shape is from KSlopeToWaterShape; blended=true uses the colour-blended variant.
func (w *WaterObject) GetWaterSpriteIndex(shape uint8, blended bool) uint32 {
	offset := uint32(FlatWaterSpriteOffset) + uint32(shape)
	if blended {
		offset = uint32(BlendedWaterSpriteOffset) + uint32(shape)
	}
	return w.ImageOffset + offset
}
