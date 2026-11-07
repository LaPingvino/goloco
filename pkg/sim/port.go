package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// CreatePort.cpp's helpers (item ST, docks): the dock a new one joins, its
// validation, and the water tests. The command itself follows.

// portBorderOffsets is kPortBorderOffsets (Station.h): the eight tiles round a
// 2×2 dock's edges.
var portBorderOffsets = [8][2]int{{-1, 0}, {-1, 1}, {0, 2}, {1, 2}, {2, 1}, {2, 0}, {1, -1}, {0, -1}}

// findNearbyStationDocks is CreatePort.cpp's (0x0049060C; ai: the AI variant
// 0x004906AC): in the 4×4 round the dock's tile (-1 … +2), the company's own
// dock — not a ghost; an AI-allocated one only for ai, else never — attached;
// else FindNearbyStation.
func (m *Map) findNearbyStationDocks(x, y, z int32, company uint8, ai bool) (int, bool) {
	tx, ty := int(x)>>5, int(y)>>5
	for py := max(ty-1, 0); py <= min(ty+2, m.Height-1); py++ {
		for px := max(tx-1, 0); px <= min(tx+2, m.Width-1); px++ {
			t := m.TileAt(px, py)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) {
					continue
				}
				el := &t.Stations[r.Index]
				if el.StationType != 3 || el.Ghost || el.AiAllocated != ai || el.Owner != company {
					continue
				}
				return int(el.StationID), true
			}
		}
	}
	return m.FindNearbyStation(x, y, z, company)
}

