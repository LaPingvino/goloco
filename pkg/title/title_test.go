package title

import (
	"testing"
)

// TestRestartResetsState verifies that Restart() sets state to Running and
// resets counters to zero without re-scaling waypoints.
func TestRestartResetsState(t *testing.T) {
	s := NewSequence(nil)
	s.Start(64, 64)

	// Advance a few ticks so counters are non-zero
	for range 30 {
		s.Update()
	}
	if s.tickCounter == 0 || s.waypointTick == 0 {
		t.Skip("ticks did not advance; skipping restart test")
	}

	s.Restart()

	if s.state != StateRunning {
		t.Errorf("Restart: state = %v, want StateRunning", s.state)
	}
	if s.tickCounter != 0 {
		t.Errorf("Restart: tickCounter = %d, want 0", s.tickCounter)
	}
	if s.waypointIdx != 0 {
		t.Errorf("Restart: waypointIdx = %d, want 0", s.waypointIdx)
	}
	if s.waypointTick != 0 {
		t.Errorf("Restart: waypointTick = %d, want 0", s.waypointTick)
	}
}

// TestRestartDoesNotRescaleWaypoints verifies that calling Restart() after
// Start() does not scale the waypoints a second time.
// Concretely: starting with map 64×64, calling Start twice must not result
// in waypoints 4× larger than a single Start call would produce.
func TestRestartDoesNotRescaleWaypoints(t *testing.T) {
	mapW, mapH := 64, 64

	s1 := NewSequence(nil)
	s1.Start(mapW, mapH)
	afterStart := make([]Waypoint, len(s1.waypoints))
	copy(afterStart, s1.waypoints)

	// Restart should not mutate waypoints
	s1.Restart()
	for i, wp := range s1.waypoints {
		if wp.X != afterStart[i].X || wp.Y != afterStart[i].Y {
			t.Errorf("waypoint %d: after Restart X=%v Y=%v, want X=%v Y=%v (Restart re-scaled!)",
				i, wp.X, wp.Y, afterStart[i].X, afterStart[i].Y)
		}
	}
}

// TestRestartCameraAtFirstWaypoint verifies that camera is positioned at
// waypoint[0] after Restart().
func TestRestartCameraAtFirstWaypoint(t *testing.T) {
	s := NewSequence(nil)
	s.Start(64, 64)

	// Move camera away
	s.SetCameraPosition(9999, 9999, 5)
	s.Restart()

	if len(s.waypoints) == 0 {
		t.Skip("no waypoints")
	}
	wp0 := s.waypoints[0]
	x, y, zoom := s.GetCameraPosition()
	if x != wp0.X || y != wp0.Y || zoom != wp0.Zoom {
		t.Errorf("Restart: camera = (%.1f, %.1f, %.2f), want (%.1f, %.1f, %.2f)",
			x, y, zoom, wp0.X, wp0.Y, wp0.Zoom)
	}
}

// TestStartThenRestartEquivalent verifies that the camera position after
// Restart() matches what it would be immediately after Start().
func TestStartThenRestartEquivalent(t *testing.T) {
	mapW, mapH := 128, 128

	s1 := NewSequence(nil)
	s1.Start(mapW, mapH)
	x1, y1, z1 := s1.GetCameraPosition()

	s2 := NewSequence(nil)
	s2.Start(mapW, mapH)
	s2.Restart()
	x2, y2, z2 := s2.GetCameraPosition()

	if x1 != x2 || y1 != y2 || z1 != z2 {
		t.Errorf("Restart camera (%.1f, %.1f, %.2f) != fresh Start camera (%.1f, %.1f, %.2f)",
			x2, y2, z2, x1, y1, z1)
	}
}
