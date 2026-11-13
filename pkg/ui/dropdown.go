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
	// ItemHeight is _dropdownItemHeight (0 = 10; show's flag 1<<6 sets it).
	ItemHeight int
	Visible    bool
	// Origin is the widget that opened it, for the release that follows the
	// opening press (MouseInput.cpp stateWidgetPressed, 0x4C7BC7): released
	// over Origin, the highlighted item is chosen (dropdownRegisterSelection
	// of kDropdownItemUndefined → the menus take getHighlightedItem); released
	// over an item, that item. Set by OpenedByPress; the first release clears it.
	Origin      image.Rectangle
	pressActive bool
	// leaveCloses is Dropdown::Flags::unk1 (showBelow outside a widget
	// press: a menu opened by hovering its button): the mouse over anything
	// but the menu and its button closes it (MouseInput.cpp 0x4C7BC7's
	// dropdownActive check).
	leaveCloses bool
}

// OpenedByPress marks the dropdown as opened by a press on origin (a toolbar
// button): the release that follows chooses (HandleOpeningRelease).
func (d *DropdownMenu) OpenedByPress(origin image.Rectangle) {
	d.Origin, d.pressActive = origin, true
}

// HoverOpened marks the dropdown as opened by hovering its origin (ToolbarTop
// onMouseHover = onMouseDown): no press to release yet.
func (d *DropdownMenu) HoverOpened() { d.pressActive, d.leaveCloses = false, true }

// LeftBehind reports whether a hover-opened menu is to close: the mouse at
// (x, y) is over neither the menu nor the button that opened it.
func (d *DropdownMenu) LeftBehind(x, y int) bool {
	if !d.Visible || !d.leaveCloses {
		return false
	}
	p := image.Pt(x, y)
	return !p.In(d.Origin) && !p.In(image.Rect(d.X, d.Y, d.X+d.Width, d.Y+d.Height))
}

// PressOnOrigin is a press on the widget that opened the dropdown while it is
// open (stateWidgetPressed: the press does nothing, its release on the widget
// chooses): it arms HandleOpeningRelease and reports the press taken.
func (d *DropdownMenu) PressOnOrigin(x, y int) bool {
	if !d.Visible || d.Origin.Empty() || !image.Pt(x, y).In(d.Origin) {
		return false
	}
	d.pressActive = true
	return true
}

// HandleOpeningRelease handles the release of the press that opened the
// dropdown: over the origin, the highlighted item's action (none
// highlighted: nothing, the menu stays open); over an item, that item; else
// the menu stays open for a click. It reports whether the dropdown closed.
func (d *DropdownMenu) HandleOpeningRelease(x, y int) bool {
	if !d.Visible || !d.pressActive {
		return false
	}
	d.pressActive = false
	if image.Pt(x, y).In(d.Origin) {
		if d.Highlighted < 0 || d.Highlighted >= len(d.Items) {
			return false
		}
		it := &d.Items[d.Highlighted]
		if it.Disabled || it.Separator {
			return false
		}
		d.Visible = false
		if it.Action != nil {
			it.Action()
		}
		if d.OnSelect != nil {
			d.OnSelect(d.Highlighted)
		}
		return true
	}
	if x >= d.X && x < d.X+d.Width && y >= d.Y && y < d.Y+d.Height {
		d.HandleClick(x, y)
		return !d.Visible
	}
	return false
}

