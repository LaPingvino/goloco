package objects

import (
	"fmt"
)

// TrackObject represents a railway track type (standard gauge, narrow gauge, etc.)
//
// OpenLoco reference: src/OpenLoco/src/Objects/TrackObject.h
// Header size: 0x36 (54 bytes)
type TrackObject struct {
	Header *ObjectHeader

	Name       string
	Image      uint32 // Base G1 sprite index (set after image table load)
	NumMods    uint8  // Number of track extra/modifications (max 4)
	NumSignals uint8  // Number of signal types
	NumBridges uint8  // Number of supported bridges (max 7)
	NumStations uint8 // Number of supported stations (max 7)
}

// GetObjectByIndex accessor sentinel — use ObjectManager.GetTrackObjectByIndex
func (t *TrackObject) GetHeader() *ObjectHeader { return t.Header }

// ParseTrackObject parses a track object from decompressed data.
func ParseTrackObject(header *ObjectHeader, data []byte) (*TrackObject, error) {
	return ParseTrackObjectWithG1(header, data, nil)
}

// ParseTrackObjectWithG1 parses a track object and loads sprites into the G1 pool.
//
// OpenLoco reference: src/OpenLoco/src/Objects/TrackObject.cpp::load
//
// Fixed struct (0x36 = 54 bytes):
//   0x00 StringId  name
//   0x02 uint16    trackPieces
//   0x04 uint16    stationTrackPieces
//   0x06 uint8     var_06
//   0x07 uint8     numCompatible
//   0x08 uint8     numMods
//   0x09 uint8     numSignals
//   0x0A uint8[4]  mods
//   0x0E uint16    signals
//   0x10 uint16    compatibleTracks
//   0x12 uint16    compatibleRoads
//   0x14 int16     buildCostFactor
//   0x16 int16     sellCostFactor
//   0x18 int16     tunnelCostFactor
//   0x1A uint8     costIndex
//   0x1B uint8     tunnel
//   0x1C uint16    curveSpeed
//   0x1E uint32    image  ← filled at load time
//   0x22 uint16    flags
//   0x24 uint8     numBridges
//   0x25 uint8[7]  bridges
//   0x2C uint8     numStations
//   0x2D uint8[7]  stations
//   0x34 uint8     displayOffset
//   0x35 uint8     pad
//
// After the fixed struct: string table, then numCompatible + numMods + numSignals + 1
// ObjectHeaders (16 bytes each) for compatible objects, then the image table.
func ParseTrackObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*TrackObject, error) {
	const fixedSize = 0x36
	if len(data) < fixedSize {
		return nil, fmt.Errorf("track object data too small: %d bytes (need %d)", len(data), fixedSize)
	}

	numCompatible := data[0x07]
	numMods := data[0x08]
	numSignals := data[0x09]
	numBridges := data[0x24]
	numStations := data[0x2C]

	track := &TrackObject{
		Header:      header,
		NumMods:     numMods,
		NumSignals:  numSignals,
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
		if track.Name == "" {
			track.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Skip embedded ObjectHeaders for:
	//   numCompatible compatible tracks/roads
	//   numMods track extra objects
	//   numSignals signal objects
	//   1 tunnel object
	//   numBridges bridge objects
	//   numStations station objects
	numHeaders := int(numCompatible) + int(numMods) + int(numSignals) + 1 + int(numBridges) + int(numStations)
	offset += numHeaders * HeaderSize

	if offset > len(data) {
		return nil, fmt.Errorf("track object: skipped past end of data (offset=%d, len=%d)", offset, len(data))
	}

	// Load sprites into G1 pool
	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading track image table to G1: %w", err)
		}
		track.Image = imgRes.ImageOffset
	}

	return track, nil
}

