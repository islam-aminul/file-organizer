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
    echo "Build complete! Run ./zensort to start the application."
else
    echo ""
    echo "CGO build failed. Building CLI-only version..."
    
    # Fallback to CLI-only build
    export CGO_ENABLED=0
    if go build -tags nocgo -o zensort-cli-only main.go; then
        echo ""
        echo "CLI-only build complete! Use ./zensort-cli-only for command-line interface."
        echo "To use GUI, install development tools:"
        echo "  Ubuntu/Debian: sudo apt install build-essential libgl1-mesa-dev xorg-dev"
        echo "  CentOS/RHEL: sudo yum groupinstall \"Development Tools\""
        echo "  macOS: xcode-select --install"
        echo "Then run: export CGO_ENABLED=1 && go build -o zensort main.go"
    else
        echo "Error: Both GUI and CLI builds failed!"
        exit 1
    fi
fi

echo ""
echo "Build process completed."