// DropdownItem represents a single item in a dropdown menu
type DropdownItem struct {
	Text     string
	Disabled bool
	Selected bool // setItemSelected: the format's next string ("»" before the text)
	// MoveX10 is dropdown_stringid's "{MOVE_X 10}{STRINGID}": the text 10 px in
	// (and "»" at the left when Selected).
	Separator bool // Is this a separator line?
	// Sprite, if > 0, is drawn at the item's (0, 1), recoloured to
	// SpriteColour; Lines are small-font lines at (45, 1 + 10·i)
	// (dropdown_bridge_stats' layout).
	Sprite       int
	SpriteColour uint8
	Lines        []string
	MoveX10      bool
	Action       func()
	// TrailSprite, if > 0, is an inline {SPRITE} drawn after the text at the
	// pen position (carrying_cargoid_sprite).
	TrailSprite int
	// CompanySelect is dropdown_company_select's layout
	// "{NEWLINE 1 2}{SPRITE}{NEWLINE 26 8}{STRINGID}": Sprite (the owner's
	// face, recoloured to SpriteColour) at the item's (1, 2), Text at (26, 8).
	CompanySelect bool
	// CompanyPerformance is dropdown_company_performance's layout
	// "{NEWLINE 0 8}{STRINGID}{NEWLINE 25 2}{SPRITE}{NEWLINE 51 2}{STRINGID}
	// {NEWLINE 51 12}{COLOUR WINDOW_1}{INT16_1DP}% “{STRINGID}”": Text (the
	// position) at (0, 8), Sprite (the face) at (25, 2), Lines[0] (the name)
	// at (51, 2), Lines[1] (the performance) at (51, 12) in WINDOW_1.
	CompanyPerformance bool
	// CompaniesListItem is dropdown_companies_list's layout: Sprite at
	// (1, 2), Text at (33, 8).
	CompaniesListItem bool
	// MenuSprite is menu_sprite_stringid's layout "{SPRITE}{NEWLINE 31 8}
	// {STRINGID}": Sprite (recoloured to SpriteColour) at the item's pen
	// origin, then " " + Text at (31, 8).
	MenuSprite bool
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
	ih := d.itemHeight()
	type small struct {
		text string
		x, y int
		c    color.Color
	}
	var smalls []small
	draw := func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 0, d.Width-1, d.Height-1, gfx8.InsetColour{Map: cmap, Translucent: d.Translucent, Trans: trans}, 0) // the Wt3 panel
		for i, it := range d.Items {
			if it.Separator {
				// Two lines at mid-row, item width - 1 wide: shades 3 and 7, or for a
				// translucent colour getTranslucent(c)+1 (highlight) then +2 (shadow).
				sy := 1 + i*ih + ih/2
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
				gfx8.FillRectTranslucent(rt, 2, 2+i*ih, 2+d.ItemWidth-1, 2+i*ih+ih-1, highlight)
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
			if it.Sprite > 0 && it.Sprite < len(g1.Elements) {
				if sp, err := gfx8.Decode(&g1.Elements[it.Sprite]); err == nil {
					m, _ := g1.GetPaletteMap(int(it.SpriteColour))
					sx, sy := 2, 1+i*ih+1
					if it.CompanySelect || it.CompaniesListItem {
						sx, sy = 2+1, 1+i*ih+2
					} else if it.CompanyPerformance {
						sx, sy = 2+25, 1+i*ih+2
					} else if it.MenuSprite {
						sx, sy = 2, 1+i*ih
					}
					gfx8.DrawSpriteNoise(rt, sp, sx, sy, gfx8.OpFor(sp, m != nil, false), m, nil)
				}
			}
			if it.CompanySelect {
				labels = append(labels, label{it.Text, 2 + 26, 1 + i*ih + 8, c, inset})
				continue
			}
			if it.CompaniesListItem {
				labels = append(labels, label{it.Text, 2 + 33, 1 + i*ih + 8, c, inset})
				continue
			}
			if it.CompanyPerformance {
				labels = append(labels, label{it.Text, 2, 1 + i*ih + 8, c, inset})
				if len(it.Lines) > 0 {
					labels = append(labels, label{it.Lines[0], 2 + 51, 1 + i*ih + 2, c, inset})
				}
				if len(it.Lines) > 1 {
					w1 := c // {COLOUR WINDOW_1}: the dropdown's colour, shade 9 (white when highlighted)
					labels = append(labels, label{it.Lines[1], 2 + 51, 1 + i*ih + 12, w1, inset})
				}
				continue
			}
			if it.MenuSprite {
				labels = append(labels, label{" " + it.Text, 2 + 31, 1 + i*ih + 8, c, inset})
				continue
			}
			if it.TrailSprite > 0 && it.Text != "" {
				tx := 2
				if it.MoveX10 {
					tx += 10
				}
				tw, _ := MeasureTextBold(it.Text)
				drawG1At(rt, it.TrailSprite, tx+tw, 1+i*ih)
			}
			for k, ln := range it.Lines {
				smalls = append(smalls, small{ln, 2 + 45, 1 + i*ih + 1 + 10*k, c})
			}
			if it.Text != "" {
				x := 2
				if it.MoveX10 {
					if it.Selected {
						labels = append(labels, label{"»", 2, 1 + i*ih, c, inset})
					}
					x += 10
				}
				labels = append(labels, label{it.Text, x, 1 + i*ih, c, inset})
			}
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
		tmp := ebiten.NewImageFromImage(img)
		screen.DrawImage(tmp, op)
		tmp.Deallocate() // a texture per frame otherwise (see draw8)
	}
	for _, l := range labels {
		if l.inset != nil {
			DrawTextBold(screen, l.text, d.X+l.x, d.Y+l.y, l.inset)
			DrawTextBold(screen, l.text, d.X+l.x+1, d.Y+l.y+1, l.color)
			continue
		}
		DrawTextBold(screen, l.text, d.X+l.x, d.Y+l.y, l.color)
	}
	for _, l := range smalls {
		DrawText(screen, l.text, d.X+l.x, d.Y+l.y, l.c)
	}
}

