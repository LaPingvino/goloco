package audio

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SoundId is Audio.h SoundId: the css1.dat samples (0…31); object sounds
// (0x8000 | object index) come with the vehicle audio (plan B step 3).
type SoundId uint16

const (
	SoundClickDown SoundId = iota
	SoundClickUp
	SoundClickPress
	SoundConstruct
	SoundDemolish
	SoundIncome
	SoundCrash
	SoundWater
	SoundSplash1
	SoundSplash2
	SoundWaypoint
	SoundNotification
	SoundOpenWindow
	SoundApplause1
	SoundError
	SoundMultiplayerConnected
	SoundMultiplayerDisconnected
	SoundDemolishTree
	SoundDemolishBuilding
	SoundVehiclePlace
	SoundVehiclePickup
	SoundConstructShip
	SoundTicker
	SoundApplause2
	SoundNewsOooh
	SoundNewsAwww
	SoundBreakdown1
	SoundBreakdown2
	SoundBreakdown3
	SoundBreakdown4
	SoundBreakdown5
	SoundBreakdown6
)

// ChannelId is AudioChannel.h ChannelId.
type ChannelId uint8

const (
	ChannelMaster ChannelId = iota
	ChannelMusic
	ChannelEffects
	ChannelVehicles
	ChannelUI
	ChannelAmbient
	channelCount
)

// KPlayAtCentre is Audio.cpp kPlayAtCentre: the pan argument for a sound in
// the middle.
const KPlayAtCentre = 0x8000

// VolumeDbMin is Audio.h kVolumeDbMin.
const VolumeDbMin = -6000

// soundVolumeTable is Audio.cpp kSoundVolumeTable: each css1 sound's volume
// in hundredths of a dB.
var soundVolumeTable = [32]int32{
	0, 0, 0, 0, 0, -400, 0, 0,
	0, 0, 0, -500, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, -1900, 0,
	0, 0, -600, -600, -600, -600, -600, -600,
}

// PercentToDb is Audio.h percentToDb.
func PercentToDb(percent int32) int32 {
	if percent <= 0 {
		return VolumeDbMin
	}
	if percent >= 100 {
		return 0
	}
	return int32(float32(2000.0) * float32(math.Log10(float64(float32(percent)/100))))
}

// VolumeToGain is AudioEngine.cpp volumeToGain: hundredths of a dB to a gain.
func VolumeToGain(volume int32) float64 { return math.Pow(10, float64(volume)/2000) }

// CalculatePan is Audio.cpp calculatePan: a screen x to a pan in −2048…2048
// (kVpSizeMin 64).
func CalculatePan(coord, screenSize int32) int32 {
	if screenSize < 64 {
		screenSize = 64
	}
	rel := (coord << 16) / screenSize
	return (rel - (1 << 15)) / 16
}

// panGains turns AudioEngine.cpp applyPan's source position (x = pan/4096 on
// the unit circle in front of the listener) into left/right gains. OpenAL
// Soft's stereo mixer does this upstream; this is a constant-power pan over
// the same angle, which matches it at the centre and both sides.
func panGains(pan int32) (l, r float64) {
	p := float64(pan) / 4096
	if p < -1 {
		p = -1
	} else if p > 1 {
		p = 1
	}
	a := (math.Asin(p) + math.Pi/2) / 2 // 0 = full left, π/2 = full right
	return math.Cos(a), math.Sin(a)
}

// Played is one sound the effects mixer started: for tests, the queue that
// PORT-PLANS B asks to assert on.
type Played struct {
	ID      SoundId
	Object  string // an object sound: its name (ID unset)
	Channel ChannelId
	Volume  int32 // the handle's volume, the sound's own volume included
	Pan     int32 // −4096…4096
}

// Effects is the sound-effect side of AudioEngine.cpp: channel volumes and
// one-shot players on the audio context. Record keeps every Played.
type Effects struct {
	ctx        *audio.Context
	samples    []Sample
	channelVol [channelCount]int32
	players    []*audio.Player
	Record     bool
	Played     []Played
	ScreenW    int32 // Ui::width(), for calculatePan
	// Disabled is !Audio::isAudioEnabled (the Audio menu's Mute):
	// playSound returns at once.
	Disabled bool

	// voices are the lasting handles (voice.go: Create … Destroy).
	voices     map[Handle]*voice
	nextHandle Handle
}

// NewEffects is the effects mixer over css1.dat's samples; ctx nil plays
// nothing (no device, GOLOCO_MUTE) but still records.
func NewEffects(ctx *audio.Context, samples []Sample) *Effects {
	return &Effects{ctx: ctx, samples: samples, ScreenW: 640}
}

// SetChannelVolume is AudioEngine.cpp setChannelVolume (hundredths of a dB).
func (e *Effects) SetChannelVolume(ch ChannelId, volume int32) {
	if e != nil && ch < channelCount {
		e.channelVol[ch] = volume
	}
}

// PlaySound is Audio.cpp playSound(id, channel, pan) for the css1 sounds: pan
// is a screen x, or KPlayAtCentre. (The world-position form, with the
// viewport's volume, comes with the first world call site.)
func (e *Effects) PlaySound(id SoundId, ch ChannelId, pan int32) {
	e.play(id, ch, 0, pan)
}

