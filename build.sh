#!/bin/bash

# ZenSort Build Script for Unix-like systems (macOS/Linux)
# Usage: ./build.sh

set -e  # Exit on any error

echo "Building ZenSort with CGO support for GUI..."

# Set CGO environment
export CGO_ENABLED=1

# Build the main executable
if go build -o zensort main.go; then
    echo ""
    echo "✓ Build successful! zensort supports both GUI and CLI modes."
    echo ""
    echo "Usage:"
    echo "  GUI Mode (default):    ./zensort"
    echo "  CLI Mode:              ./zensort -source /path/to/source -dest /path/to/dest"
    echo "  Force CLI:             ./zensort -cli -source /path/to/source -dest /path/to/dest"
    echo ""
    echo "Make executable if needed: chmod +x zensort"
else
    echo ""
    echo "✗ CGO build failed. Building CLI-only version..."
    export CGO_ENABLED=0
    
    if go build -tags nocgo -o zensort-cli-only main.go; then
        echo ""
        echo "✓ CLI-only build complete! Use ./zensort-cli-only for command-line interface."
        echo ""
        echo "To enable GUI support:"
        echo "  1. Install a C compiler (gcc, clang, or Xcode Command Line Tools)"
        echo "  2. Run: CGO_ENABLED=1 go build -o zensort main.go"
        echo ""
        echo "Usage:"
        echo "  ./zensort-cli-only -source /path/to/source -dest /path/to/dest"
    else
        echo ""
        echo "✗ Build failed completely. Check Go installation and dependencies."
        exit 1
    fi
fi
