package ui

import (
	"strings"
	"time"

	"github.com/LaPingvino/goloco/pkg/gfx8"
	"github.com/LaPingvino/goloco/pkg/paint"
	"github.com/hajimehoshi/ebiten/v2"
)

// Error.cpp's window: a title and a message, wrapped between 70 and 250
// pixels, on a translucent muted dark red panel near the cursor; it closes
// itself after 7 periodic updates (one per 1000 ms of WindowManager::update).
// PARTIAL: the competitor variant (openWithCompetitor), the error sound, and
// the text's translucent white (drawn opaque white).
const (
	errMinWidth = 70
	errMaxWidth = 250
	errPadding  = 4
)

type ErrorWindow struct {
	X, Y, W, H int
	Lines      []string
	Hold       bool // diagnostics: never expires
	opened     time.Time
}

// OpenError is createErrorWindow: title and message joined by a newline
// (either may be empty; both empty: no window), sized and placed round the
// mouse (−width/2, +26; clamped to x ∈ [0, screen−w−40], y ∈ [22, screen−h−40]).
func OpenError(title, message string, mx, my, screenW, screenH int) *ErrorWindow {
	text := title
	if message != "" {
		if text != "" {
			text += "\x05"
		}
		text += message
	}
	if text == "" {
		return nil
	}
	strW := 0
	for _, part := range strings.Split(text, "\x05") { // getStringWidthNewLined
		w, _ := MeasureTextBold(part)
		strW = max(strW, w)
	}
	strW = min(max(strW, errMinWidth), errMaxWidth)
	lines := WrapTextBold(text, strW+errPadding) // wrapString: the widest line wins
	maxW := 0
	for _, l := range lines {
		w, _ := MeasureTextBold(l)
		maxW = max(maxW, w)
	}
	e := &ErrorWindow{Lines: lines, opened: Now()}
	e.W = maxW + 2*errPadding
	e.H = len(lines)*10 + 2*errPadding
	e.X = min(max(mx-e.W/2, 0), screenW-e.W-40)
	e.Y = min(max(my+26, 22), screenH-e.H-40)
	return e
}

// Expired is onPeriodicUpdate's close: var_846 reaching 7.
func (e *ErrorWindow) Expired() bool { return !e.Hold && Now().Sub(e.opened) >= 7*time.Second }

// Draw is the panel (frame: the window's size) and the lines centred from
// y+padding (drawStringCentredRaw).
func (e *ErrorWindow) Draw(screen *ebiten.Image) {
	if globalG1Font == nil {
		return
	}
	g1 := globalG1Font.g1
	cmap, err := g1.GetPaletteMap(colourMutedDarkRed)
	if err != nil {
		return
	}
	var trans [3][]uint8
	for k := range trans {
		trans[k], _ = g1.GetPaletteMap(int(paint.TranslucentColour[colourMutedDarkRed][k]))
	}
	c := gfx8.InsetColour{Map: cmap, Translucent: true, Trans: trans}
	draw8(screen, e.X, e.Y, e.W, e.H, func(rt *gfx8.RenderTarget) {
		gfx8.FillRectInset(rt, 0, 0, e.W-1, e.H-1, c, 0)
	})
	white := g1.GetPalette()[gfx8.Shade(mustMap(colourWhite), 9)]
	for i, l := range e.Lines {
		w, _ := MeasureTextBold(l)
		DrawTextBold(screen, l, e.X+e.W/2-w/2, e.Y+errPadding+10*i, white)
	}
}

// FormatErrorArgs fills a string's arguments the way the error titles use
// them: {POP16} skips one, each {STRINGID} takes the next of args.
func FormatErrorArgs(s string, args ...string) string {
	s = strings.ReplaceAll(s, "{POP16}", "")
	for _, a := range args {
		s = strings.Replace(s, "{STRINGID}", a, 1)
	}
	return s
}
