package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// StationManager's allocation (World/StationManager.cpp): allocateNewStation,
// deallocateStation and generateNewStationName (0x0048F988) — item ST, the
// ground the station commands build on.

const (
	stationNameNull   = 0xFFFF // StringIds::null: an empty station slot (Station::empty)
	stationFlag5      = 1 << 5 // StationFlags::flag_5: not fully created yet
	stringStationTown = 281    // StringIds::station_town; the 47 names follow it in StationName order
)

// Station names (generateNewStationName's StationName), as offsets from
// station_town.
const (
	snDefault = iota
	snNorth
	snSouth
	snEast
	snWest
	snCentral
	snTransfer
	snHalt
	snValley
	snHeights
	snWoods
	snLakeside
	snExchange
	snAirport
	snOilfield
	snMines
	snDocks
	snAnnexe
	snSidings
	snBranch
	snUpperTown
	snLowerTown
	snHeliport
	snForest
	snJunction
	snCross
	snViews
	snOrd1 // … snOrd20 = snOrd1 + 19
)

// toTownName is StringManager::toTownName: a station_town* id as the town
// name variant the station keeps (kTownNamesStart + the id).
func toTownName(id int) uint16 { return uint16(townNamesStart + id) }

// generateNewStationName picks a station's name for town townID (its slot)
// at (x, y, z), mode 1 airport, 2 heliport, 3 dock: the names its other
// stations in that town don't use, in upstream's order of preference —
// airport / heliport / docks by mode, oilfield / mines near such an
// industry, lakeside (24+ water tiles around, itself dry), forest (40+
// trees) or woods (20+), heights / valley (20 above / below the town
// centre's ground), the plain town name, central (within 9 tiles), north /
// south / east / west, then transfer … views, then 1st … 20th.
func (m *Map) generateNewStationName(townID uint16, x, y, z int32, mode uint8, om *objects.ObjectManager) uint16 {
	var real [27]bool
	var ordinal [20]bool
	for i := range m.StationNames {
		st := &m.StationNames[i]
		if st.Name == stationNameNull || st.Name < townNamesStart || st.Name >= townNamesEnd || st.Town != townID { // isTownName
			continue
		}
		key := int(st.Name) - townNamesStart - stringStationTown // fromTownName(name) − station_town
		switch {
		case key < 0:
		case key < snOrd1:
			real[key] = true
		case key < snOrd1+20:
			ordinal[key-snOrd1] = true
		}
	}
	try := func(n int) (uint16, bool) {
		if !real[n] {
			return toTownName(stringStationTown + n), true
		}
		return 0, false
	}
	surfaceDry := func(px, py int32) bool {
		t := m.TileAt(int(px)>>5, int(py)>>5)
		return t != nil && t.WaterLevel == 0
	}
	switch mode {
	case 1:
		if id, ok := try(snAirport); ok {
			return id
		}
	case 2:
		if id, ok := try(snHeliport); ok {
			return id
		}
	case 3:
		if surfaceDry(x, y) {
			if id, ok := try(snDocks); ok {
				return id
			}
		}
	}
	if m.industryNearPosition(x, y, 1<<13, om) { // IndustryObjectFlags::oilfieldStationName
		if id, ok := try(snOilfield); ok {
			return id
		}
	}
	if m.industryNearPosition(x, y, 1<<14, om) { // minesStationName
		if id, ok := try(snMines); ok {
			return id
		}
	}
	if m.countSurroundingWaterTiles(int(x)>>5, int(y)>>5) >= 24 && surfaceDry(x, y) {
		if id, ok := try(snLakeside); ok {
			return id
		}
	}
	if trees := m.countSurroundingTrees(x, y); trees > 40 {
		if id, ok := try(snForest); ok {
			return id
		}
	} else if trees > 20 {
		if id, ok := try(snWoods); ok {
			return id
		}
	}
	town := m.townBySlot(townID)
	if town != nil {
		if t := m.TileAt(int(town.X)>>5, int(town.Y)>>5); t != nil {
			diff := int(z/4) - int(t.BaseZ)
			if diff > 20 {
				if id, ok := try(snHeights); ok {
					return id
				}
			} else if diff < -20 {
				if id, ok := try(snValley); ok {
					return id
				}
			}
		}
	}
	if id, ok := try(snDefault); ok {
		return id
	}
	if town != nil {
		tx, ty := int32(town.X), int32(town.Y)
		if (abs32(x-tx)+abs32(y-ty))/TileSize <= 9 {
			if id, ok := try(snCentral); ok {
				return id
			}
		}
		for _, d := range []struct {
			n    int
			cond bool
		}{{snNorth, x <= tx && y <= ty}, {snSouth, x >= tx && y >= ty}, {snEast, x <= tx && y >= ty}, {snWest, x >= tx && y <= ty}} {
			if d.cond {
				if id, ok := try(d.n); ok {
					return id
				}
			}
		}
	}
	for _, n := range []int{snTransfer, snHalt, snAnnexe, snSidings, snBranch, snJunction, snCross, snViews} {
		if id, ok := try(n); ok {
			return id
		}
	}
	for i := 0; i < 20; i++ {
		if !ordinal[i] {
			return toTownName(stringStationTown + snOrd1 + i)
		}
	}
	return toTownName(stringStationTown) // upstream falls through to station_town too
}

