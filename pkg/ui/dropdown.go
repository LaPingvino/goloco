package ui

import (
	"image"
	"image/color"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// DropdownMenu represents a dropdown menu window
//
// OpenLoco reference: src/OpenLoco/src/Ui/Dropdown.cpp
//
//	Dropdown menus are temporary windows that appear when clicking toolbar buttons.
type DropdownMenu struct {
	X         int
	Y         int
	Width     int
	Height    int
	Items     []DropdownItem
	ItemWidth int // _dropdownItemWidth
	// Colour is the window colour; Translucent as AdvancedColour::translucent.
	// Toolbar dropdowns: SetToolbarColour.
	Colour      uint8
	Translucent bool
	Highlighted int
	OnSelect    func(index int)
	Visible     bool
}

// DropdownItem represents a single item in a dropdown menu
type DropdownItem struct {
	Text      string
	Disabled  bool
	Selected  bool // Shows checkmark
	Separator bool // Is this a separator line?
	Action    func()
}

// Dropdown geometry (Dropdown.cpp common::showText): items 10 px high
// (_dropdownItemHeight default); _dropdownItemWidth = the widest item in
// medium bold + 3; the window is that + 3 + 1 wide and items×10 + 3 + 1 high,
// framed by its Wt3 panel widget (fillRectInset, raised).
const (
	dropdownItemHeight = 10
	// DropdownColour is the default (opaque) window colour of a dropdown.
	DropdownColour = 1 // Colour::grey
)

// NewDropdownMenu creates a dropdown below a widget: x, y is the widget's
// bottom-left (showBelow passes the widget top + height).
//
// OpenLoco reference: src/OpenLoco/src/Ui/Dropdown.cpp showText / showBelow.
// Place applies showText's placement for a widget and screen size.
func NewDropdownMenu(x, y int, items []DropdownItem) *DropdownMenu {
	maxW := 0
	for _, it := range items {
		if it.Separator {
			continue
		}
		if w, _ := MeasureTextBold(it.Text); w > maxW {
			maxW = w
		}
	}
	itemWidth := maxW + 3
	return &DropdownMenu{
		X: x, Y: y,
		Width:       itemWidth + 3 + 1,
		Height:      dropdownItemHeight*len(items) + 3 + 1,
		ItemWidth:   itemWidth,
		Colour:      DropdownColour,
		Items:       items,
		Highlighted: -1,
		Visible:     true,
	}
}

// Place is the placement half of common::showText for the widget at wx, wy
// (wy its top) of size ww × wh on a screen sw × sh: below the widget; if that
// leaves the screen, above it; if that fails too, beside it at the top; then x
// kept on screen. (maxStringWidth there already includes its +3: ItemWidth.)
func (d *DropdownMenu) Place(wx, wy, ww, wh, sw, sh int) {
	d.X, d.Y = uilayout.DropdownPlace(wx, wy, ww, wh, d.ItemWidth, d.Width, d.Height, sw, sh)
}

// SetToolbarColour is showBelow's colour for a toolbar button: the button's
// window colour, translucent, mapped through kDropdownColourTable.
func (d *DropdownMenu) SetToolbarColour(windowColour uint8) {
	if int(windowColour) < len(DropdownColourTable) {
		d.Colour = DropdownColourTable[windowColour]
	}
	d.Translucent = true
}

// Draw renders the dropdown (Dropdown.cpp draw) with the gfx8 window
// primitives: an opaque dropdown in its own 8-bit target, a translucent one on
// the screen region itself (draw8), so it remaps what is behind it as upstream.
func (d *DropdownMenu) Draw(screen *ebiten.Image) {
	if !d.Visible || globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	cmap, err := g1.GetPaletteMap(int(d.Colour))
	if err != nil {
		return
	}
	pal := g1.GetPalette()
	highlight, _ := g1.GetPaletteMap(extColourUnk2E)
	var trans [3][]uint8
	if d.Translucent {
		for k := range trans {
			trans[k], _ = g1.GetPaletteMap(int(paint.TranslucentColour[d.Colour][k]))
		}
	}
	type label struct {
		text  string
		x, y  int
		color color.Color
		inset color.Color // non-nil: inset text (drawn first, offset (-1,-1) from color)
	}
	var labels []label
	draw := func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 0, d.Width-1, d.Height-1, gfx8.InsetColour{Map: cmap, Translucent: d.Translucent, Trans: trans}, 0) // the Wt3 panel
		for i, it := range d.Items {
			if it.Separator {
				// Two lines at mid-row, item width - 1 wide: shades 3 and 7, or for a
				// translucent colour getTranslucent(c)+1 (highlight) then +2 (shadow).
				sy := 1 + i*dropdownItemHeight + dropdownItemHeight/2
				if d.Translucent {
					base := int(paint.TranslucentColour[d.Colour][1])
					m1, _ := g1.GetPaletteMap(base + 1)
					m2, _ := g1.GetPaletteMap(base + 2)
					gfx8.FillRectTranslucent(rt, 2, sy, 2+d.ItemWidth-2, sy, m1)
					gfx8.FillRectTranslucent(rt, 2, sy+1, 2+d.ItemWidth-2, sy+1, m2)
				} else {
					gfx8.FillRect(rt, 2, sy, 2+d.ItemWidth-2, sy, gfx8.Shade(cmap, 3))
					gfx8.FillRect(rt, 2, sy+1, 2+d.ItemWidth-2, sy+1, gfx8.Shade(cmap, 7))
				}
				continue
			}
			// TextRenderer: a plain colour draws glyphs in getShade(colour, 9)
			// (setTextColours(shade 9, black0, black0), no outline).
			var c color.Color = pal[gfx8.Shade(cmap, 9)]
			if i == d.Highlighted {
				gfx8.FillRectTranslucent(rt, 2, 2+i*dropdownItemHeight, 2+d.ItemWidth-1, 2+i*dropdownItemHeight+dropdownItemHeight-1, highlight)
				if wm, err := g1.GetPaletteMap(colourWhite); err == nil {
					c = pal[gfx8.Shade(wm, 9)] // Colour::white through the same rule
				}
			}
			var inset color.Color
			if it.Disabled {
				// colour.inset(): each glyph drawn solid in shade 6, then shade 4 at
				// (+1,+1) (TextRenderer inset, font without the dark flag).
				inset, c = pal[gfx8.Shade(cmap, 6)], pal[gfx8.Shade(cmap, 4)]
			}
			labels = append(labels, label{it.Text, 2, 1 + i*dropdownItemHeight, c, inset})
		}
	}
	if d.Translucent {
		draw8(screen, d.X, d.Y, d.Width, d.Height, draw)
	} else {
		rt := gfx8.NewRenderTarget(d.Width, d.Height)
		draw(rt)
		img := image.NewRGBA(image.Rect(0, 0, d.Width, d.Height))
		for i, p := range rt.Bits {
			img.Set(i%d.Width, i/d.Width, pal[p])
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(d.X), float64(d.Y))
		screen.DrawImage(ebiten.NewImageFromImage(img), op)
	}
	for _, l := range labels {
		if l.inset != nil {
			DrawTextBold(screen, l.text, d.X+l.x, d.Y+l.y, l.inset)
			DrawTextBold(screen, l.text, d.X+l.x+1, d.Y+l.y+1, l.color)
			continue
		}
		DrawTextBold(screen, l.text, d.X+l.x, d.Y+l.y, l.color)
	}
}

