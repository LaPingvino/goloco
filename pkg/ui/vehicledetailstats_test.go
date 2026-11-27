package ui

import "testing"

// Details::draw's three lines: power and weight for rail and road, weight
// alone (the power popped) for air and water; the rack rail speed only when
// the train uses rack rail; the car count for a rail vehicle of more cars;
// a null speed as 0 and reliability 0 as 64.
func TestVehicleDetailsStatsLines(t *testing.T) {
	for _, c := range []struct {
		s    VehicleDetailsStats
		want [3]string
	}{
		{VehicleDetailsStats{Mode: 0, TotalPower: 1200, TotalWeight: 152, MaxSpeed: 60, RackMaxSpeed: 60, Reliability: 87, TotalLength: 3 * 128, CarCount: 3},
			[3]string{"{COLOUR WINDOW_2}Total power: {COLOUR BLACK}1,200hp{COLOUR WINDOW_2} Weight: {COLOUR BLACK}152t",
				"{COLOUR WINDOW_2}Max. speed: {COLOUR BLACK}60mph{COLOUR WINDOW_2} Reliability: {COLOUR BLACK}87%",
				"{COLOUR WINDOW_2}Length: {COLOUR BLACK}3.00 tile(s), 3 cars"}},
		{VehicleDetailsStats{Mode: 0, TotalPower: 800, TotalWeight: 60, MaxSpeed: 50, RackMaxSpeed: 20, Reliability: 0, RackMods: true, TotalLength: 96, CarCount: 1},
			[3]string{"{COLOUR WINDOW_2}Total power: {COLOUR BLACK}800hp{COLOUR WINDOW_2} Weight: {COLOUR BLACK}60t",
				"{COLOUR WINDOW_2}Max. speed: {COLOUR BLACK}50mph / 20mph{COLOUR WINDOW_2} Reliability: {COLOUR BLACK}64%",
				"{COLOUR WINDOW_2}Length: {COLOUR BLACK}0.75 tile(s)"}},
		{VehicleDetailsStats{Mode: 2, TotalPower: 4000, TotalWeight: 30, MaxSpeed: -1, RackMaxSpeed: 300, Reliability: 99, TotalLength: 128, CarCount: 2},
			[3]string{"{COLOUR WINDOW_2}Weight: {COLOUR BLACK}30t",
				"{COLOUR WINDOW_2}Max. speed: {COLOUR BLACK}0mph{COLOUR WINDOW_2} Reliability: {COLOUR BLACK}99%",
				"{COLOUR WINDOW_2}Length: {COLOUR BLACK}1.00 tile(s)"}},
	} {
		if got := c.s.Lines(); got != c.want {
			t.Errorf("%+v:\n got %q\nwant %q", c.s, got, c.want)
		}
	}
}
