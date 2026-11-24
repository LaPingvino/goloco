package sim

import (
	"github.com/LaPingvino/goloco/pkg/scenario"
)

// s5export.go — the map back into a scenario.Scenario for the S5 writer
// (scenario.EncodeGameState / EncodeTileElements): the inverse of
// LoadScenario, field for field, so a save loads back into the same map.
//
// OpenLoco reference: src/OpenLoco/src/S5/S5GameState.cpp exportGameState

const surfaceFlagWave = 1 << 6 // SurfaceElement flag_6: a wave on the tile

// surfaceFlags is the flags byte of a tile's surface element.
func surfaceFlags(order []scenario.ElemRef) uint8 {
	for _, r := range order {
		if r.Kind == scenario.KindSurface {
			return r.Flags
		}
	}
	return 0
}

// LoadWaves is importGameState's waves: the table as saved (an empty slot's
// x Location::null).
func (s *State) LoadWaves(w [scenario.MaxWaves]scenario.Wave) {
	for i := range s.Waves {
		s.Waves[i] = Wave{X: int(w[i].X), Y: int(w[i].Y), Frame: w[i].Frame}
	}
}

// ExportWaves is exportGameState's waves.
func (s *State) ExportWaves() (w [scenario.MaxWaves]scenario.Wave) {
	for i, v := range s.Waves {
		w[i] = scenario.Wave{X: int16(v.X), Y: int16(v.Y), Frame: v.Frame}
	}
	return w
}

// Effect sprite extents upstream's create gives a money effect
// (MoneyEffect.cpp), which goloco's MoneyEffect doesn't keep.
const moneySpriteWidth, moneySpriteHeightNeg, moneySpriteHeightPos = 64, 20, 30

// loadEffects takes a saved game's effect entities: exhausts, crash
// particles, explosion clouds and splashes into Effects, money into
// MoneyEffects, the rest (fireball, explosion smoke) kept as read; smoke too
// goes into Effects.
func (m *Map) loadEffects(misc []scenario.MiscEntity) {
	m.Effects, m.MoneyEffects, m.OtherEffects = nil, nil, nil
	for _, e := range misc {
		x, y, z := int32(e.X), int32(e.Y), int32(e.Z)
		switch e.SubType {
		case scenario.EffectRedGreenCurrency, scenario.EffectWindowCurrency:
			if m.MoneyEffects == nil {
				m.MoneyEffects = map[uint16]*MoneyEffect{}
			}
			me := &MoneyEffect{X: x, Y: y, Z: z, Amount: e.Amount, Company: e.Var2E, SubType: e.SubType,
				NumMovements: e.Field28, Wiggle: uint8(e.Wiggle), Raw: e.Raw}
			if e.SubType == scenario.EffectRedGreenCurrency {
				me.MoveDelay = uint8(e.Field26)
			} else {
				me.Frame = e.Field26
			}
			m.MoneyEffects[e.ID] = me
		case scenario.EffectExhaust, scenario.EffectVehicleCrashParticle, scenario.EffectExplosionCloud, scenario.EffectSplash, scenario.EffectSmoke:
			if m.Effects == nil {
				m.Effects = map[uint16]*Effect{}
			}
			ef := &Effect{X: x, Y: y, Z: z, SubType: e.SubType, Width: e.SpriteWidth, HeightNeg: e.SpriteHeightNeg,
				HeightPos: e.SpriteHeightPos, Frame: e.Field28, Raw: e.Raw}
			switch e.SubType {
			case scenario.EffectExhaust:
				ef.Frame, ef.Stationary, ef.Wind, ef.ObjectID = e.Field26, uint8(e.Field28), e.WindProgress, e.ObjectID
			case scenario.EffectVehicleCrashParticle:
				ef.TimeToLive, ef.Colour, ef.SpriteBase = e.Field26, [2]uint8{e.Var2E, e.Colour2}, uint8(e.CrashedSpriteBase)
				ef.VelX, ef.VelY, ef.VelZ = e.Velocity[0], e.Velocity[1], e.Velocity[2]
				ef.AccX, ef.AccY, ef.AccZ = e.Acceleration[0], e.Acceleration[1], e.Acceleration[2]
			}
			m.Effects[e.ID] = ef
		default:
			if m.OtherEffects == nil {
				m.OtherEffects = map[uint16]scenario.MiscEntity{}
			}
			m.OtherEffects[e.ID] = e
		}
	}
}

