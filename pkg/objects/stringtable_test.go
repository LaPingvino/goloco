package objects

import (
	"bytes"
	"testing"
)

func b(s string) []byte {
	return []byte(s)
}

func TestParseStringTable_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expLen   int
		expFirst string
	}{
		{
			name:     "single language simple",
			input:    append(append([]byte{0x09}, b("Hello")...), 0x00, 0xFF),
			expLen:   1 + len("Hello") + 1 + 1, // lang byte + "Hello" + null + 0xFF
			expFirst: "Hello",
		},
		{
			name:     "two languages",
			input:    append(append(append(append([]byte{0x01}, b("One")...), 0x00, 0x02), b("Two")...), 0x00, 0xFF),
			expLen:   1 + len("One") + 1 + 1 + len("Two") + 1 + 1, // l1 + "One"+null + l2 + "Two"+null + 0xFF
			expFirst: "One",
		},
		{
			name:     "empty first string",
			input:    []byte{0x09, 0x00, 0xFF},
			expLen:   3,
			expFirst: "",
		},
		{
			name:     "no terminator",
			input:    append([]byte{0x09}, append(b("NoTerm"), 0x00)...), // missing 0xFF
			expLen:   1 + len("NoTerm") + 1,
			expFirst: "NoTerm",
		},
		{
			name:     "0xff inside string (format code) removed",
			input:    append(append([]byte{0x09}, []byte{'A', 0xFF, 'B'}...), 0x00, 0xFF),
			expLen:   1 + 3 + 1 + 1,
			expFirst: "AB",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n, s := parseStringTable(tc.input)
			if n != tc.expLen {
				t.Fatalf("length mismatch: got %d, want %d", n, tc.expLen)
			}
			if s != tc.expFirst {
				t.Fatalf("first string mismatch: got %q, want %q", s, tc.expFirst)
			}
		})
	}
}

func TestParseStringTable_EmptyAndShort(t *testing.T) {
	// empty input
	n, s := parseStringTable([]byte{})
	if n != 0 || s != "" {
		t.Fatalf("empty input: got len=%d str=%q; want len=0 str=\"\"", n, s)
	}

	// single 0xFF only
	n, s = parseStringTable([]byte{0xFF})
	if n != 1 || s != "" {
		t.Fatalf("only terminator: got len=%d str=%q; want len=1 str=\"\"", n, s)
	}

	// single language byte but no string or terminator
	n, s = parseStringTable([]byte{0x09})
	if n != 1 || s != "" {
		t.Fatalf("single lang only: got len=%d str=%q; want len=1 str=\"\"", n, s)
	}
}

func TestCleanString_RemovesNonPrintableAndFF(t *testing.T) {
	in := []byte{'H', 'i', 0x01, '!', 0xFF, 'X', 0x00} // includes non-printable 0x01 and 0xFF and trailing NUL
	got := cleanString(in)
	if got != "Hi!X" {
		t.Fatalf("cleanString: got %q want %q", got, "Hi!X")
	}
}

func TestParseStringTable_MultiLanguageEdgeCases(t *testing.T) {
	// languages with empty second string, then terminator
	data := append(append([]byte{0x01}, b("First")...), 0x00)
	data = append(data, 0x02)       // second lang id
	data = append(data, 0x00)       // empty second string (null)
	data = append(data, 0x03)       // third lang id (no string, truncated)
	data = append(data, 0x04, 0x00) // fourth lang id with empty string
	data = append(data, 0xFF)       // terminator

	n, s := parseStringTable(data)
	if s != "First" {
		t.Fatalf("expected first string to be 'First', got %q", s)
	}
	if n != len(data) {
		t.Fatalf("expected length %d got %d", len(data), n)
	}
}

func TestParseStringTable_TrimsAtFirstNull(t *testing.T) {
	// Strings must end at first NUL. Ensure embedded bytes after NUL are ignored.
	// Construct: lang id, "ab\0rest", FF
	raw := []byte{0x09, 'a', 'b', 0x00, 'r', 'e', 's', 't', 0xFF}
	n, s := parseStringTable(raw)
	if s != "ab" {
		t.Fatalf("expected 'ab', got %q", s)
	}
	// length should include up to the FF (it finds FF at byte index 8)
	if n != len(raw) {
		t.Fatalf("expected length %d got %d", len(raw), n)
	}
}

