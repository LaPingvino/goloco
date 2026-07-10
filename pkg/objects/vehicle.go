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

	// ImageOffset is the base index in the dynamic G1 sprite pool for this vehicle's
	// sprites. All FlatImageID / GentleImageID / SteepImageID values in BodySprites
	// and BogieSprites are relative offsets from this base.
	// Set by ParseVehicleObjectWithG1; zero when sprites are not loaded.
	ImageOffset uint32

	// BaseImageID kept for backward compatibility
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

// ParseVehicleObjectWithG1 parses a vehicle object and loads its sprites into
// the G1 dynamic pool via g1.LoadImageTable.
// OpenLoco reference: src/OpenLoco/src/Objects/VehicleObject.cpp::load
func ParseVehicleObjectWithG1(header *ObjectHeader, data []byte, g1 G1Loader) (*VehicleObject, error) {
	v, err := ParseVehicleObject(header, data)
	if err != nil {
		return nil, err
	}
	if g1 == nil {
		return v, nil
	}

	// Locate the image table by walking every variable-length section between
	// the fixed struct and the table, mirroring VehicleObject::load exactly.
	// OpenLoco reference: src/OpenLoco/src/Objects/VehicleObject.cpp::load
	const fixedSize = 0x15E
	const (
		flagRackRail    = 1 << 6 // VehicleObjectFlags::rackRail
		flagAnyRoadType = 1 << 9 // VehicleObjectFlags::anyRoadType
	)
	offset := fixedSize

	// 1. String table: [langID][str\0]...[0xFF]
	for offset < len(data) && data[offset] != 0xFF {
		offset++ // langID byte
		for offset < len(data) && data[offset] != 0 {
			offset++
		}
		if offset < len(data) {
			offset++ // null terminator
		}
	}
	if offset < len(data) {
		offset++ // 0xFF terminator
	}

	// 2. Track/road type header — only for rail/road vehicles without anyRoadType.
	if v.Flags&flagAnyRoadType == 0 &&
		(v.Mode == TransportModeRail || v.Mode == TransportModeRoad) {
		offset += HeaderSize
	}

	// 3. Required track-extra headers.
	offset += int(v.NumTrackExtras) * HeaderSize

	// 4. Cargo section: 2 × [maxCargo u8; if ≠0: (category u16, spriteOffset u8)*
	//    until category 0xFFFF, then the u16 terminator itself].
	for i := 0; i < 2 && offset < len(data); i++ {
		maxCargo := data[offset]
		offset++
		if maxCargo == 0 {
			continue
		}
		for offset+2 <= len(data) && binary.LittleEndian.Uint16(data[offset:offset+2]) != 0xFFFF {
			offset += 3 // cargo category u16 + sprite offset u8
		}
		offset += 2 // 0xFFFF terminator
	}

	// 5. Emitter-animation headers: one per animation slot with type ≠ none.
	//    animation[2] lives at 0x10D in the fixed struct, 3 bytes each, type at +2.
	for i := 0; i < 2; i++ {
		if data[0x10D+i*3+2] != 0 {
			offset += HeaderSize
		}
	}

	// 6. Compatible-vehicle headers.
	offset += int(v.NumCompatVehicles) * HeaderSize

	// 7. Rack-rail object header.
	if v.Flags&flagRackRail != 0 {
		offset += HeaderSize
	}

	// 8. Driving-sound object header (drivingSoundType at 0x119, none = 0).
	if data[0x119] != 0 {
		offset += HeaderSize
	}

	// 9. Start-sound headers (count at 0x15A, low 7 bits, max 3).
	numStartSounds := int(data[0x15A] & 0x7F)
	if numStartSounds > 3 {
		numStartSounds = 3
	}
	offset += numStartSounds * HeaderSize

	if offset >= len(data) {
		return v, nil // no image data section
	}

	imgRes, err := g1.LoadImageTable(data[offset:])
	if err != nil {
		return nil, fmt.Errorf("loading vehicle image table: %w", err)
	}
	v.ImageOffset = imgRes.ImageOffset

	// The image-id fields in the DAT are placeholders; compute them from the
	// frame counts, exactly like upstream. Stored relative to ImageOffset
	// (GetBodySpriteID / bogie lookups add it back).
	// OpenLoco reference: VehicleObject::load after loadImageTable.
	imgOff := uint32(0)
	for i := range v.BodySprites {
		bs := &v.BodySprites[i]
		if bs.Flags&BodyFlagHasSprites == 0 {
			continue
		}
		sym := uint32(1)
		if bs.Flags&BodyFlagRotationalSymmetry != 0 {
			sym = 2
		}
		bs.FlatImageID = imgOff
		bs.FlatYawAccuracy = yawAccuracyFlat(bs.NumFlatRotationFrames)
		nfpr := uint32(bs.NumAnimationFrames) * uint32(bs.NumCargoFrames) * uint32(bs.NumRollFrames)
		if bs.Flags&BodyFlagHasBrakingLights != 0 {
			nfpr++
		}
		bs.NumFramesPerRotation = uint8(nfpr)
		imgOff += nfpr * uint32(bs.NumFlatRotationFrames) / sym

		if bs.Flags&BodyFlagHasGentleSprites != 0 {
			bs.GentleImageID = imgOff
			imgOff += nfpr * 8 / sym // transition frames up/down deg6
			bs.SlopedYawAccuracy = yawAccuracySloped(bs.NumSlopedRotationFrames)
			imgOff += nfpr * uint32(bs.NumSlopedRotationFrames) * 2 / sym // up/down deg12

			if bs.Flags&BodyFlagHasSteepSprites != 0 {
				bs.SteepImageID = imgOff
				imgOff += nfpr * 8 / sym                                      // transition frames up/down deg18
				imgOff += uint32(bs.NumSlopedRotationFrames) * nfpr * 2 / sym // up/down deg25
			}
		}
	}
	for i := range v.BogieSprites {
		bg := &v.BogieSprites[i]
		if bg.Flags&BogieFlagHasSprites == 0 {
			continue
		}
		sym := uint32(1)
		if bg.Flags&BogieFlagRotationalSymmetry != 0 {
			sym = 2
		}
		bg.NumFramesPerRotation = bg.NumAnimationFrames
		bg.FlatImageIDs = imgOff
		imgOff += uint32(bg.NumFramesPerRotation) * 32 / sym

		if bg.Flags&BogieFlagHasGentleSprites != 0 {
			bg.GentleImageIDs = imgOff
			imgOff += uint32(bg.NumFramesPerRotation) * 32 * 2 / sym // up/down deg12

			if bg.Flags&BogieFlagHasSteepSprites != 0 {
				bg.SteepImageIDs = imgOff
				imgOff += uint32(bg.NumFramesPerRotation) * 32 * 2 / sym // up/down deg25
			}
		}
	}

	return v, nil
}

