@echo off
REM PicoClaw Quick Start - One Click to Run
REM Just double-click this file to start PicoClaw

echo 🦞 Starting PicoClaw...
echo.

REM Check if PicoClaw is installed
if not exist "%USERPROFILE%\.local\bin\picoclaw.exe" (
    echo PicoClaw not found. Installing now...
    call install-windows.bat
    echo.
)

REM Start web server and open browser
echo Starting Web Interface...
echo Opening http://localhost:8080 in your browser...
echo.

start http://localhost:8080
"%USERPROFILE%\.local\bin\picoclaw.exe" gateway

pause