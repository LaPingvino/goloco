package ui

import (
	"slices"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
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
	return strings.Replace(f, "{INT16}", Comma16(n), 1)
}

// The filter widgets (dropdownWidgets: filter type 120 × 12, cargo type 150
// × 12) as VehicleList's onResize places them from the window's bottom
// right: cargo_type.right = width − kResizeHandleSize − 2, left = right −
// 150, top = height − 14, bottom = height − 3 (so 151 wide); filter_type to
// its left (right = cargo.left − 2, left = right − 120); each button at
// right − 10 … right − 1, height − 13 … height − 4. They are hidden when the
// window is no wider than the two and 50 (enoughSpaceForCargoFilters).
// The boxes are {left, top, width, 12} (Widget::width: right − left + 1).
func vlBoxes(w *TownListWindow) (filter, cargo [4]int, shown bool) {
	cr := w.W - ResizeHandleSize - 2
	cl := cr - 150
	fr := cl - 2
	fl := fr - 120
	return [4]int{fl, w.H - 14, fr - fl + 1, 12}, [4]int{cl, w.H - 14, cr - cl + 1, 12}, w.W > 150+120+50
}

// vlPress is WinBase.Button for a filter combo box and its button (onResize:
// right − 10 … right − 1, top + 1, 10 high).
func vlPress(w *TownListWindow, box [4]int, px, py int) bool {
	r := box[0] + box[2] - 1
	return w.Button([4]int{r - 10, box[1] + 1, 10, 10}, px, py) || w.Button(box, px, py)
}

