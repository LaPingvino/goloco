package ui

import (
	"reflect"
	"testing"
)

// fixedFont gives every glyph of every font a width of 6 (space 3), so the
// wrap can be checked without the game's data.
func fixedFont(t *testing.T) {
	old := globalG1Font
	t.Cleanup(func() { globalG1Font = old })
	f := &g1Font{}
	for i := range f.widths {
		for c := range f.widths[i] {
			f.widths[i][c] = 6
		}
		f.widths[i][0] = 3
	}
	globalG1Font = f
}

func TestWrapString(t *testing.T) {
	fixedFont(t)
	for _, c := range []struct {
		text  string
		width int
		want  []string
	}{
		{"ab cd", 100, []string{"ab cd"}},
		// "ab cd ef": the width reaches 30 at 'e' (12+3+12+3 = 30): back to the last space.
		{"ab cd ef", 30, []string{"ab cd", "ef"}},
		// A word as wide as the limit: the break stands after it, its space
		// starts the next line (as upstream: only a break at a space eats it).
		{"abcde fg", 30, []string{"abcde", " fg"}},
		// No space: the break stands where the width was reached, the character that reached it included.
		{"abcdefgh", 18, []string{"abc", "def", "gh"}},
		{"one\x05two", 100, []string{"one", "two"}},
		{"one\x05", 100, []string{"one", ""}},
		{"", 100, nil},
	} {
		if got := WrapString(c.text, G1FontMediumBold, c.width); !reflect.DeepEqual(got, c.want) {
			t.Errorf("WrapString(%q, %d) = %q, want %q", c.text, c.width, got, c.want)
		}
	}
}

// wrapStringTicker: the lines needed for the first n characters grow as the
// characters pass a break at a space.
func TestWrapTickerLines(t *testing.T) {
	fixedFont(t)
	text := "ab cd ef gh" // at 30: "ab cd" / "ef gh"
	for n, want := range map[int]int{0: 1, 5: 1, 6: 2, 20: 2} {
		lines, got := wrapLines(text, G1FontSmall, 30, n)
		if got != want || len(lines) != 2 {
			t.Errorf("%d characters: %d lines (want %d) of %q", n, got, want, lines)
		}
	}
}

func TestFromLocoBytes(t *testing.T) {
	if got := FromLocoBytes("a\xAC\x05"); got != "a¬\x05" {
		t.Errorf("%q", got)
	}
}
