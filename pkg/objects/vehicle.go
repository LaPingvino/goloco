package objects

import (
	"encoding/binary"
	"fmt"
	"io"
)

// VehicleObject represents a vehicle from a DAT file
type VehicleObject struct {
	Header ObjectHeader

	// Basic properties (from the decompressed data)
	Name                uint16 // String ID
	Mode                TransportMode
	Type                VehicleType
	NumCarComponents    uint8
	TrackType           uint8
	NumTrackExtras      uint8
	CostIndex           uint8
	CostFactor          int16
	Reliability         uint8
	RunCostIndex        uint8
	RunCostFactor       int16
	ColourType          uint8
	NumCompatVehicles   uint8
	CompatibleVehicles  [8]uint16
	RequiredTrackExtras [4]uint8

	// Car components
	CarComponents [4]VehicleCarComponent

	// Sprites
	BodySprites  [4]VehicleBodySprite
	BogieSprites [2]VehicleBogieSprite

	// Performance
	Power     uint16
	Speed     uint16 // Speed16
	RackSpeed uint16
	Weight    uint16
	Flags     uint16
	MaxCargo  [2]uint8

	// Design dates
	Designed uint16
	Obsolete uint16

	// Parsed strings
	DisplayName string

	// Image IDs (base image ID for this object's sprites)
	BaseImageID uint32
}

// VehicleCarComponent describes one car in a vehicle consist
type VehicleCarComponent struct {
	FrontBogiePosition   uint8
	BackBogiePosition    uint8
	FrontBogieSpriteIdx  uint8
	BackBogieSpriteIdx   uint8
	BodySpriteIdx        uint8
	EmitterHorizontalPos uint8
}

// VehicleBodySprite describes body sprite layout
type VehicleBodySprite struct {
	NumFlatRotationFrames   uint8
	NumSlopedRotationFrames uint8
	NumAnimationFrames      uint8
	NumCargoLoadFrames      uint8
	NumCargoFrames          uint8
	NumRollFrames           uint8
	HalfLength              uint8
	Flags                   uint8
	Width                   uint8
	HeightNegative          uint8
	HeightPositive          uint8
	FlatYawAccuracy         uint8
	SlopedYawAccuracy       uint8
	NumFramesPerRotation    uint8
	FlatImageID             uint32
	UnkImageID              uint32
	GentleImageID           uint32
	SteepImageID            uint32
}

// VehicleBogieSprite describes bogie sprite layout
type VehicleBogieSprite struct {
	NumAnimationFrames   uint8
	Flags                uint8
	Width                uint8
	HeightNegative       uint8
	HeightPositive       uint8
	NumFramesPerRotation uint8
	FlatImageIDs         uint32
	GentleImageIDs       uint32
	SteepImageIDs        uint32
}

// BogieSpriteFlags
const (
	BogieFlagHasSprites         = 1 << 0
	BogieFlagRotationalSymmetry = 1 << 1
	BogieFlagHasGentleSprites   = 1 << 2
	BogieFlagHasSteepSprites    = 1 << 3
)

// BodySpriteFlags
const (
	BodyFlagHasSprites         = 1 << 0
	BodyFlagRotationalSymmetry = 1 << 1
	BodyFlagHasGentleSprites   = 1 << 3
	BodyFlagHasSteepSprites    = 1 << 4
	BodyFlagHasBrakingLights   = 1 << 5
	BodyFlagHasSpeedAnimation  = 1 << 6
)

