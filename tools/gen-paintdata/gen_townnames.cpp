// Generates pkg/scene/testdata/townnames_hash.txt: TownManager.cpp's copyTownNameToBuffer + townNameFromNamesObject
// (sed-copied) over the scenario's town names object (townnames.bin dumped by a throwaway test, never committed).
#include <cstdio>
#include <cstdint>
#include <cstring>
#include <bit>
#include <vector>
#include <type_traits>
namespace OpenLoco {
  enum class LocationFlags : uint8_t { none = 0, adjacentToLargeWaterBody = 1, notMountaineous = 2, adjacentToSmallWaterBody = 4 };
  constexpr LocationFlags operator|(LocationFlags a, LocationFlags b) { return LocationFlags(uint8_t(a) | uint8_t(b)); }
  inline LocationFlags& operator|=(LocationFlags& a, LocationFlags b) { return a = a | b; }
  struct TownNamesObject { struct Category { uint8_t count; uint8_t bias; uint16_t offset; }; uint16_t name; Category categories[6]; };
  static std::vector<uint8_t> raw;
  namespace ObjectManager { template<class T> T* get() { return (T*)raw.data(); } }
    static LocationFlags copyTownNameToBuffer(const TownNamesObject* namesObj, uint32_t categoryOffset, uint16_t index, char* buffer)
    {
        // Offset into the string table for the requested category, located just after the names object header.
        auto* offsetPtr = reinterpret_cast<const std::byte*>(namesObj) + categoryOffset;
        auto srcOffset = *reinterpret_cast<const int16_t*>(offsetPtr + index * 2);

        auto* srcPtr = reinterpret_cast<const char*>(offsetPtr + srcOffset);
        strcpy(buffer, srcPtr);

        // Location flags are stored at the end of the string
        LocationFlags flags = *reinterpret_cast<const LocationFlags*>(srcPtr + strlen(srcPtr) + 1);
        return flags;
    }
    static LocationFlags townNameFromNamesObject(uint32_t rand, const char* buffer)
    {
        auto* namesObj = ObjectManager::get<TownNamesObject>();
        LocationFlags locationFlags = LocationFlags::none;

        // Town names are concatenated from six morpheme categories:
        // {CAT1}{CAT2}{CAT3}{CAT4}{CAT5}{CAT6}
        // Seperators are defined in each of the category strings; either '-', ' ', or nothing.
        // Categories can be completely empty, or can be skipped based on randomness.
        // e.g. "Fort " + "Apple " + "Green" for "Fort Apple Green"

        for (auto& category : namesObj->categories)
        {
            if (category.count == 0)
            {
                continue;
            }

            uint16_t ax = rand;
            uint16_t dx = category.count + category.bias;
            int16_t index = ((ax * dx) >> 16) - category.bias;

            if (index >= 0)
            {
                char* strEnd = const_cast<char*>(buffer + strlen(buffer));
                locationFlags |= copyTownNameToBuffer(namesObj, category.offset, index, strEnd);
            }

            for (auto shifts = category.count + category.bias; shifts > 0; shifts >>= 1)
            {
                rand = std::rotr(rand, 1);
            }
        }

        return locationFlags;
    }

}
using namespace OpenLoco;
int main(int argc, char** argv) {
  FILE* f = fopen(argv[1], "rb"); uint8_t b[65536]; size_t n = fread(b, 1, sizeof b, f); raw.assign(b, b + n); raw.resize(n + 512, 0);
  uint64_t h = 1469598103934665603ull; uint32_t r = 7;
  for (int i = 0; i < 5000; i++) { r = r * 1664525u + 1013904223u; char buf[256]{}; auto fl = townNameFromNamesObject(r, buf);
    if (i < 5) fprintf(stderr, "%u %s %d\n", r, buf, (int)fl);
    for (char* p = buf; *p; p++) { h ^= (uint8_t)*p; h *= 1099511628211ull; } h ^= 0x100 | (uint8_t)fl; h *= 1099511628211ull; }
  printf("%llx\n", (unsigned long long)h);
}
