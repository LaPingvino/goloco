package sim

import (
	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// A station's catchment and accepted cargo (World/Station.cpp): the cargo
// map's flag_1 (setCatchmentDisplay 0x00491D70: 4 tiles round each station
// tile, an airport's extents, a dock's 2×2) and doCalcAcceptedCargo
// (0x00492026): the industries' required cargoes (8 points each) and the
// town buildings' produced / consumed quantities in the catchment, a
// cargo accepted at 8 points. Item ST (addTileToStation, Station::update).

const catchmentSize = 4

// CargoSearchState is Station.cpp's: the filter, the scores, the produced
// cargo types, the industry per cargo, and whether a station tile was found
// (cargoFilterHasBeenRecalculated).
type CargoSearchState struct {
	Filter       uint32
	Score        [32]uint32
	Produced     uint32
	Industry     [32]uint8
	foundElement bool
}

// CargoFilterHasBeenRecalculated is upstream's: a station element was met.
func (s *CargoSearchState) CargoFilterHasBeenRecalculated() bool { return s.foundElement }

func newCargoSearchState() CargoSearchState {
	var s CargoSearchState
	for i := range s.Industry {
		s.Industry[i] = 0xFF
	}
	return s
}

// getStationElement is Station.cpp's: the first station element on the
// tile at that height — nil when it is AI-allocated (or there is none).
func (m *Map) getStationElement(x, y, z int32) *scenario.StationElement {
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return nil
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindStation || int(r.Index) >= len(t.Stations) {
			continue
		}
		el := &t.Stations[r.Index]
		if int32(el.BaseZ) != z/4 {
			continue
		}
		if el.AiAllocated {
			return nil
		}
		return el
	}
	return nil
}

// setCatchmentRegion is setStationCatchmentRegion: flag_1 on the tiles of
// the rectangle, clamped to the map.
func (m *Map) setCatchmentRegion(cm []uint8, x0, y0, x1, y1 int) {
	for y := max(y0, 0); y <= min(y1, m.Height-1); y++ {
		for x := max(x0, 0); x <= min(x1, m.Width-1); x++ {
			cm[y*m.Width+x] |= 1 << 1
		}
	}
}

// catchment is setCatchmentDisplay(station, flag_1): the station's catchment
// on a fresh map.
func (m *Map) catchment(st *scenario.StationName, om *objects.ObjectManager) []uint8 {
	cm := make([]uint8, m.Width*m.Height)
	for i := 0; i < int(st.TileSize) && i < len(st.Tiles); i++ {
		p := st.Tiles[i]
		x, y, z := int32(p[0]), int32(p[1]), int32(p[2])&^3 // heightFloor
		el := m.getStationElement(x, y, z)
		if el == nil {
			continue
		}
		tx, ty := int(x)>>5, int(y)>>5
		switch el.StationType {
		case 2: // airport
			if om == nil || int(el.ObjectID) >= len(om.AirportObjects) || om.AirportObjects[el.ObjectID] == nil {
				continue
			}
			a := om.AirportObjects[el.ObjectID]
			ax0, ay0 := rotOffset(int16(a.MinX), int16(a.MinY), el.Rotation)
			ax1, ay1 := rotOffset(int16(a.MaxX), int16(a.MaxY), el.Rotation)
			x0, y0, x1, y1 := tx+int(ax0), ty+int(ay0), tx+int(ax1), ty+int(ay1)
			if x0 > x1 {
				x0, x1 = x1, x0
			}
			if y0 > y1 {
				y0, y1 = y1, y0
			}
			m.setCatchmentRegion(cm, x0-catchmentSize, y0-catchmentSize, x1+catchmentSize, y1+catchmentSize)
		case 3: // docks: always 2×2
			m.setCatchmentRegion(cm, tx-catchmentSize, ty-catchmentSize, tx+catchmentSize+1, ty+catchmentSize+1)
		default:
			m.setCatchmentRegion(cm, tx-catchmentSize, ty-catchmentSize, tx+catchmentSize, ty+catchmentSize)
		}
	}
	return cm
}

// CalcAcceptedCargo is Station::calcAcceptedCargo with doCalcAcceptedCargo:
// the cargo types the station accepts (a bit each), the search state with
// its industries and produced types.
func (m *Map) CalcAcceptedCargo(id int, om *objects.ObjectManager) (uint32, CargoSearchState) {
	s := newCargoSearchState()
	st := &m.StationNames[id]
	return m.doCalcAcceptedCargo(st, m.catchment(st, om), om, s)
}

