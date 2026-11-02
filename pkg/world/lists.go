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
func (w *World) StationList() []StationInfo { return w.Map.StationList(w.cargoObjects()) }

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
