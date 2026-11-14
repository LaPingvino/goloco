package ui

import (
	"image"
	"strings"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
	L "github.com/LaPingvino/goloco/pkg/uilayout"
	"github.com/hajimehoshi/ebiten/v2"
)

// KeyboardShortcuts — a port of Ui/Windows/KeyboardShortcuts.cpp (Options →
// Controls → Customise Keys): 420×238 centred (resizable up to 600×800), the
// skin's windowTitlebarColour and windowOptionsColour; a scroll view of every
// shortcut in 10 px rows — "Name:" and, three fifths across, its binding —
// with separator lines after the groups (kSeparatorPositions); the hovered row
// highlighted (unk30, wcolour2 text); a click on a row opens
// EditKeyboardShortcut; "Reset Keys" puts every default back.
//
// clickable(KeyboardShortcuts.cpp): close_button reset_keys_btn
//
// PARTIAL: CJK row height (13).

const (
	ksW, ksH       = 420, 238
	ksMaxW, ksMaxH = 600, 800
	ksRowHeight    = 10
)

// ShortcutRow is one row of the list: a separator, or a shortcut's name and
// binding text ("" unbound).
type ShortcutRow struct {
	Separator     bool
	ID            int
	Name, Binding string
}

// KeyboardShortcutsWindow is the window.
type KeyboardShortcutsWindow struct {
	WinBase
	TitleColour, Colour uint8
	// Rows gives the list (read each draw: a changed binding shows at once);
	// OnRow is onScrollMouseDown on a shortcut; OnReset the Reset Keys button.
	Rows    func() []ShortcutRow
	OnRow   func(id int)
	OnReset func()

	hover   int
	scroll  int
	hold    L.VScrollHold
	content [4]int
}

// NewKeyboardShortcuts is KeyboardShortcuts::open's window (createWindowCentred).
func NewKeyboardShortcuts(sw, sh int, titleColour, colour uint8) *KeyboardShortcutsWindow {
	return &KeyboardShortcutsWindow{WinBase: WinBase{WinType: WindowTypeKeyboardShortcuts, X: sw/2 - ksW/2, Y: max(28, sh/2-ksH/2), W: ksW, H: ksH,
		MinW: ksW, MinH: ksH, MaxW: ksMaxW, MaxH: ksMaxH, WinFlags: WindowFlagResizable}, TitleColour: titleColour, Colour: colour, hover: -1}
}

// list is the scroll view as onResize lays it out (x, y, w, h); reset the button.
func (w *KeyboardShortcutsWindow) list() [4]int { return [4]int{4, 19, w.W - 8, w.H - 20 - 19 + 1} }
func (w *KeyboardShortcutsWindow) reset() [4]int {
	return [4]int{w.W - 150 - 4 - 12, w.H - 4 - 12, 150, 13}
}

func (w *KeyboardShortcutsWindow) rows() []ShortcutRow {
	if w.Rows == nil {
		return nil
	}
	return w.Rows()
}

func (w *KeyboardShortcutsWindow) contentHeight() int { return ksRowHeight * len(w.rows()) }

// PrepareDraw keeps the scroll within the resized list.
func (w *KeyboardShortcutsWindow) PrepareDraw() {
	w.scroll = min(w.scroll, max(w.contentHeight()-(w.list()[3]-2), 0))
}

