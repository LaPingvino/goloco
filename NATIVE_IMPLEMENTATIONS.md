# Native Go Implementations

This file tracks C++ constructs from OpenLoco that were replaced with native Go implementations
instead of being mechanically translated.

## Math Package

### Square Root (FastSquareRoot)
- **Original**: Lookup table with 2048 entries at address 0x00500B50
- **Native**: Uses `math.Sqrt()` from Go standard library
- **Reason**: Go's math.Sqrt is highly optimized and more maintainable than a lookup table
- **Files affected**: `pkg/conversion/math/vector.go`

### Distance Functions (Distance2D, Distance3D)
- **Original**: Used FastSquareRoot lookup table
- **Native**: Uses standard library sqrt via the Sqrt wrapper function
- **Files affected**: `pkg/conversion/math/vector.go`

## Pending Native Implementations

These items may need native Go implementations as we encounter them:

- [ ] Trigonometric functions (if lookup tables exist in OpenLoco)
- [ ] Random number generation (may need compatible RNG for save compatibility)
- [ ] File I/O (Go's os package vs C++ file handling)
- [ ] String handling (Go strings vs C++ std::string)
- [ ] Memory management patterns (C++ new/delete vs Go GC)
