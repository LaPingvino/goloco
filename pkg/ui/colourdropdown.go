package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// ColourDropdown is Dropdown::showColour (Dropdown.cpp): a grid of colour
// swatches (16 × 16 cells, getItemsPerRow columns), each available colour's
// raised swatch, the highlighted one pressed (its image + 1) under a
// translucent ExtColour::unk2E cell. A click picks the colour under it
// (dropdownIndexFromPoint); a click elsewhere closes it.
type ColourDropdown struct {
	X, Y, W, H  int
	Colour      uint8   // the dropdown's colour (the window's secondary)
	Colours     []uint8 // the available colours, in order
	Columns     int
	Highlighted int
	OnSelect    func(colour uint8)
	Visible     bool
}

const (
	imgColourSwatchRaised = 2307 // ImageIds::colour_swatch_recolourable_raised
	colourDropdownCell    = 16
)

// kAppropriateImageDropdownItemsPerRow.
var imageDropdownItemsPerRow = [33]int{1, 1, 1, 1, 2, 2, 3, 3, 4, 3, 5, 4, 4, 5, 5, 5, 4, 5, 6, 5, 5, 7, 4, 5, 6, 5, 6, 6, 6, 6, 6, 8, 8}

// NewColourDropdown is showColour for a widget at (wx, wy) of height wh on a
// sw × sh screen: the colours set in available (bit per colour), the
// selected one highlighted; showImage's placement (below the widget, else
// above it, else beside it at the top; kept on screen).
func NewColourDropdown(wx, wy, wh int, available uint32, selected, colour uint8, sw, sh int) *ColourDropdown {
	d := &ColourDropdown{Colour: colour, Highlighted: -1, Visible: true}
	for i := 0; i < 32; i++ {
		if available&(1<<i) != 0 {
			if uint8(i) == selected {
				d.Highlighted = len(d.Colours)
			}
			d.Colours = append(d.Colours, uint8(i))
		}
	}
	n := len(d.Colours)
	d.Columns = imageDropdownItemsPerRow[min(n, 32)]
	rows := n / d.Columns
	if n%d.Columns != 0 {
		rows++
	}
	d.W = colourDropdownCell*d.Columns + 3 + 1
	d.H = colourDropdownCell*rows + 3 + 1
	heightOffset := wh + 1
	x, y := wx, wy+heightOffset
	if y+d.H > sh || y < 0 {
		y -= heightOffset + d.H
		bottom := y
		if y >= 0 {
			bottom = y + d.H
		}
		if y < 0 || bottom > sh {
			x += d.W - 1
			y = 0
		}
	}
	x = max(x, 0)
	x = min(x+d.W, sw) - d.W
	d.X, d.Y = x, y
	return d
}

// indexAt is dropdownIndexFromPoint: the cell under a point, −1 none.
func (d *ColourDropdown) indexAt(x, y int) int {
	left, top := x-d.X, y-d.Y-2
	if left < 0 || left >= d.W || top < 0 {
		return -1
	}
	left -= 2
	if left < 0 {
		return -1
	}
	col, row := left/colourDropdownCell, top/colourDropdownCell
	if col >= d.Columns {
		return -1
	}
	if i := row*d.Columns + col; i < len(d.Colours) {
		return i
	}
	return -1
}

// UpdateHover highlights the cell under the pointer.
func (d *ColourDropdown) UpdateHover(x, y int) {
	if i := d.indexAt(x, y); i >= 0 {
		d.Highlighted = i
	}
}

// HandleClick picks the colour clicked (true) or closes the dropdown.
func (d *ColourDropdown) HandleClick(x, y int) bool {
	if !d.Visible {
		return false
	}
	i := d.indexAt(x, y)
	d.Visible = false
	if i < 0 {
		return x >= d.X && x < d.X+d.W && y >= d.Y && y < d.Y+d.H
	}
	if d.OnSelect != nil {
		d.OnSelect(d.Colours[i])
	}
	return true
}

// Draw draws the panel and the swatches.
func (d *ColourDropdown) Draw(screen *ebiten.Image) {
	if !d.Visible || globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	cmap, err := g1.GetPaletteMap(int(d.Colour))
	if err != nil {
		return
	}
	highlight, _ := g1.GetPaletteMap(extColourUnk2E)
	draw8(screen, d.X, d.Y, d.W, d.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 0, d.W-1, d.H-1, gfx8.InsetColour{Map: cmap}, 0) // the Wt3 panel
		for i, c := range d.Colours {
			cx, cy := (i%d.Columns)*colourDropdownCell+2, (i/d.Columns)*colourDropdownCell+2
			img := imgColourSwatchRaised
			if i == d.Highlighted {
				if highlight != nil {
					gfx8.FillRectTranslucent(rt, cx, cy, cx+colourDropdownCell-1, cy+colourDropdownCell-1, highlight)
				}
				img++ // the pressed swatch
			}
			if img < len(g1.Elements) {
				if sp, err := gfx8.Decode(&g1.Elements[img]); err == nil {
					m, _ := g1.GetPaletteMap(int(c))
					gfx8.DrawSpriteNoise(rt, sp, cx, cy, gfx8.OpFor(sp, m != nil, false), m, nil)
				}
			}
		}
	})
}
