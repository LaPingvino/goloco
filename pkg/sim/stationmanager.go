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

// stationsUpdateDaily is StationManager::updateDaily (0x0048B244): every
// town's ratingAdjusted cleared; a station with no tiles counts the days and
// after 10 is deallocated; each station's cargo updated, and the first of a
// town's stations with a good cargo rating that day raises its owner's rating
// in the town by 1. PARTIAL: the day-5 disgusted emotion (goloco keeps no
// company emotions yet); the construction window's sub_49E1F1 and the
// invalidations (UI).
func (m *Map) stationsUpdateDaily(om *objects.ObjectManager) {
	for i := range m.Towns {
		m.Towns[i].Flags &^= townFlagRatingAdjusted
	}
	for id := range m.StationNames {
		st := &m.StationNames[id]
		if st.Name == stationNameNull {
			continue
		}
		if st.TileSize == 0 {
			st.NoTilesTimeout++
			if st.NoTilesTimeout >= 10 {
				m.DeallocateStation(id)
				continue // upstream goes on with the slot it just emptied; its updateCargo then sees no cargo
			}
		} else {
			st.NoTilesTimeout = 0
		}
		if m.updateStationCargo(id, om) {
			for t := range m.Towns {
				if m.Towns[t].ID == st.Town && m.Towns[t].Flags&townFlagRatingAdjusted == 0 {
					m.Towns[t].Flags |= townFlagRatingAdjusted
					m.adjustCompanyRating(t, st.Owner, 1)
				}
			}
		}
	}
}

const townFlagRatingAdjusted = 1 << 1 // TownFlags::ratingAdjusted

const (
	stationCargoOriginNull = 0xFFFF // StationId::null: an unused cargo slot
)

// updateStationCargo is Station::updateCargo (0x00492683): var_3B0 / var_3B1
// count up; per cargo the en-route age and age, the rating stepped (±2 a day)
// toward calculateCargoRating, cargo lost while the rating is poor (gPrng1, in
// upstream's order); then updateCargoDistribution. Reports whether a cargo's
// rating is good (≥ 100, 50 %).
func (m *Map) updateStationCargo(id int, om *objects.ObjectManager) bool {
	st := &m.StationNames[id]
	good := false
	st.Var3B0 = uint8(min(int(st.Var3B0)+1, 255))
	st.Var3B1 = uint8(min(int(st.Var3B1)+1, 255))
	for i := range st.Cargo {
		c := &st.Cargo[i]
		if c.Origin == stationCargoOriginNull {
			continue
		}
		if c.Quantity != 0 && int(c.Origin) != id {
			c.EnrouteAge = uint8(min(int(c.EnrouteAge)+1, 255))
		} else {
			c.EnrouteAge = 0 // OpenLoco's change: the cargo transfer bug
		}
		c.Age = uint8(min(int(c.Age)+1, 255))
		target := m.calculateCargoRating(st, c)
		c.Rating = uint8(int(c.Rating) + min(max(target-int(c.Rating), -2), 2))
		if c.Rating <= 50 {
			if c.Quantity >= 400 {
				c.Quantity -= uint16(m.Rng1.RandNextRange(1, 32))
			} else if c.Quantity >= 200 {
				c.Quantity -= uint16(m.Rng1.RandNextRange(1, 8))
			}
		}
		if c.Rating >= 100 {
			good = true
		}
		if c.Rating <= 100 && c.Quantity != 0 {
			if int32(c.Rating) <= m.Rng1.RandNextRange(0, 127) {
				c.Quantity = uint16(max(0, int(c.Quantity)-int(m.Rng1.RandNextRange(1, 4))))
			}
		}
	}
	m.UpdateCargoDistribution(id, om)
	return good
}

// calculateCargoRating is Station::calculateCargoRating (0x004927F6): fresh
// cargo, little waiting, a fast and new vehicle rate well; an AI company's
// station without flag 7 / 8 starts from 120. 0…200.
func (m *Map) calculateCargoRating(st *scenario.StationName, c *scenario.StationCargo) int {
	r := 0
	switch {
	case c.Age <= 7:
		r += 165
	case c.Age <= 15:
		r += 130
	case c.Age <= 30:
		r += 85
	case c.Age <= 45:
		r += 40
	}
	r -= 130
	switch {
	case c.Quantity <= 100:
		r += 130
	case c.Quantity <= 200:
		r += 110
	case c.Quantity <= 300:
		r += 90
	case c.Quantity <= 500:
		r += 60
	case c.Quantity <= 1000:
		r += 30
	}
	if st.Flags&(stationFlag7|stationFlag8) == 0 && !m.IsPlayerCompany(int(st.Owner)) {
		r = 120
	}
	if speed := min(int(c.VehicleSpeed), 250); speed > 35 { // Speed16 raw is mph
		r += (speed - 35) / 4
	}
	switch {
	case c.VehicleAge < 1:
		r += 33
	case c.VehicleAge < 2:
		r += 20
	case c.VehicleAge < 4:
		r += 10
	}
	return min(max(r, 0), 200)
}

// StationOwner is Station::owner: kept on the station record (StationNames,
// which AllocateNewStation fills in) — the one place, as upstream has it, so
// a station built in this game is its builder's like a saved game's.
func (m *Map) StationOwner(id int) (uint8, bool) {
	if id < 0 || id >= len(m.StationNames) || m.StationNames[id].Name == stationNameNull {
		return 0, false
	}
	return m.StationNames[id].Owner, true
}
