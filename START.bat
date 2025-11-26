@echo off
REM clariphish Startup Script for Windows
REM This script will set up and run the phishing simulation platform

echo ==================================
echo clariphish - Setup and Start
echo ==================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo X Go is not installed!
    echo.
    echo Please install Go first:
    echo Download from https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

echo + Go is installed
go version
echo.

REM Check if we're in the right directory
if not exist "main.go" (
    echo X main.go not found!
    echo Please run this script from the phishSimAI directory
    echo.
    pause
    exit /b 1
)

echo + Found main.go
echo.

REM Install dependencies
echo Installing dependencies...
go mod download
if %ERRORLEVEL% EQU 0 (
    echo + Dependencies installed
) else (
    echo ! Warning: Some dependencies may not have installed correctly
)
echo.

REM Clean database (optional - comment out if you want to keep existing data)
if exist "clariphish.db" (
    echo Removing old database...
    del /f clariphish.db
    echo + Old database removed
    echo.
)

REM Build the application
echo Building application...
go build -o clariphish.exe .
if %ERRORLEVEL% EQU 0 (
    echo + Build successful
) else (
    echo X Build failed
    pause
    exit /b 1
)
echo.

REM Start the application
echo Starting clariphish...
echo.
echo ==================================
echo Server will start on: http://localhost:3333
echo Default credentials:
echo   Username: admin
echo   Password: admin
echo ==================================
echo.
echo Press Ctrl+C to stop the server
echo.

clariphish.exe
