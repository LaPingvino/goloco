package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Station cargo — World/Station.cpp updateCargoDistribution (0x0049218F-side)
// and deliverCargoToStation (0x0042F489), over the station records' full
// StationCargoStats (Map.StationNames[id].Cargo).

const maxStationCargoDensity = 15 // Limits::kMaxStationCargoDensity

// StationCargoDensities is each cargo's densityPerTile at a station (what
// paintStationCargo draws).
func (m *Map) StationCargoDensities(id int) [32]uint8 {
	var out [32]uint8
	if id < 0 || id >= len(m.StationNames) {
		return out
	}
	for i, c := range m.StationNames[id].Cargo {
		out[i] = c.DensityPerTile
	}
	return out
}

// UpdateCargoDistribution is Station::updateCargoDistribution: per cargo,
// ceil(quantity / tiles / 2^stationCargoDensity), at most 15; when any
// changed, each station tile is invalidated (heightFloor(z) to + 32, full).
func (m *Map) UpdateCargoDistribution(id int, om *objects.ObjectManager) {
	if id < 0 || id >= len(m.StationNames) {
		return
	}
	st := &m.StationNames[id]
	changed := false
	for i := range st.Cargo {
		c := &st.Cargo[i]
		d := 0
		if c.Quantity != 0 && st.TileSize != 0 {
			d = int(c.Quantity) / int(st.TileSize)
			var shift uint8
			if i < len(om.CargoObjects) && om.CargoObjects[i] != nil {
				shift = om.CargoObjects[i].StationCargoDensity
			}
			d += 1<<shift - 1
			d >>= shift
			d = min(d, maxStationCargoDensity)
		}
		if c.DensityPerTile != uint8(d) {
			c.DensityPerTile = uint8(d)
			changed = true
		}
	}
	if changed {
		for k := 0; k < int(st.TileSize) && k < len(st.Tiles); k++ {
			tl := st.Tiles[k]
			z := int32(tl[2]) &^ 3 // World::heightFloor
			m.invalidateTile(int32(tl[0]), int32(tl[1]), z, z+32, ZoomFull, defaultRadius)
		}
	}
}

// boundAdd16 is Math::Bound::add for uint16.
func boundAdd16(a uint16, b uint32) uint16 {
	return uint16(min(uint32(a)+b, 0xFFFF))
}

// DeliverCargoToStation is Station::deliverCargoToStation: cargo produced
// here (industries, towns) joins the waiting quantity, fresh, from here.
func (m *Map) DeliverCargoToStation(id int, cargoType, qty uint8, om *objects.ObjectManager) {
	if id < 0 || id >= len(m.StationNames) || cargoType >= 32 {
		return
	}
	c := &m.StationNames[id].Cargo[cargoType]
	c.Quantity = boundAdd16(c.Quantity, uint32(qty))
	c.EnrouteAge = 0
	c.Origin = uint16(id)
	m.UpdateCargoDistribution(id, om)
}

// TownByID is TownManager::get: the town in slot id (Towns skips empty
// slots, so this looks it up by its ID).
func (m *Map) TownByID(id uint16) *scenario.Town {
	for i := range m.Towns {
		if m.Towns[i].ID == id {
			return &m.Towns[i]
		}
	}
	return nil
}

// DeliverCargoToTown is Station::deliverCargoToTown: the station's town
// counts the cargo among this month's deliveries.
func (m *Map) DeliverCargoToTown(stationID int, cargoType uint8, qty uint16) {
	if stationID < 0 || stationID >= len(m.StationNames) || cargoType >= 32 {
		return
	}
	if t := m.TownByID(m.StationNames[stationID].Town); t != nil {
		t.MonthlyCargoDelivered[cargoType] = boundAdd16(t.MonthlyCargoDelivered[cargoType], uint32(qty))
	}
}