// vlDropdown draws a combo box and its button as onResize leaves them (the
// button 10 wide at right − 10).
func vlDropdown(rt *gfx8.RenderTarget, box [4]int, sec []uint8) {
	l, t, r := box[0], box[1], box[0]+box[2]-1
	gfx8.FillRectInset(rt, l, t, r, t+12-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
	gfx8.FillRectInset(rt, r-10, t+1, r-1, t+10, gfx8.InsetColour{Map: sec}, 0)
}

// vlColumns is VehicleList's onResize for the table headers: the name
// column takes what the other three (100, 65, 67 wide as declared) leave,
// at least its own 310 — its right edge set to that width, as upstream does
// — and each next header starts one past the last one's right, its right
// at left + its width; every edge kept within width − 4.
func vlColumns(w *TownListWindow) {
	if len(w.Columns) != 4 {
		return
	}
	nameW := max(310, w.W-100-65-67)
	edge := w.W - 4
	spans := [4][2]int{{4, min(nameW, edge)}}
	for i, cw := range [3]int{100, 65, 67} {
		l := min(spans[i][1]+1, edge)
		spans[i+1] = [2]int{l, min(l+cw, edge)}
	}
	for i, s := range spans {
		c := &w.Columns[i]
		c.HeaderX, c.HeaderW = s[0], s[1]-s[0]+1
		c.TextX, c.TextW = s[0]+1-4, c.HeaderW-2 // drawScroll: header.left + 1, width() − 2
	}
}

// addVehicleListFilters adds VehicleList's filter type and cargo type
// dropdowns: "All vehicles" / "Transports cargo" / "Has cargo order"
// (var_850, af4d69c); with either of the latter, the cargo dropdown (var_852,
// none at first: "No cargo type selected") lists every cargo as "Carrying
// {cargo} {sprite}", and only the vehicles whose cars carry that cargo
// (vehicleTransportsCargo) or with an UnloadAll / WaitFor order for it
// (vehicleHasCargoOrder) stay (populateVehicleList). The rows are refiltered
// and sorted again.
// vlFilterModes are the filter types' texts: all_vehicles, transports_cargo,
// has_cargo_order.
func vlFilterMode(i int) string {
	return i18n.Str([3]int{i18n.AllVehicles, i18n.TransportsCargo, i18n.HasCargoOrder}[i])
}

// vlKeeps is populateVehicleList's filter: every vehicle for all vehicles or
// no cargo chosen; else its cars' cargo (transports cargo) or its orders'
// (has cargo order) must hold the cargo.
func vlKeeps(v VehicleRow, mode, cargo int) bool {
	if mode == 0 || cargo < 0 {
		return true
	}
	list := v.Carried
	if mode == 2 {
		list = v.Cargos
	}
	return slices.Contains(list, cargo)
}

func addVehicleListFilters(w *TownListWindow, tab int, vehicles *[]VehicleRow, all []VehicleRow, cells func(VehicleRow) []string, cargos []VehicleListCargo) {
	mode, cargo := 0, -1 // var_850, var_852 (0xFFFF = -1)
	refilter := func() {
		var out []VehicleRow
		for _, v := range all {
			if !vlKeeps(v, mode, cargo) {
				continue
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
		vlFilterBox, vlCargoBox, shown := vlBoxes(w)
		if !shown {
			return
		}
		vlDropdown(rt, vlFilterBox, sec)
		vlDropdown(rt, vlCargoBox, sec)
		if mode != 0 && cargo >= 0 { // the cargo's inline sprite after its text
			if t, sp := cargoName(); sp > 0 {
				tw, _ := MeasureTextBold(t)
				drawG1At(rt, sp, vlCargoBox[0]+1+10+tw, vlCargoBox[1]+1)
			}
		}
	}
	w.ExtraText = func(screen *ebiten.Image) {
		vlFilterBox, vlCargoBox, shown := vlBoxes(w)
		if !shown {
			return
		}
		g1 := globalG1Font.g1
		pal := g1.GetPalette()
		cmap, _ := g1.GetPaletteMap(int(w.Colour))
		black, win2 := pal[paletteBlack0], pal[gfx8.Shade(cmap, 9)]
		arrow := string(rune(170))
		aw, _ := MeasureTextBold(arrow)
		label := vlFilterMode(mode)
		DrawTextBold(screen, label, w.X+vlFilterBox[0]+1, w.Y+vlFilterBox[1]+1, win2) // wcolour2_stringid
		bx := vlFilterBox[0] + vlFilterBox[2] - 1 - 10                                // the button: right − 10 … right − 1, its text at (10 − 1) / 2 (Label: centred)
		DrawTextBold(screen, arrow, w.X+bx+(10-1)/2-aw/2, w.Y+vlFilterBox[1]+1, black)
		bx = vlCargoBox[0] + vlCargoBox[2] - 1 - 10
		if mode == 0 { // disabled (prepareDraw): no text, the button's arrow in colour.inset()
			DrawTextInset(screen, arrow, w.X+bx+(10-1)/2-aw/2, w.Y+vlCargoBox[1]+1, cmap)
			return
		}
		t, _ := cargoName()
		x := w.X + vlCargoBox[0] + 1
		if cargo >= 0 {
			x += 10 // carrying_cargoid_sprite's {MOVE_X 10}
		}
		DrawTextBold(screen, t, x, w.Y+vlCargoBox[1]+1, win2)
		DrawTextBold(screen, arrow, w.X+bx+(10-1)/2-aw/2, w.Y+vlCargoBox[1]+1, black)
	}
	w.ExtraClick = func(px, py int) bool {
		vlFilterBox, vlCargoBox, shown := vlBoxes(w)
		if w.OpenDropdown == nil || !shown {
			return false
		}
		// filter_type_btn: Dropdown::show at the widget, width() − 4.
		if vlPress(w, vlFilterBox, px, py) {
			d := NewDropdownShow(w.X+vlFilterBox[0], w.Y+vlFilterBox[1], vlFilterBox[2]-4, 11, w.Colour,
				[]DropdownItem{{Text: vlFilterMode(0), Selected: mode == 0, MoveX10: true}, {Text: vlFilterMode(1), Selected: mode == 1, MoveX10: true},
					{Text: vlFilterMode(2), Selected: mode == 2, MoveX10: true}}, 1<<20, 1<<20)
			d.OnSelect = func(i int) {
				if i >= 0 && i != mode {
					mode, cargo = i, -1
					refilter()
				}
			}
			w.OpenDropdown(d)
			return true
		}
		// cargo_type_btn (enabled but for all vehicles): Dropdown::showText,
		// every cargo "Carrying {name} {sprite}"; the selected test compares
		// the item index with var_852 (the cargo id) as upstream does.
		if mode != 0 && vlPress(w, vlCargoBox, px, py) && len(cargos) > 0 { // all vehicles: disabled
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
