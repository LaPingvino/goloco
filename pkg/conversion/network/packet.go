package network

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "GameCommands/GameCommands.h"
// #include "Network.h"
// #include <cstdint>
// #include <cstdlib>
// #include <string_view>
// namespace OpenLoco::Network
type Sequence_t = uint16
type PacketKind int

const (
	Unknown PacketKind = iota
	Ack
	Ping
	Connect
	ConnectResponse
	RequestState
	RequestStateResponse
	RequestStateResponseChunk
	SendChatMessage
	ReceiveChatMessage
	GameCommand
)
type PacketHeader struct {
	Kind PacketKind
	Sequence sequence_t
	DataSize uint16
}
const MaxPacketDataSize uint16 = kMaxPacketSize - sizeof(PacketHeader)
type Packet struct {
	Header PacketHeader
// uint8_t data[kMaxPacketDataSize]{};
// template<typename T>
// const T* cast() const
// return reinterpret_cast<const T*>(data);
}
// template<PacketKind TKind, typename T>
// const T* as() const
// if (header.kind == TKind && header.dataSize >= sizeof(T))
// return cast<T>();
// orphan member: return nullptr;
type PingPacket struct {
func (p *PingPacket) Size() int {
	// uint32_t gameCommandIndex{};
	// uint32_t tick{};
	// uint32_t srand0{};
	// uint32_t srand1{};
}
}

type PingPacketConnectPacket struct {
func (p *PingPacket) Size() int {
	// uint16_t version{};
	// char name[32]{};
}
type ConnectionResult int

const (
	Success ConnectionResult = iota
	Error
)
}

type PingPacketConnectResponsePacket struct {
func (p *PingPacket) Size() int {
	var result ConnectionResult
	// char message[256]{};
}
}

type PingPacketRequestStatePacket struct {
func (p *PingPacket) Size() int {
	// uint32_t cookie{};
}
}

type PingPacketRequestStateResponse struct {
func (p *PingPacket) Size() int {
	// uint32_t cookie{};
	// uint32_t totalSize{};
	// uint16_t numChunks{};
}
}

type PingPacketRequestStateResponseChunk struct {
func (p *PingPacket) Size() int {
	// uint32_t cookie{};
	// uint16_t index{};
	// uint32_t offset{};
	// uint32_t dataSize{};
	// uint8_t data[kMaxPacketDataSize - 14]{};
}
// static_assert(sizeof(RequestStateResponseChunk) == kMaxPacketDataSize);
// /**
// * Extra state on top of S5 that we want to send over network
// */
}

type PingPacketExtraState struct {
	GameCommandIndex uint32
	Tick uint32
}
}

type PingPacketSendChatMessage struct {
func (p *PingPacket) Size() int {
	// uint16_t length{};
	// char text[2048]{};
	// std::string_view getText() const
	return std.string_view(text, length)
	}
}
// static_assert(sizeof(SendChatMessage) <= kMaxPacketDataSize);
}

type PingPacketReceiveChatMessage struct {
func (p *PingPacket) Size() int {
	// client_id_t sender{};
	// uint16_t length{};
	// char text[2048]{};
	// std::string_view getText() const
	return std.string_view(text, length)
	}
}
// static_assert(sizeof(SendChatMessage) <= kMaxPacketDataSize);
}

type PingPacketGameCommandPacket struct {
func (p *PingPacket) Size() int {
	// uint32_t index{};
	// uint32_t tick{};
	// CompanyId company{};
	// OpenLoco::GameCommands::registers regs;
}
}