// yawAccuracyFlat mirrors upstream getYawAccuracyFlat.
func yawAccuracyFlat(numFrames uint8) uint8 {
	switch numFrames {
	case 8:
		return 1
	case 16:
		return 2
	case 32:
		return 3
	default:
		return 4
	}
}

// yawAccuracySloped mirrors upstream getYawAccuracySloped.
func yawAccuracySloped(numFrames uint8) uint8 {
	switch numFrames {
	case 4:
		return 0
	case 8:
		return 1
	case 16:
		return 2
	default:
		return 3
	}
}

// GetBodySpriteID returns the G1 sprite index for the given body sprite slot
// (bodyIdx 0-3), rotation direction (0-31 in OpenLoco convention), and flat
// (non-sloped) ground.  Returns 0 and false if the slot has no sprites.
// OpenLoco reference: src/OpenLoco/src/Paint/PaintVehicle.cpp
func (v *VehicleObject) GetBodySpriteID(bodyIdx, direction32 int) (uint32, bool) {
	if bodyIdx < 0 || bodyIdx >= 4 {
		return 0, false
	}
	bs := &v.BodySprites[bodyIdx]
	if bs.Flags&BodyFlagHasSprites == 0 {
		return 0, false
	}

	// Reduce the 32-direction to the number of unique flat frames.
	yawAcc := int(bs.FlatYawAccuracy)
	frameIdx := direction32 >> yawAcc

	// Rotational symmetry: only the first half of frames are stored; the
	// second half mirrors them (renderer should flip horizontally — for now
	// we just clamp to the stored range).
	if bs.Flags&BodyFlagRotationalSymmetry != 0 {
		maxFrame := int(bs.NumFlatRotationFrames)
		if frameIdx >= maxFrame {
			frameIdx = (2*maxFrame - 1) - frameIdx
			if frameIdx < 0 {
				frameIdx = 0
			}
		}
	}

	spriteID := v.ImageOffset + bs.FlatImageID + uint32(frameIdx)*uint32(bs.NumFramesPerRotation)
	return spriteID, true
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
