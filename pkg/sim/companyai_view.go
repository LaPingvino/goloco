package sim

import (
	"fmt"

	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// AIView is what one company's AI sees and does during a think: the
// observations (copies; nothing returned aliases the game's state), its
// PRNG and log, and Query / Do for its game commands (companyai_actions.go).
// It is valid only during the Think it is passed to.
type AIView struct {
	m  *Map
	om *objects.ObjectManager
	id uint8
	st *companyAI
}

// Self is the company the AI runs.
func (v *AIView) Self() uint8 { return v.id }

// Ticks is the scenario tick count.
func (v *AIView) Ticks() uint32 { return v.m.ScenarioTicks }

// Day is the current day (days since 1 January 1800, getCurrentDay) and
// Year the current year.
func (v *AIView) Day() uint32  { return v.m.CurrentDay }
func (v *AIView) Year() uint16 { return v.m.CurrentYear }

// Rng is the AI's own PRNG (Locomotion's generator, seeded per company):
// never the game's gPrng1, so an AI's choices leave the game's random
// sequence alone (AI-INTERFACE.md O22).
func (v *AIView) Rng() *core.Prng { return &v.st.rng }

// Logf writes a note in the AI's decision log.
func (v *AIView) Logf(format string, args ...any) {
	v.st.record(AIDecision{Tick: v.m.ScenarioTicks, Day: v.m.CurrentDay, Text: fmt.Sprintf(format, args...)})
}

// MapSize is the map's size in tiles.
func (v *AIView) MapSize() (w, h int) { return v.m.Width, v.m.Height }

// AISurface is a tile's surface: base height (SmallZ), slope corners,
// water level (0 none).
type AISurface struct {
	BaseZ, Slope, WaterLevel uint8
}

// Surface is the tile (x, y)'s surface; ok false off the map.
func (v *AIView) Surface(x, y int) (AISurface, bool) {
	t := v.m.TileAt(x, y)
	if t == nil {
		return AISurface{}, false
	}
	return AISurface{t.BaseZ, t.Slope, t.WaterLevel}, true
}

// AIRoad is a road element as the map shows it.
type AIRoad struct {
	RoadID, Rotation, BaseZ, SequenceIndex uint8
	RoadObjectID, Owner, Mods              uint8
	HasStation, HasSignal, LevelCrossing   bool
	HasBridge                              bool
}

// Roads lists the road elements on tile (x, y); ghosts and AI-allocated
// reservations are not shown.
func (v *AIView) Roads(x, y int) []AIRoad {
	t := v.m.TileAt(x, y)
	if t == nil {
		return nil
	}
	var out []AIRoad
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad {
			continue
		}
		el := t.Roads[r.Index]
		if el.Ghost || el.AiAllocated {
			continue
		}
		out = append(out, AIRoad{el.RoadID, el.Rotation, el.BaseZ, el.SequenceIndex, el.RoadObjectID, el.Owner, el.Mods,
			el.HasStationElement, el.HasSignalElement, el.HasLevelCrossing, el.HasBridge})
	}
	return out
}

// RoadConnections is getRoadConnections for the AI's company: the road
// pieces (trackAndDirection) at (x, y, z) (world units) a vehicle arriving
// with nextRotation continues onto, roadObjectID 0xFF for any road every
// vehicle may use; and the road station there (-1 none). Pass each
// connection to RoadConnectionEnd for where it leads.
func (v *AIView) RoadConnections(x, y, z int32, nextRotation, roadObjectID uint8) ([]uint16, int) {
	rc := v.m.RoadConnections(x, y, z, nextRotation, v.id, roadObjectID, 0, 0)
	return rc.Connections, rc.StationID
}

// AITown is a town as the town list shows it.
type AITown struct {
	ID         int
	Name       string
	X, Y       int // tiles
	Population uint32
}

// Towns lists the towns.
func (v *AIView) Towns() []AITown {
	var out []AITown
	for i, t := range v.m.Towns {
		if t.Name == 0xFFFF {
			continue
		}
		out = append(out, AITown{i, v.m.TownName(t.Name, i), int(t.X) / 32, int(t.Y) / 32, t.Population})
	}
	return out
}

// AIIndustry is an industry as its window shows it.
type AIIndustry struct {
	ID, ObjectID      int
	X, Y              int // tiles
	Town              int
	ProducedPrevMonth [2]uint16
}

// Industries lists the industries.
func (v *AIView) Industries() []AIIndustry {
	var out []AIIndustry
	for i, in := range v.m.Industries {
		if in.Name == 0xFFFF {
			continue
		}
		out = append(out, AIIndustry{i, int(in.ObjectID), int(in.X) / 32, int(in.Y) / 32, int(in.Town), in.ProducedPrevMonth})
	}
	return out
}

// AIStation is a station as its window shows it.
type AIStation struct {
	ID, Owner int
	Name      string
	X, Y      int // tiles
	Town      int
	// Waiting is the cargo waiting by cargo type.
	Waiting [32]uint16
}

// Stations lists every company's stations.
func (v *AIView) Stations() []AIStation {
	var out []AIStation
	for i, s := range v.m.StationNames {
		if s.Name == 0 {
			continue
		}
		st := AIStation{ID: i, Owner: int(s.Owner), Name: v.m.StationName(i), X: int(s.X) / 32, Y: int(s.Y) / 32, Town: int(s.Town)}
		for c := range s.Cargo {
			st.Waiting[c] = s.Cargo[c].Quantity
		}
		out = append(out, st)
	}
	return out
}

