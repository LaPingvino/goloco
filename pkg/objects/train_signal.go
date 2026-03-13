package objects

import "fmt"

// TrainSignalObject represents a railway signal object.
//
// OpenLoco reference: src/OpenLoco/src/Objects/TrainSignalObject.h
// sizeof(TrainSignalObject) == 0x1E
//
// Sprite layout (per direction, 8 directions total):
//
//	Offset 0: direction 0 (NE straight)
//	Offset 1: direction 1 (SE straight)
//	Offset 2: direction 2 (SW straight)
//	Offset 3: direction 3 (NW straight)
//	Offset 4-7: second rotation set (for diagonal tracks)
//
//	Per frame: 8 sprites (one per image rotation offset 0-7)
//	imageOffset = imageRotationOffset + ImageOffset + (frame << 3)
//
//	Light overlays (if hasLights):
//	  + 80: redLights
//	  + 88: redLights2
//	  + 96: greenLights
//	  + 104: greenLights2
type TrainSignalObject struct {
	Header      *ObjectHeader
	Name        string
	Flags       uint16 // TrainSignalObjectFlags
	AnimSpeed   uint8
	NumFrames   uint8
	IsLeft      bool // flags bit 0: left vs right orientation
	HasLights   bool // flags bit 1

	// ImageOffset is the base G1 pool index set by LoadImageTable.
	// All sprite lookups: ImageOffset + local_index
	ImageOffset uint32
}

// Signal-specific image offset constants (relative to ImageOffset).
// OpenLoco reference: TrainSignalObject.h TrainSignal::ImageIds
const (
	TrainSignalRedLights   = 80
	TrainSignalRedLights2  = 88
	TrainSignalGreenLights = 96
	TrainSignalGreenLights2 = 104
)

// Signal position offset tables indexed by trackRotation (0-15).
// [0] = left-side signals (_4FE830); [1] = right-side signals (_4FE870).
// Each entry is {tileX, tileY} within the 32×32 tile space.
//
// OpenLoco reference: src/OpenLoco/src/Paint/PaintSignal.cpp
var SignalPosLeft = [16][2]int8{
	{24, 4}, {4, 8}, {8, 28}, {28, 24},
	{24, 1}, {1, 8}, {8, 31}, {31, 24},
	{24, 16}, {16, 8}, {8, 16}, {16, 24},
	{6, 10}, {10, 26}, {26, 22}, {22, 6},
}

var SignalPosRight = [16][2]int8{
	{24, 28}, {28, 8}, {8, 4}, {4, 24},
	{24, 16}, {16, 8}, {8, 16}, {16, 24},
	{24, 31}, {31, 8}, {8, 1}, {31, 24},
	{22, 26}, {26, 10}, {10, 6}, {6, 22},
}

// ParseTrainSignalObject parses without loading sprites.
func ParseTrainSignalObject(header *ObjectHeader, data []byte) (*TrainSignalObject, error) {
	return ParseTrainSignalObjectWithG1(header, data, nil)
}

// ParseTrainSignalObjectWithG1 parses a TrainSignalObject from decompressed DAT data.
//
// Layout:
//
//	0x00–0x1D: fixed struct (30 bytes)
//	0x1E:      string table [langID][str\0]...[0xFF]
//	after strings: numCompatible ObjectHeaders (16 bytes each, from offset 0x12)
//	image table
//
// OpenLoco reference: src/OpenLoco/src/Objects/TrainSignalObject.cpp load()
func ParseTrainSignalObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*TrainSignalObject, error) {
	const fixedSize = 0x1E
	if len(data) < fixedSize {
		return nil, fmt.Errorf("train signal object data too small: %d bytes (need %d)", len(data), fixedSize)
	}

	flags := uint16(data[0x02]) | uint16(data[0x03])<<8
	obj := &TrainSignalObject{
		Header:    header,
		Flags:     flags,
		AnimSpeed: data[0x04],
		NumFrames: data[0x05],
		IsLeft:    flags&0x01 != 0,
		HasLights: flags&0x02 != 0,
	}
	numCompatible := data[0x12]

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

	// Skip numCompatible ObjectHeaders (HeaderSize bytes each)
	offset += int(numCompatible) * HeaderSize

	if offset > len(data) {
		return nil, fmt.Errorf("train signal object: offset %d > data len %d after headers", offset, len(data))
	}

	// Load image table
	if g1 != nil {
		result, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("train signal object: loading image table: %w", err)
		}
		obj.ImageOffset = result.ImageOffset
	}

	return obj, nil
}
