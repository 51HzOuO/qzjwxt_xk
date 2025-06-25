#!/bin/bash

# Script to compile Go application for macOS from macOS
echo "Building macOS executable..."

# Set environment variables for native compilation
export GOOS=darwin
export GOARCH=amd64

# Compile the application
go build -o main_darwin main.go

echo "Build complete: main_darwin" 