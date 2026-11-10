package objects

import "testing"

func table(entries ...any) []byte {
	var b []byte
	for i := 0; i < len(entries); i += 2 {
		b = append(b, byte(entries[i].(int)))
		b = append(b, entries[i+1].(string)...)
		b = append(b, 0)
	}
	return append(b, 0xFF, 0xAA) // 0xAA: past the table
}

// loadStringTable's choice: the selected language, else UK English (even
// after US), else US English, else the last entry seen.
func TestChooseString(t *testing.T) {
	defer func(l uint8) { SelectedLocoLanguage = l }(SelectedLocoLanguage)
	cases := []struct {
		lang uint8
		data []byte
		want string
	}{
		{3, table(0, "uk", 3, "de"), "de"},
		{3, table(1, "us", 0, "uk"), "uk"},
		{3, table(0, "uk", 1, "us"), "uk"},
		{3, table(1, "us", 2, "fr"), "us"},
		{3, table(2, "fr", 5, "it"), "it"},
		{0, table(1, "us", 0, "uk"), "uk"},
	}
	for i, c := range cases {
		SelectedLocoLanguage = c.lang
		n, got := chooseString(c.data)
		if string(got) != c.want || c.data[n] != 0xAA {
			t.Errorf("case %d: %q (length %d), want %q", i, got, n, c.want)
		}
	}
}
