package ui

import "testing"

func TestFillArgs(t *testing.T) {
	got := FillArgs("{COLOUR BLACK}{STRINGID} transported {INT16} blocks in {INT16} days = {CURRENCY32}",
		"{COMMA16} passengers", "300", "55", "76", "£27,525")
	if want := "{COLOUR BLACK}300 passengers transported 55 blocks in 76 days = £27,525"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	// A cargo's unit name as loaded: the raw byte 123 (int32_grouped) takes the quantity.
	got = FillArgs("{COLOUR BLACK}{STRINGID} transported {INT16} blocks in {INT16} days = {CURRENCY32}",
		ObjectString("{ tonnes of Coal"), "16", "17", "5", "$1,204")
	if want := "{COLOUR BLACK}16 tonnes of Coal transported 17 blocks in 5 days = $1,204"; got != want {
		t.Errorf("object unit: got %q, want %q", got, want)
	}
	if got := FillArgs("{COLOUR WINDOW_2}Last income: {COLOUR BLACK}N/A"); got != "{COLOUR WINDOW_2}Last income: {COLOUR BLACK}N/A" {
		t.Errorf("no args: %q", got)
	}
	for n, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 1234567: "1,234,567", -4500: "-4,500"} {
		if got := Comma16(n); got != want {
			t.Errorf("Comma16(%d) = %q, want %q", n, got, want)
		}
	}
}
