package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A plain straight whose track object has a mod 0 that fits straights, with
// that mod not yet on it.
func findModdableStraight(t *testing.T, m *Map, c *Cmd) (int, int, scenario.TrackElement) {
	t.Helper()
	for y := 4; y < m.Height-4; y++ {
		for x := 4; x < m.Width-4; x++ {
			tl := m.TileAt(x, y)
			if len(tl.Tracks) != 1 {
				continue
			}
			el := tl.Tracks[0]
			if el.TrackID != 0 || el.Ghost || el.Mods&1 != 0 {
				continue
			}
			tr := c.trackObject(el.TrackObjectID)
			if tr == nil || len(tr.ModNames) == 0 {
				continue
			}
			if ex := c.Objects.Extra(tr.ModNames[0]); ex != nil && ex.Pieces&TrackCompatibleFlags[0] == TrackCompatibleFlags[0] {
				return x, y, el
			}
		}
	}
	t.Skip("no moddable straight")
	return 0, 0, scenario.TrackElement{}
}

func TestCreateTrackModSingle(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	x, y, el := findModdableStraight(t, m, c)
	tl := m.TileAt(x, y)
	a := TrackModsPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation, Index: el.SequenceIndex, Type: 1, TrackObjectID: el.TrackObjectID, Section: ModSectionSingle}

	dry, ok := m.CreateTrackMod(c, 0, a)
	if !ok || dry <= 0 || tl.Tracks[0].Mods&1 != 0 {
		t.Fatalf("dry run: ok %v cost %d mods %#x", ok, dry, tl.Tracks[0].Mods)
	}
	// A ghost only flags the element (the controlling company's).
	m.ControllingCompany = el.Owner
	if _, ok := m.CreateTrackMod(c, CmdApply|CmdGhost, a); !ok || !tl.Tracks[0].HasGhostMods || tl.Tracks[0].Mods&1 != 0 {
		t.Fatalf("ghost: ok %v ghostMods %v mods %#x", ok, tl.Tracks[0].HasGhostMods, tl.Tracks[0].Mods)
	}
	if _, ok := m.RemoveTrackMod(c, CmdApply|CmdGhost, a); !ok || tl.Tracks[0].HasGhostMods {
		t.Fatal("the ghost mod wasn't taken off")
	}
	cost, ok := m.CreateTrackMod(c, CmdApply, a)
	if !ok || cost != dry || tl.Tracks[0].Mods&1 == 0 {
		t.Fatalf("apply: ok %v cost %d (dry %d) mods %#x", ok, cost, dry, tl.Tracks[0].Mods)
	}
	if again, _ := m.CreateTrackMod(c, 0, a); again != 0 {
		t.Errorf("the mod again costs %d, want 0", again)
	}
	refund, ok := m.RemoveTrackMod(c, CmdApply, a)
	if !ok || refund >= 0 || tl.Tracks[0].Mods&1 != 0 {
		t.Errorf("removal: ok %v refund %d mods %#x", ok, refund, tl.Tracks[0].Mods)
	}
}

// All connected: more than the one piece gets the mod, each paid once.
func TestCreateTrackModAllConnected(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	x, y, el := findModdableStraight(t, m, c)
	count := func() int {
		n := 0
		for ty := 0; ty < m.Height; ty++ {
			for tx := 0; tx < m.Width; tx++ {
				for _, e := range m.TileAt(tx, ty).Tracks {
					if e.Mods&1 != 0 && e.TrackObjectID == el.TrackObjectID {
						n++
					}
				}
			}
		}
		return n
	}
	before := count()
	a := TrackModsPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(el.BaseZ) * 4, Rotation: el.Rotation, Index: el.SequenceIndex, Type: 1, TrackObjectID: el.TrackObjectID, Section: ModSectionAllConnected}
	single, _ := m.CreateTrackMod(c, 0, TrackModsPlacement{X: a.X, Y: a.Y, Z: a.Z, Rotation: a.Rotation, Index: a.Index, Type: 1, TrackObjectID: a.TrackObjectID})
	dryAll, _ := m.CreateTrackMod(c, 0, a) // the reverse marked visited, or this doubles
	cost, ok := m.CreateTrackMod(c, CmdApply, a)
	if !ok {
		t.Fatal("refused")
	}
	if dryAll != cost {
		t.Errorf("all connected: dry run %d, applied %d", dryAll, cost)
	}
	after := count()
	t.Logf("tiles with the mod: %d → %d; cost %d (one straight %d)", before, after, cost, single)
	if after-before < 2 || cost <= single {
		t.Errorf("all connected modded %d tiles for %d (one straight %d)", after-before, cost, single)
	}
}
