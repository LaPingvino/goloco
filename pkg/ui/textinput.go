package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/i18n"
)

// Text entry: Ui/TextInput.cpp's InputSession and
// Ui/Windows/TextInputWindow.cpp — the window towns, stations, companies and
// vehicles are renamed through.

// textboxPadding is TextInput.h's textboxPadding.
const textboxPadding = 4

// InputSession is TextInput::InputSession: the text (as runes; each stands
// for one of the game's 8-bit characters), the caret, the caret's blink
// counter, the scroll offset and the length limit.
type InputSession struct {
	Buffer         []rune
	CursorPosition int
	CursorFrame    int
	XOffset        int
	InputLenLimit  int
}

// Input keys (the keyCodes handleInput reads besides typed characters).
const (
	KeyBackspace = iota + 1
	KeyDelete
	KeyHome
	KeyEnd
	KeyLeft
	KeyRight
)

// NewInputSession is InputSession(text, limit).
func NewInputSession(text string, limit int) *InputSession {
	b := []rune(text)
	return &InputSession{Buffer: b, CursorPosition: len(b), InputLenLimit: limit}
}

// Type is handleInput for a typed character: printable (32–126, 159–255 in
// the game's code page) inserted at the caret unless the limit is reached.
func (s *InputSession) Type(r rune) {
	c := i18n.LocoChar(r)
	if !(c >= 32 && c < 127 || c >= 159 && c <= 255) {
		return
	}
	if s.InputLenLimit > 0 && len(s.Buffer) >= s.InputLenLimit {
		return
	}
	s.Buffer = append(s.Buffer[:s.CursorPosition], append([]rune{r}, s.Buffer[s.CursorPosition:]...)...)
	s.CursorPosition++
	s.CursorFrame = 0
}

// Key is handleInput for an editing key.
func (s *InputSession) Key(k int) {
	switch k {
	case KeyBackspace:
		if s.CursorPosition == 0 {
			return
		}
		s.Buffer = append(s.Buffer[:s.CursorPosition-1], s.Buffer[s.CursorPosition:]...)
		s.CursorPosition--
	case KeyDelete:
		if s.CursorPosition == len(s.Buffer) {
			return
		}
		s.Buffer = append(s.Buffer[:s.CursorPosition], s.Buffer[s.CursorPosition+1:]...)
	case KeyHome:
		s.CursorPosition = 0
	case KeyEnd:
		s.CursorPosition = len(s.Buffer)
	case KeyLeft:
		if s.CursorPosition == 0 {
			return
		}
		s.CursorPosition--
	case KeyRight:
		if s.CursorPosition == len(s.Buffer) {
			return
		}
		s.CursorPosition++
	}
	s.CursorFrame = 0
}

// NeedsReoffsetting is needsReoffsetting: the caret outside the padding, or
// the text's end short of the right edge.
func (s *InputSession) NeedsReoffsetting(containerWidth int) bool {
	sw, _ := MeasureTextBold(string(s.Buffer))
	cx, _ := MeasureTextBold(string(s.Buffer[:s.CursorPosition]))
	x := s.XOffset + cx
	return x < textboxPadding || x > containerWidth-textboxPadding || s.XOffset+sw < containerWidth-textboxPadding
}

// CalculateTextOffset is calculateTextOffset (0x004CEB67): the caret centred,
// the text kept at the left edge, the box filled when it is longer.
func (s *InputSession) CalculateTextOffset(containerWidth int) {
	sw, _ := MeasureTextBold(string(s.Buffer))
	cx, _ := MeasureTextBold(string(s.Buffer[:s.CursorPosition]))
	s.XOffset = -(cx - containerWidth/2)
	minOffset, maxOffset := textboxPadding, textboxPadding
	if sw+textboxPadding*2 > containerWidth {
		minOffset = -sw + containerWidth - textboxPadding
	}
	s.XOffset = max(minOffset, min(s.XOffset, maxOffset))
}

