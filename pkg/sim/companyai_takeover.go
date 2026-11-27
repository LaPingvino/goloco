package sim

import (
	"fmt"
	"log"

	"github.com/LaPingvino/goloco/pkg/objects"
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// Taking over a loaded OpenLoco AI company (AI-INTERFACE.md G10, §4
// question 1, option (a)): its plans are dropped as upstream's AI drops a
// thought it gives up on (aiThinkUndoPartialAction 0x00431104's sub-states
// 2–4: removeAllAiAllocatedCompanyAssetsOnMapByChunk, removeAiAllocatedStations,
// clearThought and aiThinkState unk0), for every thought at once. Sub-states 0
// and 1 (selling the thought's vehicles, sub_48715C's land assets) are left
// out: a takeover keeps what the company really owns. Then every head of the
// company has aiThoughtId 0xFF, as a vehicle bought in play.
// OpenLoco reference: src/OpenLoco/src/World/CompanyAi/CompanyAi.cpp.

// aiStationFlagAiAllocated / aiStationFlagOperational are AiThoughtStationFlags.
const (
	aiStationFlagAiAllocated = 1 << 0
	aiStationFlagOperational = 1 << 1
)

// TakeOverCompanyAI gives company id, an OpenLoco AI company of a loaded
// game, to the registered AI name: its OpenLoco plans are removed from the
// map and its thoughts nulled first (TakeOverCompany), then SetCompanyAI.
func (m *Map) TakeOverCompanyAI(id int, name string, opts AIOptions, om *objects.ObjectManager) error {
	if _, ok := aiRegistry[name]; !ok {
		return fmt.Errorf("no AI named %q", name)
	}
	if id < 0 || id >= companyMaxCount || m.CompanyEmpty(id) || m.IsPlayerCompany(id) {
		return fmt.Errorf("company %d can't have an AI", id)
	}
	m.TakeOverCompany(uint8(id), om)
	return m.SetCompanyAI(id, name, opts)
}

// TakeOverCompany clears company id's OpenLoco AI plans: its AI-allocated
// track and road (removeAiAllocatedCompanyTracksRoadsOnTile on every tile,
// the whole map in one go — upstream does 1500 tiles a think), each thought's
// AI-allocated stations (removeAiAllocatedStations, one per call, until none
// is left), every thought nulled (clearThought), the think state unk0, and
// the company's heads' aiThoughtId 0xFF. Free: every command runs apply |
// noPayment | aiAllocated under the company, as upstream's.
func (m *Map) TakeOverCompany(id uint8, om *objects.ObjectManager) {
	if int(id) >= len(m.Companies) {
		return
	}
	cmd := &Cmd{Company: id, Objects: om, Players: func(o uint8) bool { return m.IsPlayerCompany(int(o)) }}
	var g1 G1Raw
	if om != nil {
		g1, _ = om.G1File.(G1Raw)
	}
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			m.removeAiAllocatedCompanyTracksRoadsOnTile(cmd, g1, x, y)
		}
	}
	c := &m.Companies[id]
	nonAi := 0
	for k := range c.ThoughtType {
		if c.ThoughtType[k] == 0xFF {
			continue
		}
		// removeAllAiAllocatedCompanyAssetsOnMapByChunk's end: the
		// connection flags' bits 1 and 3 cleared.
		for i := range c.ThoughtNumStations[k] {
			if int(i) >= len(c.ThoughtStations[k]) {
				break
			}
			st := &c.ThoughtStations[k][i]
			st.ConnectionFlagsA &^= 1<<1 | 1<<3
			st.ConnectionFlagsB &^= 1<<1 | 1<<3
		}
		for {
			done, n := m.removeAiAllocatedStations(cmd, g1, k)
			nonAi += n
			if done {
				break
			}
		}
		c.ThoughtType[k] = 0xFF // clearThought
	}
	c.AiThinkState = 0 // AiThinkState::unk0 (sub_431186)
	for i := range m.Entities {
		e := &m.Entities[i]
		if e.EntityType == scenario.VehicleTypeHead && e.Owner == id && len(e.Raw) > 0x60 {
			e.Raw[0x60] = 0xFF // aiThoughtId: the save copies Raw; unload.go reads it there
		}
	}
	if nonAi != 0 {
		log.Printf("[AI] takeover of company %d: %d station tiles had real track (removeAiAllocatedTrainStation's other branch)", id, nonAi)
	}
	log.Printf("[AI] company %d taken over: OpenLoco plans removed, thoughts cleared", id)
}

