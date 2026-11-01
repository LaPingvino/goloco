package uilayout

import (
	"math"
	"strconv"
	"strings"
)

// TimePanel layout — Ui/Windows/TimePanel.cpp: a 140×27 window at the
// bottom-right; widgets relative to it.
const (
	TimePanelW = 140
	TimePanelH = 27
)

// TimePanelPos is the window position on an sw × sh screen.
func TimePanelPos(sw, sh int) (x, y int) { return sw - TimePanelW, sh - TimePanelH }

// Widget rectangles (x, y, w, h) of the time panel.
var (
	TimePanelOuter   = [4]int{0, 0, 140, 29}
	TimePanelInner   = [4]int{2, 2, 136, 25}
	TimePanelMapChat = [4]int{113, 1, 26, 26}
	TimePanelDate    = [4]int{2, 2, 111, 12}
	// TimePanelSpeed: pause, normal, fast forward, extra fast forward.
	TimePanelSpeed = [4][4]int{{18, 15, 20, 12}, {38, 15, 20, 12}, {58, 15, 20, 12}, {78, 15, 20, 12}}
)

// TimePanelSpeedAt returns which speed button (0-3) is at screen point px, py,
// or -1.
func TimePanelSpeedAt(px, py, sw, sh int) int {
	wx, wy := TimePanelPos(sw, sh)
	for i, r := range TimePanelSpeed {
		if px >= wx+r[0] && px < wx+r[0]+r[2] && py >= wy+r[1] && py < wy+r[1]+r[3] {
			return i
		}
	}
	return -1
}

// InTimePanel reports whether the point is on the time panel window.
func InTimePanel(px, py, sw, sh int) bool {
	wx, wy := TimePanelPos(sw, sh)
	return px >= wx && px < wx+TimePanelW && py >= wy && py < wy+TimePanelH
}

// DateDMY is formatDateDMYFull: "<day ordinal> <month> <year>".
func DateDMY(day int, month string, year int) string {
	suffix := "th"
	if day%100 < 11 || day%100 > 13 {
		switch day % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return itoa(day) + suffix + " " + month + " " + itoa(year)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// PlayerInfoPanel layout — Ui/Windows/PlayerInfoPanel.cpp: 140×27 at the
// bottom-left; the same outer/inner frames as the time panel.
const (
	PlayerPanelW = 140
	PlayerPanelH = 27
)

// PlayerPanelPos is the window position on an sw × sh screen.
func PlayerPanelPos(sw, sh int) (x, y int) { return 0, sh - PlayerPanelH }

// Widget rectangles of the player info panel.
var (
	PlayerPanelFace        = [4]int{1, 1, 26, 26}
	PlayerPanelCompanyVal  = [4]int{27, 2, 111, 12}
	PlayerPanelPerformance = [4]int{27, 14, 111, 12}
)

// InPlayerPanel reports whether the point is on the player info panel.
func InPlayerPanel(px, py, sw, sh int) bool {
	wx, wy := PlayerPanelPos(sw, sh)
	return px >= wx && px < wx+PlayerPanelW && py >= wy && py < wy+PlayerPanelH
}

// CurrencyFormat is what formatCurrency reads of the scenario's currency
// object: the prefix and suffix symbol strings (Locomotion characters; the
// symbol itself is character 0xA3, whose glyph LoadCurrency replaces), the
// power-of-ten separator and the power-of-two factor.
type CurrencyFormat struct {
	Prefix, Suffix    string
	Separator, Factor uint8
}

// Currency is the format in use; SetCurrency sets it when a scenario loads.
// Before that it is the symbol as a prefix with no scaling.
var Currency = CurrencyFormat{Prefix: string(rune(0xA3))}

// SetCurrency takes a currency object's raw symbol strings (bytes are
// Locomotion characters) and its separator and factor.
func SetCurrency(prefix, suffix string, separator, factor uint8) {
	Currency = CurrencyFormat{Prefix: locoString(prefix), Suffix: locoString(suffix), Separator: separator, Factor: factor}
}

func locoString(b string) string {
	r := make([]rune, len(b))
	for i := range len(b) {
		r[i] = rune(b[i])
	}
	return string(r)
}

// Currency48 is Formatting.cpp formatCurrency ({CURRENCY48} and {CURRENCY32}):
// a '-' for negatives, the prefix symbol, value × 2^factor × 10^separator
// grouped in threes with ',' (Environment.cpp setLocale), the suffix symbol.
func Currency48(v int64) string {
	out := ""
	if v < 0 {
		out = "-"
		v = -v
	}
	c := Currency
	n := uint64(v) << c.Factor
	for range c.Separator {
		n *= 10
	}
	s := strconv.FormatUint(n, 10)
	g := ""
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			g += ","
		}
		g += string(ch)
	}
	return out + c.Prefix + g + c.Suffix
}

// Metric is Config::measurementFormat == metric (the Options window's
// Regional tab sets it; imperial is upstream's default): speeds in km/h
// (× 1.609375), heights in m (× 5) instead of mph and ft (× 16).
var Metric bool

// Velocity formats ControlCodes::velocity (a speed16 in mph units).
func Velocity(mph int) string {
	if Metric {
		return groupInt(int64(math.Round(float64(mph)*1.609375))) + "km/h"
	}
	return groupInt(int64(mph)) + "mph"
}

// Height formats ControlCodes::height (showHeightAsUnits off).
func Height(v int) string {
	if Metric {
		return groupInt(int64(v*5)) + "m"
	}
	return groupInt(int64(v*16)) + "ft"
}

// groupInt is formatInt32Grouped: thousands separated by ",".
func groupInt(v int64) string {
	s := strconv.FormatInt(v, 10)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		s = "-" + s
	}
	return s
}
