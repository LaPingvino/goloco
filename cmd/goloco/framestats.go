package main

import (
	"log"
	"runtime"

	"github.com/LaPingvino/goloco/pkg/ui"
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
	d8calls, d8reads         int
	parts                    int // the world's parts last frame
	d8read, d8total          time.Duration
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
	f.d8calls += ui.Draw8Stats.Calls
	f.d8reads += ui.Draw8Stats.Readbacks
	f.d8read += ui.Draw8Stats.Readback
	f.d8total += ui.Draw8Stats.Total
	ui.Draw8Stats.Calls, ui.Draw8Stats.Readbacks, ui.Draw8Stats.Readback, ui.Draw8Stats.Total = 0, 0, 0, 0
	if f.windowStart.IsZero() {
		f.windowStart = start
	}
	if time.Since(f.windowStart) >= 2*time.Second {
		n := max(len(f.draw), 1)
		log.Printf("[Frames] %d frames, %d updates | interval %s | update %s | draw %s | world %s (%d parts) | draw8 %.1f calls (%.1f read back), readback %.1f ms, total %.1f ms a frame | check: %d of %d pixels differ",
			len(f.draw), len(f.upd), summary(f.interval), summary(f.upd), summary(f.draw), summary(f.wld), f.parts,
			float64(f.d8calls)/float64(n), float64(f.d8reads)/float64(n), float64(f.d8read.Microseconds())/1000/float64(n), float64(f.d8total.Microseconds())/1000/float64(n),
			ui.Draw8Stats.CheckDiff, ui.Draw8Stats.CheckPixels)
		f.d8calls, f.d8reads, f.d8read, f.d8total = 0, 0, 0, 0
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
	memStop := startMemProfile()
	path := os.Getenv("GOLOCO_CPUPROFILE")
	if path == "" {
		return memStop
	}
	f, err := os.Create(path)
	if err != nil {
		log.Printf("[Profile] %v", err)
		return func() {}
	}
	// GOLOCO_CPUPROFILE_AFTER=20s starts it later (loading left out).
	if d, err := time.ParseDuration(os.Getenv("GOLOCO_CPUPROFILE_AFTER")); err == nil && d > 0 {
		go func() {
			time.Sleep(d)
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Printf("[Profile] %v", err)
			}
			log.Printf("[Profile] started after %v", d)
			stopAfter(f)
		}()
		return func() { pprof.StopCPUProfile(); f.Close(); memStop() }
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Printf("[Profile] %v", err)
		f.Close()
		return func() {}
	}
	return func() { pprof.StopCPUProfile(); f.Close(); memStop() }
}

// startMemProfile: GOLOCO_MEMPROFILE=file records every allocation (rate 1;
// slow) and writes the profile at exit — to find who makes GPU images each
// frame (their call stacks are in it; the CPU profile only sees the queue).
func startMemProfile() func() {
	path := os.Getenv("GOLOCO_MEMPROFILE")
	if path == "" {
		return func() {}
	}
	runtime.MemProfileRate = 1
	return func() {
		f, err := os.Create(path)
		if err != nil {
			log.Printf("[Profile] %v", err)
			return
		}
		runtime.GC() // the profile is as of the last collection
		pprof.Lookup("allocs").WriteTo(f, 0)
		f.Close()
	}
}

// stopAfter: GOLOCO_CPUPROFILE_FOR=20s stops the CPU profile and writes it
// after that long (a run ended by a timeout never reaches the game's end).
func stopAfter(f *os.File) {
	d, err := time.ParseDuration(os.Getenv("GOLOCO_CPUPROFILE_FOR"))
	if err != nil || d <= 0 {
		return
	}
	time.Sleep(d)
	pprof.StopCPUProfile()
	f.Close()
	log.Printf("[Profile] written after %v", d)
}
