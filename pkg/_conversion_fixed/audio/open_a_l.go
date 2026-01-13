package audio

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include <AL/alc.h>
// #include <cstdint>
// #include <span>
// #include <string>
// #include <vector>
// // TODO: When ubuntu dependencies upreved remove AL/alc.h and forward declare ALCcontext and ALCdevice
// namespace OpenAL
type Source struct {
	Id uint32
	// Source(uint32 id)
	// : _id(id)
}

// func Pause()
// func Play()
// func Stop()
// func SetBuffer(bufferId uint32)
// // value to be of the range 0.0f -> 1.0f
// func SetPitch(value float32)
// // value to be of the range 0.0f -> 1.0f
// func SetGain(value float32)
// func SetPosition(x float32, y float32, z float32)
// // value to be of the range -0.5f -> 0.5f
// func SetPan(value float32)
// func SetLooping(value bool)
// func IsPaused() bool
// func IsPlaying() bool
func GetId() uint32 {

type Context struct {
	// ALCcontext* _context = nullptr;
	// bool _isOpen = false;
	// ~Context();
	// method: void close();
	// method: void open(ALCdevice* device);
type Device struct {
	// ALCdevice* _device = nullptr;
	Context Context
	// bool _isOpen = false;
	// ~Device();
	// method: void open(const string& name);
	// method: void close();
	// std::vector<string> getAvailableDeviceNames() const;
type BufferManager struct {
	// std::vector<uint32> _buffers;
	// ~BufferManager();
	// method: uint32 allocate([]<const uint8> data, uint32 sampleRate, bool stereo, uint8 bits);
	// method: void deAllocate(uint32 id);
	// method: void dispose();
type SourceManager struct {
	// std::vector<uint32> _sources;
	// ~SourceManager();
	// method: Source allocate();
	// method: void deAllocate(const Source& source);
	// method: void dispose();

// func VolumeFromLoco(volume int32) float32
// func FreqFromLoco(freq int32) float32
// func PanFromLoco(pan int32) float32
