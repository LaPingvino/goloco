package i18n

import "unicode/utf8"

// Wide characters in Loco byte strings — a goloco deviation (LOOP.md N56).
// Upstream's 8-bit buffers (a message's text, a user string) can only hold
// Loco codes, so convertUnicodeToLoco makes everything else "?". goloco keeps
// a rune the Loco fonts lack (above U+00FF, LocoChar '?') in memory as
// LocoWide followed by its UTF-8 bytes, so the text renderer draws it from the
// fallback font. LocoWide is ControlCodes::adjustPalette (2), which upstream
// marks "Not used" and no language pack writes. It never reaches a save's
// 8-bit fields: StripWide makes it "?" there, and the Unicode text goes to
// goloco's own chunk (scenario GolocoTextMagic).
const LocoWide = 2

// IsWide is a rune the Loco code page has no glyph for: above Latin-1 and
// converted to '?'.
func IsWide(r rune) bool { return r > 0xFF && LocoChar(r) == '?' }

// AppendLoco appends r converted for a Loco byte string: its Loco code, or
// LocoWide + UTF-8 for a wide rune.
func AppendLoco(b []byte, r rune) []byte {
	if IsWide(r) {
		return utf8.AppendRune(append(b, LocoWide), r)
	}
	return append(b, byte(LocoChar(r)))
}

// NameLoco is a name as goloco's strings carry it (a user string: runes of
// Loco codes, plus wide runes typed by the player) as Loco bytes. A string
// that is not valid UTF-8 is Loco bytes already (an object's string) and is
// returned as it is.
func NameLoco(s string) string {
	if !utf8.ValidString(s) {
		return s
	}
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r <= 0xFF {
			b = append(b, byte(r))
		} else {
			b = AppendLoco(b, r)
		}
	}
	return string(b)
}

// DecodeWide reads a LocoWide sequence at b[i] (b[i] == LocoWide): the rune
// and the bytes it takes; ok false for a cut or malformed one (a buffer's
// length limit can cut one at the end), which takes LocoWide and the UTF-8
// bytes that follow it.
func DecodeWide(b string, i int) (r rune, n int, ok bool) {
	r, sz := utf8.DecodeRuneInString(b[i+1:])
	if r != utf8.RuneError {
		return r, 1 + sz, true
	}
	n = 1
	for n < 4 && i+n < len(b) && b[i+n] >= 0x80 {
		n++
	}
	return 0, n, false
}

// StripWide is a Loco byte string as an 8-bit field holds it: each wide
// sequence "?" (convertUnicodeToLoco's replacement_character); a cut one
// dropped.
func StripWide(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); {
		if b[i] == LocoWide {
			_, n, ok := DecodeWide(string(b), i)
			if ok {
				out = append(out, '?')
			}
			i += n
			continue
		}
		out = append(out, b[i])
		i++
	}
	return out
}

// HasWide: a string as goloco carries it (a user string: runes) holds a
// rune the Loco code page lacks.
func HasWide(s string) bool {
	for _, r := range s {
		if IsWide(r) {
			return true
		}
	}
	return false
}

// HasWideLoco: Loco bytes (a message's text) hold a LocoWide sequence. Only
// the byte counts: a Loco character above 127 isn't UTF-8.
func HasWideLoco(b []byte) bool {
	for _, c := range b {
		if c == LocoWide {
			return true
		}
	}
	return false
}

// ClipLoco is Loco bytes cut to n characters, a LocoWide sequence counting as
// one (the "?" it is in an 8-bit field).
func ClipLoco(s string, n int) string {
	for i := 0; i < len(s); n-- {
		if n == 0 {
			return s[:i]
		}
		if s[i] == LocoWide {
			_, k, _ := DecodeWide(s, i)
			i += k
		} else {
			i++
		}
	}
	return s
}
