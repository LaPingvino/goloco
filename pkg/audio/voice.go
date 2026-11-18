package audio

import (
	"io"
	"math"
	"sync"
	"time"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// Object sounds and lasting handles: the AudioEngine.cpp instances that
// Audio.cpp's object sounds (SoundId 0x8000 | object, getSoundObject) and
// VehicleAudio.cpp keep — create(buffer, channel, attribs), then setVolume,
// setPan and setPitch while it plays, and destroy (OpenLoco reference:
// src/Audio/src/AudioEngine.cpp, src/OpenLoco/src/Audio/Audio.cpp).
//
// AudioEngine's instance count is OpenAL's source count (the device's mono
// plus stereo sources): no selection rule, so none here.

// Attributes is AudioAttributes: volume in hundredths of a dB (the handle's,
// before the channel's), pan −4096…4096, frequency (applyPitch: frequency /
// 22000; under 100 either way plays as recorded), loop.
type Attributes struct {
	Volume, Pan, Frequency int32
	Loop                   bool
}

// Handle is an AudioHandle (0: null).
type Handle uint32

// ObjectSample is a sound object's sample (getSoundBuffer of an object
// SoundId: loadSoundFromWaveMemory of its pcmHeader and data).
func ObjectSample(so *objects.SoundObject) Sample { return Sample{Format: so.Format, PCM: so.PCM} }

// PlayObjectAt is playSound(makeObjectSoundId(object), channel, loc, volume,
// frequency) after the viewport's part (world.SoundAt): volume is the
// call's extra volume plus the viewport's; the object's own volume
// (getVolumeForSoundId: obj->volume) is added; quieter than −100 dB in all,
// it is skipped. Recorded with the object's name.
func (e *Effects) PlayObjectAt(so *objects.SoundObject, ch ChannelId, volume, screenX, frequency int32) {
	if e == nil || e.Disabled || so == nil {
		return
	}
	volume += int32(so.Volume)
	if volume < -10000 {
		return
	}
	pan := screenX
	if pan != 0 {
		pan = CalculatePan(pan, e.ScreenW)
	}
	if e.Record {
		e.Played = append(e.Played, Played{Object: so.Name, Channel: ch, Volume: volume, Pan: pan})
	}
	if e.ctx == nil {
		return
	}
	gain := VolumeToGain(volume + e.channelVol[ChannelMaster] + e.channelVol[ch])
	l, r := panGains(pan)
	p := e.ctx.NewPlayerFromBytes(stereo16Pitch(ObjectSample(so), gain*l, gain*r, pitchOf(frequency)))
	p.Play()
	e.players = append(e.players, p)
}

// pitchOf is applyPitch: |frequency| / 22000, or 1 under 100.
func pitchOf(frequency int32) float64 {
	if frequency >= 100 || frequency <= -100 {
		return math.Abs(float64(frequency)) / 22000
	}
	return 1
}

// voice is an instance that lasts: its attributes and, with a device, the
// stream its player reads.
type voice struct {
	ch     ChannelId
	attr   Attributes
	stream *voiceStream
	player interface {
		Play()
		Close() error
		IsPlaying() bool
	}
	done bool // a one-shot that played out (AL_STOPPED: reclaimFinishedInstances)
}

// Create is AudioEngine create + play: a handle on the channel for the
// sample with the attributes. Without a device the handle still exists
// (its attributes kept), so the callers' state runs the same.
func (e *Effects) Create(s Sample, ch ChannelId, a Attributes) Handle {
	if e == nil {
		return 0
	}
	if e.voices == nil {
		e.voices = map[Handle]*voice{}
	}
	e.nextHandle++
	h := e.nextHandle
	v := &voice{ch: ch, attr: a}
	e.voices[h] = v
	if e.ctx != nil {
		v.stream = newVoiceStream(s, a.Loop)
		e.applyVoice(v)
		if p, err := e.ctx.NewPlayer(v.stream); err == nil {
			p.SetBufferSize(50 * time.Millisecond)
			p.Play()
			v.player = p
		}
	}
	return h
}

// Destroy is AudioEngine destroy: the handle stops and goes.
func (e *Effects) Destroy(h Handle) {
	if e == nil {
		return
	}
	if v, ok := e.voices[h]; ok {
		if v.player != nil {
			_ = v.player.Close()
		}
		delete(e.voices, h)
	}
}

// SetVolume, SetPan and SetPitch are AudioEngine's setters on a live handle.
func (e *Effects) SetVolume(h Handle, volume int32) { e.setAttr(h, func(a *Attributes) { a.Volume = volume }) }
func (e *Effects) SetPan(h Handle, pan int32)       { e.setAttr(h, func(a *Attributes) { a.Pan = pan }) }
func (e *Effects) SetPitch(h Handle, frequency int32) {
	e.setAttr(h, func(a *Attributes) { a.Frequency = frequency })
}

func (e *Effects) setAttr(h Handle, f func(*Attributes)) {
	if e == nil {
		return
	}
	if v, ok := e.voices[h]; ok {
		f(&v.attr)
		e.applyVoice(v)
	}
}

// VoiceAttributes reports a live handle's attributes and channel (tests).
func (e *Effects) VoiceAttributes(h Handle) (Attributes, ChannelId, bool) {
	if e == nil {
		return Attributes{}, 0, false
	}
	v, ok := e.voices[h]
	if !ok {
		return Attributes{}, 0, false
	}
	return v.attr, v.ch, true
}

// Voices is the number of live handles.
func (e *Effects) Voices() int {
	if e == nil {
		return 0
	}
	return len(e.voices)
}

// applyVoice is applyVolume, applyPitch and applyPan on the stream: the
// effective gain (handle + master + channel), the pan's two gains, the
// pitch.
func (e *Effects) applyVoice(v *voice) {
	if v.stream == nil {
		return
	}
	gain := VolumeToGain(v.attr.Volume + e.channelVol[ChannelMaster] + e.channelVol[v.ch])
	l, r := panGains(v.attr.Pan)
	v.stream.set(gain*l, gain*r, pitchOf(v.attr.Frequency))
}

// voiceStream renders a sample as the context's 16-bit stereo at the
// voice's current gains and pitch (linear interpolation, as OpenAL Soft's
// default resampler), looping or ending.
type voiceStream struct {
	mu       sync.Mutex
	l, r     []float64 // the sample's frames, per side
	pos      float64
	gl, gr   float64
	pitch    float64
	loop     bool
	finished bool
}

func newVoiceStream(s Sample, loop bool) *voiceStream {
	st := &voiceStream{loop: loop, pitch: 1}
	ch := int(s.Format.Channels)
	bytesPer := int(s.Format.BitsPerSample / 8)
	if ch < 1 || bytesPer < 1 {
		st.finished = true
		return st
	}
	at := func(i int) float64 {
		if bytesPer == 1 {
			return float64(int(s.PCM[i])-128) * 256
		}
		return float64(int16(uint16(s.PCM[i]) | uint16(s.PCM[i+1])<<8))
	}
	frames := len(s.PCM) / (ch * bytesPer)
	st.l, st.r = make([]float64, frames), make([]float64, frames)
	for f := range frames {
		base := f * ch * bytesPer
		st.l[f] = at(base)
		st.r[f] = st.l[f]
		if ch > 1 {
			st.r[f] = at(base + bytesPer)
		}
	}
	if frames == 0 {
		st.finished = true
	}
	return st
}

func (s *voiceStream) ended() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.finished
}

func (s *voiceStream) set(gl, gr, pitch float64) {
	s.mu.Lock()
	s.gl, s.gr, s.pitch = gl, gr, pitch
	s.mu.Unlock()
}

func (s *voiceStream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.finished {
		return 0, io.EOF
	}
	n := len(s.l)
	clip := func(v float64) uint16 {
		if v > 32767 {
			v = 32767
		} else if v < -32768 {
			v = -32768
		}
		return uint16(int16(v))
	}
	w := 0
	for ; w+4 <= len(p); w += 4 {
		if s.pos >= float64(n) {
			if !s.loop {
				s.finished = true
				break
			}
			s.pos -= float64(n)
		}
		f := int(s.pos)
		t := s.pos - float64(f)
		g := f + 1
		if g >= n {
			g = 0
			if !s.loop {
				g = n - 1
			}
		}
		a := clip((s.l[f]*(1-t) + s.l[g]*t) * s.gl)
		b := clip((s.r[f]*(1-t) + s.r[g]*t) * s.gr)
		p[w], p[w+1], p[w+2], p[w+3] = byte(a), byte(a>>8), byte(b), byte(b>>8)
		s.pos += s.pitch
	}
	if w == 0 && s.finished {
		return 0, io.EOF
	}
	return w, nil
}
