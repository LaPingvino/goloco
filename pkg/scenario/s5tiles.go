package scenario

// s5tiles.go — the TileElements chunk written back from the tiles: each
// element's 8 bytes as S5TileElement.cpp toSaveElement sets them (the inverse
// of parseTileElements), in tile order, the last flag on each tile's last
// element; ghosts left out (S5.cpp removeGhostElements).
//
// OpenLoco reference: src/OpenLoco/src/S5/S5TileElement.cpp toSaveElement,
// src/OpenLoco/src/S5/S5.cpp prepareGameState / removeGhostElements

const elemFlagGhost = 0x10 // ElementFlags::ghost

// EncodeTileElements is the TileElements chunk of a map of tiles[y][x].
func EncodeTileElements(tiles [][]Tile) []byte {
	out := make([]byte, 0, len(tiles)*len(tiles)*2*tileElementSize)
	for y := range tiles {
		for x := range tiles[y] {
			t := &tiles[y][x]
			start := len(out)
			for _, r := range t.Order {
				e := t.encodeElement(r)
				// removeGhostElements: a ghost goes, unless it is the tile's
				// first element (a tile keeps at least one).
				if e[1]&elemFlagGhost != 0 && len(out) > start {
					continue
				}
				e[1] &^= elementFlagLast
				out = append(out, e[:]...)
			}
			if len(out) == start { // no elements: upstream's tiles always have their surface
				var e [tileElementSize]byte
				e[2], e[3] = t.Height, t.ClearZ
				out = append(out, e[:]...)
			}
			out[len(out)-tileElementSize+1] |= elementFlagLast
		}
	}
	return out
}

func boolBit(v bool, bit uint8) uint8 {
	if v {
		return bit
	}
	return 0
}

// flags is the element's flags byte: the ref's (quarters, ghost, AI, flag_6),
// with the bits the element itself carries (ghost / aiAllocated / flag_6)
// from the element — the simulation keeps those there.
func flagsWith(ref uint8, ghost, ai bool, flag6 *bool) uint8 {
	f := ref&^0x30 | boolBit(ghost, 0x10) | boolBit(ai, 0x20)
	if flag6 != nil {
		f = f&^0x40 | boolBit(*flag6, 0x40)
	}
	return f
}

