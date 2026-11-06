package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// CreateAirport.cpp's helpers (item ST, airports): the extents, the station
// an airport joins and its validation. The command itself follows.

// stationFlagHasAirport is StationFlags::flag_6: the station has an airport
// (set by createAirport, cleared by removeAirport) — one per station.
const stationFlagHasAirport = 1 << 6

// airportExtents is AirportObject::getAirportExtents: the object's min / max
// tile offsets rotated, about the centre tile, ordered.
func airportExtents(a *objects.AirportObject, tx, ty int, rotation uint8) (x0, y0, x1, y1 int) {
	ax0, ay0 := rotOffset(int16(a.MinX), int16(a.MinY), rotation)
	ax1, ay1 := rotOffset(int16(a.MaxX), int16(a.MaxY), rotation)
	x0, y0, x1, y1 = tx+int(ax0), ty+int(ay0), tx+int(ax1), ty+int(ay1)
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return
}

// findNearbyStationAirport is CreateAirport.cpp's (0x00490372):
// findNearbyStation over the airport's extents plus 2 tiles, skipping a
// station that has an airport already.
func (m *Map) findNearbyStationAirport(x, y, z int32, a *objects.AirportObject, rotation, company uint8) (int, bool) {
	x0, y0, x1, y1 := airportExtents(a, int(x)>>5, int(y)>>5, rotation)
	best, minDistance, attached := -1, int32(0x7FFF), false
	for py := max(y0-2, 0); py <= min(y1+2, m.Height-1); py++ {
		for px := max(x0-2, 0); px <= min(x1+2, m.Width-1); px++ {
			t := m.TileAt(px, py)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) {
					continue
				}
				el := &t.Stations[r.Index]
				if el.Ghost || int(el.StationID) >= len(m.StationNames) || m.StationNames[el.StationID].Owner != company {
					continue
				}
				d := chebyshev(int32(px)*TileSize, int32(py)*TileSize, x, y)
				if d < minDistance {
					dz := abs32(int32(el.BaseZ)*4 - z)
					if dz > kMaxStationNearbyDistance {
						continue
					}
					if m.StationNames[el.StationID].Flags&stationFlagHasAirport != 0 {
						continue
					}
					minDistance = d + dz/2
					if minDistance <= kMaxStationNearbyDistance {
						attached = true
					}
					best = int(el.StationID)
				}
			}
		}
	}
	if e := m.findNearbyEmptyStation(x, y, z, company, minDistance); e >= 0 {
		return e, attached
	}
	return best, attached
}

// validateNearbyAirport is CreateAirport.cpp's validateNearbyStation
// (0x00492E48 / 0x00492DBA): none, a new station; full, a failure when
// attached (station_too_large), else new; too spread out (not for the AI), a
// failure when attached (station_too_spread_out), else new.
func (m *Map) validateNearbyAirport(x, y, z int32, a *objects.AirportObject, rotation, company uint8, ai bool) (int, int) {
	id, attached := m.findNearbyStationAirport(x, y, z, a, rotation, company)
	if id < 0 || id >= len(m.StationNames) {
		return nearbyRequiresNewStation, -1
	}
	st := &m.StationNames[id]
	if int(st.TileSize) >= len(st.Tiles) || (!ai && m.exceedsStationSize(st, x, y)) {
		if attached {
			return nearbyFailure, -1
		}
		return nearbyRequiresNewStation, -1
	}
	return nearbyOkay, id
}

const airportFlagIsHelipad = 1 << 1 // AirportObjectFlags::isHelipad

// AirportPlacement is AirportPlacementArgs: the airport's centre tile,
// height, rotation and airport object.
type AirportPlacement struct {
	X, Y, Z  int32
	Rotation uint8
	Type     uint8
}

