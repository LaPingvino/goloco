package gfx8

// InvalidationGrid is Graphics/InvalidationGrid.cpp: the screen in blocks
// (SoftwareDrawingEngine::resize: 64 × 8 pixels), each marked when a region
// over it is invalidated (Gfx::invalidateRegion), then walked column by column
// with each vertical run of marked blocks handed over as one rectangle and
// cleared (traverseDirtyCells) — what renderDirtyRegions repaints.
type InvalidationGrid struct {
	blockWidth, blockHeight   int
	columnCount, rowCount     int
	screenWidth, screenHeight int
	blocks                    []uint8
}

// Block size SoftwareDrawingEngine uses (widthShift 6, heightShift 3).
const (
	InvalidationBlockWidth  = 1 << 6
	InvalidationBlockHeight = 1 << 3
)

// Reset is reset: a grid for a width × height screen, every block clean.
func (g *InvalidationGrid) Reset(width, height, blockWidth, blockHeight int) {
	g.blockWidth, g.blockHeight = blockWidth, blockHeight
	g.columnCount = width/blockWidth + 1
	g.rowCount = height/blockHeight + 1
	g.screenWidth, g.screenHeight = width, height
	n := g.columnCount * g.rowCount
	if cap(g.blocks) >= n {
		g.blocks = g.blocks[:n]
		clear(g.blocks)
	} else {
		g.blocks = make([]uint8, n)
	}
}

// Invalidate is invalidate: the region clamped to the screen, then every block
// from left / blockWidth to right / blockWidth (inclusive — a region ending on
// a block edge marks the next block too, as upstream) and the same in y.
func (g *InvalidationGrid) Invalidate(left, top, right, bottom int) {
	left, top = max(left, 0), max(top, 0)
	right, bottom = min(right, g.screenWidth), min(bottom, g.screenHeight)
	if left >= right || top >= bottom {
		return
	}
	left /= g.blockWidth
	right /= g.blockWidth
	top /= g.blockHeight
	bottom /= g.blockHeight
	for y := top; y <= bottom; y++ {
		row := g.blocks[y*g.columnCount:]
		for x := left; x <= right; x++ {
			row[x] = 0xFF
		}
	}
}

// TraverseDirtyCells is traverseDirtyCells: per column, top to bottom, each run
// of marked blocks as (left, top, right, bottom) in pixels (clipped to the
// screen; a run starting off-screen is skipped), cleared as it is handed over.
func (g *InvalidationGrid) TraverseDirtyCells(f func(left, top, right, bottom int)) {
	cc := g.columnCount
	for column := range cc {
		for row := 0; row < g.rowCount; row++ {
			start := row * cc
			if g.blocks[start+column] == 0 {
				continue
			}
			end, dirty := start, 0
			for row+dirty+1 < g.rowCount && g.blocks[end+column+cc] != 0 {
				dirty++
				end += cc
			}
			for off := start; off <= end; off += cc {
				g.blocks[off+column] = 0
			}
			left, top := column*g.blockWidth, row*g.blockHeight
			right, bottom := (column+1)*g.blockWidth, (row+dirty+1)*g.blockHeight
			if left < g.screenWidth && top < g.screenHeight {
				f(left, top, min(right, g.screenWidth), min(bottom, g.screenHeight))
			}
		}
	}
}
