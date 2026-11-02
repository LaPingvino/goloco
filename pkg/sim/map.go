package sim

import (
	"github.com/LaPingvino/goloco/pkg/core"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Map is the loaded map: its size, tiles, towns and the vehicle entities saved
// in the scenario (OpenLoco: Map/TileManager for tiles, Towns/TownManager,
// Entities/EntityManager — state only, no drawing).
type Map struct {
	Width, Height                                          int
	Tiles                                                  [][]Tile        // [y][x]
	Towns                                                  []scenario.Town // parsed town records
	Entities                                               []scenario.VehicleEntity
	VehicleListData                                        []scenario.VehicleListEntry // heads + veh1/veh2 data for the vehicle list
	CompanyFaces                                           []scenario.CompanyFace      // company slot → competitor, emotion, performance
	Orders                                                 []byte                      // GameState orders[]
	Routings                                               []uint16                    // GameState routings[], by RoutingHandle (routing.go)
	RoadAnyTypeCompatible                                  uint32                      // GameState roadObjectIdIsAnyRoadTypeCompatible
	RoadUsableByAll                                        uint32                      // GameState roadObjectIdIsUsableByAllCompanies
	Rng1                                                   core.Prng                   // GameState rng (gPrng1), from the save
	TitleMode                                              bool                        // SceneManager::isTitleMode (pending breakdowns don\'t happen)
	ScenarioTicks                                          uint32                      // GameState scenarioTicks: +1 per game tick (UpdateVehicles)
	Companies                                              []scenario.CompanyEconomy   // the save\'s companies (economy)
	Industries                                             []scenario.Industry         // the save\'s industries (deliveries)
	vehicleMangled113623B                                  uint8                       // VehicleHead.cpp module static _vehicleMangled_113623B (sub_4ACEF1)
	CurrentDay                                             uint32                      // GameState general.currentDay
	CompanyNames                                           []uint16                    // company slot → name StringId
	ControllingCompany                                     uint8                       // playerCompanies[0] (getControllingId)
	TrafficHandedness                                      uint8                       // GameState trafficHandedness
	UnlockedVehicles                                       [][7]uint32                 // per company: Company::unlockedVehicles
	CurrentYear                                            uint16                      // GameState general.currentYear
	CurrentMonth                                           uint8                       // GameState general.currentMonth
	CurrentDayOfMonth                                      uint8                       // GameState general.currentDayOfMonth
	DayCounter                                             uint16                      // GameState general.dayCounter
	PlayerCompanies                                        [2]uint8                    // GameState playerCompanies
	LoanInterestRate                                       uint8                       // GameState loanInterestRate
	MaxLoanSize                                            uint16                      // GameState maxLoanSize
	StartingLoanSize                                       uint16                      // GameState startingLoanSize
	SeaLevel, LastLandOption                               uint8                       // GameState seaLevel, lastLandOption
	CurrentSnowLine, CurrentSeason                         uint8                       // GameState currentSnowLine, currentSeason
	PreferredAI                                            [3]uint8                    // GameState preferredAI intelligence / aggressiveness / competitiveness
	DefaultLevelCrossingType                               uint8                       // GameState currentDefaultLevelCrossingType
	ForbiddenVehiclesPlayers, ForbiddenVehiclesCompetitors uint16
	Objective                                              scenario.Objective // the scenario objective
	ObjectiveUntilYear                                     uint16             // scenarioObjectiveProgress.timeLimitUntilYear
	ObjectiveMonthsInChallenge                             uint16             // scenarioObjectiveProgress.monthsInChallenge
	ObjectiveCompletedInMonths                             uint16             // scenarioObjectiveProgress.completedChallengeInMonths
	// CurrencyMultiplicationFactor prices cost indices (Economy).
	CurrencyMultiplicationFactor [32]uint32

	IndustryObjectIDs []uint8                // industry index → industry object (GameState industries)
	CompanyColours    []uint8                // owner → main colour (CompanyManager::updateColours); nil if none
	StationOwners     []uint8                // station id → owner (Station::owner); nil if none
	UserStrings       []string               // GameState userStrings (StringIds 0x8000+); town names
	StationNames      []scenario.StationName // station id → name StringId + town
}

// TileAt returns the tile at (x, y), or nil outside the map.
func (m *Map) TileAt(x, y int) *Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return nil
	}
	return &m.Tiles[y][x]
}

// AddTrack appends a track element to tile (x, y).
// Returns false if the tile is out of bounds or an identical element already exists.
func (m *Map) AddTrack(x, y int, te scenario.TrackElement) bool {
	t := m.TileAt(x, y)
	if t == nil {
		return false
	}
	for _, ex := range t.Tracks {
		if ex.TrackObjectID == te.TrackObjectID &&
			ex.TrackID == te.TrackID &&
			ex.Rotation == te.Rotation &&
			ex.BaseZ == te.BaseZ {
			return false
		}
	}
	t.Tracks = append(t.Tracks, te)
	t.insertRef(scenario.KindTrack, len(t.Tracks)-1, te.BaseZ)
	return true
}

