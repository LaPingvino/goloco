package objects

import "fmt"

// TrainStationObject represents a train station platform/canopy object.
//
// OpenLoco reference: src/OpenLoco/src/Objects/TrainStationObject.h
// sizeof(TrainStationObject) == 0xAE
//
// Sprite layout per sequenceIndex (Style0, 17 images each):
//
//	+0  straightBackNE              +4  straightBackSE
//	+1  straightFrontNE             +5  straightFrontSE
//	+2  straightCanopyNE            +6  straightCanopySE
//	+3  straightCanopyTranslucentNE +7  straightCanopyTranslucentSE
//	+8  diagonalNE0   +9  diagonalNE3  +10 diagonalNE1
//	+11 diagonalCanopyNE1           +12 diagonalCanopyTranslucentNE1
//	+13 diagonalSE1  +14 diagonalSE2  +15 diagonalSE3
//	+16 diagonalCanopyTranslucentSE3
type TrainStationObject struct {
	Header     *ObjectHeader
	Name       string
	PaintStyle uint8
	Flags      uint8 // TrainStationFlags: bit0=recolourable, bit1=noCanopy

	// ImageOffset is the base G1 pool index set by LoadImageTable.
	// Preview images:  ImageOffset+0 (base), ImageOffset+1 (windows overlay).
	// Sequence images: SequenceImageOffsets[seqIdx] + Style0::{sprite constant}
	ImageOffset uint32

	// SequenceImageOffsets[i] = ImageOffset + 2 + i*numImagesPerStyle
	// For Style0: numImagesPerStyle = 17
	// For Style1: numImagesPerStyle = 8
	SequenceImageOffsets [4]uint32
}

// Style0 sprite offsets (relative to SequenceImageOffsets[seqIdx]).
// OpenLoco reference: TrainStationObject.h TrainStation::ImageIds::Style0
const (
	StStraightBackNE              = 0
	StStraightFrontNE             = 1
	StStraightCanopyNE            = 2
	StStraightCanopyTranslucentNE = 3
	StStraightBackSE              = 4
	StStraightFrontSE             = 5
	StStraightCanopySE            = 6
	StStraightCanopyTranslucentSE = 7
	StStyle0NumImages             = 17
	StStyle1NumImages             = 8
)

// ParseTrainStationObject parses without loading sprites.
func ParseTrainStationObject(header *ObjectHeader, data []byte) (*TrainStationObject, error) {
	return ParseTrainStationObjectWithG1(header, data, nil)
}

// ParseTrainStationObjectWithG1 parses a TrainStationObject from decompressed DAT data.
//
// Layout after sizeof(TrainStationObject)=0xAE bytes:
//   string table: [langID][str\0]...[0xFF]
//   numCompatible ObjectHeaders (8 bytes each, read from offset 0x22)
//   4×4 cargo offset blobs  (variable length, each: z + N×(x,y,x,y) + terminator)
//   16   var_6E blobs        (same format as cargo offset blobs)
//   image table
//
// OpenLoco reference: TrainStationObject.cpp::load()
func ParseTrainStationObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*TrainStationObject, error) {
	const fixedSize = 0xAE
	if len(data) < fixedSize {
		return nil, fmt.Errorf("train station object data too small: %d bytes (need %d)", len(data), fixedSize)
	}

	obj := &TrainStationObject{
		Header:     header,
		PaintStyle: data[0x02],
		Flags:      data[0x0C], // TrainStationFlags (uint8 — only lower byte used)
	}
	numCompatible := data[0x22]

	offset := fixedSize

	// String table
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
		return nil, fmt.Errorf("train station object: compatible headers exceed data length")
	}

	// Skip 4×4 = 16 cargo offset blobs, then 16 var_6E blobs (same format).
	// Each blob: [z:int8] ([x:int8 y:int8 x:int8 y:int8]* terminator:[0xFF ...])
	// terminator: first byte is 0xFF (−1 as int8); skip 4 bytes for it.
	for blobSet := 0; blobSet < 2; blobSet++ { // set 0 = cargoOffsetBytes, set 1 = var_6E
		blobCount := 16
		for i := 0; i < blobCount; i++ {
			if offset >= len(data) {
				return nil, fmt.Errorf("train station object: blob %d/%d out of data", blobSet, i)
			}
			offset++ // z byte
			for offset+4 <= len(data) {
				if data[offset] == 0xFF { // int8(-1) terminator
					offset += 4
					break
				}
				offset += 4 // x, y, x, y
			}
		}
	}

	if offset > len(data) {
		return nil, fmt.Errorf("train station object: image table offset %d exceeds data len %d", offset, len(data))
	}

	// Load sprites into G1 pool
	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("loading train station image table to G1: %w", err)
		}
		obj.ImageOffset = imgRes.ImageOffset
	}

	// Compute per-sequence image offsets:
	//   imageOffsets[i] = image + totalPreviewImages + i * numImagesForStyle
	numPerSeq := uint32(StStyle0NumImages)
	if obj.PaintStyle == 1 {
		numPerSeq = StStyle1NumImages
	}
	for i := range obj.SequenceImageOffsets {
		obj.SequenceImageOffsets[i] = obj.ImageOffset + 2 + uint32(i)*numPerSeq
	}

	return obj, nil
}
