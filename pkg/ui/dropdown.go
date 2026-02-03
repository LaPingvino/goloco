package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// DropdownMenu represents a dropdown menu window
//
// OpenLoco reference: src/OpenLoco/src/Ui/Dropdown.cpp
//   Dropdown menus are temporary windows that appear when clicking toolbar buttons.
type DropdownMenu struct {
	X          int
	Y          int
	Width      int
	Height     int
	Items      []DropdownItem
	Highlighted int
	OnSelect   func(index int)
	Visible    bool
}

// DropdownItem represents a single item in a dropdown menu
type DropdownItem struct {
	Text      string
	Disabled  bool
	Selected  bool  // Shows checkmark
	Separator bool  // Is this a separator line?
	Action    func()
}

// NewDropdownMenu creates a new dropdown menu positioned below a widget
//
// OpenLoco reference: src/OpenLoco/src/Ui/Dropdown.cpp
//   showBelow(window, widgetIndex, count, itemHeight, flags)
func NewDropdownMenu(x, y int, items []DropdownItem) *DropdownMenu {
	const itemHeight = 16
	const padding = 4

	// Calculate width based on longest text
	maxWidth := 120
	for _, item := range items {
		textWidth := len(item.Text) * 6 + padding*2
		if textWidth > maxWidth {
			maxWidth = textWidth
		}
	}

	height := len(items) * itemHeight

	return &DropdownMenu{
		X:           x,
		Y:           y,
		Width:       maxWidth,
		Height:      height,
		Items:       items,
		Highlighted: -1,
		Visible:     true,
	}
}

// Draw renders the dropdown menu
func (d *DropdownMenu) Draw(screen *ebiten.Image) {
	if !d.Visible {
		return
	}

	const itemHeight = 16

	// Draw background
	bgColor := color.RGBA{60, 60, 60, 240}
	ebitenutil.DrawRect(screen, float64(d.X), float64(d.Y), float64(d.Width), float64(d.Height), bgColor)

	// Draw border
	borderColor := color.RGBA{120, 120, 120, 255}
	ebitenutil.DrawRect(screen, float64(d.X), float64(d.Y), float64(d.Width), 1, borderColor)
	ebitenutil.DrawRect(screen, float64(d.X), float64(d.Y), 1, float64(d.Height), borderColor)
	ebitenutil.DrawRect(screen, float64(d.X), float64(d.Y+d.Height-1), float64(d.Width), 1, borderColor)
	ebitenutil.DrawRect(screen, float64(d.X+d.Width-1), float64(d.Y), 1, float64(d.Height), borderColor)

	// Draw items
	for i, item := range d.Items {
		itemY := d.Y + i*itemHeight

		if item.Separator {
			// Draw separator line
			sepColor := color.RGBA{80, 80, 80, 255}
			ebitenutil.DrawRect(screen, float64(d.X+4), float64(itemY+itemHeight/2), float64(d.Width-8), 1, sepColor)
			continue
		}

		// Highlight if hovered
		if i == d.Highlighted {
			highlightColor := color.RGBA{80, 100, 120, 200}
			ebitenutil.DrawRect(screen, float64(d.X+1), float64(itemY), float64(d.Width-2), float64(itemHeight), highlightColor)
		}

		// Draw checkmark if selected
		if item.Selected {
			DrawText(screen, "✓", d.X+4, itemY+12, color.White)
		}

		// Draw text
		textX := d.X + 20
		textY := itemY + 12

		// Render disabled items in grey color
		var textColor color.Color = color.White
		if item.Disabled {
			textColor = color.RGBA{128, 128, 128, 255}
		}
		DrawText(screen, item.Text, textX, textY, textColor)
	}
}

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

	const itemHeight = 16
	relY := y - d.Y
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

	const itemHeight = 16
	relY := y - d.Y
	index := relY / itemHeight

	if index >= 0 && index < len(d.Items) {
		if !d.Items[index].Disabled && !d.Items[index].Separator {
			d.Highlighted = index
			return
		}
	}

	d.Highlighted = -1
}
