package math

import "math"

// IntegerSinePrecisionHigh implements a fixed-point sine similar to the
// original OpenLoco table-based implementation. The original uses a 14-bit
// direction precision (0..16383) where full circle is 16384. The function
// returns the sine multiplied by `magnitude` and divided by 32768 (0x8000)
// to match the original scaling.
// We'll provide a straightforward reference implementation using float64 but
// integer inputs/outputs so behavior is deterministic and easy to test.

// IntegerSinePrecisionHigh computes sin(direction / 16384 * 2pi) * magnitude / 32768
func IntegerSinePrecisionHigh(direction uint16, magnitude int32) int32 {
	// direction is 0..16383 (14 bits), but uint16 may carry larger values; mask
	d := int(direction) & 0x3FFF
	angle := float64(d) * 2.0 * math.Pi / 16384.0
	v := math.Sin(angle) * float64(magnitude) / 32768.0
	// Round toward zero to match integer division semantics
	if v >= 0 {
		return int32(math.Floor(v + 0.5))
	}
	return int32(math.Ceil(v - 0.5))
}

// IntegerCosinePrecisionHigh computes cosine by shifting direction by quarter-turn
func IntegerCosinePrecisionHigh(direction uint16, magnitude int32) int32 {
	// Adding 4096 (kDirectionPrecisionHigh/4) corresponds to pi/2 phase shift
	return IntegerSinePrecisionHigh(uint16((int(direction)+4096)&0x3FFF), magnitude)
}
