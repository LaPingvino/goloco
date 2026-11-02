package ui

import (
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// VehicleListCargo is one cargo object for the cargo filter: its id (object
// slot), name and unitInlineSprite (a G1 image index).
type VehicleListCargo struct {
	ID     int
	Name   string
	Sprite int
}

// vehicleListFooters are kTypeToFooterStringIds (num_trains_singular /
// _plural …) per tab.
var vehicleListFooters = [6][2]string{
	{"{INT16} train", "{INT16} trains"}, {"{INT16} bus", "{INT16} buses"}, {"{INT16} truck", "{INT16} trucks"},
	{"{INT16} tram", "{INT16} trams"}, {"{INT16} aircraft", "{INT16} aircraft"}, {"{INT16} ship", "{INT16} ships"},
}

// vehicleListFooter is prepareDraw's status bar: the singular for one row.
func vehicleListFooter(tab, n int) string {
	f := vehicleListFooters[tab%6][1]
	if n == 1 {
		f = vehicleListFooters[tab%6][0]
	}
	return strings.Replace(f, "{INT16}", itoa3(n), 1)
}

// The filter widgets (dropdownWidgets): filter type at (264, 200) 120×12,
// cargo type at (386, 200) 150×12.
var (
	vlFilterBox = [4]int{280 - 16, 200, 120, 12}
	vlCargoBox  = [4]int{402 - 16, 200, 150, 12}
)

// addVehicleListFilters adds VehicleList's filter type and cargo type
// dropdowns: "All vehicles" / "Transporting cargo" (var_850); with the
// latter, the cargo dropdown (var_852, none at first: "No cargo type
// selected") lists every cargo as "Carrying {cargo} {sprite}", and only
// vehicles with an UnloadAll / WaitFor order for that cargo stay
// (populateVehicleList). The rows are refiltered and sorted again.
func addVehicleListFilters(w *TownListWindow, tab int, vehicles *[]VehicleRow, all []VehicleRow, cells func(VehicleRow) []string, cargos []VehicleListCargo) {
	mode, cargo := 0, -1 // var_850, var_852 (0xFFFF = -1)
	refilter := func() {
		var out []VehicleRow
		for _, v := range all {
			if mode == 1 && cargo >= 0 { // isCargoFilterActive
				ok := false
				for _, c := range v.Cargos {
					if c == cargo {
						ok = true
					}
				}
				if !ok {
					continue
				}
			}
			out = append(out, v)
		}
		*vehicles = out
		w.Rows = make([][]string, len(out))
		for i, v := range out {
			w.Rows[i] = cells(v)
		}
		w.Status = vehicleListFooter(tab, len(out))
		w.Hover = -1
		w.Sort(w.SortMode)
	}
	cargoName := func() (string, int) {
		for _, c := range cargos {
			if c.ID == cargo {
				return "Carrying " + c.Name + " ", c.Sprite
			}
		}
		return "No cargo type selected", 0
	}
	w.Extra8 = func(rt *gfx8.RenderTarget, sec []uint8) {
		optDropdown(rt, vlFilterBox[0], vlFilterBox[1], vlFilterBox[2], sec)
		optDropdown(rt, vlCargoBox[0], vlCargoBox[1], vlCargoBox[2], sec)
		if mode == 1 && cargo >= 0 { // the cargo's inline sprite after its text
			if t, sp := cargoName(); sp > 0 {
				tw, _ := MeasureTextBold(t)
				drawG1At(rt, sp, vlCargoBox[0]+1+10+tw, vlCargoBox[1]+1)
			}
		}
	}
	w.ExtraText = func(screen *ebiten.Image) {
		g1 := globalG1Font.g1
		pal := g1.GetPalette()
		cmap, _ := g1.GetPaletteMap(int(w.Colour))
		black, win2 := pal[paletteBlack0], pal[gfx8.Shade(cmap, 9)]
		arrow := string(rune(170))
		aw, _ := MeasureTextBold(arrow)
		label := [2]string{"All vehicles", "Transporting cargo"}[mode]
		DrawTextBold(screen, label, w.X+vlFilterBox[0]+1, w.Y+vlFilterBox[1]+1, win2) // wcolour2_stringid
		bx := vlFilterBox[0] + vlFilterBox[2] - 12
		DrawTextBold(screen, arrow, w.X+bx+(11-1)/2-aw/2, w.Y+vlFilterBox[1]+1, black)
		bx = vlCargoBox[0] + vlCargoBox[2] - 12
		if mode != 1 { // disabled (prepareDraw): no text, the button inset
			c6, c4 := pal[gfx8.Shade(cmap, 6)], pal[gfx8.Shade(cmap, 4)]
			DrawTextBold(screen, arrow, w.X+bx+(11-1)/2-aw/2+1, w.Y+vlCargoBox[1]+2, c4)
			DrawTextBold(screen, arrow, w.X+bx+(11-1)/2-aw/2, w.Y+vlCargoBox[1]+1, c6)
			return
		}
		t, _ := cargoName()
		x := w.X + vlCargoBox[0] + 1
		if cargo >= 0 {
			x += 10 // carrying_cargoid_sprite's {MOVE_X 10}
		}
		DrawTextBold(screen, t, x, w.Y+vlCargoBox[1]+1, win2)
		DrawTextBold(screen, arrow, w.X+bx+(11-1)/2-aw/2, w.Y+vlCargoBox[1]+1, black)
	}
	w.ExtraClick = func(px, py int) bool {
		if w.OpenDropdown == nil {
			return false
		}
		// filter_type_btn: Dropdown::show at the widget, width() − 4.
		if w.in(vlFilterBox, px, py) {
			d := NewDropdownShow(w.X+vlFilterBox[0], w.Y+vlFilterBox[1], vlFilterBox[2]-4, 11, w.Colour,
				[]DropdownItem{{Text: "All vehicles", Selected: mode == 0, MoveX10: true}, {Text: "Transporting cargo", Selected: mode == 1, MoveX10: true}}, 1<<20, 1<<20)
			d.OnSelect = func(i int) {
				if i >= 0 && i != mode {
					mode, cargo = i, -1
					refilter()
				}
			}
			w.OpenDropdown(d)
			return true
		}
		// cargo_type_btn (enabled with transporting cargo): Dropdown::showText,
		// every cargo "Carrying {name} {sprite}"; the selected test compares
		// the item index with var_852 (the cargo id) as upstream does.
		if mode == 1 && w.in(vlCargoBox, px, py) && len(cargos) > 0 {
			items := make([]DropdownItem, len(cargos))
			for i, c := range cargos {
				items[i] = DropdownItem{Text: "Carrying " + c.Name + " ", MoveX10: true, TrailSprite: c.Sprite, Selected: i == cargo}
			}
			d := NewDropdownShow(w.X+vlCargoBox[0], w.Y+vlCargoBox[1], vlCargoBox[2]-4, 11, w.Colour, items, 1<<20, 1<<20)
			d.OnSelect = func(i int) {
				if i >= 0 && i < len(cargos) {
					cargo = cargos[i].ID // getItemArgument(itemIndex, 3): the cargo id
					refilter()
				}
			}
			w.OpenDropdown(d)
			return true
		}
		return false
	}
}

// drawG1At draws a G1 image at (x, y) with its own offsets ({SPRITE}).
func drawG1At(rt *gfx8.RenderTarget, id, x, y int) {
	g1 := globalG1Font.g1
	if id <= 0 || id >= len(g1.Elements) {
		return
	}
	if sp, err := gfx8.Decode(&g1.Elements[id]); err == nil {
		gfx8.DrawSpriteNoise(rt, sp, x, y, gfx8.OpFor(sp, false, false), nil, nil)
	}
}
