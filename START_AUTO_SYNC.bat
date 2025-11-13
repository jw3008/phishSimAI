@echo off
title PhishSimAI Auto-Sync
color 0A
echo.
echo ========================================
echo   PhishSimAI Auto-Sync
echo ========================================
echo.
echo Starting auto-sync for GitHub...
echo This window must remain open for auto-sync to work.
echo.
echo Press Ctrl+C to stop auto-sync.
echo ========================================
echo.

cd /d "%~dp0"
powershell.exe -ExecutionPolicy Bypass -NoProfile -File "%~dp0auto-sync.ps1"

pause