// encodeElement is toSaveElement for one element of the tile.
func (t *Tile) encodeElement(r ElemRef) [tileElementSize]byte {
	var e [tileElementSize]byte
	e[0] = r.Kind << elementTypeShift
	e[1] = r.Flags
	switch r.Kind {
	case KindSurface:
		e[0] |= boolBit(t.IsIndustrial, 0x80) | boolBit(t.Type6Flag, 0x40)
		e[2], e[3] = t.Height, t.ClearZ
		e[4] = t.Slope
		e[5] = t.Water&0x1F | t.UpdateTimer<<5
		e[6] = t.TerrainIndex&0x1F | t.GrowthStage<<5
		e[7] = t.Variation
	case KindTrack:
		el := &t.Tracks[r.Index]
		e[0] |= el.Rotation&3 | boolBit(el.HasSignal, 0x40) | boolBit(el.HasStation, 0x80)
		e[1] = flagsWith(r.Flags, el.Ghost, el.AiAllocated, &el.IsLast)
		e[2], e[3] = el.BaseZ, el.ClearZ
		e[4] = el.TrackID&0x3F | boolBit(el.HasGhostMods, 0x40) | boolBit(el.HasBridge, 0x80)
		e[5] = el.SequenceIndex&0xF | (el.TrackObjectID&0xF)<<4
		e[6] = boolBit(el.HasLevelCrossing, 0x10) | (el.BridgeID&7)<<5
		e[7] = el.Owner&0xF | el.Mods<<4
	case KindStation:
		el := &t.Stations[r.Index]
		e[0] |= el.Rotation&3 | (el.SequenceIndex&3)<<6
		e[1] = flagsWith(r.Flags, el.Ghost, el.AiAllocated, &el.Flag6)
		e[2], e[3] = el.BaseZ, el.ClearZ
		e[4] = el.Owner & 0xF // unk4SLR4 (the high nibble): toSaveElement doesn't write it
		e[5] = el.ObjectID&0x1F | (el.StationType&7)<<5
		w := el.StationID&0x3FF | uint16(el.BuildingType&0x3F)<<10
		e[6], e[7] = uint8(w), uint8(w>>8)
	case KindSignal:
		el := &t.Signals[r.Index]
		e[0] |= el.Rotation&3 | boolBit(el.RightGhost, 0x40) | boolBit(el.LeftGhost, 0x80)
		e[1] = flagsWith(r.Flags, el.Ghost, el.AiAllocated, nil)
		e[2], e[3] = el.BaseZ, el.ClearZ
		e[4], e[5], e[6], e[7] = el.Sides[0].B4, el.Sides[0].B5, el.Sides[1].B4, el.Sides[1].B5
	case KindBuilding:
		el := &t.Buildings[r.Index]
		e[0] |= el.Rotation&3 | boolBit(el.IsMisc, 0x40) | boolBit(el.IsConstructed, 0x80)
		e[2], e[3] = el.BaseZ, el.ClearZ
		e[4] = el.ObjectID
		e[5] = el.SequenceIndex&3 | el.SectionProgress<<5
		w := uint16(el.Age&0x3F) | uint16(el.Variation&0x1F)<<6 | uint16(el.Colour)<<11
		e[6], e[7] = uint8(w), uint8(w>>8)
	case KindTree:
		el := &t.Trees[r.Index]
		e[0] |= el.Rotation&3 | (el.Quadrant&3)<<6
		e[2], e[3] = el.BaseZ, el.ClearZ
		e[4] = el.TreeObjectID
		e[5] = el.Growth&0xF | (el.Unk5h&0xF)<<4
		e[6] = el.Colour&0x1F | boolBit(el.HasSnow, 0x40) | boolBit(el.IsDying, 0x80)
		e[7] = el.Unk7l&7 | el.Season<<3
	case KindWall:
		el := &t.Walls[r.Index]
		e[0] |= el.Rotation&3 | (el.EdgeSlope&3)<<6
		e[1] = r.Flags&^0x60 | (el.SecondaryColour&0x18)<<2 // setSecondaryColour: its high bits in the flags
		e[2], e[3] = el.BaseZ, el.ClearZ
		e[4] = el.WallObjectID
		e[5] = el.TertiaryColour
		e[6] = el.PrimaryColour&0x1F | (el.SecondaryColour&7)<<5
	case KindRoad:
		el := &t.Roads[r.Index]
		e[0] |= el.Rotation&3 | boolBit(el.HasSignalElement, 0x40) | boolBit(el.HasStationElement, 0x80)
		e[1] = flagsWith(r.Flags, el.Ghost, el.AiAllocated, &el.IsLast)
		e[2], e[3] = el.BaseZ, el.ClearZ
		e[4] = el.RoadID&0xF | (el.LaneOccupation&3)<<4 | boolBit(el.HasGhostMods, 0x40) | boolBit(el.HasBridge, 0x80)
		e[5] = el.SequenceIndex&3 | (el.LevelCrossingObjectID&3)<<2 | (el.RoadObjectID&0xF)<<4
		e[6] = el.AnimFrame&0xF | el.BridgeID<<5
		e[7] = el.Owner&0xF | boolBit(el.Unk7_10, 0x10) | boolBit(el.HasLevelCrossing, 0x20) | (el.Mods&3)<<6
	case KindIndustry:
		el := &t.Industries[r.Index]
		e[0] |= el.Rotation&3 | boolBit(el.IsConstructed, 0x80)
		e[2], e[3] = el.BaseZ, el.ClearZ
		e[4] = el.IndustryID
		e[5] = el.SequenceIndex&3 | el.SectionProgress<<5
		w := uint16(el.Var6Low&0x3F) | uint16(el.BuildingType&0x1F)<<6 | uint16(el.Colour)<<11
		e[6], e[7] = uint8(w), uint8(w>>8)
	}
	return e
}
