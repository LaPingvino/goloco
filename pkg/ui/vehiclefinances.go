package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/LaPingvino/goloco/pkg/gfx8"
)

// The Vehicle window's Finances tab (Ui/Windows/Vehicle.cpp Finances,
// 0x004B56CE prepareDraw / 0x004B576C draw), item 5P.g.

// VehicleFinancesSize is Finances::kMinWindowSize (= its maximum).
var VehicleFinancesSize = [2]int{400, 202}

// VehicleFinances is what Finances::draw prints, each line the game's string
// with its arguments filled in (FillArgs): the last income — its date line
// (last_income_on_date, or last_income_na) and a line per cargo
// (transported_blocks_in_days) —, the last journey's average speed (empty:
// none, not shown), the monthly running cost and profit, the sale value.
type VehicleFinances struct {
	IncomeHeader string
	Income       []string
	AverageSpeed string
	RunningCost  string
	Profit       string
	SaleValue    string
}

// VehicleFinancesLayout is Finances::widgets after prepareDraw: the common
// widgets to the window's size, the tabs left-aligned.
func VehicleFinancesLayout(width, height int, hiddenTabs uint8) (w [vwTabRoute + 1]BVWidget) {
	vehicleCommonWidgets(w[:])
	w[vwFrame].R, w[vwFrame].B = width-1, height-1
	w[vwPanel].R, w[vwPanel].B = width-1, height-1
	w[vwCaption].R = width - 2
	w[vwClose].L, w[vwClose].R = width-15, width-3
	leftAlignTabs(w[:], hiddenTabs)
	return w
}

// FillArgs fills a game string's argument codes in order — every {…} but
// the colours and NEWLINE — with args; a {STRINGID} takes its string whole,
// so the codes inside it come next (as FormatArguments are read).
func FillArgs(s string, args ...string) string {
	var b strings.Builder
	for len(s) > 0 {
		i := strings.IndexByte(s, '{')
		if i < 0 {
			b.WriteString(s)
			break
		}
		b.WriteString(s[:i])
		end := strings.IndexByte(s[i:], '}')
		if end < 0 {
			b.WriteString(s[i:])
			break
		}
		code := s[i+1 : i+end]
		s = s[i+end+1:]
		if strings.HasPrefix(code, "COLOUR") || code == "NEWLINE" || strings.HasPrefix(code, "NEWLINE ") || strings.HasPrefix(code, "SPRITE ") || strings.HasPrefix(code, "MOVE_X ") || strings.HasSuffix(code, "FONT") || len(args) == 0 { // not arguments: colours, line breaks, inline sprites, pen moves, font changes
			b.WriteString("{" + code + "}")
			continue
		}
		if code == "STRINGID" {
			s = args[0] + s
		} else {
			b.WriteString(args[0])
		}
		args = args[1:]
	}
	return b.String()
}

// objectArgCodes are the argument control codes 123–143 (Formatting.h
// ControlCodes, int32_grouped …) by name, as a string table's {…} has them.
var objectArgCodes = [...]string{"COMMA32", "INT32", "COMMA2DP32", "COMMA2DP32", "COMMA16", "UINT16", "CURRENCY32",
	"CURRENCY48", "STRINGID", "STRINGID", "STRING", "DATE", "VELOCITY", "POP16", "PUSH16", "TIME_MS", "TIME_HM",
	"LENGTH", "HEIGHT", "POWER", "SPRITE"}

// ObjectString turns an object's string (a cargo's unit name: "{ tonnes of
// Coal", the byte 123 = int32_grouped) into FillArgs' form, each raw argument
// code byte as its {NAME}.
func ObjectString(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if c := s[i]; c >= 123 && int(c)-123 < len(objectArgCodes) {
			b.WriteString("{" + objectArgCodes[c-123] + "}")
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// Comma16 is {COMMA16}: thousands separated by commas.
func Comma16(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := itoa3(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		s = "-" + s
	}
	return s
}

// drawFinances is Finances::draw: the frame, the tabs (Finances active), then
// from (4, 46): the last income — its line, then each cargo's 4 px in, 12
// apart —, 5 px, the average speed when there was one (then 15), the running
// cost (10), the monthly profit; the sale value 14 px above the bottom.
// PARTIAL: drawStringLeftWrapped's wrapping (a cargo line is one line here).
func (w *VehicleWindow) drawFinances(screen *ebiten.Image) {
	g1 := globalG1Font.g1
	prim, _ := g1.GetPaletteMap(int(w.TitleColour))
	sec, err := g1.GetPaletteMap(int(w.Colour))
	if err != nil {
		return
	}
	lw := VehicleFinancesLayout(w.W, w.H, w.State.HiddenTabs)
	DrawWindowFrame(screen, w.X, w.Y, w.W, w.H, w.TitleColour, true)
	draw8(screen, w.X, w.Y, w.W, w.H, func(rt *gfx8.RenderTarget) {
		p := lw[vwPanel]
		gfx8.FillRectInset(rt, p.L, p.T, p.R, p.B, gfx8.InsetColour{Map: sec}, 0)
		drawResizeHandle(rt, &w.WinBase, sec)
		if sp, err := gfx8.Decode(&g1.Elements[imgCloseButton]); err == nil {
			gfx8.DrawSpriteNoise(rt, sp, lw[vwClose].L, lw[vwClose].T, gfx8.OpTransparent|gfx8.OpSrc, prim, nil)
		}
		w.drawVehicleTabs(rt, lw[:], vwTabFinances, sec, prim)
	})
	DrawWindowCaptionColour(screen, w.X+1, w.Y+1, w.W-2, w.Title, int(w.TitleColour))
	pal := g1.GetPalette()
	black, win2 := pal[paletteBlack0], pal[gfx8.Shade(sec, 9)]
	f := &w.Finances
	x, y := w.X+4, w.Y+46
	drawFormatted(screen, x, y, f.IncomeHeader, win2, black)
	y += 10
	for _, l := range f.Income {
		drawFormatted(screen, x+4, y, l, win2, black)
		y += 12
	}
	y += 5
	if f.AverageSpeed != "" {
		drawFormatted(screen, x, y, f.AverageSpeed, win2, black)
		y += 10 + 5
	}
	drawFormatted(screen, x, y, f.RunningCost, win2, black)
	y += 10
	drawFormatted(screen, x, y, f.Profit, win2, black)
	drawFormatted(screen, x, w.Y+w.H-14, f.SaleValue, win2, black)
}

// Comma32 is {COMMA32}: as {COMMA16}, for a 32-bit value.
func Comma32(n int64) string { return Comma16(int(n)) }
