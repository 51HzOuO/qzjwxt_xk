#!/bin/bash

# Script to compile Go application for Linux from macOS
echo "Building Linux executable from macOS..."

# Set environment variables for cross-compilation
export GOOS=linux
export GOARCH=amd64

# Compile the application
go build -o main_linux main.go

echo "Build complete: main_linux" 