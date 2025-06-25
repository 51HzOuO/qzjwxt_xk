#!/bin/bash

# Script to compile Go application for Windows from macOS
echo "Building Windows executable from macOS..."

# Set environment variables for cross-compilation
export GOOS=windows
export GOARCH=amd64

# Compile the application
go build -o main_windows.exe main.go

echo "Build complete: main_windows.exe" 