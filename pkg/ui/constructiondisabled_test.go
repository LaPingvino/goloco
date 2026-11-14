package ui

import "testing"

// setDisabledWidgets' rules: straight and level on a plain track, nothing;
// on a slope the curves go (small curves stay with sloped curves); a curve
// chosen takes the slopes (a small curve only those its track can't build);
// beside a dual track's other track and diagonally, the rotation's own set.
func TestConstructionDisabled(t *testing.T) {
	cases := []struct {
		st      ConstructionTabState
		off, on []string
	}{
		{ConstructionTabState{Piece: "straight", Gradient: "level"}, nil, []string{"left_hand_curve", "slope_up", "s_bend_left"}},
		{ConstructionTabState{Piece: "straight", Gradient: "slope_up"}, []string{"left_hand_curve", "s_bend_left", "left_hand_curve_small"}, []string{"slope_down"}},
		{ConstructionTabState{Piece: "straight", Gradient: "slope_up", Traits: TraitSlopedCurve}, []string{"left_hand_curve"}, []string{"left_hand_curve_small"}},
		{ConstructionTabState{Piece: "left_hand_curve", Gradient: "level"}, []string{"slope_up", "steep_slope_down"}, []string{"right_hand_curve"}},
		{ConstructionTabState{Piece: "left_hand_curve_small", Gradient: "level", Traits: TraitSlope}, []string{"steep_slope_up"}, []string{"slope_up"}},
		{ConstructionTabState{IsRoad: true, Piece: "left_hand_curve", Gradient: "level"}, nil, []string{"slope_up"}},
		{ConstructionTabState{IsRoad: true, Piece: "turnaround", Gradient: "level"}, []string{"slope_up"}, nil},
		{ConstructionTabState{Piece: "straight", Gradient: "level", Rotation: 9}, []string{"s_bend_right", "left_hand_curve", "slope_up"}, []string{"s_bend_left", "steep_slope_up"}},
		{ConstructionTabState{Piece: "straight", Gradient: "level", Rotation: 13}, []string{"left_hand_curve_very_small", "steep_slope_up", "s_bend_left"}, []string{"left_hand_curve_large"}},
	}
	for i, c := range cases {
		d := c.st.Disabled()
		for _, n := range c.off {
			if !d[n] {
				t.Errorf("case %d: %s should be disabled", i, n)
			}
		}
		for _, n := range c.on {
			if d[n] {
				t.Errorf("case %d: %s should be enabled", i, n)
			}
		}
	}
}