// removeAiAllocatedCompanyTracksRoadsOnTile is CompanyAi.cpp 0x00486224:
// the tile's first AI-allocated road or track of the company goes (a road
// unless an AI-allocated station element stands on it, a track unless it has
// a station element), and the tile is searched again after each removal
// that worked.
func (m *Map) removeAiAllocatedCompanyTracksRoadsOnTile(c *Cmd, g1 G1Raw, x, y int) {
	t := &m.Tiles[y][x]
	wx, wy := int32(x)*32, int32(y)*32
	for recheck := true; recheck; {
		recheck = false
		for _, r := range t.Order {
			if r.Flags&elemFlagAiAllocated == 0 {
				continue
			}
			switch r.Kind {
			case scenario.KindRoad:
				el := t.Roads[r.Index]
				if el.Owner != c.Company {
					continue
				}
				if el.HasStationElement && m.aiAllocatedStationElement(t, el.BaseZ) {
					continue
				}
				a := RoadRemoval{X: wx, Y: wy, Z: int32(el.BaseZ) * smallZStep, Rotation: el.Rotation,
					RoadID: el.RoadID, Index: el.SequenceIndex, RoadObjectID: el.RoadObjectID}
				if _, ok := m.DoCommandFor(c, CmdApply|CmdAiAllocated|CmdNoPayment, ExpenditureConstruction, func(f uint8) (int32, bool) {
					return m.RemoveRoad(c, f, a, g1)
				}); ok {
					recheck = true
				}
			case scenario.KindTrack:
				el := t.Tracks[r.Index]
				if el.Owner != c.Company || el.HasStation {
					continue
				}
				a := TrackRemoval{X: wx, Y: wy, Z: int32(el.BaseZ) * smallZStep, Rotation: el.Rotation,
					TrackID: el.TrackID, Index: el.SequenceIndex, TrackObjectID: el.TrackObjectID}
				if _, ok := m.DoCommandFor(c, CmdApply|CmdAiAllocated|CmdNoPayment, ExpenditureConstruction, func(f uint8) (int32, bool) {
					return m.RemoveTrack(c, f, a, g1)
				}); ok {
					recheck = true
				}
			default:
				continue
			}
			if recheck {
				break
			}
		}
	}
}

// aiAllocatedStationElement is getAiAllocatedStationElement (CompanyAi.cpp
// 0x…5719): the tile's first station element at baseZ is AI-allocated.
func (m *Map) aiAllocatedStationElement(t *Tile, baseZ uint8) bool {
	for _, r := range t.Order {
		if r.Kind != scenario.KindStation || t.Stations[r.Index].BaseZ != baseZ {
			continue
		}
		return t.Stations[r.Index].AiAllocated
	}
	return false
}

// removeAiAllocatedStations is CompanyAi.cpp 0x00483602 for thought k: the
// first station of the thought that is AI-allocated and not operational is
// removed (an airport, a port, a road or a train station by the thought's
// type and trackObjId) and loses its aiAllocated flag ("Ai assumes removal
// was a success!"); done (upstream's 1) when there is none. n counts the
// station tiles removeAiAllocatedTrainStation found real track on.
func (m *Map) removeAiAllocatedStations(c *Cmd, g1 G1Raw, k int) (done bool, n int) {
	co := &m.Companies[c.Company]
	typ := co.ThoughtType[k]
	for i := range co.ThoughtNumStations[k] {
		if int(i) >= len(co.ThoughtStations[k]) {
			break
		}
		st := &co.ThoughtStations[k][i]
		if st.Flags&aiStationFlagOperational != 0 || st.Flags&aiStationFlagAiAllocated == 0 {
			continue
		}
		x, y, z := int32(st.X), int32(st.Y), int32(st.BaseZ)*smallZStep
		const flags = CmdApply | CmdAiAllocated | CmdNoPayment
		switch {
		case thoughtTypeHasFlags(typ, thoughtTypeFlagAirBased):
			m.DoCommandFor(c, flags, ExpenditureConstruction, func(f uint8) (int32, bool) {
				return m.RemoveAirport(c, f, AirportRemoval{X: x, Y: y, Z: z}, g1)
			})
		case thoughtTypeHasFlags(typ, thoughtTypeFlagWaterBased):
			m.DoCommandFor(c, flags, ExpenditureConstruction, func(f uint8) (int32, bool) {
				return m.RemovePort(c, f, PortRemoval{X: x, Y: y, Z: z}, g1)
			})
		case co.ThoughtTrackObjID[k]&(1<<7) != 0:
			m.removeAiAllocatedRoadStation(c, g1, x, y, z, st.Rotation, co.ThoughtTrackObjID[k]&^(1<<7))
		default:
			n = m.removeAiAllocatedTrainStation(c, g1, x, y, z, st.Rotation, co.ThoughtTrackObjID[k], co.ThoughtStationLength[k])
		}
		st.Flags &^= aiStationFlagAiAllocated
		return false, n
	}
	return true, 0
}

