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

// BenchmarkRenderColumn is one column a call (the adaptive parts' cap).
func BenchmarkRenderColumn(b *testing.B) {
	r, m, v := titleView(b)
	r.Render(m, v)
	n := v.W / 32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt, x0, x1 := r.RenderPart(m, v, i%n, n)
		r.ToRGBAColumns(rt, x0, x1)
	}
}

// In play: every part plus the vehicles' columns, over n frames, makes the
// whole frame; and each frame repaints every column a vehicle is in.
func TestRenderPlayWhole(t *testing.T) {
	r, m, v := titleView(t)
	want := append([]uint8(nil), r.Render(m, v).Bits...)
	clear(r.rt.Bits)
	const n = 5
	moving := 0
	for k := range n {
		_, ranges := r.RenderPlay(m, v, k, n)
		moving += len(ranges) - 1
	}
	if !bytes.Equal(want, r.rt.Bits) {
		t.Error("the parts differ from the whole frame")
	}
	t.Logf("%d extra ranges for vehicles over %d frames", moving, n)
}
