package scene

import (
	"bytes"
	"testing"
)

// BenchmarkRenderPart is one frame's work while a window is open: a third of
// the columns (RenderPart k of 3), against BenchmarkRenderTitleView's whole.
func BenchmarkRenderPart(b *testing.B) {
	r, m, v := titleView(b)
	r.Render(m, v)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt, x0, x1 := r.RenderPart(m, v, i%3, 3)
		r.ToRGBAColumns(rt, x0, x1)
	}
}

// Three parts make the whole frame.
func TestRenderPartWhole(t *testing.T) {
	r, m, v := titleView(t)
	want := append([]uint8(nil), r.Render(m, v).Bits...)
	clear(r.rt.Bits)
	for k := range 3 {
		r.RenderPart(m, v, k, 3)
	}
	if !bytes.Equal(want, r.rt.Bits) {
		t.Error("three parts differ from the whole frame")
	}
	full := append([]byte(nil), r.ToRGBA(r.rt).Pix...)
	x0, x1 := v.W/3, 2*v.W/3
	part := r.ToRGBAColumns(r.rt, x0, x1)
	for y := 0; y < v.H; y++ {
		if !bytes.Equal(part[y*(x1-x0)*4:(y+1)*(x1-x0)*4], full[(y*v.W+x0)*4:(y*v.W+x1)*4]) {
			t.Fatalf("ToRGBAColumns row %d differs from ToRGBA", y)
		}
	}
}
