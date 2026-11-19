package sim

import (
	"reflect"
	"testing"
)

// onToolUpdateAirport / onToolUpdateDock's map selection: the station's tiles
// (getClampedRange: clamped to the map, x fastest, both ends included) with
// enableConstruct, unk_03 off, and the arrow at the placement; reset, both off.
func TestStationToolSelection(t *testing.T) {
	s := &State{Map: Map{Width: 10, Height: 10}}
	s.ConsSelection.Unk03 = true
	x0, y0, x1, y1 := DockToolTiles(3*32+5, 4*32+7)
	if x0 != 3 || y0 != 4 || x1 != 4 || y1 != 5 {
		t.Fatalf("dock tiles %d,%d … %d,%d", x0, y0, x1, y1)
	}
	s.SetStationToolSelection(x0, y0, x1, y1, ConstructionArrow{X: 96, Y: 128, Z: 16, Direction: 2})
	want := [][2]int32{{96, 128}, {128, 128}, {96, 160}, {128, 160}}
	if !reflect.DeepEqual(s.ConsSelection.Tiles, want) || !s.ConsSelection.On || s.ConsSelection.Unk03 {
		t.Errorf("dock selection %+v", s.ConsSelection)
	}
	if a := s.ConsArrow; !a.On || a.X != 96 || a.Y != 128 || a.Z != 16 || a.Direction != 2 {
		t.Errorf("arrow %+v", a)
	}
	s.SetStationToolSelection(-2, 8, 0, 12, ConstructionArrow{})
	want = [][2]int32{{0, 256}, {0, 288}}
	if !reflect.DeepEqual(s.ConsSelection.Tiles, want) {
		t.Errorf("clamped selection %v", s.ConsSelection.Tiles)
	}
	s.ClearStationToolSelection()
	if s.ConsSelection.On || s.ConsArrow.On {
		t.Error("reset: still on")
	}
}
