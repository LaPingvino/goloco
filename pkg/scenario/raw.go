package scenario

import (
	"encoding/binary"

	"github.com/LaPingvino/goloco/pkg/assets"
)

// raw.go — the file's own bytes, kept from the load for the save (s5save.go):
// every byte goloco does not model is written back as it was read, and the
// modelled fields are written over them. As S5.cpp loadSave reads a file into
// one S5File: the header, the save details or scenario options, the packed
// objects, the required objects and ONE GameState — a scenario's three
// game-state chunks laid into it at their offsets, an old GameStateType2
// converted (importGameStateType2).
//
// OpenLoco reference: src/OpenLoco/src/S5/S5.cpp loadSave

// GameStateSize is sizeof(S5::GameState).
const GameStateSize = 0x4A0644

// Offsets in S5::GameState (S5GameState.h) of the chunks a scenario splits
// it into: general (to the companies), towns (+ industries, stations: 0x123480
// bytes), animations (+ waves, user strings, routings, orders: 0x79D80).
const (
	gsTownsOffset      = 0x092444
	gsTownsChunkSize   = 0x123480
	gsAnimationsOffset = 0x4268C4
	gsAnimationsSize   = 0x79D80
	generalStateSize   = 0xB96C
	fixFlagsOffset     = 0x434 // general.fixFlags
	fixFlag0, fixFlag1 = 1, 2  // S5FixFlags
)

// Raw is the loaded file's chunks, decoded.
type Raw struct {
	Header          assets.S5Header
	SaveDetails     []byte // nil without HeaderFlags::hasSaveDetails
	Options         []byte // a scenario's / landscape's S5::Options; nil for a saved game
	RequiredObjects []byte // the object list (ObjectHeader × kMaxObjects)
	// GameState is the whole S5::GameState (GameStateSize bytes, the current
	// layout); fixFlag1 set, as loadSave sets it.
	GameState []byte
	// TileElements is the TileElements chunk as read (nil when the file has
	// no map): what the save's tile elements are compared with in the tests.
	TileElements []byte
}

// gameStateFromScenarioChunks lays a scenario's three game-state chunks into
// one GameState (loadSave: readChunk(&gameState), readChunk(&gameState.towns),
// readChunk(&gameState.animations), each up to the struct's end).
func gameStateFromScenarioChunks(general, towns, animations []byte) []byte {
	gs := make([]byte, GameStateSize)
	copy(gs, general)
	copy(gs[gsTownsOffset:], towns)
	copy(gs[gsAnimationsOffset:], animations)
	setFixFlag1(gs)
	return gs
}

// gameStateFromSaveChunk is loadSave's saved-game branch: a chunk without
// fixFlag0 / fixFlag1 is a GameStateType2 and is converted; then fixFlag1.
func gameStateFromSaveChunk(chunk []byte) []byte {
	gs := make([]byte, GameStateSize)
	if len(chunk) > fixFlagsOffset+1 && binary.LittleEndian.Uint16(chunk[fixFlagsOffset:])&(fixFlag0|fixFlag1) == 0 {
		importGameStateType2(gs, chunk)
	} else {
		copy(gs, chunk)
	}
	setFixFlag1(gs)
	return gs
}

func setFixFlag1(gs []byte) {
	binary.LittleEndian.PutUint16(gs[fixFlagsOffset:], binary.LittleEndian.Uint16(gs[fixFlagsOffset:])|fixFlag1)
}

// importGameStateType2 is S5GameState.cpp importGameStateType2: the general
// state as it is, every company through importCompanyType2, everything after
// the companies as it is (the same layout, 0x1C20 bytes on).
// importCompanyType2 copies the fields up to cargoUnitsTotalDistance and
// jailStatus; cargoUnitsDistanceHistory, which CompanyType2 lacks, is zero.
// (Its field-by-field copy also zeroes CompanyType2's padding; here the
// padding is copied — no state lives there.)
func importGameStateType2(dst, src []byte) {
	const (
		companyType2Size = 0x8DC8
		jailType2        = 0x8C54
		jailType1        = 0x8E34
	)
	copy(dst[:generalStateSize], src)
	for i := range companyMaxCount {
		s := src[companiesGameStateOffset+i*companyType2Size:]
		d := dst[companiesGameStateOffset+i*companyRecordSize:]
		copy(d[:jailType2], s[:jailType2])
		copy(d[jailType1:jailType1+2], s[jailType2:jailType2+2])
	}
	rest := companiesGameStateOffset + companyMaxCount*companyType2Size
	if rest < len(src) {
		copy(dst[gsTownsOffset:], src[rest:])
	}
}
