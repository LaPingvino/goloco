package sim

// The vehiclePlace game command (VehiclePlace.cpp 0x004B01B6) with
// validateTrackPlacement / validateRoadPlacement.
// PARTIAL: the placedown sound, the error texts (which piece or extra is
// missing), the viewport invalidation of convertGhost.

// VehiclePlacement is VehiclePlacementArgs: where the front goes (pos, a
// world position with z in world units), on which piece, how far along it,
// and convertGhost (a ghost placement made real).
type VehiclePlacement struct {
	X, Y, Z           int32
	TrackAndDirection uint16
	TrackProgress     uint16
	Head              uint16
	ConvertGhost      bool
}

// checkCompanyCompatibility is GameCommands::checkCompanyCompatibility: the
// neutral company's, or the updating company is that one or neutral.
func checkCompanyCompatibility(c *Cmd, owner uint8) bool {
	return owner == CompanyNeutral || c.Company == owner || c.Company == CompanyNeutral
}

// VehiclePlace is the vehiclePlace game command: 0 and true, or false when
// refused (another company's vehicle, already placed, no cars, no fitting
// piece there, a missing track / road extra, or in the way).
func (m *Map) VehiclePlace(c *Cmd, flags uint8, a VehiclePlacement) (int32, bool) {
	if a.Head == entityIDNull {
		return 0, false
	}
	tr, err := m.TrainOf(a.Head)
	if err != nil {
		return 0, false // Bad vehicle structure
	}
	head := &m.Entities[tr.Head]
	if !checkCompanyCompatibility(c, head.Owner) {
		return 0, false
	}
	if !a.ConvertGhost && (head.TileX != -1 || len(tr.Cars) == 0) {
		return 0, false
	}
	if flags&CmdApply == 0 {
		return 0, true
	}
	if a.ConvertGhost {
		m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 &^= flags38IsGhost })
		return 0, true
	}
	ok := false
	if head.Mode == 1 {
		ok = m.validateRoadPlacement(c, a.X, a.Y, a.Z, a.TrackAndDirection, tr)
	} else {
		ok = m.validateTrackPlacement(c, a.X, a.Y, a.Z, a.TrackAndDirection, tr)
	}
	if !ok {
		return 0, false
	}
	if m.PlaceDownVehicle(a.Head, a.X, a.Y, uint8(a.Z/4), a.TrackAndDirection, a.TrackProgress, c.Objects) != PlaceDownOkay {
		return 0, false
	}
	tr, _ = m.TrainOf(a.Head)
	head = &m.Entities[tr.Head]
	head.VehicleFlags |= VehicleFlagCommandStop
	head.ManualPower = -40
	if flags&CmdGhost != 0 {
		m.eachComponent(tr, func(i int) { m.Entities[i].Flags38 |= flags38IsGhost })
	}
	return 0, true
}

// pieceStart is the placement's piece start: a reversed piece measured back
// from its far end (getUnk*.pos, kRotationOffset), then the first sub-piece's z.
func pieceStart(x, y, z int32, c TrackCoord, firstZ int16) (int32, int32, int32) {
	x, y, z = x-int32(c.DX), y-int32(c.DY), z-int32(c.DZ)
	if c.RotEnd < 12 {
		x += int32(RotationOffset[c.RotEnd][0])
		y += int32(RotationOffset[c.RotEnd][1])
	}
	return x, y, z + int32(firstZ)
}

// validateTrackPlacement is VehiclePlace.cpp's: a track element of the
// vehicle's track type, this piece and direction at this height (not ghost or
// AI), the company's, carrying every extra the vehicle needs (var_53).
func (m *Map) validateTrackPlacement(c *Cmd, x, y, z int32, tad uint16, tr Train) bool {
	head := &m.Entities[tr.Head]
	id := tad >> 3 & 0x3F
	if int(id) >= len(TrackPieces) || len(TrackPieces[id]) == 0 {
		return false
	}
	if tad&(1<<2) != 0 {
		tad &^= 1 << 2
		x, y, z = pieceStart(x, y, z, TrackCoordinates[tad&0x1FF], 0)
	}
	z += int32(TrackPieces[id][0].Z)
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return false
	}
	for _, el := range t.Tracks {
		if int32(el.BaseZ)*4 != z || el.AiAllocated || el.Ghost || el.TrackObjectID != head.TrackType ||
			uint16(el.TrackID) != id || uint16(el.Rotation) != tad&3 {
			continue
		}
		if !checkCompanyCompatibility(c, el.Owner) {
			return false
		}
		return el.Mods&head.HeadVar53 == head.HeadVar53 // this_vehicle_requires_stringid
	}
	return false // can_only_be_placed_on_stringid
}

// validateRoadPlacement is VehiclePlace.cpp's: a road element of the
// vehicle's road type (or, for an any-road-type vehicle, a road type all such
// vehicles may use), this piece and direction at this height (not ghost or
// AI), the company's unless every company may use that road, with every extra
// the vehicle needs.
func (m *Map) validateRoadPlacement(c *Cmd, x, y, z int32, tad uint16, tr Train) bool {
	head := &m.Entities[tr.Head]
	id := tad >> 3 & 0xF
	if int(id) >= len(RoadPieces) || len(RoadPieces[id]) == 0 {
		return false
	}
	if tad&(1<<2) != 0 {
		tad &^= 1 << 2
		x, y, z = pieceStart(x, y, z, RoadCoordinates[tad&0x7F], 0)
	}
	z += int32(RoadPieces[id][0].Z)
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return false
	}
	for _, el := range t.Roads {
		if int32(el.BaseZ)*4 != z || el.AiAllocated || el.Ghost {
			continue
		}
		if el.RoadObjectID != head.TrackType {
			if head.TrackType != 0xFF || m.RoadAnyTypeCompatible&(1<<el.RoadObjectID) == 0 {
				continue
			}
		}
		if uint16(el.RoadID) != id || uint16(el.Rotation) != tad&3 {
			continue
		}
		if m.RoadUsableByAll&(1<<el.RoadObjectID) == 0 && !checkCompanyCompatibility(c, el.Owner) {
			return false
		}
		return el.Mods&head.HeadVar53 == head.HeadVar53
	}
	return false
}

// VehicleIsGhost reports whether a vehicle is placed as a ghost (the pickup
// tool's preview): head->tileX != -1 && head->has38Flags(isGhost).
func (m *Map) VehicleIsGhost(headID uint16) bool {
	h := m.entity(headID)
	return h != nil && h.TileX != -1 && h.Flags38&flags38IsGhost != 0
}
