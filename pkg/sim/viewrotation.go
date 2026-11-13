package sim

// The view's rotation. Upstream keeps one for every viewport
// (Viewport::getRotation returns WindowManager::getCurrentRotation: "eventually
// this should become a variable of the viewport"), and the simulation reads it
// too — an entity's sprite bounds (EntityBase::moveTo), the name labels
// (Town / Station::updateLabel), the money effect's drift, the waves — so it
// lives on the Map as CurrentRotation.

// RotateVec is Math::Vector::rotate: (x, y) turned by direction (0–3).
func RotateVec(x, y int32, direction uint8) (int32, int32) {
	switch direction & 3 {
	case 1:
		return y, -x
	case 2:
		return -x, -y
	case 3:
		return -y, x
	}
	return x, y
}

// GameToScreen is World::gameToScreen (Map/Tile.cpp): a world position as a
// viewport position at a view rotation.
func GameToScreen(x, y, z int32, rotation uint8) (int32, int32) {
	x, y = RotateVec(x, y, rotation)
	return y - x, ((y + x) >> 1) - z
}

// ViewportCoordToMapCoord is Ui::viewportCoordToMapCoord (Window.cpp): the
// world position at height z a viewport position shows at a view rotation.
func ViewportCoordToMapCoord(x, y, z int32, rotation uint8) (int32, int32) {
	inverse := [4]uint8{0, 3, 2, 1} // inverseRotationMapping
	return RotateVec(y-(x>>1)+z, y+(x>>1)+z, inverse[rotation&3])
}

// GameToScreen is World::gameToScreen at the current rotation
// (WindowManager::getCurrentRotation).
func (m *Map) GameToScreen(x, y, z int32) (int32, int32) {
	return GameToScreen(x, y, z, m.CurrentRotation)
}

// ViewportCoordToMapCoord is Ui::viewportCoordToMapCoord at the current rotation.
func (m *Map) ViewportCoordToMapCoord(x, y, z int32) (int32, int32) {
	return ViewportCoordToMapCoord(x, y, z, m.CurrentRotation)
}

// UpdateSpatialIndex is EntityManager::updateSpatialIndex: every entity moved
// to where it is (moveTo), which invalidates it and recomputes its sprite
// bounds for the current rotation — after the view rotates and after a load
// (a save keeps the bounds of the rotation it was saved in).
func (m *Map) UpdateSpatialIndex() {
	for i := range m.Entities {
		e := &m.Entities[i]
		m.moveTo(e, int32(e.PosX), int32(e.PosY), int32(e.PosZ))
	}
	for _, e := range m.Effects {
		m.moveEffect(e, e.X, e.Y, e.Z)
	}
}

// UpdateLabels is TownManager::updateLabels and StationManager::updateLabels:
// every town's and station's name frame, for the current rotation.
func (m *Map) UpdateLabels(g1 G1Raw) {
	if g1 == nil {
		return
	}
	for i := range m.Towns {
		if m.Towns[i].Name != 0xFFFF {
			m.UpdateTownLabel(i, g1)
		}
	}
	for i := range m.StationNames {
		if m.StationNames[i].Name != 0xFFFF {
			m.UpdateStationLabel(i, g1)
		}
	}
}
