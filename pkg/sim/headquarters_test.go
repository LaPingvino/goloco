package sim

import (
	"testing"

	"github.com/LaPingvino/goloco/pkg/objects"
)

// An empty flat 2×2 near a town, not overlapping skip.
func findFlat2x2(m *Map, skipX, skipY int) (int, int, bool) {
	for _, tw := range m.Towns {
		for r := 3; r < 14; r++ {
			for dy := -r; dy <= r; dy++ {
				for dx := -r; dx <= r; dx++ {
					x, y := int(tw.X)/32+dx, int(tw.Y)/32+dy
					if skipX >= 0 && x >= skipX-2 && x <= skipX+2 && y >= skipY-2 && y <= skipY+2 {
						continue
					}
					z, ok := -1, true
					for k := 0; k < 4 && ok; k++ {
						tl := m.TileAt(x+k%2, y+k/2)
						if tl == nil || len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 || tl.IsIndustrial || (z >= 0 && int(tl.BaseZ) != z) {
							ok = false
						} else {
							z = int(tl.BaseZ)
						}
					}
					if ok {
						return x, y, true
					}
				}
			}
		}
	}
	return 0, 0, false
}

func TestBuildCompanyHeadquarters(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	hqObj := -1
	for i, b := range om.BuildingObjects {
		if b != nil && b.Flags&objects.BuildingFlagIsHeadquarters != 0 {
			hqObj = i
			break
		}
	}
	if hqObj < 0 {
		t.Skip("no headquarters object")
	}
	company := -1
	for i := range m.Companies {
		if !m.CompanyEmpty(i) {
			company = i
			break
		}
	}
	if company < 0 {
		t.Skip("no company")
	}
	c := &Cmd{Company: uint8(company), Objects: om, Editor: true} // editor: no rating / payment concerns
	m.Companies[company].HeadquartersX = -1
	x, y, ok := findFlat2x2(m, -1, -1)
	if !ok {
		t.Skip("no flat 2×2")
	}
	a := HeadquarterPlacement{X: int32(x) * 32, Y: int32(y) * 32, Z: int32(m.TileAt(x, y).BaseZ) * 4, Type: uint8(hqObj), BuildImmediately: true}
	if _, ok := m.BuildCompanyHeadquarters(c, CmdApply, a); !ok {
		t.Fatal("building the HQ refused")
	}
	hq := m.Companies[company]
	if int32(hq.HeadquartersX) != a.X || int32(hq.HeadquartersY) != a.Y || int32(hq.HeadquartersZ)*4 != a.Z {
		t.Fatalf("HQ recorded at %d,%d,%d", hq.HeadquartersX, hq.HeadquartersY, hq.HeadquartersZ)
	}
	count := func(tx, ty int) (n int) {
		for k := 0; k < 4; k++ {
			for _, b := range m.TileAt(tx+k%2, ty+k/2).Buildings {
				if int(b.ObjectID) == hqObj {
					n++
				}
			}
		}
		return
	}
	if n := count(x, y); n != 4 {
		t.Fatalf("%d HQ tiles, want 4", n)
	}
	if v := m.TileAt(x, y).Buildings[0].Variation; v != m.headquarterPerformanceVariation(company) {
		t.Errorf("variation %d, want the performance's %d", v, m.headquarterPerformanceVariation(company))
	}
	// Moving it: the old one goes.
	x2, y2, ok := findFlat2x2(m, x, y)
	if !ok {
		t.Skip("no second flat 2×2")
	}
	a2 := a
	a2.X, a2.Y, a2.Z = int32(x2)*32, int32(y2)*32, int32(m.TileAt(x2, y2).BaseZ)*4
	if _, ok := m.BuildCompanyHeadquarters(c, CmdApply, a2); !ok {
		t.Fatal("moving the HQ refused")
	}
	if n := count(x, y); n != 0 {
		t.Errorf("the old HQ left %d tiles", n)
	}
	if n := count(x2, y2); n != 4 || int32(m.Companies[company].HeadquartersX) != a2.X {
		t.Errorf("new HQ: %d tiles, recorded x %d", n, m.Companies[company].HeadquartersX)
	}
	// Removed: no HQ.
	if _, ok := m.RemoveCompanyHeadquarters(c, CmdApply, a2.X, a2.Y, a2.Z); !ok || m.Companies[company].HeadquartersX != -1 || count(x2, y2) != 0 {
		t.Errorf("removal: ok %v, x %d, tiles %d", ok, m.Companies[company].HeadquartersX, count(x2, y2))
	}
}
