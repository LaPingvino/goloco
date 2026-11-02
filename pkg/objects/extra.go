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
	// PaintStyle (0x04) picks the addition tables (track: Style0/1); Image is
	// load's image = the image table's base + 8 (0 without sprites loaded).
	PaintStyle uint8
	Image      uint32
}

// ParseExtraObject reads a track or road extra (Track/RoadExtraObject::load:
// the name string table, then the image table when a loader is given).
func ParseExtraObject(header *ObjectHeader, data []byte, g1 G1Loader) (*ExtraObject, error) {
	const fixedSize = 0x12
	if len(data) < fixedSize {
		return nil, fmt.Errorf("extra object data too small: %d bytes", len(data))
	}
	offset, name := walkStringTableLang(data, fixedSize)
	ex := &ExtraObject{Header: header, Name: name,
		Pieces:          uint16(data[0x02]) | uint16(data[0x03])<<8,
		BuildCostFactor: int16(uint16(data[0x06]) | uint16(data[0x07])<<8),
		CostIndex:       data[0x05], PaintStyle: data[0x04]}
	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("extra image table: %w", err)
		}
		ex.Image = imgRes.ImageOffset + 8
	}
	return ex, nil
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

// TrackExtraBySlot is the track extra in scenario slot i (trackObj->mods[]
// hold slots); nil if none.
func (m *ObjectManager) TrackExtraBySlot(i int) *ExtraObject {
	if i < 0 || i >= len(m.TrackExtraOrder) || m.TrackExtraOrder[i] == "" {
		return nil
	}
	return m.Extra(m.TrackExtraOrder[i])
}