func (w *KeyboardShortcutsWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

func (w *KeyboardShortcutsWindow) onBar(px, py int) bool {
	l := w.list()
	return w.in(l, px, py) && w.contentHeight() > l[3] && px >= w.X+l[0]+l[2]-L.ScrollbarSize-L.ScrollbarMargin
}

// rowAt is the row under a screen point of the list (-1: none).
func (w *KeyboardShortcutsWindow) rowAt(px, py int) int {
	l := w.list()
	if !w.in(l, px, py) || w.onBar(px, py) || py < w.Y+l[1]+1 {
		return -1
	}
	if i := (py - (w.Y + l[1] + 1) + w.scroll) / ksRowHeight; i < len(w.rows()) {
		return i
	}
	return -1
}

// GripBlocked: the list reaches the resize corner's row.
func (w *KeyboardShortcutsWindow) GripBlocked(px, py int) bool { return w.in(w.list(), px, py) }

// HandleClick is onMouseUp (close, reset) and onScrollMouseDown (a shortcut's
// row opens its editor; a separator does nothing).
func (w *KeyboardShortcutsWindow) HandleClick(px, py int) {
	switch {
	case w.in([4]int{w.W - 15, 2, 13, 13}, px, py):
		w.Close()
	case w.in(w.reset(), px, py):
		if w.OnReset != nil {
			w.OnReset()
		}
	case w.onBar(px, py):
		l := w.list()
		w.scroll = w.hold.Begin(py-(w.Y+l[1]), py, l[3], w.contentHeight(), w.scroll)
	default:
		if i := w.rowAt(px, py); i >= 0 {
			if r := w.rows()[i]; !r.Separator && w.OnRow != nil {
				w.OnRow(r.ID)
			}
		}
	}
}

// MouseOver is onScrollMouseOver: the row under the mouse.
func (w *KeyboardShortcutsWindow) MouseOver(px, py int) {
	if i := w.rowAt(px, py); i >= 0 {
		w.hover = i
	}
}

// Drag is the scrollbar's held press each tick.
func (w *KeyboardShortcutsWindow) Drag(px, py int, held bool) {
	l := w.list()
	w.scroll = w.hold.Continue(held, w.onBar(px, py), py-(w.Y+l[1]), py, l[3], w.contentHeight(), w.scroll)
}

// HandleWheel scrolls the list.
func (w *KeyboardShortcutsWindow) HandleWheel(dy float64) {
	w.scroll = min(max(w.scroll-int(dy*3*ksRowHeight), 0), max(w.contentHeight()-(w.list()[3]-2), 0))
}

// WidgetTooltip is the tooltip of the widget under a screen point (the
// list's keyboard_shortcut_list_tip; KeyboardShortcuts::tooltip's
// tooltip_scroll_list argument has no place in it).
func (w *KeyboardShortcutsWindow) WidgetTooltip(px, py int) string {
	switch {
	case w.in([4]int{w.W - 15, 2, 13, 13}, px, py):
		return i18n.Str(i18n.TooltipCloseWindow)
	case w.in(w.reset(), px, py):
		return i18n.Str(i18n.ResetKeysTip)
	case w.in(w.list(), px, py):
		return i18n.Str(i18n.KeyboardShortcutListTip)
	}
	return ""
}

// Draw draws the window.
func (w *KeyboardShortcutsWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	rows, l, rb := w.rows(), w.list(), w.reset()
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, false)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 15, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0) // Panel
		drawResizeHandle(rt, &w.WinBase, sec)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, w.W-15, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		gfx8.FillRectInset(rt, rb[0], rb[1], rb[0]+rb[2]-1, rb[1]+rb[3]-1, gfx8.InsetColour{Map: sec}, 0) // Button
		cx, cy, cw, ch, _ := drawVScrollViewPressed(rt, l[0], l[1], l[2], l[3], gfx8.InsetColour{Map: sec}, w.contentHeight(), w.scroll, w.hold.PressedPart())
		w.content = [4]int{cx, cy, cw, ch}
		if cw <= 0 || ch <= 0 {
			return
		}
		sub := gfx8.NewRenderTarget(cw, ch)
		gfx8.FillRect(sub, 0, 0, cw-1, ch-1, gfx8.Shade(sec, 4)) // clear(shade 4)
		m30, _ := g1.GetPaletteMap(0x30)                         // ExtColour::unk30
		sepColour := int(paintTranslucent(w.Colour, 1))          // Colours::getTranslucent(colour): shade 1
		trans, _ := g1.GetPaletteMap(sepColour + 1)
		trans2, _ := g1.GetPaletteMap(sepColour + 2)
		width := l[2] // widgets[list].width()
		for i, r := range rows {
			y := i*ksRowHeight - w.scroll
			if y+ksRowHeight <= 0 || y >= ch {
				continue
			}
			if r.Separator { // getTranslucent(colour) + 1 and + 2, transparent, at y + 4 and + 5
				gfx8.FillRectTranslucent(sub, 1, y+4, width-1, y+4, trans)
				gfx8.FillRectTranslucent(sub, 1, y+5, width-1, y+5, trans2)
				continue
			}
			if i == w.hover {
				gfx8.FillRectTranslucent(sub, 0, y, 799, y+ksRowHeight-1, m30)
			}
		}
		for yy := range ch {
			copy(rt.Bits[(cy+yy)*rt.Width+cx:(cy+yy)*rt.Width+cx+cw], sub.Bits[yy*cw:(yy+1)*cw])
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.KeyboardShortcuts))
	label := i18n.Label(i18n.ResetKeys)
	lw, _ := MeasureTextBold(label)
	DrawTextBold(screen, label, w.X+rb[0]+(rb[2]-1)/2-lw/2, w.Y+rb[1], black)
	if w.contentHeight() > l[3] {
		drawVScrollArrows(screen, w.X+l[0], w.Y+l[1], l[2], l[3])
	}
	ct := w.content
	clip := screen
	if c, ok := screen.SubImage(image.Rect(w.X+ct[0], w.Y+ct[1], w.X+ct[0]+ct[2], w.Y+ct[1]+ct[3])).(*ebiten.Image); ok {
		clip = c
	}
	ox, oy := w.X+ct[0], w.Y+ct[1]
	listFormat := i18n.Str(i18n.KeyboardShortcutListFormat) // "{STRINGID}:"
	for i, r := range rows {
		y := oy + i*ksRowHeight - w.scroll
		if r.Separator || y+ksRowHeight <= oy || y >= oy+ct[3] {
			continue
		}
		c := black // black_stringid; the hovered row wcolour2_stringid
		if i == w.hover {
			c = win2
		}
		DrawTextBold(clip, strings.Replace(listFormat, "{STRINGID}", r.Name, 1), ox, y, c)
		if r.Binding != "" {
			DrawTextBold(clip, r.Binding, ox+l[2]/5*3, y, c)
		}
	}
}