// colourWhite is Colour::white.
const colourWhite = 2

// extColourUnk2E is ExtColour::unk2E, the dropdown's highlight translucency.
const extColourUnk2E = 0x2E

// HandleClick processes a click on the dropdown menu
// Returns true if the click was handled
func (d *DropdownMenu) HandleClick(x, y int) bool {
	if !d.Visible {
		return false
	}

	// Check if click is inside dropdown
	if x < d.X || x >= d.X+d.Width || y < d.Y || y >= d.Y+d.Height {
		// Click outside - close dropdown
		d.Visible = false
		return false
	}

	const itemHeight = dropdownItemHeight
	relY := y - d.Y - 2 // getItemIndexAtPosition: "2px of padding on the top of the list"
	if relY < 0 {
		return false
	}
	index := relY / itemHeight

	if index < 0 || index >= len(d.Items) {
		return true
	}

	item := &d.Items[index]
	if item.Disabled || item.Separator {
		return true
	}

	// Execute action
	if item.Action != nil {
		item.Action()
	}

	// Call selection callback
	if d.OnSelect != nil {
		d.OnSelect(index)
	}

	// Close dropdown
	d.Visible = false
	return true
}

// UpdateHover updates the highlighted item based on mouse position
func (d *DropdownMenu) UpdateHover(x, y int) {
	if !d.Visible {
		return
	}

	// Check if mouse is inside dropdown
	if x < d.X || x >= d.X+d.Width || y < d.Y || y >= d.Y+d.Height {
		d.Highlighted = -1
		return
	}

	const itemHeight = dropdownItemHeight
	relY := y - d.Y - 2 // getItemIndexAtPosition: "2px of padding on the top of the list"
	if relY < 0 {
		d.Highlighted = -1
		return
	}
	index := relY / itemHeight

	if index >= 0 && index < len(d.Items) {
		if !d.Items[index].Disabled && !d.Items[index].Separator {
			d.Highlighted = index
			return
		}
	}

	d.Highlighted = -1
}

// NewDropdownShow is Dropdown::show (0x004CC807) for a window's dropdown
// widget at x, y of the given height: items `width` wide and 10 high, the
// list below the widget (above it if it doesn't fit, else beside it at the
// top), width + 4 wide, kept on the sw × sh screen; in the window's colour
// (opaque unless the colour is translucent).
func NewDropdownShow(x, y, width, height int, colour uint8, items []DropdownItem, sw, sh int) *DropdownMenu {
	h := len(items)*dropdownItemHeight + 3 + 1
	oy := y + height
	if h+oy > sh || oy < 0 {
		oy -= height + h
		bottom := oy
		if oy >= 0 {
			bottom = oy + h
		}
		if oy < 0 || bottom > sh {
			x += width + 3
			oy = 0
		}
	}
	w := width + 3 + 1
	x = max(x, 0)
	x = min(x+w, sw) - w
	return &DropdownMenu{X: x, Y: oy, Width: w, Height: h, ItemWidth: width, Colour: colour,
		Items: items, Highlighted: -1, Visible: true}
}
