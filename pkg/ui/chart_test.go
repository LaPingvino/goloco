package ui

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// drawGraph: front-to-back data walks from dataEnd − 1 − dataStart back; the
// values are shifted to fit the canvas; axes give a label per labelled tick
// and per y step; the line's pixels land where the walk puts them.
func TestChart(t *testing.T) {
	sec := make([]uint8, 256)
	for i := range sec {
		sec[i] = uint8(i)
	}
	gs := &Chart{Left: 0, Top: 0, Width: 200, Height: 117, YOffset: 17, XOffset: 40, YAxisLabelIncrement: 20,
		Lines: [][]int64{{400, 300, 200, 100}}, DataStart: []int{0}, LineColour: []uint8{77}, DataEnd: 4,
		XAxisRange: 1950 * 12, XAxisStepSize: 1, XAxisTickIncrement: 10, XAxisLabelIncrement: 12,
		Flags: ChartDataFrontToBack, PointFlags: ChartDrawLines}
	if v := gs.value(0, 0); v != 100 {
		t.Errorf("first value walked: %d, want the oldest-first 100", v)
	}
	if m := gs.maxValue(); m != 400 {
		t.Errorf("max %d", m)
	}
	rt := gfx8.NewRenderTarget(200, 130)
	labels := gs.Draw(rt, sec)
	if gs.numValueShifts != 2 { // 400 → 200 → 100 ≤ 100 (canvas height)
		t.Errorf("shifts %d", gs.numValueShifts)
	}
	var ys int
	for _, l := range labels {
		if !l.XAxis {
			ys++
		}
	}
	if ys != 5 { // 0, 20, 40, 60, 80 (100 ≥ the canvas height stops it)
		t.Errorf("%d y labels", ys)
	}
	// the first point: x = 40 + 0, y = 117 − (100 >> 2) − 17
	if got := rt.Bits[(117-25-17)*200+40]; got != 77 {
		t.Errorf("first point colour %d", got)
	}
}
