package diagnostics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Logging.h"
// #include "LogSink.h"
// #include <algorithm>
// #include <cstdio>
// #include <fmt/color.h>
// #include <stdarg.h>
// #include <vector>
// namespace OpenLoco::Diagnostics::Logging
// static std::vector<std::shared_ptr<LogSink>> _sinks;
// namespace Detail
// func Print(level Level, message string) 
// if (_sinks.empty())
// // Its possible that the logging interface is used from static initializers and
// // may not have any sinks installed at this point, we redirect the output to
// // stdout/stderr in this case.
// fmt::print(level == Level::error ? stderr : stdout, "{}", message);
// return;
// for (auto& sink : _sinks)
// if (!sink->passesLevelFilter(level))
// continue;
// sink->print(level, message);
// func PassesLevelFilter(level Level) bool
// if (_sinks.empty())
// orphan member: return true;
// for (auto& sink : _sinks)
// if (sink->passesLevelFilter(level))
// orphan member: return true;
// orphan member: return false;
// func InstallSink(sink any /* std::shared_ptr<LogSink> */ ) 
// _sinks.push_back(sink);
// func RemoveSink(sink any /* std::shared_ptr<LogSink> */ ) 
// auto it = std::find(_sinks.begin(), _sinks.end(), sink);
// if (it != _sinks.end())
// _sinks.erase(it);
// func IncrementIntend() 
// for (auto& sink : _sinks)
// sink->incrementIntendSize();
// func DecrementIntend() 
// for (auto& sink : _sinks)
// sink->decrementIntendSize();
// func EnableLevel(level Level) 
// for (auto& sink : _sinks)
// sink->enableLevel(level);
// func DisableLevel(level Level) 
// for (auto& sink : _sinks)
// sink->disableLevel(level);
