package objects

import (
	"encoding/binary"
	"fmt"
)

// CargoObject is a cargo type definition (passengers, coal, mail, …).
// OpenLoco reference: include/OpenLoco/Objects/CargoObject.h (0x1F bytes) and
// src/Objects/CargoObject.cpp load(): four string tables (name,
// unitsAndCargoName, unitNameSingular, unitNamePlural) then the image table.
type CargoObject struct {
	Header *ObjectHeader

	UnitWeight            uint16 // 0x02
	CargoTransferTime     uint16 // 0x04
	CargoCategory         uint16 // 0x10
	Flags                 uint8  // 0x12
	NumPlatformVariations uint8  // 0x13: platform images after unitInlineSprite + kStationPlatformBegin
	StationCargoDensity   uint8  // 0x14: log2 of units per visible cargo pile (updateCargoDistribution)
	PremiumDays           uint8  // 0x15
	MaxNonPremiumDays     uint8  // 0x16
	NonPremiumRate        uint16 // 0x17
	PenaltyRate           uint16 // 0x19
	PaymentFactor         uint16 // 0x1B — calculateDeliveredCargoPayment's basis
	PaymentIndex          uint8  // 0x1D
	UnitSize              uint8  // 0x1E

	// Localized display strings picked from the DAT string tables using
	// SelectedLocoLanguage (falling back to English).
	Name              string // "Passengers"
	UnitsAndCargoName string // "{COMMA16} passengers"
	UnitNameSingular  string
	UnitNamePlural    string

	ImageOffset uint32
}

// SelectedLocoLanguage is the vanilla language id used when picking strings
// from object string tables. 0 = English (UK), the base game's default.
// OpenLoco reference: Localisation::LocoLanguageId / loadStringTable.
var SelectedLocoLanguage uint8 = 0

// walkStringTableLang walks one string table starting at offset: the offset
// past its 0xFF terminator and the string loadStringTable chooses
// (chooseString), as stored.
func walkStringTableLang(data []byte, offset int) (int, string) {
	if offset >= len(data) {
		return offset, ""
	}
	n, raw := chooseString(data[offset:])
	return offset + n, string(raw)
}

// ParseCargoObjectWithG1 parses a cargo object, loading its image table into
// the G1 pool when a loader is provided.
func ParseCargoObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*CargoObject, error) {
	const fixedSize = 0x1F
	if len(data) < fixedSize {
		return nil, fmt.Errorf("cargo object too small: %d bytes", len(data))
	}
	c := &CargoObject{
		Header:                header,
		UnitWeight:            binary.LittleEndian.Uint16(data[0x02:]),
		CargoTransferTime:     binary.LittleEndian.Uint16(data[0x04:]),
		CargoCategory:         binary.LittleEndian.Uint16(data[0x10:]),
		Flags:                 data[0x12],
		NumPlatformVariations: data[0x13],
		StationCargoDensity:   data[0x14],
		PremiumDays:           data[0x15],
		MaxNonPremiumDays:     data[0x16],
		NonPremiumRate:        binary.LittleEndian.Uint16(data[0x17:]),
		PenaltyRate:           binary.LittleEndian.Uint16(data[0x19:]),
		PaymentFactor:         binary.LittleEndian.Uint16(data[0x1B:]),
		PaymentIndex:          data[0x1D],
		UnitSize:              data[0x1E],
	}

	offset := fixedSize
	offset, c.Name = walkStringTableLang(data, offset)
	offset, c.UnitsAndCargoName = walkStringTableLang(data, offset)
	offset, c.UnitNameSingular = walkStringTableLang(data, offset)
	offset, c.UnitNamePlural = walkStringTableLang(data, offset)

	if g1 != nil && offset < len(data) {
		imgRes, err := g1.LoadImageTable(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("cargo image table: %w", err)
		}
		c.ImageOffset = imgRes.ImageOffset
	}
	return c, nil
}
