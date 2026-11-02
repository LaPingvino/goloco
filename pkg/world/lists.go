package world

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/sim"
)

// The list windows' data lives in pkg/sim (lists.go); these pass the loaded
// cargo objects to the station lists.

// StationInfo, TownInfo and VehicleListRow are sim's list rows.
type (
	StationInfo    = sim.StationInfo
	TownInfo       = sim.TownInfo
	VehicleListRow = sim.VehicleListRow
)

func (w *World) cargoObjects() []*objects.CargoObject {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return nil
	}
	return w.renderer.ObjMgr.CargoObjects
}

// StationList is sim.Map.StationList with the loaded cargo objects.
func (w *World) StationList() []StationInfo {
	// Cached between game ticks (LOOP item 0: building it scans every tile, and
	// the station labels asked for it every frame): rebuilt when the sim has
	// ticked, the map changed, or a station was placed (stationsGen).
	k := stationListKey{m: &w.Map, ticks: w.Map.ScenarioTicks, gen: w.stationsGen}
	if w.stationList == nil || w.stationListKey != k {
		w.stationList = w.Map.StationList(w.cargoObjects())
		if w.stationList == nil {
			w.stationList = []StationInfo{}
		}
		w.stationListKey = k
	}
	return w.stationList
}

// stationTiles is each station's first tile, for the station labels drawn
// every frame: rebuilt only when stations change (one is placed through the
// world, or the sim's station records grow), not every tick as StationList.
func (w *World) stationTiles() [][2]int {
	k := stationListKey{m: &w.Map, ticks: uint32(len(w.Map.StationNames)), gen: w.stationsGen}
	if w.stationTileList == nil || w.stationTilesKey != k {
		w.stationTileList = w.stationTileList[:0]
		for _, st := range w.Map.StationList(nil) {
			w.stationTileList = append(w.stationTileList, [2]int{st.TileX, st.TileY})
		}
		if w.stationTileList == nil {
			w.stationTileList = [][2]int{}
		}
		w.stationTilesKey = k
	}
	return w.stationTileList
}

type stationListKey struct {
	m     *sim.Map
	ticks uint32
	gen   int
}

// StationAcceptsKey is sim.Map.StationAcceptsKey with the loaded cargo objects.
func (w *World) StationAcceptsKey(id int) string {
	return w.Map.StationAcceptsKey(id, w.cargoObjects())
}

// ControllingCompany is CompanyManager::getControllingId and its name (a user
// string; "" if unknown).
func (w *World) ControllingCompany() (int, string) {
	id := int(w.Map.ControllingCompany)
	return id, w.CompanyName(id)
}

// ControllingCompanyCash is the controlling company's cash (Company::cash,
// as PlayerInfoPanel shows it); false when the save has no companies.
func (w *World) ControllingCompanyCash() (int64, bool) {
	id := int(w.Map.ControllingCompany)
	if id >= len(w.Map.Companies) {
		return 0, false
	}
	return w.Map.Companies[id].Cash, true
}
