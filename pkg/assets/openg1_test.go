package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// The open build's pool (item O): OpenGraphics' palette and objects
// (GOLOCO_OPENGRAPHICS: the unzipped objects.zip with locopal.jasc.pal in it;
// skipped without) — the objects register their images after the empty base
// pool, and an image decodes to visible pixels in the palette.
func TestOpenG1WithOpenGraphics(t *testing.T) {
	dir := os.Getenv("GOLOCO_OPENGRAPHICS")
	if dir == "" {
		t.Skip("GOLOCO_OPENGRAPHICS not set")
	}
	g1, err := NewOpenG1(filepath.Join(dir, "locopal.jasc.pal"))
	if err != nil {
		t.Fatal(err)
	}
	if g1.Palette[10].A == 0 || g1.GetSpriteCount() != G1ExpectedCountDisc {
		t.Fatalf("palette[10] %v, base %d", g1.Palette[10], g1.GetSpriteCount())
	}
	m := objects.NewObjectManager(dir)
	m.SetBaseSpriteIndex(uint32(G1ExpectedCountDisc))
	m.G1File = g1
	n := 0
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasPrefix(d.Name(), "OG_") {
			if _, err := m.LoadObject(p); err == nil {
				n++
			}
		}
		return nil
	})
	if int(g1.GetTotalSprites()) <= G1ExpectedCountDisc {
		t.Fatalf("no object images registered (%d objects loaded)", n)
	}
	visible := 0
	for i := G1ExpectedCountDisc; i < len(g1.Elements) && visible == 0; i++ {
		img, err := g1.DecodeSprite(i)
		if err != nil || img == nil {
			continue
		}
		for k := 3; k < len(img.Pix); k += 4 {
			if img.Pix[k] != 0 {
				visible++
			}
		}
	}
	if visible == 0 {
		t.Fatal("no object image decoded to visible pixels")
	}
	t.Logf("%d objects, %d images in the pool (%d object images)", n, g1.GetTotalSprites(), int(g1.GetTotalSprites())-G1ExpectedCountDisc)
}

// The open pool's colour maps: the 16 colours with an exact rule have one
// (identity but the 12 remap slots, onto their swatch), the rest none.
func TestOpenG1Remaps(t *testing.T) {
	p := filepath.Join(t.TempDir(), "pal")
	var b strings.Builder
	b.WriteString("JASC-PAL\n0100\n256\n")
	for i := 0; i < 256; i++ {
		b.WriteString("1 2 3\n")
	}
	os.WriteFile(p, []byte(b.String()), 0o644)
	g1, err := NewOpenG1(p)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for c := 0; c < 32; c++ {
		m, err := g1.GetPaletteMap(c)
		if _, ok := OpenRemap(c); ok != (err == nil) {
			t.Errorf("colour %d: rule %v, map error %v", c, ok, err)
		}
		if err == nil {
			n++
			if m[100] != 100 || m[7] == 7 {
				t.Errorf("colour %d: map %v…", c, m[:12])
			}
		}
	}
	if n != 16 {
		t.Errorf("%d colours mapped, want 16", n)
	}
}
