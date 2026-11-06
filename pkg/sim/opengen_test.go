package sim

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// The open build's title map (item O): with OpenGraphics' objects only
// (GOLOCO_OPENGRAPHICS: a folder with locopal.jasc.pal and ObjData/ holding
// the OG objects, loaded into assets.NewOpenG1's pool;
// skipped without), the editor's default options generate a 384 × 384
// landscape with land and water. (Trees: 0 so far — an open finding, logged
// in the LOOP; asserted once found.)
func TestNewGeneratedMapOpenGraphics(t *testing.T) {
	dir := os.Getenv("GOLOCO_OPENGRAPHICS")
	if dir == "" {
		t.Skip("GOLOCO_OPENGRAPHICS not set")
	}
	g1, err := assets.NewOpenG1(filepath.Join(dir, "locopal.jasc.pal")) // the hill shapes are images in the pool
	if err != nil {
		t.Fatal(err)
	}
	om := objects.NewObjectManager(filepath.Join(dir, "ObjData"))
	om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
	om.G1File = g1
	if err := om.LoadAllObjects(); err != nil {
		t.Fatal(err)
	}
	m := NewGeneratedMap(0x1234567, 0x89ABCDE, 1900, EditorDefaultTerrainOptions(om), g1, om)
	water, land, trees, maxZ := 0, 0, 0, uint8(0)
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			t := m.TileAt(x, y)
			if t.WaterLevel != 0 {
				water++
			} else {
				land++
			}
			trees += len(t.Trees)
			maxZ = max(maxZ, t.BaseZ)
		}
	}
	if land == 0 || water == 0 {
		t.Fatalf("land %d water %d trees %d, highest baseZ %d", land, water, trees, maxZ)
	}
	t.Logf("land %d, water %d, trees %d, highest baseZ %d", land, water, trees, maxZ)
}
