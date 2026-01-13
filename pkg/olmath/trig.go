package olmath

import "math"

// IntegerSinePrecisionHigh computes a fixed-point sine value similar to
// OpenLoco's high-precision integer sine. direction is 0..16383 (14-bit).
// The returned value is (sin(angle) * magnitude) / 32768 rounded to nearest int32.
func IntegerSinePrecisionHigh(direction uint16, magnitude int32) int32 {
	d := int(direction) & 0x3FFF
	angle := float64(d) * 2.0 * math.Pi / 16384.0
	v := math.Sin(angle) * float64(magnitude)
	if v >= 0 {
		return int32(math.Floor(v + 0.5))
	}
	return int32(math.Ceil(v - 0.5))
}

// IntegerCosinePrecisionHigh computes cosine by shifting direction by a quarter-turn.
func IntegerCosinePrecisionHigh(direction uint16, magnitude int32) int32 {
	return IntegerSinePrecisionHigh(uint16((int(direction)+4096)&0x3FFF), magnitude)
}
