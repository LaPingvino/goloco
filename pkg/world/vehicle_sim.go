package world

import (
	"log"
	"os"
	"strconv"
	"time"
)

// vehicle_sim.go — the game's vehicle tick: the ported sim (pkg/sim,
// VehicleManager / VehicleHead::update) run on the scenario's own vehicles.
// goloco's earlier straight-segment SimVehicles and their invented
// board→travel→deliver loop are retired (item 5); new vehicles come with the
// Build Vehicle window (item 5b).

// GameTicks is how many game ticks this frame runs, as OpenLoco.cpp paces
// them: update() / fixedUpdate() call tick() at most 40 times a second
// (Engine::UpdateRateHz: a 25 ms accumulator, capped at MaxTimeDeltaMs 500),
// one call per frame; tick() runs numUpdates = the time since the last tick()
// (at most 500 ms) / 31 ms, clamped to 1–3 — 1 when the mouse moved
// (Input::hasPendingMouseInputUpdate). Everything the game ticks — tiles,
// waves, the date, towns, vehicles, the title sequence — steps that many
// times. Before, one Ebitengine update was one game tick: 60 a second on the
// desktop, and on the web page, which draws a few frames a second, the whole
// game crawled.
func (w *World) GameTicks(mouseMoved bool) int {
	now := Now()
	if w.lastUpdate.IsZero() {
		w.lastUpdate, w.lastTickTime = now, now
		return 1
	}
	w.tickAccumulator = min(w.tickAccumulator+now.Sub(w.lastUpdate), 500*time.Millisecond)
	w.lastUpdate = now
	if w.tickAccumulator < 25*time.Millisecond { // fixedUpdate: not yet
		return 0
	}
	w.tickAccumulator -= 25 * time.Millisecond
	since := min(now.Sub(w.lastTickTime), 500*time.Millisecond) // tick(): _time_since_last_tick
	w.lastTickTime = now
	n := min(max(int(since/(31*time.Millisecond)), 1), 3)
	if mouseMoved {
		n = 1
	}
	return n
}

// diagTicks is GOLOCO_DIAG_TICKS=n: n game ticks at once before the first
// (towns grown …), with what grew logged.
func (w *World) diagTicks() {
	if n, err := strconv.Atoi(os.Getenv("GOLOCO_DIAG_TICKS")); err == nil && !w.diagTicksDone { // diagnostics: n game ticks at once, first (towns grown …)
		w.diagTicksDone = true
		roads := func() int {
			c := 0
			for y := range w.Map.Tiles {
				for x := range w.Map.Tiles[y] {
					c += len(w.Map.Tiles[y][x].Roads)
				}
			}
			return c
		}
		had := map[[3]int]bool{}
		for y := range w.Map.Tiles {
			for x := range w.Map.Tiles[y] {
				for i := range w.Map.Tiles[y][x].Roads {
					had[[3]int{x, y, i}] = true
				}
			}
		}
		r0 := roads()
		for range n {
			w.State.Tick(w.tileOnScreen, w.renderer.ObjMgr)
		}
		var news [][2]int // where roads grew (the first few)
		for y := range w.Map.Tiles {
			for x := range w.Map.Tiles[y] {
				for i := range w.Map.Tiles[y][x].Roads {
					if !had[[3]int{x, y, i}] && len(news) < 8 {
						news = append(news, [2]int{x, y})
					}
				}
			}
		}
		log.Printf("[World] GOLOCO_DIAG_TICKS: %d ticks, road pieces %d → %d, towns %d; new road at %v", n, r0, roads(), len(w.Map.Towns), news)
	}
}
