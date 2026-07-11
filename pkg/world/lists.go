package world

import (
	"fmt"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// lists.go — read-only summaries of the world's towns, stations and vehicles
// for the toolbar list windows (Vehicles / Stations / Towns).  Each entry
// carries a tile coordinate so the caller can centre the camera on it via
// World.SetCamera.

// townSizeNames maps the S5 TownSize enum to a display label.
// OpenLoco reference: include/OpenLoco/World/Town.h enum class TownSize.
var townSizeNames = [...]string{"Hamlet", "Village", "Town", "City", "Metropolis"}

// TownInfo is one row of the Towns list window.
type TownInfo struct {
	Name         string // "Town N"
	TileX, TileY int    // town centre, tile grid
	Population   uint32
	Size         string
	NumBuildings int
}

// TownList returns the loaded scenario's towns.  Town names are stored as
// dynamically-registered StringIds that are non-trivial to resolve, so each
// town is labelled "Town N" and shown with its population and coordinates.
func (w *World) TownList() []TownInfo {
	out := make([]TownInfo, 0, len(w.towns))
	for i, t := range w.towns {
		size := ""
		if int(t.Size) < len(townSizeNames) {
			size = townSizeNames[t.Size]
		}
		out = append(out, TownInfo{
			Name:         fmt.Sprintf("Town %d", i+1),
			TileX:        int(t.X) / 32,
			TileY:        int(t.Y) / 32,
			Population:   t.Population,
			Size:         size,
			NumBuildings: int(t.NumBuildings),
		})
	}
	return out
}

// stationKindNames maps the StationElement StationType field to a label.
// OpenLoco reference: src/OpenLoco/src/Map/StationElement.h.
var stationKindNames = [...]string{"Train", "Road", "Airport", "Docks"}

// StationInfo is one row of the Stations list window.
type StationInfo struct {
	ID           int
	TileX, TileY int
	Kind         string
}

// StationList returns every placed/loaded station on the map.  A station spans
// multiple tiles that share one StationID; only the first tile encountered is
// reported.  This covers both scenario-loaded stations and stations placed at
// runtime (both append a StationElement to their tile).
func (w *World) StationList() []StationInfo {
	seen := make(map[int]bool)
	var out []StationInfo
	for y := range w.tiles {
		for x := range w.tiles[y] {
			for i := range w.tiles[y][x].stations {
				se := &w.tiles[y][x].stations[i]
				id := int(se.StationID)
				if seen[id] {
					continue
				}
				seen[id] = true
				kind := "Station"
				if int(se.StationType) < len(stationKindNames) {
					kind = stationKindNames[se.StationType]
				}
				out = append(out, StationInfo{ID: id, TileX: x, TileY: y, Kind: kind})
			}
		}
	}
	return out
}

// VehicleInfo is one row of the Vehicles list window.
type VehicleInfo struct {
	Name         string
	TileX, TileY int
}

// VehicleList returns the vehicles present in the world: save-loaded entities
// (one entry per VehicleHead, named after its first car's object) followed by
// runtime SimVehicles.
func (w *World) VehicleList() []VehicleInfo {
	var out []VehicleInfo

	// Map each head entity ID to the object name of its first body car.
	bodyName := make(map[uint16]string)
	for i := range w.entities {
		e := &w.entities[i]
		if e.EntityType != scenario.VehicleTypeBody {
			continue
		}
		if _, ok := bodyName[e.HeadID]; !ok {
			bodyName[e.HeadID] = w.vehicleObjectName(e.ObjectID)
		}
	}

	n := 0
	for i := range w.entities {
		e := &w.entities[i]
		if e.EntityType != scenario.VehicleTypeHead {
			continue
		}
		n++
		label := fmt.Sprintf("Vehicle %d", n)
		if name := bodyName[e.ID]; name != "" {
			label += " — " + name
		}
		tx := int(e.TileX) / 32
		ty := int(e.TileY) / 32
		if tx == 0 && ty == 0 {
			// Tile fields unset — fall back to sub-tile world position.
			tx = e.PosX.Tile().Int()
			ty = e.PosY.Tile().Int()
		}
		out = append(out, VehicleInfo{Name: label, TileX: tx, TileY: ty})
	}

	// Runtime simulation vehicles (placed/spawned during play).  These carry no
	// object id of their own, so they are simply numbered after the entities.
	for i, sv := range w.simVehicles {
		out = append(out, VehicleInfo{
			Name:  fmt.Sprintf("Vehicle %d", n+i+1),
			TileX: sv.tileX,
			TileY: sv.tileY,
		})
	}
	return out
}

// vehicleObjectName resolves a vehicle object slot id to its display name via
// the object manager, returning "" when unavailable.
func (w *World) vehicleObjectName(objID uint16) string {
	if w.renderer == nil || w.renderer.ObjMgr == nil {
		return ""
	}
	vs := w.renderer.ObjMgr.Vehicles
	if int(objID) < len(vs) && vs[objID] != nil {
		if name := vs[objID].DisplayName; name != "" {
			return name
		}
		return vs[objID].Header.GetName()
	}
	return ""
}
