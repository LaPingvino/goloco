package world

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
