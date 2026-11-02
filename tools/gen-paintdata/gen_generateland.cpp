// Generates pkg/sim/testdata/generateland.txt: MapGenerator.cpp generateLand body (sed-copied, tile lookup stubbed) over 40 random 24x24 height maps.
#include <cstdio>
#include <cstdint>
#include <algorithm>
#include <vector>
using MicroZ = uint8_t;
constexpr int kMicroToSmallZStep = 4, kSmallZStep = 4;
namespace SurfaceSlope { constexpr uint8_t flat = 0, doubleHeight = 1<<4;
 namespace CornerUp { constexpr uint8_t all=0xF, north=1, east=2, south=4, west=8; }
 namespace CornerDown { constexpr uint8_t west = CornerUp::all & ~CornerUp::west, south = CornerUp::all & ~CornerUp::south, east = CornerUp::all & ~CornerUp::east, north = CornerUp::all & ~CornerUp::north; } }
struct TilePos2 { int x, y; TilePos2 operator+(TilePos2 o) const { return {x+o.x, y+o.y}; } };
struct HM { int w; std::vector<uint8_t> d; uint8_t operator[](TilePos2 p) const { return d[p.y*w+p.x]; } };
struct Surf { int baseZ_=0, slope_=0, clearZ_=0; void setBaseZ(int z){baseZ_=z;} int baseZ() const {return baseZ_;} void setSlope(int s){slope_=s;} int slopeCorners() const {return slope_&0xF;} bool isSlopeDoubleHeight() const {return slope_&0x10;} void setClearZ(int z){clearZ_=z;} };
int main(){
  const int W=24,H=24; uint32_t r=12345;
  for(int map=0; map<40; map++){
    HM heightMap{W, std::vector<uint8_t>(W*H)};
    for(auto& v: heightMap.d){ r=r*1103515245u+12345u; v=(r>>16)% (map<20?4:12); }
    for(int i=0;i<W*H;i++) printf("%d ", heightMap.d[i]); printf("\n");
    for(int y=1;y<=H-2;y++) for(int x=1;x<=W-2;x++){ TilePos2 pos{x,y}; Surf s; Surf* surfaceElement=&s;
            const MicroZ q00 = heightMap[pos + TilePos2{ -1, -1 }];
            const MicroZ q01 = heightMap[pos + TilePos2{ 0, -1 }];
            const MicroZ q10 = heightMap[pos + TilePos2{ -1, 0 }];
            const MicroZ q11 = heightMap[pos + TilePos2{ 0, 0 }];


            const MicroZ baseHeight = std::min({ q00, q01, q10, q11 });
            surfaceElement->setBaseZ(baseHeight * kMicroToSmallZStep);

            uint8_t currentSlope = SurfaceSlope::flat;

            // First, figure out basic corner style
            if (q00 > baseHeight)
            {
                currentSlope |= SurfaceSlope::CornerUp::south;
            }
            if (q01 > baseHeight)
            {
                currentSlope |= SurfaceSlope::CornerUp::east;
            }
            if (q10 > baseHeight)
            {
                currentSlope |= SurfaceSlope::CornerUp::west;
            }
            if (q11 > baseHeight)
            {
                currentSlope |= SurfaceSlope::CornerUp::north;
            }

            // Now, deduce if we should go for double height
            // clang-format off
            if ((currentSlope == SurfaceSlope::CornerDown::north && q00 - baseHeight >= 2) ||
                (currentSlope == SurfaceSlope::CornerDown::west  && q01 - baseHeight >= 2) ||
                (currentSlope == SurfaceSlope::CornerDown::east  && q10 - baseHeight >= 2) ||
                (currentSlope == SurfaceSlope::CornerDown::south && q11 - baseHeight >= 2))
            {
                currentSlope |= SurfaceSlope::doubleHeight;
            }
            // clang-format on

            surfaceElement->setSlope(currentSlope);

            auto clearZ = surfaceElement->baseZ();
            if (surfaceElement->slopeCorners())
            {
                clearZ += kSmallZStep;
            }
            if (surfaceElement->isSlopeDoubleHeight())
            {
                clearZ += kSmallZStep;
            }
            surfaceElement->setClearZ(clearZ);
      printf("%d %d %d ", s.baseZ_, s.slope_, s.clearZ_); }
    printf("\n"); } }