// airportBuilding is CreateAirport.cpp's createBuilding (0x004930E1): as the
// dock's on land — the clear height from the variation's parts, rounded up to
// 4 and at least 128 ("unsure why"), a 1×1 or 2×2 footprint (largeTiles),
// trees and buildings cleared and anything else colliding, the ground paid
// per height step (and a slope) and levelled, the airport station element
// (its building type the variation) with animation 7.
// PARTIAL: an AI-allocated airport's surface flag.
func (m *Map) airportBuilding(c *Cmd, stationID int, tx, ty int, baseHeight int32, rotation, variation uint8, a *objects.AirportObject, objID uint8, removed RemovedBuildings, flags uint8) (int32, bool) {
	om := c.Objects
	clearHeight := 0
	if int(variation) < len(a.Variations) {
		for _, part := range a.Variations[variation] {
			if int(part) < len(a.PartHeights) {
				clearHeight += int(a.PartHeights[part])
			}
		}
	}
	clearHeight = max((clearHeight+3)&^3, 128)
	argBaseZ, argClearZ := uint8(baseHeight/4), uint8((baseHeight+int32(clearHeight))/4)
	var total int32
	for seq, off := range buildingTileOffsets(a.LargeTiles&(1<<variation) != 0) {
		px0, py0 := tx+int(off[0])/32, ty+int(off[1])/32
		if px0 < 0 || py0 < 0 || px0 >= m.Width || py0 >= m.Height {
			return 0, false // off_edge_of_map
		}
		px, py := int32(px0)*32, int32(py0)*32
		t := m.TileAt(px0, py0)
		if flags&CmdApply != 0 && flags&(CmdGhost|CmdAiAllocated) == 0 {
			t.removeAllWallsBelow(argClearZ)
		}
		clear := func(tt *Tile, k int) ClearResult { // 0x004933B2, as createPort's
			switch tt.Order[k].Kind {
			case scenario.KindTree:
				return m.clearTreeCollision(c, tt, k, px, py, flags, &total)
			case scenario.KindBuilding:
				return m.clearBuildingCollision(c, tt, k, px, py, removed, flags, &total)
			}
			return ClearCollision
		}
		baseZ := min(min(t.BaseZ, argBaseZ), argClearZ)
		clearZ := max(max(t.BaseZ, argBaseZ), argClearZ)
		if ok, _ := m.ApplyClearAtStandardHeight(px, py, baseZ, clearZ, 0xF, 0xF, clear); !ok {
			return 0, false
		}
		t = m.TileAt(px0, py0)
		diff := int32(argBaseZ) - int32(t.BaseZ)
		if diff < 0 {
			diff = -diff
		}
		if t.Slope&0x1F != 0 {
			diff++
		}
		if int(t.TerrainIndex) < len(om.LandObjects) && om.LandObjects[t.TerrainIndex] != nil {
			lo := om.LandObjects[t.TerrainIndex]
			total += m.InflationAdjustedCost(lo.CostFactor, lo.CostIndex, 10) * diff
		}
		if flags&CmdGhost == 0 && flags&CmdApply != 0 && (t.Slope&0x1F != 0 || baseHeight != int32(t.BaseZ)*4) && flags&CmdAiAllocated == 0 {
			t.BaseZ = argBaseZ
			t.Slope = 0
			t.GrowthStage = 0
		}
		if flags&CmdApply == 0 {
			continue
		}
		if flags&(CmdGhost|CmdAiAllocated) == 0 {
			m.removeSurfaceIndustry(px, py, t, c)
			t.setTerrainStyleAsCleared()
		}
		el := scenario.StationElement{StationType: 2, ObjectID: objID, Rotation: rotation, Owner: c.Company, BuildingType: variation,
			SequenceIndex: uint8(seq), BaseZ: argBaseZ, ClearZ: uint8(clearHeight/4) + argBaseZ,
			Ghost: flags&CmdGhost != 0, AiAllocated: flags&CmdAiAllocated != 0}
		if flags&CmdGhost == 0 && stationID >= 0 {
			el.StationID = uint16(stationID)
		}
		t.Stations = append(t.Stations, el)
		t.insertRef(scenario.KindStation, len(t.Stations)-1, argBaseZ)
		for i := range t.Order {
			if r := &t.Order[i]; r.Kind == scenario.KindStation && int(r.Index) == len(t.Stations)-1 {
				r.Flags |= 0xF
				if el.Ghost {
					r.Flags |= elemFlagGhost
				}
				if el.AiAllocated {
					r.Flags |= elemFlagAiAllocated
				}
			}
		}
		m.CreateAnimation(7, px, py, argBaseZ)
	}
	return total, true
}

