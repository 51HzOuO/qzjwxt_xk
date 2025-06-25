#!/bin/bash

# Script to compile Go application for Windows from macOS
echo "Building Windows executable from macOS..."

# Set environment variables for cross-compilation
export GOOS=windows
export GOARCH=amd64

# Compile the application
go build -o qzjwxt_xk_windows_amd64.exe main.go

echo "Build complete: qzjwxt_xk_windows_amd64.exe (Windows AMD64)" 