// exportEffects is loadEffects' inverse.
func (m *Map) exportEffects() []scenario.MiscEntity {
	var out []scenario.MiscEntity
	for id, me := range m.MoneyEffects {
		e := scenario.MiscEntity{ID: id, SubType: me.SubType, X: int16(me.X), Y: int16(me.Y), Z: int16(me.Z),
			SpriteWidth: moneySpriteWidth, SpriteHeightNeg: moneySpriteHeightNeg, SpriteHeightPos: moneySpriteHeightPos,
			Field26: me.Frame, Field28: me.NumMovements, Amount: me.Amount, Var2E: me.Company, Wiggle: uint16(me.Wiggle), Raw: me.Raw}
		if me.SubType == scenario.EffectRedGreenCurrency {
			e.Field26 = uint16(me.MoveDelay)
		}
		out = append(out, e)
	}
	for id, ef := range m.Effects {
		e := scenario.MiscEntity{ID: id, SubType: ef.SubType, X: int16(ef.X), Y: int16(ef.Y), Z: int16(ef.Z),
			SpriteWidth: ef.Width, SpriteHeightNeg: ef.HeightNeg, SpriteHeightPos: ef.HeightPos, Field28: ef.Frame, Raw: ef.Raw}
		switch ef.SubType {
		case scenario.EffectExhaust:
			e.Field26, e.Field28, e.WindProgress, e.ObjectID = ef.Frame, uint16(ef.Stationary), ef.Wind, ef.ObjectID
		case scenario.EffectVehicleCrashParticle:
			e.Field26, e.Var2E, e.Colour2, e.CrashedSpriteBase = ef.TimeToLive, ef.Colour[0], ef.Colour[1], uint16(ef.SpriteBase)
			e.Velocity = [3]int16{ef.VelX, ef.VelY, ef.VelZ}
			e.Acceleration = [3]int32{ef.AccX, ef.AccY, ef.AccZ}
		}
		out = append(out, e)
	}
	for _, e := range m.OtherEffects {
		out = append(out, e)
	}
	return out
}

// exportTile is a tile as scenario.Tile: its surface (clearZ: SurfaceClearZ,
// as the map keeps it) and elements, the surface's wave flag in its flags.
func (t *Tile) exportTile() scenario.Tile {
	order := append([]scenario.ElemRef(nil), t.Order...)
	for i := range order {
		if order[i].Kind == scenario.KindSurface {
			order[i].Flags = order[i].Flags&^surfaceFlagWave | boolU8(t.WaveFlag)*surfaceFlagWave
		}
	}
	return scenario.Tile{Height: t.BaseZ, ClearZ: t.ClearZ, Slope: t.Slope, Water: t.WaterLevel,
		UpdateTimer: t.UpdateTimer, TerrainIndex: t.TerrainIndex, GrowthStage: t.GrowthStage, Variation: t.Variation,
		IsIndustrial: t.IsIndustrial, Type6Flag: t.Type6Flag,
		Trees: t.Trees, Buildings: t.Buildings, Walls: t.Walls, Tracks: t.Tracks, Roads: t.Roads,
		Stations: t.Stations, Signals: t.Signals, Industries: t.Industries, Order: order}
}

func boolU8(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}

