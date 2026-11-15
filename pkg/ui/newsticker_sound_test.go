package ui

import "testing"

// Ticker.cpp onUpdate: when numCharsToDisplay & 7 is 0 after its += 2 (every
// second character), the ticker sound for the character at numChars >> 2 if
// it is not a space; a newline is two spaces; nothing past the end.
func TestNewsTickerSound(t *testing.T) {
	w := &NewsTicker{Text: "Ab cd\x05Xyz"}
	var ticks []int
	for i := 1; i <= 20; i++ {
		if w.Update(640, 480, false, false) {
			ticks = append(ticks, int(w.numCharsToDisplay>>2))
		}
	}
	// A(0) b(1) ' '(2) c(3) d(4) newline as ' ' ' '(5,6) X(7) y(8) z(9) end(10):
	// checked at 2, 4, 6, 8, 10 — the sound at d and y
	want := []int{4, 8}
	if len(ticks) != len(want) {
		t.Fatalf("ticks at %v, want %v", ticks, want)
	}
	for i := range want {
		if ticks[i] != want[i] {
			t.Fatalf("ticks at %v, want %v", ticks, want)
		}
	}
	if w.Update(640, 480, false, true) { // paused: nothing
		t.Error("tick while paused")
	}
}
