package audio

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// getFalloffModifier and getZoomVolumeModifier (VehicleAudio.cpp).
func TestVehicleAudioModifiers(t *testing.T) {
	for _, c := range []struct{ pan, want int32 }{
		{0, 255}, {2048, 255}, {-2048, 255}, {2049, 255}, {2100, 243}, {-3000, 18}, {3072, 0}, {3073, 0}, {-4096, 0},
	} {
		if g := FalloffModifier(c.pan); g != c.want {
			t.Errorf("FalloffModifier(%d) = %d, want %d", c.pan, g, c.want)
		}
	}
	for z, want := range []int32{0, -35, -70, -70} {
		if g := ZoomVolumeModifier(z); g != want {
			t.Errorf("ZoomVolumeModifier(%d) = %d, want %d", z, g, want)
		}
	}
}

// getVehicleAudioAttributes: volume = max(drivingVolume × max(falloff +
// underground + zoom, 0) / 8 − 8191, −10000), pan the x's, pitch the
// frequency.
func TestVehicleAudioAttributes(t *testing.T) {
	for _, c := range []struct {
		v    VehicleView
		vol  uint8
		want Attributes
	}{
		// the middle of a 640 × 480 screen: falloff 255, zoom 0
		{VehicleView{X: 320, Y: 240}, 200, Attributes{Volume: 200*255/8 - 8191, Pan: 0, Frequency: 22050}},
		// zoomed out twice and below ground: 255 − 70 − 28 = 157
		{VehicleView{X: 320, Y: 240, Zoom: 3, Underground: true}, 200, Attributes{Volume: 200*157/8 - 8191, Frequency: 22050}},
		// at x 600: pan (600 << 16 / 640 − 32768) / 16 = 1792, still 255
		{VehicleView{X: 600, Y: 240}, 100, Attributes{Volume: 100*255/8 - 8191, Pan: 1792, Frequency: 22050}},
		// y 470: pan y (470 << 16 / 480 − 32768) / 16 = 1962; x 630: 1984 — both in
		{VehicleView{X: 630, Y: 470}, 8, Attributes{Volume: 255 - 8191, Pan: 1984, Frequency: 22050}},
		// far off the screen: falloff 0 → −81.91 dB
		{VehicleView{X: 2000, Y: 240}, 255, Attributes{Volume: -8191, Pan: CalculatePan(2000, 640), Frequency: 22050}},
	} {
		if g := VehicleAudioAttributes(c.v, c.vol, 22050, 640, 480); g != c.want {
			t.Errorf("%+v vol %d: %+v, want %+v", c.v, c.vol, g, c.want)
		}
	}
}

