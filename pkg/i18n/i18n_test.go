package i18n

import (
	"strings"
	"testing"
)

const sampleYML = `header:
  locale: xx-YY
  english_name: Test
  native_name: "Tëst"
  loco_original_id: 7
strings:
  0: ""
  1: "{POP16}"
  103: Load Game
  186: Cancel
  635: January
  # a comment line
  1573: Challenging
  219: "{COLOUR WINDOW_2}Cost: {COLOUR BLACK}{CURRENCY32}"
`

func TestParse(t *testing.T) {
	strs, hdr := parse(strings.NewReader(sampleYML))
	if hdr.locoOriginalID != 7 {
		t.Errorf("locoOriginalID = %d, want 7", hdr.locoOriginalID)
	}
	if hdr.nativeName != "Tëst" {
		t.Errorf("nativeName = %q, want Tëst", hdr.nativeName)
	}
	cases := map[int]string{
		0:    "",
		1:    "{POP16}",
		103:  "Load Game",
		186:  "Cancel",
		635:  "January",
		1573: "Challenging",
	}
	for id, want := range cases {
		if got := strs[id]; got != want {
			t.Errorf("strs[%d] = %q, want %q", id, got, want)
		}
	}
}

func TestIsFormatOnly(t *testing.T) {
	for _, s := range []string{"", "{POP16}", "{COLOUR BLACK}", "  {A}{B} "} {
		if !isFormatOnly(s) {
			t.Errorf("isFormatOnly(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"Cancel", "{COLOUR WINDOW_2}Cost:", "1st"} {
		if isFormatOnly(s) {
			t.Errorf("isFormatOnly(%q) = true, want false", s)
		}
	}
}

func TestTFallback(t *testing.T) {
	// With no pack loaded, T serves English fallbacks and never returns empty.
	loaded = map[int]string{}
	if got := T("menu.new_game"); got != "New Game" {
		t.Errorf("T(menu.new_game) = %q, want New Game", got)
	}
	if got := T("unknown.key"); got != "unknown.key" {
		t.Errorf("T(unknown.key) = %q, want the key itself", got)
	}
	// Loaded pack text wins over fallback for a mapped key.
	loaded = map[int]string{186: "Abbrechen"}
	if got := T("newgame.cancel"); got != "Abbrechen" {
		t.Errorf("T(newgame.cancel) = %q, want Abbrechen", got)
	}
	// A format-only loaded value is ignored in favour of the fallback.
	loaded = map[int]string{186: "{POP16}"}
	if got := T("newgame.cancel"); got != "Cancel" {
		t.Errorf("T(newgame.cancel) with format-only pack = %q, want Cancel", got)
	}
	loaded = map[int]string{}
}

func TestMonth(t *testing.T) {
	loaded = map[int]string{}
	if got := Month(1); got != "January" {
		t.Errorf("Month(1) = %q, want January", got)
	}
	if got := Month(0); got != "" {
		t.Errorf("Month(0) = %q, want empty", got)
	}
}
