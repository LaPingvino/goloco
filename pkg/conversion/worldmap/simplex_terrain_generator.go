package worldmap

// AUTO-GENERATED FROM C++ - WILL NOT COMPILE
// This is a mechanical translation that needs manual fixing

// #include "HeightMapRange.h"
// #include <cstdint>
// #include <random>
// namespace OpenLoco::Scenario
// forward: struct Options;
// namespace OpenLoco::World::MapGenerator
type SimplexTerrainGenerator struct {
	// method: void generate(const Scenario::Options& options, HeightMapRange heightMap, uint32_t seed);
}

type SimplexTerrainGeneratorSimplexSettings struct {
// int32_t low = 2;
// int32_t high = 24;
// float baseFreq = 1.25f;
// int32_t octaves = 4;
// int32_t smooth = 2;
}
// std::mt19937 _pprng;
	// method: void initialiseRng(uint32_t seed);
	// method: void generate(const SimplexSettings& settings, HeightMapRange heightMap);
	// method: void generateSimplex(const SimplexSettings& settings, HeightMapRange heightMap);
	// method: static void smooth(int32_t iterations, HeightMapRange heightMap);
	// method: static float noiseFractal(uint8_t* perm, int32_t x, int32_t y, float frequency, int32_t octaves, float lacunarity, float persistence);
	// method: static float generateNoise(uint8_t* perm, float x, float y);
	// method: void noise(uint8_t* perm, size_t len);
	// method: static float grad(int32_t hash, float x, float y);
	// method: static int32_t fastFloor(float x);
	// method: uint32_t randomNext();
	// method: int32_t randomNext(int32_t min, int32_t max);
}