// validateNearbyDock is CreatePort.cpp's validateNearbyStation (0x0048BDCE /
// 0x0048BD40): none nearby, a new station; full, a failure when attached
// (station_too_large), else new; too spread out (not for the AI), a failure
// when attached (station_too_spread_out), else new.
func (m *Map) validateNearbyDock(x, y, z int32, company uint8, ai bool) (int, int) {
	id, attached := m.findNearbyStationDocks(x, y, z, company, ai)
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

// isValidWaterIndustryPort is CreatePort.cpp's: a non-ghost industry built on
// water (IndustryObjectFlags::builtOnWater) on one of the dock's border tiles.
func (m *Map) isValidWaterIndustryPort(x, y int32, om *objects.ObjectManager) bool {
	tx, ty := int(x)>>5, int(y)>>5
	for _, o := range portBorderOffsets {
		px, py := tx+o[0], ty+o[1]
		if px < 0 || py < 0 || px >= m.Width || py >= m.Height { // validCoords
			continue
		}
		t := m.TileAt(px, py)
		for _, r := range t.Order {
			if r.Kind != scenario.KindIndustry || int(r.Index) >= len(t.Industries) || r.Flags&elemFlagGhost != 0 {
				continue
			}
			ind := m.industryByID(t.Industries[r.Index].IndustryID)
			if ind == nil || om == nil {
				continue
			}
			if obj := om.GetIndustryObjectByIndex(int(ind.ObjectID)); obj != nil && obj.Flags&industryBuiltOnWater != 0 {
				return true
			}
		}
	}
	return false
}

// tileHasWater is CreatePort.cpp's: the tile is on the map and its surface
// has water.
func (m *Map) tileHasWater(x, y int32) bool {
	tx, ty := int(x)>>5, int(y)>>5
	if x < 0 || y < 0 || tx >= m.Width || ty >= m.Height {
		return false
	}
	t := m.TileAt(tx, ty)
	return t != nil && t.WaterLevel != 0
}

// PortPlacement is PortPlacementArgs: the dock's tile (its north corner of
// the 2×2), height, rotation and dock object.
type PortPlacement struct {
	X, Y, Z  int32
	Rotation uint8
	Type     uint8
}

// rotationToBuildingFront is kRotationToBuildingFront: the two tiles in front
// of a dock of each rotation, which must be water.
var rotationToBuildingFront = [4][2][2]int32{
	{{64, 0}, {64, 32}}, {{0, -32}, {32, -32}}, {{-32, 32}, {-32, 0}}, {{32, 64}, {0, 64}},
}

// portBuilding is CreatePort.cpp's createBuilding (0x00493F0E): on water a
// dock only next to a water industry; the 2×2's clear height from its parts
// (rounded up to 4); per tile: on the map, walls below removed, cleared —
// trees and buildings go, anything else (a surface too) collides —; on water
// at the water's height and not a ship channel, else between the ground and
// the dock, paid per height step (and a slope) by the land, the ground
// levelled; the station element (type docks, its building type and sequence)
// with its animation (type 8). PARTIAL: an AI-allocated dock's surface flag.
func (m *Map) portBuilding(c *Cmd, stationID int, a PortPlacement, flags uint8, removed RemovedBuildings, buildingType uint8) (int32, bool) {
	om := c.Objects
	waterIndustry := false
	if tx, ty := int(a.X)>>5, int(a.Y)>>5; tx >= 0 && ty >= 0 && tx < m.Width && ty < m.Height && m.TileAt(tx, ty).WaterLevel != 0 {
		if waterIndustry = m.isValidWaterIndustryPort(a.X, a.Y, om); !waterIndustry {
			return 0, false // can_only_be_built_on_water_next_to_water_based_industry
		}
	}
	obj := om.DockObjects[a.Type]
	clearHeight := 0
	if int(buildingType) < len(obj.Variations) {
		for _, part := range obj.Variations[buildingType] {
			if int(part) < len(obj.PartHeights) {
				clearHeight += int(obj.PartHeights[part])
			}
		}
	}
	clearHeight = (clearHeight + 3) &^ 3
	argBaseZ, argClearZ := uint8(a.Z/4), uint8((a.Z+int32(clearHeight))/4)
	var total int32
	for seq, off := range buildingTileOffsets(true) {
		tx, ty := int(a.X+off[0])>>5, int(a.Y+off[1])>>5
		if tx < 0 || ty < 0 || tx >= m.Width || ty >= m.Height {
			return 0, false // off_edge_of_map
		}
		px, py := int32(tx)*32, int32(ty)*32
		t := m.TileAt(tx, ty)
		if flags&CmdApply != 0 && flags&(CmdGhost|CmdAiAllocated) == 0 {
			t.removeAllWallsBelow(argClearZ)
		}
		clear := func(tt *Tile, k int) ClearResult { // 0x0049434F: tileClearFunction colliding on surfaces too
			switch tt.Order[k].Kind {
			case scenario.KindTree:
				return m.clearTreeCollision(c, tt, k, px, py, flags, &total)
			case scenario.KindBuilding:
				return m.clearBuildingCollision(c, tt, k, px, py, removed, flags, &total)
			}
			return ClearCollision
		}
		if waterIndustry {
			if int32(t.WaterLevel)*16 != a.Z {
				return 0, false
			}
			if t.Type6Flag {
				return 0, false // water_channel_currently_needed_by_ships
			}
			if ok, _ := m.ApplyClearAtStandardHeight(px, py, argBaseZ, argClearZ, 0xF, 0xF, clear); !ok {
				return 0, false
			}
		} else {
			baseZ := min(min(t.BaseZ, argBaseZ), argClearZ)
			clearZ := max(max(t.BaseZ, argBaseZ), argClearZ)
			if ok, _ := m.ApplyClearAtStandardHeight(px, py, baseZ, clearZ, 0xF, 0xF, clear); !ok {
				return 0, false
			}
			t = m.TileAt(tx, ty)
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
			if flags&CmdGhost == 0 && flags&CmdApply != 0 && (t.Slope&0x1F != 0 || a.Z != int32(t.BaseZ)*4) {
				if flags&CmdAiAllocated == 0 {
					t.BaseZ = argBaseZ // base and clear height (the sim tile keeps one)
					t.Slope = 0        // the slope and the snow coverage (its top bits)
					t.GrowthStage = 0
				}
			}
		}
		if flags&CmdApply == 0 {
			continue
		}
		if flags&(CmdGhost|CmdAiAllocated) == 0 {
			m.removeSurfaceIndustry(px, py, t, c)
			t.setTerrainStyleAsCleared()
		}
		el := scenario.StationElement{StationType: 3, ObjectID: a.Type, Rotation: a.Rotation, Owner: c.Company, BuildingType: buildingType,
			SequenceIndex: uint8(seq), BaseZ: argBaseZ, ClearZ: uint8(clearHeight/4) + argBaseZ,
			Ghost: flags&CmdGhost != 0, AiAllocated: flags&CmdAiAllocated != 0}
		if flags&CmdGhost == 0 && stationID >= 0 {
			el.StationID = uint16(stationID)
		}
		t.Stations = append(t.Stations, el)
		t.insertRef(scenario.KindStation, len(t.Stations)-1, argBaseZ)
		for i := range t.Order { // insertElement(…, 0xF): all quarters; the ghost / AI flags
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
		m.CreateAnimation(8, px, py, argBaseZ)
	}
	return total, true
}

// CreatePort is the createPort game command (CreatePort.cpp 0x00493AA7): a
// dock (2×2) of a.Type with water at its front, joining the company's dock it
// touches or a new station (allocateNewStation mode 3). The cost: the dock's
// (inflation-adjusted, base 7) plus the building's. Applied (not ghost):
// addTileToStation, the modes, centre, label and sub_48D794.
// PARTIAL: the company observation and thinking emotion, the construction
// sound, the error texts.
func (m *Map) CreatePort(c *Cmd, flags uint8, a PortPlacement, g1 G1Raw) (int32, bool) {
	c.Pos = &[3]int32{a.X + 16, a.Y + 16, a.Z} // setPosition(args.pos + (16, 16, 0)): where the money effect shows
	om := c.Objects
	if om == nil || int(a.Type) >= len(om.DockObjects) || om.DockObjects[a.Type] == nil || a.Rotation > 3 {
		return 0, false
	}
	for _, f := range rotationToBuildingFront[a.Rotation] {
		if !m.tileHasWater(a.X+f[0], a.Y+f[1]) {
			return 0, false // requires_water_in_front_of_dock
		}
	}
	ai := flags&CmdAiAllocated != 0
	m.LastAdjoiningStation, m.LastAdjoiningStationX, m.LastAdjoiningStationY = -1, -1, -1
	if flags&CmdGhost != 0 && flags&CmdApply != 0 {
		m.LastAdjoiningStationX, m.LastAdjoiningStationY = a.X, a.Y
		m.LastAdjoiningStation, _ = m.findNearbyStationDocks(a.X, a.Y, a.Z, c.Company, ai)
	}
	stationID := -1
	if flags&CmdGhost == 0 {
		res, near := m.validateNearbyDock(a.X, a.Y, a.Z, c.Company, ai)
		if flags&CmdApply != 0 {
			switch res {
			case nearbyRequiresNewStation:
				if id, ok := m.AllocateNewStation(a.X, a.Y, a.Z, c.Company, 3, om); ok {
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
				id, ok := m.AllocateNewStation(a.X, a.Y, a.Z, c.Company, 3, om)
				if !ok {
					return 0, false
				}
				m.DeallocateStation(id)
			}
		}
	}
	dock := om.DockObjects[a.Type]
	total := m.InflationAdjustedCost(dock.BuildCostFactor, dock.CostIndex, 7)
	cost, ok := m.portBuilding(c, stationID, a, flags, RemovedBuildings{}, 0)
	if !ok {
		return 0, false
	}
	total += cost
	if flags&CmdGhost == 0 && flags&CmdApply != 0 && stationID >= 0 {
		m.AddTileToStation(stationID, a.X, a.Y, a.Z, a.Rotation, om)
		m.RecalculateStationModes(stationID, om)
		m.RecalculateStationCenter(stationID, g1)
		if g1 != nil {
			m.UpdateStationLabel(stationID, g1)
		}
		m.UpdateStationCanopies(stationID, om)
	}
	return total, true
}

// PortRemoval is PortRemovalArgs: the dock's tile and height.
type PortRemoval struct {
	X, Y, Z int32
}

// stationElAt is RemovePort.cpp's getStationEl: the tile's first station
// element at that height (its order position, -1 none).
func (m *Map) stationElAt(x, y, z int32) (*Tile, int) {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return nil, -1
	}
	for k, r := range t.Order {
		if r.Kind == scenario.KindStation && int(r.Index) < len(t.Stations) && int32(t.Stations[r.Index].BaseZ)*4 == z {
			return t, k
		}
	}
	return t, -1
}

// RemovePort is the removePort game command (RemovePort.cpp 0x00494570): the
// dock at a.X, a.Y taken off — refused without a station element there, or
// while a boat lies at it (flag_6: currently_in_use_by_at_least_one_vehicle).
// As upstream, nobody's ownership is checked. The cost: the dock's sell cost
// (base 7; negative, a refund). removePortTileElements (0x00494706): each of
// the 2×2 needs its station element at that height (else refused); applied,
// it goes. Then, unless a ghost: the tile out of the station with its cargo
// recalculated, the modes, centre and label. PARTIAL: the surfaces' AI
// allocation flag reset (goloco's tiles don't keep it), the invalidations.
func (m *Map) RemovePort(c *Cmd, flags uint8, a PortRemoval, g1 G1Raw) (int32, bool) {
	om := c.Objects
	t, k := m.stationElAt(a.X, a.Y, a.Z)
	if k < 0 {
		return 0, false
	}
	el := t.Stations[t.Order[k].Index]
	stationID := -1
	if flags&CmdGhost == 0 {
		stationID = int(el.StationID)
	}
	if el.Flag6 {
		return 0, false // currently_in_use_by_at_least_one_vehicle
	}
	if om == nil || int(el.ObjectID) >= len(om.DockObjects) || om.DockObjects[el.ObjectID] == nil {
		return 0, false
	}
	dock := om.DockObjects[el.ObjectID]
	total := m.InflationAdjustedCost(dock.SellCostFactor, dock.CostIndex, 7)
	for _, off := range buildingTileOffsets(true) { // removePortTileElements
		tt, kk := m.stationElAt(a.X+off[0], a.Y+off[1], a.Z)
		if kk < 0 {
			return 0, false
		}
		if flags&CmdApply != 0 {
			tt.removeStationElement(kk)
		}
	}
	if flags&CmdGhost == 0 && flags&CmdApply != 0 && stationID >= 0 && stationID < len(m.StationNames) {
		m.RemoveTileFromStationAndRecalcCargo(stationID, a.X, a.Y, a.Z, el.Rotation, om)
		m.RecalculateStationModes(stationID, om)
		m.RecalculateStationCenter(stationID, g1)
		if g1 != nil {
			m.UpdateStationLabel(stationID, g1)
		}
	}
	return total, true
}

// dockSearchArea is getDockPlacementArgs' searchArea: per side of the 2×2,
// the two tiles beyond it.
var dockSearchArea = [4][2][2]int{{{-1, 0}, {-1, 1}}, {{0, 2}, {1, 2}}, {{2, 0}, {2, 1}}, {{0, -1}, {1, -1}}}

// DockPlacementAt is StationTab's getDockPlacementArgs for the tile at x, y
// (the surface or water under the cursor): the water round the 2×2 counted per
// side — a tile whose water sits one step above a double-height slope's base
// is skipped, though its height still counts — and a water industry's side
// noted; none: no placement. The dock at the water's height, facing the
// industry, else the first side with two water tiles, else with one, else 0.
// (The counts go to the opposite side, (side + 2) & 3, as upstream.)
func (m *Map) DockPlacementAt(x, y int32, typ uint8, om *objects.ObjectManager) (PortPlacement, bool) {
	var water [4]int
	industrySide := -1
	waterHeight := 0
	tx, ty := int(x)>>5, int(y)>>5
	for side := 0; side < 4; side++ {
		for _, o := range dockSearchArea[side] {
			px, py := tx+o[0], ty+o[1]
			if px < 0 || py < 0 || px >= m.Width || py >= m.Height {
				continue
			}
			t := m.TileAt(px, py)
			surfaceFound := false
			for _, r := range t.Order {
				if surfaceFound && r.Kind == scenario.KindIndustry && int(r.Index) < len(t.Industries) && r.Flags&elemFlagGhost == 0 {
					if ind := m.industryByID(t.Industries[r.Index].IndustryID); ind != nil && om != nil {
						if obj := om.GetIndustryObjectByIndex(int(ind.ObjectID)); obj != nil && obj.Flags&industryBuiltOnWater != 0 {
							industrySide = side
						}
					}
				}
				if r.Kind != scenario.KindSurface {
					continue
				}
				surfaceFound = true
				if t.WaterLevel == 0 {
					continue
				}
				waterHeight = int(t.WaterLevel) * 4                     // kMicroToSmallZStep
				if waterHeight-4 == int(t.BaseZ) && t.Slope&0x10 != 0 { // isSlopeDoubleHeight
					continue
				}
				water[(side+2)&3]++
			}
		}
	}
	if waterHeight == 0 {
		return PortPlacement{}, false
	}
	a := PortPlacement{X: int32(tx) * 32, Y: int32(ty) * 32, Z: int32(waterHeight) * 4, Type: typ}
	switch {
	case industrySide >= 0:
		a.Rotation = uint8(industrySide)
	default:
		for _, need := range []int{2, 1} {
			found := false
			for side, n := range water {
				if n >= need {
					a.Rotation, found = uint8(side), true
					break
				}
			}
			if found {
				break
			}
		}
	}
	return a, true
}

// UpdateAvailableAirportAndDockOptions is Construction's (Common.cpp): an
// obsolete last airport forgotten; with none, the first available airport —
// kept only when the player has an air vehicle unlocked; the same for docks
// and water vehicles (as upstream, the dock isn't checked for obsolescence).
func (m *Map) UpdateAvailableAirportAndDockOptions(om *objects.ObjectManager) {
	if om == nil {
		return
	}
	year := int(m.CurrentYear)
	if a := int(m.LastAirport); m.LastAirport != 0xFF && a < len(om.AirportObjects) && om.AirportObjects[a] != nil && year > int(om.AirportObjects[a].ObsoleteYear) {
		m.LastAirport = 0xFF
	}
	unlocked := func(mode objects.TransportMode) bool {
		for i, v := range om.Vehicles {
			if v != nil && v.Mode == mode && m.isVehicleIndexUnlocked(m.ControllingCompany, uint16(i)) {
				return true
			}
		}
		return false
	}
	if m.LastAirport == 0xFF {
		if list := om.AvailableAirports(year); len(list) > 0 {
			m.LastAirport = uint8(list[0])
			if !unlocked(objects.TransportModeAir) {
				m.LastAirport = 0xFF
			}
		}
	}
	if m.LastShipPort == 0xFF {
		if list := om.AvailableDocks(year); len(list) > 0 {
			m.LastShipPort = uint8(list[0])
			if !unlocked(objects.TransportModeWater) {
				m.LastShipPort = 0xFF
			}
		}
	}
}