// removeAiAllocatedRoadStation is CompanyAi.cpp 0x0047B107: the road
// station removed (refused: once more facing the other way), then the
// tile's first road at that height, rotation and object with road piece 0,
// if it is the company's and AI-allocated.
func (m *Map) removeAiAllocatedRoadStation(c *Cmd, g1 G1Raw, x, y, z int32, rotation, objectID uint8) {
	const flags = CmdApply | CmdAiAllocated | CmdNoPayment
	a := RoadStationRemoval{X: x, Y: y, Z: z, Rotation: rotation, RoadObjectID: objectID}
	if _, ok := m.DoCommandFor(c, flags, ExpenditureConstruction, func(f uint8) (int32, bool) {
		return m.RemoveRoadStation(c, f, a, g1)
	}); !ok {
		a.Rotation ^= 1 << 1
		m.DoCommandFor(c, flags, ExpenditureConstruction, func(f uint8) (int32, bool) {
			return m.RemoveRoadStation(c, f, a, g1)
		})
	}
	t := m.TileAt(int(x)>>5, int(y)>>5)
	if t == nil {
		return
	}
	for _, r := range t.Order {
		if r.Kind != scenario.KindRoad {
			continue
		}
		el := t.Roads[r.Index]
		if int32(el.BaseZ)*smallZStep != z || el.Rotation != rotation || el.RoadObjectID != objectID || el.RoadID != 0 || el.Owner != c.Company {
			continue
		}
		if el.AiAllocated {
			ra := RoadRemoval{X: x, Y: y, Z: z, Rotation: rotation, RoadObjectID: objectID}
			m.DoCommandFor(c, flags, ExpenditureConstruction, func(f uint8) (int32, bool) {
				return m.RemoveRoad(c, f, ra, g1)
			})
		}
		break
	}
}

// removeAiAllocatedTrainStation is CompanyAi.cpp 0x004A7CD2: along the
// station's length from pos, each tile's first track at that height,
// rotation and object, piece 0, sequence 0, of the company, is removed —
// AI-allocated under aiAllocated, and a real one ("Really? Seems a bit odd")
// without it. Returns how many real ones it found.
func (m *Map) removeAiAllocatedTrainStation(c *Cmd, g1 G1Raw, x, y, z int32, rotation, objectID, length uint8) int {
	real := 0
	for range length {
		nonAi := false
		if t := m.TileAt(int(x)>>5, int(y)>>5); t != nil {
			for _, r := range t.Order {
				if r.Kind != scenario.KindTrack {
					continue
				}
				el := t.Tracks[r.Index]
				if int32(el.BaseZ)*smallZStep != z || el.Rotation != rotation || el.TrackObjectID != objectID ||
					el.SequenceIndex != 0 || el.TrackID != 0 || el.Owner != c.Company {
					continue
				}
				a := TrackRemoval{X: x, Y: y, Z: z, Rotation: rotation, TrackObjectID: objectID}
				if el.AiAllocated {
					m.DoCommandFor(c, CmdApply|CmdAiAllocated|CmdNoPayment, ExpenditureConstruction, func(f uint8) (int32, bool) {
						return m.RemoveTrack(c, f, a, g1)
					})
				} else {
					nonAi = true
				}
				break
			}
		}
		if nonAi {
			real++
			a := TrackRemoval{X: x, Y: y, Z: z, Rotation: rotation, TrackObjectID: objectID}
			m.DoCommandFor(c, CmdApply|CmdNoPayment, ExpenditureConstruction, func(f uint8) (int32, bool) {
				return m.RemoveTrack(c, f, a, g1)
			})
		}
		x += int32(kRotationOffset[rotation][0])
		y += int32(kRotationOffset[rotation][1])
	}
	return real
}
