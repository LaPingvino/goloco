package olmath

import "testing"

func TestIntegerSineCosine(t *testing.T) {
	// Test some known values: 0, quarter, half, three-quarters
	if got := IntegerSinePrecisionHigh(0, 32768); got != 0 {
		t.Fatalf("sine(0) expected 0 got %d", got)
	}
	if got := IntegerSinePrecisionHigh(4096, 32768); got != 32768 {
		t.Fatalf("sine(pi/2) expected 32768 got %d", got)
	}
	if got := IntegerSinePrecisionHigh(8192, 32768); got != 0 {
		t.Fatalf("sine(pi) expected 0 got %d", got)
	}
	if got := IntegerSinePrecisionHigh(12288, 32768); got != -32768 {
		t.Fatalf("sine(3pi/2) expected -32768 got %d", got)
	}

	// Cosine checks
	if got := IntegerCosinePrecisionHigh(0, 32768); got != 32768 {
		t.Fatalf("cosine(0) expected 32768 got %d", got)
	}
	if got := IntegerCosinePrecisionHigh(4096, 32768); got != 0 {
		t.Fatalf("cosine(pi/2) expected 0 got %d", got)
	}
}
