package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/objects"
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
