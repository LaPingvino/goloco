#!/bin/bash
# Build script for goloco

set -e

echo "🔨 Building goloco..."
go build -o goloco ./cmd/goloco

echo "✅ Build successful!"
echo ""
echo "🎮 To run:"
echo "  ./goloco"
echo ""
echo "📝 Controls:"
echo "  WASD / Arrows: Pan camera"
echo "  Q/E: Zoom in/out"
echo "  Mouse: Drag windows by title bar"
echo "  Click: X/OK/Cancel to close windows"
echo "  Click: Checkboxes to toggle"
echo "  Esc: Quit"
