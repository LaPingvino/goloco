package network

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Network.h"
// #include "NetworkBase.h"
// #include "Socket.h"
// #include <cstdint>
// #include <list>
// #include <span>
// #include <vector>
// namespace OpenLoco::Network
// forward: class NetworkConnection;
type NetworkClientStatus int

const (
	None NetworkClientStatus = iota
	Connecting
	ConnectedSuccessfully
	WaitingForState
	Connected
	Closed
)
type NetworkClient struct {
// std::unique_ptr<INetworkEndpoint> _serverEndpoint;
// std::unique_ptr<NetworkConnection> _serverConnection;
	Status NetworkClientStatus
	Timeout uint32
	LocalGameCommandIndex uint32
	ServerGameCommandIndex uint32
	LocalTick uint32
	ServerTick uint32
// std::list<GameCommandPacket> _receivedGameCommands;
}

type NetworkClientReceivedChunk struct {
	Offset uint32
// std::vector<uint8_t> data;
}
	RequestStateCookie uint32
	RequestStateTotalSize uint32
	RequestStateNumChunks uint16
// std::vector<ReceivedChunk> _requestStateChunksReceived;
	RequestStateReceivedBytes uint32
	RequestStateReceivedChunks uint32
	// method: void onCancel();
	// method: void processReceivedPackets();
	// method: bool hasTimedOut() const;
	// method: void onReceivePacketFromServer(const Packet& packet);
	// method: void processFullState(std::span<uint8_t const> data);
	// method: void updateLocalTick();
	// method: void initStatus(std::string_view text);
	// method: void setStatus(std::string_view text);
	// method: void clearStatus();
	// method: void endStatus(std::string_view text);
	// method: void sendConnectPacket();
	// method: void sendRequestStatePacket();
	// method: void receiveConnectionResponsePacket(const ConnectResponsePacket& response);
	// method: void receiveRequestStateResponsePacket(const RequestStateResponse& response);
	// method: void receiveRequestStateResponseChunkPacket(const RequestStateResponseChunk& responseChunk);
	// method: void receiveChatMessagePacket(const ReceiveChatMessage& packet);
	// method: void receivePingPacket(const PingPacket& packet);
	// method: void receiveGameCommandPacket(const GameCommandPacket& packet);
	// method: void onClose() override;
	// method: void onUpdate() override;
	// method: void onReceivePacket(IUdpSocket& socket, std::unique_ptr<INetworkEndpoint> endpoint, const Packet& packet) override;
// ~NetworkClient() override;
	// method: NetworkClientStatus getStatus() const;
	// method: uint32_t getLocalTick() const;
	// method: void connect(std::string_view host, port_t port);
	// method: void sendChatMessage(std::string_view message) override;
	// method: void sendGameCommand(CompanyId company, const OpenLoco::GameCommands::registers& regs);
	// method: bool shouldProcessTick(uint32_t tick) const;
	// method: void runGameCommandsForTick(uint32_t tick);
}
