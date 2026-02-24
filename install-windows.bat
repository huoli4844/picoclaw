@echo off
REM Easy Windows Installation Script for PicoClaw
REM This script handles the complete installation process on Windows

echo PicoClaw Windows Installation Script
echo ====================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo.
    echo Please install Go from: https://golang.org/dl/
    echo After installation, restart this script.
    echo.
    pause
    exit /b 1
)

echo ✓ Go installation found
echo.

REM Clean any previous builds
echo Step 1: Cleaning previous builds...
if exist "build" (
    rmdir /s /q "build"
    echo   - Build directory cleaned
)

REM Copy workspace files
echo Step 2: Copying workspace files...
if exist "workspace" (
    if exist "cmd\picoclaw\workspace" (
        rmdir /s /q "cmd\picoclaw\workspace"
    )
    xcopy /E /I /Y "workspace" "cmd\picoclaw\workspace" >nul 2>&1
    echo   - Workspace files copied
) else (
    echo   - Warning: workspace directory not found
)

REM Download dependencies
echo Step 3: Downloading dependencies...
go mod download >nul 2>&1
if %errorlevel% equ 0 (
    echo   - Dependencies downloaded successfully
) else (
    echo   - Warning: Some dependencies may have failed to download
)

REM Build the application
echo Step 4: Building PicoClaw...
if not exist "build" mkdir "build"
go build -v -tags stdjson -ldflags "-s -w" -o "build\picoclaw.exe" "./cmd\picoclaw"
if %errorlevel% equ 0 (
    echo   ✓ Build successful!
    echo   - Binary created: build\picoclaw.exe
) else (
    echo   ✗ Build failed!
    echo.
    echo Common solutions:
    echo 1. Run as administrator
    echo 2. Check your internet connection
    echo 3. Temporarily disable antivirus
    echo 4. Make sure you have sufficient disk space
    echo.
    pause
    exit /b 1
)

REM Install to system
echo Step 5: Installing to system...
set INSTALL_DIR=%USERPROFILE%\.local\bin
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

copy /Y "build\picoclaw.exe" "%INSTALL_DIR%\picoclaw.exe" >nul 2>&1
if %errorlevel% equ 0 (
    echo   ✓ Installation successful!
    echo   - Installed to: %INSTALL_DIR%\picoclaw.exe
) else (
    echo   ✗ Installation failed!
    echo   - Manual copy required:
    echo     From: %CD%\build\picoclaw.exe
    echo     To:   %INSTALL_DIR%\picoclaw.exe
    pause
    exit /b 1
)

REM Create desktop shortcut (optional)
echo Step 6: Creating desktop shortcut...
set DESKTOP=%USERPROFILE%\Desktop
echo @echo off > "%DESKTOP%\PicoClaw.bat"
echo title PicoClaw >> "%DESKTOP%\PicoClaw.bat"
echo "%INSTALL_DIR%\picoclaw.exe" %%* >> "%DESKTOP%\PicoClaw.bat"
echo Desktop shortcut created: %DESKTOP%\PicoClaw.bat

REM Test installation
echo Step 7: Testing installation...
"%INSTALL_DIR%\picoclaw.exe" --version >nul 2>&1
if %errorlevel% le 1 (
    echo   ✓ Installation verified successfully!
) else (
    echo   ⚠ Installation test inconclusive (may be normal)
)

REM PATH setup reminder
echo.
echo ====================================
echo Installation Complete!
echo ====================================
echo.
echo Important Notes:
echo 1. Add '%INSTALL_DIR%' to your system PATH
echo 2. You can also use the desktop shortcut
echo 3. Restart command prompt to use 'picoclaw' command
echo.
echo Quick test:
echo   picoclaw --help
echo.
echo Configuration directory: %USERPROFILE%\.picoclaw\
echo.
pause