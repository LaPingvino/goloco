package sim

// The map window's positions (OpenLoco reference:
// src/OpenLoco/src/Ui/Windows/MapWindow.cpp): a point of the map buffer
// (the scroll view's content, kRenderedMapWidth square: the map turned 45°
// for the view's rotation) and a world position.

// mapWidthWorld is kMapWidth / kMapHeight: kMapColumns tiles of 32.
const mapWidthWorld = mapColumns * tileSize

// MapViewFrameOffsets is kViewFrameOffsetsByRotation (MapWindow.cpp l. 70,
// 0x004FDC4C): where viewport coordinate (0, 0) — the world's origin at
// height 0 — lies on the map buffer, per rotation (a viewport x / 32 and y / 16
// added to it is the map point).
var MapViewFrameOffsets = [4][2]int{
	{mapColumns, 0},
	{MapRenderedWidth, mapColumns},
	{mapColumns, MapRenderedWidth},
	{0, mapColumns},
}

// MapWindowPosToLocation is mapWindowPosToLocation (MapWindow.cpp l. 144): a
// point of the map buffer as a world position at the given view rotation
// (getCurrentRotation). C++'s truncating division and its int16 Pos2 are
// kept; positions off the map are not clamped (getHeight gives them 16).
func MapWindowPosToLocation(x, y int, rotation uint8) (int32, int32) {
	px := ((x + 8) - mapColumns) / 2
	py := (y + 8) / 2
	lx, ly := int16(py-px), int16(px+py)
	lx *= tileSize
	ly *= tileSize
	switch rotation & 3 {
	case 1:
		return int32(int16(mapWidthWorld - 1 - int(ly))), int32(lx)
	case 2:
		return int32(int16(mapWidthWorld - 1 - int(lx))), int32(int16(mapWidthWorld - 1 - int(ly)))
	case 3:
		return int32(ly), int32(int16(mapWidthWorld - 1 - int(lx)))
	}
	return int32(lx), int32(ly)
}

// LocationToMapWindowPos is locationToMapWindowPos (MapWindow.cpp l. 167): a
// world position as a point of the map buffer at the given view rotation. It
// is not mapWindowPosToLocation's inverse: a tile's point goes back to the
// tile 8 further along y (that function's + 8s), as upstream.
func LocationToMapWindowPos(wx, wy int32, rotation uint8) (int, int) {
	x, y := int(wx), int(wy)
	switch rotation & 3 {
	case 3:
		x, y = y, x
		x = mapWidthWorld - 1 - x
	case 2:
		x = mapWidthWorld - 1 - x
		y = mapWidthWorld - 1 - y
	case 1:
		x, y = y, x
		y = mapWidthWorld - 1 - y
	}
	x /= tileSize
	y /= tileSize
	return -x + y + mapColumns, x + y
}
