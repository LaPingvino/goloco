package main

import (
	"time"

	"github.com/LaPingvino/goloco/pkg/ui"
	"github.com/LaPingvino/goloco/pkg/world"
)

// The game's clock. Upstream paces its ticks, tooltips and timed windows by
// real time (Platform::getTime), and so does goloco — except while a script
// plays: then the clock is virtual and moves a fixed scriptStep per Update,
// so a playthrough ticks the same every run however fast the machine draws
// (under Xvfb frames come slowly and unevenly), and "wait N" is N seconds of
// that clock.
//
// scriptStep is 3 × 31 ms: tick()'s numUpdates is the time since the last
// tick / 31 ms, at most 3 — so every Update runs 3 game ticks (1 while the
// mouse moves), the most upstream's pacing gives a slow frame.
const scriptStep = 93 * time.Millisecond
var clock struct {
	virtual bool
	t       time.Time
}

// gameNow is the game's clock: real time, or the script's virtual time.
func gameNow() time.Time {
	if clock.virtual {
		return clock.t
	}
	return time.Now()
}

// useVirtualClock makes the clock virtual, from a fixed start, for the
// script's run; world and ui (tick pacing, timed windows) read it too.
func useVirtualClock() {
	clock.virtual, clock.t = true, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	world.Now, ui.Now = gameNow, gameNow
}

// advanceClock is one Update step of the virtual clock.
func advanceClock() {
	if clock.virtual {
		clock.t = clock.t.Add(scriptStep)
	}
}
