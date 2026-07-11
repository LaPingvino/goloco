package world

import (
	"log"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// station.go — minimal station placement, passenger supply and the
// board→travel→deliver loop that increments delivered cargo.
//
// A station is attached to a track piece (its origin / SequenceIndex-0 element
// gains HasStation and a scenario.StationElement is appended to the tile so the
// existing paintStation renderer draws it). Placed stations are also recorded in
// a lightweight registry (SimStation) that accumulates waiting passengers from
// nearby town buildings; SimVehicles pick these up on arrival and deliver them
// when they next stop at a DIFFERENT station.
//
// OpenLoco reference: src/OpenLoco/src/Map/StationElement.h (element layout),
// src/OpenLoco/src/World/Station.cpp (cargo accumulation / delivery).

const (
	// CargoSlotPass is the cargo object slot used for passengers (PASS).
	// TODO: derive real cargo typing from the vehicle/station/cargo objects
	// instead of hard-coding the passenger slot (11 in the base game).
	CargoSlotPass uint8 = 11

	// stationSupplyRadius is the Chebyshev radius (in tiles) searched for town
	// buildings that supply a station with passengers.
	stationSupplyRadius = 4
	// passengerAccumPeriod is the number of ticks between supply increments.
	passengerAccumPeriod = 100
	// stationWaitingCap caps a station's waiting-passenger queue.
	stationWaitingCap uint32 = 200
	// stationDwellTicks is how long a vehicle halts at a station to load/unload.
	stationDwellTicks = 120
	// vehiclePassengerCapacity is how many passengers one vehicle can carry.
	vehiclePassengerCapacity uint32 = 50
)

// SimStation is one placed station in the passenger-supply registry.
type SimStation struct {
	id           int
	tileX, tileY int
	baseZ        uint8
	buildings    int    // nearby building count (fixed supply, computed once)
	waiting      uint32 // passengers waiting to board
	accum        int    // tick accumulator toward the next supply increment
}

// PlaceStationAtHead attaches a train station to the track piece under the
// construction head. Returns false when the head is inactive or its tile has no
// track origin element to carry the station.
func (w *World) PlaceStationAtHead(stationObjID uint8) bool {
	h := w.consHead
	if !h.Active {
		return false
	}
	return w.PlaceStationOnTile(h.X, h.Y, stationObjID)
}

// PlaceStationOnTile marks the SequenceIndex-0 track element on tile (x,y) as a
// station, appends a matching scenario.StationElement (so paintStation renders
// it) and registers the tile in the passenger-supply registry. Returns false if
// the tile has no track origin or already carries a station.
func (w *World) PlaceStationOnTile(x, y int, stationObjID uint8) bool {
	t := w.getTile(x, y)
	if t == nil {
		return false
	}
	var te *scenario.TrackElement
	for i := range t.tracks {
		if t.tracks[i].SequenceIndex == 0 {
			te = &t.tracks[i]
			break
		}
	}
	if te == nil || te.HasStation {
		return false
	}
	te.HasStation = true
	t.stations = append(t.stations, scenario.StationElement{
		StationType:   0, // train station
		ObjectID:      stationObjID,
		Rotation:      te.Rotation & 3,
		SequenceIndex: 0,
		StationID:     uint16(len(w.simStations)),
		Owner:         te.Owner,
		BaseZ:         te.BaseZ,
		ClearZ:        te.BaseZ + 8,
	})
	s := w.registerStation(x, y, te.BaseZ)
	log.Printf("[Station] placed at (%d,%d) z%d objSlot=%d nearbyBuildings=%d",
		x, y, te.BaseZ, stationObjID, s.buildings)
	return true
}

// registerStation adds (or returns the existing) SimStation for a tile, snapping
// its passenger supply from the surrounding town buildings.
func (w *World) registerStation(x, y int, baseZ uint8) *SimStation {
	for _, s := range w.simStations {
		if s.tileX == x && s.tileY == y && s.baseZ == baseZ {
			return s
		}
	}
	s := &SimStation{
		id:        len(w.simStations),
		tileX:     x,
		tileY:     y,
		baseZ:     baseZ,
		buildings: w.countNearbyBuildings(x, y, stationSupplyRadius),
	}
	w.simStations = append(w.simStations, s)
	return s
}

// countNearbyBuildings sums the building elements on tiles within a Chebyshev
// radius of (cx,cy) — a stand-in for a real catchment/town-population model.
func (w *World) countNearbyBuildings(cx, cy, radius int) int {
	n := 0
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if t := w.getTile(cx+dx, cy+dy); t != nil {
				n += len(t.buildings)
			}
		}
	}
	return n
}

