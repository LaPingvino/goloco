package world

import (
	"testing"
	"time"
)

// GameTicks paces game ticks as OpenLoco.cpp: at 60 frames a second about 40
// ticks a second (one tick() per 25 ms, one tick each); at 7 frames a second
// one tick() a frame with 3 ticks (≈ 21 a second); with the mouse moving 1.
func TestGameTicks(t *testing.T) {
	run := func(frame time.Duration, frames int, mouse bool) int {
		w := &World{}
		now := time.Now()
		w.lastUpdate, w.lastTickTime = now, now
		ticks := 0
		for i := 1; i <= frames; i++ {
			// simulate the clock: shift the stored times back instead of sleeping
			w.lastUpdate = w.lastUpdate.Add(-frame)
			w.lastTickTime = w.lastTickTime.Add(-frame)
			ticks += w.GameTicks(mouse)
		}
		return ticks
	}
	if n := run(time.Second/60, 600, false); n < 380 || n > 420 { // 10 s at 60 fps
		t.Errorf("60 fps, 10 s: %d ticks (want ≈ 400)", n)
	}
	if n := run(150*time.Millisecond, 67, false); n < 190 || n > 210 { // 10 s at ~7 fps
		t.Errorf("7 fps, 10 s: %d ticks (want 3 a frame ≈ 200)", n)
	}
	if n := run(150*time.Millisecond, 67, true); n != 67 {
		t.Errorf("7 fps with the mouse moving: %d ticks (want 1 a frame)", n)
	}
}
