#!/bin/bash

# Script to compile Go application for all platforms from macOS
echo "Building executables for all platforms..."

# Build for macOS (ARM64)
echo "Building for macOS (ARM64)..."
GOOS=darwin GOARCH=arm64 go build -o qzjwxt_xk_macos_arm64 main.go

# Build for Linux (AMD64)
echo "Building for Linux (AMD64)..."
GOOS=linux GOARCH=amd64 go build -o qzjwxt_xk_linux_amd64 main.go

# Build for Windows (AMD64)
echo "Building for Windows (AMD64)..."
GOOS=windows GOARCH=amd64 go build -o qzjwxt_xk_windows_amd64.exe main.go

echo "All builds complete!"
echo "- qzjwxt_xk_macos_arm64 (macOS ARM64)"
echo "- qzjwxt_xk_linux_amd64 (Linux AMD64)"
echo "- qzjwxt_xk_windows_amd64.exe (Windows AMD64)" 