// CreateAirport is the createAirport game command (CreateAirport.cpp
// 0x00492C41): refused when the nearest town has 4 airports (a rival's: 2) —
// town_will_not_allow_airport_to_be_built_here; applied (not AI, not ghost)
// the town counts one more, as upstream, even if a building then fails. The
// station: the nearby one without an airport, else new (allocateNewStation
// mode 2 for a helipad, else 1). The cost: the airport's (base 6) plus each
// building's (at its rotated position — a 2×2 shifted by the footprint's
// rotation offset — and rotation). Applied (not ghost): addTileToStation,
// flag_6, the airport's start / rotation, its occupied edges cleared, modes,
// centre, label, sub_48D794. PARTIAL: the company observation and emotion,
// the construction sound, the error texts.
func (m *Map) CreateAirport(c *Cmd, flags uint8, a AirportPlacement, g1 G1Raw) (int32, bool) {
	om := c.Objects
	if om == nil || int(a.Type) >= len(om.AirportObjects) || om.AirportObjects[a.Type] == nil || a.Rotation > 3 {
		return 0, false
	}
	ap := om.AirportObjects[a.Type]
	m.LastAdjoiningStation, m.LastAdjoiningStationX, m.LastAdjoiningStationY = -1, -1, -1
	town, _, ok := m.ClosestTownAndDensity(a.X, a.Y)
	if !ok || town < 0 || town >= len(m.Towns) {
		return 0, false
	}
	tw := &m.Towns[town]
	if tw.NumberOfAirports >= 4 || (!m.IsPlayerCompany(int(c.Company)) && tw.NumberOfAirports >= 2) {
		return 0, false // town_will_not_allow_airport_to_be_built_here
	}
	if flags&CmdApply != 0 && flags&(CmdAiAllocated|CmdGhost) == 0 {
		tw.NumberOfAirports++
	}
	ai := flags&CmdAiAllocated != 0
	if flags&CmdGhost != 0 && flags&CmdApply != 0 {
		m.LastAdjoiningStationX, m.LastAdjoiningStationY = a.X, a.Y
		m.LastAdjoiningStation, _ = m.findNearbyStationAirport(a.X, a.Y, a.Z, ap, a.Rotation, c.Company)
	}
	stationID := -1
	if flags&CmdGhost == 0 {
		nameMode := uint8(1)
		if ap.Flags&airportFlagIsHelipad != 0 {
			nameMode = 2
		}
		res, near := m.validateNearbyAirport(a.X, a.Y, a.Z, ap, a.Rotation, c.Company, ai)
		if flags&CmdApply != 0 {
			switch res {
			case nearbyRequiresNewStation:
				if id, ok := m.AllocateNewStation(a.X, a.Y, a.Z, c.Company, nameMode, om); ok {
					stationID = id
					if g1 != nil {
						m.UpdateStationLabel(id, g1)
					}
				}
			case nearbyOkay:
				stationID = near
			}
		} else {
			switch res {
			case nearbyFailure:
				return 0, false
			case nearbyRequiresNewStation:
				id, ok := m.AllocateNewStation(a.X, a.Y, a.Z, c.Company, nameMode, om)
				if !ok {
					return 0, false
				}
				m.DeallocateStation(id)
			}
		}
	}
	total := m.InflationAdjustedCost(ap.BuildCostFactor, ap.CostIndex, 6)
	removed := RemovedBuildings{}
	tx, ty := int(a.X)>>5, int(a.Y)>>5
	for _, b := range ap.Buildings {
		ox, oy := rotOffset(int16(b.X), int16(b.Y), a.Rotation)
		bx, by := tx+int(ox), ty+int(oy)
		if ap.LargeTiles&(1<<b.Index) != 0 {
			off := buildingTileOffsets(true)[a.Rotation]
			bx, by = bx-int(off[0])/32, by-int(off[1])/32
		}
		cost, ok := m.airportBuilding(c, stationID, bx, by, a.Z, (a.Rotation+b.Rotation)&3, b.Index, ap, a.Type, removed, flags)
		if !ok {
			return 0, false
		}
		total += cost
	}
	if flags&CmdGhost == 0 && flags&CmdApply != 0 && stationID >= 0 {
		m.AddTileToStation(stationID, a.X, a.Y, a.Z, a.Rotation, om)
		st := &m.StationNames[stationID]
		st.Flags |= stationFlagHasAirport
		st.AirportStartPos = [3]int16{int16(a.X), int16(a.Y), int16(a.Z)}
		st.AirportRotation = a.Rotation
		st.AirportOccupiedEdges = 0
		m.RecalculateStationModes(stationID, om)
		m.RecalculateStationCenter(stationID, g1)
		if g1 != nil {
			m.UpdateStationLabel(stationID, g1)
		}
		m.UpdateStationCanopies(stationID, om)
	}
	return total, true
}
