package ui

import "testing"

func TestInputSession(t *testing.T) {
	s := NewInputSession("Train 1", 5+32)
	if s.CursorPosition != 7 {
		t.Fatalf("caret %d, want at the end", s.CursorPosition)
	}
	s.Key(KeyHome)
	for _, r := range "Night " {
		s.Type(r)
	}
	s.Key(KeyEnd)
	s.Key(KeyBackspace)
	s.Key(KeyLeft)
	s.Key(KeyDelete) // the space before the removed "1"
	if got := string(s.Buffer); got != "Night Train" {
		t.Errorf("edited to %q", got)
	}
	s.Key(KeyLeft)
	s.Key(KeyLeft)
	s.Type('é') // 159–255: allowed
	if got := string(s.Buffer); got != "Night Traéin" {
		t.Errorf("é inserted: %q", got)
	}
	lim := NewInputSession("abc", 3)
	lim.Type('d')
	if string(lim.Buffer) != "abc" {
		t.Errorf("past the limit: %q", string(lim.Buffer))
	}
	if got := (&InputSession{Buffer: []rune("a{b|c«d»")}).Sanitized(); got != "abc«d»" {
		t.Errorf("sanitized %q", got)
	}
}
