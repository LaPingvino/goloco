package objects

import (
	"fmt"
)

// RoadObject represents a road type (country road, town road, etc.)
//
// OpenLoco reference: src/OpenLoco/src/Objects/RoadObject.h
// Header size: 0x30 (48 bytes)
type RoadObject struct {
	Header *ObjectHeader

	Name        string
	Image       uint32 // Base G1 sprite index (set after image table load)
	PaintStyle  uint8  // Sprite style: 0=normal, 1=ballast/sleeper, 2=left-curve
	NumMods     uint8  // Number of road extra/modifications (max 2)
	NumBridges  uint8  // Number of supported bridges (max 7)
	NumStations uint8  // Number of supported stations (max 7)
}

func (r *RoadObject) GetHeader() *ObjectHeader { return r.Header }

// ParseRoadObject parses a road object from decompressed data.
func ParseRoadObject(header *ObjectHeader, data []byte) (*RoadObject, error) {
	return ParseRoadObjectWithG1(header, data, nil)
}

// ParseRoadObjectWithG1 parses a road object and loads sprites into the G1 pool.
//
// OpenLoco reference: src/OpenLoco/src/Objects/RoadObject.cpp::load
//
// Fixed struct (0x30 = 48 bytes):
//   0x00 StringId  name
//   0x02 uint16    roadPieces
//   0x04 int16     buildCostFactor
//   0x06 int16     sellCostFactor
//   0x08 int16     tunnelCostFactor
//   0x0A uint8     costIndex
//   0x0B uint8     tunnel
//   0x0C uint16    maxSpeed
//   0x0E uint32    image  ← filled at load time
//   0x12 uint16    flags
//   0x14 uint8     numBridges
//   0x15 uint8[7]  bridges
//   0x1C uint8     numStations
//   0x1D uint8[7]  stations
//   0x24 uint8     paintStyle  ← 0=normal, 1=ballast style
//   0x25 uint8     numMods
//   0x26 uint8[2]  mods
//   0x28 uint8     numCompatible
//   0x29 uint8     displayOffset
//   0x2A uint16    compatibleRoads
//   0x2C uint16    compatibleTracks
//   0x2E uint8     targetTownSize
//   0x2F uint8     pad
//
// After the fixed struct: string table, then numCompatible + numMods + 1
// ObjectHeaders (16 bytes each) for compatible objects, then the image table.
func ParseRoadObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*RoadObject, error) {
	const fixedSize = 0x30
	if len(data) < fixedSize {
		return nil, fmt.Errorf("road object data too small: %d bytes (need %d)", len(data), fixedSize)
	}

	numBridges := data[0x14]
	numStations := data[0x1C]
	paintStyle := data[0x24]
	numMods := data[0x25]
	numCompatible := data[0x28]

	road := &RoadObject{
		Header:      header,
		PaintStyle:  paintStyle,
		NumMods:     numMods,
		NumBridges:  numBridges,
		NumStations: numStations,
	}

	offset := fixedSize

	// String table: (language_id + null-terminated string)*, ends with 0xFF
	for offset < len(data) && data[offset] != 0xFF {
		offset++ // skip language ID
		strStart := offset
		for offset < len(data) && data[offset] != 0 {
			offset++
		}
		if road.Name == "" {
			road.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Skip embedded ObjectHeaders for:
	//   numCompatible compatible roads/tracks
	//   numMods road extra objects
	//   1 tunnel object
	//   numBridges bridge objects
	//   numStations station objects
	numHeaders := int(numCompatible) + int(numMods) + 1 + int(numBridges) + int(numStations)
	offset += numHeaders * HeaderSize

	if offset > len(data) {
		return nil, fmt.Errorf("road object: skipped past end of data (offset=%d, len=%d)", offset, len(data))
	}

	// Load sprites into G1 pool
	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading road image table to G1: %w", err)
		}
		road.Image = imgRes.ImageOffset
	}

	return road, nil
}