// AICompany is the AI's own company, as its Company window shows it.
type AICompany struct {
	Cash             int64
	Loan, MaxLoan    int32
	Bankrupt         bool
	LoanInterestRate uint8
}

// Company is the AI's own company.
func (v *AIView) Company() AICompany {
	c := v.m.Companies[v.id]
	return AICompany{Cash: c.Cash, Loan: c.CurrentLoan, MaxLoan: v.m.InflationAdjustedCost(int16(v.m.MaxLoanSize), 0, 8),
		Bankrupt: v.m.CompanyFaces[v.id].ChallengeFlags&companyFlagBankrupt != 0, LoanInterestRate: v.m.LoanInterestRate}
}

// AIVehicle is one of the AI's own vehicles.
type AIVehicle struct {
	Head        uint16
	Mode        uint8 // TransportMode: 0 rail, 1 road, 2 air, 3 water
	VehicleType uint8
	Placed      bool
	Status      uint8
	StationID   uint16 // the station it is at, 0xFFFF none
	Cars        int
}

// Vehicles lists the AI's own vehicles.
func (v *AIView) Vehicles() []AIVehicle {
	trains, _ := v.m.Trains()
	var out []AIVehicle
	for _, tr := range trains {
		h := v.m.Entities[tr.Head]
		if h.Owner != v.id {
			continue
		}
		out = append(out, AIVehicle{h.ID, h.Mode, h.VehicleType, h.TileX != -1, h.Status, h.StationID, len(tr.Cars)})
	}
	return out
}

// AIVehicleObject is a vehicle the AI may buy, as the Build Vehicle window
// lists it.
type AIVehicleObject struct {
	ID          uint16
	Name        string
	Power       uint16
	Speed       uint16
	Cargo       uint32 // a bit per cargo type it carries
	Capacity    uint8
	DesignedYr  uint16
	ObsoleteYr  uint16
	CostFactor  int16
	CostIndex   uint8
	RoadOrTrack uint8 // its road (bit 7 set) or track type
}

// BuildableVehicles is the Build Vehicle window's list for the AI's company:
// vehicleType (0 train, 1 bus, 2 truck, 3 tram, 4 aircraft, 5 ship) on the
// tab trackType (0xFF all; a road object with bit 7; a track object),
// unlocked ones only.
func (v *AIView) BuildableVehicles(vehicleType, trackType uint8) []AIVehicleObject {
	if v.om == nil {
		return nil
	}
	o := DefaultBuildVehicleOptions(v.id)
	o.Filter = BuildFilterPowered | BuildFilterUnpowered | BuildFilterUnlocked
	var out []AIVehicleObject
	for _, id := range v.m.BuildableVehicles(v.om, vehicleType, trackType, o) {
		obj := v.om.Vehicles[id]
		ci := v.om.CargoFor(obj)
		var cargo uint32
		for k := range int(ci.NumSimultaneousCargoTypes) {
			cargo |= ci.CompatibleCargoCategories[k] // a bit per cargo type, despite the name (CargoFor)
		}
		rt := v.om.TrackTypeFor(obj)
		if obj.Mode == objects.TransportModeRoad && rt != 0xFF {
			rt |= 1 << 7
		}
		out = append(out, AIVehicleObject{ID: id, Name: obj.DisplayName, Power: obj.Power, Speed: obj.Speed, Cargo: cargo,
			Capacity: ci.MaxCargo[0], DesignedYr: obj.Designed, ObsoleteYr: obj.Obsolete, CostFactor: obj.CostFactor,
			CostIndex: obj.CostIndex, RoadOrTrack: rt})
	}
	return out
}

// AvailableRoads is companyGetAvailableRoads for the AI's company: the road
// object ids (bit 7 set) its vehicles may use, as the toolbar's road menu.
func (v *AIView) AvailableRoads() []uint8 { return v.m.CompanyAvailableRoads(v.id, v.om) }

// AIRoadStationObject is a road station type the AI may build.
type AIRoadStationObject struct {
	ID    uint8
	Name  string
	Cargo uint32 // the cargo types it handles (RoadStationCargoFilter)
}

// RoadStations is getAvailableCompatibleStations for a road object (without
// bit 7) this year: the road stop types the construction window offers.
func (v *AIView) RoadStations(roadObjectID uint8) []AIRoadStationObject {
	if v.om == nil {
		return nil
	}
	var out []AIRoadStationObject
	for _, id := range v.om.AvailableCompatibleStations(int(roadObjectID), true, int(v.m.CurrentYear)) {
		if o := v.om.RoadStationObjects[id]; o != nil {
			out = append(out, AIRoadStationObject{uint8(id), o.Name, RoadStationCargoFilter(v.om, id)})
		}
	}
	return out
}

// PassengerCargo is IndustryManager::getMostCommonBuildingCargoType: the
// cargo towns' buildings make most (passengers, in practice; AI-INTERFACE.md
// O4).
func (v *AIView) PassengerCargo() uint8 { return mostCommonBuildingCargoType(v.om) }