func TestParseStringTable_FirstStringOnlyWhenPrintable(t *testing.T) {
	// If first string contains only non-printable characters, it should be returned as empty.
	input := []byte{0x09, 0x01, 0x02, 0x00, 0xFF} // language id + two non-printable bytes + NUL + FF
	n, s := parseStringTable(input)
	if n != len(input) {
		t.Fatalf("unexpected length: got %d want %d", n, len(input))
	}
	if s != "" {
		t.Fatalf("expected empty first string, got %q", s)
	}
}

func TestCleanString_PreservesPrintable(t *testing.T) {
	in := []byte("Hello, World!")
	out := cleanString(in)
	if out != "Hello, World!" {
		t.Fatalf("cleanString modified printable string: got %q", out)
	}
}

func TestParseStringTable_DoesNotOverrun(t *testing.T) {
	// Construct intentionally truncated input where language byte points to EOF.
	data := []byte{0x09}
	n, s := parseStringTable(data)
	if n != 1 {
		t.Fatalf("expected length 1 for truncated input, got %d", n)
	}
	if s != "" {
		t.Fatalf("expected empty string for truncated input, got %q", s)
	}

	// truncated after language + some string bytes but missing NUL/FF
	data2 := []byte{0x09, 'a', 'b'}
	n2, s2 := parseStringTable(data2)
	if n2 != 3 {
		t.Fatalf("expected length 3 got %d", n2)
	}
	if s2 != "ab" {
		// since there was no NUL, parseStringTable takes everything until EOF as the string
		t.Fatalf("expected 'ab', got %q", s2)
	}
}

func TestParseStringTable_RepeatedRun(t *testing.T) {
	// multiple runs should consistently return the first string
	raw := append(append([]byte{0x01}, b("One")...), 0x00, 0x02)
	raw = append(raw, b("Two")...)
	raw = append(raw, 0x00, 0xFF)

	for i := 0; i < 3; i++ {
		n, s := parseStringTable(raw)
		if s != "One" {
			t.Fatalf("iteration %d: expected 'One', got %q", i, s)
		}
		if n != len(raw) {
			t.Fatalf("iteration %d: expected len %d got %d", i, len(raw), n)
		}
	}
}

func TestCleanString_WithEmbeddedNulls(t *testing.T) {
	// cleanString should stop at the first NUL byte
	in := []byte{'A', 'B', 0x00, 'C', 0xFF}
	out := cleanString(in)
	if out != "AB" {
		t.Fatalf("cleanString did not stop at NUL: got %q", out)
	}
}

func TestParseStringTable_IntegrationWithCleanString(t *testing.T) {
	// Ensure parseStringTable returns a cleaned string when format bytes are present.
	raw := append(append([]byte{0x09}, []byte{'X', 0xFF, 'Y'}...), 0x00, 0xFF)
	n, s := parseStringTable(raw)
	if n != len(raw) {
		t.Fatalf("expected length %d got %d", len(raw), n)
	}
	if s != "XY" {
		t.Fatalf("expected cleaned 'XY', got %q", s)
	}
}

func TestParseStringTable_NoUnexpectedBufferReads(t *testing.T) {
	// Create a buffer with some trailing garbage the parser should ignore after terminator.
	prefix := append(append([]byte{0x09}, b("Lead")...), 0x00, 0xFF)
	trailer := []byte{0x99, 0x80, 0x00}
	data := append(prefix, trailer...)
	n, s := parseStringTable(data)
	if n != len(prefix) {
		t.Fatalf("expected n=%d got %d", len(prefix), n)
	}
	if s != "Lead" {
		t.Fatalf("expected 'Lead', got %q", s)
	}
	// ensure trailer not considered part of the table
	if !bytes.Equal(data[n:], trailer) {
		t.Fatalf("trailer was altered or mis-indexed")
	}
}
