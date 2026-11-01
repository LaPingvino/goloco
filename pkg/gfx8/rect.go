package gfx8

// Rectangles — a port of Graphics/SoftwareDrawingContext.cpp drawRectImpl
// (regular and transparent fill), fillRectInset / drawRectInset (the 3D window
// bevels) and Colours::getShade. Coordinates are relative to the target (rt.x
// = rt.y = 0), zoom 0, as the UI is drawn.

// RectInsetFlags (Graphics/DrawingContext.h).
const (
	InsetFillTransparent = 1 << 2
	InsetBorderNone      = 1 << 3
	InsetFillNone        = 1 << 4
	InsetBorderInset     = 1 << 5
	InsetFillDarker      = 1 << 6
	InsetColourLight     = 1 << 7
)

// Shade is Colours::getShade: the colour's palette map at primaryRemap2+shade
// (shades 0-7 from _colourMapA, 8-9 from _colourMapB — the same indices).
func Shade(colourMap []uint8, shade int) uint8 {
	return colourMap[PrimaryRemap[2+shade]]
}

// FillRect is drawRectImpl with RectFlags::none, on the inclusive rectangle
// left..right × top..bottom, clipped to the target.
func FillRect(rt *RenderTarget, left, top, right, bottom int, colour uint8) {
	l, t, r, b, ok := clipRect(rt, left, top, right, bottom)
	if !ok {
		return
	}
	for y := t; y < b; y++ {
		row := rt.Bits[y*rt.Width+l : y*rt.Width+r]
		for i := range row {
			row[i] = colour
		}
	}
}

// FillRectTranslucent is drawRectImpl with RectFlags::transparent: every
// covered pixel is remapped through the translucent colour's palette map.
func FillRectTranslucent(rt *RenderTarget, left, top, right, bottom int, palMap []uint8) {
	l, t, r, b, ok := clipRect(rt, left, top, right, bottom)
	if !ok || palMap == nil {
		return
	}
	for y := t; y < b; y++ {
		row := rt.Bits[y*rt.Width+l : y*rt.Width+r]
		for i, p := range row {
			row[i] = palMap[p]
		}
	}
}

func clipRect(rt *RenderTarget, left, top, right, bottom int) (l, t, r, b int, ok bool) {
	if left > right || top > bottom || right < 0 || left >= rt.Width || bottom < 0 || top >= rt.Height {
		return
	}
	return max(left, 0), max(top, 0), min(right+1, rt.Width), min(bottom+1, rt.Height), true
}

// InsetColour is what fillRectInset needs of an AdvancedColour: the base
// colour's palette map (for getShade) or, when translucent, the palette maps of
// getTranslucent(colour, 0/1/2).
type InsetColour struct {
	Map         []uint8    // the colour's palette map (opaque colours)
	Translucent bool       // AdvancedColour::isTranslucent
	Trans       [3][]uint8 // palette maps of getTranslucent(colour, shade 0..2)
}

// FillRectInset is fillRectInset (inclusive left..right × top..bottom).
func FillRectInset(rt *RenderTarget, left, top, right, bottom int, c InsetColour, flags int) {
	if c.Translucent {
		t := func(l, tp, r, b, shade int) { FillRectTranslucent(rt, l, tp, r, b, c.Trans[shade]) }
		switch {
		case flags&InsetBorderNone != 0:
			t(left, top, right, bottom, 1)
		case flags&InsetBorderInset != 0:
			t(left, top, left, bottom, 2)
			t(left, top, right, top, 2)
			t(right, top, right, bottom, 0)
			t(left, bottom, right, bottom, 0)
			if flags&InsetFillNone == 0 {
				t(left+1, top+1, right-1, bottom-1, 1)
			}
		default:
			t(left, top, left, bottom, 0)
			t(left, top, right, top, 0)
			t(right, top, right, bottom, 2)
			t(left, bottom, right, bottom, 2)
			if flags&InsetFillNone == 0 {
				t(left+1, top+1, right-1, bottom-1, 1)
			}
		}
		return
	}
	shadow, fill, fill2, hilight := Shade(c.Map, 3), Shade(c.Map, 5), Shade(c.Map, 6), Shade(c.Map, 7)
	if flags&InsetColourLight != 0 {
		shadow, fill, fill2, hilight = Shade(c.Map, 1), Shade(c.Map, 3), Shade(c.Map, 4), Shade(c.Map, 5)
	}
	switch {
	case flags&InsetBorderNone != 0:
		FillRect(rt, left, top, right, bottom, fill)
	case flags&InsetBorderInset != 0:
		FillRect(rt, left, top, left, bottom, shadow)
		FillRect(rt, left+1, top, right, top, shadow)
		FillRect(rt, right, top+1, right, bottom-1, hilight)
		FillRect(rt, left+1, bottom, right, bottom, hilight)
		if flags&InsetFillNone == 0 {
			if flags&InsetFillDarker == 0 {
				fill = fill2
			}
			if flags&InsetFillTransparent != 0 {
				fill = Transparent
			}
			FillRect(rt, left+1, top+1, right-1, bottom-1, fill)
		}
	default:
		FillRect(rt, left, top, left, bottom-1, hilight)
		FillRect(rt, left+1, top, right-1, top, hilight)
		FillRect(rt, right, top, right, bottom-1, shadow)
		FillRect(rt, left, bottom, right, bottom, shadow)
		if flags&InsetFillNone == 0 {
			if flags&InsetFillTransparent != 0 {
				fill = Transparent
			}
			FillRect(rt, left+1, top+1, right-1, bottom-1, fill)
		}
	}
}

// DrawRectInset is drawRectInset: x, y, width, height.
func DrawRectInset(rt *RenderTarget, x, y, w, h int, c InsetColour, flags int) {
	FillRectInset(rt, x, y, x+w-1, y+h-1, c, flags)
}
