package s5

// S5 save file format definitions

// S5Type indicates the type of S5 file
type S5Type uint8

const (
	S5TypeSavedGame S5Type = 0
	S5TypeScenario  S5Type = 1
	S5TypeObjects   S5Type = 2
	S5TypeLandscape S5Type = 3
)

// HeaderFlags are flags in the S5 header
type HeaderFlags uint32

const (
	HeaderFlagNone            HeaderFlags = 0
	HeaderFlagIsRaw           HeaderFlags = 1 << 0
	HeaderFlagIsDump          HeaderFlags = 1 << 1
	HeaderFlagIsTitleSequence HeaderFlags = 1 << 2
	HeaderFlagHasSaveDetails  HeaderFlags = 1 << 3
)

// Header is the S5 file header (32 bytes)
type Header struct {
	Type             S5Type
	Flags            HeaderFlags
	NumPackedObjects uint16
	Version          uint32
	Magic            uint32
	Padding          [20]byte
}

// HasFlags checks if the header has specific flags
func (h *Header) HasFlags(flags HeaderFlags) bool {
	return (h.Flags & flags) != 0
}

// CompanyFlags are flags for company state. Stub — only the zero value is
// defined; the full flag set has not been mapped yet.
//
// OpenLoco reference: src/OpenLoco/src/World/Company.h
//
//	enum class CompanyFlags : uint32_t
//
// Known flags in OpenLoco include (bit positions approximate):
//
//	unk0, unk1, unk2, sorted,
//	increasedPerformance, decreasedPerformance,
//	challengeCompleted, challengeFailed, challengeBeatenByOpponent,
//	bankrupt, autopayLoan
type CompanyFlags uint32

const (
	CompanyFlagNone CompanyFlags = 0
)

// SaveDetails contains save game metadata
type SaveDetails struct {
	Company           [256]byte
	Owner             [256]byte
	Date              uint32
	PerformanceIndex  uint16
	Scenario          [0x40]byte
	ChallengeProgress uint8
	Padding1          byte
	Image             [250 * 200]byte
	ChallengeFlags    CompanyFlags
	Padding2          [0x7C]byte // Pad to 0xC618
}

// S5FixFlags are flags for S5 file fixes
type S5FixFlags uint32

const (
	S5FixFlagNone S5FixFlags = 0
	S5FixFlag0    S5FixFlags = 1 << 0
	S5FixFlag1    S5FixFlags = 1 << 1
)

// LoadFlags control S5 file loading
type LoadFlags uint32

const (
	LoadFlagNone          LoadFlags = 0
	LoadFlagTitleSequence LoadFlags = 1 << 0
	LoadFlagTwoPlayer     LoadFlags = 1 << 1
	LoadFlagScenario      LoadFlags = 1 << 2
	LoadFlagLandscape     LoadFlags = 1 << 3
)

// SaveFlags control S5 file saving
type SaveFlags uint32

const (
	SaveFlagNone              SaveFlags = 0
	SaveFlagPackCustomObjects SaveFlags = 1 << 0
	SaveFlagScenario          SaveFlags = 1 << 1
	SaveFlagLandscape         SaveFlags = 1 << 2
	SaveFlagIsAutosave        SaveFlags = 1 << 28
	SaveFlagNoWindowClose     SaveFlags = 1 << 29
	SaveFlagRaw               SaveFlags = 1 << 30
	SaveFlagDump              SaveFlags = 1 << 31
)

// LoadError represents an error during S5 loading
type LoadError struct {
	Message string
}

func (e *LoadError) Error() string {
	return e.Message
}