// ParseVehicleObject parses a vehicle from decompressed object data
func ParseVehicleObject(header *ObjectHeader, data []byte) (*VehicleObject, error) {
	if len(data) < 0x15E {
		return nil, fmt.Errorf("vehicle data too short: %d bytes", len(data))
	}

	v := &VehicleObject{
		Header: *header,
	}

	// Read the fixed-size portion using a reader
	r := newByteReader(data)

	v.Name = r.readU16()
	v.Mode = TransportMode(r.readU8())
	v.Type = VehicleType(r.readU8())
	v.NumCarComponents = r.readU8()
	v.TrackType = r.readU8()
	v.NumTrackExtras = r.readU8()
	v.CostIndex = r.readU8()
	v.CostFactor = r.readI16()
	v.Reliability = r.readU8()
	v.RunCostIndex = r.readU8()
	v.RunCostFactor = r.readI16()
	v.ColourType = r.readU8()
	v.NumCompatVehicles = r.readU8()

	for i := 0; i < 8; i++ {
		v.CompatibleVehicles[i] = r.readU16()
	}
	for i := 0; i < 4; i++ {
		v.RequiredTrackExtras[i] = r.readU8()
	}

	// Car components at 0x24
	for i := 0; i < 4; i++ {
		v.CarComponents[i] = VehicleCarComponent{
			FrontBogiePosition:   r.readU8(),
			BackBogiePosition:    r.readU8(),
			FrontBogieSpriteIdx:  r.readU8(),
			BackBogieSpriteIdx:   r.readU8(),
			BodySpriteIdx:        r.readU8(),
			EmitterHorizontalPos: r.readU8(),
		}
	}

	// Body sprites at 0x3C
	for i := 0; i < 4; i++ {
		v.BodySprites[i] = VehicleBodySprite{
			NumFlatRotationFrames:   r.readU8(),
			NumSlopedRotationFrames: r.readU8(),
			NumAnimationFrames:      r.readU8(),
			NumCargoLoadFrames:      r.readU8(),
			NumCargoFrames:          r.readU8(),
			NumRollFrames:           r.readU8(),
			HalfLength:              r.readU8(),
			Flags:                   r.readU8(),
			Width:                   r.readU8(),
			HeightNegative:          r.readU8(),
			HeightPositive:          r.readU8(),
			FlatYawAccuracy:         r.readU8(),
			SlopedYawAccuracy:       r.readU8(),
			NumFramesPerRotation:    r.readU8(),
			FlatImageID:             r.readU32(),
			UnkImageID:              r.readU32(),
			GentleImageID:           r.readU32(),
			SteepImageID:            r.readU32(),
		}
	}

	// Bogie sprites at 0xB4
	for i := 0; i < 2; i++ {
		v.BogieSprites[i] = VehicleBogieSprite{
			NumAnimationFrames:   r.readU8(),
			Flags:                r.readU8(),
			Width:                r.readU8(),
			HeightNegative:       r.readU8(),
			HeightPositive:       r.readU8(),
			NumFramesPerRotation: r.readU8(),
			FlatImageIDs:         r.readU32(),
			GentleImageIDs:       r.readU32(),
			SteepImageIDs:        r.readU32(),
		}
	}

	// Performance data at 0xD8
	v.Power = r.readU16()
	v.Speed = r.readU16()
	v.RackSpeed = r.readU16()
	v.Weight = r.readU16()
	v.Flags = r.readU16()
	v.MaxCargo[0] = r.readU8()
	v.MaxCargo[1] = r.readU8()

	// Skip to designed/obsolete at 0x114
	r.skip(0x114 - r.pos)
	v.Designed = r.readU16()
	v.Obsolete = r.readU16()

	return v, nil
}

// Helper for reading binary data
type byteReader struct {
	data []byte
	pos  int
}

func newByteReader(data []byte) *byteReader {
	return &byteReader{data: data}
}

func (r *byteReader) readU8() uint8 {
	if r.pos >= len(r.data) {
		return 0
	}
	v := r.data[r.pos]
	r.pos++
	return v
}

func (r *byteReader) readU16() uint16 {
	if r.pos+2 > len(r.data) {
		return 0
	}
	v := binary.LittleEndian.Uint16(r.data[r.pos:])
	r.pos += 2
	return v
}

func (r *byteReader) readI16() int16 {
	return int16(r.readU16())
}

func (r *byteReader) readU32() uint32 {
	if r.pos+4 > len(r.data) {
		return 0
	}
	v := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return v
}

func (r *byteReader) skip(n int) {
	r.pos += n
	if r.pos > len(r.data) {
		r.pos = len(r.data)
	}
}

// String returns a description of the vehicle
func (v *VehicleObject) String() string {
	name := v.DisplayName
	if name == "" {
		name = v.Header.GetName()
	}
	return fmt.Sprintf("Vehicle{%s, Type: %s, Mode: %s, Power: %d, Speed: %d}",
		name, v.Type, v.Mode, v.Power, v.Speed)
}

// GetSpeedKmh returns the speed in km/h
func (v *VehicleObject) GetSpeedKmh() float64 {
	// Speed16 is in 1/32 mph, convert to km/h
	mph := float64(v.Speed) / 32.0
	return mph * 1.60934
}

// HasBodySprites returns true if this vehicle has body sprites
func (v *VehicleObject) HasBodySprites(idx int) bool {
	if idx < 0 || idx >= 4 {
		return false
	}
	return v.BodySprites[idx].Flags&BodyFlagHasSprites != 0
}

// HasBogieSprites returns true if this vehicle has bogie sprites
func (v *VehicleObject) HasBogieSprites(idx int) bool {
	if idx < 0 || idx >= 2 {
		return false
	}
	return v.BogieSprites[idx].Flags&BogieFlagHasSprites != 0
}

// Dummy usage to avoid unused import error
var _ io.Reader = nil
