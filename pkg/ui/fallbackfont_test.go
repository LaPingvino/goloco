package ui

import (
	"os"
	"strings"
	"testing"
)

// A two-glyph BDF in Unifont's layout: 日 (16×16, BBX 16 16 0 -2) and a
// small box at the baseline (8 wide, BBX 4 2 2 0: rows 12 and 13).
const testBDF = `STARTFONT 2.1
FONTBOUNDINGBOX 16 16 0 -2
CHARS 2
STARTCHAR U+65E5
ENCODING 26085
DWIDTH 16 0
BBX 16 16 0 -2
BITMAP
0000
1FF0
1010
1010
1010
1010
1010
1FF0
1010
1010
1010
1010
1010
1010
1FF0
1010
ENDCHAR
STARTCHAR U+0416
ENCODING 1046
DWIDTH 8 0
BBX 4 2 2 0
BITMAP
F0
90
ENDCHAR
ENDFONT
`

func bitsString(g glyphBits) string {
	var b strings.Builder
	for _, row := range g.rows {
		for _, v := range row {
			if v {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestParseBDFPlacesByBaseline(t *testing.T) {
	m := parseBDF(strings.NewReader(testBDF))
	ri, ok := m.glyph(0x65E5)
	if !ok || ri.width != 16 || len(ri.rows) != fallbackCell {
		t.Fatalf("日: ok %v width %d rows %d", ok, ri.width, len(ri.rows))
	}
	if got := bitsString(ri); !strings.HasPrefix(got, "................\n...#########....\n...#.......#....\n") {
		t.Errorf("日 rows:\n%s", got)
	}
	zh, ok := m.glyph(0x0416)
	if !ok || zh.width != 8 {
		t.Fatalf("box: ok %v width %d", ok, zh.width)
	}
	// BBX 4 2 2 0: its bottom row sits on the baseline (row 13), 2 px in
	want := strings.Repeat("........\n", 12) + "..####..\n..#..#..\n" + strings.Repeat("........\n", 2)
	if got := bitsString(zh); got != want {
		t.Errorf("box rows:\n%s\nwant\n%s", got, want)
	}
	if _, ok := m.glyph('?'); ok {
		t.Error("a glyph the font lacks")
	}
}

func TestParseHexMatchesBDF(t *testing.T) {
	hex := "65E5:0000" + "1FF0" + strings.Repeat("1010", 5) + "1FF0" + strings.Repeat("1010", 6) + "1FF0" + "1010\n"
	h, ok := parseHex(strings.NewReader(hex)).glyph(0x65E5)
	b, _ := parseBDF(strings.NewReader(testBDF)).glyph(0x65E5)
	if !ok || bitsString(h) != bitsString(b) {
		t.Errorf("hex:\n%s\nbdf:\n%s", bitsString(h), bitsString(b))
	}
}

// The outline planes follow the Loco glyphs' rule: textRemap1 above / left
// of the body, textRemap2 below / right, no diagonals.
func TestFramedGlyphPlanes(t *testing.T) {
	g := glyphBits{width: 1, rows: make([][]bool, fallbackCell)}
	for y := range g.rows {
		g.rows[y] = []bool{y == 5}
	}
	fb, fo, fi := framedMasks(g)
	w, h := fb.Bounds().Dx(), fb.Bounds().Dy()
	if w != 3 || h != fallbackCell+2 {
		t.Fatalf("framed %d×%d", w, h)
	}
	type pt struct{ x, y int }
	body, inset, outline := map[pt]bool{}, map[pt]bool{}, map[pt]bool{}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if _, _, _, a := fb.At(x, y).RGBA(); a != 0 {
				body[pt{x, y}] = true
			}
			if _, _, _, a := fi.At(x, y).RGBA(); a != 0 {
				inset[pt{x, y}] = true
			}
			if _, _, _, a := fo.At(x, y).RGBA(); a != 0 {
				outline[pt{x, y}] = true
			}
		}
	}
	if len(body) != 1 || !body[pt{1, 6}] {
		t.Errorf("body %v", body)
	}
	if len(inset) != 2 || !inset[pt{2, 6}] || !inset[pt{1, 7}] {
		t.Errorf("inset (textRemap2) %v", inset)
	}
	if len(outline) != 4 || !outline[pt{0, 6}] || !outline[pt{1, 5}] {
		t.Errorf("outline %v", outline)
	}
}

// Text with a fallback rune measures with its glyphs and a taller line;
// without a source it stays as upstream ("?").
func TestFallbackMeasure(t *testing.T) {
	defer setFallbackSource(nil)
	setFallbackSource(parseBDF(strings.NewReader(testBDF)))
	if g := fallbackGlyph(0x65E5); g == nil || g.advance != 16 {
		t.Fatalf("日: %+v", g)
	}
	if fallbackGlyph('A') != nil || fallbackGlyph('é') != nil {
		t.Error("a Loco character from the fallback")
	}
	if got := charWidth(G1FontMediumNormal, 0x65E5); got != 16 {
		t.Errorf("charWidth 日 = %d", got)
	}
	if TextLineHeight("日", G1FontMediumNormal) != fallbackLineH || TextLineHeight("A", G1FontMediumNormal) != 10 {
		t.Error("line heights")
	}
	if lines := WrapString("日日 日日", G1FontMediumNormal, 40); len(lines) != 2 {
		t.Errorf("wrapped %q", lines)
	}
	setFallbackSource(nil)
	if hasFallback("日") {
		t.Error("a fallback glyph without a source")
	}
}

// Unifont's OTF gives the BDF's pixels, when both are installed here.
func TestUnifontOTFMatchesBDF(t *testing.T) {
	const bdf, otf = "/usr/share/fonts/misc/unifont.bdf", "/usr/share/fonts/Unifont/Unifont.otf"
	if _, err := os.Stat(bdf); err != nil {
		t.Skip("no unifont.bdf")
	}
	data, err := os.ReadFile(otf)
	if err != nil {
		t.Skip("no Unifont.otf")
	}
	o := newOTFSource(data)
	b := loadFontFile(bdf)
	if o == nil || b == nil {
		t.Fatal("load")
	}
	for _, r := range "日本語中文한국어Русскийぁア" {
		bg, ok1 := b.glyph(r)
		og, ok2 := o.glyph(r)
		if !ok1 || !ok2 || bitsString(bg) != bitsString(og) {
			t.Errorf("%c: bdf %v otf %v\n%s\n%s", r, ok1, ok2, bitsString(bg), bitsString(og))
		}
	}
}

// Typed names and news (N56 leftovers): a wide rune the fallback draws can be
// typed and is kept by sanitizeInput; a message's LocoWide sequence decodes.
func TestFallbackTypedAndNews(t *testing.T) {
	defer setFallbackSource(nil)
	setFallbackSource(parseBDF(strings.NewReader(testBDF)))
	s := NewInputSession("", 10)
	for _, r := range "A日一" { // 一 isn't in the test font: "?"
		s.Type(r)
	}
	if got := s.Sanitized(); got != "A日?" {
		t.Errorf("typed %q", got)
	}
	if got := FromLocoBytes("At \x02\xe6\x97\xa5\xe9\x02\xe6"); got != "At 日é" {
		t.Errorf("FromLocoBytes %q", got)
	}
}
