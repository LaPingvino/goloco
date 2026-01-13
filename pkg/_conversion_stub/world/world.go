// Package world - stub implementations for world data structures
package world

// Pos2 is a 2D position
type Pos2 struct {
	X, Y int16
}

// Pos3 is a 3D position
type Pos3 struct {
	X, Y, Z int16
}

// Company represents a company in the game
type Company struct {
	Name        string
	OwnerName   string
	Cash        uint32
	CurrentLoan uint32
}

// Station represents a train/bus/airport station
type Station struct {
	Name string
}

// Vehicle represents a train, bus, ship, or airplane
type Vehicle struct {
	Name string
}

// TownManager manages all towns in the world
type TownManager struct{}

// StationManager manages all stations
type StationManager struct{}

// VehicleManager manages all vehicles
type VehicleManager struct{}

// CompanyManager manages all companies
type CompanyManager struct{}
