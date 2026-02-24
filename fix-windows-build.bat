@echo off
REM Fix Windows Build Issues for PicoClaw
REM This script fixes common Windows build problems

echo PicoClaw Windows Build Fix Script
echo ==================================
echo.

REM Check Go installation
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

echo Step 1: Checking and creating necessary directories...
if not exist "cmd\picoclaw" (
    echo Error: cmd\picoclaw directory not found
    echo Please run this script from the picoclaw project root
    pause
    exit /b 1
)

echo Step 2: Copying workspace files (Windows compatible)...
if exist "workspace" (
    if exist "cmd\picoclaw\workspace" (
        rmdir /s /q "cmd\picoclaw\workspace"
    )
    xcopy /E /I /Y "workspace" "cmd\picoclaw\workspace" >nul 2>&1
    echo Workspace files copied successfully
) else (
    echo Warning: workspace directory not found
)

echo Step 3: Downloading Go dependencies...
go mod download
go mod verify

echo Step 4: Building with verbose output...
echo.
echo Build output:
echo ============
go build -v -tags stdjson -o "build\picoclaw.exe" "./cmd/picoclaw"
if %errorlevel% equ 0 (
    echo.
    echo ✓ Build successful!
    echo Output: build\picoclaw.exe
) else (
    echo.
    echo ✗ Build failed with error code %errorlevel%
    echo.
    echo Common solutions:
    echo 1. Make sure Go is properly installed
    echo 2. Run as administrator if needed
    echo 3. Check if Antivirus is blocking the build
    echo 4. Verify all source files are present
)

echo.
echo Step 5: Testing installation...
if exist "build\picoclaw.exe" (
    echo Testing the built binary...
    "build\picoclaw.exe" --help >nul 2>&1
    if %errorlevel% equ 0 (
        echo ✓ Binary test successful
    ) else (
        echo ⚠ Binary test failed (may be normal if --help not supported)
    )
)

echo.
echo Windows build fix complete!
echo.
echo If build succeeded, you can now run:
echo   build.bat install
echo.
pause