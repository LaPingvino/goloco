package objects

import (
	"encoding/binary"
	"fmt"
)

// LandObjectFlags define special behavior for land types
type LandObjectFlags uint8

const (
	LandFlagNone                        LandObjectFlags = 0
	LandFlagHasGrowthStages             LandObjectFlags = 1 << 0
	LandFlagHasReplacementLandHeader    LandObjectFlags = 1 << 1
	LandFlagIsDesert                    LandObjectFlags = 1 << 2
	LandFlagNoTrees                     LandObjectFlags = 1 << 3
	LandFlagHasSharpSlopeTransition     LandObjectFlags = 1 << 4
	LandFlagDisableSmoothTileTransition LandObjectFlags = 1 << 5
)

// Terrain image constants from OpenLoco
const (
	NumTerrainImages                = 19
	NumTerrainBlendImages           = 6
	NumTerrainImageZoomLevels       = 4
	TerrainFlatImageOffset          = NumTerrainImages * (NumTerrainImageZoomLevels - 1)                 // 57
	NumImagesPerGrowthStage         = NumTerrainImages + NumTerrainBlendImages                           // 25
	NumImagesPerGrowthStagePlusZoom = NumTerrainImages*NumTerrainImageZoomLevels + NumTerrainBlendImages // 82
)

// LandObject represents a land/terrain type (grass, desert, etc.)
type LandObject struct {
	Header *ObjectHeader

	// From file (offsets relative to object data start)
	Name                     string
	CostIndex                uint8
	NumGrowthStages          uint8
	NumImageAngles           uint8
	Flags                    LandObjectFlags
	CliffEdgeHeaderIdx       uint8 // Index into loaded cliff edge objects
	ReplacementLandHeaderIdx uint8 // Index into loaded land objects
	CostFactor               int16
	DistributionPattern      uint8
	NumVariations            uint8
	VariationLikelihood      uint8

	// Calculated after loading
	ImageOffset             uint32 // Base sprite index in global pool
	NumImagesPerGrowthStage uint32
	CliffEdgeImage          uint32
	MapPixelImage           uint32

	// Embedded sprite data
	Sprites []*SpriteElement
}

// SpriteElement represents a single sprite from an object's image table
type SpriteElement struct {
	Width      int16
	Height     int16
	XOffset    int16
	YOffset    int16
	Flags      uint16
	ZoomOffset int16
	Data       []byte
}

// G1ElementFlagHasZoomSprites indicates sprite has zoom variants
const G1ElementFlagHasZoomSprites = 0x10

