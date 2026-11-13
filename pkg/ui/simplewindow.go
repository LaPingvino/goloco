package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// SimpleWindow is a window whose owner draws it and takes its clicks through
// callbacks (the Construction window, drawn by cmd/goloco from the
// construction tabs): PortedDraw draws it at its position; the close button
// is upstream's (width − 15, 2, 13 × 13); other presses go to OnWindowClick
// in window coordinates. The stack moves it by its caption.
type SimpleWindow struct {
	WinBase
	PortedDraw    func(screen *ebiten.Image, x, y int)
	OnWindowClick func(relX, relY int)
	// OnScrollWheel is the mouse wheel over the window (dy > 0: away from
	// the user).
	OnScrollWheel func(dy float64)
}

// NewSimpleWindow creates a window of a type at x, y.
func NewSimpleWindow(typ WindowType, x, y, width, height int) *SimpleWindow {
	return &SimpleWindow{WinBase: WinBase{WinType: typ, X: x, Y: y, W: width, H: height}}
}

// HandleClick is a left press on the window.
func (w *SimpleWindow) HandleClick(x, y int) {
	switch rx, ry := x-w.X, y-w.Y; {
	case rx >= w.W-15 && rx < w.W-2 && ry >= 2 && ry < 15:
		w.Close()
	case w.OnWindowClick != nil:
		w.OnWindowClick(rx, ry)
	}
}

// Draw draws the window. It is marked foreign for draw8's screen copy: its
// owner draws past that copy.
func (w *SimpleWindow) Draw(screen *ebiten.Image) {
	if w.PortedDraw != nil {
		w.PortedDraw(screen, w.X, w.Y)
	}
	MarkForeign(image.Rect(w.X, w.Y, w.X+w.W, w.Y+w.H))
}
