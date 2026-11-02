// Generates pkg/sim/testdata/quartertile_rotate.txt: QuarterTile::rotate and Numerics::rotl4bit (upstream headers) for every value and rotation, FNV-1a hashed.
#include <cstdio>
#include <cstdint>
#include <OpenLoco/Map/QuarterTile.h>
#include <OpenLoco/Core/Numerics.hpp>
int main(){ uint64_t h=1469598103934665603ull; for(int v=0;v<256;v++) for(int r=0;r<4;r++){ auto q=OpenLoco::World::QuarterTile(static_cast<uint8_t>(v)); uint8_t b; auto rq=q.rotate(r); __builtin_memcpy(&b,&rq,1); uint8_t l=OpenLoco::Numerics::rotl4bit(v&0xF,r); h^=b; h*=1099511628211ull; h^=l; h*=1099511628211ull; } printf("%llx\n",(unsigned long long)h); }
