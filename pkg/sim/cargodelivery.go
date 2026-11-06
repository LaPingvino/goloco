package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Cargo from producers to stations (item CP): StationManager.cpp's
// findStationsForCargoType, deliverCargoToStations and
// deliverCargoToNearbyStations (0x0042F2FE) — what town buildings and
// industries hand their production to.

// stationCargoAcceptedFromProducer is StationCargoStatsFlags::acceptedFromProducer.
const stationCargoAcceptedFromProducer uint8 = 1 << 1

// stationShare is a found station and its rating for the cargo.
type stationShare struct {
	id     uint16
	rating uint8
}

// findStationsForCargoType is the stations whose catchment covers the
// producer — the tiles from 4 before (x, y)'s tile over its size plus 8 —
// that take this cargo from producers: each station once, in the order its
// elements are met (tile by tile, in each tile's stored order), not ghost or
// AI-allocated, at most 16.
func (m *Map) findStationsForCargoType(cargoType uint8, x, y int32, sizeX, sizeY int) []stationShare {
	tx0, ty0 := int(x)>>5-4, int(y)>>5-4
	var found []stationShare
	for oy := 0; oy < sizeY+8; oy++ {
		for ox := 0; ox < sizeX+8; ox++ {
			tx, ty := tx0+ox, ty0+oy
			if tx < 0 || ty < 0 || tx >= m.Width || ty >= m.Height { // validCoords
				continue
			}
			t := m.TileAt(tx, ty)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) {
					continue
				}
				el := &t.Stations[r.Index]
				if el.AiAllocated || el.Ghost {
					continue
				}
				if len(found) > 15 {
					break
				}
				dup := false
				for _, f := range found {
					if f.id == el.StationID {
						dup = true
						break
					}
				}
				if dup || int(el.StationID) >= len(m.StationNames) || int(cargoType) >= len(m.StationNames[el.StationID].Cargo) {
					continue
				}
				cs := &m.StationNames[el.StationID].Cargo[cargoType]
				if cs.Flags&stationCargoAcceptedFromProducer == 0 {
					continue
				}
				found = append(found, stationShare{el.StationID, cs.Rating})
			}
		}
	}
	return found
}

// deliverCargoToStations shares the quantity among the stations by rating
// squared, each share no more than rating / 256 of it, one more over a
// rating of 66; the total delivered (at most the quantity).
func (m *Map) deliverCargoToStations(found []stationShare, cargoType, qty uint8, om *objects.ObjectManager) uint16 {
	total := 0
	for _, f := range found {
		total += int(f.rating) * int(f.rating)
	}
	if total == 0 {
		return 0
	}
	delivered := 0
	for _, f := range found {
		r := int(f.rating)
		share := min(r*r*int(qty)/total, r*int(qty)/256)
		if r > 66 {
			share++
		}
		delivered += share
		m.DeliverCargoToStation(int(f.id), cargoType, uint8(share), om)
	}
	return uint16(min(delivered, int(qty)))
}

// DeliverCargoToNearbyStations is deliverCargoToNearbyStations: a producer
// of size × size tiles at (x, y) hands qty of the cargo to the stations
// around it; how much they took.
func (m *Map) DeliverCargoToNearbyStations(cargoType, qty uint8, x, y int32, sizeX, sizeY int, om *objects.ObjectManager) uint16 {
	found := m.findStationsForCargoType(cargoType, x, y, sizeX, sizeY)
	if len(found) == 0 {
		return 0
	}
	return m.deliverCargoToStations(found, cargoType, qty, om)
}