// ParseLandObject parses a land object from decompressed data
func ParseLandObject(header *ObjectHeader, data []byte) (*LandObject, error) {
	if len(data) < 0x1E {
		return nil, fmt.Errorf("land object data too small: %d bytes", len(data))
	}

	land := &LandObject{
		Header:                   header,
		CostIndex:                data[0x02],
		NumGrowthStages:          data[0x03],
		NumImageAngles:           data[0x04],
		Flags:                    LandObjectFlags(data[0x05]),
		CliffEdgeHeaderIdx:       data[0x06],
		ReplacementLandHeaderIdx: data[0x07],
		CostFactor:               int16(binary.LittleEndian.Uint16(data[0x08:0x0A])),
		DistributionPattern:      data[0x1A],
		NumVariations:            data[0x1B],
		VariationLikelihood:      data[0x1C],
	}

	// Calculate derived values
	land.NumImagesPerGrowthStage = uint32(land.NumImageAngles) * NumImagesPerGrowthStage

	// Skip past the fixed structure (0x1E bytes)
	offset := 0x1E

	// Parse string table: (language_id + null-terminated string)*, ends with 0xFF
	for offset < len(data) && data[offset] != 0xFF {
		// Skip language ID
		offset++
		// Find and skip the null-terminated string
		strStart := offset
		for offset < len(data) && data[offset] != 0 {
			offset++
		}
		// Save the first string as the name
		if land.Name == "" {
			land.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Skip ObjectHeader for CliffEdge (16 bytes)
	if offset+16 > len(data) {
		return nil, fmt.Errorf("missing CliffEdge header at offset %d", offset)
	}
	offset += 16

	// Skip ObjectHeader for ReplacementLand if present
	if land.Flags&LandFlagHasReplacementLandHeader != 0 {
		if offset+16 > len(data) {
			return nil, fmt.Errorf("missing ReplacementLand header at offset %d", offset)
		}
		offset += 16
	}

	// Parse embedded G1 image table
	sprites, err := parseImageTable(data[offset:])
	if err != nil {
		return nil, fmt.Errorf("parsing image table: %w", err)
	}
	land.Sprites = sprites

	return land, nil
}

// parseImageTable parses a G1 image table embedded in object data
func parseImageTable(data []byte) ([]*SpriteElement, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("image table data too small: %d bytes", len(data))
	}

	numEntries := int(binary.LittleEndian.Uint32(data[0:4]))
	totalSize := int(binary.LittleEndian.Uint32(data[4:8]))

	if numEntries <= 0 || numEntries > 10000 {
		return nil, fmt.Errorf("invalid entry count: %d", numEntries)
	}

	const elementHeaderSize = 16
	headersOffset := 8
	if headersOffset+numEntries*elementHeaderSize > len(data) {
		return nil, fmt.Errorf("headers truncated: need %d, have %d",
			headersOffset+numEntries*elementHeaderSize, len(data))
	}

	elementDataOffset := headersOffset + numEntries*elementHeaderSize
	if elementDataOffset > len(data) {
		return nil, fmt.Errorf("element data offset beyond data")
	}
	elementData := data[elementDataOffset:]

	// Clamp totalSize if it exceeds available data
	if totalSize > len(elementData) {
		totalSize = len(elementData)
	}

	sprites := make([]*SpriteElement, numEntries)

	for i := 0; i < numEntries; i++ {
		hOff := headersOffset + i*elementHeaderSize

		spriteOffset := int(binary.LittleEndian.Uint32(data[hOff : hOff+4]))
		width := int16(binary.LittleEndian.Uint16(data[hOff+4 : hOff+6]))
		height := int16(binary.LittleEndian.Uint16(data[hOff+6 : hOff+8]))
		xOffset := int16(binary.LittleEndian.Uint16(data[hOff+8 : hOff+10]))
		yOffset := int16(binary.LittleEndian.Uint16(data[hOff+10 : hOff+12]))
		flags := binary.LittleEndian.Uint16(data[hOff+12 : hOff+14])
		zoomOffset := int16(binary.LittleEndian.Uint16(data[hOff+14 : hOff+16]))

		// Calculate end offset
		var endOffset int
		if i+1 < numEntries {
			nextHOff := headersOffset + (i+1)*elementHeaderSize
			endOffset = int(binary.LittleEndian.Uint32(data[nextHOff : nextHOff+4]))
		} else {
			endOffset = totalSize
		}

		var spriteData []byte
		if spriteOffset >= 0 && endOffset > spriteOffset && endOffset <= len(elementData) {
			spriteData = elementData[spriteOffset:endOffset]
		}

		sprites[i] = &SpriteElement{
			Width:      width,
			Height:     height,
			XOffset:    xOffset,
			YOffset:    yOffset,
			Flags:      flags,
			ZoomOffset: zoomOffset,
			Data:       spriteData,
		}
	}

	return sprites, nil
}

// GetFlatTerrainSpriteIndex returns the local sprite index for the zoom-0
// flat terrain tile within this object's image table.
//
// OpenLoco reference: src/OpenLoco/src/Objects/LandObject.cpp
//   image = numImageAngles * numGrowthStages * kTerrainFlatImageOffset + imgRes.imageOffset
//
// imgRes.imageOffset is the global sprite-pool base; for local indexing it is 0.
func (l *LandObject) GetFlatTerrainSpriteIndex() int {
	return int(l.NumImageAngles) * int(l.NumGrowthStages) * TerrainFlatImageOffset
}

// HasFlags checks if the land object has the given flags
func (l *LandObject) HasFlags(flags LandObjectFlags) bool {
	return l.Flags&flags == flags
}