// townBySlot is TownManager::get(TownId): the town in that slot, or nil.
func (m *Map) townBySlot(id uint16) *scenario.Town {
	for i := range m.Towns {
		if m.Towns[i].ID == id {
			return &m.Towns[i]
		}
	}
	return nil
}

// AllocateNewStation is StationManager::allocateNewStation: the first empty
// station slot for a station at (x, y, z) of owner — off the map or no town
// (town_must_be_built_first) or no free slot (too_many_stations_in_game):
// false. The slot gets the closest town, the owner, a generated name, fresh
// cargo stats (rating 150, origin and industry none), its position, an
// empty label frame and flag_5 (not fully created), the rest cleared.
func (m *Map) AllocateNewStation(x, y, z int32, owner, mode uint8, om *objects.ObjectManager) (int, bool) {
	if x < 0 || y < 0 || int(x)>>5 >= m.Width || int(y)>>5 >= m.Height { // validCoords
		return 0, false
	}
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name != stationNameNull {
			continue
		}
		ti, _, ok := m.ClosestTownAndDensity(x, y)
		if !ok {
			return 0, false
		}
		st.Town = m.Towns[ti].ID
		st.Owner = owner
		st.Name = m.generateNewStationName(st.Town, x, y, z, mode, om)
		for c := range st.Cargo {
			st.Cargo[c] = scenario.StationCargo{Origin: 0xFFFF, Rating: 150, IndustryID: 0xFF}
		}
		st.X, st.Y, st.Z = int16(x), int16(y), int16(z)
		st.LabelFrame = scenario.LabelFrame{}
		st.Flags = stationFlag5
		st.TileSize, st.NoTilesTimeout = 0, 0
		st.Var3B0, st.Var3B1, st.VisitedTypes = 0, 0, 0
		st.AirportRotation, st.AirportStartPos, st.AirportOccupiedEdges = 0, [3]int16{}, 0
		return id, true
	}
	return 0, false
}

// DeallocateStation is StationManager::deallocateStation: a fully created
// station's town has one station fewer; its name (a user string emptied)
// becomes null — the slot is free. PARTIAL: the station window's closing,
// the station list, removeOrdersForStation, removeStationFromCargoStats,
// resetIfHeadingForStation and the messages' references.
func (m *Map) DeallocateStation(id int) {
	if id < 0 || id >= len(m.StationNames) {
		return
	}
	st := &m.StationNames[id]
	if st.Flags&stationFlag5 == 0 {
		if t := m.townBySlot(st.Town); t != nil {
			t.NumStations--
		}
	}
	m.emptyUserString(st.Name)
	st.Name = stationNameNull
}

// stationsUpdateDaily is StationManager::updateDaily (0x0048B244): a station
// with no tiles counts the days; after 10 it is deallocated (a player
// company's owner is disgusted on the 5th). PARTIAL: the towns'
// ratingAdjusted flags and Station::updateCargo with its rating bonus; the
// emotion event (goloco keeps no company emotions yet); the construction
// window's sub_49E1F1 and the invalidation (UI).
func (m *Map) stationsUpdateDaily() {
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull {
			continue
		}
		if st.TileSize != 0 {
			st.NoTilesTimeout = 0
			continue
		}
		st.NoTilesTimeout++
		if st.NoTilesTimeout >= 10 {
			m.DeallocateStation(id)
		}
	}
}
