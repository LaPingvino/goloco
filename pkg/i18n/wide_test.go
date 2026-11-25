package i18n

import "testing"

func TestWideLocoBytes(t *testing.T) {
	b := string(AppendLoco(AppendLoco([]byte("A"), '日'), 'é'))
	if b != "A\x02\xe6\x97\xa5\xe9" {
		t.Fatalf("AppendLoco: %q", b)
	}
	if r, n, ok := DecodeWide(b, 1); !ok || r != '日' || n != 4 {
		t.Errorf("DecodeWide: %q %d %v", r, n, ok)
	}
	if _, n, ok := DecodeWide(b[:3], 1); ok || n != 2 { // cut by a length limit
		t.Errorf("cut DecodeWide: %d %v", n, ok)
	}
	if got := string(StripWide([]byte(b))); got != "A?\xe9" {
		t.Errorf("StripWide: %q", got)
	}
	if got := NameLoco("Gare é 日"); got != "Gare \xe9 \x02\xe6\x97\xa5" {
		t.Errorf("NameLoco: %q", got)
	}
	if raw := "Coal \xe9"; NameLoco(raw) != raw { // Loco bytes already
		t.Error("NameLoco changed Loco bytes")
	}
	if HasWide("Café") || HasWide("Čech") || !HasWide("日") || HasWideLoco([]byte("Caf\xe9")) || !HasWideLoco([]byte(b)) {
		t.Error("HasWide")
	}
	if got := ClipLoco(b, 2); got != "A\x02\xe6\x97\xa5" {
		t.Errorf("ClipLoco: %q", got)
	}
}