func (d *DropdownMenu) itemHeight() int {
	if d.ItemHeight > 0 {
		return d.ItemHeight
	}
	return dropdownItemHeight
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

	itemHeight := d.itemHeight()
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

	// Only over an item does the highlight move (MouseInput.cpp 0x4C7F02:
	// setHighlightedItem when dropdownIndexFromPoint has one); elsewhere it
	// stays — the menu's own highlighted item, chosen by a release on its
	// button, survives the mouse leaving the list.
	if x < d.X || x >= d.X+d.Width || y < d.Y || y >= d.Y+d.Height {
		return
	}

	itemHeight := d.itemHeight()
	relY := y - d.Y - 2 // getItemIndexAtPosition: "2px of padding on the top of the list"
	if relY < 0 {
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
	return NewDropdownShowItemHeight(x, y, width, height, colour, items, dropdownItemHeight, sw, sh)
}

// NewDropdownShowItemHeight is Dropdown::show with flag 1<<6: items of the
// given height.
func NewDropdownShowItemHeight(x, y, width, height int, colour uint8, items []DropdownItem, itemHeight, sw, sh int) *DropdownMenu {
	h := len(items)*itemHeight + 3 + 1
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
		Items: items, Highlighted: -1, Visible: true, ItemHeight: itemHeight}
}

// itemTextWidth is getMaxStringWidth of an item's formatted string: the text
// in medium bold (after MOVE_X 10), or for company select max(1 + the face's
// width, 26 + the name's).
func itemTextWidth(it DropdownItem) int {
	tw, _ := MeasureTextBold(it.Text)
	if it.CompanyPerformance {
		w := tw
		if g1 := globalG1Font.g1; it.Sprite > 0 && it.Sprite < len(g1.Elements) {
			w = max(w, 25+int(g1.Elements[it.Sprite].Width))
		}
		for _, l := range it.Lines {
			lw, _ := MeasureTextBold(l)
			w = max(w, 51+lw)
		}
		return w
	}
	if it.CompaniesListItem {
		sw := 0
		if g1 := globalG1Font.g1; it.Sprite > 0 && it.Sprite < len(g1.Elements) {
			sw = int(g1.Elements[it.Sprite].Width)
		}
		return max(1+sw, 33+tw)
	}
	if it.CompanySelect {
		sw := 0
		if g1 := globalG1Font.g1; it.Sprite > 0 && it.Sprite < len(g1.Elements) {
			sw = int(g1.Elements[it.Sprite].Width)
		}
		return max(1+sw, 26+tw)
	}
	if it.MenuSprite {
		sw := 0
		if g1 := globalG1Font.g1; it.Sprite > 0 && it.Sprite < len(g1.Elements) {
			sw = int(g1.Elements[it.Sprite].Width)
		}
		sp, _ := MeasureTextBold(" " + it.Text)
		return max(sw, 31+sp)
	}
	if it.MoveX10 {
		tw += 10
	}
	if g1 := globalG1Font.g1; it.TrailSprite > 0 && it.TrailSprite < len(g1.Elements) {
		tw += int(g1.Elements[it.TrailSprite].Width) // getStringWidth: an inline {SPRITE} adds its width
	}
	return tw
}

// NewDropdownShowText is Dropdown::showText (common::showText): items
// itemHeight high (flag 1<<6), as wide as the widest formatted item + 3; the
// list below the widget, above it if it doesn't fit, else beside it at the
// top (to the left when that runs off the screen); width + 3 + 1 wide.
func NewDropdownShowText(x, y, width, height int, colour uint8, items []DropdownItem, itemHeight, sw, sh int) *DropdownMenu {
	maxW := 0
	for _, it := range items {
		maxW = max(maxW, itemTextWidth(it))
	}
	maxW += 3
	h := len(items)*itemHeight + 3 + 1
	ox, oy := x, y+height
	if h+oy > sh || oy < 0 {
		oy -= height + h
		bottom := oy
		if oy >= 0 {
			bottom = oy + h
		}
		if oy < 0 || bottom > sh {
			ox += width + maxW
			if ox > sw {
				ox = x - (maxW + 4)
			}
			oy = 0
		}
	}
	w := maxW + 3 + 1
	ox = max(ox, 0)
	ox = min(ox+w, sw) - w
	return &DropdownMenu{X: ox, Y: oy, Width: w, Height: h, ItemWidth: maxW, Colour: colour,
		Items: items, Highlighted: -1, Visible: true, ItemHeight: itemHeight}
}
