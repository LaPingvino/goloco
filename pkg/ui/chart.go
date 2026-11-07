package ui

import (
	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// Ui/Chart.cpp: drawGraph (0x004CF824) and its parts, for the company list's
// graph tabs and the like. A line's data is its values in order (what the
// yData pointer walks); with ChartDataFrontToBack the walk starts at
// DataEnd − 1 − DataStart and goes back, as all charts but cargo delivery do.

// GraphFlags.
const (
	ChartShowNegativeValues = 1 << 0
	ChartDataFrontToBack    = 1 << 1
	ChartHideAxesAndLabels  = 1 << 2
)

// GraphPointFlags.
const (
	ChartDrawLines  = 1 << 0
	ChartDrawPoints = 1 << 1
)

// Chart is GraphSettings (positions in the render target's coordinates).
type Chart struct {
	Left, Top, Width, Height int
	YOffset, XOffset         int
	YAxisLabelIncrement      int
	Lines                    [][]int64 // yData, one per line
	DataStart                []int
	LinesToExclude           uint32
	LineColour               []uint8
	DataEnd                  int
	XAxisRange               int
	XAxisStepSize            int
	XAxisTickIncrement       int
	XAxisLabelIncrement      int
	YAxisStepSize            int
	Flags                    int
	PointFlags               int

	canvasLeft, canvasBottom, canvasHeight int
	numValueShifts                         uint
}

// ChartLabel is a label the chart asks for (graph_label_format with the
// x or y label's string): its anchor, centred (x axis) or right-aligned
// (y axis), and the value (x: the axis value; y: the scaled value).
type ChartLabel struct {
	X, Y  int
	XAxis bool
	Value int64
}

// value is the walk's value at dataIndex (graphGetValueFromPointer).
func (gs *Chart) value(line, dataIndex int) int64 {
	d := gs.Lines[line]
	i := dataIndex
	if gs.Flags&ChartDataFrontToBack != 0 {
		i = gs.DataEnd - 1 - dataIndex
	}
	if i < 0 || i >= len(d) {
		return 0
	}
	return d[i]
}

func (gs *Chart) start(line int) int {
	if gs.Flags&ChartDataFrontToBack != 0 && line < len(gs.DataStart) {
		return gs.DataStart[line]
	}
	return 0
}

// maxValue is graphGetMaxValue (0x004CF869): the largest |value| drawn.
func (gs *Chart) maxValue() int64 {
	var m int64
	for l := range gs.Lines {
		for i := gs.start(l); i < gs.DataEnd; i++ {
			v := gs.value(l, i)
			if v < 0 {
				v = -v
			}
			m = max(m, v)
		}
	}
	return m
}

// Draw is drawGraph: the canvas from the settings, the value shift that fits
// the largest value into the height, the axes and their labels (unless
// hidden), then the lines and / or points of each line not excluded. sec is
// the window's secondary colour map. It returns the labels to draw.
func (gs *Chart) Draw(rt *gfx8.RenderTarget, sec []uint8) []ChartLabel {
	gs.canvasLeft = gs.XOffset + gs.Left
	gs.canvasHeight = gs.Height - gs.YOffset
	gs.canvasBottom = gs.Top + gs.Height - gs.YOffset
	height := int64(gs.canvasHeight)
	if gs.Flags&ChartShowNegativeValues != 0 {
		height >>= 1
	}
	gs.numValueShifts = 0
	for v := gs.maxValue(); v > height; v >>= 1 {
		gs.numValueShifts++
	}
	var labels []ChartLabel
	if gs.Flags&ChartHideAxesAndLabels == 0 {
		labels = gs.drawAxesAndLabels(rt, sec)
	}
	for j := 0; j < 2; j++ {
		pf := 1 << j
		if gs.PointFlags&pf == 0 {
			continue
		}
		for i := range gs.Lines {
			if gs.LinesToExclude&(1<<i) != 0 {
				continue
			}
			gs.drawLineSegments(rt, i, pf)
		}
	}
	return labels
}

// drawAxesAndLabels is graphDrawAxesAndLabels (0x004CFA49).
func (gs *Chart) drawAxesAndLabels(rt *gfx8.RenderTarget, sec []uint8) []ChartLabel {
	var labels []ChartLabel
	xv := gs.XAxisRange
	if gs.Flags&ChartDataFrontToBack != 0 {
		xv -= (gs.DataEnd - 1) * gs.XAxisStepSize
	}
	for t := 0; t < gs.DataEnd; t++ {
		rem := 1
		if gs.XAxisLabelIncrement != 0 {
			rem = xv % gs.XAxisLabelIncrement
		}
		x := t*gs.XAxisTickIncrement + gs.Left + gs.XOffset
		h, shade := gs.canvasHeight, 4
		if rem == 0 {
			h, shade = h+3, 6
		}
		gfx8.FillRect(rt, x, gs.Top, x, gs.Top+h-1, gfx8.Shade(sec, shade))
		if rem == 0 {
			labels = append(labels, ChartLabel{X: x, Y: gs.canvasBottom + 5, XAxis: true, Value: int64(xv)})
		}
		xv += gs.XAxisStepSize
	}
	for y := 0; ; {
		x, w := gs.Left+gs.XOffset-2, gs.Width-gs.XOffset+3
		yy := -y + gs.canvasHeight + gs.Top
		if gs.Flags&ChartShowNegativeValues != 0 {
			yy -= gs.canvasHeight / 2
		}
		gfx8.FillRect(rt, x, yy, x+w-1, yy, gfx8.Shade(sec, 6))
		if gs.Flags&ChartShowNegativeValues != 0 {
			yn := y + gs.Top + gs.canvasHeight/2
			gfx8.FillRect(rt, x, yn, x+w-1, yn, gfx8.Shade(sec, 6))
		}
		ly := -y + gs.canvasHeight + gs.Top - 5
		if gs.Flags&ChartShowNegativeValues != 0 {
			ly -= gs.canvasHeight / 2
		}
		v := int64(y) << gs.numValueShifts
		labels = append(labels, ChartLabel{X: gs.Left + gs.XOffset - 3, Y: ly, Value: v})
		if gs.Flags&ChartShowNegativeValues != 0 {
			labels = append(labels, ChartLabel{X: gs.Left + gs.XOffset - 3, Y: y + gs.Top + gs.canvasHeight/2 - 5, Value: -v})
		}
		y += gs.YAxisLabelIncrement
		q := y
		if gs.Flags&ChartShowNegativeValues != 0 {
			q *= 2
		}
		if q >= gs.canvasHeight || gs.YAxisLabelIncrement <= 0 {
			break
		}
	}
	return labels
}

// drawLineSegments is drawGraphLineSegments (0x004CFD87): each value a
// pixel joined to the last (lines) or a 2 × 2 point.
func (gs *Chart) drawLineSegments(rt *gfx8.RenderTarget, line, pf int) {
	px, py := -1, 0
	c := uint8(0)
	if line < len(gs.LineColour) {
		c = gs.LineColour[line]
	}
	for i := gs.start(line); i < gs.DataEnd; i++ {
		v := gs.value(line, i) >> gs.numValueShifts
		x := gs.canvasLeft + i*gs.XAxisTickIncrement
		y := gs.Height - int(v) - gs.YOffset
		if gs.Flags&ChartShowNegativeValues != 0 {
			y -= gs.canvasHeight / 2
		}
		y += gs.Top
		switch pf {
		case ChartDrawLines:
			gfx8.FillRect(rt, x, y, x, y, c)
			if px != -1 {
				chartDrawLine(rt, px, py, x, y, c)
			}
			px, py = x, y
		case ChartDrawPoints:
			gfx8.FillRect(rt, x, y, x+1, y+1, c)
		}
	}
}

// chartDrawLine is SoftwareDrawingContext's drawLine: Bresenham, steep lines
// a pixel at a time, others in horizontal runs — with upstream's swap for a
// steep line (b.y ↔ a.x, a.y ↔ b.x) kept as it is.
func chartDrawLine(rt *gfx8.RenderTarget, ax, ay, bx, by int, c uint8) {
	abs := func(v int) int {
		if v < 0 {
			return -v
		}
		return v
	}
	steep := abs(ay-by) > abs(ax-bx)
	if steep {
		by, ax = ax, by
		ay, bx = bx, ay
	}
	if ax > bx {
		ax, bx = bx, ax
		by, ay = ay, by
	}
	dx, dy := bx-ax, abs(by-ay)
	e := dx / 2
	ys := -1
	if ay < by {
		ys = 1
	}
	y := ay
	for x, xs, n := ax, ax, 1; x < bx; x, n = x+1, n+1 {
		if steep {
			gfx8.FillRect(rt, y, x, y, x, c)
		}
		e -= dy
		if e < 0 {
			if !steep {
				gfx8.FillRect(rt, xs, y, xs+n-1, y, c)
			}
			xs, n = x+1, 0
			y += ys
			e += dx
		}
		if x+1 == bx && !steep {
			gfx8.FillRect(rt, xs, y, xs+n-1, y, c)
		}
	}
}
