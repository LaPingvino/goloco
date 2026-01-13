package network

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "Network.h"
// #include "NetworkBase.h"
// #include "NetworkConnection.h"
// #include "Socket.h"
// #include <mutex>
// namespace OpenLoco::Network
// forward: class NetworkConnection;
type Client struct {
	Id client_id_t
// std::unique_ptr<NetworkConnection> connection;
// std::string name;
}
type ChatMessage struct {
	Sender client_id_t
// std::string message;
}
type NetworkServer struct {
// std::mutex _incomingConnectionsSync;
// std::mutex _chatMessageQueueSync;
// std::vector<std::unique_ptr<NetworkConnection>> _incomingConnections;
// std::vector<std::unique_ptr<Client>> _clients;
// std::queue<ChatMessage> _chatMessageQueue;
// client_id_t _nextClientId = 1;
	LastPing uint32
	GameCommandIndex uint32
// std::queue<GameCommandPacket> _gameCommands;
// Client* findClient(const INetworkEndpoint& endpoint);
	// method: void createNewClient(std::unique_ptr<NetworkConnection> conn, const ConnectPacket& packet);
	// method: void onReceivePacketFromClient(Client& client, const Packet& packet);
	// method: void onReceiveStateRequestPacket(Client& client, const RequestStatePacket& packet);
	// method: void onReceiveSendChatMessagePacket(Client& client, const SendChatMessage& packet);
	// method: void onReceiveGameCommandPacket(Client& client, const GameCommandPacket& packet);
	// method: void removedTimedOutClients();
	// method: void sendPings();
	// method: void sendChatMessages();
	// method: void processIncomingConnections();
	// method: void processPackets();
	// method: void updateClients();
// template<typename T>
	// method: void sendPacketToAll(const T& packet)
// for (auto& client : _clients)
// client->connection->sendPacket(packet);
}
// protected:
// func OnClose() 
// func OnReceivePacket(socket IUdpSocket, endpoint any /* std::unique_ptr<INetworkEndpoint> */ , packet Packet) 
// func OnUpdate() 
// public:
// ~NetworkServer() override;
// func Listen(bind string, port port_t) 
// func SendChatMessage(message string) 
// func SendGameCommand(index uint32, tick uint32, company CompanyId, regs OpenLoco::GameCommands::registers) 
// func QueueGameCommand(company CompanyId, regs OpenLoco::GameCommands::registers) 
// func RunGameCommands() 
