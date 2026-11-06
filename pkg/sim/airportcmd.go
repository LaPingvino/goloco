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
