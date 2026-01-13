package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "MemoryStream.h"
// #include "Exception.hpp"
// #include <algorithm>
// #include <cstring>
// namespace OpenLoco
// MemoryStream::~MemoryStream()
// if (_data != nullptr)
// std::free(_data);
// void MemoryStream::reserve(size_t len)
// if (len == 0 || len < _capacity)
// return;
// auto* newData = static_cast<std::byte*>(std::realloc(_data, len));
// if (newData == nullptr)
// throw Exception::BadAllocation();
// _data = newData;
// _capacity = len;
// void MemoryStream::resize(size_t len)
// if (len == 0)
// clear();
// return;
// if (len < _capacity)
// _length = len;
// return;
// reserve(len);
// _length = len;
// void MemoryStream::clear()
// _offset = 0;
// _length = 0;
// const std::byte* MemoryStream::data() const
// orphan member: return _data;
// std::byte* MemoryStream::data()
// orphan member: return _data;
// std::span<std::byte> MemoryStream::getSpan()
// return std::span(_data, _length);
// std::span<const std::byte> MemoryStream::getSpan() const
// return std::span(_data, _length);
// size_t MemoryStream::getLength() const noexcept
// orphan member: return _length;
// size_t MemoryStream::getPosition() const noexcept
// orphan member: return _offset;
// void MemoryStream::setPosition(size_t position)
// _offset = std::min(_length, position);
// void MemoryStream::read(void* buffer, size_t len)
// const auto maxReadLength = std::min(len, _length - _offset);
// if (len > maxReadLength)
// throw Exception::RuntimeError("Failed to read data");
// std::memcpy(buffer, _data + _offset, len);
// _offset += maxReadLength;
// // TODO: Move this somewhere more sensible.
// template<typename T>
// func AlignTo(val T, align T) T
// if (val % align == 0)
// orphan member: return val;
// return val + (align - (val % align));
// void MemoryStream::write(const void* buffer, size_t len)
// const auto spaceLeft = _capacity - _offset;
// // Check if we have to expand the current buffer.
// if (len > spaceLeft)
var GrowthFactor = 2.0 // auto
// constexpr std::size_t kPageSize = 0x1000;
// const auto newCapacity = _capacity + len;
// const auto finalCapacity = alignTo(static_cast<std::size_t>(newCapacity * kGrowthFactor), kPageSize);
// auto* newData = static_cast<std::byte*>(std::realloc(_data, finalCapacity));
// if (newData == nullptr)
// throw Exception::BadAllocation();
// _data = newData;
// _capacity = finalCapacity;
// // Copy the data into the buffer.
// std::memcpy(_data + _offset, buffer, len);
// const auto spaceAvailable = _length - _offset;
// if (len > spaceAvailable)
// _length += len - spaceAvailable;
// _offset += len;
