package diagnostics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "OpenLoco/Diagnostics/LogTerminal.h"
// #include <OpenLoco/Platform/Platform.h>
// #include <fmt/chrono.h>
// #include <fmt/color.h>
// #include <fmt/format.h>
// namespace OpenLoco::Diagnostics::Logging
var ColourInfo = fmt.fg(fmt.color.light_gray) // auto
var ColourWarning = fmt.fg(fmt.color.yellow) // auto
var ColourError = fmt.fg(fmt.color.red) // auto
var ColourVerbose = fmt.fg(fmt.color.gray) // auto
// LogTerminal::LogTerminal()
// _vt100Enabled = Platform::enableVT100TerminalMode();
// static FILE* getOutputStream(Level level)
// switch (level)
// case Level::info:
// case Level::warning:
// case Level::verbose:
// orphan member: return stdout;
// case Level::error:
// orphan member: return stderr;
// default:
// break;
// orphan member: return stdout;
// static fmt::text_style getTextStyle(Level level)
// fmt::text_style textStyle{};
// switch (level)
// case Level::info:
// textStyle |= kColourInfo;
// break;
// case Level::warning:
// textStyle |= kColourWarning;
// break;
// case Level::error:
// textStyle |= kColourError;
// break;
// case Level::verbose:
// textStyle |= kColourVerbose;
// break;
// default:
// break;
// orphan member: return textStyle;
// void LogTerminal::print(Level level, std::string_view message)
// if (!passesLevelFilter(level))
// return;
// std::string timestamp;
// if (getWriteTimestamps())
// auto now = std::chrono::system_clock::now();
// timestamp = fmt::format("[{:%Y-%m-%d %H:%M:%S}] ", std::chrono::floor<std::chrono::seconds>(now));
// const int intendSize = getIntendSize();
// if (_vt100Enabled)
// fmt::print(getOutputStream(level), getTextStyle(level), "{}{:>{}}\n", timestamp, message, intendSize);
// else
// fmt::print(getOutputStream(level), "{}{}{:<{}}\n", timestamp, getLevelPrefix(level), message, intendSize);