// AddRoad appends a road element to tile (x, y).
// Returns false if the tile is out of bounds or an identical element already exists.
func (m *Map) AddRoad(x, y int, re scenario.RoadElement) bool {
	t := m.TileAt(x, y)
	if t == nil {
		return false
	}
	for _, ex := range t.Roads {
		if ex.RoadObjectID == re.RoadObjectID &&
			ex.RoadID == re.RoadID &&
			ex.Rotation == re.Rotation &&
			ex.BaseZ == re.BaseZ {
			return false
		}
	}
	t.Roads = append(t.Roads, re)
	t.insertRef(scenario.KindRoad, len(t.Roads)-1, re.BaseZ)
	return true
}

// LoadScenario fills the map from a loaded scenario or save: size, every
// tile's surface and elements, towns and the saved vehicle entities.
func (m *Map) LoadScenario(sc *scenario.Scenario) {
	m.Width, m.Height = sc.MapWidth, sc.MapHeight
	m.Tiles = make([][]Tile, m.Height)
	for y := 0; y < m.Height; y++ {
		m.Tiles[y] = make([]Tile, m.Width)
		for x := 0; x < m.Width; x++ {
			st := sc.GetTile(x, y)
			if st == nil {
				continue
			}
			m.Tiles[y][x] = Tile{
				TerrainIndex: st.TerrainIndex,
				BaseZ:        st.Height,
				Slope:        st.Slope,
				WaterLevel:   st.Water,
				GrowthStage:  st.GrowthStage,
				Variation:    st.Variation,
				IsIndustrial: st.IsIndustrial,
				Trees:        st.Trees,
				Buildings:    st.Buildings,
				Walls:        st.Walls,
				Tracks:       st.Tracks,
				Roads:        st.Roads,
				Stations:     st.Stations,
				Signals:      st.Signals,
				Industries:   st.Industries,
				Order:        st.Order,
			}
		}
	}
	m.Towns = sc.Towns
	m.Entities = sc.Entities
	m.VehicleListData = sc.VehicleList
	m.Orders = sc.Orders
	m.Routings = sc.Routings
	m.RoadAnyTypeCompatible, m.RoadUsableByAll = sc.RoadAnyTypeCompatible, sc.RoadUsableByAll
	m.Rng1 = *core.NewPrngSeed(sc.Rng[0], sc.Rng[1])
	m.ScenarioTicks = sc.ScenarioTicks
	m.Companies = sc.Companies
	m.Industries = sc.Industries
	m.Companies = sc.Companies
	m.CurrentDay = sc.CurrentDay
	m.CompanyNames = sc.CompanyNames
	m.CompanyFaces = sc.CompanyFaces
	m.ControllingCompany = sc.ControllingCompany
	m.TrafficHandedness = sc.TrafficHandedness
	m.UnlockedVehicles = sc.UnlockedVehicles
	m.CurrentYear = sc.CurrentYear
	m.CurrentMonth = sc.CurrentMonth
	m.CurrentDayOfMonth = sc.CurrentDayOfMonth
	m.DayCounter = sc.DayCounter
	m.PlayerCompanies, m.LoanInterestRate, m.MaxLoanSize = sc.PlayerCompanies, sc.LoanInterestRate, sc.MaxLoanSize
	m.DefaultLevelCrossingType = sc.DefaultLevelCrossingType
	m.StartingLoanSize = sc.StartingLoanSize
	m.SeaLevel, m.LastLandOption = sc.SeaLevel, sc.LastLandOption
	m.CurrentSnowLine, m.CurrentSeason = sc.CurrentSnowLine, sc.CurrentSeason
	m.PreferredAI = sc.PreferredAI
	m.ForbiddenVehiclesPlayers, m.ForbiddenVehiclesCompetitors = sc.ForbiddenVehiclesPlayers, sc.ForbiddenVehiclesCompetitors
	m.Objective = sc.Objective
	m.ObjectiveUntilYear = sc.ObjectiveTimeLimitUntilYear
	m.ObjectiveMonthsInChallenge = sc.ObjectiveMonthsInChallenge
	m.CurrencyMultiplicationFactor = sc.CurrencyMultiplicationFactor
	m.IndustryObjectIDs = sc.IndustryObjectIDs
	m.CompanyColours = sc.CompanyColours
	m.StationOwners = sc.StationOwners
	m.UserStrings = sc.UserStrings
	m.StationNames = sc.StationNames
}
