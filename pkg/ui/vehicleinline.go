package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/LaPingvino/goloco/pkg/objects"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// DrawVehicleInline is VehicleDraw.cpp drawVehicleInline for a vehicle object
// (no car): at loc + (0, 19 + yOffset), every bogie, then every body, each at
// its screen distance along x, in the colour scheme (primary, secondary).
// Returns the vehicle's screen length.
func DrawVehicleInline(rt *gfx8.RenderTarget, v *objects.VehicleObject, x, y, yOffset int, primary, secondary uint8) int {
	g1 := globalG1Font.g1
	y += 19 + yOffset
	items, total := v.InlineDrawItems(40, 0) // yaw 40
	pal, _ := g1.GetPaletteMap2(int(primary), int(secondary))
	for _, bodies := range [2]bool{false, true} {
		for _, it := range items {
			if it.IsBody != bodies {
				continue
			}
			id := int(v.ImageOffset + it.Image)
			if id <= 0 || id >= len(g1.Elements) {
				continue
			}
			if sp, err := gfx8.Decode(&g1.Elements[id]); err == nil {
				gfx8.DrawSpriteNoise(rt, sp, x+int(it.Dist), y, gfx8.OpFor(sp, pal != nil, false), pal, nil)
			}
		}
	}
	return int(total)
}

// DrawVehiclesInlineDiag draws vehicle objects one per 30 px row in a grey
// panel at (x, y) (diagnostics for DrawVehicleInline).
func DrawVehiclesInlineDiag(screen *ebiten.Image, x, y int, vs []*objects.VehicleObject, yOffsets []int) {
	h := 30*len(vs) + 4
	draw8(screen, x, y, 300, h, func(rt *gfx8.RenderTarget) {
		gfx8.FillRect(rt, 0, 0, 299, h-1, 10)
		for i, v := range vs {
			DrawVehicleInline(rt, v, 4, 2+30*i, yOffsets[i], 7, 2) // ColourScheme: blue / grey
		}
	})
}

// TrainCarDraw is one car for DrawTrainInline (world.TrainCar's fields).
type TrainCarDraw struct {
	Object     *objects.VehicleObject
	YOffset    int
	Components []objects.CarComponent
	Reversed   bool
	Brake      bool
	// The car tooltip's data: its front bogie's creationDay, refundCost and
	// reliability, and its rack rail's name.
	Built       uint32
	Value       int64
	Reliability uint16
	RackName    string
}

// carTooltip is the Vehicle window's car tooltip (Details / Cargo
// scrollMouseOver, buffer_337): the vehicle's name, then a line each — built
// (month and year), power (rail and road, when it has any), length in tiles
// (rail and road: getLength, two decimals), weight, top speed (and on its rack
// rail), value, reliability (when not 0).
func carTooltip(c TrainCarDraw) string {
	v := c.Object
	if v == nil {
		return ""
	}
	line := func(id int, args ...string) string {
		return "\n" + strings.TrimPrefix(FillArgs(i18n.Str(id), args...), "{NEWLINE}")
	}
	s := "{SMALLFONT}" + v.DisplayName // tooltip_stringid
	y, m, _ := L.CalcDate(c.Built)
	s += line(i18n.VehicleDetailsTooltipBuilt, i18n.Label(i18n.MonthLongJanuary+m)+" "+strconv.Itoa(y))
	landed := v.Mode == objects.TransportModeRail || v.Mode == objects.TransportModeRoad
	if v.Power != 0 && landed {
		s += line(i18n.VehicleDetailsTooltipPower, L.Power(int(v.Power)))
	}
	if landed {
		length := 0 // getLength: each car component's body sprite's halfLength × 2
		for i := 0; i < int(v.NumCarComponents) && i < len(v.CarComponents); i++ {
			if b := v.CarComponents[i].BodySpriteIdx; b != 0xFF {
				length += int(v.BodySprites[int(b)&(len(v.BodySprites)-1)].HalfLength) * 2
			}
		}
		n := length * 100 / 4 / 32 // internalLengthToComma1DP
		s += line(i18n.VehicleDetailsTooltipLength, fmt.Sprintf("%d.%02d", n/100, n%100))
	}
	s += line(i18n.VehicleDetailsTooltipWeight, strconv.Itoa(int(v.Weight)))
	s += line(i18n.VehicleDetailsTooltipMaxSpeed, L.Velocity(int(v.Speed)))
	if c.RackName != "" {
		s += FillArgs(i18n.Str(i18n.VehicleDetailsTooltipSpeedOnStringid), L.Velocity(int(v.RackSpeed)), c.RackName)
	}
	s += line(i18n.VehicleDetailsTooltipValue, currency32(c.Value))
	if c.Reliability != 0 {
		s += line(i18n.VehicleDetailsTooltipReliability, strconv.Itoa(int(c.Reliability)/256))
	}
	return s
}

// drawCarInline is drawVehicleInline for a Car (animated): y + 19 + the
// object's offset, the parts asked for (bogies or bodies) at their screen
// distances, each in its own colour scheme. Returns the car's screen length.
func drawCarInline(rt *gfx8.RenderTarget, c TrainCarDraw, x, y int, bodies bool) int {
	g1 := globalG1Font.g1
	y += 19 + c.YOffset
	items, total := c.Object.CarDrawItems(40, c.Components, c.Reversed, true, c.Brake)
	for _, it := range items {
		if it.IsBody != bodies {
			continue
		}
		id := int(c.Object.ImageOffset + it.Image)
		if id <= 0 || id >= len(g1.Elements) {
			continue
		}
		sp, err := gfx8.Decode(&g1.Elements[id])
		if err != nil {
			continue
		}
		pal, _ := g1.GetPaletteMap2(int(it.Colour[0]), int(it.Colour[1]))
		gfx8.DrawSpriteNoise(rt, sp, x+int(objects.ScreenDistance(it.Dist, 40)), y, gfx8.OpFor(sp, pal != nil, false), pal, nil)
	}
	return int(objects.ScreenDistance(total, 40))
}

// DrawTrainInline is VehicleDraw.cpp drawTrainInline: every car's bogies
// left to right, then from the start again every car's bodies. Returns the
// end x.
func DrawTrainInline(rt *gfx8.RenderTarget, cars []TrainCarDraw, x, y int) int {
	start := x
	for _, c := range cars {
		x += drawCarInline(rt, c, x, y, false)
	}
	x = start
	for _, c := range cars {
		x += drawCarInline(rt, c, x, y, true)
	}
	return x
}
