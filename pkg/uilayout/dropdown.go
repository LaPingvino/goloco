// Package uilayout holds the UI's layout arithmetic, ported from upstream Ui/,
// free of Ebitengine so it can be tested without a display.
package uilayout

// DropdownPlace is the placement half of Dropdown.cpp common::showText for a
// dropdown of width × height (itemWidth = _dropdownItemWidth, i.e. the
// maxStringWidth that already includes its +3) below the widget at wx, wy (wy
// its top) of size ww × wh, on a screen sw × sh: below the widget; if that
// leaves the screen, above it; if that fails too, beside it at the top; then x
// kept on screen.
func DropdownPlace(wx, wy, ww, wh, itemWidth, width, height, sw, sh int) (x, y int) {
	x, y = wx, wy+wh
	if height+y > sh || y < 0 {
		y -= wh + height
		bottom := y
		if y >= 0 {
			bottom = y + height
		}
		if y < 0 || bottom > sh {
			x += ww + itemWidth
			if x > sw {
				x = wx - (itemWidth + 4)
			}
			y = 0
		}
	}
	x = max(x, 0)
	x = min(x+width, sw) - width
	return x, y
}
