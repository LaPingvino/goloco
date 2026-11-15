package audio

import (
	"math"
	"testing"
)

// The maths against upstream's formulas, by hand.
func TestEffectsMaths(t *testing.T) {
	for _, c := range []struct{ p, db int32 }{{0, -6000}, {100, 0}, {50, -602}, {10, -2000}} {
		if got := PercentToDb(c.p); got != c.db {
			t.Errorf("PercentToDb(%d) = %d, want %d", c.p, got, c.db)
		}
	}
	if g := VolumeToGain(-2000); math.Abs(g-0.1) > 1e-9 {
		t.Errorf("VolumeToGain(-2000) = %v", g)
	}
	// calculatePan: the centre of a 640 screen is 0, its edges ±2048.
	for _, c := range []struct{ x, want int32 }{{320, 0}, {0, -2048}, {640, 2048}, {160, -1024}} {
		if got := CalculatePan(c.x, 640); got != c.want {
			t.Errorf("CalculatePan(%d, 640) = %d, want %d", c.x, got, c.want)
		}
	}
	l, r := panGains(0)
	if math.Abs(l-r) > 1e-9 {
		t.Errorf("centre gains %v %v", l, r)
	}
	if l, r := panGains(4096); l > 1e-9 || math.Abs(r-1) > 1e-9 {
		t.Errorf("right gains %v %v", l, r)
	}
}

// A queued sound carries its table volume and pan (Audio.cpp playSound).
func TestEffectsRecord(t *testing.T) {
	e := NewEffects(nil, nil)
	e.Record = true
	e.PlaySound(SoundTicker, ChannelUI, KPlayAtCentre)
	e.PlaySound(SoundClickDown, ChannelUI, 160)
	want := []Played{{SoundTicker, ChannelUI, -1900, 0}, {SoundClickDown, ChannelUI, 0, -1024}}
	if len(e.Played) != 2 || e.Played[0] != want[0] || e.Played[1] != want[1] {
		t.Errorf("Played = %+v, want %+v", e.Played, want)
	}
}

func TestStereo16(t *testing.T) {
	s := Sample{PCM: []byte{0x00, 0x40, 0x00, 0xC0}}
	s.Format.Channels, s.Format.BitsPerSample = 1, 16
	out := stereo16(s, 0.5, 1)
	// 0x4000 → 0x2000 left, 0x4000 right; 0xC000 (−16384) → −8192, −16384.
	want := []byte{0x00, 0x20, 0x00, 0x40, 0x00, 0xE0, 0x00, 0xC0}
	if string(out) != string(want) {
		t.Errorf("stereo16 = % x, want % x", out, want)
	}
}
