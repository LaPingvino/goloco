package ui

import (
	"image"
	"image/color"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/hajimehoshi/ebiten/v2"
)

// DropdownMenu represents a dropdown menu window
//
// OpenLoco reference: src/OpenLoco/src/Ui/Dropdown.cpp
//
//	Dropdown menus are temporary windows that appear when clicking toolbar buttons.
type DropdownMenu struct {
	X           int
	Y           int
	Width       int
	Height      int
	Items       []DropdownItem
	ItemWidth   int // _dropdownItemWidth
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
	// DropdownColour is the window colour the dropdown panel is drawn in.
	// PARTIAL: upstream's toolbar dropdowns are translucent (kDropdownColourTable
	// + translucent(), remapping the view behind them); that needs the UI drawn
	// into the 8-bit frame — until then the opaque colour is used.
	DropdownColour = 1 // Colour::grey
)

// NewDropdownMenu creates a dropdown below a widget: x, y is the widget's
// bottom-left (showBelow passes the widget top + height).
//
// OpenLoco reference: src/OpenLoco/src/Ui/Dropdown.cpp showText / showBelow.
// PARTIAL: the flip-up / beside-the-widget placement when the dropdown would
// leave the screen is not applied (toolbar dropdowns always fit below).
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
		Items:       items,
		Highlighted: -1,
		Visible:     true,
	}
}

// Draw renders the dropdown (Dropdown.cpp draw) into an 8-bit target with the
// gfx8 window primitives and puts it on screen through the game palette.
func (d *DropdownMenu) Draw(screen *ebiten.Image) {
	if !d.Visible || globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	cmap, err := g1.GetPaletteMap(DropdownColour)
	if err != nil {
		return
	}
	rt := gfx8.NewRenderTarget(d.Width, d.Height)
	gfx8.FillRectInset(rt, 0, 0, d.Width-1, d.Height-1, gfx8.InsetColour{Map: cmap}, 0) // the Wt3 panel
	highlight, _ := g1.GetPaletteMap(extColourUnk2E)
	pal := g1.GetPalette()
	// Text is drawn after the frame is on screen (pkg/ui draws text with ebiten).
	type label struct {
		text  string
		x, y  int
		color color.Color
	}
	var labels []label
	for i, it := range d.Items {
		x, y := 2, 1+i*dropdownItemHeight
		if it.Separator {
			// Two lines: shade 3 then shade 7, item width - 1 wide, at mid-row.
			sy := 1 + i*dropdownItemHeight + dropdownItemHeight/2
			gfx8.FillRect(rt, 2, sy, 2+d.ItemWidth-2, sy, gfx8.Shade(cmap, 3))
			gfx8.FillRect(rt, 2, sy+1, 2+d.ItemWidth-2, sy+1, gfx8.Shade(cmap, 7))
			continue
		}
		// PARTIAL: upstream draws the string in the window's primary colour through
		// the text renderer's colour handling (not ported); shade 9 stands in.
		var c color.Color = pal[gfx8.Shade(cmap, 9)]
		if i == d.Highlighted {
			gfx8.FillRectTranslucent(rt, 2, 2+i*dropdownItemHeight, 2+d.ItemWidth-1, 2+i*dropdownItemHeight+dropdownItemHeight-1, highlight)
			c = color.White
		}
		if it.Disabled {
			c = pal[gfx8.Shade(cmap, 7)] // PARTIAL: upstream draws the inset (embossed) text style
		}
		labels = append(labels, label{it.Text, x, y, c})
	}
	img := image.NewRGBA(image.Rect(0, 0, d.Width, d.Height))
	for i, p := range rt.Bits {
		img.Set(i%d.Width, i/d.Width, pal[p])
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(d.X), float64(d.Y))
	screen.DrawImage(ebiten.NewImageFromImage(img), op)
	for _, l := range labels {
		DrawTextBold(screen, l.text, d.X+l.x, d.Y+l.y, l.color)
	}
}

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
