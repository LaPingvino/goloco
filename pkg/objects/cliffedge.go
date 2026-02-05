package objects

import (
	"fmt"
)

// CliffEdgeObject represents cliff edge sprites for height transitions
// OpenLoco reference: src/OpenLoco/src/Objects/CliffEdgeObject.h
type CliffEdgeObject struct {
	Header *ObjectHeader

	Name  string
	Image uint32 // Base sprite index in G1 pool
}

// ParseCliffEdgeObject parses a cliff edge object from decompressed data
func ParseCliffEdgeObject(header *ObjectHeader, data []byte) (*CliffEdgeObject, error) {
	return ParseCliffEdgeObjectWithG1(header, data, nil)
}

// ParseCliffEdgeObjectWithG1 parses a cliff edge object and optionally loads sprites into the G1 pool
// OpenLoco reference: src/OpenLoco/src/Objects/CliffEdgeObject.cpp::load
func ParseCliffEdgeObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*CliffEdgeObject, error) {
	if len(data) < 0x06 {
		return nil, fmt.Errorf("cliff edge object data too small: %d bytes", len(data))
	}

	cliff := &CliffEdgeObject{
		Header: header,
	}

	// Skip StringId name field (2 bytes) at offset 0x00
	// Image field at offset 0x02 will be populated after loading

	// String table starts at offset 0x06
	offset := 0x06

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
		if cliff.Name == "" {
			cliff.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Load sprites into G1 pool if G1 loader is available
	if g1 != nil {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading image table to G1: %w", err)
		}

		// Store the base sprite index
		cliff.Image = imgRes.ImageOffset
	} else {
		// Fallback: no dynamic loading
		cliff.Image = 0
	}

	return cliff, nil
}
