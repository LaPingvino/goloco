package network

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Network.h"
// #include "Packet.h"
// #include "Socket.h"
// #include <cstdint>
// #include <memory>
// #include <thread>
// namespace OpenLoco::Network
type NetworkBase struct {
// std::thread _receivePacketThread;
	EndReceivePacketLoop bool
	IsClosed bool
	// method: void receivePacketLoop();
// std::vector<std::unique_ptr<IUdpSocket>> _sockets;
	// method: void beginReceivePacketLoop();
	// method: void endReceivePacketLoop();
// virtual void onClose();
// virtual void onUpdate();
// virtual void onReceivePacket(IUdpSocket& socket, std::unique_ptr<INetworkEndpoint> endpoint, const Packet& packet);
// NetworkBase();
// virtual ~NetworkBase();
	// method: bool isClosed() const;
	// method: void close();
	// method: void update();
// virtual void sendChatMessage(std::string_view message) = 0;
}
