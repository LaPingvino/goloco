package sim

import "testing"

// A flat, empty, dry 2×2 patch on the title map.
func findDryPatch(t *testing.T, m *Map) (int, int) {
	t.Helper()
	for y := 8; y < m.Height-8; y++ {
		for x := 8; x < m.Width-8; x++ {
			ok := true
			z := m.TileAt(x, y).BaseZ
			for dy := 0; dy < 2 && ok; dy++ {
				for dx := 0; dx < 2; dx++ {
					tl := m.TileAt(x+dx, y+dy)
					if len(tl.Order) != 1 || tl.Slope != 0 || tl.WaterLevel != 0 || tl.BaseZ != z || tl.IsIndustrial {
						ok = false
						break
					}
				}
			}
			if ok {
				return x, y
			}
		}
	}
	t.Skip("no dry flat patch")
	return 0, 0
}

// Raised once, a dry flat patch gets water a step above its land, at the
// water object's cost per tile; lowered again it's dry.
func TestRaiseLowerWater(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	x, y := findDryPatch(t, m)
	z := m.TileAt(x, y).BaseZ
	m.TakeSounds()
	dry, ok := m.RaiseWater(c, 0, x, y, x+1, y+1)
	if !ok || dry <= 0 || m.TileAt(x, y).WaterLevel != 0 {
		t.Fatalf("dry run: ok %v cost %d", ok, dry)
	}
	if s := m.TakeSounds(); len(s) != 0 {
		t.Errorf("a dry run's sound: %+v", s)
	}
	cost, ok := m.RaiseWater(c, CmdApply, x, y, x+1, y+1)
	if !ok || cost != dry {
		t.Fatalf("raise: ok %v cost %d (dry %d)", ok, cost, dry)
	}
	// the water sound at the range's centre (raiseWater, applied and paid)
	if s := m.TakeSounds(); len(s) != 1 || s[0].ID != soundWater || s[0].X != int32((2*x+1)*tileSize/2) || s[0].Y != int32((2*y+1)*tileSize/2) {
		t.Errorf("raise's sound: %+v", s)
	}
	one := m.InflationAdjustedCost(om.WaterObj.CostFactor, om.WaterObj.CostIndex, 10)
	if cost != 4*one {
		t.Errorf("cost %d, want 4 × %d", cost, one)
	}
	for dy := range 2 {
		for dx := range 2 {
			if w := m.TileAt(x+dx, y+dy).WaterLevel; int(w)*microToSmallZStep != int(z)+smallZStep {
				t.Errorf("tile %d,%d water %d, want %d", x+dx, y+dy, w, (int(z)+smallZStep)/microToSmallZStep)
			}
		}
	}
	if _, ok := m.LowerWater(c, CmdApply, x, y, x+1, y+1); !ok {
		t.Fatal("lower refused")
	}
	if s := m.TakeSounds(); len(s) != 1 || s[0].ID != soundWater || c.Pos == nil || c.Pos[0] != s[0].X+tileSize/2 {
		t.Errorf("lower's sound %+v, position %v", s, c.Pos)
	}
	for dy := range 2 {
		for dx := range 2 {
			if w := m.TileAt(x+dx, y+dy).WaterLevel; w != 0 {
				t.Errorf("tile %d,%d still has water %d", x+dx, y+dy, w)
			}
		}
	}
}

// Water at the ships' channel can't be changed.
func TestAdjustWaterShipChannel(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	x, y := findDryPatch(t, m)
	m.TileAt(x, y).Type6Flag = true
	if _, ok := m.RaiseWater(c, 0, x, y, x, y); ok {
		t.Error("a ship route's tile took water")
	}
}

func TestChangeLandMaterial(t *testing.T) {
	m, om := loadDataMap(t, "title.dat")
	c := &Cmd{Company: CompanyNeutral, Objects: om}
	x, y := findDryPatch(t, m)
	land := -1
	for i, lo := range om.LandObjects {
		if lo != nil && uint8(i) != m.TileAt(x, y).TerrainIndex {
			land = i
			break
		}
	}
	if land < 0 {
		t.Skip("one land object only")
	}
	if _, ok := m.ChangeLandMaterial(c, 0, x, y, x+1, y+1, uint8(land)); !ok || m.TileAt(x, y).TerrainIndex == uint8(land) {
		t.Fatal("the dry run changed the land")
	}
	m.ChangeLandMaterial(c, CmdApply, x, y, x+1, y+1, uint8(land))
	want := om.LandObjects[land].NumGrowthStages - 1
	for dy := range 2 {
		for dx := range 2 {
			tl := m.TileAt(x+dx, y+dy)
			if tl.TerrainIndex != uint8(land) || tl.GrowthStage != want {
				t.Errorf("tile %d,%d land %d growth %d, want %d / %d", x+dx, y+dy, tl.TerrainIndex, tl.GrowthStage, land, want)
			}
		}
	}
}
