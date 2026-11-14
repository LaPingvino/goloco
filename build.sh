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
echo "  Arrows: Scroll the view; PageUp/PageDown: Zoom out/in; Return: Rotate"
echo "  The original's shortcut keys (README.md)"
echo "  Mouse: Drag windows by title bar"
echo "  Click: X/OK/Cancel to close windows"
echo "  Click: Checkboxes to toggle"
echo "  Esc: Cancel the tool"
