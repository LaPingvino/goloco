package objects

import (
	"fmt"
	"strings"
)

// ExtraObject is a TrackExtraObject or RoadExtraObject (overhead wires,
// third rail, rack rail…): both are 0x12 bytes fixed (sizeof from the
// compiler: 18), then the name's string table.
// PARTIAL: the image table isn't read (the overhead tab's preview needs it).
type ExtraObject struct {
	Header *ObjectHeader
	Name   string
	// Pieces is trackPieces/roadPieces (0x02): the piece kinds it can go on;
	// BuildCostFactor (0x06) and CostIndex (0x05) price it (offsets from
	// the compiler).
	Pieces          uint16
	BuildCostFactor int16
	CostIndex       uint8
}

// ParseExtraObject reads a track or road extra's name.
func ParseExtraObject(header *ObjectHeader, data []byte) (*ExtraObject, error) {
	const fixedSize = 0x12
	if len(data) < fixedSize {
		return nil, fmt.Errorf("extra object data too small: %d bytes", len(data))
	}
	_, name := walkStringTableLang(data, fixedSize)
	return &ExtraObject{Header: header, Name: name,
		Pieces:          uint16(data[0x02]) | uint16(data[0x03])<<8,
		BuildCostFactor: int16(uint16(data[0x06]) | uint16(data[0x07])<<8),
		CostIndex:       data[0x05]}, nil
}

// Extra returns the track/road extra with this DAT name (nil if not loaded).
func (m *ObjectManager) Extra(datName string) *ExtraObject {
	return m.extras[upperName(datName)]
}

// ExtraName is the display name of the track/road extra with this DAT name
// ("" if not loaded).
func (m *ObjectManager) ExtraName(datName string) string {
	return m.extraNames[upperName(datName)]
}

func upperName(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }
