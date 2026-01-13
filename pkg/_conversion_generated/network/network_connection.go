package network

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Network.h"
// #include "Packet.h"
// #include "Socket.h"
// #include <cassert>
// #include <cstdint>
// #include <deque>
// #include <memory>
// #include <mutex>
// #include <optional>
// #include <queue>
// #include <thread>
// namespace OpenLoco::Network
type NetworkConnection struct {
}

type NetworkConnectionSentPacket struct {
	Timestamp uint32
	Packet Packet
}
// IUdpSocket* _socket;
// std::unique_ptr<INetworkEndpoint> _endpoint;
// std::mutex _sentPacketsSync;
// std::mutex _receivedPacketsSync;
// std::vector<SentPacket> _sentPackets;
// std::queue<Packet> _receivedPackets;
// std::deque<sequence_t> _receivedSequences;
	SendSequence uint16
	TimeOfLastReceivedPacket uint32
	// method: static uint32_t getTime();
	// method: bool checkOrRecordReceivedSequence(sequence_t sequence);
	// method: void receiveAcknowledgePacket(sequence_t sequence);
	// method: void sendAcknowledgePacket(sequence_t sequence);
	// method: void resendUndeliveredPackets();
	// method: void sendPacket(PacketKind kind, size_t dataSize, const void* packetData);
	// method: void logPacket(const Packet& packet, bool sent, bool resend);
// NetworkConnection(IUdpSocket* socket, std::unique_ptr<INetworkEndpoint> endpoint);
// const INetworkEndpoint& getEndpoint() const;
	// method: bool hasTimedOut() const;
	// method: void update();
	// method: void receivePacket(const Packet& packet);
	// method: void sendPacket(const Packet& packet);
// std::optional<Packet> takeNextPacket();
// template<typename T>
	// method: void sendPacket(const T& packetData)
// sendPacket(T::kind, packetData.size(), &packetData);
}
