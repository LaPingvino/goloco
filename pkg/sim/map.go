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
}

// TileAt returns the tile at (x, y), or nil outside the map.
func (m *Map) TileAt(x, y int) *Tile {
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return nil
	}
	return &m.Tiles[y][x]
}
