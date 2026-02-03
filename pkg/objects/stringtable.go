package objects

// parseStringTable extracts the first string from a Locomotion object string table.
//
// OpenLoco reference: src/OpenLoco/src/Objects/ObjectStringTable.cpp
//
//	loadStringTable() at address 0x00472172
//
// Format: Sequence of [languageID byte][null-terminated string], ending with 0xFF byte.
//
//	Example: 09 "Hello\0" 09 "World\0" FF
//
// Returns: (tableLength, firstString)
func parseStringTable(data []byte) (int, string) {
	if len(data) < 1 {
		return 0, ""
	}

	pos := 0
	firstString := ""

	// Iterate through language/string pairs until we hit 0xFF terminator
	for pos < len(data) {
		if data[pos] == 0xFF {
			// Found terminator - include it in the length
			return pos + 1, firstString
		}

		// Skip language ID byte
		pos++
		if pos >= len(data) {
			break
		}

		// Find null terminator of string
		stringStart := pos
		for pos < len(data) && data[pos] != 0 {
			pos++
		}

		// Extract string if this is the first one
		if firstString == "" && pos > stringStart {
			firstString = cleanString(data[stringStart:pos])
		}

		// Skip the null terminator
		if pos < len(data) {
			pos++
		}
	}

	// If we reached end without finding 0xFF, return what we have
	return pos, firstString
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
