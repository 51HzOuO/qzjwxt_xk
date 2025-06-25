#!/bin/bash

# Script to compile Go application for macOS from macOS
echo "Building macOS executable..."

# Set environment variables for native compilation
export GOOS=darwin
export GOARCH=arm64

# Compile the application
go build -o qzjwxt_xk_macos_arm64 main.go

echo "Build complete: qzjwxt_xk_macos_arm64 (macOS ARM64)" 