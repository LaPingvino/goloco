package ui

import (
	"image"
	"testing"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// The Town window's viewport is one pixel in from its widget (3,44 to
// width − 4, height − 14), only on the first tab; findWidgetAt over it is
// the viewport but for the centre-view button in its bottom-right corner.
func TestViewWindowViewportWidget(t *testing.T) {
	w := &ViewWindow{WinBase: WinBase{X: 100, Y: 200, W: 223, H: 161}}
	if _, ok := w.ViewportRect(); ok {
		t.Fatal("a viewport without a view")
	}
	w.View = func(int, int) *gfx8.RenderTarget { return nil }
	r, ok := w.ViewportRect()
	if want := image.Rect(104, 245, 100+219, 200+147); !ok || r != want {
		t.Errorf("ViewportRect = %v %v, want %v", r, ok, want)
	}
	for _, c := range []struct {
		x, y int
		want bool
	}{
		{103, 244, true},              // the widget's top-left edge
		{100 + 219, 200 + 147, true},  // its bottom-right edge, outside the button
		{100 + 218, 200 + 146, false}, // the centre-view button (right − 24 … right − 1)
		{100 + 195, 200 + 123, false}, // the button's top-left
		{100 + 194, 200 + 146, true},  // left of the button
		{100 + 218, 200 + 122, true},  // above the button
		{102, 244, false},             // outside
	} {
		if got := w.ViewportWidgetAt(c.x, c.y); got != c.want {
			t.Errorf("ViewportWidgetAt(%d, %d) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
	w.Tab = 1
	if w.ViewportWidgetAt(150, 300) {
		t.Error("a viewport widget on the second tab")
	}
}
