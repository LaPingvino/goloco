package render

// Isometric helpers: convert tile coordinates (tx,ty) to screen pixel coords.
// Using standard 2:1 isometric projection with tile width/height.

// TileToScreen converts tile (col,row) to screen pixel (x,y) with given tileW,tileH and origin offset.
func TileToScreen(col, row, tileW, tileH, originX, originY int) (int, int) {
	// isometric: x = (col - row) * tileW/2 ; y = (col + row) * tileH/2
	x := (col - row) * (tileW / 2)
	y := (col + row) * (tileH / 2)
	return originX + x, originY + y
}

// ScreenToTile approximates inverse mapping. Works for primary grid snapping.
func ScreenToTile(x, y, tileW, tileH, originX, originY int) (int, int) {
	x = x - originX
	y = y - originY
	col := (x/(tileW/2) + y/(tileH/2)) / 2
	row := (y/(tileH/2) - x/(tileW/2)) / 2
	return col, row
}
