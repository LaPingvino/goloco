package objects

// chooseString is ObjectStringTable.cpp loadStringTable's choice over one
// [langID][str\0]…[0xFF] table: the SelectedLocoLanguage string, else the
// English backup (english_uk always; english_us only while no backup is
// set), else anyStr — updated on every entry until an English or target
// string has been seen. It returns the table's length (past the 0xFF) and the
// chosen string's bytes as stored (control codes kept; nil for none).
// OpenLoco reference: src/OpenLoco/src/Objects/ObjectStringTable.cpp
// (loadStringTable, 0x00472172).
func chooseString(data []byte) (int, []byte) {
	var target, eng, anyStr []byte
	pos := 0
	for pos < len(data) && data[pos] != 0xFF {
		lang := data[pos]
		pos++
		start := pos
		for pos < len(data) && data[pos] != 0 {
			pos++
		}
		str := data[start:pos]
		if pos < len(data) {
			pos++ // the null
		}
		if lang == 0 { // english_uk
			eng = str
		} else if lang == 1 && eng == nil { // english_us
			eng = str
		}
		if lang == SelectedLocoLanguage {
			target = str
		}
		if eng == nil && target == nil {
			anyStr = str
		}
	}
	if pos < len(data) {
		pos++ // the 0xFF
	}
	switch {
	case target != nil:
		return pos, target
	case eng != nil:
		return pos, eng
	}
	return pos, anyStr
}

// parseStringTable is chooseString with the string cleaned of control codes
// (printable ASCII only). Returns (tableLength, string).
func parseStringTable(data []byte) (int, string) {
	n, raw := chooseString(data)
	return n, cleanString(raw)
}

// cleanString removes non-printable characters and 0xFF format codes from a string
func cleanString(data []byte) string {
	result := make([]byte, 0, len(data))
	for _, b := range data {
		if b == 0xFF {
			continue // Skip FF bytes (format codes)
		}
		if b == 0 {
			break
		}
		if isPrintable(b) {
			result = append(result, b)
		}
	}
	return string(result)
}

func isPrintable(b byte) bool {
	return b >= 0x20 && b < 0x7F
}

// chosenRawString is the table's chosen string as stored (control codes
// kept), a copy.
func chosenRawString(data []byte) []byte {
	_, raw := chooseString(data)
	return append([]byte(nil), raw...)
}
