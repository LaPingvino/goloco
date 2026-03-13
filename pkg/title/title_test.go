package title

import (
	"testing"
)

func TestRestartResetsState(t *testing.T) {
	s := NewSequence(nil)
	s.Start(512, 512)

	for range 30 {
		s.Update()
	}
	if s.tickCounter == 0 || s.waypointTick == 0 {
		t.Skip("ticks did not advance")
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

func TestRestartDoesNotMutateWaypoints(t *testing.T) {
	s := NewSequence(nil)
	s.Start(512, 512)
	before := make([]Waypoint, len(s.waypoints))
	copy(before, s.waypoints)

	s.Restart()
	for i, wp := range s.waypoints {
		if wp != before[i] {
			t.Errorf("waypoint %d changed after Restart", i)
		}
	}
}

func TestRestartCameraAtFirstWaypoint(t *testing.T) {
	s := NewSequence(nil)
	s.Start(512, 512)

	s.SetCameraPosition(9999, 9999)
	s.Restart()

	if len(s.waypoints) == 0 {
		t.Skip("no waypoints")
	}
	wp0 := s.waypoints[0]
	x, y := s.GetCameraPosition()
	wantX, wantY := float64(wp0.TileX), float64(wp0.TileY)
	if x != wantX || y != wantY {
		t.Errorf("camera = (%.1f, %.1f), want (%.1f, %.1f)", x, y, wantX, wantY)
	}
}

func TestStartThenRestartEquivalent(t *testing.T) {
	s1 := NewSequence(nil)
	s1.Start(512, 512)
	x1, y1 := s1.GetCameraPosition()

	s2 := NewSequence(nil)
	s2.Start(512, 512)
	s2.Restart()
	x2, y2 := s2.GetCameraPosition()

	if x1 != x2 || y1 != y2 {
		t.Errorf("after Restart (%.1f,%.1f) != after Start (%.1f,%.1f)", x2, y2, x1, y1)
	}
}