// PlaySoundAt is playSound(id, channel, loc) after the viewport's part
// (world.SoundAt): viewVolume is calculateVolumeFromViewport's, screenX the
// point's x on the screen. Quieter than −100 dB in all, it is skipped.
func (e *Effects) PlaySoundAt(id SoundId, ch ChannelId, viewVolume, screenX int32) {
	e.PlaySoundAtVF(id, ch, viewVolume, screenX, 0)
}

// PlaySoundAtVF is playSound(id, channel, loc, volume, frequency) after the
// viewport's part: volume is the call's extra volume plus the viewport's;
// frequency sets the pitch (AudioEngine applyPitch: frequency / 22000;
// 0, or under 100, plays as recorded).
func (e *Effects) PlaySoundAtVF(id SoundId, ch ChannelId, volume, screenX, frequency int32) {
	if e == nil || int(id) >= len(soundVolumeTable) || soundVolumeTable[id]+volume < -10000 {
		return
	}
	e.playF(id, ch, volume, screenX, frequency)
}

func (e *Effects) play(id SoundId, ch ChannelId, volume, pan int32) { e.playF(id, ch, volume, pan, 0) }

func (e *Effects) playF(id SoundId, ch ChannelId, volume, pan, frequency int32) {
	if e == nil || e.Disabled || int(id) >= len(soundVolumeTable) {
		return
	}
	volume += soundVolumeTable[id]
	if pan == KPlayAtCentre {
		pan = 0
	}
	if pan != 0 {
		pan = CalculatePan(pan, e.ScreenW)
	}
	if e.Record {
		e.Played = append(e.Played, Played{ID: id, Channel: ch, Volume: volume, Pan: pan})
	}
	if e.ctx == nil || int(id) >= len(e.samples) {
		return
	}
	gain := VolumeToGain(volume + e.channelVol[ChannelMaster] + e.channelVol[ch])
	l, r := panGains(pan)
	pitch := 1.0
	if frequency >= 100 || frequency <= -100 {
		pitch = math.Abs(float64(frequency)) / 22000
	}
	p := e.ctx.NewPlayerFromBytes(stereo16Pitch(e.samples[id], gain*l, gain*r, pitch))
	p.Play()
	e.players = append(e.players, p)
}

// Update closes the players that have finished.
func (e *Effects) Update() {
	if e == nil {
		return
	}
	live := e.players[:0]
	for _, p := range e.players {
		if p.IsPlaying() {
			live = append(live, p)
		} else {
			p.Close()
		}
	}
	e.players = live
	for _, v := range e.voices { // reclaimFinishedInstances: a one-shot that ended
		if v.player != nil && !v.done && !v.player.IsPlaying() && !v.attr.Loop {
			_ = v.player.Close()
			v.done = true
		}
	}
}

// stereo16 renders a mono or stereo, 8- or 16-bit PCM sample as the context's
// 16-bit stereo with the two gains. The samples are 22050 Hz, the context's
// rate; a pitch (playSound's frequency) comes with the vehicle sounds.
func stereo16(s Sample, gl, gr float64) []byte {
	ch := int(s.Format.Channels)
	bytesPer := int(s.Format.BitsPerSample / 8)
	if ch < 1 || bytesPer < 1 {
		return nil
	}
	frames := len(s.PCM) / (ch * bytesPer)
	out := make([]byte, frames*4)
	at := func(i int) float64 {
		if bytesPer == 1 {
			return float64(int(s.PCM[i])-128) * 256
		}
		return float64(int16(uint16(s.PCM[i]) | uint16(s.PCM[i+1])<<8))
	}
	clip := func(v float64) uint16 {
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		return uint16(int16(v))
	}
	for f := 0; f < frames; f++ {
		base := f * ch * bytesPer
		lv := at(base)
		rv := lv
		if ch > 1 {
			rv = at(base + bytesPer)
		}
		a, b := clip(lv*gl), clip(rv*gr)
		out[f*4], out[f*4+1], out[f*4+2], out[f*4+3] = byte(a), byte(a>>8), byte(b), byte(b>>8)
	}
	return out
}

// stereo16Pitch is stereo16 played at pitch × its speed (linear
// interpolation, as OpenAL Soft's default resampler); pitch 1 is stereo16.
func stereo16Pitch(s Sample, gl, gr, pitch float64) []byte {
	out := stereo16(s, gl, gr)
	if pitch == 1 || pitch <= 0 || len(out) < 8 {
		return out
	}
	frames := len(out) / 4
	n := int(float64(frames) / pitch)
	res := make([]byte, n*4)
	at := func(f, c int) float64 { return float64(int16(uint16(out[f*4+c*2]) | uint16(out[f*4+c*2+1])<<8)) }
	for i := 0; i < n; i++ {
		pos := float64(i) * pitch
		f := int(pos)
		t := pos - float64(f)
		g := min(f+1, frames-1)
		for c := 0; c < 2; c++ {
			v := uint16(int16(at(f, c)*(1-t) + at(g, c)*t))
			res[i*4+c*2], res[i*4+c*2+1] = byte(v), byte(v>>8)
		}
	}
	return res
}
