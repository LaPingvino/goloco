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

	// Camera position and animation
	cameraX    float64
	cameraY    float64
	zoom       float64
	targetX    float64
	targetY    float64
	targetZoom float64

	// Waypoints for camera path
	waypoints    []Waypoint
	waypointIdx  int
	waypointTick uint32

	// Map dimensions for bounds checking
	mapWidth  int
	mapHeight int
}

// Waypoint defines a camera position and duration
type Waypoint struct {
	X        float64
	Y        float64
	Zoom     float64
	Duration uint32 // ticks to reach this point
	Wait     uint32 // ticks to wait at this point
}

// DefaultWaypoints provides a default camera path for the title sequence
var DefaultWaypoints = []Waypoint{
	{X: 200, Y: 200, Zoom: 1.0, Duration: 0, Wait: 180},
	{X: 400, Y: 300, Zoom: 1.2, Duration: 300, Wait: 180},
	{X: 600, Y: 200, Zoom: 0.8, Duration: 300, Wait: 180},
	{X: 300, Y: 500, Zoom: 1.0, Duration: 300, Wait: 180},
	{X: 500, Y: 400, Zoom: 1.5, Duration: 300, Wait: 180},
	{X: 200, Y: 200, Zoom: 1.0, Duration: 300, Wait: 60}, // Loop back
}

// NewSequence creates a new title sequence
func NewSequence(audioMgr *audio.Manager) *Sequence {
	return &Sequence{
		state:      StateNone,
		audioMgr:   audioMgr,
		cameraX:    200,
		cameraY:    200,
		zoom:       1.0,
		targetX:    200,
		targetY:    200,
		targetZoom: 1.0,
		waypoints:  append([]Waypoint(nil), DefaultWaypoints...), // copy so scaleWaypoints doesn't mutate the global
	}
}

// Start begins the title sequence
func (s *Sequence) Start(mapWidth, mapHeight int) {
	s.state = StateBegin
	s.tickCounter = 0
	s.mapWidth = mapWidth
	s.mapHeight = mapHeight

	// Scale waypoints to map size
	s.scaleWaypoints()

	// Set initial position
	if len(s.waypoints) > 0 {
		s.cameraX = s.waypoints[0].X
		s.cameraY = s.waypoints[0].Y
		s.zoom = s.waypoints[0].Zoom
		s.targetX = s.cameraX
		s.targetY = s.cameraY
		s.targetZoom = s.zoom
	}

	s.waypointIdx = 0
	s.waypointTick = 0
	s.state = StateRunning
}

// scaleWaypoints adjusts waypoint positions based on map size
func (s *Sequence) scaleWaypoints() {
	// Scale waypoints to fit the actual map
	scaleX := float64(s.mapWidth) * 16 / 800  // 800 was the reference width
	scaleY := float64(s.mapHeight) * 16 / 600 // 600 was the reference height

	for i := range s.waypoints {
		s.waypoints[i].X *= scaleX
		s.waypoints[i].Y *= scaleY
	}
}

// Update advances the title sequence by one tick
func (s *Sequence) Update() {
	if s.state != StateRunning {
		return
	}

	s.tickCounter++
	s.waypointTick++

	if len(s.waypoints) == 0 {
		return
	}

	wp := &s.waypoints[s.waypointIdx]
	totalDuration := wp.Duration + wp.Wait

	if s.waypointTick >= totalDuration {
		// Move to next waypoint
		s.waypointIdx = (s.waypointIdx + 1) % len(s.waypoints)
		s.waypointTick = 0

		// Set new target
		nextWP := &s.waypoints[s.waypointIdx]
		s.targetX = nextWP.X
		s.targetY = nextWP.Y
		s.targetZoom = nextWP.Zoom
	} else if s.waypointTick < wp.Duration && wp.Duration > 0 {
		// Interpolate towards target
		progress := float64(s.waypointTick) / float64(wp.Duration)
		progress = easeInOutQuad(progress)

		// Smooth interpolation
		s.cameraX += (s.targetX - s.cameraX) * 0.02
		s.cameraY += (s.targetY - s.cameraY) * 0.02
		s.zoom += (s.targetZoom - s.zoom) * 0.02
	}
	// During wait period, camera stays at current position
}

// easeInOutQuad provides smooth easing for camera movement
func easeInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - (-2*t+2)*(-2*t+2)/2
}

// GetCameraPosition returns the current camera position and zoom
func (s *Sequence) GetCameraPosition() (x, y, zoom float64) {
	return s.cameraX, s.cameraY, s.zoom
}

// SetCameraPosition allows external control of camera
func (s *Sequence) SetCameraPosition(x, y, zoom float64) {
	s.cameraX = x
	s.cameraY = y
	s.zoom = zoom
	s.targetX = x
	s.targetY = y
	s.targetZoom = zoom
}

// Stop ends the title sequence
func (s *Sequence) Stop() {
	s.state = StateNone
}

// Restart resumes the title sequence from waypoint 0 without re-scaling
// the waypoints (safe to call after Start has already been called once).
func (s *Sequence) Restart() {
	s.tickCounter = 0
	s.waypointIdx = 0
	s.waypointTick = 0
	if len(s.waypoints) > 0 {
		s.cameraX = s.waypoints[0].X
		s.cameraY = s.waypoints[0].Y
		s.zoom = s.waypoints[0].Zoom
		s.targetX = s.cameraX
		s.targetY = s.cameraY
		s.targetZoom = s.zoom
	}
	s.state = StateRunning
}

// IsRunning returns true if the title sequence is active
func (s *Sequence) IsRunning() bool {
	return s.state == StateRunning
}

// GetState returns the current state
func (s *Sequence) GetState() State {
	return s.state
}
