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
		r.ToRGBA(r.RenderPart(m, v, i%3, 3))
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
}
