package scene

import (
	"os"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// Every steam object loads (SteamObject::load walks exactly to its image
// table), with frames and measured sprites; every rail vehicle's steam-puff
// emitter names a loaded steam object.
func TestSteamObjectsLoad(t *testing.T) {
	const data = "../../../locomotion"
	if _, err := os.Stat(data + "/Data/g1.DAT"); err != nil {
		t.Skip("game data not present")
	}
	g1, err := assets.LoadG1(data + "/Data/g1.DAT")
	if err != nil {
		t.Fatal(err)
	}
	om := objects.NewObjectManager(data + "/ObjData")
	om.SetBaseSpriteIndex(uint32(g1.GetSpriteCount()))
	om.G1File = g1
	if err := om.LoadAllObjects(); err != nil {
		t.Fatal(err)
	}
	if len(om.SteamObjects) < 4 {
		t.Fatalf("%d steam objects loaded", len(om.SteamObjects))
	}
	for _, s := range om.SteamObjects {
		if len(s.Frames[0]) == 0 || s.Image == 0 || s.SpriteWidth == 0 {
			t.Errorf("%s: frames %d/%d, image %d, width %d", s.Header.GetName(), len(s.Frames[0]), len(s.Frames[1]), s.Image, s.SpriteWidth)
		}
		for k, run := range s.Frames {
			for _, f := range run {
				if int(f.ImageOffset) >= int(s.NumImages) {
					t.Errorf("%s run %d: image %d of %d", s.Header.GetName(), k, f.ImageOffset, s.NumImages)
				}
			}
		}
	}
	puffs := 0
	for _, v := range om.Vehicles {
		for _, a := range v.Animations {
			if a.Type >= objects.EmitterSteamPuff1 && a.Type <= objects.EmitterSteamPuff3 {
				puffs++
				if om.SteamObjectIndex(a.SteamName) < 0 {
					t.Errorf("%s: steam puff names %q, not loaded", v.Header.GetName(), a.SteamName)
				}
			}
		}
	}
	if puffs == 0 {
		t.Error("no vehicle has a steam puff emitter")
	}
	t.Logf("%d steam objects; %d steam-puff emitters", len(om.SteamObjects), puffs)
}
