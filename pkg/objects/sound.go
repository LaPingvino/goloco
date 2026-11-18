package objects

import (
	"encoding/binary"
	"errors"
	"strings"
)

// SoundObject is Objects/SoundObject.h (0x0C bytes: name, dataOffset,
// shouldLoop, volume), then its string table, an unknown table (a count of
// 16-byte structs and a length), and the sample: SoundObjectData (0x1E:
// var_00, offset, length and a WAVEFORMATEX) followed by length bytes of PCM.
// The vehicle and ambient sounds (Audio.cpp getSoundObject).
type SoundObject struct {
	Header     *ObjectHeader
	Name       string
	ShouldLoop bool   // 0x06
	Volume     uint32 // 0x08
	Format     WaveFormat
	PCM        []byte
}

// WaveFormat is Audio::WAVEFORMATEX (18 bytes).
type WaveFormat struct {
	FormatTag      uint16
	Channels       uint16
	SamplesPerSec  uint32
	AvgBytesPerSec uint32
	BlockAlign     uint16
	BitsPerSample  uint16
	CbSize         uint16
}

// ParseWaveFormat reads an 18-byte little-endian WAVEFORMATEX.
func ParseWaveFormat(f []byte) (WaveFormat, error) {
	if len(f) < 18 {
		return WaveFormat{}, errors.New("WAVEFORMATEX: short")
	}
	return WaveFormat{FormatTag: binary.LittleEndian.Uint16(f), Channels: binary.LittleEndian.Uint16(f[2:]),
		SamplesPerSec: binary.LittleEndian.Uint32(f[4:]), AvgBytesPerSec: binary.LittleEndian.Uint32(f[8:]),
		BlockAlign: binary.LittleEndian.Uint16(f[12:]), BitsPerSample: binary.LittleEndian.Uint16(f[14:]), CbSize: binary.LittleEndian.Uint16(f[16:])}, nil
}

// ParseSoundObject is SoundObject::load (0x0048AFAF); nil when the data is
// too short for its sample.
func ParseSoundObject(header *ObjectHeader, data []byte) *SoundObject {
	if len(data) < 0x0C {
		return nil
	}
	s := &SoundObject{Header: header, ShouldLoop: data[0x06] != 0, Volume: binary.LittleEndian.Uint32(data[0x08:])}
	off, name := walkStringTableLang(data, 0x0C)
	s.Name = name
	if off+8 > len(data) {
		return nil
	}
	numUnk := int(binary.LittleEndian.Uint32(data[off:]))
	off += 8 + numUnk*16 // the count, the (unused) length, the 16-byte structs
	if off+0x1E > len(data) {
		return nil
	}
	length := int(binary.LittleEndian.Uint32(data[off+8:]))
	s.Format, _ = ParseWaveFormat(data[off+12:])
	pcm := off + 0x1E
	if length < 0 || pcm+length > len(data) {
		return nil
	}
	s.PCM = data[pcm : pcm+length]
	return s
}

// ReorderSoundObjects puts the sound objects in the scenario's slot order
// (ObjectType::sound, slots 1–128): a slot whose object isn't loaded is nil.
func (m *ObjectManager) ReorderSoundObjects(order []string) {
	re := make([]*SoundObject, len(order))
	for i, name := range order {
		if name == "" {
			continue
		}
		if lo := m.GetObject(strings.ToUpper(name)); lo != nil {
			if so, ok := lo.Object.(*SoundObject); ok {
				re[i] = so
			}
		}
	}
	m.SoundObjects = re
}

// SoundObjectIndex is the slot of the sound object a header names
// (findObjectHandle: VehicleObject::load's soundObjectId / startSounds,
// SteamObject's soundEffects); -1 when it isn't loaded.
func (m *ObjectManager) SoundObjectIndex(name string) int {
	for i, s := range m.SoundObjects {
		if s != nil && s.Header != nil && strings.EqualFold(strings.TrimSpace(s.Header.GetName()), strings.TrimSpace(name)) {
			return i
		}
	}
	return -1
}

// GetSoundObject is ObjectManager::get<SoundObject>(slot); nil when none.
func (m *ObjectManager) GetSoundObject(slot int) *SoundObject {
	if m == nil || slot < 0 || slot >= len(m.SoundObjects) {
		return nil
	}
	return m.SoundObjects[slot]
}
