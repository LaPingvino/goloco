package ui

import (
	"bufio"
	"compress/gzip"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/LaPingvino/goloco/pkg/i18n"
	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// The fallback font (LOOP.md N56): a deliberate deviation from upstream.
// OpenLoco draws every character without a Locomotion glyph as "?"
// (Localisation/Conversion.cpp convertUnicodeToLoco); goloco draws runes
// above U+00FF that have no Loco code from GNU Unifont instead — its 16 px
// bitmaps unscaled (Joop 2026-11-20: the smallest readable CJK is the
// starting point, the UI grows where text is taller). Latin keeps the Loco
// sprite fonts. Unifont is not shipped: it is read from the system at run
// time (fallbackfont_sys.go: the font folders; fallbackfont_js.go: the
// browser's fonts through a canvas), only the glyphs a string needs.
// Without one, the rune stays "?" as upstream.
//
// Licence (checked 2026-11-25, Unifont 17.0.03): dual SIL OFL 1.1 and
// GPLv2+ with the GNU Font Embedding Exception; no Reserved Font Name.

const (
	fallbackCell     = 16 // Unifont's cell height (FONTBOUNDINGBOX 16 16 0 -2)
	fallbackAscent   = 14 // rows above the baseline
	fallbackLineH    = fallbackCell + 2
	fallbackBaseline = 1 + fallbackAscent // in the framed glyph: a 1 px frame as the Loco glyphs'
)

// glyphBits is one 1-bit glyph in a fallbackCell-row cell: rows of
// width pixels, the baseline fallbackAscent rows down.
type glyphBits struct {
	width int
	rows  [][]bool
}

// glyphSource gives the bits of a rune's glyph, ok false without one.
type glyphSource interface {
	glyph(r rune) (glyphBits, bool)
}

// fbGlyph is a fallback glyph as the text styles draw it: the body
// (textRemap0), the outline plane (textRemap1 and 2) and the inset plane
// (textRemap2), white shapes tinted at draw time, framed by 1 px as the Loco
// glyph sprites are.
type fbGlyph struct {
	body, outline, inset *ebiten.Image
	advance              int
}

var (
	fallbackOnce   sync.Once
	fallbackSrc    glyphSource
	fallbackMu     sync.Mutex
	fallbackGlyphs = map[rune]*fbGlyph{}
)

// setFallbackSource replaces the source (tests) and forgets cached glyphs.
func setFallbackSource(s glyphSource) {
	fallbackOnce.Do(func() {})
	fallbackMu.Lock()
	fallbackSrc, fallbackGlyphs = s, map[rune]*fbGlyph{}
	fallbackMu.Unlock()
}

// needsFallback: a rune the Loco fonts have no glyph for (convertUnicodeToLoco
// gives "?"), above Latin-1 — the Loco code page's own range stays Loco.
func needsFallback(r rune) bool {
	return r > 0xFF && i18n.LocoChar(r) == '?'
}

// fallbackGlyph is r's fallback glyph, nil when r has a Loco glyph or no
// fallback font has it. The source is looked for on the first such rune.
func fallbackGlyph(r rune) *fbGlyph {
	if !needsFallback(r) {
		return nil
	}
	fallbackOnce.Do(func() {
		fallbackSrc = findFallbackSource()
		if fallbackSrc == nil {
			log.Printf("[Font] no GNU Unifont found: characters without a Locomotion glyph stay \"?\" (install Unifont, e.g. otf-unifont / bdf-unifont / fonts-unifont, or set GOLOCO_FALLBACK_FONT to a .bdf, .hex or .otf)")
		}
	})
	fallbackMu.Lock()
	defer fallbackMu.Unlock()
	if g, ok := fallbackGlyphs[r]; ok {
		return g
	}
	var g *fbGlyph
	if fallbackSrc != nil {
		if b, ok := fallbackSrc.glyph(r); ok {
			g = framedGlyph(b)
		}
	}
	fallbackGlyphs[r] = g
	return g
}

// framedGlyph builds the three planes of a glyph. The outline rule is goloco's
// own, read off the Loco glyph sprites (medium "I": 0x02 above and left of the
// body, 0x03 below and right, no diagonals): a pixel next to the body takes
// textRemap2 (0x03) when the body is above or left of it, else textRemap1
// (0x02) when the body is below or right of it.
func framedGlyph(b glyphBits) *fbGlyph {
	body, outline, inset := framedMasks(b)
	return &fbGlyph{
		body:    ebiten.NewImageFromImage(body),
		outline: ebiten.NewImageFromImage(outline),
		inset:   ebiten.NewImageFromImage(inset),
		advance: b.width,
	}
}

// framedMasks are framedGlyph's planes as masks.
func framedMasks(b glyphBits) (body, outline, inset *image.Alpha) {
	w, h := b.width+2, fallbackCell+2
	on := func(x, y int) bool {
		x, y = x-1, y-1
		return y >= 0 && y < len(b.rows) && x >= 0 && x < len(b.rows[y]) && b.rows[y][x]
	}
	body = image.NewAlpha(image.Rect(0, 0, w, h))
	outline = image.NewAlpha(image.Rect(0, 0, w, h))
	inset = image.NewAlpha(image.Rect(0, 0, w, h))
	white := color.Alpha{255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			switch {
			case on(x, y):
				body.SetAlpha(x, y, white)
			case on(x-1, y) || on(x, y-1): // textRemap2
				outline.SetAlpha(x, y, white)
				inset.SetAlpha(x, y, white)
			case on(x+1, y) || on(x, y+1): // textRemap1
				outline.SetAlpha(x, y, white)
			}
		}
	}
	return body, outline, inset
}

