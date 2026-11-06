package assets

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// NewOpenG1 is the open build's stand-in for g1.DAT (item O): no original
// images — the base pool is the disc layout's G1ExpectedCountDisc empty
// entries, so the fixed g1 ids stay in range (they draw nothing) and object
// images append after them as with the real file — and the palette from a
// JASC-PAL file (OpenGraphics' palettes/locopal.jasc.pal, MIT: identical to
// Locomotion's over the sprite range 10–244). Index 0 is transparent, as
// loadPalette makes it.
func NewOpenG1(palettePath string) (*G1File, error) {
	f, err := os.Open(palettePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	var header []string
	for len(header) < 3 && sc.Scan() {
		header = append(header, strings.TrimSpace(sc.Text()))
	}
	if len(header) < 3 || header[0] != "JASC-PAL" || header[2] != "256" {
		return nil, fmt.Errorf("%s: not a 256-colour JASC-PAL palette", palettePath)
	}
	g1 := &G1File{Elements: make([]G1Element, G1ExpectedCountDisc)}
	g1.Header.NumEntries = G1ExpectedCountDisc
	for i := 0; i < 256; i++ {
		if !sc.Scan() {
			return nil, fmt.Errorf("%s: %d colours, want 256", palettePath, i)
		}
		p := strings.Fields(sc.Text())
		if len(p) != 3 {
			return nil, fmt.Errorf("%s: colour %d: %q", palettePath, i, sc.Text())
		}
		var c [3]uint8
		for k := range c {
			v, err := strconv.Atoi(p[k])
			if err != nil || v < 0 || v > 255 {
				return nil, fmt.Errorf("%s: colour %d: %q", palettePath, i, sc.Text())
			}
			c[k] = uint8(v)
		}
		g1.Palette[i] = color.RGBA{c[0], c[1], c[2], 255}
	}
	g1.Palette[0] = color.RGBA{}
	g1.totalSprites = G1ExpectedCountDisc
	return g1, nil
}
