package world

// SoundAt is the world-position half of Audio.cpp playSound: where the main
// view shows (x, y, z) (findBestViewportForSound's main-window case), the
// extra volume calculateVolumeFromViewport gives it (−1024 per zoom level,
// ×256 below the surface by more than 5), and the screen x it pans to. ok is
// false when the main view doesn't show the point, and the sound is skipped.
// PARTIAL: the other windows' viewports (vehicle, station windows) as a second
// choice.
func (w *World) SoundAt(x, y, z int32) (volume int32, screenX int, ok bool) {
	vx, vy := w.Map.GameToScreen(x, y, z)
	if w.viewW == 0 || w.viewH == 0 {
		return 0, 0, false
	}
	rx, ry := int(vx)-int(w.camX), int(vy)-int(w.camY) // Viewport::contains: view x … x + view width
	if rx < 0 || ry < 0 || rx >= w.viewW<<w.zoom || ry >= w.viewH<<w.zoom {
		return 0, 0, false
	}
	if t := w.Map.TileAt(int(x/32), int(y/32)); t != nil {
		zVol := 0
		if int32(t.BaseZ)*4-5 > z {
			zVol = 8
		}
		volume = ((-1024*int32(max(w.zoom, 0)) - 1) << zVol) + 1
	}
	return volume, rx >> w.zoom, true
}

// VehicleViewAt is findBestViewportForEntity's main-window case for a vehicle
// part at (x, y, z) (VehicleAudio.cpp): where the main view shows it on the
// screen, the view's zoom, and isUnderground (the tile's surface above z).
// ok is false when the main view doesn't show it. PARTIAL: the main view
// following the train (savedView.entityId) and the other windows' viewports.
func (w *World) VehicleViewAt(x, y, z int32) (sx, sy int32, zoom int, underground, ok bool) {
	vx, vy := w.Map.GameToScreen(x, y, z)
	if w.viewW == 0 || w.viewH == 0 {
		return 0, 0, 0, false, false
	}
	rx, ry := int(vx)-int(w.camX), int(vy)-int(w.camY)
	if rx < 0 || ry < 0 || rx >= w.viewW<<w.zoom || ry >= w.viewH<<w.zoom {
		return 0, 0, 0, false, false
	}
	if x >= 0 && y >= 0 && x < int32(w.Map.Width)*32 && y < int32(w.Map.Height)*32 { // validCoords
		if t := w.Map.TileAt(int(x/32), int(y/32)); t != nil {
			underground = int32(t.BaseZ)*4 > z
		}
	}
	return int32(rx >> w.zoom), int32(ry >> w.zoom), w.zoom, underground, true
}
