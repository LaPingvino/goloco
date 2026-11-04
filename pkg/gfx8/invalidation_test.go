package gfx8

import (
	"reflect"
	"testing"
)

func dirtyRects(g *InvalidationGrid) [][4]int {
	var out [][4]int
	g.TraverseDirtyCells(func(l, t, r, b int) { out = append(out, [4]int{l, t, r, b}) })
	return out
}

func TestInvalidationGrid(t *testing.T) {
	var g InvalidationGrid
	g.Reset(800, 600, InvalidationBlockWidth, InvalidationBlockHeight)
	if g.columnCount != 13 || g.rowCount != 76 {
		t.Fatalf("grid %dx%d, want 13x76 (width / 64 + 1, height / 8 + 1)", g.columnCount, g.rowCount)
	}
	if r := dirtyRects(&g); r != nil {
		t.Fatalf("clean grid gave %v", r)
	}
	// inside one block: that block only
	g.Invalidate(10, 10, 20, 12)
	if r := dirtyRects(&g); !reflect.DeepEqual(r, [][4]int{{0, 8, 64, 16}}) {
		t.Errorf("one block: %v", r)
	}
	// traversal clears: nothing the second time
	if r := dirtyRects(&g); r != nil {
		t.Errorf("not cleared: %v", r)
	}
	// ending exactly on a block edge marks the next block too (upstream's inclusive right / bottom)
	g.Invalidate(0, 0, 64, 8)
	if r := dirtyRects(&g); !reflect.DeepEqual(r, [][4]int{{0, 0, 64, 16}, {64, 0, 128, 16}}) {
		t.Errorf("edge: %v", r)
	}
	// a tall region: one rectangle per column (the vertical run merged), columns left to right
	g.Invalidate(70, 20, 140, 50)
	if r := dirtyRects(&g); !reflect.DeepEqual(r, [][4]int{{64, 16, 128, 56}, {128, 16, 192, 56}}) {
		t.Errorf("tall: %v", r)
	}
	// clamped to the screen; the last column / row clipped to its size
	g.Invalidate(-50, -50, 2000, 3)
	r := dirtyRects(&g)
	if len(r) != 13 || r[0] != [4]int{0, 0, 64, 8} || r[12] != [4]int{768, 0, 800, 8} {
		t.Errorf("clamped: %d rects, first %v last %v", len(r), r[0], r[len(r)-1])
	}
	// empty regions mark nothing
	g.Invalidate(100, 100, 100, 200)
	g.Invalidate(900, 0, 1000, 10)
	if r := dirtyRects(&g); r != nil {
		t.Errorf("empty regions marked %v", r)
	}
	// two separate runs in one column stay two rectangles
	g.Invalidate(0, 0, 10, 4)
	g.Invalidate(0, 40, 10, 44)
	if r := dirtyRects(&g); !reflect.DeepEqual(r, [][4]int{{0, 0, 64, 8}, {0, 40, 64, 48}}) {
		t.Errorf("two runs: %v", r)
	}
}
