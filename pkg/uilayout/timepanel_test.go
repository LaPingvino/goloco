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
