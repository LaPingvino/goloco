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
