package game

import (
	"time"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

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
	CompanyName string
	PlayerMoney int64

	// CargoDelivered counts units delivered per cargo slot, for the
	// cargoDelivery objective (Objective type 3).
	CargoDelivered [32]uint32
	CurrentLoan    int64

	// Economy accumulators. MonthlyIncome/MonthlyExpenses are summed over the
	// current game month; at the month rollover Update snapshots
	// LastMonthProfit = income - expenses and resets both accumulators. The
	// monthly vehicle-profit objective (type 1) reads LastMonthProfit.
	MonthlyIncome   int64
	MonthlyExpenses int64
	LastMonthProfit int64

	// VehicleCount returns the number of company vehicles at month rollover,
	// used to charge the flat monthly running cost. Set by the game loop
	// (cmd/goloco loadScenario); nil in tests unless supplied.
	VehicleCount func() int

	// Time tracking
	GameDate    time.Time
	TickCount   uint64
	TicksPerDay uint64

	// World
	MapWidth  int
	MapHeight int
	MapTiles  [][]Tile

	// Entities
	Vehicles  []*Vehicle
	Stations  []*Station
	Buildings []*Building

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
	Type       TileType
	Height     uint8
	Owner      int // company ID, -1 for unowned
	BuildingID int // -1 for no building
}

// Vehicle represents a train, bus, ship, or plane
type Vehicle struct {
	ID       int
	Name     string
	Type     VehicleType
	OwnerID  int
	X, Y     float64 // pixel position
	TileX    int
	TileY    int
	Speed    float64
	RouteID  int
	RoutePos int
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
		CompanyName: "Player Transport Co.",
		PlayerMoney: 100000,
		CurrentLoan: 0,
		GameDate:    time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC),
		TickCount:   0,
		TicksPerDay: 74, // OpenLoco uses 74 ticks per day
		MapWidth:    width,
		MapHeight:   height,
		MapTiles:    make([][]Tile, height),
		Vehicles:    make([]*Vehicle, 0),
		Stations:    make([]*Station, 0),
		Buildings:   make([]*Building, 0),
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

// vehicleRunningCost is the flat monthly running cost charged per company
// vehicle. Simplified: upstream derives running cost from each vehicle
// object's runCostFactor scaled by a cost index and inflation.
// TODO: upstream reference src/OpenLoco/src/Vehicles/Vehicle.cpp
// (updateMonthly / VehicleObject::runCostFactor).
const vehicleRunningCost int64 = 120

// cargoFare returns the simplified flat fare paid per unit of the given cargo
// slot on delivery. Upstream computes payment from cargo type, distance and
// transit time (a payment curve); this is a flat per-cargo stand-in.
// TODO: upstream reference src/OpenLoco/src/Economy/ (calculateTransportPayment
// / CargoObject payment factors and the age/distance payment table).
func cargoFare(cargoSlot uint8) int64 {
	switch cargoSlot {
	case 0, 11: // passengers (PASS is slot 11 in the base game — see
		// world.CargoSlotPass; slot 0 kept for generic/test callers)
		return 6
	case 1: // mail
		return 8
	default:
		return 10
	}
}

// CreditDelivery pays the company for delivering amount units of the given
// cargo slot, crediting both PlayerMoney and the current month's income.
// Wired from World.OnCargoDelivered (cmd/goloco loadScenario).
func (gs *GameState) CreditDelivery(cargoSlot uint8, amount uint32) {
	income := cargoFare(cargoSlot) * int64(amount)
	gs.PlayerMoney += income
	gs.MonthlyIncome += income
}

// Update updates the game state for one tick
func (gs *GameState) Update() {
	gs.TickCount++

	// Update date, detecting a month rollover to run the monthly economy.
	if gs.TickCount%gs.TicksPerDay == 0 {
		prevMonth := gs.GameDate.Month()
		gs.GameDate = gs.GameDate.AddDate(0, 0, 1)
		if gs.GameDate.Month() != prevMonth {
			gs.onMonthRollover()
		}
	}

	// Update vehicles
	for _, v := range gs.Vehicles {
		v.Update()
	}
}

// onMonthRollover charges monthly vehicle running costs, snapshots the
// finished month's profit (income - expenses) into LastMonthProfit, and
// resets the accumulators for the new month.
func (gs *GameState) onMonthRollover() {
	n := 0
	if gs.VehicleCount != nil {
		n = gs.VehicleCount()
	}
	runningCost := int64(n) * vehicleRunningCost
	gs.MonthlyExpenses += runningCost
	gs.PlayerMoney -= runningCost

	gs.LastMonthProfit = gs.MonthlyIncome - gs.MonthlyExpenses
	gs.MonthlyIncome = 0
	gs.MonthlyExpenses = 0
}

// Update updates a vehicle's position
func (v *Vehicle) Update() {
	// Simple movement for now - move in a circle
	v.X += v.Speed
	if v.X > 500 {
		v.X = 0
	}
}

// LoadFromScenario populates game state from a loaded scenario
func (gs *GameState) LoadFromScenario(sc *scenario.Scenario) error {
	if sc == nil {
		return nil
	}

	// Update map dimensions
	gs.MapWidth = sc.MapWidth
	gs.MapHeight = sc.MapHeight

	// Convert scenario tiles to game tiles
	gs.MapTiles = make([][]Tile, gs.MapHeight)
	for y := 0; y < gs.MapHeight; y++ {
		gs.MapTiles[y] = make([]Tile, gs.MapWidth)
		for x := 0; x < gs.MapWidth; x++ {
			scTile := sc.GetTile(x, y)
			if scTile != nil {
				gs.MapTiles[y][x] = Tile{
					Type:       convertSurfaceType(scTile.Surface),
					Height:     scTile.Height,
					Owner:      -1,
					BuildingID: -1,
				}
			} else {
				gs.MapTiles[y][x] = Tile{
					Type:       TileGrass,
					Height:     0,
					Owner:      -1,
					BuildingID: -1,
				}
			}
		}
	}

	// Update scenario options if available
	if sc.Options.StartYear > 0 {
		gs.GameDate = time.Date(int(sc.Options.StartYear), 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if sc.Options.StartMoney > 0 {
		gs.PlayerMoney = sc.Options.StartMoney
	}

	return nil
}

// convertSurfaceType converts scenario surface type to game tile type
func convertSurfaceType(surface scenario.SurfaceType) TileType {
	switch surface {
	case scenario.SurfaceGrass:
		return TileGrass
	case scenario.SurfaceDirt, scenario.SurfaceSand:
		return TileDirt
	case scenario.SurfaceWater:
		return TileWater
	case scenario.SurfaceRock, scenario.SurfaceSnow:
		return TileDirt // Use dirt for now
	default:
		return TileGrass
	}
}

// NewGameStateFromScenario creates a game state from a scenario file
func NewGameStateFromScenario(filePath string) (*GameState, error) {
	sc, err := scenario.LoadScenarioData(filePath)
	if err != nil {
		return nil, err
	}

	gs := NewGameState(sc.MapWidth, sc.MapHeight)
	err = gs.LoadFromScenario(sc)
	if err != nil {
		return nil, err
	}

	return gs, nil
}
