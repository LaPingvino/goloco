package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Objects/Object.h"
// #include "Types.hpp"
// #include <OpenLoco/Core/EnumFlags.hpp>
// #include <OpenLoco/Core/FileSystem.hpp>
// #include <cstdint>
// #include <memory>
// #include <vector>
// namespace OpenLoco
type CompanyFlags int

const (
type GameStateFlags int

const (
// forward: class Stream;
)
// namespace OpenLoco::Scenario
// forward: struct Options;
// namespace OpenLoco::S5
// forward: struct Options;
type S5Type int

const (
	SavedGame S5Type = 0
	Scenario S5Type = 1
	Objects S5Type = 2
	Landscape S5Type = 3
)
type HeaderFlags int

const (
	None HeaderFlags = 0
	IsRaw HeaderFlags = 1 << 0
	IsDump HeaderFlags = 1 << 1
	IsTitleSequence HeaderFlags = 1 << 2
	HasSaveDetails HeaderFlags = 1 << 3
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(HeaderFlags);
type Header struct {
	Type S5Type
	Flags HeaderFlags
	NumPackedObjects uint16
	Version uint32
	Magic uint32
// std::byte padding[20];
	// method: constexpr bool hasFlags(HeaderFlags flagsToTest) const
// return (flags & flagsToTest) != HeaderFlags::none;
}
// static_assert(sizeof(Header) == 0x20);
type SaveDetails struct {
// char company[256];                   // 0x000
// char owner[256];                     // 0x100
	Date uint32
	PerformanceIndex uint16
// char scenario[0x40];                 // 0x206
	ChallengeProgress uint8
// std::byte pad_247;                   // 0x247
// uint8_t image[250 * 200];            // 0x248
	ChallengeFlags CompanyFlags
// std::byte pad_C59C[0xC618 - 0xC59C]; // 0xC59C
}
// static_assert(sizeof(SaveDetails) == 0xC618);
type S5FixFlags int

const (
	None S5FixFlags = 0
	FixFlag0 S5FixFlags = 1 << 0
	FixFlag1 S5FixFlags = 1 << 1
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(S5FixFlags);
// forward: struct S5File;
type LoadFlags int

const (
	None LoadFlags = 0
	TitleSequence LoadFlags = 1 << 0
	TwoPlayer LoadFlags = 1 << 1
	Scenario LoadFlags = 1 << 2
	Landscape LoadFlags = 1 << 3
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(LoadFlags);
type SaveFlags int

const (
	None SaveFlags = 0
	PackCustomObjects SaveFlags = 1 << 0
	Scenario SaveFlags = 1 << 1
	Landscape SaveFlags = 1 << 2
	IsAutosave SaveFlags = 1 << 28
	NoWindowClose SaveFlags = 1 << 29
	Raw SaveFlags = 1 << 30
	Dump SaveFlags = 1 << 31
)
// OPENLOCO_ENABLE_ENUM_OPERATORS(SaveFlags);
type LoadError struct {
	ErrorCode int8
	ErrorMessage StringId
// std::vector<ObjectHeader> objectList{};
}
const MagicNumber uint32 = 0x62300
// constexpr const char* extensionSC5 = ".SC5";
// constexpr const char* extensionSV5 = ".SV5";
// constexpr const char* filterSC5 = "*.SC5";
// constexpr const char* filterSV5 = "*.SV5";
// func ExportGameStateToFile(path fs::path, flags SaveFlags) bool
// func ExportGameStateToFile(stream Stream, flags SaveFlags) bool
// const LoadError& getLastLoadError();
// func ResetLastLoadError() 
// std::unique_ptr<S5File> importSave(Stream& stream);
// func ImportSaveToGameState(path fs::path, flags LoadFlags) bool
// func ImportSaveToGameState(stream Stream, flags LoadFlags) bool
// std::unique_ptr<SaveDetails> readSaveDetails(const fs::path& path);
// std::unique_ptr<Scenario::Options> readScenarioOptions(const fs::path& path);
