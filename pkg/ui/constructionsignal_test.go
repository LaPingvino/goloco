package ui

import "testing"

// The newline control code (5) in object strings starts a new line.
func TestWrapTextBoldNewlineCode(t *testing.T) {
	got := WrapTextBold("Colourlight Signal,\x05American style", 1000)
	if len(got) != 2 || got[0] != "Colourlight Signal," || got[1] != "American style" {
		t.Errorf("%q", got)
	}
}
