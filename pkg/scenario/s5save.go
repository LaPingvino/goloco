package scenario

import (
	"encoding/binary"
	"io"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// s5save.go — the S5 writer (S5.cpp exportGameStateToFile / exportGameState):
// the game state is the loaded file's own (Raw.GameState) with every field
// goloco models written over it — the inverse of each parse* in loader.go,
// s5entities.go and vehiclelist.go, at the same offsets — and the upstream
// save's normalisations (EntityManager::zeroUnused, StationManager::zeroUnused,
// OrderManager::zeroUnusedOrderTable, the entities' sprite bounds null).
//
// OpenLoco reference: src/OpenLoco/src/S5/S5.cpp, S5GameState.cpp,
// S5Entity.cpp, S5Company.cpp, S5Town.cpp, S5Industry.cpp, S5Station.cpp

// kCurrentVersion and kMagicNumber (S5.cpp, S5.h).
const (
	S5Version = 0x62262
	S5Magic   = 0x62300 // S5::kMagicNumber
)

// S5File is what exportGameState writes, chunk by chunk.
type S5File struct {
	Header          assets.S5Header
	Options         []byte // scenario / landscape only
	SaveDetails     []byte // with HeaderFlags::hasSaveDetails
	PackedObjects   []PackedObjectRaw
	RequiredObjects []byte
	GameState       []byte // GameStateSize bytes
	TileElements    []byte
}

// Write is exportGameState: the header, options or save details, packed
// objects (writePackedObjects: the raw header, then the data chunk), the
// object list, the game state (a scenario's in its three chunks), the tile
// elements, the checksum.
func (f *S5File) Write(w io.Writer) error {
	s := assets.NewS5ChunkWriter(w)
	s.WriteChunk(assets.EncodingRotate, f.Header.Bytes())
	isScenario := f.Header.Type == assets.S5TypeScenario || f.Header.Type == assets.S5TypeLandscape
	if isScenario {
		s.WriteChunk(assets.EncodingRotate, f.Options)
	}
	if f.Header.Flags&assets.HeaderFlagHasSaveDetails != 0 {
		s.WriteChunk(assets.EncodingRotate, f.SaveDetails)
	}
	for _, po := range f.PackedObjects {
		s.Write(po.HeaderBytes[:])
		s.WriteChunk(assets.EncodingRunLengthSingle, po.Data) // ObjectManager::writePackedObjects
	}
	s.WriteChunk(assets.EncodingRotate, f.RequiredObjects)
	if f.Header.Type == assets.S5TypeScenario {
		s.WriteChunk(assets.EncodingRunLengthSingle, f.GameState[:generalStateSize])
		s.WriteChunk(assets.EncodingRunLengthSingle, f.GameState[gsTownsOffset:gsTownsOffset+gsTownsChunkSize])
		s.WriteChunk(assets.EncodingRunLengthSingle, f.GameState[gsAnimationsOffset:gsAnimationsOffset+gsAnimationsSize])
	} else {
		s.WriteChunk(assets.EncodingRunLengthSingle, f.GameState)
	}
	s.WriteChunk(assets.EncodingRunLengthMulti, f.TileElements)
	return s.Close()
}

// SavedGameHeader is prepareHeader for a saved game: type savedGame, the
// current version and magic, hasSaveDetails.
func SavedGameHeader(numPacked int) assets.S5Header {
	return assets.S5Header{Type: assets.S5TypeSavedGame, Flags: assets.HeaderFlagHasSaveDetails,
		NumPackedObjects: uint16(numPacked), Version: S5Version, Magic: S5Magic}
}

// little-endian writers
func put16(b []byte, o int, v uint16) { binary.LittleEndian.PutUint16(b[o:], v) }
func put32(b []byte, o int, v uint32) { binary.LittleEndian.PutUint32(b[o:], v) }
func putCurrency48(b []byte, o int, v int64) {
	put32(b, o, uint32(v))
	put16(b, o+4, uint16(v>>32))
}

// putCString writes s (Latin-1, as parse reads it) and its terminator into a
// field of n bytes, the rest left as it was.
func putCString(b []byte, s string, n int) {
	i := 0
	for _, r := range s {
		if i >= n-1 {
			break
		}
		b[i] = byte(r)
		i++
	}
	b[i] = 0
}

// EncodeGameState is the save's S5::GameState: Raw.GameState (zeros without
// one) with every modelled field of sc written over it.
func (sc *Scenario) EncodeGameState() []byte {
	gs := make([]byte, GameStateSize)
	if sc.Raw != nil && len(sc.Raw.GameState) == GameStateSize {
		copy(gs, sc.Raw.GameState)
	}
	setFixFlag1(gs)
	sc.putGeneralState(gs)
	sc.putCompanies(gs)
	sc.putTowns(gs)
	sc.putIndustries(gs)
	sc.putStations(gs)
	sc.putEntities(gs)
	sc.putAnimationsAndWaves(gs)
	sc.putUserStrings(gs)
	sc.putOrdersAndRoutings(gs)
	return gs
}

func (sc *Scenario) putGeneralState(gs []byte) {
	put32(gs, 0, sc.Rng[0])
	put32(gs, 4, sc.Rng[1])
	put32(gs, 8, sc.UnkRng[0])
	put32(gs, 12, sc.UnkRng[1])
	put32(gs, GeneralFlagsOffset, sc.GameStateFlags)
	put32(gs, gameStateCurrentDayOffset, sc.CurrentDay)
	put16(gs, 0x18, sc.DayCounter)
	put16(gs, GeneralCurrentYearOffset, sc.CurrentYear)
	gs[0x1C], gs[0x1D] = sc.CurrentMonth, sc.CurrentDayOfMonth
	put16(gs, 0x1E, uint16(sc.SavedViewX))
	put16(gs, 0x20, uint16(sc.SavedViewY))
	gs[0x22], gs[0x23] = sc.SavedViewZoom, sc.SavedViewRotation
	gs[playerCompaniesOffset], gs[playerCompaniesOffset+1] = sc.PlayerCompanies[0], sc.PlayerCompanies[1]
	gs[playerCompaniesOffset] = sc.ControllingCompany
	for i, f := range sc.CurrencyMultiplicationFactor {
		put32(gs, 0x46+4*i, f)
	}
	put32(gs, 0x146, sc.ScenarioTicks)
	put16(gs, 0x156, uint16(sc.TileUpdateStart[0]))
	put16(gs, 0x158, uint16(sc.TileUpdateStart[1]))
	gs[0x192], gs[0x193], gs[0x1AD] = sc.LastRailroadOption, sc.LastRoadOption, sc.LastTrackTypeOption
	gs[0x194], gs[0x195] = sc.LastAirport, sc.LastShipPort
	gs[0x196], gs[0x197], gs[0x198] = sc.TrafficHandedness, sc.LastVehicleType, sc.PickupDirection
	put16(gs, 0x19A, uint16(sc.SeaLevel))
	gs[0x19C], gs[0x19D], gs[0x19E] = sc.CurrentSnowLine, sc.CurrentSeason, sc.LastLandOption
	put32(gs, 0x1A0, sc.OrderTableLength)
	put32(gs, 0x1A4, sc.RoadAnyTypeCompatible)
	put32(gs, 0x1A8, sc.RoadUsableByAll)
	gs[0x1AC], gs[0x1AE] = sc.DefaultLevelCrossingType, sc.LoanInterestRate
	putCString(gs[0x1BC:], sc.ScenarioFileName, 256)
	putCString(gs[0x2BC:], sc.ScenarioName, 64)
	putCString(gs[0x2FC:], sc.ScenarioDetails, 256)
	gs[0x3FD], gs[0x3FE], gs[0x3FF] = sc.PreferredAI[0], sc.PreferredAI[1], sc.PreferredAI[2]
	put16(gs, 0x400, sc.StartingLoanSize)
	put16(gs, 0x402, sc.MaxLoanSize)
	gs[0x414], gs[0x415] = sc.LastBuildVehiclesOption, sc.NumberOfIndustries
	o := sc.Objective
	gs[0x418], gs[0x419] = o.Type, o.Flags
	put32(gs, 0x41A, o.CompanyValue)
	put32(gs, 0x41E, o.MonthlyVehicleProfit)
	gs[0x422], gs[0x423] = o.PerformanceIndex, o.DeliveredCargoType
	put32(gs, 0x424, o.DeliveredCargoAmount)
	gs[0x428] = o.TimeLimitYears
	put16(gs, 0x429, sc.ObjectiveTimeLimitUntilYear)
	put16(gs, 0x42B, sc.ObjectiveMonthsInChallenge)
	put16(gs, 0x42D, sc.ObjectiveCompletedInMonths)
	gs[0x42F] = sc.IndustryManagerFlags
	put16(gs, 0x430, sc.ForbiddenVehiclesPlayers)
	put16(gs, 0x432, sc.ForbiddenVehiclesCompetitors)
	for i := range 3 {
		put16(gs, 0x436+2*i, sc.CompanyRecords.Speed[i])
		gs[0x43C+i] = sc.CompanyRecords.Company[i]
		put32(gs, 0x440+4*i, sc.CompanyRecords.Date[i])
	}
	// The messages: numMessages, activeMessageIndex, the records.
	const base, size, maxMessages = 0x13BA, 0xD4, 200
	n := min(len(sc.Messages), maxMessages)
	put16(gs, 0x13B6, uint16(n))
	put16(gs, 0x13B8, sc.ActiveMessageIndex)
	for i, m := range sc.Messages[:n] {
		r := gs[base+i*size : base+(i+1)*size]
		r[0] = m.Type
		t := m.Text
		if len(t) > 197 {
			t = t[:197]
		}
		copy(r[1:], t)
		r[1+len(t)] = 0
		r[0xC7] = m.CompanyID
		put16(r, 0xC8, m.TimeActive)
		for k, s := range m.Subjects {
			put16(r, 0xCA+2*k, s)
		}
		put32(r, 0xD0, m.Date)
	}
}

// Company fields createCompany sets that goloco doesn't model: written when
// a company slot that was empty in the loaded file has one now.
func initCreatedCompany(r []byte) {
	const aiThoughts, aiThoughtSize = 0x4A8, 0x8C
	for k := range 60 {
		r[aiThoughts+k*aiThoughtSize] = 0xFF // AiThoughtType::null
	}
	r[0x25BE] = 0xFF         // aiSavedThoughtType
	put16(r, 0x8BBC, 0xFFFF) // observationEntity
	put16(r, 0x8BBE, 0xFFFF) // observationX −1
	put16(r, 0x8BC0, 0xFFFF) // observationY −1
	put16(r, 0x8BC2, 0xFFFF) // observationObject
}

func (sc *Scenario) putCompanies(gs []byte) {
	if len(sc.CompanyNames) < companyMaxCount {
		return // no companies modelled (a scenario not started)
	}
	var baseNames []uint16
	if sc.Raw != nil && len(sc.Raw.GameState) == GameStateSize {
		baseNames = parseCompanyNames(sc.Raw.GameState)
	}
	for i := range companyMaxCount {
		r := gs[companiesGameStateOffset+i*companyRecordSize : companiesGameStateOffset+(i+1)*companyRecordSize]
		put16(r, 0x00, sc.CompanyNames[i])
		if sc.CompanyNames[i] != 0 && (baseNames == nil || baseNames[i] == 0) {
			initCreatedCompany(r)
		}
		if i < len(sc.CompanyFaces) {
			f := sc.CompanyFaces[i]
			put16(r, 0x02, f.OwnerName)
			put32(r, 0x04, f.ChallengeFlags)
			put16(r, 0x16, uint16(f.Performance))
			r[0x18], r[0x19] = f.Competitor, f.Emotion
		}
		if i < len(sc.CompanyColourSchemes) {
			c := sc.CompanyColourSchemes[i]
			r[0x1A], r[0x1B] = c.Main[0], c.Main[1]
			for k, v := range c.Vehicle {
				r[0x1C+2*k], r[0x1C+2*k+1] = v[0], v[1]
			}
			put32(r, 0x30, c.CustomSet)
		}
		if i < len(sc.UnlockedVehicles) {
			for w, v := range sc.UnlockedVehicles[i] {
				put32(r, 0x34+4*w, v)
			}
		}
		if i >= len(sc.Companies) {
			continue
		}
		c := &sc.Companies[i]
		putCurrency48(r, 0x08, c.Cash)
		put32(r, 0x0E, uint32(c.CurrentLoan))
		put16(r, 0x50, c.AvailableVehicles)
		r[0x57] = c.NumExpenditureYears
		for mo := range c.Expenditures {
			for t, v := range c.Expenditures[mo] {
				put32(r, 0x58+4*(mo*17+t), uint32(v))
			}
		}
		put32(r, 0x498, c.StartedDate)
		put32(r, 0x49C, c.Var49C)
		put32(r, 0x4A0, c.Var4A0)
		r[0x2579] = c.HeadquartersZ
		put16(r, 0x257A, uint16(c.HeadquartersX))
		put16(r, 0x257C, uint16(c.HeadquartersY))
		r[0x25BF] = c.CurrentRating
		put32(r, 0x85F8, c.CargoUnitsTotalDelivered)
		for k := range 120 {
			put32(r, 0x85FC+4*k, c.CargoUnitsDeliveredHist[k])
			put16(r, 0x87DC+2*k, uint16(c.PerformanceIndexHist[k]))
			putCurrency48(r, 0x88CE+6*k, c.CompanyValueHistory[k])
			put32(r, 0x8C54+4*k, c.CargoUnitsDistanceHist[k])
		}
		put16(r, 0x88CC, c.HistorySize)
		putCurrency48(r, 0x8B9E, c.VehicleProfit)
		for k, v := range c.TransportTypeCount {
			put16(r, 0x8BA4+2*k, v)
		}
		for k, v := range c.CargoDelivered {
			put32(r, 0x8BCE+4*k, v)
		}
		r[0x8C4E], r[0x8C4F] = c.ChallengeProgress, c.NumMonthsInTheRed
		put32(r, 0x8C50, c.CargoUnitsTotalDistance)
		put16(r, 0x8E34, c.JailStatus)
	}
}

func putLabelFrame(b []byte, f LabelFrame) {
	for z := range 4 {
		put16(b, 2*z, uint16(f.Left[z]))
		put16(b, 8+2*z, uint16(f.Right[z]))
		put16(b, 16+2*z, uint16(f.Top[z]))
		put16(b, 24+2*z, uint16(f.Bottom[z]))
	}
}

func (sc *Scenario) putTowns(gs []byte) {
	if sc.Towns == nil && sc.Raw == nil {
		return
	}
	present := make([]bool, townMaxCount)
	for _, tw := range sc.Towns {
		if int(tw.ID) >= townMaxCount {
			continue
		}
		present[tw.ID] = true
		r := gs[townsGameStateOffset+int(tw.ID)*townRecordSize:]
		put16(r, townNameOffset, tw.Name)
		put16(r, townXOffset, uint16(tw.X))
		put16(r, townYOffset, uint16(tw.Y))
		put16(r, 0x06, tw.Flags)
		putLabelFrame(r[0x08:], tw.LabelFrame)
		put32(r, 0x28, tw.Prng[0])
		put32(r, 0x2C, tw.Prng[1])
		put32(r, townPopulationOffset, tw.Population)
		put32(r, 0x34, tw.PopulationCapacity)
		put16(r, townNumBuildingsOffset, uint16(tw.NumBuildings))
		for c, v := range tw.CompanyRatings {
			put16(r, 0x3A+2*c, uint16(v))
		}
		put16(r, 0x58, tw.CompaniesWithRating)
		r[townSizeOffset], r[0x5B] = tw.Size, tw.HistorySize
		copy(r[0x5C:0x5C+240], tw.History[:])
		put32(r, 0x14C, uint32(tw.HistoryMinPopulation))
		copy(r[0x150:0x158], tw.AmenityCounts[:])
		for c, v := range tw.MonthlyCargoDelivered {
			put16(r, 0x158+2*c, v)
		}
		put32(r, 0x198, tw.CargoInfluenceFlags)
		for k := range 4 {
			put16(r, 0x19C+2*k, tw.Var19C[k/2][k%2])
		}
		r[0x1A4], r[0x1A5] = tw.BuildSpeed, tw.NumberOfAirports
		put16(r, townNumStationsOffset, uint16(tw.NumStations))
		put32(r, 0x1A8, tw.Var1A8)
	}
	for i, p := range present {
		if !p {
			put16(gs, townsGameStateOffset+i*townRecordSize+townNameOffset, 0xFFFF) // Town::empty
		}
	}
}

func (sc *Scenario) putIndustries(gs []byte) {
	if sc.Industries == nil && sc.Raw == nil {
		return
	}
	present := make([]bool, industryMaxCount)
	for _, ind := range sc.Industries {
		present[ind.ID] = true
		r := gs[industriesGameStateOffset+int(ind.ID)*industryRecordSize:]
		put16(r, 0x00, ind.Name)
		put16(r, 0x02, uint16(ind.X))
		put16(r, 0x04, uint16(ind.Y))
		put16(r, 0x06, ind.Flags)
		put32(r, 0x08, ind.Prng[0])
		put32(r, 0x0C, ind.Prng[1])
		r[0x10], r[0x11] = ind.ObjectID, ind.UnderConstruction
		put16(r, 0x12, ind.FoundingYear)
		n := min(len(ind.Tiles), 32)
		r[0x14] = uint8(n)
		for k, tl := range ind.Tiles[:n] {
			for c := range 3 {
				put16(r, 0x15+6*k+2*c, uint16(tl[c]))
			}
		}
		put16(r, 0xD5, ind.Town)
		put16(r, 0xD7, uint16(ind.TileLoop[0]))
		put16(r, 0xD9, uint16(ind.TileLoop[1]))
		put16(r, 0xDB, uint16(ind.NumFarmTiles))
		put16(r, 0xDD, uint16(ind.NumIdleFarmTiles))
		r[0xDF], r[0xE0] = ind.ProductionRate, ind.Owner
		for k, v := range ind.StationsInRange {
			put32(r, 0xE1+4*k, v)
		}
		for k := range 2 {
			for j := range 4 {
				put16(r, 0x161+8*k+2*j, ind.ProducedStations[k][j])
				r[0x171+4*k+j] = ind.ProducedRatings[k][j]
			}
			put16(r, 0x179+2*k, ind.DailyProductionTarget[k])
			put16(r, 0x17D+2*k, ind.DailyProduction[k])
			put16(r, 0x181+2*k, ind.OutputBuffer[k])
			put16(r, 0x185+2*k, ind.ProducedMonthlyTotal[k])
			put16(r, 0x189+2*k, ind.ProducedPrevMonth[k])
			put16(r, 0x19F+2*k, ind.DeliveredMonthlyTotal[k])
			put16(r, 0x1A3+2*k, ind.DeliveredPrevMonth[k])
			r[0x1A7+k], r[0x1A9+k] = ind.PercentTransported[k], ind.HistorySize[k]
			copy(r[0x1AB+0xF0*k:0x1AB+0xF0*(k+1)], ind.History[k][:])
		}
		for k := range 3 {
			put16(r, 0x18D+2*k, ind.ReceivedMonthlyTotal[k])
			put16(r, 0x193+2*k, ind.ReceivedPrevMonth[k])
			put16(r, 0x199+2*k, ind.ReceivedDailyTotal[k])
		}
		put32(r, 0x38B, ind.HistoryMinProduction)
	}
	for i, p := range present {
		if !p {
			put16(gs, industriesGameStateOffset+i*industryRecordSize, 0xFFFF) // Industry::empty
		}
	}
}

// stationNameNull is StringIds::null: a free station slot.
const stationNameNull = 0xFFFF

func (sc *Scenario) putStations(gs []byte) {
	if len(sc.StationNames) < stationMaxCount {
		return
	}
	for i := range stationMaxCount {
		s := &sc.StationNames[i]
		r := gs[stationsGameStateOffset+i*stationRecordSize : stationsGameStateOffset+(i+1)*stationRecordSize]
		if s.Name == stationNameNull {
			// StationManager::zeroUnused: station = {} (name null, every
			// cargo's origin null, the rest zero).
			clear(r)
			put16(r, 0x00, stationNameNull)
			for c := range 32 {
				put16(r, 0x2E+c*13+2, 0xFFFF)
			}
			continue
		}
		put16(r, 0x00, s.Name)
		put16(r, 0x02, uint16(s.X))
		put16(r, 0x04, uint16(s.Y))
		put16(r, 0x06, uint16(s.Z))
		putLabelFrame(r[0x08:], s.LabelFrame)
		r[0x28], r[0x29] = s.Owner, s.NoTilesTimeout
		put16(r, 0x2A, s.Flags)
		put16(r, 0x2C, s.Town)
		for c, cs := range s.Cargo {
			e := r[0x2E+c*13:]
			put16(e, 0, cs.Quantity)
			put16(e, 2, cs.Origin)
			e[4], e[5], e[6], e[7] = cs.Flags, cs.Age, cs.Rating, cs.EnrouteAge
			put16(e, 8, uint16(cs.VehicleSpeed))
			e[10], e[11], e[12] = cs.VehicleAge, cs.IndustryID, cs.DensityPerTile
		}
		put16(r, 0x1CE, s.TileSize)
		for k := range s.Tiles {
			for c := range 3 {
				v := uint16(s.Tiles[k][c])
				if k >= int(s.TileSize) {
					v = 0 // zeroUnused: the station's unused tiles
				}
				put16(r, 0x1D0+6*k+2*c, v)
			}
		}
		r[0x3B0], r[0x3B1], r[0x3B2], r[0x3B3] = s.Var3B0, s.Var3B1, s.VisitedTypes, s.AirportRotation
		for c := range 3 {
			put16(r, 0x3B4+2*c, uint16(s.AirportStartPos[c]))
		}
		put32(r, 0x3BA, s.AirportOccupiedEdges)
	}
}

func (sc *Scenario) putAnimationsAndWaves(gs []byte) {
	n := min(len(sc.Animations), 8192)
	put16(gs, 0x154, uint16(n))
	for i, a := range sc.Animations[:n] {
		r := gs[animationsOffset+6*i:]
		r[0], r[1] = a.BaseZ, a.Type
		put16(r, 2, uint16(a.X))
		put16(r, 4, uint16(a.Y))
	}
	for i, w := range sc.Waves {
		r := gs[wavesOffset+6*i:]
		put16(r, 0, uint16(w.X))
		put16(r, 2, uint16(w.Y))
		put16(r, 4, w.Frame)
	}
}

func (sc *Scenario) putUserStrings(gs []byte) {
	for i, s := range sc.UserStrings {
		if i >= userStringsCount {
			break
		}
		putCString(gs[userStringsOffset+i*userStringSize:], s, userStringSize)
	}
}

func (sc *Scenario) putOrdersAndRoutings(gs []byte) {
	if sc.Orders != nil {
		o := gs[gameStateOrders : gameStateOrders+maxOrders]
		copy(o, sc.Orders)
		if int(sc.OrderTableLength) < len(o) {
			clear(o[sc.OrderTableLength:]) // OrderManager::zeroUnusedOrderTable
		}
	}
	for i, r := range sc.Routings {
		if i >= MaxRoutingVehicles*RoutingsPerVehicle {
			break
		}
		put16(gs, gameStateRoutings+2*i, r)
	}
}

// SavedGameFile is the S5 file of a saved game (prepareGameState): the
// header, the given save details, the object list, the game state and the
// tile elements.
func (sc *Scenario) SavedGameFile(details, requiredObjects []byte) *S5File {
	return &S5File{Header: SavedGameHeader(0), SaveDetails: details, RequiredObjects: requiredObjects,
		GameState: sc.EncodeGameState(), TileElements: EncodeTileElements(sc.Tiles)}
}

// SaveDetailsSize is sizeof(S5::SaveDetails).
const SaveDetailsSize = 0xC618

// EncodeSaveDetails is S5::SaveDetails (prepareSaveDetails): the player
// company's name and owner (formatted), the date, its performance index,
// the scenario's name, its challenge progress and flags, and the 250×200
// preview.
func EncodeSaveDetails(d SaveDetails, performanceIndex uint16, scenarioName string) []byte {
	b := make([]byte, SaveDetailsSize)
	putCString(b[0x000:], d.Company, 256)
	putCString(b[0x100:], d.Owner, 256)
	put32(b, 0x200, d.Date)
	put16(b, 0x204, performanceIndex)
	putCString(b[0x206:], scenarioName, 0x40)
	b[0x246] = d.ChallengeProgress
	copy(b[0x248:0x248+250*200], d.Image)
	put32(b, 0xC598, d.ChallengeFlags)
	return b
}
