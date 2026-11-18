package audio

import (
	"os"
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// The choice (the sea before the wilderness before the forest, each past
// its count) and getAmbientMaxVolume.
func TestChooseAmbient(t *testing.T) {
	for _, c := range []struct{ water, wild, trees, want int }{
		{61, 100, 100, AmbientOcean}, {60, 61, 100, AmbientWilderness}, {60, 60, 31, AmbientForest},
		{60, 60, 30, AmbientNone}, {0, 0, 0, AmbientNone},
	} {
		if g := ChooseAmbient(c.water, c.wild, c.trees); g != c.want {
			t.Errorf("ChooseAmbient(%d, %d, %d) = %d, want %d", c.water, c.wild, c.trees, g, c.want)
		}
	}
	for z, want := range []int32{-1200, -1200, -1200, -2000, -3000, -3000} {
		if g := AmbientMaxVolume(z); g != want {
			t.Errorf("AmbientMaxVolume(%d) = %d, want %d", z, g, want)
		}
	}
}

// updateAmbientNoise's fades: a choice starts its loop at −35 dB and rises
// 1 dB a tick to the zoom's loudest; another choice while it plays falls
// 1 dB a tick and, under −35 dB, goes, so the next tick starts the new one;
// none falls the same way. Stopped (pause) with the same choice, it stays
// silent: the handle is gone but the choice is kept, as upstream.
func TestUpdateAmbient(t *testing.T) {
	e := NewEffects(nil, nil)
	var a Ambient
	smp := func(int) (Sample, bool) { return Sample{Format: objects.WaveFormat{Channels: 1, BitsPerSample: 16}}, true }
	e.UpdateAmbient(&a, AmbientForest, -1200, smp)
	at, ch, ok := e.VoiceAttributes(a.Handle)
	if !ok || a.Chosen != AmbientForest || a.Volume != -3500 || !at.Loop || ch != ChannelAmbient {
		t.Fatalf("start: %+v %+v ch %d", a, at, ch)
	}
	for range 30 {
		e.UpdateAmbient(&a, AmbientForest, -1200, smp)
	}
	if at, _, _ = e.VoiceAttributes(a.Handle); a.Volume != -1200 || at.Volume != -1200 {
		t.Errorf("risen: %d (handle %d), want -1200", a.Volume, at.Volume)
	}
	e.UpdateAmbient(&a, AmbientOcean, -1200, smp)
	if a.Volume != -1300 || a.Chosen != AmbientForest {
		t.Errorf("another choice: %+v, want falling forest", a)
	}
	for range 22 { // −1300 → −3500
		e.UpdateAmbient(&a, AmbientOcean, -1200, smp)
	}
	if a.Handle == 0 || a.Volume != -3500 {
		t.Fatalf("at the floor: %+v", a)
	}
	e.UpdateAmbient(&a, AmbientOcean, -1200, smp) // under −35 dB: gone
	if a.Handle != 0 || a.Chosen != AmbientNone || e.Voices() != 0 {
		t.Fatalf("gone: %+v, %d voices", a, e.Voices())
	}
	e.UpdateAmbient(&a, AmbientOcean, -1200, smp)
	if a.Handle == 0 || a.Chosen != AmbientOcean {
		t.Fatalf("new: %+v", a)
	}
	e.StopAmbient(&a)
	e.UpdateAmbient(&a, AmbientOcean, -1200, smp)
	if a.Handle != 0 || a.Volume != -3400 {
		t.Errorf("after a stop: %+v, want no handle at -3400", a)
	}
}

// loadMusicSample reads css2–4.dat as RIFF WAVE PCM.
func TestReadMusicSample(t *testing.T) {
	for _, f := range AmbientFiles[1:] {
		p := "../../../locomotion/Data/" + f
		if _, err := os.Stat(p); err != nil {
			t.Skip("game data not present")
		}
		s, err := ReadMusicSample(p)
		if err != nil || len(s.PCM) == 0 || s.Format.Channels == 0 || s.Format.BitsPerSample == 0 {
			t.Errorf("%s: %v (%d bytes, %+v)", f, err, len(s.PCM), s.Format)
		}
	}
}
