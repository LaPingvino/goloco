package construct

import "testing"

// Cases read off getTrackPieceId's branches.
func TestGetTrackPieceID(t *testing.T) {
	cases := []struct {
		piece, gradient, rotation uint8
		ok                        bool
		id, rot                   uint8
	}{
		{Straight, Level, 1, true, 0, 1},
		{Straight, Level, 13, true, 1, 13},   // diagonal keeps its rotation
		{Straight, SlopeUp, 13, false, 0, 0}, // no diagonal slopes
		{Straight, SteepSlopeUp, 5, true, 34, 1},
		{Straight, SlopeUp, 5, false, 0, 0},
		{Straight, SteepSlopeDown, 2, true, 17, 2},
		{LeftCurveVerySmall, Level, 9, true, 29, 1},
		{RightCurveVerySmall, Level, 6, true, 30, 2},
		{RightCurveVerySmall, Level, 12, false, 0, 0},
		{LeftCurveSmall, SlopeDown, 3, true, 20, 3},
		{RightCurveSmall, SteepSlopeUp, 0, true, 23, 0},
		{LeftCurve, Level, 4, false, 0, 0},
		{LeftCurveLarge, Level, 14, true, 10, 14},
		{RightCurveLarge, Level, 2, true, 9, 2},
		{RightCurveLarge, Level, 6, false, 0, 0},
		{SBendLeft, Level, 0, true, 12, 0},
		{SBendLeft, Level, 9, true, 33, 1},
		{SBendLeft, Level, 5, false, 0, 0},
		{SBendRight, Level, 5, true, 32, 1},
		{SBendRight, Level, 8, false, 0, 0},
		{SBendToDualTrack, Level, 1, true, 38, 1},
		{SBendToSingleTrack, Level, 10, true, 41, 2},
		{Turnaround, Level, 3, true, 42, 3},
		{Turnaround, Level, 8, true, 43, 0},
		{Turnaround, SlopeUp, 0, false, 0, 0},
		{PieceNull, Level, 0, false, 0, 0},
	}
	for _, c := range cases {
		got, ok := GetTrackPieceID(c.piece, c.gradient, c.rotation)
		if ok != c.ok || (ok && (got.ID != c.id || got.Rotation != c.rot)) {
			t.Errorf("piece %d gradient %d rotation %d: %v %v, want %v {%d %d}", c.piece, c.gradient, c.rotation, got, ok, c.ok, c.id, c.rot)
		}
	}
}

// Cases read off getRoadPieceId's branches.
func TestGetRoadPieceID(t *testing.T) {
	cases := []struct {
		piece, gradient, rotation uint8
		ok                        bool
		id, rot                   uint8
	}{
		{Straight, Level, 2, true, 0, 2},
		{Straight, SteepSlopeDown, 1, true, 8, 1},
		{Straight, Level, 5, false, 0, 0},
		{LeftCurveVerySmall, Level, 3, true, 1, 3},
		{RightCurveSmall, Level, 0, true, 4, 0},
		{RightCurveSmall, SlopeUp, 0, false, 0, 0},
		{LeftCurve, Level, 0, false, 0, 0},
		{SBendLeft, Level, 0, false, 0, 0},
		{Turnaround, Level, 6, true, 9, 2},
		{Turnaround, Level, 12, false, 0, 0},
	}
	for _, c := range cases {
		got, ok := GetRoadPieceID(c.piece, c.gradient, c.rotation)
		if ok != c.ok || (ok && (got.ID != c.id || got.Rotation != c.rot)) {
			t.Errorf("piece %d gradient %d rotation %d: %v %v, want %v {%d %d}", c.piece, c.gradient, c.rotation, got, ok, c.ok, c.id, c.rot)
		}
	}
}
