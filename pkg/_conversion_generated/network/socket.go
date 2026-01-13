package network

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <memory>
// #include <string>
// #include <vector>
// namespace OpenLoco::Network
type Protocol int

const (
	Any Protocol = iota
	Ipv4
	Ipv6
)
type SocketStatus int

const (
	Closed SocketStatus = iota
	Waiting
	Resolving
	Connecting
	Connected
	Listening
)
type NetworkReadPacket int

const (
	Success NetworkReadPacket = iota
	NoData
	MoreData
	Disconnected
)
// /**
// * Represents an address and port.
// */
type INetworkEndpoint struct {
// virtual ~INetworkEndpoint()
}
// virtual Protocol getProtocol() const = 0;
// virtual std::string getIpAddress() const = 0;
// virtual std::string getHostname() const = 0;
// virtual std::unique_ptr<INetworkEndpoint> clone() const = 0;
// virtual bool equals(const INetworkEndpoint& other) const = 0;
// /**
// * Represents a UDP socket / listener.
// */
type IUdpSocket struct {
// virtual ~IUdpSocket() = default;
// virtual SocketStatus getStatus() const = 0;
// virtual const char* getError() const = 0;
// virtual const char* getHostName() const = 0;
// virtual Protocol getProtocol() const = 0;
// virtual std::string getIpAddress() const = 0;
// virtual void listen(Protocol procotol, uint16_t port) = 0;
// virtual void listen(Protocol procotol, const std::string& address, uint16_t port) = 0;
// virtual size_t sendData(Protocol protocol, const std::string& address, uint16_t port, const void* buffer, size_t size) = 0;
// virtual size_t sendData(const INetworkEndpoint& destination, const void* buffer, size_t size) = 0;
// virtual NetworkReadPacket receiveData(
// void* buffer, size_t size, size_t* sizeReceived, std::unique_ptr<INetworkEndpoint>* sender)
// = 0;
// virtual void close() = 0;
}
// namespace Socket
// [[nodiscard]] std::unique_ptr<IUdpSocket> createUdp();
// std::unique_ptr<INetworkEndpoint> resolve(Protocol protocol, const std::string& address, uint16_t port);
