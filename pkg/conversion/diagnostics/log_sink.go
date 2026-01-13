package diagnostics

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "OpenLoco/Diagnostics/LogSink.h"
// #include <cassert>
// namespace OpenLoco::Diagnostics::Logging
// void LogSink::setWriteTimestamps(bool value)
// _writeTimestamps = value;
// bool LogSink::getWriteTimestamps() const noexcept
// orphan member: return _writeTimestamps;
// int LogSink::getIntendSize() const noexcept
// orphan member: return _intendSize;
// void LogSink::setIntendSize(int size)
// _intendSize = size;
// assert(_intendSize >= 0);
// void LogSink::incrementIntendSize()
// _intendSize++;
// void LogSink::decrementIntendSize()
// _intendSize--;
// assert(_intendSize >= 0);
// void LogSink::enableLevel(Level level)
// const auto levelMask = getLevelMask(level);
// _levelMask |= levelMask;
// void LogSink::disableLevel(Level level)
// const auto levelMask = getLevelMask(level);
// _levelMask &= ~levelMask;
// bool LogSink::passesLevelFilter(Level level) const noexcept
// const auto levelMask = getLevelMask(level);
// return (_levelMask & levelMask) != 0;
// void LogSink::setLevelMask(LevelMask mask)
// _levelMask = mask;
