package game

import "time"

// TileType represents different types of terrain
type TileType uint8

const (
	TileGrass TileType = iota
	TileDirt
	TileWater
	TileRoad
	TileRail
	TileBuilding
	TileStation
)

// GameState holds the overall game state
type GameState struct {
	// Player company
	CompanyName  string
	PlayerMoney  int64
	CurrentLoan  int64

	// Time tracking
	GameDate     time.Time
	TickCount    uint64
	TicksPerDay  uint64

	// World
	MapWidth     int
	MapHeight    int
	MapTiles     [][]Tile

	// Entities
	Vehicles     []*Vehicle
	Stations     []*Station
	Buildings    []*Building

	// UI state
	SelectedTile *TilePos
	CameraX      float64
	CameraY      float64
}

// TilePos represents a tile position
type TilePos struct {
	X, Y int
}

// Tile represents a single map tile
type Tile struct {
	Type      TileType
	Height    uint8
	Owner     int // company ID, -1 for unowned
	BuildingID int // -1 for no building
}

// Vehicle represents a train, bus, ship, or plane
type Vehicle struct {
	ID        int
	Name      string
	Type      VehicleType
	OwnerID   int
	X, Y      float64 // pixel position
	TileX     int
	TileY     int
	Speed     float64
	RouteID   int
	RoutePos  int
}

type VehicleType uint8

const (
	VehicleTypeTrain VehicleType = iota
	VehicleTypeBus
	VehicleTypeShip
	VehicleTypePlane
)

// Station represents a station
type Station struct {
	ID      int
	Name    string
	TileX   int
	TileY   int
	Type    StationType
	OwnerID int
}

type StationType uint8

const (
	StationTypeRail StationType = iota
	StationTypeRoad
	StationTypeAirport
	StationTypeDock
)

// Building represents a building
type Building struct {
	ID      int
	Name    string
	TileX   int
	TileY   int
	Type    BuildingType
	OwnerID int
}

type BuildingType uint8

const (
	BuildingTypeDepot BuildingType = iota
	BuildingTypeWarehouse
	BuildingTypeFactory
	BuildingTypeTown
)

// NewGameState creates a new game state with default values
func NewGameState(width, height int) *GameState {
	gs := &GameState{
		CompanyName:  "Player Transport Co.",
		PlayerMoney:  100000,
		CurrentLoan:  0,
		GameDate:     time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC),
		TickCount:    0,
		TicksPerDay:  74, // OpenLoco uses 74 ticks per day
		MapWidth:     width,
		MapHeight:    height,
		MapTiles:     make([][]Tile, height),
		Vehicles:     make([]*Vehicle, 0),
		Stations:     make([]*Station, 0),
		Buildings:    make([]*Building, 0),
	}

	// Initialize map tiles
	for y := 0; y < height; y++ {
		gs.MapTiles[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			// Create varied terrain
			tileType := TileGrass
			if (x+y)%7 == 0 {
				tileType = TileDirt
			} else if x < 3 || y < 3 {
				tileType = TileWater
			}

			gs.MapTiles[y][x] = Tile{
				Type:       tileType,
				Height:     0,
				Owner:      -1,
				BuildingID: -1,
			}
		}
	}

	return gs
}

// Update updates the game state for one tick
func (gs *GameState) Update() {
	gs.TickCount++

	// Update date
	if gs.TickCount%gs.TicksPerDay == 0 {
		gs.GameDate = gs.GameDate.AddDate(0, 0, 1)
	}

	// Update vehicles
	for _, v := range gs.Vehicles {
		v.Update()
	}
}

// Update updates a vehicle's position
func (v *Vehicle) Update() {
	// Simple movement for now - move in a circle
	v.X += v.Speed
	if v.X > 500 {
		v.X = 0
	}
}
