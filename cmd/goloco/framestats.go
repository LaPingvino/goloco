package main

import (
	"log"
	"os"
	"runtime/pprof"
	"slices"
	"time"
)

// Frame timing for LOOP item 0 (responsiveness): GOLOCO_FRAMESTATS=1 logs,
// every 2 s, the average / 95th percentile / worst of the frame interval (Draw
// to Draw — what the player feels), Update, Draw in total (prewarm included)
// and the world's part of Draw (pkg/scene's CPU renderer). CPU time only; the
// GPU's work is not in it.
type frameStats struct {
	on                       bool
	lastDraw, windowStart    time.Time
	interval, upd, draw, wld []time.Duration
}

var fstats = frameStats{on: os.Getenv("GOLOCO_FRAMESTATS") != ""}

func (f *frameStats) addUpdate(d time.Duration) {
	if f.on {
		f.upd = append(f.upd, d)
	}
}

func (f *frameStats) addWorld(d time.Duration) {
	if f.on {
		f.wld = append(f.wld, d)
	}
}

// addDraw ends a frame: its Draw time, and the interval since the last one.
func (f *frameStats) addDraw(start time.Time, d time.Duration) {
	if !f.on {
		return
	}
	if !f.lastDraw.IsZero() {
		f.interval = append(f.interval, start.Sub(f.lastDraw))
	}
	f.lastDraw = start
	f.draw = append(f.draw, d)
	if f.windowStart.IsZero() {
		f.windowStart = start
	}
	if time.Since(f.windowStart) >= 2*time.Second {
		log.Printf("[Frames] %d frames | interval %s | update %s | draw %s | world %s",
			len(f.draw), summary(f.interval), summary(f.upd), summary(f.draw), summary(f.wld))
		f.interval, f.upd, f.draw, f.wld = f.interval[:0], f.upd[:0], f.draw[:0], f.wld[:0]
		f.windowStart = time.Now()
	}
}

// summary is "avg/p95/max" in milliseconds.
func summary(d []time.Duration) string {
	if len(d) == 0 {
		return "-"
	}
	s := slices.Clone(d)
	slices.Sort(s)
	var sum time.Duration
	for _, x := range s {
		sum += x
	}
	ms := func(x time.Duration) float64 { return float64(x.Microseconds()) / 1000 }
	return fmtMs(ms(sum/time.Duration(len(s)))) + "/" + fmtMs(ms(s[len(s)*95/100])) + "/" + fmtMs(ms(s[len(s)-1]))
}

func fmtMs(v float64) string {
	if v >= 100 {
		return itoa(int(v + 0.5))
	}
	return itoa(int(v)) + "." + itoa(int(v*10)%10)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}

// startCPUProfile writes a CPU profile to GOLOCO_CPUPROFILE until the returned
// stop runs (the game's end).
func startCPUProfile() func() {
	path := os.Getenv("GOLOCO_CPUPROFILE")
	if path == "" {
		return func() {}
	}
	f, err := os.Create(path)
	if err != nil {
		log.Printf("[Profile] %v", err)
		return func() {}
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Printf("[Profile] %v", err)
		f.Close()
		return func() {}
	}
	return func() { pprof.StopCPUProfile(); f.Close() }
}
