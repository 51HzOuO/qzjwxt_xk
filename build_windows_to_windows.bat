@echo off
REM Script to compile Go application for Windows from Windows
echo Building Windows executable...

REM Set environment variables for native compilation
set GOOS=windows
set GOARCH=amd64

REM Compile the application
go build -o main_windows.exe main.go

echo Build complete: main_windows.exe 