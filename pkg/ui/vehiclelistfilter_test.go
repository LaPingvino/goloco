package ui

import "testing"

// The vehicle list's filter types (af4d69c): all vehicles keeps everything;
// transports cargo looks at what the cars carry, has cargo order at the
// orders; without a cargo chosen neither filters.
func TestVehicleListFilterModes(t *testing.T) {
	v := VehicleRow{Carried: []int{3}, Cargos: []int{5}}
	for _, c := range []struct {
		mode, cargo int
		want        bool
	}{
		{0, 3, true}, {1, -1, true}, {1, 3, true}, {1, 5, false}, {2, 5, true}, {2, 3, false},
	} {
		if got := vlKeeps(v, c.mode, c.cargo); got != c.want {
			t.Errorf("mode %d cargo %d: %v, want %v", c.mode, c.cargo, got, c.want)
		}
	}
}
