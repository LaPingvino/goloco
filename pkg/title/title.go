package title

import (
	"github.com/LaPingvino/goloco/pkg/audio"
)

// State represents the current state of the title sequence
type State int

const (
	StateNone State = iota
	StateBegin
	StateLoadScenario
	StateRunning
	StateRestart
)

// Sequence manages the title screen state and camera animation
type Sequence struct {
	state       State
	tickCounter uint32
	audioMgr    *audio.Manager

	// Camera position in tile coordinates
	cameraX float64
	cameraY float64

	// Current waypoint state
	waypoints    []Waypoint
	waypointIdx  int
	waypointTick uint32

	// Map dimensions (tiles) — unused for tile-coord waypoints but kept for API compat
	mapWidth  int
	mapHeight int
}

// Waypoint defines a camera cut position in tile coordinates and how many
// ticks to show it before cutting to the next.
// Taken directly from OpenLoco Title.cpp _titleSequence (MoveStep + WaitStep pairs).
// RotateStep and ReloadStep are skipped (rotation not yet supported).
// OpenLoco reference: src/OpenLoco/src/Title.cpp
type Waypoint struct {
	TileX uint16
	TileY uint16
	Wait  uint32 // ticks to show this position (~40 ticks/sec in OpenLoco)
}

// defaultWaypoints is the exact OpenLoco title sequence camera positions.
var defaultWaypoints = []Waypoint{
	{TileX: 231, TileY: 160, Wait: 368},
	{TileX: 93, TileY: 59, Wait: 370},
	{TileX: 314, TileY: 230, Wait: 370},
	{TileX: 83, TileY: 314, Wait: 370},
	{TileX: 224, TileY: 253, Wait: 375},
	{TileX: 210, TileY: 322, Wait: 375},
	{TileX: 125, TileY: 338, Wait: 375},
	{TileX: 72, TileY: 138, Wait: 375},
	{TileX: 178, TileY: 158, Wait: 375},
	{TileX: 203, TileY: 316, Wait: 375},
}

// NewSequence creates a new title sequence.
func NewSequence(audioMgr *audio.Manager) *Sequence {
	return &Sequence{
		state:     StateNone,
		audioMgr:  audioMgr,
		waypoints: append([]Waypoint(nil), defaultWaypoints...),
	}
}

// Start begins the title sequence for a map of the given tile dimensions.
func (s *Sequence) Start(mapWidth, mapHeight int) {
	s.mapWidth = mapWidth
	s.mapHeight = mapHeight
	s.waypointIdx = 0
	s.waypointTick = 0
	if len(s.waypoints) > 0 {
		s.cameraX = float64(s.waypoints[0].TileX)
		s.cameraY = float64(s.waypoints[0].TileY)
	} else {
		s.cameraX = float64(mapWidth) / 2
		s.cameraY = float64(mapHeight) / 2
	}
	s.state = StateRunning
}

// Update advances the title sequence by one tick.
// The camera cuts instantly to each waypoint after Wait ticks have elapsed.
// OpenLoco reference: src/OpenLoco/src/Title.cpp update()
func (s *Sequence) Update() {
	if s.state != StateRunning || len(s.waypoints) == 0 {
		return
	}

	s.tickCounter++
	s.waypointTick++

	wp := &s.waypoints[s.waypointIdx]
	if s.waypointTick >= wp.Wait {
		s.waypointIdx = (s.waypointIdx + 1) % len(s.waypoints)
		s.waypointTick = 0
		next := &s.waypoints[s.waypointIdx]
		s.cameraX = float64(next.TileX)
		s.cameraY = float64(next.TileY)
	}
}

// GetCameraPosition returns the current camera tile position.
func (s *Sequence) GetCameraPosition() (tileX, tileY float64) {
	return s.cameraX, s.cameraY
}

// SetCameraPosition allows external override of the camera.
func (s *Sequence) SetCameraPosition(tileX, tileY float64) {
	s.cameraX = tileX
	s.cameraY = tileY
}

// Stop ends the title sequence.
func (s *Sequence) Stop() {
	s.state = StateNone
}

// Restart resumes the title sequence from waypoint 0.
func (s *Sequence) Restart() {
	s.tickCounter = 0
	s.waypointIdx = 0
	s.waypointTick = 0
	if len(s.waypoints) > 0 {
		s.cameraX = float64(s.waypoints[0].TileX)
		s.cameraY = float64(s.waypoints[0].TileY)
	}
	s.state = StateRunning
}

// IsRunning returns true if the title sequence is active.
func (s *Sequence) IsRunning() bool {
	return s.state == StateRunning
}

// GetState returns the current state.
func (s *Sequence) GetState() State {
	return s.state
}
