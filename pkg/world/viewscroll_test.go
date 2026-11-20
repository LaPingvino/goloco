package world

import "testing"

// glideStep: an eighth of the way, rounded up away from the view, its sign
// kept (Window.cpp l. 328–363).
func TestGlideStep(t *testing.T) {
	for _, c := range []struct{ view, saved, want int }{
		{0, 100, 13}, {93, 100, 94}, {99, 100, 100}, {0, -20, -3}, {-13, -20, -14}, {5, -3, 4}, {7, 7, 7},
	} {
		if got, _ := glideStep(c.view, c.saved); got != c.want {
			t.Errorf("glideStep(%d, %d) = %d, want %d", c.view, c.saved, got, c.want)
		}
	}
}

// A glide from (0, 0) to (100, −20), worked by hand: x 13, 24, 34, 43, 51 …
// reaches 100 on the 24th pass, y −3, −6, −8 … −20 on the 13th; the 25th pass
// finds both there and clears scrollingToLocation.
func TestGlide(t *testing.T) {
	w := &World{}
	w.moveWindowToLocation(100, -20)
	var xs, ys []int
	for range 24 {
		w.ViewportsUpdatePosition()
		xs, ys = append(xs, int(w.camX)), append(ys, int(w.camY))
	}
	if xs[0] != 13 || xs[1] != 24 || xs[4] != 51 || ys[0] != -3 || ys[2] != -8 {
		t.Errorf("first steps: x %v y %v", xs[:5], ys[:5])
	}
	if xs[22] != 99 || xs[23] != 100 || ys[11] != -19 || ys[12] != -20 {
		t.Errorf("last steps: x %v y %v", xs[20:], ys[10:])
	}
	if !w.Gliding() {
		t.Fatal("cleared on arriving: upstream clears on the pass after")
	}
	w.ViewportsUpdatePosition()
	if w.Gliding() || w.camX != 100 || w.camY != -20 {
		t.Fatalf("25th pass: gliding %v at %v, %v", w.Gliding(), w.camX, w.camY)
	}
}

// While it glides a pan moves the glide's end (savedView += delta), a
// right drag's begin stops it, a jump (centreOn) ends it, and a view
// following a vehicle neither starts nor steps one.
func TestGlideInteractions(t *testing.T) {
	w := &World{}
	w.moveWindowToLocation(80, 0)
	w.ViewportsUpdatePosition() // 10
	w.PanCamera(16, 8)
	if w.camX != 10 || w.glide.x != 96 || w.glide.y != 8 {
		t.Errorf("pan: cam %v saved %d,%d", w.camX, w.glide.x, w.glide.y)
	}
	w.ViewportsUpdatePosition() // 10 + ceil(86 / 8) = 21
	if w.camX != 21 || w.camY != 1 {
		t.Errorf("after the pan: %v, %v", w.camX, w.camY)
	}
	w.StopGlide()
	w.ViewportsUpdatePosition()
	if w.camX != 21 {
		t.Errorf("stopped glide moved: %v", w.camX)
	}
	w.PanCamera(4, 0)
	if w.camX != 25 {
		t.Errorf("pan without a glide: %v", w.camX)
	}
	w.following = true
	w.moveWindowToLocation(500, 500)
	if w.Gliding() {
		t.Error("a followed view took a glide")
	}
}