// CalcAcceptedCargoTrainStationGhost is calcAcceptedCargoTrainStationGhost
// (Station.cpp): the catchment of the ghost's station (id -1: none) plus the
// ghost's own (sub_491BF5: 4 tiles round the tile at x, y), filtered by
// filter; the accepted and the produced cargo types.
func (m *Map) CalcAcceptedCargoTrainStationGhost(id int, x, y int32, filter uint32, om *objects.ObjectManager) (accepted, produced uint32) {
	s := newCargoSearchState()
	s.Filter = filter
	var st *scenario.StationName
	cm := make([]uint8, m.Width*m.Height)
	if id >= 0 && id < len(m.StationNames) {
		st = &m.StationNames[id]
		cm = m.catchment(st, om)
	}
	tx, ty := int(x)>>5, int(y)>>5
	m.setCatchmentRegion(cm, tx-catchmentSize, ty-catchmentSize, tx+catchmentSize, ty+catchmentSize)
	accepted, s = m.doCalcAcceptedCargo(st, cm, om, s)
	return accepted, s.Produced
}

// doCalcAcceptedCargo is Station.cpp's: the station's tiles set the filter
// (none, or no station: everything), then the buildings and industries in
// the catchment score and produce.
func (m *Map) doCalcAcceptedCargo(st *scenario.StationName, cm []uint8, om *objects.ObjectManager, s CargoSearchState) (uint32, CargoSearchState) {
	for i := 0; st != nil && i < int(st.TileSize) && i < len(st.Tiles); i++ {
		p := st.Tiles[i]
		el := m.getStationElement(int32(p[0]), int32(p[1]), int32(p[2]))
		if el == nil {
			continue
		}
		s.foundElement = true
		if el.StationType == 1 { // road station: its object's cargo
			if ro := roadStationObject(om, el.ObjectID); ro != nil {
				slot := cargoSlotByName(om, ro.CargoName)
				cargo := uint(slot) & 31
				switch {
				case slot < 0:
				case ro.Flags&roadStationFlagPassenger != 0:
					s.Filter |= 1 << cargo
				case ro.Flags&roadStationFlagFreight != 0:
					s.Filter |= ^(uint32(1) << cargo)
				}
			}
		} else {
			s.Filter = ^uint32(0)
		}
	}
	if s.Filter == 0 {
		s.Filter = ^uint32(0)
	}
	in := func(c uint8) bool { return c != 0xFF && c < 32 && s.Filter&(1<<c) != 0 }
	for ty := 0; ty < m.Height; ty++ {
		for tx := 0; tx < m.Width; tx++ {
			if cm[ty*m.Width+tx]&(1<<1) == 0 {
				continue
			}
			t := m.TileAt(tx, ty)
			if t == nil {
				continue
			}
			for _, r := range t.Order {
				if r.Flags&0x10 != 0 { // ghost
					continue
				}
				switch r.Kind {
				case scenario.KindIndustry:
					if int(r.Index) >= len(t.Industries) {
						break
					}
					ind := m.industryByID(t.Industries[r.Index].IndustryID)
					if ind == nil || ind.UnderConstruction != 0xFF || om == nil {
						break
					}
					obj := om.GetIndustryObjectByIndex(int(ind.ObjectID))
					if obj == nil {
						break
					}
					produced, required := om.IndustryCargoIDs(obj)
					for _, c := range required {
						if in(c) {
							s.Score[c] += 8
							s.Industry[c] = ind.ID
						}
					}
					for _, c := range produced {
						if in(c) {
							s.Produced |= 1 << c
						}
					}
				case scenario.KindBuilding:
					if int(r.Index) >= len(t.Buildings) || om == nil {
						break
					}
					b := &t.Buildings[r.Index]
					if b.IsMisc || !b.IsConstructed {
						break
					}
					obj := om.GetBuildingObjectByIndex(int(b.ObjectID))
					if obj == nil {
						break
					}
					produced, consumed := om.ProducedCargoIDs(obj), om.ConsumedCargoIDs(obj)
					for i, c := range produced {
						if in(c) {
							s.Score[c] += uint32(obj.ProducedCargoQty[i])
							if obj.ProducedQuantity[i] != 0 {
								s.Produced |= 1 << c
							}
						}
					}
					for i, c := range consumed {
						if in(c) {
							s.Score[c] += uint32(obj.ConsumedCargoQty[i])
						}
					}
					if obj.Flags&objects.BuildingFlagLargeTile != 0 { // a large building counts once
						off := buildingTileOffsets(true)[b.SequenceIndex&3]
						x0, y0 := (tx*32-int(off[0]))/32, (ty*32-int(off[1]))/32
						for _, d := range [4][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}} {
							if x, y := x0+d[0], y0+d[1]; x >= 0 && y >= 0 && x < m.Width && y < m.Height {
								cm[y*m.Width+x] &^= 1 << 1
							}
						}
					}
				}
			}
		}
	}
	var accepted uint32
	for c := 0; c < 32; c++ {
		if s.Score[c] >= 8 {
			accepted |= 1 << c
		}
	}
	return accepted, s
}

// RoadStationFlags::passenger / freight.
const (
	roadStationFlagPassenger = 1 << 1
	roadStationFlagFreight   = 1 << 2
)

// roadStationObject is ObjectManager::get<RoadStationObject>(id), or nil.
func roadStationObject(om *objects.ObjectManager, id uint8) *objects.RoadStationObject {
	if om == nil || int(id) >= len(om.RoadStationObjects) {
		return nil
	}
	return om.RoadStationObjects[id]
}