// hasFallback reports whether str holds a rune drawn from the fallback font:
// its line is then fallbackLineH tall and the Loco glyphs move down onto
// the fallback's baseline.
func hasFallback(str string) bool {
	for _, r := range str {
		if fallbackGlyph(r) != nil {
			return true
		}
	}
	return false
}

// TextLineHeight is the line height of str in a G1 font: the font's own, or
// the fallback's when str has a fallback glyph (never less than the font's).
func TextLineHeight(str string, font int) int {
	if hasFallback(str) {
		return max(g1FontHeight(font), fallbackLineH)
	}
	return g1FontHeight(font)
}

// --- sources ---------------------------------------------------------------

// rawGlyph is a glyph as the file holds it (its hex rows), decoded when
// drawn: a whole Unifont stays a few MB in memory.
type rawGlyph struct {
	dw, bw, bh, bx, by int8 // advance, bitmap box and its offset from the baseline (BDF's DWIDTH / BBX)
	bits               string
}

// mapSource is a parsed bitmap font.
type mapSource map[rune]rawGlyph

func (m mapSource) glyph(r rune) (glyphBits, bool) {
	g, ok := m[r]
	if !ok || g.bh <= 0 || len(g.bits)%int(g.bh) != 0 {
		return glyphBits{}, false
	}
	dw, perRow := int(g.dw), len(g.bits)/int(g.bh)
	out := glyphBits{width: dw}
	top := fallbackAscent - (int(g.bh) + int(g.by)) // the box's first row in the cell
	for y := 0; y < fallbackCell; y++ {
		row := make([]bool, dw)
		if sy := y - top; sy >= 0 && sy < int(g.bh) {
			for x, v := range hexRow(g.bits[sy*perRow:(sy+1)*perRow], perRow*4) {
				if dx := x + int(g.bx); v && x < int(g.bw) && dx >= 0 && dx < dw {
					row[dx] = true
				}
			}
		}
		out.rows = append(out.rows, row)
	}
	return out, true
}

// openMaybeGzip opens a font file, through gzip when it ends in .gz.
func openMaybeGzip(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(strings.ToLower(path), ".gz") {
		return f, nil
	}
	z, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return struct {
		io.Reader
		io.Closer
	}{z, f}, nil
}

// parseHex reads GNU Unifont's .hex format: "CODEPOINT:BITS" lines, BITS 32
// hex digits for an 8×16 glyph or 64 for a 16×16 one (rows top down).
func parseHex(r io.Reader) mapSource {
	out := mapSource{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		cp, bits, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		c, err := strconv.ParseUint(cp, 16, 32)
		if err != nil || len(bits) == 0 || len(bits)%fallbackCell != 0 {
			continue
		}
		w := int8(len(bits) / fallbackCell * 4)
		out[rune(c)] = rawGlyph{dw: w, bw: w, bh: fallbackCell, by: fallbackAscent - fallbackCell, bits: bits}
	}
	return out
}

