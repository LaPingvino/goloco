package objects

import "fmt"

// RoadStationObject represents a road station (bus stop) platform/canopy object.
//
// OpenLoco reference: src/OpenLoco/src/Objects/RoadStationObject.h
// sizeof(RoadStationObject) == 0x6E
//
// Sprite layout (Style0, 8 images per platform section):
//
//	+0  straightBackNE    +4  straightBackSE
//	+1  straightFrontNE   +5  straightFrontSE
//	+2  straightCanopyNE  +6  straightCanopySE
//	+3  straightCanopyTranslucentNE  +7  straightCanopyTranslucentSE
type RoadStationObject struct {
	Header     *ObjectHeader
	Name       string
	PaintStyle uint8

	// ImageOffset is the base G1 pool index set by LoadImageTable.
	// Preview images:  ImageOffset+0 (base), ImageOffset+1 (windows overlay).
	// Sequence images: SequenceImageOffsets[seqIdx] + Style0 sprite constant
	ImageOffset uint32

	// SequenceImageOffsets[i] = ImageOffset + 2 + i * 8 (Style0 always has 8 images)
	// OpenLoco reference: RoadStationObject.cpp::load()
	SequenceImageOffsets [4]uint32
}

// Style0 sprite offsets (relative to SequenceImageOffsets[seqIdx]).
// OpenLoco reference: RoadStationObject.h RoadStation::ImageIds::Style0
const (
	RsStraightBackNE              = 0
	RsStraightFrontNE             = 1
	RsStraightCanopyNE            = 2
	RsStraightCanopyTranslucentNE = 3
	RsStraightBackSE              = 4
	RsStraightFrontSE             = 5
	RsStraightCanopySE            = 6
	RsStraightCanopyTranslucentSE = 7
	RsStyle0NumImages             = 8
	RsTotalPreviewImages          = 2
)

// ParseRoadStationObject parses without loading sprites.
func ParseRoadStationObject(header *ObjectHeader, data []byte) (*RoadStationObject, error) {
	return ParseRoadStationObjectWithG1(header, data, nil)
}

// ParseRoadStationObjectWithG1 parses a RoadStationObject from decompressed DAT data.
//
// Layout after sizeof(RoadStationObject)=0x6E bytes:
//
//	string table: [langID][str\0]...[0xFF]
//	numCompatible ObjectHeaders (8 bytes each, read from offset 0x20)
//	optional 1 cargo ObjectHeader (if passenger|freight flags set — bit1 or bit2 of flags)
//	4×4 = 16 cargo offset blobs (each: [z:int8] [x,y,x,y]* [0xFF 0x00 0x00 0x00])
//	image table
//
// OpenLoco reference: src/OpenLoco/src/Objects/RoadStationObject.cpp::load()
func ParseRoadStationObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*RoadStationObject, error) {
	const fixedSize = 0x6E
	if len(data) < fixedSize {
		return nil, fmt.Errorf("road station object data too small: %d bytes (need %d)", len(data), fixedSize)
	}

	obj := &RoadStationObject{
		Header:     header,
		PaintStyle: data[0x02],
	}
	numCompatible := data[0x20]
	flags := data[0x0B] // RoadStationFlags

	offset := fixedSize

	// String table: [langID][str\0]...[0xFF]
	for offset < len(data) && data[offset] != 0xFF {
		offset++ // skip language ID
		strStart := offset
		for offset < len(data) && data[offset] != 0 {
			offset++
		}
		if obj.Name == "" {
			obj.Name = string(data[strStart:offset])
		}
		offset++ // skip null terminator
	}
	if offset < len(data) {
		offset++ // skip 0xFF end marker
	}

	// Skip numCompatible ObjectHeaders (HeaderSize bytes each)
	offset += int(numCompatible) * HeaderSize
	if offset > len(data) {
		return nil, fmt.Errorf("road station object: compatible headers exceed data length")
	}

	// If passenger or freight flags set (bits 1 or 2), skip 1 cargo ObjectHeader.
	// RoadStationFlags: passenger=bit1, freight=bit2
	if flags&0x06 != 0 {
		offset += HeaderSize
		if offset > len(data) {
			return nil, fmt.Errorf("road station object: cargo header exceeds data length")
		}
	}

	// Skip 4×4 = 16 cargo offset blobs.
	// Each blob: [z:int8] ([x:int8 y:int8 x:int8 y:int8]* terminator:[0xFF + 3 bytes])
	// The terminator is 4 bytes: first byte 0xFF (== int8(-1)), then 3 pad bytes.
	for i := range 16 {
		if offset >= len(data) {
			return nil, fmt.Errorf("road station object: cargo blob %d out of data", i)
		}
		offset++ // z byte
		for offset+4 <= len(data) {
			if data[offset] == 0xFF {
				offset += 4 // consume the 4-byte terminator
				break
			}
			offset += 4 // x, y, x, y
		}
	}

	if offset > len(data) {
		return nil, fmt.Errorf("road station object: image table offset %d exceeds data len %d", offset, len(data))
	}

	// Load sprites into G1 pool
	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading road station image table to G1: %w", err)
		}
		obj.ImageOffset = imgRes.ImageOffset
	}

	// Compute per-sequence image offsets:
	//   imageOffsets[i] = ImageOffset + totalPreviewImages + i * RsStyle0NumImages
	// OpenLoco reference: RoadStationObject.cpp::load() — imageOffset += kDrawStyleTotalNumImages[paintStyle]
	for i := range obj.SequenceImageOffsets {
		obj.SequenceImageOffsets[i] = obj.ImageOffset + RsTotalPreviewImages + uint32(i)*RsStyle0NumImages
	}

	return obj, nil
}
