package objects

import (
	"fmt"
	"strings"
)

// ExtraObject is a TrackExtraObject or RoadExtraObject (overhead wires,
// third rail, rack rail…): both are 0x12 bytes fixed (sizeof from the
// compiler: 18), then the name's string table.
// PARTIAL: only the name is read (the image table and the cost factors are
// not; the overhead tab's preview and mod costs need them).
type ExtraObject struct {
	Header *ObjectHeader
	Name   string
}

// ParseExtraObject reads a track or road extra's name.
func ParseExtraObject(header *ObjectHeader, data []byte) (*ExtraObject, error) {
	const fixedSize = 0x12
	if len(data) < fixedSize {
		return nil, fmt.Errorf("extra object data too small: %d bytes", len(data))
	}
	_, name := walkStringTableLang(data, fixedSize)
	return &ExtraObject{Header: header, Name: name}, nil
}

// ExtraName is the display name of the track/road extra with this DAT name
// ("" if not loaded).
func (m *ObjectManager) ExtraName(datName string) string {
	return m.extraNames[upperName(datName)]
}

func upperName(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }
