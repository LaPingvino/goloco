package main

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/audio"
)

// An error window plays the error sound at its centre (Error.cpp
// createErrorWindow).
func TestOpenErrorPlaysSound(t *testing.T) {
	g := &Game{sw: 640, sh: 480, mouseX: 100, mouseY: 100, sfx: audio.NewEffects(nil, nil)}
	g.sfx.Record = true
	g.openError("Can't build here", "")
	if len(g.sfx.Played) != 1 || g.sfx.Played[0].ID != audio.SoundError || g.sfx.Played[0].Channel != audio.ChannelUI {
		t.Fatalf("Played = %+v", g.sfx.Played)
	}
	e := g.errorWin
	if want := audio.CalculatePan(int32(e.X+e.W/2), 640); g.sfx.Played[0].Pan != want {
		t.Errorf("pan %d, want %d", g.sfx.Played[0].Pan, want)
	}
}
