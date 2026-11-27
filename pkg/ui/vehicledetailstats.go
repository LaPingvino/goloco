package ui

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
)

// The Vehicle window's Details tab text under the car list (Ui/Windows/
// Vehicle.cpp Details::draw 0x004B3542, from "Draw power and weight" on):
// three lines kVehicleDetailsLineHeight (12) apart, the first at
// height − kVehicleDetailsTextHeight + kVehicleDetailsOffset.

// VehicleDetailsStats is what Details::draw reads from the train: Vehicle2's
// mode, totalPower, totalWeight, maxSpeed, rackRailMaxSpeed and reliability,
// Vehicle1's var_49 (the rack rail mods in use), the head's
// getVehicleTotalLength and getCarCount.
type VehicleDetailsStats struct {
	Mode                   int // TransportMode: 0 rail, 1 road, 2 air, 3 water
	TotalPower             int
	TotalWeight            int
	MaxSpeed, RackMaxSpeed int16 // speed16, -1 (kSpeed16Null) shown as 0
	Reliability            int   // Vehicle2's reliability, 0 shown as 64
	RackMods               bool  // veh1->var_49 != 0
	TotalLength, CarCount  int
}

// Lines are the three strings with their arguments filled in (FillArgs):
// vehicle_details_total_power_and_weight (rail and road; else
// vehicle_details_weight, which pops the power), vehicle_details_max_speed_
// and_reliability (…_and_rack_rail_… when the train uses rack rail; the plain
// one pops the rack speed), vehicle_length (vehicle_car_count_and_length for
// a rail vehicle of more than one car).
func (s VehicleDetailsStats) Lines() [3]string {
	speed := func(v int16) string {
		if v == -1 {
			v = 0
		}
		return L.Velocity(int(v))
	}
	var out [3]string
	weight := strconv.Itoa(s.TotalWeight)
	if s.Mode == 0 || s.Mode == 1 {
		out[0] = FillArgs(i18n.Str(i18n.VehicleDetailsTotalPowerAndWeight), L.Power(s.TotalPower), weight)
	} else {
		out[0] = FillArgs(i18n.Str(i18n.VehicleDetailsWeight), "", weight) // {POP16}: the power
	}
	rel := s.Reliability
	if rel == 0 {
		rel = 64
	}
	if s.RackMods {
		out[1] = FillArgs(i18n.Str(i18n.VehicleDetailsMaxSpeedAndRackRailAndReliability), speed(s.MaxSpeed), speed(s.RackMaxSpeed), strconv.Itoa(rel))
	} else {
		out[1] = FillArgs(i18n.Str(i18n.VehicleDetailsMaxSpeedAndReliability), speed(s.MaxSpeed), "", strconv.Itoa(rel)) // {POP16}: the rack speed
	}
	if s.Mode == 0 && s.CarCount > 1 {
		out[2] = FillArgs(i18n.Str(i18n.VehicleCarCountAndLength), length2DP(s.TotalLength), strconv.Itoa(s.CarCount))
	} else {
		out[2] = FillArgs(i18n.Str(i18n.VehicleLength), length2DP(s.TotalLength))
	}
	return out
}

// drawCodedLeftClipped is drawStringLeftClipped (TextRenderer.cpp) in medium
// bold for a string with colour codes, starting in black: clipString's cut
// over the visible characters (the codes have no width; "..." after the
// longest prefix that stays narrower than width, nothing under 6 px).
func drawCodedLeftClipped(screen *ebiten.Image, s string, x, y, width int, win2, black color.Color) {
	const font = G1FontMediumBold
	runes := []rune(codedRunes(s))
	visible := func(rs []rune) string {
		var v []rune
		for _, r := range rs {
			if r >= 32 {
				v = append(v, r)
			}
		}
		return string(v)
	}
	if width < 6 {
		return
	}
	if MeasureTextFont(visible(runes), font) > width {
		cut := []rune(nil)
		for i := 1; i < len(runes); i++ {
			try := append(append([]rune{}, runes[:i]...), []rune("...")...)
			if MeasureTextFont(visible(try), font) >= width {
				break
			}
			cut = try
		}
		runes = cut
	}
	px, col := x, black
	var seg []rune
	flush := func() {
		if len(seg) > 0 {
			px += DrawTextFont(screen, string(seg), font, px, y, col)
			seg = seg[:0]
		}
	}
	for _, r := range runes {
		switch r {
		case codeWindow2:
			flush()
			col = win2
		case codeBlack:
			flush()
			col = black
		default:
			if r >= 32 {
				seg = append(seg, r)
			}
		}
	}
	flush()
}
