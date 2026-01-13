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
func init() {
	Unknown PacketKind = iota
}

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
// SKIPPED CONSTRUCTOR: const MaxPacketDataSize uint16 = kMaxPacketSize - sizeof(PacketHeader)
type Packet struct {
	Header PacketHeader
// uint8 data[kMaxPacketDataSize]{};
// template<typename T>
// const T* cast() const
// return reinterpret_cast<const T*>(data);
// template<PacketKind TKind, typename T>
// const T* as() const
// if (header.kind == TKind && header.dataSize >= sizeof(T))
// return cast<T>();
// orphan member: return nullptr;
type PingPacket struct {
func (p *PingPacket) Size() int {
	// uint32 gameCommandIndex{};
	// uint32 tick{};
	// uint32 srand0{};
	// uint32 srand1{};

type PingPacketConnectPacket struct {
func (p *PingPacket) Size() int {
	// uint16 version{};
	// char name[32]{};
type ConnectionResult int

	Success ConnectionResult = iota
	Error

type PingPacketConnectResponsePacket struct {
func (p *PingPacket) Size() int {
	result ConnectionResult
	// char message[256]{};

type PingPacketRequestStatePacket struct {
func (p *PingPacket) Size() int {
	// uint32 cookie{};

type PingPacketRequestStateResponse struct {
func (p *PingPacket) Size() int {
	// uint32 cookie{};
	// uint32 totalSize{};
	// uint16 numChunks{};

type PingPacketRequestStateResponseChunk struct {
func (p *PingPacket) Size() int {
	// uint32 cookie{};
	// uint16 index{};
	// uint32 offset{};
	// uint32 dataSize{};
	// uint8 data[kMaxPacketDataSize - 14]{};
// static_assert(sizeof(RequestStateResponseChunk) == kMaxPacketDataSize);
// /**
// * Extra state on top of S5 that we want to send over network
// */

type PingPacketExtraState struct {
	GameCommandIndex uint32
	Tick uint32

type PingPacketSendChatMessage struct {
func (p *PingPacket) Size() int {
	// uint16 length{};
	// char text[2048]{};
	// string_view getText() const
// SKIPPED CONSTRUCTOR: 	return std.string_view(text, length)
// static_assert(sizeof(SendChatMessage) <= kMaxPacketDataSize);

type PingPacketReceiveChatMessage struct {
func (p *PingPacket) Size() int {
	// client_id_t sender{};
	// uint16 length{};
	// char text[2048]{};
	// string_view getText() const
// static_assert(sizeof(SendChatMessage) <= kMaxPacketDataSize);

type PingPacketGameCommandPacket struct {
func (p *PingPacket) Size() int {
	// uint32 index{};
	// uint32 tick{};
	// CompanyId company{};
	// OpenLoco::GameCommands::registers regs;