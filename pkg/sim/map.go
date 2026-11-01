package sim

import "github.com/LaPingvino/goloco/pkg/scenario"

// Map is the loaded map: its size, tiles, towns and the vehicle entities saved
// in the scenario (OpenLoco: Map/TileManager for tiles, Towns/TownManager,
// Entities/EntityManager — state only, no drawing).
type Map struct {
	Width, Height int
	Tiles         [][]Tile                 // [y][x]
	Towns         []scenario.Town          // parsed town records
	Entities      []scenario.VehicleEntity // real vehicle entities from the scenario / saved game

	IndustryObjectIDs []uint8 // industry index → industry object (GameState industries)
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
			}
		}
	}
	m.Towns = sc.Towns
	m.Entities = sc.Entities
	m.IndustryObjectIDs = sc.IndustryObjectIDs
}
