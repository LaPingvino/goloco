// Package coords defines distinct types for the different coordinate systems
// used in OpenLoco / GoLoco to prevent accidental unit confusion.
//
// There are two primary systems:
//
//	WorldUnits — sub-tile world coordinates. 1 tile = 32 WorldUnits.
//	             Entity PosX/PosY and TileX/TileY are in this unit.
//	             Max map: 384 tiles * 32 = 12 288 WorldUnits (fits in int16).
//
//	TileCoord  — tile-grid index (0–383 for a standard 384×384 map).
//	             Used for tile element lookups and render tile loops.
//	             Derived from WorldUnits by dividing by 32.
//
// The types are distinct so that the Go compiler catches mixing them.
//
// OpenLoco reference: src/OpenLoco/src/World/World.hpp
//   kTileSize = 32, kMapColumns = 384, kMapRows = 384
package coords

// WorldUnits is a world-space coordinate in sub-tile units.
// 1 tile = 32 WorldUnits.
// Stored as int16 in entity structs (OpenLoco: coord_t = int16_t).
type WorldUnits int16

// Tile returns the tile-grid index (WorldUnits / 32).
func (w WorldUnits) Tile() TileCoord { return TileCoord(int(w) / 32) }

// SubOffset returns the fractional position within the current tile (0–31).
func (w WorldUnits) SubOffset() int { return int(w) & 31 }

// Float64 returns the WorldUnits value as float64 for rendering arithmetic.
func (w WorldUnits) Float64() float64 { return float64(w) }

// TileCoord is a tile-grid index (0–383 for a standard map).
// Always derived from WorldUnits by dividing by 32, never stored directly
// in entity slots (those use WorldUnits).
type TileCoord int

// World converts the tile index back to the tile-aligned WorldUnits origin.
func (t TileCoord) World() WorldUnits { return WorldUnits(int(t) * 32) }

// Int returns the TileCoord as a plain int for use in slice indexing / bounds.
func (t TileCoord) Int() int { return int(t) }