// updateSingleVehicleSound: none or no view stops; no handle starts one
// (looping as the object says) with the attributes; a changed driving
// sound stops this update and starts the new one the next; a running one
// takes the new attributes.
func TestUpdateVehicleSound(t *testing.T) {
	e := NewEffects(nil, nil)
	id, active := uint8(soundObjectNull), uint8(soundObjectNull)
	var handle uint32
	s := func(vol uint8, freq uint16) VehicleSound {
		return VehicleSound{DrivingSoundID: &id, Volume: vol, Frequency: freq, Handle: &handle, ActiveSoundID: &active}
	}
	loops := map[uint8]bool{3: true, 4: false}
	sample := func(slot uint8) (Sample, bool, bool) {
		l, ok := loops[slot]
		return Sample{Format: objects.WaveFormat{Channels: 1, BitsPerSample: 16}, PCM: make([]byte, 4)}, l, ok
	}
	view := &VehicleView{X: 320, Y: 240}

	e.UpdateVehicleSound(s(100, 20000), view, 640, 480, sample)
	if handle != 0 || e.Voices() != 0 {
		t.Fatalf("no driving sound: handle %d, %d voices", handle, e.Voices())
	}
	id = 9 // no such sample: nothing starts
	e.UpdateVehicleSound(s(100, 20000), view, 640, 480, sample)
	if handle != 0 {
		t.Fatalf("no sample: handle %d", handle)
	}
	id = 3
	e.UpdateVehicleSound(s(100, 20000), view, 640, 480, sample)
	a, ch, ok := e.VoiceAttributes(Handle(handle))
	if !ok || active != 3 || ch != ChannelVehicles || !a.Loop || a.Volume != 100*255/8-8191 || a.Frequency != 20000 {
		t.Fatalf("start: handle %d active %d %+v on %d", handle, active, a, ch)
	}
	first := handle
	e.UpdateVehicleSound(s(160, 24000), &VehicleView{X: 600, Y: 240}, 640, 480, sample)
	a, _, _ = e.VoiceAttributes(Handle(handle))
	if handle != first || a.Volume != 160*255/8-8191 || a.Pan != 1792 || a.Frequency != 24000 {
		t.Errorf("update: handle %d %+v", handle, a)
	}
	id = 4 // the driving sound changed: stop now, start next time (not looping)
	e.UpdateVehicleSound(s(160, 24000), view, 640, 480, sample)
	if handle != 0 || active != soundObjectNull || e.Voices() != 0 {
		t.Fatalf("change: handle %d active %d, %d voices", handle, active, e.Voices())
	}
	e.UpdateVehicleSound(s(160, 24000), view, 640, 480, sample)
	if a, _, ok = e.VoiceAttributes(Handle(handle)); !ok || active != 4 || a.Loop {
		t.Fatalf("restart: handle %d active %d %+v", handle, active, a)
	}
	e.UpdateVehicleSound(s(160, 24000), nil, 640, 480, sample) // out of view
	if handle != 0 || active != soundObjectNull || e.Voices() != 0 {
		t.Errorf("out of view: handle %d active %d, %d voices", handle, active, e.Voices())
	}
}

// The looping stream plays the sample over and over at the pitch; a
// one-shot ends.
func TestVoiceStream(t *testing.T) {
	pcm := []byte{0, 0x10, 0, 0x20} // two 16-bit mono frames: 4096, 8192
	s := newVoiceStream(Sample{Format: objects.WaveFormat{Channels: 1, BitsPerSample: 16}, PCM: pcm}, true)
	s.set(1, 0.5, 1)
	buf := make([]byte, 4*5)
	if n, err := s.Read(buf); n != 20 || err != nil {
		t.Fatalf("loop read %d %v", n, err)
	}
	for f, want := range []int16{4096, 8192, 4096, 8192, 4096} {
		l := int16(uint16(buf[f*4]) | uint16(buf[f*4+1])<<8)
		r := int16(uint16(buf[f*4+2]) | uint16(buf[f*4+3])<<8)
		if l != want || r != want/2 {
			t.Errorf("frame %d: %d %d, want %d %d", f, l, r, want, want/2)
		}
	}
	one := newVoiceStream(Sample{Format: objects.WaveFormat{Channels: 1, BitsPerSample: 16}, PCM: pcm}, false)
	one.set(1, 1, 1)
	if n, _ := one.Read(buf); n != 8 {
		t.Errorf("one-shot read %d bytes, want 8", n)
	}
	if _, err := one.Read(buf); err == nil {
		t.Error("one-shot: no EOF after the end")
	}
}

// An object sound plays at its own volume (getVolumeForSoundId: obj->volume)
// plus the call's, and is skipped under −100 dB.
func TestPlayObjectAt(t *testing.T) {
	e := NewEffects(nil, nil)
	e.Record = true
	so := &objects.SoundObject{Name: "chuff", Volume: uint32(0xFFFFFFFF - 299)} // −300
	e.PlayObjectAt(so, ChannelVehicles, -1000, 0, 22050)
	e.PlayObjectAt(so, ChannelVehicles, -9800, 0, 22050)
	if len(e.Played) != 1 || e.Played[0].Object != "chuff" || e.Played[0].Volume != -1300 || e.Played[0].Channel != ChannelVehicles {
		t.Errorf("Played = %+v", e.Played)
	}
}
