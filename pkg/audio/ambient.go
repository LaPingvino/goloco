package audio

import (
	"encoding/binary"
	"fmt"
	"os"
)

// The ambient sound: Audio/AmbientAudio.cpp (OpenLoco reference:
// src/OpenLoco/src/Audio/AmbientAudio.cpp). What the main view looks at
// chooses the sea (css3.dat: more than 60 water tiles), the wilderness
// (css2.dat: more than 60 snowy or high tiles) or the forest (css4.dat: more
// than 30 trees), one looping handle on the ambient channel fading in and
// out by 1 dB a tick between −35 dB and the zoom's loudest.

const (
	ambientMinVolume           = -3500 // kAmbientMinVolume
	ambientVolumeChangePerTick = 100   // kAmbientVolumeChangePerTick
	ambientWaterForOcean       = 60    // kAmbientNumWaterTilesForOcean
	ambientTreesForForest      = 30    // kAmbientNumTreeTilesForForest
	ambientMountainsForWild    = 60    // kAmbientNumMountainTilesForWilderness
)

// AmbientNone, AmbientWilderness, AmbientOcean, AmbientForest are the chosen
// sound (std::nullopt, PathId::css2, css3, css4).
const (
	AmbientNone = iota
	AmbientWilderness
	AmbientOcean
	AmbientForest
)

// AmbientFiles are the Data files of the ambient sounds (Environment's
// css2, css3, css4).
var AmbientFiles = [...]string{AmbientWilderness: "css2.dat", AmbientOcean: "css3.dat", AmbientForest: "css4.dat"}

// AmbientMaxVolume is getAmbientMaxVolume: −12 dB up to zoom 2, then −20,
// −30.
func AmbientMaxVolume(zoom int) int32 {
	v := [...]int32{-1200, -1200, -1200, -2000, -3000, -3000}
	return v[min(max(zoom, 0), len(v)-1)]
}

// ChooseAmbient is updateAmbientNoise's choice from the counts: the sea
// first, then the wilderness, then the forest.
func ChooseAmbient(water, wilderness, trees int) int {
	switch {
	case water > ambientWaterForOcean:
		return AmbientOcean
	case wilderness > ambientMountainsForWild:
		return AmbientWilderness
	case trees > ambientTreesForForest:
		return AmbientForest
	}
	return AmbientNone
}

// Ambient is AmbientAudio.cpp's state: _ambientHandle, _ambientVolume,
// _chosenAmbientNoisePathId.
type Ambient struct {
	Handle Handle
	Volume int32
	Chosen int
}

// UpdateAmbient is updateAmbientNoise after the scan: chosen is the
// sound the view asks for (AmbientNone without a map or main view, with
// maxVolume then kAmbientMinVolume). With none, or another while one plays,
// the volume falls a step — below −35 dB the handle goes; a new choice
// starts its loop at −35 dB (sample's ok: the file loaded); the same one
// rises a step towards maxVolume.
func (e *Effects) UpdateAmbient(a *Ambient, chosen int, maxVolume int32, sample func(int) (Sample, bool)) {
	playing := a.Handle != 0 && e.Playing(a.Handle)
	if chosen == AmbientNone || playing && a.Chosen != chosen {
		v := a.Volume - ambientVolumeChangePerTick
		if v < ambientMinVolume {
			a.Chosen = AmbientNone
			if a.Handle != 0 {
				e.Destroy(a.Handle)
				a.Handle = 0
			}
		} else {
			a.Volume = v
			if a.Handle != 0 {
				e.SetVolume(a.Handle, v)
			}
		}
		return
	}
	if a.Chosen != chosen {
		if s, ok := sample(chosen); ok {
			if a.Handle != 0 {
				e.Destroy(a.Handle)
			}
			a.Handle = e.Create(s, ChannelAmbient, Attributes{Volume: ambientMinVolume, Loop: true})
			a.Volume = ambientMinVolume
			a.Chosen = chosen
		}
		return
	}
	a.Volume = min(a.Volume+ambientVolumeChangePerTick, maxVolume)
	if a.Handle != 0 {
		e.SetVolume(a.Handle, a.Volume)
	}
}

// StopAmbient is stopAmbientNoise: the handle goes (the choice and volume
// stay, as upstream's).
func (e *Effects) StopAmbient(a *Ambient) {
	if a.Handle != 0 {
		e.Destroy(a.Handle)
		a.Handle = 0
	}
}

// Playing is AudioEngine isPlaying: a live handle that hasn't ended.
func (e *Effects) Playing(h Handle) bool {
	if e == nil {
		return false
	}
	v, ok := e.voices[h]
	return ok && !v.done && (v.stream == nil || !v.stream.ended())
}

// ReadMusicSample is Audio.cpp loadMusicSample: a RIFF WAVE file read as
// its fixed header lays out (RIFF, size, WAVE, fmt, its length, PCM format
// 1, channels, rate, byte rate, block align, bits, data, its length, PCM).
func ReadMusicSample(path string) (Sample, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Sample{}, err
	}
	if len(b) < 44 {
		return Sample{}, fmt.Errorf("%s: too short", path)
	}
	u32 := func(o int) uint32 { return binary.LittleEndian.Uint32(b[o:]) }
	u16 := func(o int) uint16 { return binary.LittleEndian.Uint16(b[o:]) }
	switch {
	case u32(0) != 0x46464952:
		return Sample{}, fmt.Errorf("%s: invalid signature", path)
	case u32(8) != 0x45564157:
		return Sample{}, fmt.Errorf("%s: invalid format", path)
	case u32(12) != 0x20746d66 && u32(12) != 0x00746d66:
		return Sample{}, fmt.Errorf("%s: invalid format marker", path)
	case u16(20) != 1:
		return Sample{}, fmt.Errorf("%s: not PCM", path)
	case u32(36) != 0x61746164:
		return Sample{}, fmt.Errorf("%s: invalid data marker", path)
	}
	s := Sample{}
	s.Format.FormatTag, s.Format.Channels, s.Format.SamplesPerSec = 1, u16(22), u32(24)
	s.Format.BitsPerSample = u16(34)
	n := int(u32(40))
	if 44+n > len(b) {
		n = len(b) - 44
	}
	s.PCM = b[44 : 44+n]
	return s, nil
}
