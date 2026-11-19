package sim

import (
	"strconv"
	"strings"
	"testing"

	"github.com/LaPingvino/goloco/pkg/assets"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// checkIndustryColours (0x0046D0C3) drops every kIndustryColours entry
// within 2 of the colour, either side.
func TestCheckIndustryColours(t *testing.T) {
	got := checkIndustryColours(palBlack0+1, 0x7FFFFFFF) // black0 (0) within 1; black4 (1) 3 away
	if got&1 != 0 || got&2 == 0 {
		t.Errorf("black1: mask %#x", got)
	}
	got = checkIndustryColours(palMutedPurple7, 0x7FFFFFFF) // mutedPurple7 is entry 14
	if got != 0x7FFFFFFF&^(1<<14) {
		t.Errorf("mutedPurple7: mask %#x", got)
	}
}

func mapTabsFill(t *testing.T) (*Map, *MapFill) {
	m, om := loadDataMap(t, "title.dat")
	g1, ok := om.G1File.(*assets.G1File)
	if !ok {
		t.Skip("g1 not present")
	}
	pixel := func(img uint32, i int) uint8 { return g1.Elements[img].Data[i] }
	shade := func(colour uint8, s int) uint8 { return colour*16 + uint8(s) } // a recognisable stand-in
	c := &MapColours{}
	c.AssignIndustryColours(om, pixel)
	m.AssignRouteColours(c, om, pixel)
	return m, &MapFill{OM: om, Pixel: pixel, Shade: shade, Colours: c}
}

// The colour tables on the title game: every loaded industry object a
// distinct colour, its preferred one when no other took it and it is not
// near the land's; the route list's types, then aircraft, ships, the end.
func TestAssignMapColours(t *testing.T) {
	m, f := mapTabsFill(t)
	c, om := f.Colours, f.OM
	seen := map[uint8]int{}
	for i := range c.Industry {
		io := om.GetIndustryObjectByIndex(i)
		if io == nil {
			if c.Industry[i] != 0xFF {
				t.Errorf("industry %d: none loaded but colour %d", i, c.Industry[i])
			}
			continue
		}
		if j, dup := seen[c.Industry[i]]; dup {
			t.Errorf("industries %d and %d share colour %d", j, i, c.Industry[i])
		}
		seen[c.Industry[i]] = i
		if checkIndustryColours(MapIndustryColours[c.Industry[i]], 1<<c.Industry[i])&(1<<c.Industry[i]) != 0 {
			t.Errorf("industry %d: colour %d not its own entry", i, c.Industry[i])
		}
	}
	n := 0
	for c.RouteToObjectID[n] != MapRouteAircraft {
		n++
	}
	want := len(m.companyAvailableRailTracks(m.ControllingCompany, om)) + len(m.companyAvailableRoads(m.ControllingCompany, om))
	if n != want || c.RouteToObjectID[n+1] != MapRouteShips || c.RouteToObjectID[n+2] != MapRouteEnd {
		t.Errorf("route list %v: %d types, want %d", c.RouteToObjectID, n, want)
	}
	if c.Route[n] != 0xD3 || c.Route[n+1] != 0x8B {
		t.Errorf("aircraft / ships colours %#x %#x", c.Route[n], c.Route[n+1])
	}
	for i := range n {
		id := c.RouteToObjectID[i]
		tab := c.Track[id&7]
		if id&0x80 != 0 {
			tab = c.Road[id&7]
		}
		if tab != c.Route[i] {
			t.Errorf("route %d (id %#x): table colour %#x, list %#x", i, id, tab, c.Route[i])
		}
	}
}

// Each tab's colours for the kinds of tile it marks, on the title map.
func TestMapPixelsTabs(t *testing.T) {
	m, f := mapTabsFill(t)
	var land, building, industry, track, owned bool
	for ty := 1; ty < m.Height-1; ty++ {
		for tx := 1; tx < m.Width-1; tx++ {
			tl := m.TileAt(tx, ty)
			only := len(tl.Order) == 1
			switch {
			case !land && only && tl.WaterLevel == 0:
				lo := f.OM.GetLandObjectByIndex(int(tl.TerrainIndex))
				p0, p1 := f.Pixel(lo.MapPixelImage, 0), f.Pixel(lo.MapPixelImage, 1)
				for tab, want := range map[int][2]uint8{MapTabVehicles: {p0, p0}, MapTabIndustries: {p0, p0}, MapTabRoutes: {p0, p1}, MapTabOwnership: {p0, p1}} {
					if c0, c1, _, _, _ := m.MapPixelsTab(tab, tx, ty, f); c0 != want[0] || c1 != want[1] {
						t.Errorf("tab %d land %d,%d: %#x %#x, want %v", tab, tx, ty, c0, c1, want)
					}
				}
				land = true
			case !building && len(tl.Buildings) == 1 && len(tl.Order) == 2:
				for tab, want := range map[int]uint8{MapTabVehicles: palMutedDarkRed2, MapTabIndustries: palMutedDarkRed2, MapTabRoutes: palMutedDarkRed2, MapTabOwnership: palBlack1} {
					if c0, c1, _, _, _ := m.MapPixelsTab(tab, tx, ty, f); c0 != want || c1 != want {
						t.Errorf("tab %d building %d,%d: %#x %#x, want %#x", tab, tx, ty, c0, c1, want)
					}
				}
				building = true
			case !industry && len(tl.Industries) > 0 && tl.Order[len(tl.Order)-1].Kind == scenario.KindIndustry:
				ind := m.industryByID(tl.Industries[len(tl.Industries)-1].IndustryID)
				want := MapIndustryColours[f.Colours.Industry[ind.ObjectID]]
				f.Flashing = 1 << ind.ObjectID
				c0, c1, f0, f1, _ := m.MapPixelsTab(MapTabIndustries, tx, ty, f)
				f.Flashing = 0
				if c0 != want || c1 != want || f0 != palBlack0 || f1 != palBlack0 {
					t.Errorf("industry %d,%d: %#x %#x flash %#x %#x, want %#x", tx, ty, c0, c1, f0, f1, want)
				}
				industry = true
			case !track && len(tl.Tracks) == 1 && tl.Order[len(tl.Order)-1].Kind == scenario.KindTrack && tl.Order[len(tl.Order)-1].Flags&(elemFlagGhost|elemFlagAiAllocated) == 0:
				id := tl.Tracks[0].TrackObjectID
				row := -1
				for i, r := range f.Colours.RouteToObjectID {
					if r == id {
						row = i
						break
					}
				}
				if row < 0 {
					continue // a track type the player can't build: no key row
				}
				f.Flashing = 1 << row
				c0, c1, f0, _, _ := m.MapPixelsTab(MapTabRoutes, tx, ty, f)
				f.Flashing = 0
				if want := f.Colours.Track[id]; c0 != want || c1 != want || f0 != mapFlashColour(want) {
					t.Errorf("routes track %d,%d: %#x %#x flash %#x, want %#x", tx, ty, c0, c1, f0, want)
				}
				if c0, _, _, _, _ := m.MapPixelsTab(MapTabVehicles, tx, ty, f); c0 != palBlack2 {
					t.Errorf("vehicles track %d,%d: %#x", tx, ty, c0)
				}
				track = true
			case !owned && len(tl.Tracks) == 1 && tl.Order[len(tl.Order)-1].Kind == scenario.KindTrack && tl.Tracks[0].Owner < 15:
				o := tl.Tracks[0].Owner
				f.Flashing = 1 << o
				c0, c1, f0, f1, _ := m.MapPixelsTab(MapTabOwnership, tx, ty, f)
				f.Flashing = 0
				want := f.Shade(m.CompanyColours[o], 5)
				if c0 != want || c1 != want || f0 != mapFlashColour(want) || f1 != f0 {
					t.Errorf("ownership %d,%d: %#x %#x flash %#x %#x, want %#x", tx, ty, c0, c1, f0, f1, want)
				}
				owned = true
			}
		}
	}
	if !land || !building || !industry || !track || !owned {
		t.Errorf("not every kind met: land %v building %v industry %v track %v owned %v", land, building, industry, track, owned)
	}
}

// drawVehiclesOnMap's marks: a pixel per placed car component in the tab's
// colour (blackB on the overall tab, the type's on the vehicles tab), the
// count by type there only.
func TestMapVehicles(t *testing.T) {
	m, f := mapTabsFill(t)
	shade := f.Shade
	marks, _, counted := m.MapVehicles(MapTabOverall, f.Colours, 0, 0, 0, shade)
	if counted || len(marks) == 0 {
		t.Fatalf("overall: counted %v, %d marks", counted, len(marks))
	}
	for _, mk := range marks {
		if mk.Line || mk.Colour != palBlackB {
			t.Fatalf("overall mark %+v", mk)
		}
	}
	marks, counts, counted := m.MapVehicles(MapTabVehicles, f.Colours, 0, 0, 0, shade)
	var total int
	for _, n := range counts {
		total += int(n)
	}
	trains, _ := m.Trains()
	var want int
	for _, tr := range trains {
		if h := m.Entities[tr.Head]; h.Flags38&flags38IsGhost == 0 && h.PosX != locationNull {
			want++
		}
	}
	if !counted || total != want || want == 0 {
		t.Errorf("vehicles: counted %v, %d (%v), want %d", counted, total, counts, want)
	}
	types := map[uint8]bool{}
	for _, c := range MapVehicleTypeColours {
		types[c] = true
	}
	for _, mk := range marks {
		if !types[mk.Colour] {
			t.Fatalf("vehicles mark colour %#x", mk.Colour)
		}
	}
	// A hovered type flashes on the frames with bit 2.
	marks, _, _ = m.MapVehicles(MapTabVehicles, f.Colours, 1<<0, 4, 0, shade)
	for _, mk := range marks {
		if mk.Colour == palRed7 {
			t.Fatalf("a train still red while flashing")
		}
	}
}

// formatIndustryString / formatVehicleString: the totals, a hovered row's
// count by its name with its cargo in brackets.
func TestMapStatusStrings(t *testing.T) {
	m, f := mapTabsFill(t)
	s := m.MapIndustryStatus(0, f.OM)
	n := 0
	for k := range m.Industries {
		if m.Industries[k].Name != 0xFFFF {
			n++
		}
	}
	if want := strconv.Itoa(n) + " industries"; s != want {
		t.Errorf("total: %q, want %q", s, want)
	}
	obj := m.Industries[0].ObjectID
	s = m.MapIndustryStatus(1<<obj, f.OM)
	if !strings.Contains(s, " (") || !strings.HasSuffix(s, ")") || strings.Contains(s, "{") {
		t.Errorf("industry %d: %q", obj, s)
	}
	t.Logf("industry %d: %q", obj, s)
	if s := MapVehicleStatus([6]uint16{1, 2}, 0); s != "3 vehicles" {
		t.Errorf("vehicles: %q", s)
	}
	if s := MapVehicleStatus([6]uint16{1, 2}, 1); s != "1 train" {
		t.Errorf("trains: %q", s)
	}
	if s := MapVehicleStatus([6]uint16{1, 2}, 2); s != "2 buses" {
		t.Errorf("buses: %q", s)
	}
}
