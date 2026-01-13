package s5

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Core/FileStream.h>
// #include <OpenLoco/Core/FileSystem.hpp>
// #include <OpenLoco/Core/MemoryStream.h>
// #include <cstdint>
// #include <memory>
// #include <span>
// namespace OpenLoco
type SawyerEncoding int

const (
	Uncompressed SawyerEncoding = iota
	RunLengthSingle
	RunLengthMulti
	Rotate
)
type SawyerStreamReader struct {
// Stream& _stream;
	DecodeBuffer MemoryStream
	DecodeBuffer2 MemoryStream
// std::span<const std::byte> decode(SawyerEncoding encoding, std::span<const std::byte> data);
	// method: static void decodeRunLengthSingle(MemoryStream& buffer, std::span<const std::byte> data);
	// method: static void decodeRunLengthMulti(MemoryStream& buffer, std::span<const std::byte> data);
	// method: static void decodeRotate(MemoryStream& buffer, std::span<const std::byte> data);
// SawyerStreamReader(Stream& stream);
// std::span<const std::byte> readChunk();
	// method: size_t readChunk(void* data, size_t maxDataLen);
	// method: void read(void* data, size_t dataLen);
	// method: bool validateChecksum();
}
type SawyerStreamWriter struct {
// Stream& _stream;
	Checksum uint32
	EncodeBuffer MemoryStream
	EncodeBuffer2 MemoryStream
	// method: void writeStream(const void* data, size_t dataLen);
// std::span<const std::byte> encode(SawyerEncoding encoding, std::span<const std::byte> data);
	// method: static void encodeRunLengthSingle(MemoryStream& buffer, std::span<const std::byte> data);
	// method: static void encodeRunLengthMulti(MemoryStream& buffer, std::span<const std::byte> data);
	// method: static void encodeRotate(MemoryStream& buffer, std::span<const std::byte> data);
// SawyerStreamWriter(Stream& stream);
	// method: void writeChunk(SawyerEncoding chunkType, const void* data, size_t dataLen);
	// method: void write(const void* data, size_t dataLen);
	// method: void writeChecksum();
// template<typename T>
	// method: void writeChunk(SawyerEncoding chunkType, const T& data)
// writeChunk(chunkType, &data, sizeof(T));
}
// template<typename T>
// func Write(data T) 
// write(&data, sizeof(T));
