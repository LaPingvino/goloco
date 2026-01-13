package core

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Traits.hpp"
// #include <cstddef>
// #include <cstdint>
// #include <istream>
// namespace OpenLoco
// namespace Utility
// // Obsolete methods, use new Stream APIs
// template<typename T1, typename T2, typename T3>
// std::basic_istream<T1, T2>& readData(std::basic_istream<T1, T2>& stream, T3* dst, size_t count)
// return stream.read((char*)dst, static_cast<uint64_t>(count) * sizeof(T3));
// template<typename T1, typename T2, typename T3>
// std::basic_istream<T1, T2>& readData(std::basic_istream<T1, T2>& stream, T3& dst)
// func ReadData(stream, dst, 1) return
// template<typename T3, typename T1, typename T2>
// func ReadValue(any /* std::basic_istream<T1 */ , stream T2>) T3
// orphan member: T3 result{};
// readData(stream, result);
// orphan member: return result;
type StreamMode int

const (
	None StreamMode = 0
	Read
	Write
)
type Stream struct {
// virtual ~Stream() = default;
// virtual size_t getLength() const noexcept = 0;
// virtual size_t getPosition() const noexcept = 0;
// virtual void setPosition(size_t) = 0;
// virtual void read(void*, size_t) = 0;
// virtual void write(const void*, size_t) = 0;
// template<typename T>
	// method: T readValue()
// static_assert(Traits::IsPrimitive<T>::value || Traits::IsPOD<T>::value, "Type must be primitive or POD");
	Tmp T
// read(&tmp, sizeof(T));
	Tmp return
}
// template<typename T>
// func WriteValue(src T) 
// static_assert(Traits::IsPrimitive<T>::value || Traits::IsPOD<T>::value, "Type must be primitive or POD");
// write(&src, sizeof(T));