// Sanitized is sanitizeInput (0x004CEBFB): control characters and 123–190
// dropped, but « (171) and » (187) kept.
func (s *InputSession) Sanitized() string {
	var out []rune
	for _, r := range s.Buffer {
		c := i18n.LocoChar(r)
		if c < ' ' || c > 'z' && c != 171 && c != 187 && c < 191 {
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

// TextInputWindow is TextInputWindow.cpp: 330×90, centred, in the caller's
// colours (a company's: the lighter frame and a coloured caption); the
// message centred and wrapped at y 30, the text box (4, 58, 322×14) with the
// caret blinking (cursorFrame % 32 < 16), "n/limit" right-aligned at
// (150, 75, 100) and OK. Enter is OK, Escape closes.
type TextInputWindow struct {
	X, Y                int
	Title, Message      string
	TitleColour, Colour uint8
	Owned               bool // window->owner != null: lighterFrame, colourText caption
	Session             *InputSession
	OnOK                func(text string)
	Closed              bool
}

const tiW, tiH = 330, 90

var (
	tiClose = [4]int{315, 2, 13, 13}
	tiInput = [4]int{4, 58, 322, 14}
	tiLimit = [4]int{150, 75, 100, 10}
	tiOK    = [4]int{256, 74, 70, 12}
)

// OpenTextInput is openTextInput: the window centred on an sw × sh screen,
// the session on the current text (inputSize: its limit), the text offset
// for the box.
func OpenTextInput(title, message, value string, limit, sw, sh int, titleColour, colour uint8, owned bool, onOK func(string)) *TextInputWindow {
	w := &TextInputWindow{X: (sw - tiW) / 2, Y: (sh - tiH) / 2, Title: title, Message: message, TitleColour: titleColour, Colour: colour,
		Owned: owned, Session: NewInputSession(value, limit), OnOK: onOK}
	w.Session.CalculateTextOffset(tiInput[2] - 2)
	return w
}

func (w *TextInputWindow) in(r [4]int, px, py int) bool {
	return px >= w.X+r[0] && px < w.X+r[0]+r[2] && py >= w.Y+r[1] && py < w.Y+r[1]+r[3]
}

// HandleClick is onMouseUp: close, or OK (sanitizeInput, the caller's
// textInput with the text, closed).
func (w *TextInputWindow) HandleClick(px, py int) {
	switch {
	case w.in(tiClose, px, py):
		w.Closed = true
	case w.in(tiOK, px, py):
		w.Closed = true
		if w.OnOK != nil {
			w.OnOK(w.Session.Sanitized())
		}
	}
}

// Contains reports whether a screen point is on the window.
func (w *TextInputWindow) Contains(px, py int) bool {
	return w.in([4]int{0, 0, tiW, tiH}, px, py)
}

// Update is onUpdate: the caret's blink counter.
func (w *TextInputWindow) Update() { w.Session.CursorFrame++ }

// Keys is keyUp: Enter is OK, Escape the close button; else the typed
// characters and editing keys go to the session, re-offset when needed.
func (w *TextInputWindow) Keys(typed []rune, keys []int, enter, escape bool) {
	switch {
	case enter:
		w.HandleClick(w.X+tiOK[0], w.Y+tiOK[1])
		return
	case escape:
		w.Closed = true
		return
	}
	if len(typed) == 0 && len(keys) == 0 {
		return
	}
	for _, r := range typed {
		w.Session.Type(r)
	}
	for _, k := range keys {
		w.Session.Key(k)
	}
	w.Session.CursorFrame = 0
	if cw := tiInput[2] - 2; w.Session.NeedsReoffsetting(cw) {
		w.Session.CalculateTextOffset(cw)
	}
}

// Draw is prepareDraw and draw.
func (w *TextInputWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	DrawWindowFrame(screen, w.X, w.Y, tiW, tiH, w.TitleColour, w.Owned)
	draw8(screen, w.X, w.Y, tiW, tiH, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 15, tiW-1, tiH-1, gfx8.InsetColour{Map: sec}, 0) // panel
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, tiClose[0], tiClose[1], gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		b, o := tiInput, tiOK
		gfx8.FillRectInset(rt, b[0], b[1], b[0]+b[2]-1, b[1]+b[3]-1, gfx8.InsetColour{Map: sec}, gfx8.InsetBorderInset|gfx8.InsetFillDarker)
		gfx8.FillRectInset(rt, o[0], o[1], o[0]+o[2]-1, o[1]+o[3]-1, gfx8.InsetColour{Map: sec}, 0)
	})
	if w.Owned {
		DrawWindowCaptionColour(screen, w.X+1, w.Y+1, tiW-2, w.Title, int(w.TitleColour))
	} else {
		DrawWindowCaption(screen, w.X+1, w.Y+1, tiW-2, w.Title)
	}
	pal := g1.GetPalette()
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	for i, line := range WrapTextBold(w.Message, tiW-8) { // wcolour2_stringid, centred, wrapped
		lw, _ := MeasureTextBold(line)
		DrawTextBold(screen, line, w.X+tiW/2-lw/2, w.Y+30+10*i, win2)
	}
	b := tiInput
	clip := screen.SubImage(image.Rect(w.X+b[0]+1, w.Y+b[1]+1, w.X+b[0]+b[2]-1, w.Y+b[1]+b[3]-1)).(*ebiten.Image)
	tx, ty := w.X+b[0]+1+w.Session.XOffset, w.Y+b[1]+1+1
	DrawTextBold(clip, string(w.Session.Buffer), tx, ty, black)
	if w.Session.CursorFrame%32 < 16 {
		cx, _ := MeasureTextBold(string(w.Session.Buffer[:w.Session.CursorPosition]))
		if c, ok := clip.SubImage(image.Rect(tx+cx, ty, tx+cx+1, ty+10)).(*ebiten.Image); ok {
			c.Fill(win2) // fillRect(x, 1, x, 10, getShade(secondary, 9))
		}
	}
	limit := FillArgs(i18n.Str(i18n.NumCharactersLeftIntInt), Comma16(len(w.Session.Buffer)), Comma16(w.Session.InputLenLimit))
	lw := measureFormatted(limit)
	drawFormatted(screen, w.X+tiLimit[0]+tiLimit[2]-lw, w.Y+tiLimit[1], limit, win2, black)
	ok := i18n.Label(i18n.LabelButtonOk)
	ow, _ := MeasureTextBold(ok)
	DrawTextBold(screen, ok, w.X+tiOK[0]+(tiOK[2]-1)/2-ow/2, w.Y+tiOK[1], win2)
}

// measureFormatted is the width of a game string drawn by drawFormatted
// (its codes skipped).
func measureFormatted(s string) int {
	n := 0
	for len(s) > 0 {
		if s[0] == '{' {
			if end := indexByte(s, '}'); end >= 0 {
				s = s[end+1:]
				continue
			}
		}
		next := indexByte(s, '{')
		if next < 0 {
			next = len(s)
		}
		w, _ := MeasureTextBold(s[:next])
		n += w
		s = s[next:]
	}
	return n
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
