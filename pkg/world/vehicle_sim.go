package world

import (
	"log"
	"os"
	"strconv"
)

// vehicle_sim.go — the game's vehicle tick: the ported sim (pkg/sim,
// VehicleManager / VehicleHead::update) run on the scenario's own vehicles.
// goloco's earlier straight-segment SimVehicles and their invented
// board→travel→deliver loop are retired (item 5); new vehicles come with the
// Build Vehicle window (item 5b).

// TickVehicles is the gameplay tick: money floats, then the vehicles.
func (w *World) TickVehicles() {
	w.tickMoneyFloats()
	w.Map.TitleMode = false
	w.TickSavedVehicles()
}

// TickSavedVehicles runs the ported vehicle tick (sim.Map.UpdateVehicles) on
// the scenario's own vehicles at upstream's rate: one game tick per 31 ms of
// elapsed time, 1–3 per frame (OpenLoco.cpp: numUpdates = clamp(elapsed /
// 31, 1, 3)); goloco's update runs at Ebitengine's 60 per second.
func (w *World) TickSavedVehicles() {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return
	}
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
			w.Map.UpdateVehicles(w.renderer.ObjMgr)
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
	w.vehicleTickMs += 1000.0 / 60
	n := 0
	for w.vehicleTickMs >= 31 && n < 3 {
		w.vehicleTickMs -= 31
		w.Map.UpdateVehicles(w.renderer.ObjMgr)
		n++
	}
	if n == 3 {
		w.vehicleTickMs = 0
	}
}