// ExportScenario is the map as a scenario.Scenario: LoadScenario's inverse
// (the file it was loaded from as the base, Raw).
func (s *State) ExportScenario() *scenario.Scenario {
	m := &s.Map
	sc := &scenario.Scenario{Raw: m.Raw, MapWidth: m.Width, MapHeight: m.Height}
	sc.Tiles = make([][]scenario.Tile, m.Height)
	for y := range sc.Tiles {
		sc.Tiles[y] = make([]scenario.Tile, m.Width)
		for x := range sc.Tiles[y] {
			sc.Tiles[y][x] = m.Tiles[y][x].exportTile()
		}
	}
	for _, msg := range m.Messages {
		sc.Messages = append(sc.Messages, scenario.MessageRecord{Type: uint8(msg.Type), Text: []byte(msg.Text), CompanyID: msg.CompanyID,
			TimeActive: msg.TimeActive, Subjects: msg.Subjects, Date: msg.Date})
	}
	sc.ActiveMessageIndex = m.ActiveMessageIndex
	sc.Towns = m.Towns
	sc.Entities = m.Entities
	if l := m.EntityLists; l != nil {
		sc.EntityLists = &scenario.EntityLists{Heads: l.Heads, Counts: l.Counts, Links: l.Links}
	}
	sc.MiscEntities = m.exportEffects()
	sc.Orders, sc.OrderTableLength, sc.Routings = m.Orders, m.OrderTableLength, m.Routings
	if len(sc.Routings) == 0 { // none allocated yet: the table as resetRoutingTable leaves it
		sc.Routings = make([]uint16, scenario.MaxRoutingVehicles*scenario.RoutingsPerVehicle)
		for i := range sc.Routings {
			sc.Routings[i] = RoutingNull
		}
	}
	sc.RoadAnyTypeCompatible, sc.RoadUsableByAll = m.RoadAnyTypeCompatible, m.RoadUsableByAll
	sc.Rng = [2]uint32{m.Rng1.Srand0(), m.Rng1.Srand1()}
	sc.UnkRng = [2]uint32{m.Rng2.Srand0(), m.Rng2.Srand1()}
	sc.ScenarioTicks = m.ScenarioTicks
	sc.TileUpdateStart = [2]int16{int16(m.tileUpdateX << 5), int16(m.tileUpdateY << 5)}
	sc.Companies, sc.Industries = m.Companies, m.Industries
	sc.CurrentDay = m.CurrentDay
	sc.CompanyNames, sc.CompanyFaces = m.CompanyNames, m.CompanyFaces
	sc.ControllingCompany = m.ControllingCompany
	sc.TrafficHandedness = m.TrafficHandedness
	sc.LastRailroadOption, sc.LastRoadOption, sc.LastTrackTypeOption = m.LastRailroadOption, m.LastRoadOption, m.LastTrackTypeOption
	sc.LastBuildVehiclesOption, sc.LastVehicleType = m.LastBuildVehiclesOption, m.LastVehicleType
	sc.LastAirport, sc.LastShipPort = m.LastAirport, m.LastShipPort
	sc.Construction = m.Construction
	sc.LastMapWindow = m.LastMapWindow
	sc.PickupDirection = m.PickupDirection
	sc.UnlockedVehicles = m.UnlockedVehicles
	sc.Animations, sc.AnimationSlots = m.Animations, m.AnimationSlots
	sc.Waves = s.ExportWaves()
	sc.CurrentYear, sc.CurrentMonth, sc.CurrentDayOfMonth = m.CurrentYear, m.CurrentMonth, m.CurrentDayOfMonth
	sc.GameStateFlags = m.GameStateFlags
	sc.IndustryManagerFlags = m.IndustryManagerFlags
	sc.CompanyRecords = m.CompanyRecords
	sc.NumberOfIndustries = m.NumberOfIndustries
	sc.DayCounter = m.DayCounter
	sc.PlayerCompanies, sc.LoanInterestRate, sc.MaxLoanSize = m.PlayerCompanies, m.LoanInterestRate, m.MaxLoanSize
	sc.DefaultLevelCrossingType = m.DefaultLevelCrossingType
	sc.StartingLoanSize = m.StartingLoanSize
	sc.SeaLevel, sc.LastLandOption = m.SeaLevel, m.LastLandOption
	sc.DefaultTreeObjectId, sc.DefaultWallObjectId = m.DefaultTreeObjectId, m.DefaultWallObjectId
	sc.CurrentSnowLine, sc.CurrentSeason = m.CurrentSnowLine, m.CurrentSeason
	sc.PreferredAI = m.PreferredAI
	sc.MaxCompetingCompanies, sc.CompetitorStartDelay, sc.ProduceAICompanyTimeout = m.MaxCompetingCompanies, m.CompetitorStartDelay, m.ProduceAICompanyTimeout
	sc.GolocoAI = m.EncodeAIChunk()
	sc.ForbiddenVehiclesPlayers, sc.ForbiddenVehiclesCompetitors = m.ForbiddenVehiclesPlayers, m.ForbiddenVehiclesCompetitors
	sc.Objective = m.Objective
	sc.ObjectiveTimeLimitUntilYear, sc.ObjectiveMonthsInChallenge = m.ObjectiveUntilYear, m.ObjectiveMonthsInChallenge
	sc.ObjectiveCompletedInMonths = m.ObjectiveCompletedInMonths
	sc.CurrencyMultiplicationFactor = m.CurrencyMultiplicationFactor
	sc.CompanyColourSchemes = m.CompanyColourSchemes
	sc.UserStrings = m.UserStrings
	sc.StationNames = m.StationNames
	sc.ScenarioFileName, sc.ScenarioName, sc.ScenarioDetails = m.ScenarioFileName, m.ScenarioName, m.ScenarioDetails
	return sc
}

// PrepareSave is what exportGameStateToFile does to the live game before it
// writes it: StationManager::zeroUnused (an empty station's record reset —
// name null, every cargo's origin null — and every station's tiles past its
// count zeroed), OrderManager::zeroUnusedOrderTable and TileManager::
// reorganise (entriesEnd the element count). (EntityManager::
// resetSpatialIndex and zeroUnused touch nothing goloco keeps: the entity
// slots it doesn't model are written zeroed.)
func (m *Map) PrepareSave() {
	m.EntriesEnd = m.countElements() // TileManager::reorganise: the entries packed
	for i := range m.StationNames {
		s := &m.StationNames[i]
		if s.Name == stationNameNull {
			*s = scenario.StationName{Name: stationNameNull}
			for c := range s.Cargo {
				s.Cargo[c].Origin = stationIDNull
			}
			continue
		}
		for k := int(s.TileSize); k < len(s.Tiles); k++ {
			s.Tiles[k] = [3]int16{}
		}
	}
	if int(m.OrderTableLength) < len(m.Orders) {
		clear(m.Orders[m.OrderTableLength:])
	}
}

// resetSurfaceClearance is TileManager::resetSurfaceClearance (every load,
// S5::importSaveToGameState): a flat surface's clearZ its baseZ.
func (m *Map) resetSurfaceClearance() {
	for y := range m.Tiles {
		for x := range m.Tiles[y] {
			if t := &m.Tiles[y][x]; t.Slope&0x1F == 0 {
				t.ClearZ = t.BaseZ
			}
		}
	}
}
