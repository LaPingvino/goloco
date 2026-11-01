package scene

import (
	"os"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/objects"
)

// Every industry object in the game data must walk to exactly its image table:
// upstream asserts remainingData.size() == imgRes.tableLength in load().
func TestIndustryObjectsWalkExactly(t *testing.T) {
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
	if len(om.IndustryObjects) == 0 {
		t.Fatal("no industry objects loaded")
	}
	for _, o := range om.IndustryObjects {
		if int(o.TableLength) != o.RemainingAtTable {
			t.Errorf("%s: image table %d bytes, %d left at the table", o.Name, o.TableLength, o.RemainingAtTable)
		}
	}
	t.Logf("%d industry objects walk exactly", len(om.IndustryObjects))
}
