package uilayout

import "testing"

func TestDateDMY(t *testing.T) {
	for _, c := range []struct {
		d    int
		want string
	}{{1, "1st January 1900"}, {2, "2nd January 1900"}, {3, "3rd January 1900"}, {11, "11th January 1900"}, {22, "22nd January 1900"}, {31, "31st January 1900"}} {
		if got := DateDMY(c.d, "January", 1900); got != c.want {
			t.Errorf("%d: %q", c.d, got)
		}
	}
}

func TestTimePanelSpeedAt(t *testing.T) {
	sw, sh := 800, 600
	wx, wy := TimePanelPos(sw, sh)
	if i := TimePanelSpeedAt(wx+40, wy+20, sw, sh); i != 1 {
		t.Errorf("normal button: %d", i)
	}
	if i := TimePanelSpeedAt(wx+5, wy+20, sw, sh); i != -1 {
		t.Errorf("left of the buttons: %d", i)
	}
}

func TestCurrency48(t *testing.T) {
	for _, c := range []struct {
		v    int64
		want string
	}{{0, "£0"}, {999, "£999"}, {1000, "£1,000"}, {2500000, "£2,500,000"}, {-12345, "-£12,345"}} {
		if got := Currency48(c.v); got != c.want {
			t.Errorf("%d: %q", c.v, got)
		}
	}
}

func TestCurrencyObjectFormat(t *testing.T) {
	defer func(c CurrencyFormat) { Currency = c }(Currency)
	sign := "\xa3"
	for _, c := range []struct {
		prefix, suffix string
		sep, factor    uint8
		v              int64
		want           string
	}{
		{sign, "", 0, 1, 1234, "\u00a32,468"},   // CURRDOLL: factor 1
		{"", sign, 0, 2, -1000, "-4,000\u00a3"}, // CURRDMRK: suffix, factor 2
		{sign, "", 2, 0, 1234, "\u00a3123,400"}, // CURRLIRA: separator 2
		{"", sign, 1, 0, 5, "50\u00a3"},         // CURRFREN: separator 1
	} {
		SetCurrency(c.prefix, c.suffix, c.sep, c.factor)
		if got := Currency48(c.v); got != c.want {
			t.Errorf("%+v: %q", c, got)
		}
	}
}

// ControlCodes::velocity and height in both measurement formats.
func TestVelocityHeight(t *testing.T) {
	defer func(m bool) { Metric = m }(Metric)
	Metric = false
	if Velocity(60) != "60mph" || Height(10) != "160ft" {
		t.Errorf("imperial: %s %s", Velocity(60), Height(10))
	}
	Metric = true
	if Velocity(60) != "97km/h" || Height(10) != "50m" || Velocity(1000) != "1,609km/h" {
		t.Errorf("metric: %s %s %s", Velocity(60), Height(10), Velocity(1000))
	}
	defer func(u bool) { HeightUnits = u }(HeightUnits)
	HeightUnits = true
	if Height(10) != "10 units" {
		t.Errorf("units: %s", Height(10))
	}
}

// calcDate against hand-counted days (1800 is a leap year in Loco's Julian
// calendar: day 59 is 29 February, day 366 is 1 January 1801).
func TestCalcDate(t *testing.T) {
	for _, c := range []struct {
		days    uint32
		y, m, d int
	}{{0, 1800, 0, 1}, {58, 1800, 1, 28}, {59, 1800, 1, 29}, {60, 1800, 2, 1}, {365, 1800, 11, 31}, {366, 1801, 0, 1},
		{366 + 59, 1801, 2, 1}, {1461, 1804, 0, 1}} {
		y, m, d := CalcDate(c.days)
		if y != c.y || m != c.m || d != c.d {
			t.Errorf("day %d: %d-%d-%d, want %d-%d-%d", c.days, y, m+1, d, c.y, c.m+1, c.d)
		}
	}
}
