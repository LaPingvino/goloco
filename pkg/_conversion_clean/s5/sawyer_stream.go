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
func init() {
	Uncompressed SawyerEncoding = iota
}

	RunLengthSingle
	RunLengthMulti
	Rotate
)

type SawyerStreamReader struct {
	// Stream& _stream;
	DecodeBuffer  MemoryStream
	DecodeBuffer2 MemoryStream
	// []<const std::byte> decode(SawyerEncoding encoding, []<const std::byte> data);
	// method: static void decodeRunLengthSingle(MemoryStream& buffer, []<const std::byte> data);
	// method: static void decodeRunLengthMulti(MemoryStream& buffer, []<const std::byte> data);
	// method: static void decodeRotate(MemoryStream& buffer, []<const std::byte> data);
	// SawyerStreamReader(Stream& stream);
	// []<const std::byte> readChunk();
	// method: size_t readChunk(void* data, size_t maxDataLen);
	// method: void read(void* data, size_t dataLen);
	// method: bool validateChecksum();
type SawyerStreamWriter struct {
	// Stream& _stream;
	Checksum      uint32
	EncodeBuffer  MemoryStream
	EncodeBuffer2 MemoryStream
	// method: void writeStream(const void* data, size_t dataLen);
	// []<const std::byte> encode(SawyerEncoding encoding, []<const std::byte> data);
	// method: static void encodeRunLengthSingle(MemoryStream& buffer, []<const std::byte> data);
	// method: static void encodeRunLengthMulti(MemoryStream& buffer, []<const std::byte> data);
	// method: static void encodeRotate(MemoryStream& buffer, []<const std::byte> data);
	// SawyerStreamWriter(Stream& stream);
	// method: void writeChunk(SawyerEncoding chunkType, const void* data, size_t dataLen);
	// method: void write(const void* data, size_t dataLen);
	// method: void writeChecksum();
	// template<typename T>
	// method: void writeChunk(SawyerEncoding chunkType, const T& data)
	// writeChunk(chunkType, &data, sizeof(T));

// template<typename T>
// func Write(data T)
// write(&data, sizeof(T));