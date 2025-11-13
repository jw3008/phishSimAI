@echo off
REM Setup phishSimAI with Gemini API Key
echo ==================================
echo phishSimAI - Quick Setup with API Key
echo ==================================
echo.

REM Check if API key is provided as argument
if "%~1"=="" (
    echo Usage: SETUP_WITH_API_KEY.bat YOUR_API_KEY
    echo.
    echo Example: SETUP_WITH_API_KEY.bat AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    echo.
    echo Get your free API key from: https://aistudio.google.com/app/apikey
    echo.
    pause
    exit /b 1
)

set API_KEY=%~1

echo Step 1: Building application...
go build -o phishSimAI.exe .
if %ERRORLEVEL% NEQ 0 (
    echo X Build failed
    pause
    exit /b 1
)
echo + Build successful
echo.

echo Step 2: Starting application temporarily to create database...
start /B phishSimAI.exe
echo Waiting for database to initialize...
timeout /t 5 /nobreak >nul
echo + Database should be initialized
echo.

echo Step 3: Configuring Gemini API key...
go run tools/setup_api_key.go "%API_KEY%"
if %ERRORLEVEL% NEQ 0 (
    echo X Failed to configure API key
    taskkill /F /IM phishSimAI.exe >nul 2>&1
    pause
    exit /b 1
)
echo.

echo Step 4: Stopping temporary instance...
taskkill /F /IM phishSimAI.exe >nul 2>&1
timeout /t 2 /nobreak >nul
echo + Stopped
echo.

echo ==================================
echo Setup Complete!
echo ==================================
echo.
echo To start the application, run:
echo   START.bat
echo.
echo Or start manually:
echo   phishSimAI.exe
echo.
echo The application will be available at:
echo   http://localhost:3333
echo.
echo Default credentials:
echo   Username: admin
echo   Password: admin
echo ==================================
echo.
pause