// EditKeyboardShortcut — a port of Ui/Windows/EditKeyboardShortcut.cpp:
// 280×78 centred, "Press new shortcut key for:", the shortcut's name in
// quotes and the modifiers held now; the next key (not an arrow, Num Lock or a
// system key) becomes the shortcut's binding, and the window closes.
// clickable(EditKeyboardShortcut.cpp): close
type EditShortcutWindow struct {
	WinBase
	TitleColour, Colour uint8
	ID                  int    // the shortcut being edited
	Name                string // its display name
	Modifiers           string // the modifiers held (pushModifierStrings), set each update
}

// NewEditShortcut is EditKeyboardShortcut::open's window.
func NewEditShortcut(sw, sh int, id int, name string, titleColour, colour uint8) *EditShortcutWindow {
	const w, h = 280, 78
	return &EditShortcutWindow{WinBase: WinBase{WinType: WindowTypeEditKeyboardShortcut, X: sw/2 - w/2, Y: max(28, sh/2-h/2), W: w, H: h},
		TitleColour: titleColour, Colour: colour, ID: id, Name: name}
}

// HandleClick is onMouseUp: the close button.
func (w *EditShortcutWindow) HandleClick(px, py int) {
	if px >= w.X+265 && px < w.X+278 && py >= w.Y+2 && py < w.Y+15 {
		w.Close()
	}
}

// Draw draws the window: the frame, the caption, and the three centred labels
// (Label::draw: x + (w − 1) / 2, the text at the widget's top).
func (w *EditShortcutWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	pal := g1.GetPalette()
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, false)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 15, w.W-1, w.H-1, gfx8.InsetColour{Map: sec}, 0)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, 265, 2, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
	})
	DrawWindowCaption(screen, w.X+1, w.Y+1, w.W-2, i18n.Label(i18n.ChangeKeyboardShortcut))
	centred := func(s string, y int, c any) {
		tw, _ := MeasureTextBold(s)
		col := pal[paletteBlack0]
		if c != nil {
			col = pal[gfx8.Shade(sec, 9)]
		}
		DrawTextBold(screen, s, w.X+4+(w.W-8-1)/2-tw/2, w.Y+y, col)
	}
	centred(i18n.Label(i18n.ChangeKeyboardShortcutDesc), 20, true)                                   // {COLOUR WINDOW_2}
	centred(strings.Replace(i18n.Label(i18n.BlackQuotedStringid), "{STRINGID}", w.Name, 1), 34, nil) // black_quoted_stringid
	centred(w.Modifiers, 54, nil)
}
