@echo off
title clariphish - Phishing Simulation Platform
color 0A

echo ========================================
echo   clariphish - Starting Server
echo ========================================
echo.
echo Server will be available at:
echo.
echo   http://localhost:3333
echo.
echo Default Admin Login:
echo   Username: admin
echo   Password: changeme
echo.
echo ========================================
echo.
echo Server is starting...
echo Press Ctrl+C to stop the server
echo.

cd /d "%~dp0"
clariphish.exe
