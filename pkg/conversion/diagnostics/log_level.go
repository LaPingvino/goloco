package diagnostics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <cstdint>
// #include <string_view>
// namespace OpenLoco::Diagnostics::Logging
type Level int

const (
	Info Level = iota
	Warning
	Error
	Verbose
	All
)
type LevelMask = uint32
// template<typename... TArgs>
// func GetLevelMask(args TArgs...) LevelMask
// orphan member: LevelMask mask{};
// const auto maskForLevel = [](Level level) {
// if (level == Level::all)
// return ~0U;
// else
// return 1U << static_cast<LevelMask>(level);
// ((mask = mask | maskForLevel(args)), ...);
// orphan member: return mask;
// func GetLevelMaskFromName(name string) LevelMask
// if (name == "info")
// func GetLevelMask(Level::info) return
// func If("warning" name ==) else
// func GetLevelMask(Level::warning) return
// func If("error" name ==) else
// func GetLevelMask(Level::error) return
// func If("verbose" name ==) else
// func GetLevelMask(Level::verbose) return
// func If("all" name ==) else
// func GetLevelMask(Level::all) return
// else
// orphan member: return 0;
// constexpr std::string_view getLevelPrefix(Level level)
// switch (level)
// case Level::info:
// return "[INF] ";
// case Level::warning:
// return "[WRN] ";
// case Level::error:
// return "[ERR] ";
// case Level::verbose:
// return "[VER] ";
// default:
// break;
// return "[INVALID] ";