// stationAt returns the SimStation registered at (x,y,baseZ), or nil.
func (w *World) stationAt(x, y int, baseZ uint8) *SimStation {
	for _, s := range w.simStations {
		if s.tileX == x && s.tileY == y && s.baseZ == baseZ {
			return s
		}
	}
	return nil
}

// tickStations accumulates waiting passengers at every station proportional to
// its nearby-building supply: roughly buildings/4 per 100 ticks, capped.
func (w *World) tickStations() {
	for _, s := range w.simStations {
		if s.buildings <= 0 {
			continue
		}
		s.accum++
		if s.accum < passengerAccumPeriod {
			continue
		}
		s.accum = 0
		add := uint32(s.buildings)
		if add == 0 {
			add = 1 // trickle when a few buildings are in range
		}
		s.waiting += add
		if s.waiting > stationWaitingCap {
			s.waiting = stationWaitingCap
		}
	}
}

// serviceStation handles a vehicle's arrival at a station: it first delivers any
// passengers that boarded at a DIFFERENT station (incrementing delivered cargo
// via OnCargoDelivered), then boards fresh passengers up to capacity.
func (w *World) serviceStation(sv *SimVehicle, st *SimStation) {
	if sv.boarded > 0 && sv.boardedStation != st.id {
		delivered := sv.boarded
		sv.boarded = 0
		sv.boardedStation = -1
		if w.OnCargoDelivered != nil {
			w.OnCargoDelivered(CargoSlotPass, delivered)
		}
		// Spawn a rising "+£N" income float at the delivering vehicle's tile.
		// The fare mirrors game.cargoFare(CargoSlotPass); see passengerFare.
		w.SpawnMoneyFloat(sv.tileX, sv.tileY, sv.baseZ, int64(delivered)*passengerFare)
	}
	space := vehiclePassengerCapacity - sv.boarded
	take := st.waiting
	if take > space {
		take = space
	}
	if take > 0 {
		st.waiting -= take
		sv.boarded += take
		sv.boardedStation = st.id
	}
}

// SeedStationWaiting adds waiting passengers to the station registered at
// (tileX,tileY) (any height). It is a debug/demo aid used by the headless
// GOLOCO_OPEN=track hook so the board→deliver loop fires without waiting for
// slow building-based accumulation. Returns false if no station is there.
func (w *World) SeedStationWaiting(tileX, tileY int, n uint32) bool {
	for _, s := range w.simStations {
		if s.tileX == tileX && s.tileY == tileY {
			s.waiting += n
			if s.waiting > stationWaitingCap {
				s.waiting = stationWaitingCap
			}
			return true
		}
	}
	return false
}

// FindTownsideRun locates a straight, flat, empty west-heading run of n tiles
// whose endpoints have town buildings nearby — a viable site for a passenger
// shuttle line. Returns the origin tile and ok.
func (w *World) FindTownsideRun(n int) (int, int, bool) {
	buildingsNear := func(x, y int) int {
		c := 0
		for dy := -4; dy <= 4; dy++ {
			for dx := -4; dx <= 4; dx++ {
				if t := w.getTile(x+dx, y+dy); t != nil && len(t.buildings) > 0 {
					c++
				}
			}
		}
		return c
	}
	for y := 2; y < w.height-2; y++ {
		for x := n + 1; x < w.width-2; x++ {
			ok := true
			var z uint8
			for i := 0; i < n; i++ {
				t := w.getTile(x-i, y)
				if t == nil || t.slope&0x1F != 0 || t.waterLevel > 0 ||
					len(t.buildings) > 0 || len(t.tracks) > 0 || len(t.roads) > 0 || t.isIndustrial {
					ok = false
					break
				}
				if i == 0 {
					z = t.baseZ
				} else if t.baseZ != z {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			if buildingsNear(x, y) >= 4 && buildingsNear(x-n+1, y) >= 1 {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}
