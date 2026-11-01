package uilayout

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

// Currency48 formats money as {CURRENCY48} with a pound prefix and comma
// thousands separators. PARTIAL: upstream takes the symbol, its position, the
// separator and the factor from the scenario's currency object.
func Currency48(v int64) string {
	neg := v < 0
	if neg {
		v = -v
	}
	s := itoa(int(v))
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	if neg {
		return "-£" + out
	}
	return "£" + out
}
