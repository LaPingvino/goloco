package core

import "math"

// StrLogicalCmp is Utility::strlogicalcmp (Utility/src/String.cpp): a
// case-insensitive comparison where runs of digits compare as numbers
// ("Train 2" before "Train 10"); '.' and ',' inside a number are skipped, and
// a number past int32 counts as its maximum. Negative, zero or positive as
// s1 sorts before, with or after s2.
func StrLogicalCmp(s1, s2 string) int {
	for {
		switch {
		case s2 == "":
			if s1 != "" {
				return 1
			}
			return 0
		case s1 == "":
			return -1
		case !(isDigit(s1[0]) && isDigit(s2[0])):
			if a, b := upper(s1[0]), upper(s2[0]); a != b {
				return int(a) - int(b)
			}
			s1, s2 = s1[1:], s2[1:]
		default:
			n1, r1 := parseNextNumber(s1)
			n2, r2 := parseNextNumber(s2)
			if n1 != n2 {
				if n1 > n2 {
					return 1
				}
				return -1
			}
			s1, s2 = r1, r2
		}
	}
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func upper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}

// skipDigits skips digits, '.' and ','.
func skipDigits(s string) string {
	for s != "" && (s[0] == '.' || s[0] == ',' || isDigit(s[0])) {
		s = s[1:]
	}
	return s
}

// parseNextNumber reads the number at s's start ('.' and ',' skipped),
// clamped to int32's maximum on overflow, and what follows it.
func parseNextNumber(s string) (int32, string) {
	var num int32
	for s != "" {
		if s[0] != '.' && s[0] != ',' {
			if !isDigit(s[0]) {
				break
			}
			if num > (math.MaxInt32-int32(s[0]-'0'))/10 {
				return math.MaxInt32, skipDigits(s)
			}
			num = num*10 + int32(s[0]-'0')
		}
		s = s[1:]
	}
	return num, s
}