func hexRow(h string, width int) []bool {
	row := make([]bool, width)
	for i := 0; i < len(h) && i*4 < width; i++ {
		v, err := strconv.ParseUint(h[i:i+1], 16, 8)
		if err != nil {
			continue
		}
		for b := 0; b < 4 && i*4+b < width; b++ {
			row[i*4+b] = v&(8>>b) != 0
		}
	}
	return row
}

// parseBDF reads a BDF bitmap font (Unifont's unifont.bdf): each glyph's
// ENCODING, DWIDTH, BBX and BITMAP rows.
func parseBDF(r io.Reader) mapSource {
	out := mapSource{}
	sc := bufio.NewScanner(r)
	enc := -1
	var g rawGlyph
	var bits strings.Builder
	inBitmap := false
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 0 {
			continue
		}
		if inBitmap {
			if f[0] != "ENDCHAR" {
				bits.WriteString(f[0])
				continue
			}
			inBitmap = false
			if enc >= 0 {
				g.bits = bits.String()
				out[rune(enc)] = g
			}
			continue
		}
		num := func(i int) int8 {
			if i < len(f) {
				v, _ := strconv.Atoi(f[i])
				return int8(v)
			}
			return 0
		}
		switch f[0] {
		case "STARTCHAR":
			enc, g = -1, rawGlyph{}
			bits.Reset()
		case "ENCODING":
			enc = -1
			if len(f) > 1 {
				enc, _ = strconv.Atoi(f[1])
			}
		case "DWIDTH":
			g.dw = num(1)
		case "BBX":
			g.bw, g.bh, g.bx, g.by = num(1), num(2), num(3), num(4)
		case "BITMAP":
			inBitmap = true
		}
	}
	return out
}

// otfSource is an OpenType/TrueType Unifont (Unifont.otf, unifont.ttf)
// rasterised at its 16 px cell without anti-aliasing: Unifont's outlines
// are its pixels, so each lands on whole pixels (TestUnifontOTFMatchesBDF).
type otfSource struct {
	f    *sfnt.Font
	face font.Face
	buf  sfnt.Buffer
}

func newOTFSource(data []byte) *otfSource {
	f, err := opentype.Parse(data)
	if err != nil {
		return nil
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: fallbackCell, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return nil
	}
	return &otfSource{f: f, face: face}
}

func (s *otfSource) glyph(r rune) (glyphBits, bool) {
	if gi, err := s.f.GlyphIndex(&s.buf, r); err != nil || gi == 0 {
		return glyphBits{}, false
	}
	dr, mask, mp, adv, ok := s.face.Glyph(fixed.P(0, fallbackAscent), r)
	if !ok {
		return glyphBits{}, false
	}
	w := adv.Round()
	g := glyphBits{width: w}
	for y := 0; y < fallbackCell; y++ {
		row := make([]bool, w)
		for x := 0; x < w; x++ {
			p := image.Pt(x, y)
			if p.In(dr) {
				_, _, _, a := mask.At(mp.X+x-dr.Min.X, mp.Y+y-dr.Min.Y).RGBA()
				row[x] = a >= 0x8000
			}
		}
		g.rows = append(g.rows, row)
	}
	return g, true
}

// loadFontFile parses a .bdf or .hex font (optionally .gz) or an .otf /
// .ttf, nil if it isn't one or holds no glyphs.
func loadFontFile(path string) glyphSource {
	if ext := strings.ToLower(filepath.Ext(path)); ext == ".otf" || ext == ".ttf" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if s := newOTFSource(data); s != nil {
			log.Printf("[Font] fallback font %s", path)
			return s
		}
		return nil
	}
	rc, err := openMaybeGzip(path)
	if err != nil {
		return nil
	}
	defer rc.Close()
	var m mapSource
	switch p := strings.TrimSuffix(strings.ToLower(path), ".gz"); {
	case strings.HasSuffix(p, ".bdf"):
		m = parseBDF(rc)
	case strings.HasSuffix(p, ".hex"):
		m = parseHex(rc)
	}
	if len(m) == 0 {
		return nil
	}
	log.Printf("[Font] fallback font %s (%d glyphs)", path, len(m))
	return m
}
