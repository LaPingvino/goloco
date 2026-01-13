package localisation

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <OpenLoco/Core/Exception.hpp>
// #include <array>
// #include <sfl/small_vector.hpp>
// namespace OpenLoco
type FormatArgumentsBuffer struct {
// std::array<std::byte, 16> _buffer{};
// std::byte* data()
// return _buffer.data();
}
// const std::byte* data() const
// return _buffer.data();
// func Capacity() int
// return _buffer.size();
type FormatArguments struct {
// std::byte* _buffer;
// std::byte* _bufferStart;
	Capacity int
// FormatArguments(std::byte* buffer, size_t length)
// _bufferStart = buffer;
// _buffer = _bufferStart;
// _capacity = length;
}
// FormatArguments(FormatArgumentsBuffer& buffer)
// _buffer = buffer.data();
// _bufferStart = _buffer;
// _capacity = buffer.capacity();
// FormatArguments(const FormatArgumentsBuffer& buffer)
// // FIXME: Create a view type for FormatArgumentsBuffer.
// _buffer = const_cast<std::byte*>(buffer.data());
// _bufferStart = _buffer;
// _capacity = buffer.capacity();
// FormatArguments()
// // TODO: refactor users to use non-static buffers
// static std::byte defaultBuffer[20];
// _bufferStart = _buffer = &*defaultBuffer;
// _capacity = std::size(defaultBuffer);
// template<typename... T>
// func Common(args T...) FormatArguments
// // TODO: refactor users to use non-static buffers
// orphan member: FormatArguments formatter;
// (formatter.push(args), ...);
// orphan member: return formatter;
// func MapToolTip() FormatArguments
// // TODO: refactor users to use non-static buffers
// static std::byte mapTooltipBuffer[40];
// orphan member: FormatArguments formatter{ mapTooltipBuffer, std::size(mapTooltipBuffer) };
// orphan member: return formatter;
// template<typename... T>
// func MapToolTip(args T...) FormatArguments
// // TODO: refactor users to use non-static buffers
// auto formatter = FormatArguments::mapToolTip();
// (formatter.push(args), ...);
// orphan member: return formatter;
// // Size in bytes to skip forward the buffer
// func Skip(size int) 
// auto* const nextOffset = getNextOffset(size);
// _buffer = nextOffset;
// template<typename T>
// func Push(arg T) 
// static_assert(sizeof(T) % 2 == 0, "Tried to push an odd number of bytes onto the format args!");
// auto* nextOffset = getNextOffset(sizeof(T));
// *(T*)_buffer = arg;
// _buffer = nextOffset;
// func Rewind() 
// _buffer = _bufferStart;
// const void* operator&() const
// orphan member: return _bufferStart;
// const std::byte* getBufferStart() const
// orphan member: return _bufferStart;
// func GetLength() int
// return _buffer - _bufferStart;
// func GetCapacity() int
// orphan member: return _capacity;
// private:
// std::byte* getNextOffset(const size_t size) const
// std::byte* const nextOffset = reinterpret_cast<std::byte*>(reinterpret_cast<std::byte*>(_buffer) + size);
// if (nextOffset > _bufferStart + _capacity)
// throw Exception::OutOfRange("FormatArguments: attempting to advance outside of buffer");
// orphan member: return nextOffset;
type FormatArgumentsView struct {
// const std::byte* args{};
// const std::byte* end{};
	// method: constexpr FormatArgumentsView() = default;
// FormatArgumentsView(const FormatArguments& newargs)
// : args(newargs.getBufferStart())
// , end(newargs.getBufferStart() + newargs.getCapacity()) {};
// FormatArgumentsView(const FormatArgumentsBuffer& newargs)
// : args(newargs.data())
// , end(newargs.data() + newargs.capacity()) {};
// template<typename T>
	// method: T pop()
// if (args == nullptr)
	T return
}
// orphan member: T value;
// std::memcpy(&value, args, sizeof(T));
// args += sizeof(T);
// orphan member: return value;
// template<typename T>
// func Skip() 
// if (args == nullptr)
// return;
// args += sizeof(T);
// template<typename T>
// func Push() 
// args -= sizeof(T);
