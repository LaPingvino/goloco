package objects

import "fmt"

// BridgeObject represents a bridge structure object.
//
// OpenLoco reference: src/OpenLoco/src/Objects/BridgeObject.h
// sizeof(BridgeObject) == 0x2C
//
// Sprite layout (fixed 192 sprites, unused slots are 1×1 pixels):
//
//	 0: uiDropdown
//	 1: deckBaseNoSupport
//	 2-5: deckBaseNoSupportEdge0-3
//	 6: roofFullTile
//	 7-10: roofEdge0-3
//	 11: deckBaseWithSupportHeaderEdge3
//	 12-15: deckWallEdge3/1/0/2
//	 16-35: support segments (NE/SW sides)
//	 36-59: NE span deck+walls+roof (4 span positions × 6 images)
//	 60-83: SW span deck+walls+roof (4 span positions × 6 images)
//	 84-107: gentle slope deck images
//	 108-191: pillar segments
type BridgeObject struct {
	Header          *ObjectHeader
	Name            string
	Flags           uint8
	ClearHeight     uint16
	DeckDepth       int16 // 16 or 32
	SpanLength      uint8 // 1, 2, or 4
	PillarSpacing   uint8
	MaxSpeed        uint16
	MaxHeight       uint8
	CostIndex       uint8
	DesignedYear    uint16

	// ImageOffset is the base G1 pool index set by LoadImageTable.
	ImageOffset uint32
}

// Bridge deck sprite offsets relative to ImageOffset.
// OpenLoco reference: BridgeObject.h Bridge::ImageIds
const (
	BridgeDeckNoSupport      = 1
	BridgeDeckNoSupportEdge0 = 2
	BridgeDeckNoSupportEdge1 = 3
	BridgeDeckNoSupportEdge2 = 4
	BridgeDeckNoSupportEdge3 = 5
	BridgeDeckNEspan0        = 36 // flat NE-facing span, position 0
	BridgeDeckNEspan1        = 42 // flat NE-facing span, position 1
	BridgeDeckSWspan0        = 60 // flat SW-facing span, position 0
	BridgeDeckSWspan1        = 66 // flat SW-facing span, position 1
)

// ParseBridgeObject parses without loading sprites.
func ParseBridgeObject(header *ObjectHeader, data []byte) (*BridgeObject, error) {
	return ParseBridgeObjectWithG1(header, data, nil)
}

// ParseBridgeObjectWithG1 parses a BridgeObject from decompressed DAT data.
//
// Layout:
//
//	0x00–0x2B: fixed struct (44 bytes)
//	0x2C: string table [langID][str\0]...[0xFF]
//	after strings: trackNumCompatible ObjectHeaders (from offset 0x1A)
//	after those:  roadNumCompatible ObjectHeaders (from offset 0x22)
//	image table
//
// OpenLoco reference: src/OpenLoco/src/Objects/BridgeObject.cpp load()
func ParseBridgeObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*BridgeObject, error) {
	const fixedSize = 0x2C
	if len(data) < fixedSize {
		return nil, fmt.Errorf("bridge object data too small: %d bytes (need %d)", len(data), fixedSize)
	}

	obj := &BridgeObject{
		Header:        header,
		Flags:         data[0x02],
		ClearHeight:   uint16(data[0x04]) | uint16(data[0x05])<<8,
		DeckDepth:     int16(data[0x06]) | int16(data[0x07])<<8,
		SpanLength:    data[0x08],
		PillarSpacing: data[0x09],
		MaxSpeed:      uint16(data[0x0A]) | uint16(data[0x0B])<<8,
		MaxHeight:     data[0x0C],
		CostIndex:     data[0x0D],
		DesignedYear:  uint16(data[0x2A]) | uint16(data[0x2B])<<8,
	}
	trackNumCompatible := data[0x1A]
	roadNumCompatible := data[0x22]

	offset := fixedSize

	// Walk string table: [langID][str\0]...[0xFF]
	for offset < len(data) && data[offset] != 0xFF {
		offset++ // skip langID
		for offset < len(data) && data[offset] != 0 {
			offset++ // skip string bytes
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Skip trackNumCompatible + roadNumCompatible ObjectHeaders (HeaderSize bytes each)
	offset += (int(trackNumCompatible) + int(roadNumCompatible)) * HeaderSize

	if offset > len(data) {
		return nil, fmt.Errorf("bridge object: offset %d > data len %d after headers", offset, len(data))
	}

	// Load image table
	if g1 != nil {
		result, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("bridge object: loading image table: %w", err)
		}
		obj.ImageOffset = result.ImageOffset
	}

	return obj, nil
}
